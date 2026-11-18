package runner

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/workspace"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	noteStopped          = "stopped"
	noteFailed           = "failed"
	noteNoSubscription   = "noSubscription"
	noteUnsupportedAgent = "unsupportedAgent"

	toolRunning = "running"
	toolDone    = "done"
	toolError   = "error"
	toolStopped = "stopped"

	maxOutputLines = 400
	maxLineRunes   = 2000
	maxInputRunes  = 600
)

type textBody struct {
	Text string `json:"text"`
}

type thinkingBody struct {
	Text   string `json:"text"`
	Status string `json:"status"`
}

type noteBody struct {
	Code string `json:"code"`
	Text string `json:"text,omitempty"`
}

type toolBody struct {
	Name   string   `json:"name"`
	Input  string   `json:"input"`
	Output []string `json:"output,omitempty"`
	Status string   `json:"status"`
	Meta   string   `json:"meta,omitempty"`
}

type openTool struct {
	id   pgtype.UUID
	body toolBody
}

type transcript struct {
	r       *Runner
	owner   string
	session pgtype.UUID
	tools   map[string]*openTool
	order   []string
}

func newTranscript(r *Runner, owner string, session pgtype.UUID) *transcript {
	return &transcript{r: r, owner: owner, session: session, tools: make(map[string]*openTool)}
}

func (t *transcript) add(role, kind string, body any) (db.Message, bool) {
	raw, err := json.Marshal(body)
	if err != nil {
		t.r.log.Warn("encode message", "err", err)
		return db.Message{}, false
	}
	m, err := t.r.q.CreateMessage(t.r.ctx, db.CreateMessageParams{SessionID: t.session, Role: role, Kind: kind, Body: raw})
	if err != nil {
		t.r.log.Warn("store agent message", "err", err)
		return db.Message{}, false
	}
	t.r.hub.Publish(t.owner, "message.created", workspace.MessageOf(m), "")
	return m, true
}

func (t *transcript) update(id pgtype.UUID, body any) {
	raw, err := json.Marshal(body)
	if err != nil {
		t.r.log.Warn("encode message", "err", err)
		return
	}
	m, err := t.r.q.UpdateMessageBody(t.r.ctx, db.UpdateMessageBodyParams{ID: id, Body: raw})
	if err != nil {
		t.r.log.Warn("update agent message", "err", err)
		return
	}
	t.r.hub.Publish(t.owner, "message.updated", workspace.MessageOf(m), "")
}

func (t *transcript) text(text string) {
	if strings.TrimSpace(text) != "" {
		t.add("assistant", "text", textBody{Text: text})
	}
}

func (t *transcript) thinking(text string) {
	if strings.TrimSpace(text) != "" {
		t.add("assistant", "thinking", thinkingBody{Text: text, Status: toolDone})
	}
}

func (t *transcript) note(code, text string) {
	t.add("system", "note", noteBody{Code: code, Text: text})
}

func (t *transcript) startTool(callID, name, input string) {
	body := toolBody{Name: name, Input: clip(input, maxInputRunes), Status: toolRunning}
	m, ok := t.add("assistant", "tool", body)
	if !ok {
		return
	}
	t.tools[callID] = &openTool{id: m.ID, body: body}
	t.order = append(t.order, callID)
}

func (t *transcript) finishTool(callID, output string, failed bool) {
	tool, ok := t.tools[callID]
	if !ok {
		return
	}
	delete(t.tools, callID)
	tool.body.Output = outputLines(output)
	tool.body.Status = toolDone
	if failed {
		tool.body.Status = toolError
	}
	t.update(tool.id, tool.body)
}

func (t *transcript) closeTools(status string) {
	for _, callID := range t.order {
		tool, ok := t.tools[callID]
		if !ok {
			continue
		}
		delete(t.tools, callID)
		tool.body.Status = status
		t.update(tool.id, tool.body)
	}
	t.order = nil
}

func outputLines(output string) []string {
	output = strings.TrimRight(output, "\n")
	if output == "" {
		return nil
	}
	lines := strings.Split(output, "\n")
	if len(lines) > maxOutputLines {
		lines = append(lines[:maxOutputLines], "…")
	}
	for i, line := range lines {
		lines[i] = clip(line, maxLineRunes)
	}
	return lines
}

func clip(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	return string([]rune(s)[:limit]) + "…"
}

func uuidString(id pgtype.UUID) string {
	v, _ := id.Value()
	s, _ := v.(string)
	return s
}

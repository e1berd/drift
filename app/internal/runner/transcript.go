package runner

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/e1berd/drift/app/internal/agents"
	"github.com/e1berd/drift/app/internal/askuser"
	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/workspace"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	noteStopped          = "stopped"
	noteFailed           = "failed"
	noteNoSubscription   = "noSubscription"
	noteUnsupportedAgent = "unsupportedAgent"
	noteUsageWarning     = "usageWarning"

	kindQuestion = "question"
	kindLimit    = "limit"

	toolRunning = "running"
	toolDone    = "done"
	toolError   = "error"
	toolStopped = "stopped"

	phaseThinking = "thinking"
	phaseWriting  = "writing"
	phaseTool     = "tool"

	activityInterval = 120 * time.Millisecond

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

// Activity is what the agent is doing right now; it is sent to clients and never stored.
type Activity struct {
	SessionID string `json:"sessionId"`
	Phase     string `json:"phase"`
	Tool      string `json:"tool,omitempty"`
	Text      string `json:"text,omitempty"`
	Tokens    int    `json:"tokens,omitempty"`
}

type limitHit struct {
	window agents.Window
	text   string
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

	lastPhase    string
	lastActivity time.Time

	windows map[string]agents.Window
	limit   *limitHit
}

func newTranscript(r *Runner, owner string, session pgtype.UUID) *transcript {
	return &transcript{
		r: r, owner: owner, session: session,
		tools: make(map[string]*openTool), windows: make(map[string]agents.Window),
	}
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
	t.r.updateMessage(t.owner, id, body)
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

func (t *transcript) question(questions []askuser.Question) {
	t.add("assistant", kindQuestion, askuser.Card{Questions: questions, Status: askuser.CardPending})
}

func (t *transcript) usage(window agents.Window) {
	t.windows[window.ID] = window
}

func (t *transcript) limitReached(window agents.Window, text string) {
	t.limit = &limitHit{window: window, text: text}
}

func (t *transcript) activity(phase, tool, text string, tokens int) {
	now := time.Now()
	if phase == t.lastPhase && phase != phaseTool && now.Sub(t.lastActivity) < activityInterval {
		return
	}
	t.lastPhase, t.lastActivity = phase, now
	t.r.hub.Publish(t.owner, "session.activity", Activity{
		SessionID: uuidString(t.session), Phase: phase, Tool: tool, Text: text, Tokens: tokens,
	}, "")
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

func (r *Runner) updateMessage(owner string, id pgtype.UUID, body any) {
	raw, err := json.Marshal(body)
	if err != nil {
		r.log.Warn("encode message", "err", err)
		return
	}
	m, err := r.q.UpdateMessageBody(r.ctx, db.UpdateMessageBodyParams{ID: id, Body: raw})
	if err != nil {
		r.log.Warn("update agent message", "err", err)
		return
	}
	r.hub.Publish(owner, "message.updated", workspace.MessageOf(m), "")
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

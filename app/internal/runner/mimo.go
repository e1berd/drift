package runner

import (
	"encoding/json"
	"strings"

	"github.com/e1berd/drift/app/internal/workspace"
)

func mimoArgs(req workspace.RunRequest, agentSessionID string, images []string) []string {
	args := []string{"run", "--format", "json", "--dangerously-skip-permissions"}
	if req.Model != "" && req.Model != "default" {
		args = append(args, "-m", req.Model)
	}
	if agentSessionID != "" {
		args = append(args, "-s", agentSessionID)
	}
	for _, image := range images {
		args = append(args, "-f", image)
	}
	return args
}

type mimoEvent struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionID"`
	Text      string          `json:"text"`
	Part      *mimoPart       `json:"part"`
	Error     *mimoError      `json:"error"`
}

type mimoPart struct {
	ID    string          `json:"id"`
	Type  string          `json:"type"`
	Text  string          `json:"text"`
	Tool  string          `json:"tool"`
	State mimoToolState   `json:"state"`
}

type mimoToolState struct {
	Status string          `json:"status"`
	Input  json.RawMessage `json:"input"`
	Output json.RawMessage `json:"output"`
	Error  string          `json:"error"`
}

type mimoError struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Data    struct {
		Message string `json:"message"`
	} `json:"data"`
}

type mimoStream struct {
	t             transcriptSink
	projectDir    string
	saveSessionID func(string)
	finished      bool
	reported      bool
	pendingText   strings.Builder
}

func (s *mimoStream) done() bool { return s.finished }

func (s *mimoStream) handle(line []byte) {
	var event mimoEvent
	if json.Unmarshal(line, &event) != nil {
		return
	}
	if event.SessionID != "" && !s.reported {
		s.reported = true
		s.saveSessionID(event.SessionID)
	}
	switch event.Type {
	case "text":
		if event.Part != nil && strings.TrimSpace(event.Part.Text) != "" {
			s.pendingText.WriteString(event.Part.Text)
			s.pendingText.WriteString("\n\n")
			s.t.activity(phaseWriting, "", s.pendingText.String(), 0)
		}
	case "tool_use":
		if event.Part != nil {
			s.tool(*event.Part)
		}
	case "step_finish":
		s.flushText()
		s.t.activity(phaseThinking, "", "", 0)
	case "step_start":
		s.t.activity(phaseThinking, "", "", 0)
	case "error":
		s.fail(event.Error)
	}
}

func (s *mimoStream) flushText() {
	text := strings.TrimSpace(s.pendingText.String())
	s.pendingText.Reset()
	if text != "" {
		s.t.text(text)
	}
}

func (s *mimoStream) tool(part mimoPart) {
	input := strings.ReplaceAll(toolInput(part.State.Input), s.projectDir+"/", "")
	failed := part.State.Status == "error"
	output := firstNonEmpty(mimoToolOutput(part.State.Output), part.State.Error)
	s.t.startTool(part.ID, part.Tool, input)
	s.t.finishTool(part.ID, output, failed)
	s.t.activity(phaseThinking, "", "", 0)
}

func (s *mimoStream) fail(err *mimoError) {
	s.flushText()
	s.finished = true
	s.t.closeTools(toolError)
	if err == nil {
		s.t.note(noteFailed, "")
		return
	}
	s.t.note(noteFailed, firstNonEmpty(err.Data.Message, err.Message, err.Name))
}

func mimoToolOutput(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var fields map[string]any
	if json.Unmarshal(raw, &fields) == nil {
		for _, key := range toolInputKeys {
			if value, ok := fields[key].(string); ok && value != "" {
				return value
			}
		}
	}
	return string(raw)
}

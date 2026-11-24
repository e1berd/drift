package runner

import (
	"encoding/json"
	"strings"

	"github.com/e1berd/drift/app/internal/workspace"
)

// Kimi Code 2.7 rejects every permission mode flag (--auto, --yolo, --plan) together with
// --prompt and approves everything in prompt mode, so the session mode is not passed.
func kimiArgs(req workspace.RunRequest, agentSessionID string) []string {
	args := []string{"-p", req.Text, "--output-format", "stream-json"}
	switch req.Model {
	case "", "default", "kimi-code":
	default:
		args = append(args, "-m", req.Model)
	}
	if agentSessionID != "" {
		args = append(args, "-S", agentSessionID)
	}
	return args
}

type kimiEvent struct {
	Role       string         `json:"role"`
	Type       string         `json:"type"`
	SessionID  string         `json:"session_id"`
	Content    string         `json:"content"`
	ToolCallID string         `json:"tool_call_id"`
	ToolCalls  []kimiToolCall `json:"tool_calls"`
	Message    string         `json:"message"`
	Error      string         `json:"error"`
}

type kimiToolCall struct {
	ID       string `json:"id"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type kimiStream struct {
	t             transcriptSink
	projectDir    string
	saveSessionID func(string)
	finished      bool
	started       map[string]bool
}

func (s *kimiStream) done() bool { return s.finished }

func (s *kimiStream) handle(line []byte) {
	var event kimiEvent
	if json.Unmarshal(line, &event) != nil {
		return
	}
	if event.Type == "error" {
		s.finished = true
		s.t.closeTools(toolError)
		s.t.note(noteFailed, firstNonEmpty(event.Message, event.Error))
		return
	}
	switch event.Role {
	case "meta":
		s.meta(event)
	case "assistant":
		s.assistant(event)
	case "tool":
		s.toolResult(event)
	}
}

func (s *kimiStream) meta(event kimiEvent) {
	if event.Type == "session.resume_hint" && event.SessionID != "" {
		s.saveSessionID(event.SessionID)
	}
}

func (s *kimiStream) assistant(event kimiEvent) {
	if strings.TrimSpace(event.Content) != "" {
		s.t.text(event.Content)
	}
	for _, call := range event.ToolCalls {
		if s.started == nil {
			s.started = make(map[string]bool)
		}
		if s.started[call.ID] {
			continue
		}
		s.started[call.ID] = true
		input := strings.ReplaceAll(kimiToolInput(call.Function.Arguments), s.projectDir+"/", "")
		s.t.startTool(call.ID, call.Function.Name, input)
		s.t.activity(phaseTool, call.Function.Name, "", 0)
	}
}

func (s *kimiStream) toolResult(event kimiEvent) {
	if event.ToolCallID == "" {
		return
	}
	s.t.finishTool(event.ToolCallID, event.Content, false)
	s.t.activity(phaseThinking, "", "", 0)
}

func kimiToolInput(arguments string) string {
	var fields map[string]any
	if json.Unmarshal([]byte(arguments), &fields) != nil {
		return arguments
	}
	for _, key := range toolInputKeys {
		if value, ok := fields[key].(string); ok && value != "" {
			return value
		}
	}
	if len(fields) == 0 {
		return ""
	}
	return arguments
}

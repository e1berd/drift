package runner

import (
	"encoding/json"
	"strings"
)

type claudeEvent struct {
	Type            string  `json:"type"`
	Subtype         string  `json:"subtype"`
	SessionID       string  `json:"session_id"`
	ParentToolUseID *string `json:"parent_tool_use_id"`
	Message         struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
	IsError bool   `json:"is_error"`
	Result  string `json:"result"`
}

type claudeBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Thinking  string          `json:"thinking"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
	IsError   bool            `json:"is_error"`
}

var toolInputKeys = []string{"command", "file_path", "notebook_path", "pattern", "url", "query", "description", "prompt"}

type transcriptSink interface {
	text(text string)
	thinking(text string)
	note(code, text string)
	startTool(callID, name, input string)
	finishTool(callID, output string, failed bool)
	closeTools(status string)
}

type claudeStream struct {
	t              transcriptSink
	projectDir     string
	agentSessionID string
	saveSessionID  func(string)
	finished       bool
}

func (s *claudeStream) handle(line []byte) {
	var event claudeEvent
	if json.Unmarshal(line, &event) != nil {
		return
	}
	if event.ParentToolUseID != nil {
		return
	}
	switch event.Type {
	case "system":
		if event.Subtype == "init" && event.SessionID != "" && event.SessionID != s.agentSessionID {
			s.agentSessionID = event.SessionID
			s.saveSessionID(event.SessionID)
		}
	case "assistant":
		for _, block := range blocks(event.Message.Content) {
			s.assistantBlock(block)
		}
	case "user":
		for _, block := range blocks(event.Message.Content) {
			if block.Type == "tool_result" {
				s.t.finishTool(block.ToolUseID, resultText(block.Content), block.IsError)
			}
		}
	case "result":
		s.finished = true
		if event.IsError || event.Subtype != "success" {
			s.t.closeTools(toolError)
			s.t.note(noteFailed, firstNonEmpty(event.Result, event.Subtype))
		}
	}
}

func (s *claudeStream) assistantBlock(block claudeBlock) {
	switch block.Type {
	case "text":
		s.t.text(block.Text)
	case "thinking":
		s.t.thinking(block.Thinking)
	case "tool_use":
		s.t.startTool(block.ID, block.Name, strings.ReplaceAll(toolInput(block.Input), s.projectDir+"/", ""))
	}
}

func blocks(raw json.RawMessage) []claudeBlock {
	var out []claudeBlock
	if json.Unmarshal(raw, &out) != nil {
		return nil
	}
	return out
}

func resultText(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	parts := make([]string, 0)
	for _, block := range blocks(raw) {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func toolInput(raw json.RawMessage) string {
	var fields map[string]any
	if json.Unmarshal(raw, &fields) != nil {
		return string(raw)
	}
	for _, key := range toolInputKeys {
		if value, ok := fields[key].(string); ok && value != "" {
			return value
		}
	}
	if len(fields) == 0 {
		return ""
	}
	return string(raw)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

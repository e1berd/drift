package runner

import (
	"encoding/json"
	"strings"

	"github.com/e1berd/drift/app/internal/agents"
	"github.com/e1berd/drift/app/internal/askuser"
)

const claudeRateLimitError = "rate_limit"

var claudeAskTool = "mcp__" + askuser.ServerName + "__" + askuser.ToolName

type claudeEvent struct {
	Type            string  `json:"type"`
	Subtype         string  `json:"subtype"`
	SessionID       string  `json:"session_id"`
	ParentToolUseID *string `json:"parent_tool_use_id"`
	Message         struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
	IsError         bool                 `json:"is_error"`
	Error           string               `json:"error"`
	Result          string               `json:"result"`
	EstimatedTokens int                  `json:"estimated_tokens"`
	Event           *streamEvent         `json:"event"`
	RateLimit       *claudeRateLimitInfo `json:"rate_limit_info"`
}

type streamEvent struct {
	Type         string       `json:"type"`
	ContentBlock *claudeBlock `json:"content_block"`
	Delta        struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
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
	activity(phase, tool, text string, tokens int)
	question(questions []askuser.Question)
	usage(window agents.Window)
	limitReached(window agents.Window, text string)
}

type claudeStream struct {
	t              transcriptSink
	projectDir     string
	agentSessionID string
	saveSessionID  func(string)
	finished       bool
	partial        strings.Builder
	thinkingTokens int
	limits         claudeLimits
}

func (s *claudeStream) done() bool { return s.finished }

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
		s.systemEvent(event)
	case "stream_event":
		if event.Event != nil {
			s.streamEvent(*event.Event)
		}
	case "rate_limit_event":
		if event.RateLimit != nil {
			s.limits.observe(*event.RateLimit, s.t)
		}
	case "assistant":
		if event.Error == claudeRateLimitError {
			s.limits.refusal = resultText(event.Message.Content)
			return
		}
		for _, block := range blocks(event.Message.Content) {
			s.assistantBlock(block)
		}
	case "user":
		for _, block := range blocks(event.Message.Content) {
			if block.Type == "tool_result" {
				s.t.finishTool(block.ToolUseID, resultText(block.Content), block.IsError)
				s.thinkingTokens = 0
				s.t.activity(phaseThinking, "", "", 0)
			}
		}
	case "result":
		s.finished = true
		if event.IsError || event.Subtype != "success" {
			s.failed(firstNonEmpty(event.Result, s.limits.refusal, event.Subtype))
		}
	}
}

func (s *claudeStream) failed(message string) {
	s.t.closeTools(toolError)
	if window, ok := s.limits.exhausted(message); ok {
		s.t.limitReached(window, message)
		return
	}
	s.t.note(noteFailed, message)
}

func (s *claudeStream) systemEvent(event claudeEvent) {
	switch event.Subtype {
	case "init":
		if event.SessionID != "" && event.SessionID != s.agentSessionID {
			s.agentSessionID = event.SessionID
			s.saveSessionID(event.SessionID)
		}
	case "thinking_tokens":
		s.thinkingTokens = event.EstimatedTokens
		s.t.activity(phaseThinking, "", "", s.thinkingTokens)
	}
}

func (s *claudeStream) streamEvent(event streamEvent) {
	switch event.Type {
	case "content_block_start":
		if event.ContentBlock != nil {
			s.blockStarted(*event.ContentBlock)
		}
	case "content_block_delta":
		if event.Delta.Type == "text_delta" {
			s.partial.WriteString(event.Delta.Text)
			s.t.activity(phaseWriting, "", s.partial.String(), 0)
		}
	}
}

func (s *claudeStream) blockStarted(block claudeBlock) {
	switch block.Type {
	case "thinking":
		s.thinkingTokens = 0
		s.t.activity(phaseThinking, "", "", 0)
	case "text":
		s.partial.Reset()
		s.t.activity(phaseWriting, "", "", 0)
	case "tool_use":
		if block.Name != claudeAskTool {
			s.t.activity(phaseTool, block.Name, "", 0)
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
		if block.Name == claudeAskTool {
			askQuestions(s.t, block.Input)
			return
		}
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

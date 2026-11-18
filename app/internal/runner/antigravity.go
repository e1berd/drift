package runner

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/e1berd/drift/app/internal/workspace"
)

var antigravityModes = map[string]string{
	"acceptEdits": "accept-edits",
	"auto":        "auto",
	"plan":        "plan",
}

func antigravityArgs(req workspace.RunRequest, conversationID string) []string {
	args := []string{"--output-format", "stream-json"}
	if req.Model != "" && req.Model != "default" {
		args = append(args, "--model", req.Model)
	}
	if mode, ok := antigravityModes[req.Mode]; ok {
		args = append(args, "--mode", mode)
	}
	if req.Mode == "bypassPermissions" {
		args = append(args, "--dangerously-skip-permissions")
	}
	if conversationID != "" {
		args = append(args, "--conversation", conversationID)
	}
	return append(args, "--prompt="+req.Text)
}

type antigravityEvent struct {
	Type       string          `json:"type"`
	Event      string          `json:"event"`
	StepUpdate json.RawMessage `json:"step_update"`
	Result     json.RawMessage `json:"result"`
}

type antigravityStep struct {
	ConversationID string          `json:"conversation_id"`
	StepIndex      *int            `json:"step_index"`
	State          string          `json:"state"`
	StepType       string          `json:"step_type"`
	ToolName       string          `json:"tool_name"`
	TextDelta      string          `json:"text_delta"`
	ToolInfo       antigravityTool `json:"tool_info"`
	SubagentInfo   json.RawMessage `json:"subagent_info"`
	Usage          struct {
		ThinkingTokens int `json:"thinking_tokens"`
	} `json:"usage"`
}

type antigravityTool struct {
	Name       string          `json:"name"`
	Parameters json.RawMessage `json:"parameters"`
	Output     json.RawMessage `json:"output"`
	Status     string          `json:"status"`
	Error      string          `json:"error"`
}

type antigravityFinal struct {
	ConversationID string `json:"conversation_id"`
	Status         string `json:"status"`
	Error          string `json:"error"`
	FinishedReason string `json:"finished_reason"`
}

type antigravityStream struct {
	t              transcriptSink
	projectDir     string
	conversationID string
	saveSessionID  func(string)
	finished       bool
	replies        map[int]*strings.Builder
	tools          map[int]bool
}

func (s *antigravityStream) done() bool { return s.finished }

func (s *antigravityStream) handle(line []byte) {
	var event antigravityEvent
	if json.Unmarshal(line, &event) != nil {
		return
	}
	switch firstNonEmpty(event.Type, event.Event) {
	case "step_update":
		var step antigravityStep
		if json.Unmarshal(payload(event.StepUpdate, line), &step) == nil {
			s.remember(step.ConversationID)
			s.step(step)
		}
	case "result":
		var final antigravityFinal
		if json.Unmarshal(payload(event.Result, line), &final) == nil {
			s.remember(final.ConversationID)
			s.finish(final)
		}
	}
}

func payload(nested json.RawMessage, line []byte) []byte {
	if len(nested) > 0 && nested[0] == '{' {
		return nested
	}
	return line
}

func (s *antigravityStream) remember(conversationID string) {
	if conversationID != "" && conversationID != s.conversationID {
		s.conversationID = conversationID
		s.saveSessionID(conversationID)
	}
}

func (s *antigravityStream) step(step antigravityStep) {
	if step.StepIndex == nil {
		return
	}
	index := *step.StepIndex
	switch step.StepType {
	case "agent_response":
		if len(step.SubagentInfo) == 0 || string(step.SubagentInfo) == "null" {
			s.reply(index, step)
		}
	case "tool":
		s.tool(index, step)
	}
}

func (s *antigravityStream) reply(index int, step antigravityStep) {
	if s.replies == nil {
		s.replies = make(map[int]*strings.Builder)
	}
	text, ok := s.replies[index]
	if !ok {
		text = &strings.Builder{}
		s.replies[index] = text
	}
	if step.TextDelta != "" {
		text.WriteString(step.TextDelta)
		s.t.activity(phaseWriting, "", text.String(), 0)
	} else if step.State != "DONE" && text.Len() == 0 {
		s.t.activity(phaseThinking, "", "", step.Usage.ThinkingTokens)
	}
	if step.State == "DONE" {
		if strings.TrimSpace(text.String()) != "" {
			s.t.text(text.String())
		}
		delete(s.replies, index)
		s.t.activity(phaseThinking, "", "", 0)
	}
}

func (s *antigravityStream) tool(index int, step antigravityStep) {
	if s.tools == nil {
		s.tools = make(map[int]bool)
	}
	callID := "step-" + strconv.Itoa(index)
	name := firstNonEmpty(step.ToolInfo.Name, step.ToolName)
	if !s.tools[index] {
		s.tools[index] = true
		input := strings.ReplaceAll(toolInput(step.ToolInfo.Parameters), s.projectDir+"/", "")
		s.t.startTool(callID, name, input)
		s.t.activity(phaseTool, name, "", 0)
	}
	if step.State == "DONE" {
		failed := step.ToolInfo.Error != "" || strings.EqualFold(step.ToolInfo.Status, "error")
		s.t.finishTool(callID, firstNonEmpty(toolOutput(step.ToolInfo.Output), step.ToolInfo.Error), failed)
		s.t.activity(phaseThinking, "", "", 0)
	}
}

func (s *antigravityStream) finish(final antigravityFinal) {
	s.finished = true
	for index, text := range s.replies {
		s.t.text(text.String())
		delete(s.replies, index)
	}
	switch final.Status {
	case "", "SUCCESS", "CANCELED", "INTERRUPTED":
		return
	}
	s.t.closeTools(toolError)
	s.t.note(noteFailed, firstNonEmpty(final.Error, final.FinishedReason, final.Status))
}

func toolOutput(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	return resultText(raw)
}

package runner

import (
	"encoding/json"
	"strings"

	"github.com/e1berd/drift/app/internal/agents"
	"github.com/e1berd/drift/app/internal/askuser"
	"github.com/e1berd/drift/app/internal/workspace"
)

var codexSandboxes = map[string]string{
	"":            "read-only",
	"default":     "read-only",
	"plan":        "read-only",
	"acceptEdits": "workspace-write",
	"auto":        "workspace-write",
}

func codexArgs(req workspace.RunRequest, threadID string, images []string, askMCP mcpCommand) []string {
	args := []string{"exec"}
	if threadID != "" {
		args = append(args, "resume")
	}
	args = append(args, "--json", "--skip-git-repo-check")
	if req.Model != "" && req.Model != "default" {
		args = append(args, "--model", req.Model)
	}
	if req.Mode == "bypassPermissions" {
		args = append(args, "--dangerously-bypass-approvals-and-sandbox")
	} else {
		args = append(args, "-c", `sandbox_mode="`+codexSandboxes[req.Mode]+`"`, "-c", `approval_policy="never"`)
	}
	args = append(args, askMCP.codexArgs()...)
	for _, image := range images {
		args = append(args, "--image="+image)
	}
	args = append(args, "--")
	if threadID != "" {
		args = append(args, threadID)
	}
	return append(args, "-")
}

type codexEvent struct {
	Type     string     `json:"type"`
	ThreadID string     `json:"thread_id"`
	Message  string     `json:"message"`
	Error    *codexErr  `json:"error"`
	Item     *codexItem `json:"item"`
}

type codexErr struct {
	Message string `json:"message"`
}

type codexItem struct {
	ID               string          `json:"id"`
	Type             string          `json:"type"`
	Text             string          `json:"text"`
	Status           string          `json:"status"`
	Command          string          `json:"command"`
	AggregatedOutput string          `json:"aggregated_output"`
	ExitCode         *int            `json:"exit_code"`
	Changes          []codexChange   `json:"changes"`
	Server           string          `json:"server"`
	Tool             string          `json:"tool"`
	Arguments        json.RawMessage `json:"arguments"`
	Result           *codexMCPResult `json:"result"`
	Error            *codexErr       `json:"error"`
	Query            string          `json:"query"`
	Items            []codexTodo     `json:"items"`
}

type codexChange struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

type codexMCPResult struct {
	Content json.RawMessage `json:"content"`
}

type codexTodo struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

type codexStream struct {
	t             transcriptSink
	projectDir    string
	threadID      string
	saveSessionID func(string)
	finished      bool
	started       map[string]bool
}

func (s *codexStream) done() bool { return s.finished }

func (s *codexStream) handle(line []byte) {
	var event codexEvent
	if json.Unmarshal(line, &event) != nil {
		return
	}
	switch event.Type {
	case "thread.started":
		if event.ThreadID != "" && event.ThreadID != s.threadID {
			s.threadID = event.ThreadID
			s.saveSessionID(event.ThreadID)
		}
	case "turn.started":
		s.t.activity(phaseThinking, "", "", 0)
	case "item.started", "item.updated":
		if event.Item != nil {
			s.itemStarted(*event.Item)
		}
	case "item.completed":
		if event.Item != nil {
			s.itemCompleted(*event.Item)
		}
	case "turn.completed":
		s.finished = true
	case "turn.failed":
		s.finished = true
		s.t.closeTools(toolError)
		message := ""
		if event.Error != nil {
			message = event.Error.Message
		}
		s.failed(firstNonEmpty(message, event.Message))
	}
}

func (s *codexStream) failed(message string) {
	if limitExhaustedMessage.MatchString(message) {
		s.t.limitReached(agents.Window{}, message)
		return
	}
	s.t.note(noteFailed, message)
}

func (s *codexStream) itemStarted(item codexItem) {
	if item.asksUser() {
		return
	}
	switch item.Type {
	case "reasoning":
		s.t.activity(phaseThinking, "", "", 0)
	case "agent_message":
		s.t.activity(phaseWriting, "", item.Text, 0)
	default:
		name, input, ok := s.toolCall(item)
		if !ok || s.started[item.ID] {
			return
		}
		if s.started == nil {
			s.started = make(map[string]bool)
		}
		s.started[item.ID] = true
		s.t.startTool(item.ID, name, input)
		s.t.activity(phaseTool, name, "", 0)
	}
}

func (s *codexStream) itemCompleted(item codexItem) {
	if item.asksUser() {
		askQuestions(s.t, item.Arguments)
		return
	}
	switch item.Type {
	case "reasoning":
		s.t.thinking(item.Text)
	case "agent_message":
		s.t.text(item.Text)
	default:
		if _, _, ok := s.toolCall(item); !ok {
			return
		}
		s.itemStarted(item)
		delete(s.started, item.ID)
		s.t.finishTool(item.ID, s.toolOutput(item), codexFailed(item))
		s.t.activity(phaseThinking, "", "", 0)
	}
}

func (s *codexStream) toolCall(item codexItem) (name, input string, ok bool) {
	switch item.Type {
	case "command_execution":
		return "Bash", s.relative(item.Command), true
	case "file_change":
		lines := make([]string, 0, len(item.Changes))
		for _, change := range item.Changes {
			lines = append(lines, change.Kind+" "+s.relative(change.Path))
		}
		return "Edit", strings.Join(lines, "\n"), true
	case "mcp_tool_call":
		return item.Server + "." + item.Tool, s.relative(toolInput(item.Arguments)), true
	case "web_search":
		return "WebSearch", item.Query, true
	case "todo_list":
		return "TodoWrite", todoLines(item.Items), true
	}
	return "", "", false
}

func (s *codexStream) toolOutput(item codexItem) string {
	switch {
	case item.Error != nil && item.Error.Message != "":
		return item.Error.Message
	case item.Type == "command_execution":
		return item.AggregatedOutput
	case item.Type == "mcp_tool_call" && item.Result != nil:
		return resultText(item.Result.Content)
	case item.Type == "todo_list":
		return todoLines(item.Items)
	}
	return ""
}

func (s *codexStream) relative(text string) string {
	return strings.ReplaceAll(text, s.projectDir+"/", "")
}

func (item codexItem) asksUser() bool {
	return item.Type == "mcp_tool_call" && item.Server == askuser.ServerName && item.Tool == askuser.ToolName
}

func codexFailed(item codexItem) bool {
	if item.Status == "failed" || item.Status == "declined" || item.Error != nil {
		return true
	}
	return item.Type == "command_execution" && item.ExitCode != nil && *item.ExitCode != 0
}

func todoLines(items []codexTodo) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		mark := "[ ] "
		if item.Completed {
			mark = "[x] "
		}
		lines = append(lines, mark+item.Text)
	}
	return strings.Join(lines, "\n")
}

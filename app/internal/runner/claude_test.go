package runner

import (
	"strings"
	"testing"

	"github.com/e1berd/drift/app/internal/workspace"
)

type recorded struct {
	kind, a, b string
}

type fakeSink struct{ got []recorded }

func (f *fakeSink) text(text string)       { f.got = append(f.got, recorded{"text", text, ""}) }
func (f *fakeSink) thinking(text string)   { f.got = append(f.got, recorded{"thinking", text, ""}) }
func (f *fakeSink) note(code, text string) { f.got = append(f.got, recorded{"note", code, text}) }
func (f *fakeSink) startTool(callID, name, input string) {
	f.got = append(f.got, recorded{"tool", callID, name + " " + input})
}
func (f *fakeSink) finishTool(callID, output string, failed bool) {
	status := "done"
	if failed {
		status = "error"
	}
	f.got = append(f.got, recorded{"result", callID, status + " " + output})
}
func (f *fakeSink) closeTools(status string) { f.got = append(f.got, recorded{"close", status, ""}) }

const sampleRun = `{"type":"system","subtype":"init","session_id":"abc"}
{"type":"rate_limit_event"}
{"type":"assistant","message":{"content":[{"type":"thinking","thinking":"","signature":"x"}]},"parent_tool_use_id":null}
{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"/w/p/a.txt"}}]},"parent_tool_use_id":null}
{"type":"assistant","message":{"content":[{"type":"text","text":"inner"}]},"parent_tool_use_id":"t9"}
{"type":"user","message":{"content":[{"tool_use_id":"t1","type":"tool_result","content":"1\thello"}]},"parent_tool_use_id":null}
{"type":"user","message":{"content":[{"tool_use_id":"t2","type":"tool_result","content":[{"type":"text","text":"boom"}],"is_error":true}]},"parent_tool_use_id":null}
{"type":"assistant","message":{"content":[{"type":"text","text":"It says hello"}]},"parent_tool_use_id":null}
not json
{"type":"result","subtype":"success","is_error":false,"result":"It says hello"}`

func TestClaudeStreamTurnsEventsIntoTranscript(t *testing.T) {
	sink := &fakeSink{}
	var saved string
	stream := claudeStream{t: sink, projectDir: "/w/p", saveSessionID: func(id string) { saved = id }}
	for line := range strings.SplitSeq(sampleRun, "\n") {
		stream.handle([]byte(line))
	}
	want := []recorded{
		{"thinking", "", ""},
		{"tool", "t1", "Read a.txt"},
		{"result", "t1", "done 1\thello"},
		{"result", "t2", "error boom"},
		{"text", "It says hello", ""},
	}
	if len(sink.got) != len(want) {
		t.Fatalf("got %v, want %v", sink.got, want)
	}
	for i := range want {
		if sink.got[i] != want[i] {
			t.Errorf("event %d: got %v, want %v", i, sink.got[i], want[i])
		}
	}
	if saved != "abc" || !stream.finished {
		t.Errorf("saved %q, finished %v", saved, stream.finished)
	}
}

func TestClaudeStreamReportsFailedResult(t *testing.T) {
	sink := &fakeSink{}
	stream := claudeStream{t: sink, agentSessionID: "abc", saveSessionID: func(string) { t.Error("session id saved again") }}
	stream.handle([]byte(`{"type":"system","subtype":"init","session_id":"abc"}`))
	stream.handle([]byte(`{"type":"result","subtype":"error_during_execution","is_error":true,"result":""}`))
	want := []recorded{{"close", "error", ""}, {"note", "failed", "error_during_execution"}}
	if len(sink.got) != 2 || sink.got[0] != want[0] || sink.got[1] != want[1] {
		t.Fatalf("got %v", sink.got)
	}
}

func TestClaudeArgsResumeAndModes(t *testing.T) {
	args := strings.Join(claudeArgs(runRequest("opus", "acceptEdits"), "abc"), " ")
	if args != "-p --output-format stream-json --verbose --model opus --permission-mode acceptEdits --resume abc" {
		t.Errorf("args %q", args)
	}
	args = strings.Join(claudeArgs(runRequest("default", "default"), ""), " ")
	if args != "-p --output-format stream-json --verbose" {
		t.Errorf("args %q", args)
	}
}

func runRequest(model, mode string) workspace.RunRequest {
	return workspace.RunRequest{Model: model, Mode: mode}
}

package runner

import (
	"strings"
	"testing"
)

const kimiRun = `{"role":"meta","type":"system.version","version":"2.6.3"}
{"role":"meta","type":"session.resume_hint","session_id":"k1","command":"kimi -r k1","content":"To resume this session: kimi -r k1"}
{"role":"assistant","content":"Looking at the file.","tool_calls":[{"type":"function","id":"call_1","function":{"name":"Read","arguments":"{\"file_path\":\"/w/p/a.txt\"}"}}]}
{"role":"tool","tool_call_id":"call_1","content":"1\thello"}
{"role":"assistant","content":"It says hello"}
not json
{"type":"error","message":"boom"}`

func TestKimiStreamTurnsMessagesIntoTranscript(t *testing.T) {
	sink := &fakeSink{}
	var saved string
	stream := kimiStream{t: sink, projectDir: "/w/p", saveSessionID: func(id string) { saved = id }}
	lines := strings.SplitSeq(kimiRun, "\n")
	for line := range lines {
		stream.handle([]byte(line))
	}
	want := []recorded{
		{"text", "Looking at the file.", ""},
		{"tool", "call_1", "Read a.txt"},
		{"result", "call_1", "done 1\thello"},
		{"text", "It says hello", ""},
		{"close", "error", ""},
		{"note", "failed", "boom"},
	}
	if len(sink.got) != len(want) {
		t.Fatalf("got %v", sink.got)
	}
	for i := range want {
		if sink.got[i] != want[i] {
			t.Errorf("event %d: got %v, want %v", i, sink.got[i], want[i])
		}
	}
	if saved != "k1" || !stream.done() {
		t.Errorf("saved %q, done %v", saved, stream.done())
	}
}

func TestKimiFailureShowsCauseBeforeLogPath(t *testing.T) {
	const cause = `error: failed to run prompt: Model "old/model" is not configured in config.toml.`
	const logPath = "See log: /tmp/kimi/logs/kimi-code.log"
	for _, tt := range []struct {
		output string
		want   string
	}{
		{cause + "\n\n" + logPath + "\n", cause},
		{logPath + "\n", logPath},
		{"ordinary stderr\n", "ordinary stderr"},
		{" \n", "exit status 1"},
	} {
		if got := failureMessage(tt.output, "exit status 1"); got != tt.want {
			t.Errorf("failureMessage(%q) = %q, want %q", tt.output, got, tt.want)
		}
	}
}

func TestKimiArgs(t *testing.T) {
	req := runRequest("kimi-code/k3.1-pro", "acceptEdits")
	req.Text = "hi"
	args := kimiArgs(req, "k1")
	got := strings.Join(args, " ")
	want := "-p hi --output-format stream-json -m kimi-code/k3.1-pro -S k1"
	if got != want {
		t.Errorf("args %q, want %q", got, want)
	}
	req = runRequest("default", "plan")
	req.Text = "hi"
	args = kimiArgs(req, "")
	if strings.Join(args, " ") != "-p hi --output-format stream-json" {
		t.Errorf("args %q", args)
	}
	if args = kimiArgs(runRequest("default", "default"), ""); len(args) != 4 {
		t.Errorf("default model must pass no flag: %q", args)
	}
}

func TestKimiProviderAliasUsesConfiguredDefault(t *testing.T) {
	req := runRequest("kimi-code", "auto")
	req.Text = "hi"
	if got := strings.Join(kimiArgs(req, ""), " "); got != "-p hi --output-format stream-json" {
		t.Errorf("new session args %q", got)
	}
	if got := strings.Join(kimiArgs(req, "k1"), " "); got != "-p hi --output-format stream-json -S k1" {
		t.Errorf("resumed session args %q", got)
	}
}

const mimoRun = `{"type":"step_start","timestamp":1,"sessionID":"m1","part":{"id":"p0","type":"step-start"}}
{"type":"text","timestamp":2,"sessionID":"m1","part":{"id":"p1","type":"text","text":"Checking it."}}
{"type":"tool_use","timestamp":3,"sessionID":"m1","part":{"id":"p2","type":"tool","tool":"bash","state":{"status":"completed","input":{"command":"cat /w/p/a.txt"},"output":"hello"}}}
{"type":"text","timestamp":4,"sessionID":"m1","part":{"id":"p3","type":"text","text":"It says hello"}}
{"type":"step_finish","timestamp":5,"sessionID":"m1","part":{"id":"p4","type":"step-finish"}}
{"type":"tool_use","timestamp":6,"sessionID":"m1","part":{"id":"p5","type":"tool","tool":"bash","state":{"status":"error","input":{"command":"false"},"error":"exit 1"}}}
{"type":"error","timestamp":7,"sessionID":"m1","error":{"name":"MessageOutputLengthError","data":{"message":"too long"}}}`

func TestMimoStreamTurnsEventsIntoTranscript(t *testing.T) {
	sink := &fakeSink{}
	var saved string
	stream := mimoStream{t: sink, projectDir: "/w/p", saveSessionID: func(id string) { saved = id }}
	for line := range strings.SplitSeq(mimoRun, "\n") {
		stream.handle([]byte(line))
	}
	want := []recorded{
		{"tool", "p2", "bash cat a.txt"},
		{"result", "p2", "done hello"},
		{"text", "Checking it.\n\nIt says hello", ""},
		{"tool", "p5", "bash false"},
		{"result", "p5", "error exit 1"},
		{"close", "error", ""},
		{"note", "failed", "too long"},
	}
	if len(sink.got) != len(want) {
		t.Fatalf("got %v", sink.got)
	}
	for i := range want {
		if sink.got[i] != want[i] {
			t.Errorf("event %d: got %v, want %v", i, sink.got[i], want[i])
		}
	}
	if saved != "m1" || !stream.done() {
		t.Errorf("saved %q, done %v", saved, stream.done())
	}
}

func TestMimoArgs(t *testing.T) {
	req := runRequest("default", "bypassPermissions")
	req.Text = "hi"
	args := mimoArgs(req, "m1", []string{"/tmp/img.png"})
	got := strings.Join(args, " ")
	want := "run --format json --dangerously-skip-permissions --agent build -s m1 -f /tmp/img.png"
	if got != want {
		t.Errorf("args %q, want %q", got, want)
	}
	if args = mimoArgs(runRequest("", "default"), "", nil); strings.Join(args, " ") != "run --format json --dangerously-skip-permissions --agent build" {
		t.Errorf("args %q", args)
	}
	if args = mimoArgs(runRequest("", "plan"), "", nil); strings.Join(args, " ") != "run --format json --dangerously-skip-permissions --agent plan" {
		t.Errorf("plan mode must run the plan agent: %q", args)
	}
}

package runner

import (
	"strings"
	"testing"
)

const antigravityRun = `{"type":"init","init":{"cwd":"/w/p","tools":["view_file"],"permission_mode":"request-review"}}
{"type":"step_update","step_update":{"conversation_id":"c1","step_index":0,"state":"DONE","step_type":"user_input"}}
{"type":"step_update","step_update":{"conversation_id":"c1","step_index":1,"state":"ACTIVE","step_type":"agent_response","usage":{"thinking_tokens":120}}}
{"type":"step_update","step_update":{"conversation_id":"c1","step_index":1,"state":"DONE","step_type":"agent_response"}}
{"type":"step_update","step_update":{"conversation_id":"c1","step_index":2,"state":"ACTIVE","step_type":"tool","tool_name":"view_file","tool_info":{"name":"view_file","parameters":{"file_path":"/w/p/a.txt"}}}}
{"type":"step_update","step_update":{"conversation_id":"c1","step_index":2,"state":"DONE","step_type":"tool","tool_name":"view_file","tool_info":{"name":"view_file","parameters":{"file_path":"/w/p/a.txt"},"output":"hello"}}}
{"type":"step_update","step_update":{"conversation_id":"c1","step_index":3,"state":"DONE","step_type":"tool","tool_name":"run_command","tool_info":{"name":"run_command","parameters":{"command":"false"},"output":"exit 1","status":"error"}}}
{"type":"step_update","step_update":{"conversation_id":"c1","step_index":4,"state":"ACTIVE","step_type":"agent_response","subagent_info":{"name":"browser"},"text_delta":"inner"}}
{"type":"step_update","step_update":{"conversation_id":"c1","step_index":5,"state":"ACTIVE","step_type":"agent_response","text_delta":"It says "}}
{"type":"step_update","step_update":{"conversation_id":"c1","step_index":5,"state":"ACTIVE","step_type":"agent_response","text_delta":"hello"}}
not json
{"type":"result","result":{"conversation_id":"c1","status":"SUCCESS","response":"It says hello","num_turns":1}}`

func TestAntigravityStreamTurnsStepsIntoTranscript(t *testing.T) {
	sink := &fakeSink{}
	var saved []string
	stream := antigravityStream{t: sink, projectDir: "/w/p", saveSessionID: func(id string) { saved = append(saved, id) }}
	for line := range strings.SplitSeq(antigravityRun, "\n") {
		stream.handle([]byte(line))
	}
	want := []recorded{
		{"tool", "step-2", "view_file a.txt"},
		{"result", "step-2", "done hello"},
		{"tool", "step-3", "run_command false"},
		{"result", "step-3", "error exit 1"},
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
	if len(saved) != 1 || saved[0] != "c1" || !stream.done() {
		t.Errorf("saved %v, done %v", saved, stream.done())
	}
	if last := sink.activities[len(sink.activities)-1]; last != (recorded{"writing", "", "It says hello"}) {
		t.Errorf("last activity %v", last)
	}
}

func TestAntigravityStreamReadsFlatEventsAndFailures(t *testing.T) {
	sink := &fakeSink{}
	stream := antigravityStream{t: sink, conversationID: "c1", saveSessionID: func(string) { t.Error("conversation id saved again") }}
	stream.handle([]byte(`{"event":"result","conversation_id":"c1","status":"ERROR","error":"authentication required"}`))
	want := []recorded{{"close", "error", ""}, {"note", "failed", "authentication required"}}
	if len(sink.got) != 2 || sink.got[0] != want[0] || sink.got[1] != want[1] {
		t.Fatalf("got %v", sink.got)
	}
}

func TestAntigravityStreamLeavesStoppedRunsToTheRunner(t *testing.T) {
	sink := &fakeSink{}
	stream := antigravityStream{t: sink, saveSessionID: func(string) {}}
	stream.handle([]byte(`{"type":"result","status":"CANCELED","finished_reason":"interrupted"}`))
	if len(sink.got) != 0 || !stream.done() {
		t.Fatalf("got %v", sink.got)
	}
}

func TestAntigravityArgs(t *testing.T) {
	args := antigravityArgs(runRequest("gemini-4-argon", "acceptEdits"), "c1")
	got := strings.Join(args[:len(args)-1], " ")
	if got != "--output-format stream-json --model gemini-4-argon --mode accept-edits --conversation c1" {
		t.Errorf("args %q", got)
	}
	req := runRequest("default", "bypassPermissions")
	req.Text = "-h\n/usage"
	args = antigravityArgs(req, "")
	if strings.Join(args, " ") != "--output-format stream-json --dangerously-skip-permissions --prompt=-h\n/usage" {
		t.Errorf("args %q", args)
	}
	if args = antigravityArgs(runRequest("", "default"), ""); len(args) != 3 {
		t.Errorf("default mode must pass no flag: %q", args)
	}
}

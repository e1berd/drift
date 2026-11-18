package runner

import (
	"strings"
	"testing"
	"time"

	"github.com/e1berd/drift/app/internal/agents"
)

func replay(stream agentStream, lines string) {
	for line := range strings.SplitSeq(lines, "\n") {
		stream.handle([]byte(line))
	}
}

func expectRecorded(t *testing.T, got, want []recorded) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("event %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

const claudeWarningRun = `{"type":"rate_limit_event","rate_limit_info":{"status":"allowed_warning","resetsAt":1763470800,"rateLimitType":"five_hour","utilization":0.8,"unifiedWindows":{"seven_day":{"utilization":0.42,"resetsAt":1763900000},"five_hour":{"utilization":0.8,"resetsAt":1763470800}}}}
{"type":"rate_limit_event","rate_limit_info":{"status":"allowed","resetsAt":1763900000,"rateLimitType":"seven_day_opus","utilization":0.91}}
{"type":"result","subtype":"success","is_error":false,"result":"ok"}`

func TestClaudeStreamReportsUsageWindows(t *testing.T) {
	sink := &fakeSink{}
	replay(&claudeStream{t: sink, saveSessionID: func(string) {}}, claudeWarningRun)
	expectRecorded(t, sink.got, []recorded{
		{"usage", "five_hour", "80"},
		{"usage", "seven_day", "42"},
		{"usage", "seven_day_opus", "91"},
	})
}

const claudeExhaustedRun = `{"type":"rate_limit_event","rate_limit_info":{"status":"rejected","resetsAt":1763470800,"rateLimitType":"five_hour","utilization":1}}
{"type":"assistant","message":{"content":[{"type":"text","text":"You've hit your session limit · resets 3pm"}]},"parent_tool_use_id":null,"error":"rate_limit"}
{"type":"result","subtype":"success","is_error":true,"result":"You've hit your session limit · resets 3pm"}`

func TestClaudeStreamTurnsRejectionIntoLimit(t *testing.T) {
	sink := &fakeSink{}
	replay(&claudeStream{t: sink, saveSessionID: func(string) {}}, claudeExhaustedRun)
	expectRecorded(t, sink.got, []recorded{
		{"usage", "five_hour", "100"},
		{"close", "error", ""},
		{"limit", "five_hour 1763470800", "You've hit your session limit · resets 3pm"},
	})
}

func TestClaudeStreamReadsResetFromLegacyLimitText(t *testing.T) {
	sink := &fakeSink{}
	replay(&claudeStream{t: sink, saveSessionID: func(string) {}}, `{"type":"result","subtype":"success","is_error":true,"result":"Claude AI usage limit reached|1763470800"}`)
	expectRecorded(t, sink.got, []recorded{
		{"close", "error", ""},
		{"limit", " 1763470800", "Claude AI usage limit reached|1763470800"},
	})
}

func TestClaudeStreamTurnsAskToolIntoQuestion(t *testing.T) {
	sink := &fakeSink{}
	replay(&claudeStream{t: sink, saveSessionID: func(string) {}}, `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t1","name":"mcp__drift__ask_user","input":{"questions":[{"question":"Which DB?","options":[{"label":"Postgres"},{"label":"SQLite"}]}]}}]},"parent_tool_use_id":null}
{"type":"user","message":{"content":[{"tool_use_id":"t1","type":"tool_result","content":"shown"}]},"parent_tool_use_id":null}`)
	expectRecorded(t, sink.got, []recorded{{"question", "Which DB?", "1"}, {"result", "t1", "done shown"}})
}

func TestCodexStreamTurnsUsageLimitIntoLimit(t *testing.T) {
	sink := &fakeSink{}
	message := "You’ve hit your usage limit. Upgrade to Pro or try again at 3:45 PM."
	replay(&codexStream{t: sink, saveSessionID: func(string) {}}, `{"type":"turn.failed","error":{"message":"`+message+`"}}`)
	expectRecorded(t, sink.got, []recorded{{"close", "error", ""}, {"limit", " ", message}})
}

func TestCodexStreamTurnsAskToolIntoQuestion(t *testing.T) {
	sink := &fakeSink{}
	call := `"type":"mcp_tool_call","server":"drift","tool":"ask_user","arguments":{"questions":[{"question":"Ship it?","options":[{"label":"Yes"},{"label":"No"}]}]}`
	replay(&codexStream{t: sink, saveSessionID: func(string) {}}, `{"type":"item.started","item":{"id":"i1",`+call+`,"status":"in_progress"}}
{"type":"item.completed","item":{"id":"i1",`+call+`,"status":"completed"}}`)
	expectRecorded(t, sink.got, []recorded{{"question", "Ship it?", "1"}})
}

func TestNewWarningsFireOncePerThreshold(t *testing.T) {
	r := &Runner{warned: make(map[string]float64)}
	window := func(used float64) []agents.Window { return []agents.Window{{ID: "five_hour", UsedPercent: used}} }
	steps := []struct {
		used float64
		warn bool
	}{{50, false}, {76, true}, {80, false}, {91, true}, {96, true}, {99, false}, {10, false}, {77, true}}
	for _, step := range steps {
		if got := len(r.newWarnings("sub", window(step.used))) == 1; got != step.warn {
			t.Errorf("at %v%% warned %v, want %v", step.used, got, step.warn)
		}
	}
}

func TestExhaustedWindowTakesResetFromReadUsage(t *testing.T) {
	reset := time.Unix(1763470800, 0)
	read := []agents.Window{
		{ID: "primary", UsedPercent: 100, WindowMinutes: 300, ResetTime: &reset},
		{ID: "secondary", UsedPercent: 40, WindowMinutes: weekMinutes, ResetTime: &time.Time{}},
	}
	got := exhaustedWindow(agents.Window{}, read)
	if got.ID != "primary" || got.ResetTime != &reset {
		t.Errorf("window %+v", got)
	}
	if percentLeft(91.5) != 8 || percentLeft(130) != 0 {
		t.Error("percent left")
	}
}

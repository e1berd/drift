package runner

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/e1berd/drift/app/internal/workspace"
	"github.com/jackc/pgx/v5/pgtype"
)

const sampleCodexRun = `{"type":"thread.started","thread_id":"th-1"}
{"type":"turn.started"}
{"type":"item.completed","item":{"id":"item_0","type":"reasoning","text":"Looking at the file"}}
{"type":"item.started","item":{"id":"item_1","type":"command_execution","command":"bash -lc 'cat /w/p/a.txt'","aggregated_output":"","exit_code":null,"status":"in_progress"}}
{"type":"item.completed","item":{"id":"item_1","type":"command_execution","command":"bash -lc 'cat /w/p/a.txt'","aggregated_output":"hello\n","exit_code":0,"status":"completed"}}
{"type":"item.completed","item":{"id":"item_2","type":"file_change","changes":[{"path":"/w/p/b.txt","kind":"add"}],"status":"completed"}}
{"type":"item.started","item":{"id":"item_3","type":"command_execution","command":"false","aggregated_output":"","exit_code":null,"status":"in_progress"}}
{"type":"item.completed","item":{"id":"item_3","type":"command_execution","command":"false","aggregated_output":"","exit_code":1,"status":"failed"}}
{"type":"item.completed","item":{"id":"item_4","type":"agent_message","text":"It says hello"}}
not json
{"type":"turn.completed","usage":{"input_tokens":10,"cached_input_tokens":0,"output_tokens":3}}`

func TestCodexStreamTurnsEventsIntoTranscript(t *testing.T) {
	sink := &fakeSink{}
	var saved string
	stream := codexStream{t: sink, projectDir: "/w/p", saveSessionID: func(id string) { saved = id }}
	for line := range strings.SplitSeq(sampleCodexRun, "\n") {
		stream.handle([]byte(line))
	}
	want := []recorded{
		{"thinking", "Looking at the file", ""},
		{"tool", "item_1", "Bash bash -lc 'cat a.txt'"},
		{"result", "item_1", "done hello\n"},
		{"tool", "item_2", "Edit add b.txt"},
		{"result", "item_2", "done "},
		{"tool", "item_3", "Bash false"},
		{"result", "item_3", "error "},
		{"text", "It says hello", ""},
	}
	if len(sink.got) != len(want) {
		t.Fatalf("got %v", sink.got)
	}
	for i := range want {
		if sink.got[i] != want[i] {
			t.Errorf("event %d: got %v, want %v", i, sink.got[i], want[i])
		}
	}
	if saved != "th-1" || !stream.done() {
		t.Errorf("saved %q, done %v", saved, stream.done())
	}
}

func TestCodexStreamReportsFailedTurn(t *testing.T) {
	sink := &fakeSink{}
	stream := codexStream{t: sink, saveSessionID: func(string) {}}
	stream.handle([]byte(`{"type":"error","message":"stream disconnected"}`))
	stream.handle([]byte(`{"type":"turn.failed","error":{"message":"stream disconnected"}}`))
	want := []recorded{{"close", "error", ""}, {"note", "failed", "stream disconnected"}}
	if len(sink.got) != 2 || sink.got[0] != want[0] || sink.got[1] != want[1] || !stream.done() {
		t.Fatalf("got %v", sink.got)
	}
}

func TestCodexArgs(t *testing.T) {
	args := strings.Join(codexArgs(runRequest("gpt-5.5", "acceptEdits"), "", []string{"/tmp/a.png"}, testMCP), " ")
	want := `exec --json --skip-git-repo-check --model gpt-5.5 -c sandbox_mode="workspace-write" -c approval_policy="never" ` + codexMCP + ` --image=/tmp/a.png -- -`
	if args != want {
		t.Errorf("args %q", args)
	}
	args = strings.Join(codexArgs(runRequest("default", "bypassPermissions"), "th-1", nil, testMCP), " ")
	want = "exec resume --json --skip-git-repo-check --dangerously-bypass-approvals-and-sandbox " + codexMCP + " -- th-1 -"
	if args != want {
		t.Errorf("args %q", args)
	}
	args = strings.Join(codexArgs(runRequest("", "plan"), "", nil, testMCP), " ")
	if !strings.Contains(args, `sandbox_mode="read-only"`) {
		t.Errorf("args %q", args)
	}
}

func TestClaudeInputCarriesImages(t *testing.T) {
	if got := claudeInput("hi", nil); got != "hi" {
		t.Fatalf("plain text input %q", got)
	}
	line := claudeInput("what is this", []attachedImage{{mediaType: "image/png", data: []byte{1, 2}}})
	var message claudeUserMessage
	if err := json.Unmarshal([]byte(line), &message); err != nil || !strings.HasSuffix(line, "\n") {
		t.Fatalf("input %q: %v", line, err)
	}
	content := message.Message.Content
	if message.Type != "user" || len(content) != 2 || content[0].Text != "what is this" {
		t.Fatalf("message %+v", message)
	}
	if content[1].Source.MediaType != "image/png" || content[1].Source.Data != base64.StdEncoding.EncodeToString([]byte{1, 2}) {
		t.Errorf("image block %+v", content[1].Source)
	}
	args := strings.Join(claudeArgs(runRequest("", ""), "", true, testMCP), " ")
	if !strings.Contains(args, "--input-format stream-json") {
		t.Errorf("args %q", args)
	}
}

func TestMergeKeepsAttachmentsOfQueuedMessages(t *testing.T) {
	a, b := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	got := merge([]workspace.RunRequest{
		{Text: "first", Attachments: []pgtype.UUID{a}},
		{Attachments: []pgtype.UUID{b}},
		{Text: "last", Mode: "plan"},
	})
	if got.Text != "first\n\nlast" || len(got.Attachments) != 2 || got.Attachments[1] != b || got.Mode != "plan" {
		t.Errorf("merged %+v", got)
	}
}

func TestAttachmentsSplitIntoInlineImagesAndPaths(t *testing.T) {
	files := []attachedFile{
		{name: "shot.png", mediaType: "image/png", size: 10, path: "/files/a/shot.png"},
		{name: "huge.png", mediaType: "image/png", size: maxInlineImageBytes + 1, path: "/files/b/huge.png"},
		{name: "logs.zip", mediaType: "application/zip", size: 10, path: "/files/c/logs.zip"},
	}
	images, others := splitInlineImages(files)
	if len(images) != 1 || images[0].name != "shot.png" || len(others) != 2 {
		t.Fatalf("images %v, others %v", images, others)
	}
	if got := withFilePaths("look", others); got != "look\n\nAttached files:\n/files/b/huge.png\n/files/c/logs.zip" {
		t.Errorf("text %q", got)
	}
	if got := withFilePaths("look", nil); got != "look" {
		t.Errorf("text without files %q", got)
	}
}

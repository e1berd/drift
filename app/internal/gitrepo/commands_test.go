package gitrepo

import (
	"reflect"
	"testing"
)

func TestCreatedRefsFindsBranchesAndWorktreesInShellCommands(t *testing.T) {
	calls := []ToolCall{
		{MessageID: "m1", Name: "Bash", Input: "git switch -c feat/status && git add . && git commit -m 'x && y'"},
		{MessageID: "m2", Name: "Bash", Input: "cd app; GIT_TRACE=0 git -C .. checkout -b fix/one"},
		{MessageID: "m3", Name: "Bash", Input: "git branch -D old && git branch --list && git branch topic main"},
		{MessageID: "m4", Name: "Bash", Input: `git worktree add -b wt/feature "../drift wt" main`},
		{MessageID: "m5", Name: "Bash", Input: "git worktree add ../hotfix"},
		{MessageID: "m6", Name: "Bash", Input: "git worktree add --detach ../scratch HEAD~1"},
		{MessageID: "m7", Name: "EnterWorktree", Input: `{"name":"agent-tree"}`},
		{MessageID: "m8", Name: "Read", Input: "git checkout -b ignored"},
		{MessageID: "m9", Name: "Bash", Input: "echo git checkout -b not-git"},
	}
	want := []CreatedRef{
		{MessageID: "m1", Branch: "feat/status"},
		{MessageID: "m2", Branch: "fix/one", Dir: "."},
		{MessageID: "m3", Branch: "topic"},
		{MessageID: "m4", Branch: "wt/feature", Worktree: "../drift wt"},
		{MessageID: "m5", Branch: "hotfix", Worktree: "../hotfix"},
		{MessageID: "m6", Worktree: "../scratch"},
		{MessageID: "m7", Worktree: "agent-tree"},
	}
	if got := CreatedRefs(calls); !reflect.DeepEqual(got, want) {
		t.Fatalf("CreatedRefs:\n got %+v\nwant %+v", got, want)
	}
}

func TestParseNumstatSkipsMalformedRecords(t *testing.T) {
	out := []byte("3\t1\ta.go\x00-\t-\timage.png\x00junk\x00")
	want := []FileStat{{Path: "a.go", Additions: 3, Deletions: 1}, {Path: "image.png"}}
	if got := parseNumstat(out); !reflect.DeepEqual(got, want) {
		t.Fatalf("parseNumstat = %+v", got)
	}
}

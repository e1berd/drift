package gitrepo

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func (h *harness) git(args ...string) string {
	h.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = h.folder
	out, err := cmd.CombinedOutput()
	if err != nil {
		h.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestSummaryCountsChangesSinceBaseAndListsSessionRefs(t *testing.T) {
	h := newHarness(t)
	h.git("init", "-q")
	h.write("a.txt", "one\ntwo\n")
	h.git("add", ".")
	h.git("commit", "-qm", "base")
	base := h.git("rev-parse", "HEAD")

	h.git("switch", "-qc", "feat/status")
	h.write("a.txt", "one\nTWO\nthree\n")
	h.git("commit", "-qam", "edit")
	h.write("new.txt", "x\ny")
	h.git("worktree", "add", "-q", "../tree", "-b", "wt")
	h.git("branch", "unrelated")

	calls := []ToolCall{
		{MessageID: "m1", Name: "Bash", Input: "git switch -c feat/status"},
		{MessageID: "m2", Name: "Bash", Input: "git worktree add ../tree -b wt"},
		{MessageID: "m3", Name: "Bash", Input: "git checkout -b deleted-later"},
	}
	r, err := open(h.folder)
	if err != nil {
		t.Fatal(err)
	}
	s, err := r.summary(context.Background(), base, calls)
	if err != nil {
		t.Fatal(err)
	}
	if s.Changes.Additions != 4 || s.Changes.Deletions != 1 || s.Changes.FileCount != 2 {
		t.Fatalf("changes = %+v", s.Changes)
	}
	if len(s.Branches) != 2 || s.Branches[0].Name != "feat/status" || !s.Branches[0].Current ||
		s.Branches[0].Commits != 1 || s.Branches[1].Name != "wt" || s.Branches[1].MessageID != "m2" {
		t.Fatalf("branches = %+v", s.Branches)
	}
	if len(s.Worktrees) != 1 || s.Worktrees[0].Path != "tree" || s.Worktrees[0].Branch != "wt" {
		t.Fatalf("worktrees = %+v (folder %s)", s.Worktrees, filepath.Dir(h.folder))
	}
}

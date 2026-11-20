package gitrepo

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
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

func (h *harness) repo() repo {
	h.t.Helper()
	r, err := open(h.folder)
	if err != nil {
		h.t.Fatal(err)
	}
	return r
}

func activityKeys(entries []ActivityEntry) []string {
	keys := make([]string, 0, len(entries))
	for _, entry := range entries {
		state := "present"
		if entry.Removed {
			state = "removed"
		}
		keys = append(keys, entry.Key+"="+state)
	}
	return keys
}

func sessionWithBranchWorktreeAndFile(t *testing.T) (*harness, History) {
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
	return h, History{Base: base, Calls: []ToolCall{
		{MessageID: "m1", Name: "Bash", Input: "git switch -c feat/status"},
		{MessageID: "m2", Name: "Bash", Input: "git worktree add ../tree -b wt"},
		{MessageID: "m3", Name: "Write", Input: "new.txt", Created: true},
		{MessageID: "m4", Name: "Write", Input: "a.txt"},
		{MessageID: "m5", Name: "Bash", Input: "git checkout -b deleted-later"},
	}}
}

func TestSummaryCountsChangesAndListsActivityInOrder(t *testing.T) {
	h, history := sessionWithBranchWorktreeAndFile(t)
	s, err := h.repo().summary(context.Background(), history)
	if err != nil {
		t.Fatal(err)
	}
	if s.Changes.Additions != 4 || s.Changes.Deletions != 1 || s.Changes.FileCount != 2 {
		t.Fatalf("changes = %+v", s.Changes)
	}
	want := "branch:feat/status=present worktree:tree=present branch:wt=present file:new.txt=present branch:deleted-later=removed"
	if got := strings.Join(activityKeys(s.Activity), " "); got != want {
		t.Fatalf("activity:\n got %s\nwant %s", got, want)
	}
	if first := s.Activity[0]; !first.Current || first.Commits != 1 {
		t.Fatalf("branch entry = %+v", first)
	}
}

func TestRemovingAWorktreeNeedsForceForChangesAndCanDropItsBranch(t *testing.T) {
	h, history := sessionWithBranchWorktreeAndFile(t)
	if err := os.WriteFile(filepath.Join(filepath.Dir(h.folder), "tree", "dirty.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := h.repo()
	ctx := context.Background()
	var httpErr *echo.HTTPError
	if _, err := r.removeWorktree(ctx, history, "tree", false, true); !errors.As(err, &httpErr) || httpErr.Code != http.StatusConflict {
		t.Fatalf("removing a dirty worktree without force = %v", err)
	}
	removal, err := r.removeWorktree(ctx, history, "tree", true, true)
	if err != nil {
		t.Fatal(err)
	}
	if removal.Target != KindWorktree || removal.Branch != "wt" || removal.SourceID != "m2" {
		t.Fatalf("removal = %+v", removal)
	}
	history.Removals = append(history.Removals, removal)
	s, err := r.summary(ctx, history)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range s.Activity {
		if (entry.Kind == KindWorktree || entry.Name == "wt") && (!entry.Removed || entry.RemovedBy != RemovedByUser) {
			t.Fatalf("entry after removal = %+v", entry)
		}
	}
}

func TestRemovingAFileOnlyAcceptsFilesTheSessionCreated(t *testing.T) {
	h, history := sessionWithBranchWorktreeAndFile(t)
	r := h.repo()
	ctx := context.Background()
	var httpErr *echo.HTTPError
	if _, err := r.removeFile(ctx, history, "a.txt"); !errors.As(err, &httpErr) || httpErr.Code != http.StatusNotFound {
		t.Fatalf("removing a file the session did not create = %v", err)
	}
	if _, err := r.removeFile(ctx, history, "new.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(h.folder, "new.txt")); !os.IsNotExist(err) {
		t.Fatalf("new.txt still exists: %v", err)
	}
	if _, err := r.removeFile(ctx, history, "new.txt"); !errors.As(err, &httpErr) || httpErr.Code != http.StatusConflict {
		t.Fatalf("removing it twice = %v", err)
	}
}

func TestActivityListsCreatedFilesOutsideRepositories(t *testing.T) {
	h := newHarness(t)
	h.write("notes.md", "hi\n")
	s, err := h.repo().summary(context.Background(), History{Calls: []ToolCall{
		{MessageID: "m1", Name: "Write", Input: "notes.md", Created: true},
		{MessageID: "m2", Name: "Edit", Input: "add gone.txt\nupdate notes.md"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(activityKeys(s.Activity), " "); s.Repository || got != "file:notes.md=present file:gone.txt=removed" {
		t.Fatalf("repository=%v activity=%s", s.Repository, got)
	}
}

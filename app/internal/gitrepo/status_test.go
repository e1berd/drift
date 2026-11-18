package gitrepo

import (
	"strings"
	"testing"
)

func TestParseStatusSplitsStagedAndUnstagedSides(t *testing.T) {
	out := strings.Join([]string{
		"# branch.oid 0123456789abcdef0123456789abcdef01234567",
		"# branch.head main",
		"# branch.upstream origin/main",
		"# branch.ab +2 -1",
		"1 MM N... 100644 100644 100644 aaaa bbbb src/both changed.go",
		"1 .D N... 100644 100644 000000 aaaa aaaa gone.txt",
		"2 R. N... 100644 100644 100644 aaaa aaaa R100 new name.txt",
		"old name.txt",
		"u UU N... 100644 100644 100644 100644 aaaa bbbb cccc conflict.txt",
		"? notes/todo.md",
		"",
	}, "\x00")

	s := ParseStatus([]byte(out))

	if !s.Repository || s.Branch != "main" || s.Upstream != "origin/main" || s.Ahead != 2 || s.Behind != 1 || s.Head != "0123456789ab" {
		t.Fatalf("header = %+v", s)
	}
	wantStaged := []FileChange{
		{Path: "src/both changed.go", Status: Modified},
		{Path: "new name.txt", OrigPath: "old name.txt", Status: Renamed},
	}
	wantUnstaged := []FileChange{
		{Path: "src/both changed.go", Status: Modified},
		{Path: "gone.txt", Status: Deleted},
		{Path: "conflict.txt", Status: Conflicted},
		{Path: "notes/todo.md", Status: Untracked},
	}
	if !equal(s.Staged, wantStaged) {
		t.Errorf("staged = %+v, want %+v", s.Staged, wantStaged)
	}
	if !equal(s.Unstaged, wantUnstaged) {
		t.Errorf("unstaged = %+v, want %+v", s.Unstaged, wantUnstaged)
	}
}

func TestParseStatusOfAFreshRepository(t *testing.T) {
	s := ParseStatus([]byte("# branch.oid (initial)\x00# branch.head (detached)\x00"))
	if s.Head != "" || !s.Detached || s.Branch != "" || len(s.Staged)+len(s.Unstaged) != 0 {
		t.Fatalf("status = %+v", s)
	}
}

func TestCleanPathsRejectsEscapes(t *testing.T) {
	for _, bad := range []string{"", "../x", "a/../../x", "/etc/passwd", `a\b`, ".git/config", ".git"} {
		if _, err := CleanPaths([]string{bad}); err == nil {
			t.Errorf("CleanPaths(%q) accepted", bad)
		}
	}
	got, err := CleanPaths([]string{"a/./b.txt", "dir/x"})
	if err != nil || got[0] != "a/b.txt" || got[1] != "dir/x" {
		t.Fatalf("CleanPaths = %v, %v", got, err)
	}
}

func TestPageTokenRoundTrip(t *testing.T) {
	page := commitPage{head: strings.Repeat("ab", 20), skip: 30}
	got, ok := parsePageToken(page.token())
	if !ok || got != page {
		t.Fatalf("parsePageToken = %+v, %v", got, ok)
	}
	for _, bad := range []string{"%%", "bm90LWEtaGFzaDox", commitPage{head: page.head, skip: -1}.token()} {
		if _, ok := parsePageToken(bad); ok {
			t.Errorf("parsePageToken(%q) accepted", bad)
		}
	}
}

func equal(a, b []FileChange) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

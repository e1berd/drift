package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/e1berd/drift/app/internal/workspace"
)

func TestCreatesFileOnlyForWritesToMissingPaths(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "old.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	tr := &transcript{dir: dir}
	cases := map[string]bool{"new.txt": true, "old.txt": false, filepath.Join(dir, "abs.txt"): true}
	for input, want := range cases {
		if got := tr.createsFile("Write", input); got != want {
			t.Errorf("createsFile(Write, %q) = %v", input, got)
		}
	}
	if tr.createsFile("Edit", "new.txt") {
		t.Error("an edit never creates a file")
	}
}

func TestRemovalNoticeNamesEveryRemoval(t *testing.T) {
	text := removalNotice([]string{
		removalLine(workspace.RemovalNote{Target: "worktree", Path: "../tree", Branch: "wt"}),
		removalLine(workspace.RemovalNote{Target: "file", Path: "notes.md"}),
	})
	for _, want := range []string{"worktree ../tree and its branch wt", "- file notes.md", "do not recreate"} {
		if !strings.Contains(text, want) {
			t.Errorf("notice %q lacks %q", text, want)
		}
	}
}

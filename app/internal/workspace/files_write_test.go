package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateEntryMakesFileWithParents(t *testing.T) {
	root, dir := openTestRoot(t)
	entry, err := createEntry(root, "src/app/main.go", false)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Name != "main.go" || entry.Dir {
		t.Fatalf("unexpected entry %+v", entry)
	}
	if info, err := os.Stat(filepath.Join(dir, "src", "app", "main.go")); err != nil || info.Size() != 0 {
		t.Fatalf("file was not created empty: %v", err)
	}
}

func TestCreateEntryMakesFolder(t *testing.T) {
	root, dir := openTestRoot(t)
	if _, err := createEntry(root, "docs", true); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(filepath.Join(dir, "docs")); err != nil || !info.IsDir() {
		t.Fatalf("folder was not created: %v", err)
	}
}

func TestCreateEntryRefusesExisting(t *testing.T) {
	root, dir := openTestRoot(t)
	if err := os.WriteFile(filepath.Join(dir, "notes.md"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, asDir := range []bool{false, true} {
		if _, err := createEntry(root, "notes.md", asDir); !errors.Is(err, errEntryExists) {
			t.Fatalf("want conflict, got %v", err)
		}
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "notes.md")); string(got) != "keep" {
		t.Fatalf("existing file changed: %q", got)
	}
}

func TestCreateEntryStaysInsideRoot(t *testing.T) {
	root, dir := openTestRoot(t)
	if err := os.Symlink(filepath.Dir(dir), filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err := createEntry(root, "link/escape.txt", false); err == nil {
		t.Fatal("symlink escape must fail")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "escape.txt")); err == nil {
		t.Fatal("created a file outside the root")
	}
}

func TestReplaceFileKeepsModeAndLeavesNoTemp(t *testing.T) {
	root, dir := openTestRoot(t)
	target := filepath.Join(dir, "run.sh")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := replaceFile(root, "run.sh", []byte("new")); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(target)
	if err != nil || info.Mode().Perm() != 0o755 {
		t.Fatalf("mode changed: %v %v", info.Mode(), err)
	}
	if got, _ := os.ReadFile(target); string(got) != "new" {
		t.Fatalf("content not replaced: %q", got)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("temporary file left behind: %v", entries)
	}
}

func TestReplaceFileRequiresExistingFile(t *testing.T) {
	root, dir := openTestRoot(t)
	if err := os.Mkdir(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := replaceFile(root, "missing.txt", nil); err == nil {
		t.Fatal("missing file must not be created")
	}
	if err := replaceFile(root, "src", nil); err == nil {
		t.Fatal("folder must not be replaced")
	}
}

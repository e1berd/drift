package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func openTestRoot(t *testing.T) (*os.Root, string) {
	t.Helper()
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	return root, dir
}

func TestEnsureProjectFolderCreatesMissingFolders(t *testing.T) {
	root, dir := openTestRoot(t)
	if err := ensureProjectFolder(root, "evixdesk/backend"); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(filepath.Join(dir, "evixdesk", "backend")); err != nil || !info.IsDir() {
		t.Fatalf("folder was not created: %v", err)
	}
}

func TestEnsureProjectFolderKeepsExistingFolder(t *testing.T) {
	root, dir := openTestRoot(t)
	marker := filepath.Join(dir, "app", "main.go")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureProjectFolder(root, "app"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("existing content lost: %v", err)
	}
}

func TestEnsureProjectFolderRejectsFile(t *testing.T) {
	root, dir := openTestRoot(t)
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureProjectFolder(root, "notes.txt"); err == nil {
		t.Fatal("a file must not become a project folder")
	}
}

func TestEnsureProjectFolderStaysInsideRoot(t *testing.T) {
	root, dir := openTestRoot(t)
	outside := filepath.Join(filepath.Dir(dir), "escape-"+filepath.Base(dir))
	if err := os.Symlink(filepath.Dir(dir), filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	if err := ensureProjectFolder(root, "link/escape-"+filepath.Base(dir)); err == nil {
		t.Fatal("symlink escape must fail")
	}
	if _, err := os.Stat(outside); err == nil {
		t.Fatal("created a folder outside the root")
	}
}

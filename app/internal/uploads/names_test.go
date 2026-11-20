package uploads

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAttachmentNameKeepsOnlyTheBaseName(t *testing.T) {
	cases := map[string]string{
		"logs.zip":            "logs.zip",
		"../../etc/passwd":    "passwd",
		`C:\Users\me\pic.jpg`: "pic.jpg",
		"":                    fallbackName,
		"   ":                 fallbackName,
		"/":                   fallbackName,
		"..":                  fallbackName,
	}
	for in, want := range cases {
		if got := AttachmentName(in); got != want {
			t.Errorf("AttachmentName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDiskNameFitsTheFileSystemAndKeepsTheExtension(t *testing.T) {
	long := strings.Repeat("я", 200) + ".tar.gz"
	got := diskName(long)
	if len(got) > maxDiskNameBytes || !strings.HasSuffix(got, ".gz") || !strings.HasPrefix(got, "яя") {
		t.Fatalf("diskName = %q (%d bytes)", got, len(got))
	}
	if got := diskName("a\x00b"); got != "a_b" {
		t.Errorf("control characters kept: %q", got)
	}
}

func TestIDsAreUUIDs(t *testing.T) {
	id, err := newID()
	if err != nil || !isUUID(id) {
		t.Fatalf("id %q, err %v", id, err)
	}
	for _, bad := range []string{"", "../x", strings.ToUpper(id), id + "0"} {
		if isUUID(bad) {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestStoreKeepsPathsInsideFiles(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got := store.Path("../../etc/passwd"); !strings.HasPrefix(got, store.filesDir()+"/") {
		t.Errorf("path escaped: %q", got)
	}
	if _, err := store.Open("../partial"); err == nil {
		t.Error("opened a path outside files/")
	}
}

func TestDetectMediaTypeFallsBackToTheExtension(t *testing.T) {
	dir := t.TempDir()
	zip := filepath.Join(dir, "a")
	if err := os.WriteFile(zip, []byte("PK\x03\x04rest"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := detectMediaType(zip, "a.zip"); got != "application/zip" {
		t.Errorf("zip detected as %q", got)
	}
	blob := filepath.Join(dir, "b")
	if err := os.WriteFile(blob, []byte{0, 1, 2, 3}, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := detectMediaType(blob, "b.unknownext"); got != "application/octet-stream" {
		t.Errorf("blob detected as %q", got)
	}
}

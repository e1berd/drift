package workspace

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestListFoldersShowsOnlyVisibleFolders(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"beta", "Alpha/nested", ".cache", ".git"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	got := requestFolders(t, &Handler{root: root}, "/folders")
	if len(got.Entries) != 2 || got.Entries[0].Name != "Alpha" || got.Entries[1].Name != "beta" {
		t.Fatalf("unexpected entries %+v", got.Entries)
	}

	nested := requestFolders(t, &Handler{root: root}, "/folders?path=Alpha")
	if len(nested.Entries) != 1 || nested.Entries[0].Path != "Alpha/nested" {
		t.Fatalf("unexpected nested entries %+v", nested.Entries)
	}
}

func TestListFoldersStaysInsideRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "inside"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := requestFolders(t, &Handler{root: root}, "/folders?path=../..")
	if got.Path != "." || len(got.Entries) != 1 {
		t.Fatalf("escaped the root: %+v", got)
	}
}

func requestFolders(t *testing.T, h *Handler, target string) treeResponse {
	t.Helper()
	e := echo.New()
	e.GET("/folders", h.listFolders)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var out treeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

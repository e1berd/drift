package gitrepo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

type harness struct {
	t      *testing.T
	e      *echo.Echo
	folder string
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	config := "[user]\n\tname = Drift Test\n\temail = test@example.com\n[init]\n\tdefaultBranch = main\n"
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	folder := filepath.Join(t.TempDir(), "project")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	resolve := func(_ context.Context, _, session string) (string, string, error) {
		if session != "s1" {
			return "", "", echo.NewHTTPError(http.StatusNotFound, "not found")
		}
		return session, folder, nil
	}
	e := echo.New()
	New(resolve).Register(e.Group("/api"))
	return &harness{t: t, e: e, folder: folder}
}

func (h *harness) do(method, target, body string) *httptest.ResponseRecorder {
	h.t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.e.ServeHTTP(rec, req)
	return rec
}

func (h *harness) status(method, target, body string) Status {
	h.t.Helper()
	rec := h.do(method, target, body)
	if rec.Code != http.StatusOK {
		h.t.Fatalf("%s %s = %d: %s", method, target, rec.Code, rec.Body)
	}
	var s Status
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		h.t.Fatal(err)
	}
	return s
}

func (h *harness) write(name, content string) {
	h.t.Helper()
	full := filepath.Join(h.folder, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		h.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		h.t.Fatal(err)
	}
}

func TestStageUnstageAndCommitThroughEscapedColonRoutes(t *testing.T) {
	h := newHarness(t)

	if s := h.status(http.MethodGet, "/api/sessions/s1/git", ""); s.Repository {
		t.Fatalf("empty folder reported as a repository: %+v", s)
	}
	if s := h.status(http.MethodPost, "/api/sessions/s1/git:init", ""); !s.Repository || s.Branch != "main" {
		t.Fatalf("after init: %+v", s)
	}

	h.write("a.txt", "hello\n")
	h.write("sub/b.txt", "bee\n")
	s := h.status(http.MethodGet, "/api/sessions/s1/git", "")
	if len(s.Unstaged) != 2 || s.Unstaged[0].Status != Untracked {
		t.Fatalf("untracked files: %+v", s.Unstaged)
	}

	s = h.status(http.MethodPost, "/api/sessions/s1/git:stage", `{"paths":["a.txt"]}`)
	if len(s.Staged) != 1 || s.Staged[0] != (FileChange{Path: "a.txt", Status: Added}) || len(s.Unstaged) != 1 {
		t.Fatalf("after stage: %+v", s)
	}
	s = h.status(http.MethodPost, "/api/sessions/s1/git:unstage", `{"paths":[]}`)
	if len(s.Staged) != 0 || len(s.Unstaged) != 2 {
		t.Fatalf("after unstage all without HEAD: %+v", s)
	}

	h.status(http.MethodPost, "/api/sessions/s1/git:stage", `{"paths":[]}`)
	s = h.status(http.MethodPost, "/api/sessions/s1/git:commit", `{"message":"first commit"}`)
	if len(s.Staged)+len(s.Unstaged) != 0 || s.Head == "" {
		t.Fatalf("after commit: %+v", s)
	}

	h.write("a.txt", "hello\nworld\n")
	h.status(http.MethodPost, "/api/sessions/s1/git:stage", `{"paths":["a.txt"]}`)
	s = h.status(http.MethodPost, "/api/sessions/s1/git:unstage", `{"paths":["a.txt"]}`)
	if len(s.Staged) != 0 || len(s.Unstaged) != 1 || s.Unstaged[0].Status != Modified {
		t.Fatalf("after unstage with HEAD: %+v", s)
	}

	if rec := h.do(http.MethodPost, "/api/sessions/s1/git:commit", `{"message":"  "}`); rec.Code != http.StatusBadRequest {
		t.Errorf("empty message = %d", rec.Code)
	}
	if rec := h.do(http.MethodPost, "/api/sessions/s1/git:commit", `{"message":"nothing staged"}`); rec.Code != http.StatusConflict {
		t.Errorf("commit with nothing staged = %d: %s", rec.Code, rec.Body)
	}
}

func TestDiffShowsWorktreeStagedAndUntrackedFiles(t *testing.T) {
	h := newHarness(t)
	h.status(http.MethodPost, "/api/sessions/s1/git:init", "")
	h.write("a.txt", "one\n")
	h.status(http.MethodPost, "/api/sessions/s1/git:stage", `{"paths":["a.txt"]}`)
	h.status(http.MethodPost, "/api/sessions/s1/git:commit", `{"message":"one"}`)
	h.write("a.txt", "one\ntwo\n")
	h.write("new file.txt", "fresh\n")

	cases := map[string]string{
		"/api/sessions/s1/git/diff?path=a.txt":             "+two",
		"/api/sessions/s1/git/diff?path=new%20file.txt":    "+fresh",
		"/api/sessions/s1/git/diff?path=a.txt&staged=true": "",
	}
	for target, want := range cases {
		rec := h.do(http.MethodGet, target, "")
		var d Diff
		if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &d) != nil {
			t.Fatalf("GET %s = %d: %s", target, rec.Code, rec.Body)
		}
		if want == "" && d.Diff != "" || !strings.Contains(d.Diff, want) {
			t.Errorf("GET %s diff = %q, want %q", target, d.Diff, want)
		}
	}
}

func TestDiscardRestoresTrackedAndDeletesUntrackedFiles(t *testing.T) {
	h := newHarness(t)
	h.status(http.MethodPost, "/api/sessions/s1/git:init", "")
	h.write("a.txt", "one\n")
	h.status(http.MethodPost, "/api/sessions/s1/git:stage", `{"paths":[]}`)
	h.status(http.MethodPost, "/api/sessions/s1/git:commit", `{"message":"one"}`)
	h.write("a.txt", "changed\n")
	h.write("tmp/scratch.txt", "x\n")

	s := h.status(http.MethodPost, "/api/sessions/s1/git:discard", `{"paths":["a.txt","tmp/scratch.txt"]}`)
	if len(s.Unstaged) != 0 {
		t.Fatalf("after discard: %+v", s)
	}
	if got, _ := os.ReadFile(filepath.Join(h.folder, "a.txt")); string(got) != "one\n" {
		t.Errorf("a.txt = %q", got)
	}
	if _, err := os.Stat(filepath.Join(h.folder, "tmp/scratch.txt")); !os.IsNotExist(err) {
		t.Errorf("untracked file survived: %v", err)
	}
}

func TestPathsOutsideTheRepositoryAreRejected(t *testing.T) {
	h := newHarness(t)
	h.status(http.MethodPost, "/api/sessions/s1/git:init", "")
	outside := filepath.Join(filepath.Dir(h.folder), "secret.txt")
	if err := os.WriteFile(outside, []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{`{"paths":["../secret.txt"]}`, `{"paths":["` + outside + `"]}`, `{"paths":[".git/config"]}`} {
		for _, method := range []string{"stage", "discard"} {
			if rec := h.do(http.MethodPost, "/api/sessions/s1/git:"+method, body); rec.Code != http.StatusBadRequest {
				t.Errorf("%s %s = %d: %s", method, body, rec.Code, rec.Body)
			}
		}
	}
	if rec := h.do(http.MethodGet, "/api/sessions/s1/git/diff?path=../secret.txt", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("diff outside = %d", rec.Code)
	}
	if _, err := os.Stat(outside); err != nil {
		t.Errorf("file outside the repository was touched: %v", err)
	}
}

func TestBranchesCheckoutAndPagedCommits(t *testing.T) {
	h := newHarness(t)
	h.status(http.MethodPost, "/api/sessions/s1/git:init", "")
	for i, text := range []string{"one", "two", "three"} {
		h.write("a.txt", strings.Repeat(text, i+1))
		h.status(http.MethodPost, "/api/sessions/s1/git:stage", `{"paths":[]}`)
		h.status(http.MethodPost, "/api/sessions/s1/git:commit", `{"message":"commit `+text+`"}`)
	}

	s := h.status(http.MethodPost, "/api/sessions/s1/git:checkout", `{"branch":"feature/x","create":true}`)
	if s.Branch != "feature/x" {
		t.Fatalf("after create: %+v", s)
	}
	if s := h.status(http.MethodPost, "/api/sessions/s1/git:checkout", `{"branch":"main"}`); s.Branch != "main" {
		t.Fatalf("after switch: %+v", s)
	}
	if rec := h.do(http.MethodPost, "/api/sessions/s1/git:checkout", `{"branch":"--force"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("option as branch = %d", rec.Code)
	}

	var branches branchesResponse
	rec := h.do(http.MethodGet, "/api/sessions/s1/git/branches", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &branches); err != nil || len(branches.Branches) != 2 {
		t.Fatalf("branches = %s", rec.Body)
	}

	var first, second commitsResponse
	rec = h.do(http.MethodGet, "/api/sessions/s1/git/commits?page_size=2", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &first); err != nil || len(first.Commits) != 2 || first.NextPageToken == "" {
		t.Fatalf("first page = %s", rec.Body)
	}
	if first.Commits[0].Subject != "commit three" || first.Commits[0].AuthorName != "Drift Test" || first.Commits[0].AuthorTime.IsZero() {
		t.Errorf("first commit = %+v", first.Commits[0])
	}
	h.write("a.txt", "four")
	h.status(http.MethodPost, "/api/sessions/s1/git:stage", `{"paths":[]}`)
	h.status(http.MethodPost, "/api/sessions/s1/git:commit", `{"message":"commit four"}`)
	rec = h.do(http.MethodGet, "/api/sessions/s1/git/commits?page_size=2&page_token="+first.NextPageToken, "")
	if err := json.Unmarshal(rec.Body.Bytes(), &second); err != nil || len(second.Commits) != 1 || second.NextPageToken != "" {
		t.Fatalf("second page = %s", rec.Body)
	}
	if second.Commits[0].Subject != "commit one" {
		t.Errorf("second page shifted: %+v", second.Commits)
	}
}

func TestFolderInsideAnotherCheckoutIsNotARepository(t *testing.T) {
	h := newHarness(t)
	parent := filepath.Dir(h.folder)
	if out, err := exec.Command("git", "-C", parent, "init", "--quiet").CombinedOutput(); err != nil {
		t.Fatalf("git init parent: %v %s", err, out)
	}
	if s := h.status(http.MethodGet, "/api/sessions/s1/git", ""); s.Repository {
		t.Fatalf("parent checkout leaked into the session folder: %+v", s)
	}
	if rec := h.do(http.MethodPost, "/api/sessions/s1/git:stage", `{"paths":[]}`); rec.Code != http.StatusConflict {
		t.Errorf("stage outside a repository = %d", rec.Code)
	}
}

func TestUnknownSessionIsNotFound(t *testing.T) {
	h := newHarness(t)
	for _, target := range []string{"/api/sessions/other/git", "/api/sessions/other/git/branches"} {
		if rec := h.do(http.MethodGet, target, ""); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d", target, rec.Code)
		}
	}
	if rec := h.do(http.MethodPost, "/api/sessions/other/git:pull", ""); rec.Code != http.StatusNotFound {
		t.Errorf("POST git:pull = %d", rec.Code)
	}
}

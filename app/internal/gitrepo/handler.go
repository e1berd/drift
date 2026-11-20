package gitrepo

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/e1berd/drift/app/internal/identity"
	"github.com/labstack/echo/v5"
)

const (
	maxPaths         = 1000
	maxMessageBytes  = 64 << 10
	maxBranchNameLen = 250
)

// FolderResolver checks that the session belongs to owner and returns its canonical id and the folder it works in.
// Errors should be echo HTTP errors, e.g. 404 for a session of somebody else.
type FolderResolver func(ctx context.Context, owner, session string) (id, folder string, err error)

// Handler serves /sessions/{session}/git and its custom methods.
type Handler struct {
	resolve FolderResolver
	history History
	locks   sync.Map
}

// New returns the handler; resolve finds the folder of a session.
func New(resolve FolderResolver) *Handler {
	return &Handler{resolve: resolve}
}

// WithHistory enables the session summary, which needs the transcript and the base commit of a session.
func (h *Handler) WithHistory(history History) *Handler {
	h.history = history
	return h
}

// Register mounts the routes on a group that already requires a signed-in identity.
// Colons of custom methods are escaped, otherwise echo reads them as path parameters.
func (h *Handler) Register(g *echo.Group) {
	g.GET("/sessions/:session/git", h.getStatus)
	g.GET("/sessions/:session/git/summary", h.getSummary)
	g.GET("/sessions/:session/git/diff", h.getDiff)
	g.GET("/sessions/:session/git/branches", h.listBranches)
	g.GET("/sessions/:session/git/commits", h.listCommits)
	g.POST(`/sessions/:session/git\:init`, h.mutate(initRepo))
	g.POST(`/sessions/:session/git\:stage`, h.mutate(stage))
	g.POST(`/sessions/:session/git\:unstage`, h.mutate(unstage))
	g.POST(`/sessions/:session/git\:discard`, h.mutate(discard))
	g.POST(`/sessions/:session/git\:commit`, h.mutate(commit))
	g.POST(`/sessions/:session/git\:checkout`, h.mutate(checkout))
	g.POST(`/sessions/:session/git\:pull`, h.mutate(pull))
	g.POST(`/sessions/:session/git\:push`, h.mutate(push))
}

func (h *Handler) open(c *echo.Context) (repo, error) {
	_, folder, err := h.resolve(c.Request().Context(), identity.From(c), c.Param("session"))
	if err != nil {
		return repo{}, err
	}
	return open(folder)
}

func (h *Handler) reply(c *echo.Context, value any, err error) error {
	var gitErr *Error
	if errors.As(err, &gitErr) {
		return echo.NewHTTPError(http.StatusConflict, gitErr.Message)
	}
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, value)
}

func (h *Handler) getStatus(c *echo.Context) error {
	r, err := h.open(c)
	if err != nil {
		return err
	}
	status, err := r.status(c.Request().Context())
	return h.reply(c, status, err)
}

func (h *Handler) getSummary(c *echo.Context) error {
	ctx := c.Request().Context()
	id, folder, err := h.resolve(ctx, identity.From(c), c.Param("session"))
	if err != nil {
		return err
	}
	r, err := open(folder)
	if err != nil {
		return err
	}
	var base string
	var calls []ToolCall
	if h.history != nil {
		if base, calls, err = h.history(ctx, id); err != nil {
			return err
		}
	}
	summary, err := r.summary(ctx, base, calls)
	return h.reply(c, summary, err)
}

func (h *Handler) getDiff(c *echo.Context) error {
	r, err := h.open(c)
	if err != nil {
		return err
	}
	path, err := cleanPath(c.QueryParam("path"))
	if err != nil {
		return err
	}
	staged, _ := strconv.ParseBool(c.QueryParam("staged"))
	diff, err := r.diff(c.Request().Context(), path, staged)
	return h.reply(c, diff, err)
}

type branchesResponse struct {
	Branches []Branch `json:"branches"`
}

func (h *Handler) listBranches(c *echo.Context) error {
	r, err := h.open(c)
	if err != nil {
		return err
	}
	list, err := r.branches(c.Request().Context())
	return h.reply(c, branchesResponse{Branches: list}, err)
}

type commitsResponse struct {
	Commits       []Commit `json:"commits"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
}

func (h *Handler) listCommits(c *echo.Context) error {
	page, ok := parsePageToken(c.QueryParam("page_token"))
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid page_token")
	}
	r, err := h.open(c)
	if err != nil {
		return err
	}
	list, next, err := r.commits(c.Request().Context(), page, pageSize(c))
	return h.reply(c, commitsResponse{Commits: list, NextPageToken: next}, err)
}

func pageSize(c *echo.Context) int {
	n, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil || n <= 0 {
		return defaultPageSize
	}
	return min(n, maxPageSize)
}

type mutation func(ctx context.Context, r repo, c *echo.Context) error

func (h *Handler) mutate(change mutation) echo.HandlerFunc {
	return func(c *echo.Context) error {
		r, err := h.open(c)
		if err != nil {
			return err
		}
		lock, _ := h.locks.LoadOrStore(r.dir, &sync.Mutex{})
		mu := lock.(*sync.Mutex)
		mu.Lock()
		defer mu.Unlock()
		ctx := c.Request().Context()
		if err := change(ctx, r, c); err != nil {
			return h.reply(c, nil, err)
		}
		status, err := r.status(ctx)
		return h.reply(c, status, err)
	}
}

func requireRepository(ctx context.Context, r repo) error {
	ok, err := r.isRepository(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return &Error{Message: "not a git repository"}
	}
	return nil
}

func initRepo(ctx context.Context, r repo, _ *echo.Context) error {
	if ok, err := r.isRepository(ctx); err != nil || ok {
		return err
	}
	_, err := r.git(ctx, "init")
	return err
}

type pathsRequest struct {
	Paths []string `json:"paths"`
}

func bindPaths(c *echo.Context) ([]string, error) {
	var req pathsRequest
	if err := c.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if len(req.Paths) > maxPaths {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "too many paths")
	}
	return CleanPaths(req.Paths)
}

func orAll(paths []string) []string {
	if len(paths) == 0 {
		return []string{"."}
	}
	return paths
}

func stage(ctx context.Context, r repo, c *echo.Context) error {
	paths, err := bindPaths(c)
	if err != nil {
		return err
	}
	if err := requireRepository(ctx, r); err != nil {
		return err
	}
	_, err = r.git(ctx, append([]string{"add", "--all", "--"}, orAll(paths)...)...)
	return err
}

func unstage(ctx context.Context, r repo, c *echo.Context) error {
	paths, err := bindPaths(c)
	if err != nil {
		return err
	}
	if err := requireRepository(ctx, r); err != nil {
		return err
	}
	if r.hasHead(ctx) {
		_, err = r.git(ctx, append([]string{"restore", "--staged", "--"}, orAll(paths)...)...)
	} else {
		_, err = r.git(ctx, append([]string{"rm", "--cached", "-r", "--quiet", "--ignore-unmatch", "--"}, orAll(paths)...)...)
	}
	return err
}

func discard(ctx context.Context, r repo, c *echo.Context) error {
	paths, err := bindPaths(c)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "paths are required")
	}
	if err := requireRepository(ctx, r); err != nil {
		return err
	}
	untracked, err := r.untracked(ctx, paths)
	if err != nil {
		return err
	}
	tracked := make([]string, 0, len(paths))
	for _, p := range paths {
		if !untracked[p] {
			tracked = append(tracked, p)
		}
	}
	if len(tracked) > 0 {
		if _, err := r.git(ctx, append([]string{"restore", "--worktree", "--"}, tracked...)...); err != nil {
			return err
		}
	}
	return r.removeUntracked(untracked)
}

func (r repo) removeUntracked(paths map[string]bool) error {
	if len(paths) == 0 {
		return nil
	}
	root, err := os.OpenRoot(r.dir)
	if err != nil {
		return err
	}
	defer root.Close()
	for p := range paths {
		if err := root.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return &Error{Message: "cannot delete " + p}
		}
	}
	return nil
}

type commitRequest struct {
	Message string `json:"message"`
}

func commit(ctx context.Context, r repo, c *echo.Context) error {
	var req commitRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	message := strings.TrimSpace(req.Message)
	if message == "" || len(message) > maxMessageBytes {
		return echo.NewHTTPError(http.StatusBadRequest, "commit message is required")
	}
	if err := requireRepository(ctx, r); err != nil {
		return err
	}
	_, err := r.git(ctx, "commit", "--quiet", "--message", message)
	return err
}

type checkoutRequest struct {
	Branch string `json:"branch"`
	Create bool   `json:"create"`
}

func checkout(ctx context.Context, r repo, c *echo.Context) error {
	var req checkoutRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	name := strings.TrimSpace(req.Branch)
	if name == "" || strings.HasPrefix(name, "-") || len(name) > maxBranchNameLen {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid branch name")
	}
	if err := requireRepository(ctx, r); err != nil {
		return err
	}
	if _, err := r.git(ctx, "check-ref-format", "--branch", name); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid branch name")
	}
	args := []string{"switch", "--quiet"}
	if req.Create {
		args = append(args, "--create")
	}
	_, err := r.git(ctx, append(args, name)...)
	return err
}

func pull(ctx context.Context, r repo, _ *echo.Context) error {
	if err := requireRepository(ctx, r); err != nil {
		return err
	}
	return r.remote(ctx, "pull", "--ff-only", "--quiet")
}

func push(ctx context.Context, r repo, _ *echo.Context) error {
	if err := requireRepository(ctx, r); err != nil {
		return err
	}
	status, err := r.status(ctx)
	if err != nil {
		return err
	}
	if status.Detached || status.Branch == "" {
		return &Error{Message: "HEAD is detached: switch to a branch to push"}
	}
	if status.Upstream != "" {
		return r.remote(ctx, "push", "--quiet")
	}
	return r.remote(ctx, "push", "--quiet", "--set-upstream", "origin", status.Branch)
}

package gitrepo

import (
	"context"
	"net/http"
	"sync"

	"github.com/e1berd/drift/app/internal/identity"
	"github.com/labstack/echo/v5"
)

type sessionRepo struct {
	id      string
	repo    repo
	history History
}

func (h *Handler) openSession(c *echo.Context) (sessionRepo, error) {
	ctx := c.Request().Context()
	id, folder, err := h.resolve(ctx, identity.From(c), c.Param("session"))
	if err != nil {
		return sessionRepo{}, err
	}
	r, err := open(folder)
	if err != nil {
		return sessionRepo{}, err
	}
	var history History
	if h.journal != nil {
		if history, err = h.journal.History(ctx, id); err != nil {
			return sessionRepo{}, err
		}
	}
	return sessionRepo{id: id, repo: r, history: history}, nil
}

func (h *Handler) getSummary(c *echo.Context) error {
	s, err := h.openSession(c)
	if err != nil {
		return err
	}
	summary, err := s.repo.summary(c.Request().Context(), s.history)
	return h.reply(c, summary, err)
}

type removeFileRequest struct {
	Name string `json:"name"`
}

type removeWorktreeRequest struct {
	Name         string `json:"name"`
	Force        bool   `json:"force"`
	DeleteBranch bool   `json:"deleteBranch"`
}

func (h *Handler) removeFile(c *echo.Context) error {
	var body removeFileRequest
	if err := c.Bind(&body); err != nil || body.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	return h.remove(c, func(ctx context.Context, s sessionRepo) (Removal, error) {
		return s.repo.removeFile(ctx, s.history, body.Name)
	})
}

func (h *Handler) removeWorktree(c *echo.Context) error {
	var body removeWorktreeRequest
	if err := c.Bind(&body); err != nil || body.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	return h.remove(c, func(ctx context.Context, s sessionRepo) (Removal, error) {
		return s.repo.removeWorktree(ctx, s.history, body.Name, body.Force, body.DeleteBranch)
	})
}

func (h *Handler) remove(c *echo.Context, removeEntry func(context.Context, sessionRepo) (Removal, error)) error {
	if h.journal == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "removals are not available")
	}
	s, err := h.openSession(c)
	if err != nil {
		return err
	}
	lock, _ := h.locks.LoadOrStore(s.repo.dir, &sync.Mutex{})
	mu := lock.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()
	ctx := c.Request().Context()
	removal, removeErr := removeEntry(ctx, s)
	if removal.Target != "" {
		if err := h.journal.RecordRemoval(ctx, identity.From(c), s.id, removal); err != nil {
			return err
		}
		s.history.Removals = append(s.history.Removals, removal)
	}
	if removeErr != nil {
		return h.reply(c, nil, removeErr)
	}
	summary, err := s.repo.summary(ctx, s.history)
	return h.reply(c, summary, err)
}

// Package settings serves the instance-wide settings singleton (AIP-156): every signed-in user
// reads it, only the instance owner changes it.
package settings

import (
	"context"
	"net/http"
	"time"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

const (
	minUploadBytes = 10_000_000
	maxUploadBytes = 1_000_000_000_000
)

// Settings is the instance configuration editable from the web UI.
type Settings struct {
	MaxUploadBytes int64     `json:"maxUploadBytes"`
	UpdateTime     time.Time `json:"updateTime"`
}

type settingsPatch struct {
	MaxUploadBytes *int64 `json:"maxUploadBytes"`
}

// Handler owns the settings API.
type Handler struct {
	q *db.Queries
}

// New returns the settings handler.
func New(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

// Register mounts the routes on a group that already requires a signed-in identity.
func (h *Handler) Register(g *echo.Group) {
	g.GET("/settings", h.get)
	g.PATCH("/settings", h.update, identity.RequireSuperuser)
}

// MaxUploadBytes is the largest file a user may upload right now.
func (h *Handler) MaxUploadBytes(ctx context.Context) (int64, error) {
	row, err := h.q.GetInstanceSettings(ctx)
	return row.MaxUploadBytes, err
}

func (h *Handler) get(c *echo.Context) error {
	row, err := h.q.GetInstanceSettings(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, Settings{MaxUploadBytes: row.MaxUploadBytes, UpdateTime: row.UpdateTime.Time})
}

func (h *Handler) update(c *echo.Context) error {
	var patch settingsPatch
	if err := c.Bind(&patch); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid settings")
	}
	ctx := c.Request().Context()
	current, err := h.q.GetInstanceSettings(ctx)
	if err != nil {
		return err
	}
	limit := current.MaxUploadBytes
	if patch.MaxUploadBytes != nil {
		limit = *patch.MaxUploadBytes
	}
	if limit < minUploadBytes || limit > maxUploadBytes {
		return echo.NewHTTPError(http.StatusBadRequest, "maxUploadBytes must be between 10 MB and 1 TB")
	}
	row, err := h.q.UpdateInstanceSettings(ctx, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, Settings{MaxUploadBytes: row.MaxUploadBytes, UpdateTime: row.UpdateTime.Time})
}

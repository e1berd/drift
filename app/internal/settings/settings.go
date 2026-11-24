// Package settings serves the instance-wide settings singleton (AIP-156): every signed-in user
// reads it, only the instance owner changes it.
package settings

import (
	"context"
	"net/http"
	"time"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/frontend"
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
	MaxUploadBytes  int64     `json:"maxUploadBytes"`
	DefaultFrontend string    `json:"defaultFrontend"`
	UpdateTime      time.Time `json:"updateTime"`
}

type settingsPatch struct {
	MaxUploadBytes  *int64  `json:"maxUploadBytes"`
	DefaultFrontend *string `json:"defaultFrontend"`
}

// Handler owns the settings API.
type Handler struct {
	q         *db.Queries
	frontends func(id string) bool
}

// New returns the settings handler; frontends reports whether an interface id exists.
func New(pool *pgxpool.Pool, frontends func(id string) bool) *Handler {
	return &Handler{q: db.New(pool), frontends: frontends}
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

// DefaultFrontend is the interface pages get when the visitor has not picked one.
func (h *Handler) DefaultFrontend(ctx context.Context) (string, error) {
	row, err := h.q.GetInstanceSettings(ctx)
	return row.DefaultFrontend, err
}

func present(maxUpload int64, defaultFrontend string, updated time.Time) Settings {
	return Settings{
		MaxUploadBytes:  maxUpload,
		DefaultFrontend: frontend.ResourceName(defaultFrontend),
		UpdateTime:      updated,
	}
}

func (h *Handler) get(c *echo.Context) error {
	row, err := h.q.GetInstanceSettings(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, present(row.MaxUploadBytes, row.DefaultFrontend, row.UpdateTime.Time))
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
	next := db.UpdateInstanceSettingsParams{
		MaxUploadBytes:  current.MaxUploadBytes,
		DefaultFrontend: current.DefaultFrontend,
	}
	if patch.MaxUploadBytes != nil {
		next.MaxUploadBytes = *patch.MaxUploadBytes
	}
	if next.MaxUploadBytes < minUploadBytes || next.MaxUploadBytes > maxUploadBytes {
		return echo.NewHTTPError(http.StatusBadRequest, "maxUploadBytes must be between 10 MB and 1 TB")
	}
	if patch.DefaultFrontend != nil {
		next.DefaultFrontend = frontend.IDFromResourceName(*patch.DefaultFrontend)
		if !h.frontends(next.DefaultFrontend) {
			return echo.NewHTTPError(http.StatusBadRequest, "defaultFrontend must name an installed interface")
		}
	}
	row, err := h.q.UpdateInstanceSettings(ctx, next)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, present(row.MaxUploadBytes, row.DefaultFrontend, row.UpdateTime.Time))
}

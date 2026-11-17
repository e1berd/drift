package health

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Register(e *echo.Echo) {
	e.GET("/healthz", h.Health)
	e.GET("/readyz", h.Ready)
}

func (h *Handler) Health(c *echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *Handler) Ready(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(
		c.Request().Context(),
		500*time.Millisecond,
	)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		return c.NoContent(http.StatusServiceUnavailable)
	}
	return c.NoContent(http.StatusOK)
}

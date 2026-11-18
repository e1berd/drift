// Package workspace serves the JSON API for projects, sessions, messages, drafts and project files.
// Names follow the Google API design guide: list methods take page_size/page_token and return
// next_page_token (AIP-158), resources expose createTime/updateTime (AIP-142).
package workspace

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/events"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

const (
	defaultPageSize = 30
	maxPageSize     = 100
	deviceHeader    = "X-Drift-Device"
)

// Handler owns the workspace API.
type Handler struct {
	pool    *pgxpool.Pool
	q       *db.Queries
	hub     *events.Hub
	root    string
	folders Folders
	log     *slog.Logger
	runner  AgentRunner
}

// New returns the handler; project paths are relative to folders.Workspace.
func New(pool *pgxpool.Pool, hub *events.Hub, folders Folders, log *slog.Logger) *Handler {
	return &Handler{pool: pool, q: db.New(pool), hub: hub, root: folders.Workspace, folders: folders, log: log}
}

// WithRunner lets new user messages start the session's agent.
func (h *Handler) WithRunner(runner AgentRunner) *Handler {
	h.runner = runner
	return h
}

// Register mounts the routes on a group that already requires a signed-in identity.
func (h *Handler) Register(g *echo.Group) {
	g.GET("/folders", h.listFolders)
	g.GET("/projects", h.listProjects)
	g.POST("/projects", h.createProject)
	g.PATCH("/projects/:project", h.renameProject)
	g.DELETE("/projects/:project", h.deleteProject)
	g.GET("/projects/:project/sessions", h.listSessions)
	g.POST("/projects/:project/sessions", h.createSession)
	g.GET("/projects/:project/tree", h.listTree)
	g.POST("/projects/:project/tree", h.createTreeEntry)
	g.GET("/projects/:project/file", h.readFile)
	g.PATCH("/projects/:project/file", h.updateFile)

	g.GET("/sessions", h.listStandaloneSessions)
	g.POST("/sessions", h.createStandaloneSession)
	g.GET("/sessions/:session", h.getSession)
	g.PATCH("/sessions/:session", h.updateSession)
	g.POST("/sessions/:session", h.sessionCustomMethod)
	g.DELETE("/sessions/:session", h.deleteSession)
	g.GET("/sessions/:session/tree", h.listTree)
	g.POST("/sessions/:session/tree", h.createTreeEntry)
	g.GET("/sessions/:session/file", h.readFile)
	g.PATCH("/sessions/:session/file", h.updateFile)
	g.GET("/sessions/:session/messages", h.listMessages)
	g.POST("/sessions/:session/messages", h.createMessage)

	g.GET("/drafts", h.listDrafts)
	g.PUT("/drafts/:key", h.putDraft)
	g.DELETE("/drafts/:key", h.deleteDraft)
}

func (h *Handler) publish(c *echo.Context, typ string, data any) {
	h.hub.Publish(identity.From(c), typ, data, device(c))
}

func device(c *echo.Context) string {
	d := c.Request().Header.Get(deviceHeader)
	if len(d) > 64 {
		d = d[:64]
	}
	return d
}

func parseID(c *echo.Context, name string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(c.Param(name)); err != nil {
		return id, echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	return id, nil
}

func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	return err
}

func badRequest(msg string) error {
	return echo.NewHTTPError(http.StatusBadRequest, msg)
}

func pageSize(c *echo.Context) int32 {
	n, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil || n <= 0 {
		return defaultPageSize
	}
	return int32(min(n, maxPageSize))
}

type cursor struct {
	T  time.Time `json:"t"`
	ID string    `json:"i"`
}

func encodeCursor(t pgtype.Timestamptz, id pgtype.UUID) string {
	b, _ := json.Marshal(cursor{T: t.Time, ID: uuidString(id)})
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(c *echo.Context) (pgtype.Timestamptz, pgtype.UUID, error) {
	var ts pgtype.Timestamptz
	var id pgtype.UUID
	token := c.QueryParam("page_token")
	if token == "" {
		return ts, id, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return ts, id, badRequest("invalid page_token")
	}
	var cur cursor
	if err := json.Unmarshal(raw, &cur); err != nil || id.Scan(cur.ID) != nil {
		return ts, id, badRequest("invalid page_token")
	}
	ts = pgtype.Timestamptz{Time: cur.T, Valid: true}
	return ts, id, nil
}

func uuidString(id pgtype.UUID) string {
	v, _ := id.Value()
	s, _ := v.(string)
	return s
}

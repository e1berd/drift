package workspace

import (
	"net/http"
	"strings"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

// RunRequest asks the agent of a session to answer the user's latest message.
type RunRequest struct {
	Owner       string
	Session     pgtype.UUID
	Text        string
	Attachments []pgtype.UUID
	Model       string
	Mode        string
}

// AgentRunner runs agent CLIs for sessions; requests for a busy session wait in its queue.
type AgentRunner interface {
	Enqueue(RunRequest)
	Stop(owner string, session pgtype.UUID) bool
}

var permissionModes = map[string]bool{
	"": true, "default": true, "acceptEdits": true, "auto": true, "plan": true, "bypassPermissions": true,
}

func validModelName(model string) bool {
	return len(model) <= 100 && !strings.ContainsAny(model, " \t\r\n")
}

type runOptions struct {
	Model string `json:"model"`
	Mode  string `json:"mode"`
}

func (o runOptions) validate() error {
	if !permissionModes[o.Mode] {
		return badRequest("unknown permission mode")
	}
	if !validModelName(o.Model) {
		return badRequest("invalid model")
	}
	return nil
}

func (h *Handler) startRun(c *echo.Context, session pgtype.UUID, text string, attachments []pgtype.UUID, opts runOptions) {
	if h.runner == nil || text == "" && len(attachments) == 0 {
		return
	}
	h.runner.Enqueue(RunRequest{
		Owner: identity.From(c), Session: session, Text: text, Attachments: attachments,
		Model: opts.Model, Mode: opts.Mode,
	})
}

// sessionCustomMethod serves AIP-136 custom methods: POST /sessions/{id}:stop.
func (h *Handler) sessionCustomMethod(c *echo.Context) error {
	raw, method, _ := strings.Cut(c.Param("session"), ":")
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	owner := identity.From(c)
	if _, err := h.q.GetSession(c.Request().Context(), db.GetSessionParams{ID: id, OwnerID: owner}); err != nil {
		return notFound(err)
	}
	switch method {
	case "stop":
		if h.runner != nil {
			h.runner.Stop(owner, id)
		}
		return c.NoContent(http.StatusNoContent)
	}
	return echo.NewHTTPError(http.StatusNotFound, "not found")
}

// SessionOf is the API view of a stored session.
func SessionOf(s db.Session) Session { return sessionDTO(s) }

// MessageOf is the API view of a stored message.
func MessageOf(m db.Message) Message { return messageDTO(m) }

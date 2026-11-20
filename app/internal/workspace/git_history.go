package workspace

import (
	"context"
	"net/http"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/gitrepo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

// SessionGitHistory returns the commit a session started from and its git-related tool calls; the
// session id must already be checked against its owner, as SessionFolder does.
func (h *Handler) SessionGitHistory(ctx context.Context, session string) (string, []gitrepo.ToolCall, error) {
	var id pgtype.UUID
	if err := id.Scan(session); err != nil {
		return "", nil, echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	base, err := h.q.GetSessionGitBase(ctx, id)
	if err != nil {
		return "", nil, notFound(err)
	}
	rows, err := h.q.ListSessionGitCalls(ctx, id)
	if err != nil {
		return "", nil, err
	}
	return base, toolCalls(rows), nil
}

func toolCalls(rows []db.ListSessionGitCallsRow) []gitrepo.ToolCall {
	calls := make([]gitrepo.ToolCall, 0, len(rows))
	for _, row := range rows {
		calls = append(calls, gitrepo.ToolCall{MessageID: uuidString(row.ID), Name: row.Name, Input: row.Input})
	}
	return calls
}

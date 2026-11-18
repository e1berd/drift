package workspace

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/labstack/echo/v5"
)

const maxSearchQueryRunes = 200

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func likeLiteral(text string) string {
	return likeEscaper.Replace(text)
}

func (h *Handler) searchSessions(c *echo.Context) error {
	query := strings.TrimSpace(c.QueryParam("query"))
	if query == "" {
		return badRequest("query is required")
	}
	if utf8.RuneCountInString(query) > maxSearchQueryRunes {
		return badRequest("query is too long")
	}
	before, beforeID, err := decodeCursor(c)
	if err != nil {
		return err
	}
	size := pageSize(c)
	rows, err := h.q.SearchSessions(c.Request().Context(), db.SearchSessionsParams{
		OwnerID: identity.From(c), Limit: size + 1, Pattern: likeLiteral(query),
		BeforeUpdated: before, BeforeID: beforeID,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sessionPage(rows, size))
}

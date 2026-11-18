package workspace

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

const standaloneDraftKey = "new:chat"

// Folders says where agents work: a project session in its project folder under Workspace,
// a standalone session (no project) in a private empty folder of its own under Chats.
type Folders struct {
	Workspace string
	Chats     string
}

// Session returns the folder the session's agent works in, creating a standalone one on demand.
func (f Folders) Session(projectPath pgtype.Text, session pgtype.UUID) (string, error) {
	if projectPath.Valid {
		if projectPath.String != "." && !filepath.IsLocal(projectPath.String) {
			return "", errors.New("project path is outside the workspace")
		}
		return filepath.Join(f.Workspace, projectPath.String), nil
	}
	dir := filepath.Join(f.Chats, uuidString(session))
	return dir, os.MkdirAll(dir, 0o755)
}

// RemoveChat deletes the folder of a standalone session.
func (f Folders) RemoveChat(session pgtype.UUID) error {
	id := uuidString(session)
	if id == "" {
		return errors.New("invalid session id")
	}
	return os.RemoveAll(filepath.Join(f.Chats, id))
}

func (h *Handler) listStandaloneSessions(c *echo.Context) error {
	before, beforeID, err := decodeCursor(c)
	if err != nil {
		return err
	}
	size := pageSize(c)
	rows, err := h.q.ListStandaloneSessions(c.Request().Context(), db.ListStandaloneSessionsParams{
		OwnerID: identity.From(c), Limit: size + 1, BeforeUpdated: before, BeforeID: beforeID,
	})
	if err != nil {
		return err
	}
	out := listSessionsResponse{Sessions: make([]Session, 0, len(rows))}
	for i, s := range rows {
		if int32(i) == size {
			last := rows[i-1]
			out.NextPageToken = encodeCursor(last.UpdatedAt, last.ID)
			break
		}
		out.Sessions = append(out.Sessions, sessionDTO(s))
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) createStandaloneSession(c *echo.Context) error {
	return h.insertSession(c, pgtype.UUID{}, standaloneDraftKey)
}

func (h *Handler) sessionRoot(c *echo.Context) (*os.Root, error) {
	id, err := parseID(c, "session")
	if err != nil {
		return nil, err
	}
	rc, err := h.q.GetRunContext(c.Request().Context(), db.GetRunContextParams{ID: id, OwnerID: identity.From(c)})
	if err != nil {
		return nil, notFound(err)
	}
	if rc.ProjectPath.Valid {
		return openProjectRoot(h.root, rc.ProjectPath.String)
	}
	dir, err := h.folders.Session(rc.ProjectPath, rc.ID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusServiceUnavailable, "session folder is not available")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "session folder is missing")
	}
	return root, nil
}

// SessionFolder returns the canonical id and working folder of a session that belongs to owner.
func (h *Handler) SessionFolder(ctx context.Context, owner, session string) (string, string, error) {
	var id pgtype.UUID
	if err := id.Scan(session); err != nil {
		return "", "", echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	rc, err := h.q.GetRunContext(ctx, db.GetRunContextParams{ID: id, OwnerID: owner})
	if err != nil {
		return "", "", notFound(err)
	}
	dir, err := h.folders.Session(rc.ProjectPath, rc.ID)
	if err != nil {
		return "", "", echo.NewHTTPError(http.StatusServiceUnavailable, "session folder is not available")
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "", "", echo.NewHTTPError(http.StatusNotFound, "session folder is missing")
	}
	return uuidString(rc.ID), dir, nil
}

package workspace

import (
	"errors"
	"net/http"
	"os"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/labstack/echo/v5"
)

type listProjectsResponse struct {
	Projects      []Project `json:"projects"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
}

func (h *Handler) listProjects(c *echo.Context) error {
	after, afterID, err := decodeCursor(c)
	if err != nil {
		return err
	}
	size := pageSize(c)
	rows, err := h.q.ListProjects(c.Request().Context(), db.ListProjectsParams{
		OwnerID: identity.From(c), Limit: size + 1, AfterCreated: after, AfterID: afterID,
	})
	if err != nil {
		return err
	}
	out := listProjectsResponse{Projects: make([]Project, 0, len(rows))}
	for i, r := range rows {
		if int32(i) == size {
			last := rows[i-1]
			out.NextPageToken = encodeCursor(last.CreatedAt, last.ID)
			break
		}
		p := db.Project{ID: r.ID, OwnerID: r.OwnerID, Name: r.Name, Path: r.Path, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
		out.Projects = append(out.Projects, projectDTO(p, r.SessionCount))
	}
	return c.JSON(http.StatusOK, out)
}

type projectInput struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func cleanName(s string, max int) (string, bool) {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	return s, n > 0 && n <= max
}

func ensureProjectFolder(root *os.Root, rel string) error {
	if err := root.MkdirAll(rel, 0o755); err != nil {
		return err
	}
	info, err := root.Stat(rel)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("not a folder")
	}
	return nil
}

func (h *Handler) createProject(c *echo.Context) error {
	var in projectInput
	if err := c.Bind(&in); err != nil {
		return badRequest("invalid body")
	}
	name, ok := cleanName(in.Name, 120)
	if !ok {
		return badRequest("name must be 1–120 characters")
	}
	rel := path.Clean(strings.TrimSpace(strings.ReplaceAll(in.Path, "\\", "/")))
	if rel == "." || rel == "" || path.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, "../") {
		return badRequest("path must be a folder inside the workspace root")
	}
	root, err := os.OpenRoot(h.root)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "workspace root is not available")
	}
	defer root.Close()
	if err := ensureProjectFolder(root, rel); err != nil {
		return badRequest("cannot use " + rel + " as a project folder: " + err.Error())
	}

	p, err := h.q.CreateProject(c.Request().Context(), db.CreateProjectParams{OwnerID: identity.From(c), Name: name, Path: rel})
	if err != nil {
		return err
	}
	dto := projectDTO(p, 0)
	h.publish(c, "project.created", dto)
	return c.JSON(http.StatusCreated, dto)
}

func (h *Handler) renameProject(c *echo.Context) error {
	id, err := parseID(c, "project")
	if err != nil {
		return err
	}
	var in projectInput
	if err := c.Bind(&in); err != nil {
		return badRequest("invalid body")
	}
	name, ok := cleanName(in.Name, 120)
	if !ok {
		return badRequest("name must be 1–120 characters")
	}
	ctx := c.Request().Context()
	p, err := h.q.RenameProject(ctx, db.RenameProjectParams{ID: id, OwnerID: identity.From(c), Name: name})
	if err != nil {
		return notFound(err)
	}
	count, err := h.q.CountSessions(ctx, p.ID)
	if err != nil {
		return err
	}
	dto := projectDTO(p, count)
	h.publish(c, "project.updated", dto)
	return c.JSON(http.StatusOK, dto)
}

func (h *Handler) deleteProject(c *echo.Context) error {
	id, err := parseID(c, "project")
	if err != nil {
		return err
	}
	owner := identity.From(c)
	ctx := c.Request().Context()
	n, err := h.q.DeleteProject(ctx, db.DeleteProjectParams{ID: id, OwnerID: owner})
	if err != nil {
		return err
	}
	if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	if err := h.q.DeleteDraftsForProject(ctx, db.DeleteDraftsForProjectParams{OwnerID: owner, ProjectID: uuidString(id)}); err != nil {
		h.log.Warn("delete project drafts", "err", err)
	}
	h.publish(c, "project.deleted", deleted{ID: uuidString(id)})
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) project(c *echo.Context) (db.Project, error) {
	id, err := parseID(c, "project")
	if err != nil {
		return db.Project{}, err
	}
	p, err := h.q.GetProject(c.Request().Context(), db.GetProjectParams{ID: id, OwnerID: identity.From(c)})
	return p, notFound(err)
}

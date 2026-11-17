package workspace

import (
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
)

func isVisibleFolder(entry fs.DirEntry) bool {
	return entry.IsDir() && !strings.HasPrefix(entry.Name(), ".")
}

func (h *Handler) listFolders(c *echo.Context) error {
	dir, err := cleanRel(c.QueryParam("path"))
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(h.root)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "workspace root is not available")
	}
	defer root.Close()
	out, err := readEntries(root.FS(), dir, isVisibleFolder)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

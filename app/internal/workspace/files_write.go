package workspace

import (
	"crypto/rand"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path"

	"github.com/labstack/echo/v5"
)

type entryInput struct {
	Path string `json:"path"`
	Dir  bool   `json:"dir"`
}

type fileInput struct {
	Content string `json:"content"`
}

var errEntryExists = echo.NewHTTPError(http.StatusConflict, "a file or folder with this name already exists")

// createTreeEntry creates an empty file or a folder inside the project (AIP-133 Create).
func (h *Handler) createTreeEntry(c *echo.Context) error {
	var in entryInput
	if err := c.Bind(&in); err != nil {
		return badRequest("invalid body")
	}
	name, err := cleanRel(in.Path)
	if err != nil {
		return err
	}
	if name == "." || hiddenDirs[path.Base(name)] {
		return badRequest("invalid path")
	}
	root, err := h.projectRoot(c)
	if err != nil {
		return err
	}
	defer root.Close()
	entry, err := createEntry(root, name, in.Dir)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, entry)
}

func createEntry(root *os.Root, name string, dir bool) (Entry, error) {
	entry := Entry{Name: path.Base(name), Path: name, Dir: dir}
	if _, err := root.Lstat(name); err == nil {
		return entry, errEntryExists
	}
	if parent := path.Dir(name); parent != "." {
		if err := ensureProjectFolder(root, parent); err != nil {
			return entry, badRequest("cannot create the parent folder: " + err.Error())
		}
	}
	if dir {
		return entry, mapCreateError(root.Mkdir(name, 0o755))
	}
	f, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return entry, mapCreateError(err)
	}
	return entry, f.Close()
}

func mapCreateError(err error) error {
	if errors.Is(err, fs.ErrExist) {
		return errEntryExists
	}
	if err != nil {
		return badRequest(err.Error())
	}
	return nil
}

// updateFile replaces the text of an existing file (AIP-134 Update; content is its only mutable field).
func (h *Handler) updateFile(c *echo.Context) error {
	name, err := cleanRel(c.QueryParam("path"))
	if err != nil {
		return err
	}
	req := c.Request()
	req.Body = http.MaxBytesReader(c.Response(), req.Body, maxFileBytes*2)
	var in fileInput
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "file is too large to edit here")
	}
	if len(in.Content) > maxFileBytes {
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "file is too large to edit here")
	}
	root, err := h.projectRoot(c)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := replaceFile(root, name, []byte(in.Content)); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, File{Path: name, Size: int64(len(in.Content)), Content: in.Content})
}

func replaceFile(root *os.Root, name string, content []byte) error {
	info, err := root.Stat(name)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	if !info.Mode().IsRegular() {
		return badRequest("not a file")
	}
	temp := path.Join(path.Dir(name), "."+path.Base(name)+".drift-"+rand.Text())
	if err := root.WriteFile(temp, content, info.Mode().Perm()); err != nil {
		return err
	}
	if err := root.Rename(temp, name); err != nil {
		return errors.Join(err, root.Remove(temp))
	}
	return nil
}

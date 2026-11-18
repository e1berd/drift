package workspace

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v5"
)

const (
	maxTreeEntries = 2000
	maxFileBytes   = 1 << 20
)

var hiddenDirs = map[string]bool{".git": true}

// Entry is one item of a directory listing.
type Entry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Dir  bool   `json:"dir"`
	Size int64  `json:"size,omitempty"`
}

type treeResponse struct {
	Path      string  `json:"path"`
	Entries   []Entry `json:"entries"`
	Truncated bool    `json:"truncated,omitempty"`
}

// File is the content of a text file, or a marker that it is binary or too large.
type File struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Content   string `json:"content"`
	Binary    bool   `json:"binary,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

func (h *Handler) projectRoot(c *echo.Context) (*os.Root, error) {
	p, err := h.project(c)
	if err != nil {
		return nil, err
	}
	return openProjectRoot(h.root, p.Path)
}

func openProjectRoot(workspaceRoot, projectPath string) (*os.Root, error) {
	root, err := os.OpenRoot(workspaceRoot)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusServiceUnavailable, "workspace root is not available")
	}
	defer root.Close()
	sub, err := root.OpenRoot(projectPath)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "project folder is missing")
	}
	return sub, nil
}

func cleanRel(raw string) (string, error) {
	p := path.Clean("/" + strings.ReplaceAll(raw, "\\", "/"))
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		p = "."
	}
	if !fs.ValidPath(p) {
		return "", badRequest("invalid path")
	}
	return p, nil
}

func (h *Handler) listTree(c *echo.Context) error {
	root, err := h.projectRoot(c)
	if err != nil {
		return err
	}
	defer root.Close()
	fsys := root.FS()
	dir, err := cleanRel(c.QueryParam("path"))
	if err != nil {
		return err
	}
	out, err := readEntries(fsys, dir, func(fs.DirEntry) bool { return true })
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func readEntries(fsys fs.FS, dir string, keep func(fs.DirEntry) bool) (treeResponse, error) {
	items, err := fs.ReadDir(fsys, dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return treeResponse{}, echo.NewHTTPError(http.StatusNotFound, "not found")
		}
		return treeResponse{}, badRequest("not a folder")
	}
	out := treeResponse{Path: dir, Entries: make([]Entry, 0, min(len(items), maxTreeEntries))}
	for _, it := range items {
		if it.IsDir() && hiddenDirs[it.Name()] || !keep(it) {
			continue
		}
		if len(out.Entries) == maxTreeEntries {
			out.Truncated = true
			break
		}
		e := Entry{Name: it.Name(), Path: path.Join(dir, it.Name()), Dir: it.IsDir()}
		if !e.Dir {
			if info, err := it.Info(); err == nil {
				e.Size = info.Size()
			}
		}
		out.Entries = append(out.Entries, e)
	}
	sort.SliceStable(out.Entries, func(i, j int) bool {
		a, b := out.Entries[i], out.Entries[j]
		if a.Dir != b.Dir {
			return a.Dir
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})
	return out, nil
}

func (h *Handler) readFile(c *echo.Context) error {
	root, err := h.projectRoot(c)
	if err != nil {
		return err
	}
	defer root.Close()
	fsys := root.FS()
	name, err := cleanRel(c.QueryParam("path"))
	if err != nil {
		return err
	}
	f, err := fsys.Open(name)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		return badRequest("not a file")
	}
	buf, err := io.ReadAll(io.LimitReader(f, maxFileBytes))
	if err != nil {
		return err
	}
	out := File{Path: name, Size: info.Size(), Truncated: info.Size() > maxFileBytes}
	head := buf[:min(len(buf), 8000)]
	if bytes.IndexByte(head, 0) >= 0 || !utf8.Valid(trimPartialRune(head)) {
		out.Binary = true
	} else {
		out.Content = string(buf)
	}
	return c.JSON(http.StatusOK, out)
}

func trimPartialRune(b []byte) []byte {
	for i := 0; i < utf8.UTFMax && len(b) > 0; i++ {
		r, size := utf8.DecodeLastRune(b)
		if r != utf8.RuneError || size > 1 {
			break
		}
		b = b[:len(b)-1]
	}
	return b
}

// Package uploads receives files over the tus resumable upload protocol (https://tus.io) and
// keeps them on disk as message attachments: partial uploads survive dropped connections and
// page reloads, finished ones become attachments that a message or session then claims.
package uploads

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	maxDiskNameBytes = 200
	fallbackName     = "file"
)

// Store is the folder that holds attachments: partial/ for uploads in progress and files/ for
// finished ones, each under files/<attachment id>/<original name>.
type Store struct {
	root string
}

// NewStore creates the folders under root when they are missing.
func NewStore(root string) (Store, error) {
	store := Store{root: root}
	for _, dir := range []string{store.partialDir(), store.filesDir()} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return Store{}, fmt.Errorf("create uploads folder: %w", err)
		}
	}
	return store, nil
}

// Path is the absolute path of a stored attachment given its path relative to files/.
func (s Store) Path(rel string) string {
	return filepath.Join(s.filesDir(), filepath.Clean("/"+rel))
}

// Open opens a stored attachment, refusing paths that would leave files/.
func (s Store) Open(rel string) (*os.File, error) {
	return os.OpenInRoot(s.filesDir(), rel)
}

func (s Store) partialDir() string {
	return filepath.Join(s.root, "partial")
}

func (s Store) filesDir() string {
	return filepath.Join(s.root, "files")
}

func relativePath(id, name string) string {
	return id + "/" + diskName(name)
}

func diskName(name string) string {
	name = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r < 0x20 || r == 0x7f {
			return '_'
		}
		return r
	}, strings.TrimSpace(name))
	if name == "" || name == "." || name == ".." {
		return fallbackName
	}
	if len(name) <= maxDiskNameBytes {
		return name
	}
	ext := filepath.Ext(name)
	if len(ext) > maxDiskNameBytes/4 {
		ext = ""
	}
	return truncateBytes(strings.TrimSuffix(name, ext), maxDiskNameBytes-len(ext)) + ext
}

func truncateBytes(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	for limit > 0 && !utf8.RuneStart(s[limit]) {
		limit--
	}
	return s[:limit]
}

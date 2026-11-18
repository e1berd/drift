//go:build !dev

package frontend

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"strings"
)

//go:embed all:dist
var dist embed.FS

type manifestChunk struct {
	File    string   `json:"file"`
	CSS     []string `json:"css"`
	Imports []string `json:"imports"`
}

type embedded struct {
	files    fs.FS
	manifest map[string]manifestChunk
	index    []byte
	static   http.Handler
}

// New returns the frontend embedded into the binary. The dev server URL is ignored.
func New(*url.URL) (Frontend, error) {
	files, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, err
	}
	raw, err := fs.ReadFile(files, ".vite/manifest.json")
	if err != nil {
		return nil, fmt.Errorf("frontend: web build missing, run `task web:build`: %w", err)
	}
	e := &embedded{files: files, static: http.FileServerFS(files)}
	if err := json.Unmarshal(raw, &e.manifest); err != nil {
		return nil, fmt.Errorf("frontend: parse manifest: %w", err)
	}
	if e.index, err = fs.ReadFile(files, "index.html"); err != nil {
		return nil, fmt.Errorf("frontend: read index.html: %w", err)
	}
	return e, nil
}

func (e *embedded) EntryTags(entry string) (template.HTML, error) {
	chunk, ok := e.manifest[entry]
	if !ok {
		return "", fmt.Errorf("frontend: entry %q not in manifest", entry)
	}

	var b strings.Builder
	seen := map[string]bool{}
	var walk func(key string)
	walk = func(key string) {
		if seen[key] {
			return
		}
		seen[key] = true
		c := e.manifest[key]
		for _, css := range c.CSS {
			fmt.Fprintf(&b, `<link rel="stylesheet" href="/%s">`, template.HTMLEscapeString(css))
		}
		for _, imp := range c.Imports {
			if !seen[imp] {
				fmt.Fprintf(&b, `<link rel="modulepreload" href="/%s">`, template.HTMLEscapeString(e.manifest[imp].File))
			}
			walk(imp)
		}
	}
	walk(entry)
	fmt.Fprintf(&b, `<script type="module" src="/%s"></script>`, template.HTMLEscapeString(chunk.File))
	return template.HTML(b.String()), nil
}

func (e *embedded) Favicon() string {
	return "/favicon-production.svg"
}

func (e *embedded) IsPage(r *http.Request) bool {
	return e.asset(r.URL.Path) == ""
}

func (e *embedded) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := e.asset(r.URL.Path)
	if name == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(e.index)
		return
	}
	if strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	e.static.ServeHTTP(w, r)
}

func (e *embedded) asset(urlPath string) string {
	name := strings.TrimPrefix(path.Clean("/"+urlPath), "/")
	if name == "" || name == "index.html" || strings.HasPrefix(name, ".") {
		return ""
	}
	info, err := fs.Stat(e.files, name)
	if err != nil || info.IsDir() {
		return ""
	}
	return name
}

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

// New returns the interfaces embedded into the binary. The dev server settings are ignored.
func New(*url.URL, string) (Frontend, error) {
	files, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, err
	}
	return load(files)
}

type embedded struct {
	apps    map[string]*build
	catalog []Info
}

func load(root fs.FS) (*embedded, error) {
	entries, err := fs.ReadDir(root, ".")
	if err != nil {
		return nil, fmt.Errorf("frontend: web build missing, run `task web:build`: %w", err)
	}
	e := &embedded{apps: map[string]*build{}}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		files, err := fs.Sub(root, entry.Name())
		if err != nil {
			return nil, err
		}
		app, err := loadBuild(entry.Name(), files)
		if err != nil {
			return nil, err
		}
		if app == nil {
			continue
		}
		e.apps[app.info.ID] = app
		e.catalog = append(e.catalog, app.info)
	}
	if e.apps[DefaultID] == nil {
		return nil, fmt.Errorf("frontend: no %q interface in the web build, run `task web:build`", DefaultID)
	}
	e.catalog = sortDefaultFirst(e.catalog)
	return e, nil
}

func (e *embedded) EntryTags(entry string) (template.HTML, error) {
	return e.apps[DefaultID].entryTags(entry)
}

func (e *embedded) Favicon() string {
	return basePath(DefaultID) + "favicon-production.svg"
}

func (e *embedded) IsPage(r *http.Request) bool {
	if _, _, ok := underAssets(r.URL.Path); ok {
		return false
	}
	return e.apps[DefaultID].asset(r.URL.Path) == ""
}

func (e *embedded) Catalog() []Info {
	return e.catalog
}

func (e *embedded) Has(id string) bool {
	return e.apps[id] != nil
}

func (e *embedded) Handler(defaultID DefaultResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, rest, ok := underAssets(r.URL.Path); ok {
			e.serveAsset(w, r, id, rest)
			return
		}
		app := e.apps[Chosen(r, e.Has, defaultID)]
		if name := app.asset(r.URL.Path); name != "" {
			app.serveFile(w, r, name)
			return
		}
		app.serveIndex(w)
	})
}

func (e *embedded) serveAsset(w http.ResponseWriter, r *http.Request, id, rest string) {
	app := e.apps[id]
	if app == nil {
		http.NotFound(w, r)
		return
	}
	name := app.asset(rest)
	if name == "" {
		http.NotFound(w, r)
		return
	}
	app.serveFile(w, r, name)
}

type manifestChunk struct {
	File    string   `json:"file"`
	CSS     []string `json:"css"`
	Imports []string `json:"imports"`
}

type build struct {
	info     Info
	files    fs.FS
	manifest map[string]manifestChunk
	index    []byte
}

func loadBuild(id string, files fs.FS) (*build, error) {
	rawInfo, err := fs.ReadFile(files, ManifestName)
	if err != nil {
		return nil, nil
	}
	info, err := parseInfo(id, rawInfo)
	if err != nil {
		return nil, err
	}
	b := &build{info: info, files: files}
	rawManifest, err := fs.ReadFile(files, ".vite/manifest.json")
	if err != nil {
		return nil, fmt.Errorf("frontend: %q has no Vite manifest: %w", id, err)
	}
	if err := json.Unmarshal(rawManifest, &b.manifest); err != nil {
		return nil, fmt.Errorf("frontend: parse manifest of %q: %w", id, err)
	}
	if b.index, err = fs.ReadFile(files, "index.html"); err != nil {
		return nil, fmt.Errorf("frontend: read index.html of %q: %w", id, err)
	}
	return b, nil
}

func (b *build) entryTags(entry string) (template.HTML, error) {
	chunk, ok := b.manifest[entry]
	if !ok {
		return "", fmt.Errorf("frontend: entry %q not in manifest of %q", entry, b.info.ID)
	}
	base := basePath(b.info.ID)
	var tags strings.Builder
	seen := map[string]bool{}
	var walk func(key string)
	walk = func(key string) {
		if seen[key] {
			return
		}
		seen[key] = true
		c := b.manifest[key]
		for _, css := range c.CSS {
			fmt.Fprintf(&tags, `<link rel="stylesheet" href="%s">`, template.HTMLEscapeString(base+css))
		}
		for _, imp := range c.Imports {
			if !seen[imp] {
				fmt.Fprintf(&tags, `<link rel="modulepreload" href="%s">`, template.HTMLEscapeString(base+b.manifest[imp].File))
			}
			walk(imp)
		}
	}
	walk(entry)
	fmt.Fprintf(&tags, `<script type="module" src="%s"></script>`, template.HTMLEscapeString(base+chunk.File))
	return template.HTML(tags.String()), nil
}

func (b *build) asset(urlPath string) string {
	name := strings.TrimPrefix(path.Clean("/"+urlPath), "/")
	if name == "" || name == "index.html" || name == ManifestName || strings.HasPrefix(name, ".") {
		return ""
	}
	info, err := fs.Stat(b.files, name)
	if err != nil || info.IsDir() {
		return ""
	}
	return name
}

func (b *build) serveFile(w http.ResponseWriter, r *http.Request, name string) {
	if strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	http.ServeFileFS(w, r, b.files, name)
}

func (b *build) serveIndex(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(b.index)
}

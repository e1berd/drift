// Package frontend serves the Drift interfaces built from ../web/apps: each one is a Vite build
// under dist/<id>/ with its assets at /ui/<id>/. A page request gets the interface the visitor
// picked (cookie "ui") or the instance default. Release builds embed ./dist; builds with the
// `dev` tag proxy one interface to the Vite dev server instead.
package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

const (
	// AuthEntry is the Vite entry, in the default interface, loaded by the server-rendered auth pages.
	AuthEntry = "src/auth/main.ts"
	// DefaultID is the interface that carries the auth pages and serves when nothing else is chosen.
	DefaultID = "default"
	// CookieName holds the interface the visitor picked; every interface writes the same cookie.
	CookieName = "ui"
	// ManifestName is the file in an interface's public folder that describes it.
	ManifestName = "drift-frontend.json"
	assetsPrefix = "/ui/"
)

// Info describes one interface; names and descriptions are keyed by language.
type Info struct {
	ID          string            `json:"-"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
}

// DefaultResolver returns the instance's default interface id.
type DefaultResolver func(ctx context.Context) string

// Frontend serves the interfaces and renders tags for the auth entry.
type Frontend interface {
	// EntryTags returns the <script>/<link> tags that load a Vite entry of the default interface.
	EntryTags(entry string) (template.HTML, error)
	// Favicon returns the tab icon path; development and release builds have different icons.
	Favicon() string
	// IsPage reports whether the request navigates to an SPA page rather than fetching an asset.
	IsPage(r *http.Request) bool
	// Catalog lists the interfaces, the default one first.
	Catalog() []Info
	// Has reports whether an interface with this id exists.
	Has(id string) bool
	// Handler serves assets and the SPA shell of the chosen interface.
	Handler(defaultID DefaultResolver) http.Handler
}

func parseInfo(id string, raw []byte) (Info, error) {
	var info Info
	if err := json.Unmarshal(raw, &info); err != nil {
		return Info{}, fmt.Errorf("frontend: parse %s of %q: %w", ManifestName, id, err)
	}
	info.ID = id
	return info, nil
}

func sortDefaultFirst(infos []Info) []Info {
	for i, info := range infos {
		if info.ID == DefaultID && i > 0 {
			infos[0], infos[i] = infos[i], infos[0]
		}
	}
	return infos
}

func basePath(id string) string {
	return assetsPrefix + id + "/"
}

func underAssets(urlPath string) (id, rest string, ok bool) {
	trimmed, found := strings.CutPrefix(urlPath, assetsPrefix)
	if !found {
		return "", "", false
	}
	id, rest, _ = strings.Cut(trimmed, "/")
	return id, rest, id != ""
}

// Chosen picks the interface for a page request: the visitor's cookie, else the instance default.
func Chosen(r *http.Request, has func(string) bool, defaultID DefaultResolver) string {
	if c, err := r.Cookie(CookieName); err == nil && has(c.Value) {
		return c.Value
	}
	if id := defaultID(r.Context()); has(id) {
		return id
	}
	return DefaultID
}

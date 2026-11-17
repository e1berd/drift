// Package frontend serves the Vite build of ../web: the Vue SPA and the asset bundle of the
// server-rendered auth pages. Release builds embed web output from ./dist; builds with the `dev`
// tag proxy everything to the Vite dev server instead.
package frontend

import (
	"html/template"
	"net/http"
)

// AuthEntry is the Vite entry loaded by the server-rendered auth pages.
const AuthEntry = "src/auth/main.ts"

// Frontend serves the web build and renders tags for its entries.
type Frontend interface {
	// EntryTags returns the <script>/<link> tags that load a Vite entry.
	EntryTags(entry string) (template.HTML, error)
	// IsPage reports whether the request navigates to an SPA page rather than fetching an asset.
	IsPage(r *http.Request) bool
	// ServeHTTP serves assets and the SPA shell.
	http.Handler
}

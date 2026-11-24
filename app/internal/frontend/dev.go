//go:build dev

package frontend

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const devAppsDir = "../web/apps"

type devServer struct {
	proxy   *httputil.ReverseProxy
	id      string
	catalog []Info
}

// New returns a frontend that proxies the interface `id` to its Vite dev server, including the
// HMR websocket. The catalog lists every interface in ../web/apps, but only `id` is served.
func New(viteURL *url.URL, id string) (Frontend, error) {
	if viteURL == nil {
		return nil, fmt.Errorf("frontend: VITE_DEV_URL is required in dev builds")
	}
	catalog, err := devCatalog(devAppsDir)
	if err != nil {
		return nil, err
	}
	d := &devServer{id: id, catalog: catalog}
	if !d.Has(id) {
		return nil, fmt.Errorf("frontend: no interface %q in %s", id, devAppsDir)
	}
	d.proxy = &httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
		r.SetURL(viteURL)
		r.SetXForwarded()
	}}
	return d, nil
}

func devCatalog(dir string) ([]Info, error) {
	manifests, err := filepath.Glob(filepath.Join(dir, "*", "public", ManifestName))
	if err != nil {
		return nil, err
	}
	var catalog []Info
	for _, manifest := range manifests {
		raw, err := os.ReadFile(manifest)
		if err != nil {
			return nil, err
		}
		info, err := parseInfo(filepath.Base(filepath.Dir(filepath.Dir(manifest))), raw)
		if err != nil {
			return nil, err
		}
		catalog = append(catalog, info)
	}
	return sortDefaultFirst(catalog), nil
}

func (d *devServer) EntryTags(entry string) (template.HTML, error) {
	base := basePath(d.id)
	return template.HTML(fmt.Sprintf(
		`<script type="module" src="%s@vite/client"></script><script type="module" src="%s%s"></script>`,
		template.HTMLEscapeString(base), template.HTMLEscapeString(base), template.HTMLEscapeString(entry),
	)), nil
}

func (d *devServer) Favicon() string {
	return basePath(d.id) + "favicon-development.svg"
}

func (d *devServer) IsPage(r *http.Request) bool {
	if _, _, ok := underAssets(r.URL.Path); ok {
		return false
	}
	if dest := r.Header.Get("Sec-Fetch-Dest"); dest != "" {
		return dest == "document"
	}
	return r.Method == http.MethodGet && strings.Contains(r.Header.Get("Accept"), "text/html")
}

func (d *devServer) Catalog() []Info {
	return d.catalog
}

func (d *devServer) Has(id string) bool {
	for _, info := range d.catalog {
		if info.ID == id {
			return true
		}
	}
	return false
}

func (d *devServer) Handler(DefaultResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := underAssets(r.URL.Path); !ok {
			r.URL.Path = basePath(d.id)
			r.URL.RawPath = ""
		}
		d.proxy.ServeHTTP(w, r)
	})
}

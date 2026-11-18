//go:build dev

package frontend

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type devServer struct {
	proxy *httputil.ReverseProxy
}

// New returns a frontend that proxies to the Vite dev server, including the HMR websocket.
func New(viteURL *url.URL) (Frontend, error) {
	if viteURL == nil {
		return nil, fmt.Errorf("frontend: VITE_DEV_URL is required in dev builds")
	}
	return &devServer{proxy: &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(viteURL)
			r.SetXForwarded()
		},
	}}, nil
}

func (d *devServer) EntryTags(entry string) (template.HTML, error) {
	return template.HTML(fmt.Sprintf(
		`<script type="module" src="/@vite/client"></script><script type="module" src="/%s"></script>`,
		template.HTMLEscapeString(entry),
	)), nil
}

func (d *devServer) Favicon() string {
	return "/favicon-development.svg"
}

func (d *devServer) IsPage(r *http.Request) bool {
	if dest := r.Header.Get("Sec-Fetch-Dest"); dest != "" {
		return dest == "document"
	}
	return r.Method == http.MethodGet && strings.Contains(r.Header.Get("Accept"), "text/html")
}

func (d *devServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.proxy.ServeHTTP(w, r)
}

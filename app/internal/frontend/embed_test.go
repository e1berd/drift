//go:build !dev

package frontend

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func interfaceFiles(id, page string) fstest.MapFS {
	return fstest.MapFS{
		id + "/" + ManifestName:     {Data: []byte(`{"name":{"en":"` + id + `","ru":"` + id + ` ru"},"description":{"en":"about ` + id + `"}}`)},
		id + "/.vite/manifest.json": {Data: []byte(`{"src/auth/main.ts":{"file":"assets/auth.js","css":["assets/auth.css"]}}`)},
		id + "/index.html":          {Data: []byte(page)},
		id + "/favicon.ico":         {Data: []byte("icon of " + id)},
		id + "/assets/app.js":       {Data: []byte("script of " + id)},
	}
}

func testBuilds(t *testing.T) *embedded {
	t.Helper()
	files := interfaceFiles("vuetify", "vuetify page")
	for name, file := range interfaceFiles(DefaultID, "default page") {
		files[name] = file
	}
	files["stray/readme.txt"] = &fstest.MapFile{Data: []byte("not an interface")}
	e, err := load(files)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func fixedDefault(id string) DefaultResolver {
	return func(context.Context) string { return id }
}

func get(t *testing.T, handler http.Handler, target string, cookie string) (int, string, http.Header) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	if cookie != "" {
		request.AddCookie(&http.Cookie{Name: CookieName, Value: cookie})
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	body, _ := io.ReadAll(recorder.Result().Body)
	return recorder.Code, string(body), recorder.Header()
}

func TestCatalogListsInterfacesDefaultFirst(t *testing.T) {
	catalog := testBuilds(t).Catalog()
	if len(catalog) != 2 || catalog[0].ID != DefaultID || catalog[1].ID != "vuetify" {
		t.Fatalf("catalog %+v", catalog)
	}
	if catalog[1].Name["ru"] != "vuetify ru" {
		t.Errorf("name %v", catalog[1].Name)
	}
}

func TestPagesFollowCookieThenInstanceDefault(t *testing.T) {
	e := testBuilds(t)
	for _, tt := range []struct {
		name, cookie, instance, want string
	}{
		{"instance default", "", "vuetify", "vuetify page"},
		{"cookie wins", "default", "vuetify", "default page"},
		{"unknown cookie", "react", "vuetify", "vuetify page"},
		{"unknown instance default", "", "react", "default page"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			code, body, header := get(t, e.Handler(fixedDefault(tt.instance)), "/s/ses_1", tt.cookie)
			if code != http.StatusOK || body != tt.want {
				t.Errorf("got %d %q, want %q", code, body, tt.want)
			}
			if header.Get("Cache-Control") != "no-cache" {
				t.Errorf("cache %q", header.Get("Cache-Control"))
			}
		})
	}
}

func TestAssetsComeFromTheirInterface(t *testing.T) {
	handler := testBuilds(t).Handler(fixedDefault(DefaultID))
	code, body, header := get(t, handler, "/ui/vuetify/assets/app.js", "default")
	if code != http.StatusOK || body != "script of vuetify" {
		t.Errorf("asset %d %q", code, body)
	}
	if !strings.Contains(header.Get("Cache-Control"), "immutable") {
		t.Errorf("cache %q", header.Get("Cache-Control"))
	}
	for _, target := range []string{"/ui/react/assets/app.js", "/ui/vuetify/missing.js", "/ui/vuetify/" + ManifestName} {
		if code, _, _ := get(t, handler, target, ""); code != http.StatusNotFound {
			t.Errorf("%s: %d", target, code)
		}
	}
	if _, body, _ := get(t, handler, "/favicon.ico", "vuetify"); body != "icon of vuetify" {
		t.Errorf("root file %q", body)
	}
}

func TestAuthEntryLoadsFromDefaultInterface(t *testing.T) {
	e := testBuilds(t)
	tags, err := e.EntryTags(AuthEntry)
	if err != nil {
		t.Fatal(err)
	}
	want := `<link rel="stylesheet" href="/ui/default/assets/auth.css"><script type="module" src="/ui/default/assets/auth.js"></script>`
	if string(tags) != want {
		t.Errorf("tags %s", tags)
	}
	if e.IsPage(httptest.NewRequest(http.MethodGet, "/ui/default/assets/auth.js", nil)) {
		t.Error("asset taken for a page")
	}
	if !e.IsPage(httptest.NewRequest(http.MethodGet, "/settings", nil)) {
		t.Error("page taken for an asset")
	}
}

func TestLoadRequiresTheDefaultInterface(t *testing.T) {
	if _, err := load(interfaceFiles("vuetify", "page")); err == nil {
		t.Error("loaded a build without the default interface")
	}
}

func TestResourceNames(t *testing.T) {
	if IDFromResourceName(ResourceName("vuetify")) != "vuetify" {
		t.Error("round trip")
	}
	for _, name := range []string{"vuetify", "frontends/", "frontends/a/b", "settings/x"} {
		if IDFromResourceName(name) != "" {
			t.Errorf("%q accepted", name)
		}
	}
}

package i18n

import (
	"net/http/httptest"
	"slices"
	"testing"
)

func TestLocalesHaveSameKeys(t *testing.T) {
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	base := b.messages[Default]
	for _, lang := range Supported {
		for key := range base {
			if _, ok := b.messages[lang][key]; !ok {
				t.Errorf("%s.yml: missing %q", lang, key)
			}
		}
		for key := range b.messages[lang] {
			if _, ok := base[key]; !ok {
				t.Errorf("%s.yml: %q is not in %s.yml", lang, key, Default)
			}
		}
	}
}

func TestDetect(t *testing.T) {
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ cookie, accept, want string }{
		{"", "", Default},
		{"en", "ru-RU,ru;q=0.9", "en"},
		{"de", "en-US,en;q=0.9", "en"},
		{"", "de-DE,ru;q=0.5", "ru"},
		{"", "fr", Default},
	}
	for _, c := range cases {
		r := httptest.NewRequest("GET", "/login", nil)
		if c.cookie != "" {
			r.Header.Set("Cookie", CookieName+"="+c.cookie)
		}
		if c.accept != "" {
			r.Header.Set("Accept-Language", c.accept)
		}
		if got := b.Detect(r); got != c.want {
			t.Errorf("Detect(cookie=%q, accept=%q) = %q, want %q", c.cookie, c.accept, got, c.want)
		}
	}
}

func TestPrinter(t *testing.T) {
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	p := b.Printer("en")
	if got := p.T("setup.password_hint", "min", 8); got != "At least 8 characters." {
		t.Errorf("T = %q", got)
	}
	if got := p.T("no.such.key"); got != "no.such.key" {
		t.Errorf("missing key = %q", got)
	}
	if b.Printer("xx").Lang != Default || !slices.Contains(Supported, Default) {
		t.Error("unsupported language must fall back to Default")
	}
}

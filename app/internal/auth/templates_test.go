package auth

import (
	"bytes"
	"strings"
	"testing"

	"github.com/e1berd/drift/app/internal/i18n"
)

func TestPagesRenderInEveryLanguage(t *testing.T) {
	p, err := parsePages()
	if err != nil {
		t.Fatal(err)
	}
	tr, err := i18n.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, lang := range i18n.Supported {
		layout := layoutData{L: tr.Printer(lang), Title: "t", Theme: "dark"}
		renders := map[string]func(*bytes.Buffer) error{
			"setup": func(b *bytes.Buffer) error {
				return p.setup.ExecuteTemplate(b, "layout", setupData{layoutData: layout})
			},
			"setup-form": func(b *bytes.Buffer) error {
				return p.setup.ExecuteTemplate(b, "setup-form", setupData{layoutData: layout})
			},
			"login": func(b *bytes.Buffer) error {
				return p.login.ExecuteTemplate(b, "layout", loginData{layoutData: layout, Password: true, Passkey: "{}"})
			},
			"error": func(b *bytes.Buffer) error {
				return p.login.ExecuteTemplate(b, "layout", loginData{layoutData: layout, Error: "boom"})
			},
		}
		for name, render := range renders {
			var b bytes.Buffer
			if err := render(&b); err != nil {
				t.Fatalf("%s/%s: %v", lang, name, err)
			}
			out := b.String()
			// An untranslated key renders as itself, e.g. "common.password".
			for _, prefix := range []string{"common.", "login.", "setup.", "error.", "theme.", "language.", "preferences."} {
				if strings.Contains(out, ">"+prefix) || strings.Contains(out, `"`+prefix) {
					t.Errorf("%s/%s: untranslated key with prefix %q", lang, name, prefix)
				}
			}
		}
	}
}

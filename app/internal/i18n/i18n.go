// Package i18n holds the server-side translations (locales/*.yml) and picks the language of a
// request: the `lang` cookie the language switcher sets, then Accept-Language, then Default.
package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"go.yaml.in/yaml/v3"
	"golang.org/x/text/language"
)

// CookieName is the cookie the language switcher writes; the SPA reads the same cookie.
const CookieName = "lang"

// Default is the language used when nothing else matches; its file is the fallback for missing keys.
const Default = "ru"

// Supported lists the languages in switcher order. Each one needs locales/<code>.yml.
var Supported = []string{"ru", "en"}

//go:embed locales/*.yml
var localeFS embed.FS

// Bundle maps a language to its flattened messages ("login.title" → "Вход").
type Bundle struct {
	messages map[string]map[string]string
	matcher  language.Matcher
}

// Load parses the embedded locale files.
func Load() (*Bundle, error) {
	b := &Bundle{messages: map[string]map[string]string{}}
	tags := make([]language.Tag, 0, len(Supported))
	for _, lang := range Supported {
		raw, err := fs.ReadFile(localeFS, path.Join("locales", lang+".yml"))
		if err != nil {
			return nil, fmt.Errorf("i18n: %w", err)
		}
		var tree map[string]any
		if err := yaml.Unmarshal(raw, &tree); err != nil {
			return nil, fmt.Errorf("i18n: parse %s.yml: %w", lang, err)
		}
		flat := map[string]string{}
		flatten("", tree, flat)
		b.messages[lang] = flat
		tags = append(tags, language.MustParse(lang))
	}
	b.matcher = language.NewMatcher(tags)
	return b, nil
}

func flatten(prefix string, node map[string]any, out map[string]string) {
	for k, v := range node {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch v := v.(type) {
		case map[string]any:
			flatten(key, v, out)
		default:
			out[key] = fmt.Sprint(v)
		}
	}
}

// IsSupported reports whether lang is one of Supported.
func IsSupported(lang string) bool {
	for _, l := range Supported {
		if l == lang {
			return true
		}
	}
	return false
}

// Detect picks the request language: a supported `lang` cookie, then Accept-Language, then Default.
func (b *Bundle) Detect(r *http.Request) string {
	if c, err := r.Cookie(CookieName); err == nil && IsSupported(c.Value) {
		return c.Value
	}
	if accept := r.Header.Get("Accept-Language"); accept != "" {
		tags, _, err := language.ParseAcceptLanguage(accept)
		if err == nil && len(tags) > 0 {
			_, idx, conf := b.matcher.Match(tags...)
			if conf != language.No {
				return Supported[idx]
			}
		}
	}
	return Default
}

// Printer translates into one language.
func (b *Bundle) Printer(lang string) *Printer {
	if !IsSupported(lang) {
		lang = Default
	}
	return &Printer{Lang: lang, bundle: b}
}

// Printer translates keys into Lang, falling back to Default and then to the key itself.
type Printer struct {
	Lang   string
	bundle *Bundle
}

// T returns the message for key with {name} placeholders filled from name/value pairs:
// T("setup.too_long", "max", 200).
func (p *Printer) T(key string, args ...any) string {
	msg, ok := p.Lookup(key)
	if !ok {
		return key
	}
	return Format(msg, args...)
}

// Lookup returns the raw message for key and whether it exists in Lang or Default.
func (p *Printer) Lookup(key string) (string, bool) {
	if msg, ok := p.bundle.messages[p.Lang][key]; ok {
		return msg, true
	}
	msg, ok := p.bundle.messages[Default][key]
	return msg, ok
}

// Format fills {name} placeholders in msg from name/value pairs.
func Format(msg string, args ...any) string {
	if len(args) < 2 {
		return msg
	}
	pairs := make([]string, 0, len(args))
	for i := 0; i+1 < len(args); i += 2 {
		pairs = append(pairs, "{"+fmt.Sprint(args[i])+"}", fmt.Sprint(args[i+1]))
	}
	return strings.NewReplacer(pairs...).Replace(msg)
}

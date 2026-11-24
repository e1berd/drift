package auth

import (
	"embed"
	"html/template"

	"github.com/e1berd/drift/app/internal/i18n"
)

//go:embed templates/*.html
var templateFS embed.FS

type pages struct {
	setup *template.Template
	login *template.Template
}

func parsePages() (*pages, error) {
	setup, err := template.ParseFS(templateFS, "templates/layout.html", "templates/setup.html")
	if err != nil {
		return nil, err
	}
	login, err := template.ParseFS(templateFS, "templates/layout.html", "templates/login.html")
	if err != nil {
		return nil, err
	}
	return &pages{setup: setup, login: login}, nil
}

// Color schemes of the theme switcher; the browser keeps the choice in colorSchemeCookie
// (createColorScheme + createCookieStorage in web/apps/default/src/shared/preferences.ts).
const colorSchemeCookie = "color-scheme"

var colorSchemes = []string{"system", "light", "dark"}

type layoutData struct {
	L       *i18n.Printer
	Title   string
	Lead    string
	Assets  template.HTML
	Favicon string
	Theme   string
}

// Languages lists the language switcher options.
func (layoutData) Languages() []string { return i18n.Supported }

type setupData struct {
	layoutData
	Name     string
	Email    string
	Messages []message
}

func (setupData) MinPasswordLength() int { return minPasswordLength }
func (setupData) MaxPasswordBytes() int  { return maxPasswordBytes }
func (setupData) MaxNameLength() int     { return maxNameLength }

type loginData struct {
	layoutData
	Action     string
	CSRF       string
	Identifier string
	Password   bool
	Passkey    string
	Refresh    bool
	Messages   []message
	Error      string
}

type message struct {
	Type string
	Text string
}

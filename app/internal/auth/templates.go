package auth

import (
	"embed"
	"html/template"
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

type layoutData struct {
	Title  string
	Lead   string
	Assets template.HTML
}

type setupData struct {
	layoutData
	Name     string
	Email    string
	Messages []message
}

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

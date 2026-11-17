// Package auth serves the two server-rendered pages that sit in front of the SPA: first-run
// superuser setup and Kratos login (password or passkey), plus the session guard for the SPA.
package auth

import (
	"bytes"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf8"

	"github.com/e1berd/drift/app/internal/frontend"
	"github.com/e1berd/drift/app/internal/kratos"
	"github.com/labstack/echo/v5"
)

// SuperuserRole is stored in metadata_public.role of the identity created on first run.
const SuperuserRole = "superuser"

const (
	minPasswordLength = 8
	maxPasswordBytes  = 72
)

// Handler serves /setup, /login, /error, /logout and the Kratos public API proxy.
type Handler struct {
	kratos    *kratos.Client
	frontend  frontend.Frontend
	publicURL *url.URL
	pages     *pages
	log       *slog.Logger

	setupMu  sync.Mutex
	hasUsers atomic.Bool
}

// New parses the templates and returns the handler.
func New(k *kratos.Client, fe frontend.Frontend, publicURL *url.URL, log *slog.Logger) (*Handler, error) {
	p, err := parsePages()
	if err != nil {
		return nil, err
	}
	return &Handler{kratos: k, frontend: fe, publicURL: publicURL, pages: p, log: log}, nil
}

// Register mounts the auth routes.
func (h *Handler) Register(e *echo.Echo) {
	e.Any(kratos.PublicPrefix+"/*", echo.WrapHandler(h.kratos.PublicProxy()))
	e.GET("/setup", h.setupPage)
	e.POST("/setup", h.setupSubmit)
	e.GET("/login", h.login)
	e.GET("/error", h.selfServiceError)
	e.GET("/logout", h.logout)
}

// RequireSession redirects page navigations without a Kratos session to /login (or /setup on first run).
func (h *Handler) RequireSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		r := c.Request()
		if !h.frontend.IsPage(r) {
			return next(c)
		}
		session, err := h.kratos.Session(r.Context(), r.Header.Get("Cookie"))
		if err != nil {
			return err
		}
		if session == nil {
			return c.Redirect(http.StatusSeeOther, "/login?return_to="+url.QueryEscape(r.URL.RequestURI()))
		}
		return next(c)
	}
}

func (h *Handler) usersExist(c *echo.Context) (bool, error) {
	if h.hasUsers.Load() {
		return true, nil
	}
	ok, err := h.kratos.HasIdentities(c.Request().Context())
	if err != nil {
		return false, err
	}
	if ok {
		h.hasUsers.Store(true)
	}
	return ok, nil
}

func (h *Handler) layout(title, lead string) (layoutData, error) {
	assets, err := h.frontend.EntryTags(frontend.AuthEntry)
	return layoutData{Title: title, Lead: lead, Assets: assets}, err
}

func render(c *echo.Context, status int, t *template.Template, name string, data any) error {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.HTMLBlob(status, buf.Bytes())
}

func (h *Handler) setupPage(c *echo.Context) error {
	exists, err := h.usersExist(c)
	if err != nil {
		return err
	}
	if exists {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	layout, err := h.layout("Первый запуск", "Создайте суперпользователя. После этого регистрация закрывается.")
	if err != nil {
		return err
	}
	return render(c, http.StatusOK, h.pages.setup, "layout", setupData{layoutData: layout})
}

func (h *Handler) setupSubmit(c *echo.Context) error {
	if c.Request().Header.Get("HX-Request") != "true" {
		return echo.NewHTTPError(http.StatusForbidden, "htmx request required")
	}

	h.setupMu.Lock()
	defer h.setupMu.Unlock()

	exists, err := h.usersExist(c)
	if err != nil {
		return err
	}
	if exists {
		c.Response().Header().Set("HX-Redirect", "/login")
		return c.NoContent(http.StatusOK)
	}

	form := setupData{
		Name:  strings.TrimSpace(c.FormValue("name")),
		Email: strings.TrimSpace(c.FormValue("email")),
	}
	password := c.FormValue("password")

	if msg := validateSetup(form.Name, form.Email, password, c.FormValue("password_confirm")); msg != "" {
		form.Messages = []message{{Type: "error", Text: msg}}
		return render(c, http.StatusUnprocessableEntity, h.pages.setup, "setup-form", form)
	}

	_, err = h.kratos.CreateIdentity(c.Request().Context(), kratos.CreateIdentityParams{
		Traits:         kratos.Traits{Email: form.Email, Name: form.Name},
		Password:       password,
		MetadataPublic: map[string]any{"role": SuperuserRole},
	})
	var apiErr *kratos.APIError
	if errors.As(err, &apiErr) && apiErr.Status < http.StatusInternalServerError {
		h.log.Warn("superuser setup rejected", "status", apiErr.Status, "reason", apiErr.Reason)
		form.Messages = []message{{Type: "error", Text: "Kratos отклонил данные: " + firstNonEmpty(apiErr.Reason, apiErr.Message)}}
		return render(c, http.StatusUnprocessableEntity, h.pages.setup, "setup-form", form)
	}
	if err != nil {
		return err
	}

	h.hasUsers.Store(true)
	h.log.Info("superuser created")
	c.Response().Header().Set("HX-Redirect", "/login")
	return c.NoContent(http.StatusOK)
}

func validateSetup(name, email, password, confirm string) string {
	switch {
	case utf8.RuneCountInString(name) > 200:
		return "Имя длиннее 200 символов"
	case email == "":
		return "Введите email"
	}
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		return "Неверный формат email"
	}
	switch {
	case utf8.RuneCountInString(password) < minPasswordLength:
		return "Пароль короче 8 символов"
	case len(password) > maxPasswordBytes:
		return "Пароль длиннее 72 байт"
	case password != confirm:
		return "Пароли не совпадают"
	}
	return ""
}

func (h *Handler) login(c *echo.Context) error {
	r := c.Request()
	exists, err := h.usersExist(c)
	if err != nil {
		return err
	}
	if !exists {
		return c.Redirect(http.StatusSeeOther, "/setup")
	}

	cookie := r.Header.Get("Cookie")
	flowID := c.QueryParam("flow")
	if flowID == "" {
		refresh := c.QueryParam("refresh") == "true"
		if !refresh {
			session, err := h.kratos.Session(r.Context(), cookie)
			if err != nil {
				return err
			}
			if session != nil {
				return c.Redirect(http.StatusSeeOther, safeReturnTo(c.QueryParam("return_to")))
			}
		}
		q := url.Values{}
		if refresh {
			q.Set("refresh", "true")
		}
		if rt := c.QueryParam("return_to"); rt != "" {
			q.Set("return_to", strings.TrimSuffix(h.publicURL.String(), "/")+safeReturnTo(rt))
		}
		return c.Redirect(http.StatusSeeOther, kratos.PublicPrefix+"/self-service/login/browser?"+q.Encode())
	}

	flow, err := h.kratos.LoginFlow(r.Context(), flowID, cookie)
	if errors.Is(err, kratos.ErrFlowExpired) {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	if err != nil {
		return err
	}

	title, lead := "Вход", ""
	if flow.Refresh {
		title, lead = "Подтвердите вход", "Для этого действия нужно войти ещё раз."
	}
	layout, err := h.layout(title, lead)
	if err != nil {
		return err
	}

	_, hasPassword := flow.UI.Node("password")
	data := loginData{
		layoutData: layout,
		Action:     flow.UI.Action,
		CSRF:       flow.UI.Value("csrf_token"),
		Identifier: flow.UI.Value("identifier"),
		Password:   hasPassword,
		Passkey:    flow.UI.Value("passkey_challenge"),
		Refresh:    flow.Refresh,
	}
	for _, m := range flow.UI.AllMessages() {
		data.Messages = append(data.Messages, message{Type: m.Type, Text: m.Text})
	}
	return render(c, http.StatusOK, h.pages.login, "layout", data)
}

func (h *Handler) selfServiceError(c *echo.Context) error {
	text := "Что-то пошло не так. Попробуйте войти ещё раз."
	if id := c.QueryParam("id"); id != "" {
		e, err := h.kratos.SelfServiceError(c.Request().Context(), id)
		if err != nil {
			h.log.Warn("fetch self-service error", "err", err)
		} else if msg := firstNonEmpty(e.Error.Reason, e.Error.Message); msg != "" {
			text = msg
		}
	}
	layout, err := h.layout("Ошибка входа", "")
	if err != nil {
		return err
	}
	return render(c, http.StatusOK, h.pages.login, "layout", loginData{layoutData: layout, Error: text})
}

func (h *Handler) logout(c *echo.Context) error {
	logoutURL, err := h.kratos.LogoutURL(c.Request().Context(), c.Request().Header.Get("Cookie"))
	var apiErr *kratos.APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, logoutURL)
}

func safeReturnTo(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || raw == "" || u.IsAbs() || u.Host != "" || !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(raw, "//") || strings.HasPrefix(raw, "/\\") {
		return "/"
	}
	return u.RequestURI()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

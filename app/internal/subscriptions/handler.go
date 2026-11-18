// Package subscriptions serves the agent subscriptions API: the accounts Drift signs in to
// through the official CLIs, which of them is current per agent, and their usage limits.
// Custom methods follow AIP-136 (`POST /subscriptions/{id}:activate`), creation AIP-133.
package subscriptions

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/e1berd/drift/app/internal/agents"
	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/events"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

const deviceHeader = "X-Drift-Device"

// Handler owns the subscriptions API and the sign-ins running in the background.
type Handler struct {
	pool *pgxpool.Pool
	q    *db.Queries
	hub  *events.Hub
	cli  agents.CLI
	log  *slog.Logger

	mu     sync.Mutex
	logins map[string]*pendingLogin
	usage  map[string]cachedUsage
}

// New returns the handler and marks sign-ins cut off by a previous shutdown as failed.
func New(ctx context.Context, pool *pgxpool.Pool, hub *events.Hub, cli agents.CLI, log *slog.Logger) (*Handler, error) {
	q := db.New(pool)
	if err := q.FailInterruptedLogins(ctx); err != nil {
		return nil, err
	}
	return &Handler{
		pool: pool, q: q, hub: hub, cli: cli, log: log,
		logins: make(map[string]*pendingLogin),
		usage:  make(map[string]cachedUsage),
	}, nil
}

// Register mounts the routes on a group that already requires a signed-in identity.
func (h *Handler) Register(g *echo.Group) {
	g.GET("/subscriptions", h.list)
	g.POST("/subscriptions", h.create)
	g.POST("/subscriptions/:subscription", h.customMethod)
	g.DELETE("/subscriptions/:subscription", h.remove)
	g.GET("/subscriptions/:subscription/usage", h.getUsage)
}

// LoginInfo is the sign-in step the user finishes in the browser.
type LoginInfo struct {
	URL       string `json:"url"`
	UserCode  string `json:"userCode,omitempty"`
	NeedsCode bool   `json:"needsCode"`
}

// Subscription is the API view of one signed-in (or signing-in) agent account.
type Subscription struct {
	ID         string     `json:"id"`
	Agent      string     `json:"agent"`
	Method     string     `json:"method"`
	State      string     `json:"state"`
	Account    string     `json:"account"`
	Plan       string     `json:"plan"`
	Error      string     `json:"error,omitempty"`
	Current    bool       `json:"current"`
	Login      *LoginInfo `json:"login,omitempty"`
	CreateTime time.Time  `json:"createTime"`
	UpdateTime time.Time  `json:"updateTime"`
}

func (h *Handler) dto(s db.Subscription) Subscription {
	out := Subscription{
		ID: uuidString(s.ID), Agent: s.Agent, Method: s.Method, State: s.State,
		Account: s.Account, Plan: s.Plan, Error: s.Error, Current: s.IsCurrent,
		CreateTime: s.CreatedAt.Time, UpdateTime: s.UpdatedAt.Time,
	}
	if s.State == "pending" {
		out.Login = h.loginInfo(out.ID)
	}
	return out
}

func (h *Handler) publish(c *echo.Context, typ string, s Subscription) {
	h.hub.Publish(identity.From(c), typ, s, c.Request().Header.Get(deviceHeader))
}

func (h *Handler) list(c *echo.Context) error {
	rows, err := h.q.ListSubscriptions(c.Request().Context(), identity.From(c))
	if err != nil {
		return err
	}
	out := make([]Subscription, 0, len(rows))
	for _, row := range rows {
		out = append(out, h.dto(row))
	}
	return c.JSON(http.StatusOK, map[string]any{"subscriptions": out})
}

type createInput struct {
	Agent string `json:"agent"`
	Token string `json:"token"`
}

func (h *Handler) create(c *echo.Context) error {
	var in createInput
	if err := c.Bind(&in); err != nil {
		return badRequest("invalid body")
	}
	driver, ok := h.cli.Driver(in.Agent)
	if !ok {
		return badRequest("signing in to " + in.Agent + " is not supported yet")
	}
	token := strings.TrimSpace(in.Token)
	if token != "" && (in.Agent != "claude" || !strings.HasPrefix(token, "sk-ant-oat")) {
		return badRequest("expected a token from `claude setup-token` (sk-ant-oat…)")
	}
	method := "login"
	if token != "" {
		method = "token"
	}
	ctx := c.Request().Context()
	row, err := h.q.CreateSubscription(ctx, db.CreateSubscriptionParams{OwnerID: identity.From(c), Agent: in.Agent, Method: method})
	if err != nil {
		return err
	}
	id := uuidString(row.ID)
	home, err := h.cli.Prepare(id)
	if err != nil {
		return h.discard(ctx, row, err)
	}
	if token != "" {
		row, err = h.activateWithToken(ctx, row, driver, home, token)
	} else {
		row, err = h.startLogin(row, driver, home)
	}
	if err != nil {
		return h.discard(ctx, row, err)
	}
	dto := h.dto(row)
	h.publish(c, "subscription.created", dto)
	return c.JSON(http.StatusCreated, dto)
}

func (h *Handler) activateWithToken(ctx context.Context, row db.Subscription, driver agents.Driver, home, token string) (db.Subscription, error) {
	if err := agents.CheckClaudeToken(ctx, http.DefaultClient, agents.ClaudeModelsURL, token); err != nil {
		if errors.Is(err, agents.ErrTokenRejected) {
			return row, badRequest(err.Error())
		}
		return row, err
	}
	if err := h.cli.SaveClaudeToken(home, token); err != nil {
		return row, err
	}
	account, err := driver.Account(ctx, home)
	if err != nil {
		return row, badRequest("Claude Code did not accept the token: " + err.Error())
	}
	return h.q.MarkSubscriptionActive(ctx, db.MarkSubscriptionActiveParams{
		ID: row.ID, OwnerID: row.OwnerID, Account: account.Email, Plan: account.Plan,
	})
}

func (h *Handler) discard(ctx context.Context, row db.Subscription, cause error) error {
	_, _ = h.q.DeleteSubscription(ctx, db.DeleteSubscriptionParams{ID: row.ID, OwnerID: row.OwnerID})
	if err := h.cli.Remove(uuidString(row.ID)); err != nil {
		h.log.Warn("remove subscription home", "err", err)
	}
	var httpErr *echo.HTTPError
	if errors.As(cause, &httpErr) {
		return cause
	}
	return echo.NewHTTPError(http.StatusBadGateway, cause.Error())
}

func (h *Handler) customMethod(c *echo.Context) error {
	raw, method, _ := strings.Cut(c.Param("subscription"), ":")
	row, err := h.subscription(c, raw)
	if err != nil {
		return err
	}
	switch method {
	case "activate":
		return h.activate(c, row)
	case "startLogin":
		return h.restartLogin(c, row)
	case "completeLogin":
		return h.completeLogin(c, row)
	}
	return echo.NewHTTPError(http.StatusNotFound, "unknown method")
}

func (h *Handler) activate(c *echo.Context, row db.Subscription) error {
	if row.State != "active" {
		return badRequest("only a signed-in subscription can be current")
	}
	ctx := c.Request().Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := h.q.WithTx(tx)
	if err := q.ClearCurrentSubscription(ctx, db.ClearCurrentSubscriptionParams{OwnerID: row.OwnerID, Agent: row.Agent}); err != nil {
		return err
	}
	row, err = q.SetCurrentSubscription(ctx, db.SetCurrentSubscriptionParams{ID: row.ID, OwnerID: row.OwnerID})
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	dto := h.dto(row)
	h.publish(c, "subscription.updated", dto)
	return c.JSON(http.StatusOK, dto)
}

func (h *Handler) remove(c *echo.Context) error {
	row, err := h.subscription(c, c.Param("subscription"))
	if err != nil {
		return err
	}
	h.cancelLogin(uuidString(row.ID))
	ctx := c.Request().Context()
	if _, err := h.q.DeleteSubscription(ctx, db.DeleteSubscriptionParams{ID: row.ID, OwnerID: row.OwnerID}); err != nil {
		return notFound(err)
	}
	if err := h.cli.Remove(uuidString(row.ID)); err != nil {
		h.log.Warn("remove subscription home", "err", err)
	}
	if row.IsCurrent {
		next, err := h.q.PromoteNextSubscription(ctx, db.PromoteNextSubscriptionParams{OwnerID: row.OwnerID, Agent: row.Agent})
		if err == nil {
			h.publish(c, "subscription.updated", h.dto(next))
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}
	h.publish(c, "subscription.deleted", h.dto(row))
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) subscription(c *echo.Context, raw string) (db.Subscription, error) {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		return db.Subscription{}, echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	row, err := h.q.GetSubscription(c.Request().Context(), db.GetSubscriptionParams{ID: id, OwnerID: identity.From(c)})
	return row, notFound(err)
}

func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	return err
}

func badRequest(msg string) error {
	return echo.NewHTTPError(http.StatusBadRequest, msg)
}

func uuidString(id pgtype.UUID) string {
	v, _ := id.Value()
	s, _ := v.(string)
	return s
}

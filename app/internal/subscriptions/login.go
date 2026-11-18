package subscriptions

import (
	"context"
	"net/http"
	"time"

	"github.com/e1berd/drift/app/internal/agents"
	"github.com/e1berd/drift/app/internal/db"
	"github.com/labstack/echo/v5"
)

const (
	maxRunningLogins = 4
	codeWaitTimeout  = time.Minute
)

type pendingLogin struct {
	login   *agents.Login
	settled chan struct{}
}

func (h *Handler) loginInfo(id string) *LoginInfo {
	h.mu.Lock()
	defer h.mu.Unlock()
	pending, ok := h.logins[id]
	if !ok {
		return nil
	}
	return &LoginInfo{URL: pending.login.URL, UserCode: pending.login.UserCode, NeedsCode: pending.login.NeedsCode}
}

func (h *Handler) cancelLogin(id string) {
	h.mu.Lock()
	pending, ok := h.logins[id]
	delete(h.logins, id)
	h.mu.Unlock()
	if ok {
		pending.login.Cancel()
	}
}

func (h *Handler) startLogin(row db.Subscription, driver agents.Driver, home string) (db.Subscription, error) {
	id := uuidString(row.ID)
	h.cancelLogin(id)
	h.mu.Lock()
	running := len(h.logins)
	h.mu.Unlock()
	if running >= maxRunningLogins {
		return row, echo.NewHTTPError(http.StatusTooManyRequests, "too many sign-ins in progress, finish or cancel one first")
	}
	login, err := driver.StartLogin(home)
	if err != nil {
		return row, err
	}
	pending := &pendingLogin{login: login, settled: make(chan struct{})}
	h.mu.Lock()
	h.logins[id] = pending
	h.mu.Unlock()
	go h.settle(row, driver, home, pending)
	return row, nil
}

func (h *Handler) settle(row db.Subscription, driver agents.Driver, home string, pending *pendingLogin) {
	defer close(pending.settled)
	<-pending.login.Finished()
	id := uuidString(row.ID)
	h.mu.Lock()
	current := h.logins[id] == pending
	if current {
		delete(h.logins, id)
	}
	h.mu.Unlock()
	if !current {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	result := pending.login.Result()
	var account agents.Account
	if result == nil {
		account, result = driver.Account(ctx, home)
	}
	var (
		updated db.Subscription
		err     error
	)
	if result == nil {
		updated, err = h.q.MarkSubscriptionActive(ctx, db.MarkSubscriptionActiveParams{
			ID: row.ID, OwnerID: row.OwnerID, Account: account.Email, Plan: account.Plan,
		})
	} else {
		updated, err = h.q.MarkSubscriptionFailed(ctx, db.MarkSubscriptionFailedParams{
			ID: row.ID, OwnerID: row.OwnerID, Error: result.Error(),
		})
	}
	if err != nil {
		h.log.Warn("store sign-in result", "subscription", id, "err", err)
		return
	}
	h.hub.Publish(row.OwnerID, "subscription.updated", h.dto(updated), "")
}

func (h *Handler) restartLogin(c *echo.Context, row db.Subscription) error {
	if row.Method != "login" {
		return badRequest("this subscription uses a token; remove it and add it again")
	}
	driver, ok := h.cli.Driver(row.Agent)
	if !ok {
		return badRequest("signing in to " + row.Agent + " is not supported yet")
	}
	home, err := h.cli.Prepare(uuidString(row.ID))
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if row, err = h.q.MarkSubscriptionPending(ctx, db.MarkSubscriptionPendingParams{ID: row.ID, OwnerID: row.OwnerID}); err != nil {
		return err
	}
	if _, err := h.startLogin(row, driver, home); err != nil {
		failed, _ := h.q.MarkSubscriptionFailed(ctx, db.MarkSubscriptionFailedParams{ID: row.ID, OwnerID: row.OwnerID, Error: err.Error()})
		h.publish(c, "subscription.updated", h.dto(failed))
		return err
	}
	dto := h.dto(row)
	h.publish(c, "subscription.updated", dto)
	return c.JSON(http.StatusOK, dto)
}

type codeInput struct {
	Code string `json:"code"`
}

func (h *Handler) completeLogin(c *echo.Context, row db.Subscription) error {
	var in codeInput
	if err := c.Bind(&in); err != nil || in.Code == "" {
		return badRequest("code is required")
	}
	id := uuidString(row.ID)
	h.mu.Lock()
	pending, ok := h.logins[id]
	h.mu.Unlock()
	if !ok {
		return echo.NewHTTPError(http.StatusConflict, "this sign-in is no longer running, start it again")
	}
	if err := pending.login.Submit(in.Code); err != nil {
		return err
	}
	select {
	case <-pending.settled:
	case <-time.After(codeWaitTimeout):
		return echo.NewHTTPError(http.StatusGatewayTimeout, "the agent did not confirm the code in time")
	case <-c.Request().Context().Done():
		return c.Request().Context().Err()
	}
	updated, err := h.q.GetSubscription(c.Request().Context(), db.GetSubscriptionParams{ID: row.ID, OwnerID: row.OwnerID})
	if err != nil {
		return notFound(err)
	}
	return c.JSON(http.StatusOK, h.dto(updated))
}

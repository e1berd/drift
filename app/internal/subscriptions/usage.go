package subscriptions

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/e1berd/drift/app/internal/agents"
	"github.com/labstack/echo/v5"
)

const usageTTL = time.Minute

// UsageWindow is one limit window of a subscription, e.g. five hours or a week.
type UsageWindow struct {
	ID            string     `json:"id"`
	Label         string     `json:"label,omitempty"`
	UsedPercent   float64    `json:"usedPercent"`
	WindowMinutes int        `json:"windowMinutes,omitempty"`
	ResetTime     *time.Time `json:"resetTime,omitempty"`
}

// Usage is the usage singleton of a subscription (AIP-156). Unavailable carries a reason code
// when the agent does not report limits for this kind of sign-in.
type Usage struct {
	Windows     []UsageWindow `json:"windows"`
	Unavailable string        `json:"unavailable,omitempty"`
	ReadTime    time.Time     `json:"readTime"`
}

type cachedUsage struct {
	usage Usage
	until time.Time
}

func (h *Handler) getUsage(c *echo.Context) error {
	row, err := h.subscription(c, c.Param("subscription"))
	if err != nil {
		return err
	}
	if row.State != "active" {
		return c.JSON(http.StatusOK, Usage{Windows: []UsageWindow{}, Unavailable: "signed_out", ReadTime: time.Now()})
	}
	id := uuidString(row.ID)
	refresh, _ := strconv.ParseBool(c.QueryParam("refresh"))
	if !refresh {
		h.mu.Lock()
		cached, ok := h.usage[id]
		h.mu.Unlock()
		if ok && time.Now().Before(cached.until) {
			return c.JSON(http.StatusOK, cached.usage)
		}
	}
	driver, ok := h.cli.Driver(row.Agent)
	if !ok {
		return c.JSON(http.StatusOK, Usage{Windows: []UsageWindow{}, Unavailable: "unsupported", ReadTime: time.Now()})
	}
	home := h.cli.Home(id)
	if _, err := os.Stat(home); err != nil {
		return c.JSON(http.StatusOK, Usage{Windows: []UsageWindow{}, Unavailable: "signed_out", ReadTime: time.Now()})
	}
	ctx := c.Request().Context()
	if refresh {
		ctx = agents.WithForcedUsage(ctx)
	}
	usage := h.readUsage(ctx, driver, home)
	h.cacheUsage(id, usage)
	return c.JSON(http.StatusOK, usage)
}

func (h *Handler) cacheUsage(id string, usage Usage) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if usage.Unavailable == "error" {
		delete(h.usage, id)
		return
	}
	h.usage[id] = cachedUsage{usage: usage, until: time.Now().Add(usageTTL)}
}

func (h *Handler) readUsage(ctx context.Context, driver agents.Driver, home string) Usage {
	usage := Usage{Windows: []UsageWindow{}, ReadTime: time.Now()}
	windows, err := driver.Usage(ctx, home)
	var unavailable agents.UsageUnavailable
	switch {
	case errors.As(err, &unavailable):
		usage.Unavailable = unavailable.Reason
	case err != nil:
		h.log.Warn("read usage", "err", err)
		usage.Unavailable = "error"
	}
	for _, window := range windows {
		usage.Windows = append(usage.Windows, UsageWindow(window))
	}
	return usage
}

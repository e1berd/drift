package agents

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	usageFreshFor      = 5 * time.Minute
	usageRetryGap      = time.Minute
	usageRateLimitWait = 10 * time.Minute
)

type forcedUsageKey struct{}

// WithForcedUsage lets a manual refresh skip a fresh reading, at most once per usageRetryGap.
func WithForcedUsage(ctx context.Context) context.Context {
	return context.WithValue(ctx, forcedUsageKey{}, true)
}

func usageForced(ctx context.Context) bool {
	forced, _ := ctx.Value(forcedUsageKey{}).(bool)
	return forced
}

type usageRateLimited struct {
	service    string
	retryAfter time.Duration
}

func (e usageRateLimited) Error() string {
	return fmt.Sprintf("%s usage: %d %s", e.service, http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
}

func rateLimitedResponse(service string, resp *http.Response) usageRateLimited {
	seconds, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
	return usageRateLimited{service: service, retryAfter: time.Duration(seconds) * time.Second}
}

type usageReading struct {
	mu           sync.Mutex
	windows      []Window
	fetched      time.Time
	attempted    time.Time
	lastErr      error
	blockedUntil time.Time
}

func (r *usageReading) fresh(now time.Time, forced bool) bool {
	if r.fetched.IsZero() {
		return false
	}
	if forced {
		return now.Sub(r.attempted) < usageRetryGap
	}
	return now.Sub(r.fetched) < usageFreshFor
}

func (r *usageReading) waiting(now time.Time) bool {
	return now.Before(r.blockedUntil) || (r.lastErr != nil && now.Sub(r.attempted) < usageRetryGap)
}

func (r *usageReading) lastKnown() ([]Window, error) {
	if r.windows != nil {
		return r.windows, nil
	}
	return nil, r.lastErr
}

func (r *usageReading) record(now time.Time, windows []Window, err error) ([]Window, error) {
	r.attempted = now
	var limited usageRateLimited
	var unavailable UsageUnavailable
	switch {
	case err == nil:
		r.windows, r.fetched, r.lastErr = windows, now, nil
		return windows, nil
	case errors.As(err, &unavailable):
		r.windows, r.fetched, r.lastErr = nil, time.Time{}, nil
		return nil, err
	case errors.As(err, &limited):
		r.blockedUntil = now.Add(max(limited.retryAfter, usageRateLimitWait))
	}
	r.lastErr = err
	return r.lastKnown()
}

type usageThrottle struct {
	mu       sync.Mutex
	readings map[string]*usageReading
	now      func() time.Time
}

func newUsageThrottle() *usageThrottle {
	return &usageThrottle{readings: map[string]*usageReading{}, now: time.Now}
}

func (t *usageThrottle) reading(key string) *usageReading {
	t.mu.Lock()
	defer t.mu.Unlock()
	reading, ok := t.readings[key]
	if !ok {
		reading = &usageReading{}
		t.readings[key] = reading
	}
	return reading
}

func (t *usageThrottle) read(ctx context.Context, key string, fetch func(context.Context) ([]Window, error)) ([]Window, error) {
	reading := t.reading(key)
	reading.mu.Lock()
	defer reading.mu.Unlock()
	now := t.now()
	if reading.fresh(now, usageForced(ctx)) {
		return reading.windows, nil
	}
	if reading.waiting(now) {
		return reading.lastKnown()
	}
	windows, err := fetch(ctx)
	return reading.record(now, windows, err)
}

var claudeUsage = newUsageThrottle()

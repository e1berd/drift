package runner

import (
	"context"
	"math"
	"slices"
	"time"

	"github.com/e1berd/drift/app/internal/agents"
	"github.com/e1berd/drift/app/internal/askuser"
	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/workspace"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	usageCacheTTL = time.Minute

	limitReached    = "reached"
	limitScheduled  = "scheduled"
	limitContinued  = "continued"
	limitSuperseded = "superseded"
)

var warnAtUsedPercent = []float64{75, 90, 95}

type usageWarningBody struct {
	Code          string     `json:"code"`
	Window        string     `json:"window"`
	WindowMinutes int        `json:"windowMinutes,omitempty"`
	PercentLeft   int        `json:"percentLeft"`
	ResetTime     *time.Time `json:"resetTime,omitempty"`
}

type limitBody struct {
	Window        string     `json:"window,omitempty"`
	WindowMinutes int        `json:"windowMinutes,omitempty"`
	ResetTime     *time.Time `json:"resetTime,omitempty"`
	Text          string     `json:"text,omitempty"`
	Status        string     `json:"status"`
	Model         string     `json:"model,omitempty"`
	Mode          string     `json:"mode,omitempty"`
}

type cachedUsage struct {
	windows []agents.Window
	until   time.Time
}

func (r *Runner) settleLimits(ctx context.Context, agent, subscription string, req workspace.RunRequest, t *transcript) {
	windows := sortedWindows(t.windows)
	if len(windows) == 0 || t.limit != nil && t.limit.window.ResetTime == nil {
		windows = mergeWindows(windows, r.readUsage(ctx, agent, subscription))
	}
	if t.limit != nil {
		window := exhaustedWindow(t.limit.window, windows)
		t.add("system", kindLimit, limitBody{
			Window: window.ID, WindowMinutes: window.WindowMinutes, ResetTime: window.ResetTime,
			Text: t.limit.text, Status: limitReached, Model: req.Model, Mode: req.Mode,
		})
		return
	}
	for _, window := range r.newWarnings(subscription, windows) {
		t.add("system", "note", usageWarningBody{
			Code: noteUsageWarning, Window: window.ID, WindowMinutes: window.WindowMinutes,
			PercentLeft: percentLeft(window.UsedPercent), ResetTime: window.ResetTime,
		})
	}
}

func (r *Runner) readUsage(ctx context.Context, agent, subscription string) []agents.Window {
	r.mu.Lock()
	cached, ok := r.usage[subscription]
	r.mu.Unlock()
	if ok && time.Now().Before(cached.until) {
		return cached.windows
	}
	driver, ok := r.cli.Driver(agent)
	if !ok {
		return nil
	}
	windows, err := driver.Usage(ctx, r.cli.Home(subscription))
	if err != nil {
		r.log.Debug("read usage after run", "err", err, "agent", agent)
	}
	r.mu.Lock()
	r.usage[subscription] = cachedUsage{windows: windows, until: time.Now().Add(usageCacheTTL)}
	r.mu.Unlock()
	return windows
}

func (r *Runner) newWarnings(subscription string, windows []agents.Window) []agents.Window {
	r.mu.Lock()
	defer r.mu.Unlock()
	var due []agents.Window
	for _, window := range windows {
		key := subscription + "/" + window.ID
		threshold := crossedThreshold(window.UsedPercent)
		if threshold > r.warned[key] {
			due = append(due, window)
		}
		if threshold == 0 {
			delete(r.warned, key)
			continue
		}
		r.warned[key] = threshold
	}
	return due
}

func crossedThreshold(usedPercent float64) float64 {
	crossed := 0.0
	for _, threshold := range warnAtUsedPercent {
		if usedPercent >= threshold {
			crossed = threshold
		}
	}
	return crossed
}

func percentLeft(usedPercent float64) int {
	return int(math.Max(0, math.Floor(100-usedPercent)))
}

func sortedWindows(byID map[string]agents.Window) []agents.Window {
	windows := make([]agents.Window, 0, len(byID))
	for _, window := range byID {
		windows = append(windows, window)
	}
	slices.SortFunc(windows, func(a, b agents.Window) int { return a.WindowMinutes - b.WindowMinutes })
	return windows
}

func mergeWindows(observed, read []agents.Window) []agents.Window {
	merged := slices.Clone(observed)
	for _, window := range read {
		if !slices.ContainsFunc(merged, func(known agents.Window) bool { return known.ID == window.ID }) {
			merged = append(merged, window)
		}
	}
	return merged
}

func exhaustedWindow(hit agents.Window, windows []agents.Window) agents.Window {
	if hit.ResetTime != nil && hit.ID != "" {
		return hit
	}
	var fullest *agents.Window
	for i, window := range windows {
		matches := hit.ID == "" || window.ID == hit.ID
		if matches && window.ResetTime != nil && (fullest == nil || window.UsedPercent > fullest.UsedPercent) {
			fullest = &windows[i]
		}
	}
	if fullest == nil {
		return hit
	}
	completed := *fullest
	if hit.ResetTime != nil {
		completed.ResetTime = hit.ResetTime
	}
	return completed
}

func (r *Runner) supersedeOpenCards(owner string, session pgtype.UUID) {
	replacements := []db.ReplaceMessageStatusParams{
		{SessionID: session, Kind: kindQuestion, FromStatus: askuser.CardPending, ToStatus: askuser.CardDismissed},
		{SessionID: session, Kind: kindLimit, FromStatus: limitReached, ToStatus: limitSuperseded},
	}
	for _, replacement := range replacements {
		messages, err := r.q.ReplaceMessageStatus(r.ctx, replacement)
		if err != nil {
			r.log.Warn("close open cards", "err", err, "kind", replacement.Kind)
			continue
		}
		for _, m := range messages {
			r.hub.Publish(owner, "message.updated", workspace.MessageOf(m), "")
		}
	}
}

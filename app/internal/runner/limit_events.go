package runner

import (
	"encoding/json"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/e1berd/drift/app/internal/agents"
	"github.com/e1berd/drift/app/internal/askuser"
)

const (
	claudeRejected  = "rejected"
	claudeOverage   = "overage"
	weekMinutes     = 7 * 24 * 60
	fiveHourMinutes = 5 * 60
)

var (
	limitExhaustedMessage = regexp.MustCompile(`(?i)(you['’]ve hit your [^.]{0,40}limit|usage limit)`)
	limitResetInMessage   = regexp.MustCompile(`\|(\d{10})\b`)
)

type claudeRateLimitInfo struct {
	Status         string                         `json:"status"`
	ResetsAt       int64                          `json:"resetsAt"`
	RateLimitType  string                         `json:"rateLimitType"`
	Utilization    *float64                       `json:"utilization"`
	UnifiedWindows map[string]claudeUnifiedWindow `json:"unifiedWindows"`
}

type claudeUnifiedWindow struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    int64   `json:"resetsAt"`
}

type claudeLimits struct {
	rejected *agents.Window
	refusal  string
}

func (l *claudeLimits) observe(info claudeRateLimitInfo, t transcriptSink) {
	for _, id := range slices.Sorted(maps.Keys(info.UnifiedWindows)) {
		window := info.UnifiedWindows[id]
		t.usage(claudeWindow(id, window.Utilization, window.ResetsAt))
	}
	_, listed := info.UnifiedWindows[info.RateLimitType]
	if info.Utilization != nil && !listed && info.RateLimitType != "" && info.RateLimitType != claudeOverage {
		t.usage(claudeWindow(info.RateLimitType, *info.Utilization, info.ResetsAt))
	}
	if info.Status == claudeRejected {
		window := claudeWindow(info.RateLimitType, 1, info.ResetsAt)
		l.rejected = &window
	}
}

func (l claudeLimits) exhausted(message string) (agents.Window, bool) {
	if l.rejected != nil {
		return *l.rejected, true
	}
	if l.refusal != "" || limitExhaustedMessage.MatchString(message) {
		return windowFromMessage(message), true
	}
	return agents.Window{}, false
}

func claudeWindow(id string, utilization float64, resetsAt int64) agents.Window {
	window := agents.Window{ID: id, UsedPercent: utilization * 100, WindowMinutes: claudeWindowMinutes(id)}
	if resetsAt > 0 {
		reset := time.Unix(resetsAt, 0).UTC()
		window.ResetTime = &reset
	}
	return window
}

func claudeWindowMinutes(id string) int {
	switch {
	case id == "five_hour":
		return fiveHourMinutes
	case strings.HasPrefix(id, "seven_day"):
		return weekMinutes
	}
	return 0
}

func windowFromMessage(message string) agents.Window {
	match := limitResetInMessage.FindStringSubmatch(message)
	if match == nil {
		return agents.Window{}
	}
	seconds, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return agents.Window{}
	}
	reset := time.Unix(seconds, 0).UTC()
	return agents.Window{ResetTime: &reset}
}

func askQuestions(t transcriptSink, arguments json.RawMessage) {
	if questions, err := askuser.ParseQuestions(arguments); err == nil {
		t.question(questions)
	}
}

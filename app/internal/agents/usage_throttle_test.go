package agents

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeUsage struct {
	calls   int
	windows []Window
	err     error
}

func (f *fakeUsage) fetch(context.Context) ([]Window, error) {
	f.calls++
	return f.windows, f.err
}

func newTestThrottle() (*usageThrottle, *time.Time) {
	now := time.Date(2026, 11, 20, 5, 0, 0, 0, time.UTC)
	throttle := newUsageThrottle()
	throttle.now = func() time.Time { return now }
	return throttle, &now
}

func TestUsageThrottleServesFreshReadingWithoutFetching(t *testing.T) {
	throttle, now := newTestThrottle()
	source := &fakeUsage{windows: []Window{{ID: "five_hour", UsedPercent: 52}}}
	ctx := context.Background()
	for range 3 {
		if _, err := throttle.read(ctx, "home", source.fetch); err != nil {
			t.Fatal(err)
		}
	}
	*now = now.Add(usageFreshFor - time.Second)
	if _, err := throttle.read(ctx, "home", source.fetch); err != nil || source.calls != 1 {
		t.Fatalf("calls %d err %v", source.calls, err)
	}
	*now = now.Add(2 * time.Second)
	if _, err := throttle.read(ctx, "home", source.fetch); err != nil || source.calls != 2 {
		t.Fatalf("stale reading must be fetched again, calls %d err %v", source.calls, err)
	}
}

func TestUsageThrottleKeepsLastWindowsWhileRateLimited(t *testing.T) {
	throttle, now := newTestThrottle()
	source := &fakeUsage{windows: []Window{{ID: "five_hour", UsedPercent: 52}}}
	ctx := context.Background()
	if _, err := throttle.read(ctx, "home", source.fetch); err != nil {
		t.Fatal(err)
	}
	*now = now.Add(usageFreshFor)
	source.windows, source.err = nil, usageRateLimited{service: "claude"}
	windows, err := throttle.read(ctx, "home", source.fetch)
	if err != nil || len(windows) != 1 || windows[0].UsedPercent != 52 {
		t.Fatalf("want last windows, got %v %v", windows, err)
	}
	*now = now.Add(usageRateLimitWait - time.Second)
	if _, err := throttle.read(WithForcedUsage(ctx), "home", source.fetch); err != nil || source.calls != 2 {
		t.Fatalf("rate limit must hold off even forced reads, calls %d err %v", source.calls, err)
	}
	*now = now.Add(2 * time.Second)
	source.windows, source.err = []Window{{ID: "five_hour", UsedPercent: 60}}, nil
	windows, err = throttle.read(ctx, "home", source.fetch)
	if err != nil || source.calls != 3 || windows[0].UsedPercent != 60 {
		t.Fatalf("calls %d windows %v err %v", source.calls, windows, err)
	}
}

func TestUsageThrottleHonoursLongerRetryAfter(t *testing.T) {
	throttle, now := newTestThrottle()
	source := &fakeUsage{err: usageRateLimited{service: "claude", retryAfter: time.Hour}}
	ctx := context.Background()
	if _, err := throttle.read(ctx, "home", source.fetch); !errors.As(err, &usageRateLimited{}) {
		t.Fatalf("want rate limit error, got %v", err)
	}
	*now = now.Add(30 * time.Minute)
	if _, err := throttle.read(ctx, "home", source.fetch); err == nil || source.calls != 1 {
		t.Fatalf("calls %d err %v", source.calls, err)
	}
}

func TestUsageThrottleForcedReadAtMostOncePerGap(t *testing.T) {
	throttle, now := newTestThrottle()
	source := &fakeUsage{windows: []Window{{ID: "five_hour"}}}
	forced := WithForcedUsage(context.Background())
	if _, err := throttle.read(forced, "home", source.fetch); err != nil {
		t.Fatal(err)
	}
	*now = now.Add(usageRetryGap / 2)
	if _, err := throttle.read(forced, "home", source.fetch); err != nil || source.calls != 1 {
		t.Fatalf("calls %d err %v", source.calls, err)
	}
	*now = now.Add(usageRetryGap)
	if _, err := throttle.read(forced, "home", source.fetch); err != nil || source.calls != 2 {
		t.Fatalf("calls %d err %v", source.calls, err)
	}
}

func TestUsageThrottleWaitsAfterFailure(t *testing.T) {
	throttle, now := newTestThrottle()
	source := &fakeUsage{err: errors.New("claude usage: 502 Bad Gateway")}
	ctx := context.Background()
	for range 3 {
		if _, err := throttle.read(ctx, "home", source.fetch); err == nil {
			t.Fatal("want error")
		}
	}
	if source.calls != 1 {
		t.Fatalf("calls %d", source.calls)
	}
	*now = now.Add(usageRetryGap)
	source.windows, source.err = []Window{{ID: "five_hour"}}, nil
	if _, err := throttle.read(ctx, "home", source.fetch); err != nil || source.calls != 2 {
		t.Fatalf("calls %d err %v", source.calls, err)
	}
}

func TestUsageThrottleDropsWindowsWhenSignedOut(t *testing.T) {
	throttle, now := newTestThrottle()
	source := &fakeUsage{windows: []Window{{ID: "five_hour"}}}
	ctx := context.Background()
	if _, err := throttle.read(ctx, "home", source.fetch); err != nil {
		t.Fatal(err)
	}
	*now = now.Add(usageFreshFor)
	source.windows, source.err = nil, UsageUnavailable{Reason: "expired"}
	if windows, err := throttle.read(ctx, "home", source.fetch); !errors.As(err, &UsageUnavailable{}) || windows != nil {
		t.Fatalf("windows %v err %v", windows, err)
	}
}

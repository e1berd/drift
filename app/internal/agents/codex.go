package agents

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type codex struct{ cli CLI }

func (d codex) start(ctx context.Context, home string) (*rpcServer, error) {
	server, err := startRPC(ctx, d.cli.Codex, d.cli.Env(home), home, "app-server")
	if err != nil {
		return nil, err
	}
	params := map[string]any{"clientInfo": map[string]string{"name": "drift", "version": "1"}}
	if err := server.call(ctx, "initialize", params, nil); err != nil {
		server.close()
		return nil, err
	}
	if err := server.notify("initialized"); err != nil {
		server.close()
		return nil, err
	}
	return server, nil
}

type codexDeviceLogin struct {
	LoginID         string `json:"loginId"`
	VerificationURL string `json:"verificationUrl"`
	UserCode        string `json:"userCode"`
}

type codexLoginCompleted struct {
	LoginID string `json:"loginId"`
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// StartLogin asks `codex app-server` for a ChatGPT device-code sign-in and waits for the
// account/login/completed notification.
func (d codex) StartLogin(home string) (*Login, error) {
	ctx, cancel := context.WithTimeout(context.Background(), loginTimeout)
	server, err := d.start(ctx, home)
	if err != nil {
		cancel()
		return nil, err
	}
	var started codexDeviceLogin
	startCtx, stopStart := context.WithTimeout(ctx, commandTimeout)
	err = server.call(startCtx, "account/login/start", map[string]string{"type": "chatgptDeviceCode"}, &started)
	stopStart()
	if err != nil {
		server.close()
		cancel()
		return nil, err
	}

	login := newLogin(func() {
		server.close()
		cancel()
	})
	login.URL, login.UserCode = started.VerificationURL, started.UserCode
	go func() {
		defer login.stop()
		login.finish(waitCodexLogin(ctx, server, started.LoginID))
	}()
	return login, nil
}

func waitCodexLogin(ctx context.Context, server *rpcServer, loginID string) error {
	for {
		select {
		case note := <-server.notes:
			if note.Method != "account/login/completed" {
				continue
			}
			var done codexLoginCompleted
			if err := json.Unmarshal(note.Params, &done); err != nil {
				return err
			}
			if done.LoginID != "" && done.LoginID != loginID {
				continue
			}
			if !done.Success {
				return errors.New(strings.TrimSpace("sign-in failed " + done.Error))
			}
			return nil
		case <-server.closed:
			return errors.New("codex stopped before the sign-in finished")
		case <-ctx.Done():
			return errors.New("the sign-in code expired")
		}
	}
}

// Account reads account/read of the app server.
func (d codex) Account(ctx context.Context, home string) (Account, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	server, err := d.start(ctx, home)
	if err != nil {
		return Account{}, err
	}
	defer server.close()
	var out struct {
		Account *struct {
			Type     string `json:"type"`
			Email    string `json:"email"`
			PlanType string `json:"planType"`
		} `json:"account"`
	}
	if err := server.call(ctx, "account/read", map[string]any{}, &out); err != nil {
		return Account{}, err
	}
	if out.Account == nil {
		return Account{}, errors.New("codex is not signed in")
	}
	return Account{Email: out.Account.Email, Plan: out.Account.PlanType}, nil
}

type codexWindow struct {
	UsedPercent        float64 `json:"usedPercent"`
	WindowDurationMins int     `json:"windowDurationMins"`
	ResetsAt           *int64  `json:"resetsAt"`
}

type codexRateLimits struct {
	RateLimits struct {
		Primary   *codexWindow `json:"primary"`
		Secondary *codexWindow `json:"secondary"`
	} `json:"rateLimits"`
}

// Usage reads account/rateLimits/read of the app server.
func (d codex) Usage(ctx context.Context, home string) ([]Window, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	server, err := d.start(ctx, home)
	if err != nil {
		return nil, err
	}
	defer server.close()
	var out codexRateLimits
	if err := server.call(ctx, "account/rateLimits/read", nil, &out); err != nil {
		var rpcErr *rpcError
		if errors.As(err, &rpcErr) && strings.Contains(rpcErr.Message, "authentication required") {
			return nil, UsageUnavailable{Reason: "signed_out"}
		}
		return nil, err
	}
	return codexWindows(out), nil
}

func codexWindows(limits codexRateLimits) []Window {
	windows := []Window{}
	named := []struct {
		id     string
		window *codexWindow
	}{{"primary", limits.RateLimits.Primary}, {"secondary", limits.RateLimits.Secondary}}
	for _, entry := range named {
		id, window := entry.id, entry.window
		if window == nil {
			continue
		}
		converted := Window{ID: id, UsedPercent: window.UsedPercent, WindowMinutes: window.WindowDurationMins}
		if window.ResetsAt != nil {
			reset := time.Unix(*window.ResetsAt, 0).UTC()
			converted.ResetTime = &reset
		}
		windows = append(windows, converted)
	}
	return windows
}

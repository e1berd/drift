package agents

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

var claudeAuthorizeURL = regexp.MustCompile(`https://\S+/oauth/authorize\S*`)

// ClaudeUsageURL is the endpoint behind Claude Code's /usage panel. It is not a documented
// public API (checked against Claude Code 2.1.327, 2026-11-18) and needs a full OAuth login:
// tokens from `claude setup-token` lack the user:profile scope and get 403.
var ClaudeUsageURL = "https://api.anthropic.com/api/oauth/usage"

// ClaudeModelsURL answers 401 to a revoked or mistyped OAuth token without spending usage.
var ClaudeModelsURL = "https://api.anthropic.com/v1/models"

// ErrTokenRejected means Anthropic refused the token.
var ErrTokenRejected = errors.New("Anthropic rejected the token")

// CheckClaudeToken refuses only tokens Anthropic answers with 401; other answers are let
// through, since a valid setup-token may lack scopes for this endpoint.
func CheckClaudeToken(ctx context.Context, client *http.Client, url, token string) error {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrTokenRejected
	}
	return nil
}

type claude struct{ cli CLI }

func (d claude) command(ctx context.Context, home string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, d.cli.Claude, args...)
	cmd.Env = d.cli.Env(home)
	cmd.Dir = home
	cmd.WaitDelay = 2 * time.Second
	return cmd
}

// StartLogin runs `claude auth login`, which prints an authorize URL and reads the code
// that the browser shows after sign-in from stdin.
func (d claude) StartLogin(home string) (*Login, error) {
	ctx, cancel := context.WithTimeout(context.Background(), loginTimeout)
	cmd := d.command(ctx, home, "auth", "login", "--claudeai")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	reader, writer := io.Pipe()
	cmd.Stdout, cmd.Stderr = writer, writer
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start claude: %w", err)
	}

	login := newLogin(cancel)
	login.NeedsCode = true
	var mu sync.Mutex
	login.submit = func(code string) error {
		mu.Lock()
		defer mu.Unlock()
		_, err := io.WriteString(stdin, strings.TrimSpace(code)+"\n")
		return err
	}

	found := make(chan string, 1)
	var output bytes.Buffer
	scanned := make(chan struct{})
	go func() {
		defer close(scanned)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")
			if url := claudeAuthorizeURL.FindString(line); url != "" {
				select {
				case found <- url:
				default:
				}
			}
		}
	}()
	go func() {
		err := cmd.Wait()
		writer.Close()
		<-scanned
		if err != nil {
			err = loginFailure(output.String(), err)
		}
		cancel()
		login.finish(err)
	}()

	select {
	case url := <-found:
		login.URL = url
		return login, nil
	case <-login.Finished():
		return nil, login.Result()
	case <-time.After(commandTimeout):
		login.Cancel()
		return nil, errors.New("claude did not print a sign-in link")
	}
}

func loginFailure(output string, err error) error {
	if message := lastMessage(output); message != "" {
		return errors.New(message)
	}
	return err
}

type claudeStatus struct {
	LoggedIn         bool   `json:"loggedIn"`
	Email            string `json:"email"`
	SubscriptionType string `json:"subscriptionType"`
	AuthMethod       string `json:"authMethod"`
}

// Account reads `claude auth status --json`.
func (d claude) Account(ctx context.Context, home string) (Account, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	out, runErr := d.command(ctx, home, "auth", "status", "--json").Output()
	var status claudeStatus
	if err := json.Unmarshal(out, &status); err != nil {
		return Account{}, errors.Join(errors.New("cannot read claude auth status"), runErr)
	}
	if !status.LoggedIn {
		return Account{}, errors.New("claude is not signed in")
	}
	return Account{Email: status.Email, Plan: status.SubscriptionType}, nil
}

type claudeCredentials struct {
	OAuth struct {
		AccessToken string `json:"accessToken"`
		ExpiresAt   int64  `json:"expiresAt"`
	} `json:"claudeAiOauth"`
}

func readClaudeCredentials(home string) (claudeCredentials, error) {
	var creds claudeCredentials
	raw, err := os.ReadFile(filepath.Join(home, ".claude", ".credentials.json"))
	if err != nil {
		return creds, err
	}
	return creds, json.Unmarshal(raw, &creds)
}

func (c claudeCredentials) expired(now time.Time) bool {
	return c.OAuth.ExpiresAt > 0 && time.UnixMilli(c.OAuth.ExpiresAt).Before(now)
}

// Usage reads the limit windows with the access token the CLI stored after sign-in.
// An expired token is refreshed by asking the CLI for its status first.
func (d claude) Usage(ctx context.Context, home string) ([]Window, error) {
	creds, err := readClaudeCredentials(home)
	if err != nil {
		if _, tokenErr := os.Stat(filepath.Join(home, tokenFile)); tokenErr == nil {
			return nil, UsageUnavailable{Reason: "token_login"}
		}
		return nil, UsageUnavailable{Reason: "signed_out"}
	}
	if creds.expired(time.Now()) {
		_, _ = d.Account(ctx, home)
		if creds, err = readClaudeCredentials(home); err != nil || creds.expired(time.Now()) {
			return nil, UsageUnavailable{Reason: "expired"}
		}
	}
	return fetchClaudeUsage(ctx, http.DefaultClient, ClaudeUsageURL, creds.OAuth.AccessToken)
}

func fetchClaudeUsage(ctx context.Context, client *http.Client, url, token string) ([]Window, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, UsageUnavailable{Reason: "expired"}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("claude usage: %s", resp.Status)
	}
	return parseClaudeUsage(io.LimitReader(resp.Body, 1<<20))
}

var claudeWindowMinutes = map[string]int{"five_hour": 300}

func parseClaudeUsage(body io.Reader) ([]Window, error) {
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		return nil, err
	}
	windows := []Window{}
	for _, id := range slices.Sorted(maps.Keys(raw)) {
		var entry struct {
			Utilization *float64 `json:"utilization"`
			ResetsAt    string   `json:"resets_at"`
		}
		if json.Unmarshal(raw[id], &entry) != nil || entry.Utilization == nil {
			continue
		}
		window := Window{ID: id, UsedPercent: *entry.Utilization, WindowMinutes: claudeWindowMinutes[id]}
		if strings.HasPrefix(id, "seven_day") {
			window.WindowMinutes = 7 * 24 * 60
		}
		if reset, err := time.Parse(time.RFC3339, entry.ResetsAt); err == nil {
			window.ResetTime = &reset
		}
		windows = append(windows, window)
	}
	return windows, nil
}

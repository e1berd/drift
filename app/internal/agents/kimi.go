package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var kimiDeviceURL = regexp.MustCompile(`https://www\.kimi\.[a-z.]+/code/authorize_device\?\S+`)

const (
	kimiOAuthClientID = "17e5f671-d194-4dfb-9706-5516cb48c098"
	kimiDefaultBase   = "https://api.kimi.ai/coding/v1"
	kimiDefaultOAuth  = "https://auth.kimi.ai"
	kimiTokenSkew     = time.Minute
)

// ErrKimiUnauthorized means Kimi rejected the stored credentials.
var ErrKimiUnauthorized = errors.New("kimi rejected the credentials")

type kimi struct{ cli CLI }

func (d kimi) command(ctx context.Context, home string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, d.cli.KimiPath(), args...)
	cmd.Env = d.cli.KimiEnv(home)
	cmd.Dir = home
	cmd.WaitDelay = 2 * time.Second
	return cmd
}

// StartLogin runs `kimi login`, the device-code flow: it prints an authorize link with the
// code embedded and exits by itself once the browser sign-in finishes.
func (d kimi) StartLogin(home string) (*Login, error) {
	return startDeviceLogin(kimiDeviceURL, func(ctx context.Context) *exec.Cmd {
		return d.command(ctx, home, "login")
	})
}

type kimiConfig struct {
	baseURL   string
	oauthHost string
	oauthKey  string
}

func readKimiConfig(home string) kimiConfig {
	cfg := kimiConfig{baseURL: kimiDefaultBase, oauthHost: kimiDefaultOAuth}
	raw, err := os.ReadFile(filepath.Join(home, "config.toml"))
	if err != nil {
		return cfg
	}
	section := ""
	for line := range strings.Lines(string(raw)) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			section = line
			continue
		}
		if section != `[providers."managed:kimi-code"]` && section != `[providers."managed:kimi-code".oauth]` {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"`)
		switch strings.TrimSpace(key) {
		case "base_url":
			cfg.baseURL = strings.TrimRight(value, "/")
		case "oauth_host":
			cfg.oauthHost = strings.TrimRight(value, "/")
		case "key":
			cfg.oauthKey = value
		}
	}
	return cfg
}

type kimiCredentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func readKimiCredentials(home string) (kimiCredentials, error) {
	var creds kimiCredentials
	key := readKimiConfig(home).oauthKey
	if key == "" {
		entries, err := os.ReadDir(filepath.Join(home, "credentials"))
		if err != nil || len(entries) == 0 {
			return creds, errors.New("kimi is not signed in")
		}
		key = "credentials/" + entries[0].Name()
	} else if !strings.HasPrefix(key, "credentials/") {
		key = "credentials/" + filepath.Base(key) + ".json"
	}
	raw, err := os.ReadFile(filepath.Join(home, key))
	if err != nil {
		return creds, errors.New("kimi is not signed in")
	}
	if err := json.Unmarshal(raw, &creds); err != nil {
		return creds, err
	}
	if creds.AccessToken == "" {
		return creds, errors.New("kimi is not signed in")
	}
	return creds, nil
}

func (c kimiCredentials) expired(now time.Time) bool {
	return c.ExpiresAt > 0 && time.Unix(c.ExpiresAt, 0).Before(now.Add(kimiTokenSkew))
}

func writeKimiCredentials(home, key string, creds kimiCredentials) error {
	raw, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(home, key), raw, 0o600)
}

// ensureKimiToken returns a usable access token, refreshing it against the OAuth host when the
// stored one is close to expiry.
func ensureKimiToken(ctx context.Context, home string) (string, error) {
	cfg := readKimiConfig(home)
	creds, err := readKimiCredentials(home)
	if err != nil {
		return "", err
	}
	if !creds.expired(time.Now()) {
		return creds.AccessToken, nil
	}
	refreshed, err := refreshKimiToken(ctx, http.DefaultClient, cfg.oauthHost, creds.RefreshToken)
	if err != nil {
		return "", err
	}
	key := cfg.oauthKey
	if key == "" {
		entries, _ := os.ReadDir(filepath.Join(home, "credentials"))
		if len(entries) == 0 {
			return "", errors.New("kimi is not signed in")
		}
		key = "credentials/" + entries[0].Name()
	}
	if err := writeKimiCredentials(home, key, refreshed); err != nil {
		return "", err
	}
	return refreshed.AccessToken, nil
}

func refreshKimiToken(ctx context.Context, client *http.Client, oauthHost, refreshToken string) (kimiCredentials, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {kimiOAuthClientID},
		"refresh_token": {refreshToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthHost+"/api/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return kimiCredentials{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return kimiCredentials{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest {
		return kimiCredentials{}, ErrKimiUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return kimiCredentials{}, fmt.Errorf("kimi token refresh: %s", resp.Status)
	}
	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out); err != nil {
		return kimiCredentials{}, err
	}
	if out.AccessToken == "" {
		return kimiCredentials{}, errors.New("kimi token refresh returned no access token")
	}
	if out.RefreshToken == "" {
		out.RefreshToken = refreshToken
	}
	return kimiCredentials{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(out.ExpiresIn) * time.Second).Unix(),
	}, nil
}

type kimiUser struct {
	Email         string `json:"email"`
	UserID        string `json:"user_id"`
	Nickname      string `json:"nickname"`
	UserLevelName string `json:"user_level_name"`
}

// Account reads the managed profile with the stored OAuth token.
func (d kimi) Account(ctx context.Context, home string) (Account, error) {
	token, err := ensureKimiToken(ctx, home)
	if err != nil {
		return Account{}, err
	}
	user, err := fetchKimiAccount(ctx, http.DefaultClient, readKimiConfig(home).baseURL, token)
	if err != nil {
		return Account{}, err
	}
	return Account{Email: firstNonEmptyString(user.Email, user.Nickname, user.UserID), Plan: user.UserLevelName}, nil
}

func fetchKimiAccount(ctx context.Context, client *http.Client, base, token string) (kimiUser, error) {
	var user kimiUser
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	req, err := authenticatedKimiRequest(ctx, base+"/me", token)
	if err != nil {
		return user, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return user, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return user, ErrKimiUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return user, fmt.Errorf("kimi profile: %s", resp.Status)
	}
	return user, json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&user)
}

var kimiWindowMinutes = map[string]int{
	"limit_5h":          5 * 60,
	"limit_7d":          7 * 24 * 60,
	"limit_month_total": 30 * 24 * 60,
	"limit_month_code":  30 * 24 * 60,
}

var kimiWindowLabels = map[string]string{
	"limit_5h":          "5h",
	"limit_7d":          "7d",
	"limit_month_total": "Month",
	"limit_month_code":  "Month · code",
}

// Usage reads the managed quota windows the Kimi backend reports for the subscription.
func (d kimi) Usage(ctx context.Context, home string) ([]Window, error) {
	token, err := ensureKimiToken(ctx, home)
	if err != nil {
		if errors.Is(err, ErrKimiUnauthorized) {
			return nil, UsageUnavailable{Reason: "signed_out"}
		}
		return nil, err
	}
	windows, err := fetchKimiUsage(ctx, http.DefaultClient, readKimiConfig(home).baseURL, token)
	if errors.Is(err, ErrKimiUnauthorized) {
		return nil, UsageUnavailable{Reason: "signed_out"}
	}
	return windows, err
}

func fetchKimiUsage(ctx context.Context, client *http.Client, base, token string) ([]Window, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	req, err := authenticatedKimiRequest(ctx, base+"/usages", token)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrKimiUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kimi usage: %s", resp.Status)
	}
	return parseKimiUsage(io.LimitReader(resp.Body, 1<<20))
}

func authenticatedKimiRequest(ctx context.Context, rawURL, token string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func parseKimiUsage(body io.Reader) ([]Window, error) {
	var raw struct {
		Usages map[string]struct {
			UsedRatio float64 `json:"used_ratio"`
			ResetTime string  `json:"reset_time"`
		} `json:"usages"`
	}
	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		return nil, err
	}
	windows := make([]Window, 0, len(raw.Usages))
	for id, entry := range raw.Usages {
		window := Window{
			ID:            id,
			Label:         kimiWindowLabels[id],
			UsedPercent:   entry.UsedRatio * 100,
			WindowMinutes: kimiWindowMinutes[id],
		}
		if reset, err := time.Parse(time.RFC3339, entry.ResetTime); err == nil {
			window.ResetTime = &reset
		}
		windows = append(windows, window)
	}
	return windows, nil
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

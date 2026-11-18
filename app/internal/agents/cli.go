// Package agents drives the official agent CLIs (Claude Code, Codex, Antigravity) on behalf of Drift:
// signing in to a subscription, reading the signed-in account and its usage limits.
// Every subscription gets its own home folder, so several accounts of one agent can coexist.
package agents

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	loginTimeout   = 15 * time.Minute
	commandTimeout = 30 * time.Second
	tokenFile      = "oauth-token"
)

var passthroughEnv = []string{
	"PATH", "LANG", "TZ", "SSL_CERT_FILE", "SSL_CERT_DIR",
	"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "no_proxy",
}

var (
	ansiEscape  = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)
	inputPrompt = regexp.MustCompile(`^.*?(?:> |press Enter:\s*)`)
)

// Account is who a subscription is signed in as.
type Account struct {
	Email string
	Plan  string
}

// Window is one usage limit window, e.g. the five-hour or the weekly quota, or the quota of
// one model; Label is the agent's own name for it when the agent names its windows.
type Window struct {
	ID            string
	Label         string
	UsedPercent   float64
	WindowMinutes int
	ResetTime     *time.Time
}

// UsageUnavailable explains why limits cannot be read; Reason is a stable code for the UI.
type UsageUnavailable struct{ Reason string }

func (e UsageUnavailable) Error() string { return "usage unavailable: " + e.Reason }

// Driver is what Drift needs from one agent CLI.
type Driver interface {
	StartLogin(home string) (*Login, error)
	Account(ctx context.Context, home string) (Account, error)
	Usage(ctx context.Context, home string) ([]Window, error)
}

// CLI locates the agent binaries and the folder that holds one home per subscription.
type CLI struct {
	Root        string
	Claude      string
	Codex       string
	Antigravity string
}

// Driver returns the driver of an agent id, or false when Drift cannot sign in to it yet.
func (c CLI) Driver(agent string) (Driver, bool) {
	switch agent {
	case "claude":
		return claude{c}, true
	case "codex":
		return codex{c}, true
	case "antigravity":
		return antigravity{c}, true
	}
	return nil, false
}

// Home is the private folder of one subscription.
func (c CLI) Home(id string) string {
	return filepath.Join(c.Root, id)
}

// Prepare creates the subscription home with the config folders of every CLI.
func (c CLI) Prepare(id string) (string, error) {
	home := c.Home(id)
	for _, dir := range []string{home, filepath.Join(home, ".claude"), filepath.Join(home, ".codex")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return "", err
		}
	}
	if err := prepareAntigravity(home); err != nil {
		return "", err
	}
	return home, nil
}

// Remove deletes the subscription home with its credentials.
func (c CLI) Remove(id string) error {
	if id == "" || strings.ContainsAny(id, `/\.`) {
		return errors.New("invalid subscription id")
	}
	return os.RemoveAll(c.Home(id))
}

// SaveClaudeToken stores a long-lived token from `claude setup-token` for the subscription.
func (c CLI) SaveClaudeToken(home, token string) error {
	return os.WriteFile(filepath.Join(home, tokenFile), []byte(strings.TrimSpace(token)), 0o600)
}

// Env is the environment an agent process of this subscription runs with.
func (c CLI) Env(home string) []string {
	env := []string{
		"HOME=" + home,
		"CLAUDE_CONFIG_DIR=" + filepath.Join(home, ".claude"),
		"CODEX_HOME=" + filepath.Join(home, ".codex"),
		"NO_COLOR=1",
		"TERM=dumb",
	}
	for _, key := range passthroughEnv {
		if value, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+value)
		}
	}
	if token, err := os.ReadFile(filepath.Join(home, tokenFile)); err == nil && len(token) > 0 {
		env = append(env, "CLAUDE_CODE_OAUTH_TOKEN="+string(token))
	}
	return env
}

func lastMessage(output string) string {
	lines := strings.Split(ansiEscape.ReplaceAllString(output, ""), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		line = strings.TrimSpace(inputPrompt.ReplaceAllString(line, ""))
		if line != "" {
			return line
		}
	}
	return ""
}

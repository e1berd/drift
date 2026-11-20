package agents

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var mimoAuthorizeURL = regexp.MustCompile(`https://platform\.xiaomimimo\.com/authorize\?\S+`)

const mimoPlan = "MiMo"

type mimo struct{ cli CLI }

func (d mimo) command(ctx context.Context, home string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, d.cli.MimoPath(), args...)
	cmd.Env = d.cli.MimoEnv(home)
	cmd.Dir = home
	cmd.WaitDelay = 2 * time.Second
	return cmd
}

// StartLogin runs `mimo providers login`, the MiMo account sign-in: it prints an authorize
// link, and the code the browser shows after sign-in goes back to the CLI through stdin.
func (d mimo) StartLogin(home string) (*Login, error) {
	return startPasteLogin(home, mimoAuthorizeURL, func(ctx context.Context) *exec.Cmd {
		return d.command(ctx, home, "providers", "login", "-p", "xiaomi")
	})
}

type mimoProvider struct {
	Metadata struct {
		UID string `json:"uid"`
	} `json:"metadata"`
}

func xiaomiProviderName(name string) bool {
	return name == "xiaomi" || strings.HasPrefix(name, "xiaomi-")
}

func xiaomiAccountUID(providers map[string]mimoProvider) string {
	for name, provider := range providers {
		if xiaomiProviderName(name) && provider.Metadata.UID != "" {
			return provider.Metadata.UID
		}
	}
	return ""
}

func readMimoProviders(home string) (map[string]mimoProvider, error) {
	raw, err := os.ReadFile(filepath.Join(home, ".local", "share", "mimocode", "auth.json"))
	if err != nil {
		return nil, err
	}
	var providers map[string]mimoProvider
	if err := json.Unmarshal(raw, &providers); err != nil {
		return nil, err
	}
	return providers, nil
}

const mimoSignedInModel = "mimo-v2.6-pro"

// MimoModel picks the model of the provider the account signed in with. Without -m the
// CLI falls back to the plain "xiaomi" provider, which token plan sign-ins such as
// "xiaomi-token-plan-sgp" have no key for, and every run fails with "Invalid API Key".
func MimoModel(home, requested string) string {
	if requested != "" && requested != "default" {
		return requested
	}
	providers, err := readMimoProviders(home)
	if err != nil {
		return requested
	}
	if _, ok := providers["xiaomi"]; ok {
		return requested
	}
	for name := range providers {
		if xiaomiProviderName(name) {
			return name + "/" + mimoSignedInModel
		}
	}
	return requested
}

// Account reads the signed-in MiMo account from the credentials the CLI stored.
func (d mimo) Account(ctx context.Context, home string) (Account, error) {
	providers, err := readMimoProviders(home)
	if err != nil {
		return Account{}, errors.New("mimo is not signed in")
	}
	uid := xiaomiAccountUID(providers)
	if uid == "" {
		return Account{}, errors.New("mimo is not signed in")
	}
	return Account{Email: uid, Plan: mimoPlan}, nil
}

// Usage is not reported by the MiMo CLI for subscription sign-ins.
func (d mimo) Usage(ctx context.Context, home string) ([]Window, error) {
	return nil, UsageUnavailable{Reason: "unsupported"}
}

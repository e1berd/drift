package agents

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

type mimoAuth struct {
	Xiaomi *struct {
		Metadata struct {
			UID string `json:"uid"`
		} `json:"metadata"`
	} `json:"xiaomi"`
}

// Account reads the signed-in MiMo account from the credentials the CLI stored.
func (d mimo) Account(ctx context.Context, home string) (Account, error) {
	raw, err := os.ReadFile(filepath.Join(home, ".local", "share", "mimocode", "auth.json"))
	if err != nil {
		return Account{}, errors.New("mimo is not signed in")
	}
	var auth mimoAuth
	if json.Unmarshal(raw, &auth) != nil || auth.Xiaomi == nil || auth.Xiaomi.Metadata.UID == "" {
		return Account{}, errors.New("mimo is not signed in")
	}
	return Account{Email: auth.Xiaomi.Metadata.UID, Plan: mimoPlan}, nil
}

// Usage is not reported by the MiMo CLI for subscription sign-ins.
func (d mimo) Usage(ctx context.Context, home string) ([]Window, error) {
	return nil, UsageUnavailable{Reason: "unsupported"}
}

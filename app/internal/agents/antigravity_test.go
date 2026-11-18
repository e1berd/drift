package agents

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const agyUsageOutput = "Account\tdev@example.com\nPlan\tGoogle AI Pro\nModel\tRemaining\tResets\nGemini 4 Argon\t80%\t2026-11-18T15:00:00Z\nGemini 3.8 Flash\t100%\n"

func fakeAntigravity() {
	args := strings.Join(os.Args[1:], " ")
	signedIn := func() bool {
		_, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".gemini", "signed-in"))
		return err == nil
	}
	switch args {
	case "-p /usage":
		if !signedIn() {
			fmt.Fprintln(os.Stderr, "Error: authentication required. Sign in once with the Antigravity CLI on this machine (agy).")
			os.Exit(1)
		}
		fmt.Print(agyUsageOutput)
		os.Exit(0)
	case "acp --login":
		fmt.Println("To sign in, open this URL in a browser:")
		fmt.Println("  https://accounts.google.com/o/oauth2/v2/auth?client_id=x&redirect_uri=y")
		fmt.Print("Or, paste the authorization code here and press Enter: ")
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if strings.TrimSpace(line) != "4/good" {
			fmt.Println("couldn't log in: invalid grant")
			os.Exit(1)
		}
		_ = os.WriteFile(filepath.Join(os.Getenv("HOME"), ".gemini", "signed-in"), nil, 0o600)
		fmt.Println("Logged in.")
		os.Exit(0)
	}
	os.Exit(2)
}

func TestAntigravityLoginThenAccountAndQuota(t *testing.T) {
	cli := fakeCLI(t, "antigravity")
	home, err := cli.Prepare("s1")
	if err != nil {
		t.Fatal(err)
	}
	driver, ok := cli.Driver("antigravity")
	if !ok {
		t.Fatal("antigravity driver missing")
	}
	var unavailable UsageUnavailable
	if _, err := driver.Usage(context.Background(), home); !errors.As(err, &unavailable) || unavailable.Reason != "signed_out" {
		t.Fatalf("want signed_out before sign-in, got %v", err)
	}
	login, err := driver.StartLogin(home)
	if err != nil {
		t.Fatal(err)
	}
	if !login.NeedsCode || !strings.HasPrefix(login.URL, "https://accounts.google.com/o/oauth2/") {
		t.Fatalf("unexpected login %+v", login)
	}
	if err := login.Submit(" 4/good "); err != nil {
		t.Fatal(err)
	}
	if err := waitLogin(t, login); err != nil {
		t.Fatal(err)
	}
	account, err := driver.Account(context.Background(), home)
	if err != nil || account.Email != "dev@example.com" || account.Plan != "Google AI Pro" {
		t.Fatalf("account %+v, %v", account, err)
	}
	windows, err := driver.Usage(context.Background(), home)
	if err != nil || len(windows) != 2 {
		t.Fatalf("windows %+v, %v", windows, err)
	}
	argon := windows[0]
	if argon.ID != "gemini_4_argon" || argon.Label != "Gemini 4 Argon" || argon.UsedPercent != 20 {
		t.Fatalf("argon %+v", argon)
	}
	if argon.ResetTime == nil || !argon.ResetTime.Equal(time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)) {
		t.Fatalf("argon reset %v", argon.ResetTime)
	}
	if windows[1].UsedPercent != 0 || windows[1].ResetTime != nil {
		t.Fatalf("flash %+v", windows[1])
	}
}

func TestAntigravityLoginReportsRejectedCode(t *testing.T) {
	cli := fakeCLI(t, "antigravity")
	home, _ := cli.Prepare("s1")
	driver, _ := cli.Driver("antigravity")
	login, err := driver.StartLogin(home)
	if err != nil {
		t.Fatal(err)
	}
	_ = login.Submit("bad")
	if err := waitLogin(t, login); err == nil || err.Error() != "couldn't log in: invalid grant" {
		t.Fatalf("want the CLI message, got %v", err)
	}
}

func TestParseAntigravityUsageReadsUsedColumnsAndFractions(t *testing.T) {
	report := parseAntigravityUsage("model\tused\treset\nGemini 4 Argon\t35%\t2026-11-18 15:00:00 UTC\nClaude Sonnet 5.5\t0.25\n")
	if len(report.windows) != 2 || report.windows[0].UsedPercent != 35 || report.windows[0].ResetTime == nil {
		t.Fatalf("windows %+v", report.windows)
	}
	if report.windows[1].ID != "claude_sonnet_5_5" || report.windows[1].UsedPercent != 25 {
		t.Fatalf("fraction %+v", report.windows[1])
	}
}

func TestPrepareKeepsAntigravitySettings(t *testing.T) {
	cli := CLI{Root: t.TempDir()}
	home, err := cli.Prepare("s1")
	if err != nil {
		t.Fatal(err)
	}
	settings := filepath.Join(home, ".gemini", "antigravity-cli", "settings.json")
	if raw, err := os.ReadFile(settings); err != nil || string(raw) != antigravitySettings {
		t.Fatalf("settings %q, %v", raw, err)
	}
	if err := os.WriteFile(settings, []byte(`{"agentMode":"auto"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.Prepare("s1"); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(settings); string(raw) != `{"agentMode":"auto"}` {
		t.Fatalf("own settings overwritten: %q", raw)
	}
}

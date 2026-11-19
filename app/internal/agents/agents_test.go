package agents

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	switch os.Getenv("DRIFT_FAKE_CLI") {
	case "claude":
		fakeClaude()
	case "codex":
		fakeCodex()
	case "antigravity":
		fakeAntigravity()
	case "kimi":
		fakeKimi()
	case "mimo":
		fakeMimo()
	default:
		os.Exit(m.Run())
	}
}

func fakeClaude() {
	args := strings.Join(os.Args[1:], " ")
	if strings.HasPrefix(args, "auth status") {
		fmt.Println(`{"loggedIn":true,"authMethod":"claude.ai","email":"dev@example.com","subscriptionType":"max"}`)
		os.Exit(0)
	}
	fmt.Println("If the browser didn't open, visit: https://claude.com/cai/oauth/authorize?code=true&state=x")
	fmt.Print("Paste code here if prompted > ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	if strings.TrimSpace(line) == "good#state" {
		os.Exit(0)
	}
	fmt.Println("Couldn't log in: That code wasn't accepted.")
	os.Exit(1)
}

func fakeCodex() {
	out := json.NewEncoder(os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var msg rpcMessage
		_ = json.Unmarshal(scanner.Bytes(), &msg)
		if msg.ID == nil {
			continue
		}
		switch msg.Method {
		case "initialize":
			_ = out.Encode(map[string]any{"id": *msg.ID, "result": map[string]any{}})
		case "account/login/start":
			_ = out.Encode(map[string]any{"id": *msg.ID, "result": map[string]any{
				"type": "chatgptDeviceCode", "loginId": "l1",
				"verificationUrl": "https://auth.openai.com/codex/device", "userCode": "ABCD-1234",
			}})
			_ = out.Encode(map[string]any{"method": "account/login/completed", "params": map[string]any{"loginId": "l1", "success": true}})
		case "account/read":
			_ = out.Encode(map[string]any{"id": *msg.ID, "result": map[string]any{
				"account": map[string]any{"type": "chatgpt", "email": "dev@example.com", "planType": "pro"},
			}})
		case "account/rateLimits/read":
			_ = out.Encode(map[string]any{"id": *msg.ID, "result": map[string]any{"rateLimits": map[string]any{
				"primary":   map[string]any{"usedPercent": 42, "windowDurationMins": 300, "resetsAt": 1794970000},
				"secondary": map[string]any{"usedPercent": 7, "windowDurationMins": 10080},
			}}})
		}
	}
	os.Exit(0)
}

func fakeKimi() {
	args := strings.Join(os.Args[1:], " ")
	if strings.HasPrefix(args, "login") {
		fmt.Println("Opening browser for Kimi device login: https://www.kimi.ai/code/authorize_device?user_code=KO87-A7XC")
		fmt.Println("If the browser did not open, paste the URL above and enter code: KO87-A7XC")
		fmt.Println("Code expires in 1800s.")
		os.Exit(0)
	}
	fmt.Println("unknown command")
	os.Exit(1)
}

func fakeMimo() {
	args := strings.Join(os.Args[1:], " ")
	if strings.HasPrefix(args, "providers login") {
		fmt.Println("Browser didn't open? Use the url below to sign in:")
		fmt.Println("https://platform.xiaomimimo.com/authorize?pk=abc&redirect_uri=callback&kn=mimocode&key_name=mimo-code-cli-key-test")
		fmt.Print("Paste code here if prompted > ")
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if strings.TrimSpace(line) == "good-code" {
			dir := filepath.Join(os.Getenv("XDG_DATA_HOME"), "mimocode")
			_ = os.MkdirAll(dir, 0o700)
			auth := `{"xiaomi":{"type":"api","key":"k","metadata":{"uid":"6890014132","base_url":"https://token-plan-sgp.xiaomimimo.com/v1"}}}`
			_ = os.WriteFile(filepath.Join(dir, "auth.json"), []byte(auth), 0o600)
			os.Exit(0)
		}
		fmt.Println("Sign-in failed: wrong code")
		os.Exit(1)
	}
	fmt.Println("unknown command")
	os.Exit(1)
}

func fakeCLI(t *testing.T, agent string) CLI {
	t.Helper()
	t.Setenv("DRIFT_FAKE_CLI", agent)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	passthroughEnv = append(passthroughEnv, "DRIFT_FAKE_CLI")
	t.Cleanup(func() { passthroughEnv = passthroughEnv[:len(passthroughEnv)-1] })
	return CLI{Root: t.TempDir(), Claude: self, Codex: self, Antigravity: self, Kimi: self, Mimo: self}
}

func waitLogin(t *testing.T, login *Login) error {
	t.Helper()
	select {
	case <-login.Finished():
		return login.Result()
	case <-time.After(10 * time.Second):
		t.Fatal("login did not finish")
		return nil
	}
}

func TestClaudeLoginAcceptsCode(t *testing.T) {
	cli := fakeCLI(t, "claude")
	home, err := cli.Prepare("s1")
	if err != nil {
		t.Fatal(err)
	}
	driver, _ := cli.Driver("claude")
	login, err := driver.StartLogin(home)
	if err != nil {
		t.Fatal(err)
	}
	if !login.NeedsCode || !strings.Contains(login.URL, "/oauth/authorize") {
		t.Fatalf("unexpected login %+v", login)
	}
	if err := login.Submit("good#state"); err != nil {
		t.Fatal(err)
	}
	if err := waitLogin(t, login); err != nil {
		t.Fatal(err)
	}
	account, err := driver.Account(context.Background(), home)
	if err != nil || account.Email != "dev@example.com" || account.Plan != "max" {
		t.Fatalf("account %+v, %v", account, err)
	}
}

func TestClaudeLoginReportsRejectedCode(t *testing.T) {
	cli := fakeCLI(t, "claude")
	home, _ := cli.Prepare("s1")
	driver, _ := cli.Driver("claude")
	login, err := driver.StartLogin(home)
	if err != nil {
		t.Fatal(err)
	}
	_ = login.Submit("bad")
	err = waitLogin(t, login)
	if err == nil || err.Error() != "Couldn't log in: That code wasn't accepted." {
		t.Fatalf("want the CLI message, got %v", err)
	}
}

func TestCodexDeviceLoginAndLimits(t *testing.T) {
	cli := fakeCLI(t, "codex")
	home, _ := cli.Prepare("s1")
	driver, _ := cli.Driver("codex")
	login, err := driver.StartLogin(home)
	if err != nil {
		t.Fatal(err)
	}
	if login.UserCode != "ABCD-1234" || login.NeedsCode {
		t.Fatalf("unexpected login %+v", login)
	}
	if err := waitLogin(t, login); err != nil {
		t.Fatal(err)
	}
	account, err := driver.Account(context.Background(), home)
	if err != nil || account.Plan != "pro" {
		t.Fatalf("account %+v, %v", account, err)
	}
	windows, err := driver.Usage(context.Background(), home)
	if err != nil || len(windows) != 2 || windows[0].ID != "primary" || windows[0].UsedPercent != 42 {
		t.Fatalf("windows %+v, %v", windows, err)
	}
	if windows[0].ResetTime == nil || windows[1].ResetTime != nil || windows[1].WindowMinutes != 10080 {
		t.Fatalf("reset times %+v", windows)
	}
}

func TestCancelStopsLogin(t *testing.T) {
	cli := fakeCLI(t, "claude")
	home, _ := cli.Prepare("s1")
	driver, _ := cli.Driver("claude")
	login, err := driver.StartLogin(home)
	if err != nil {
		t.Fatal(err)
	}
	login.Cancel()
	if err := waitLogin(t, login); err == nil || err.Error() != "sign-in cancelled" {
		t.Fatalf("got %v", err)
	}
}

func TestFetchClaudeUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer good" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		fmt.Fprint(w, `{"five_hour":{"utilization":31.5,"resets_at":"2026-11-18T10:00:00+00:00"},
			"seven_day":{"utilization":12,"resets_at":null},"seven_day_opus":null,"extra_usage":{"is_enabled":false}}`)
	}))
	defer server.Close()

	windows, err := fetchClaudeUsage(context.Background(), server.Client(), server.URL, "good")
	if err != nil {
		t.Fatal(err)
	}
	if len(windows) != 2 || windows[0].ID != "five_hour" || windows[0].WindowMinutes != 300 || windows[0].ResetTime == nil {
		t.Fatalf("unexpected windows %+v", windows)
	}
	if windows[1].ID != "seven_day" || windows[1].WindowMinutes != 10080 || windows[1].ResetTime != nil {
		t.Fatalf("unexpected weekly window %+v", windows[1])
	}

	_, err = fetchClaudeUsage(context.Background(), server.Client(), server.URL, "stale")
	var unavailable UsageUnavailable
	if !errors.As(err, &unavailable) || unavailable.Reason != "expired" {
		t.Fatalf("want expired, got %v", err)
	}
}

func TestClaudeUsageWithSetupTokenIsUnavailable(t *testing.T) {
	cli := fakeCLI(t, "claude")
	home, _ := cli.Prepare("s1")
	if err := cli.SaveClaudeToken(home, " sk-ant-oat01-x \n"); err != nil {
		t.Fatal(err)
	}
	if env := cli.Env(home); !slices.Contains(env, "CLAUDE_CODE_OAUTH_TOKEN=sk-ant-oat01-x") {
		t.Fatalf("token not passed to the CLI: %v", env)
	}
	driver, _ := cli.Driver("claude")
	_, err := driver.Usage(context.Background(), home)
	var unavailable UsageUnavailable
	if !errors.As(err, &unavailable) || unavailable.Reason != "token_login" {
		t.Fatalf("want token_login, got %v", err)
	}
	if info, err := os.Stat(filepath.Join(home, tokenFile)); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("token file must be private: %v", err)
	}
}

func TestRemoveRefusesPaths(t *testing.T) {
	cli := CLI{Root: t.TempDir()}
	for _, id := range []string{"", "..", "a/b"} {
		if cli.Remove(id) == nil {
			t.Fatalf("%q must be refused", id)
		}
	}
}

func TestAgentBinaryFindsCLIsInstalledAfterTheFirstLookup(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", t.TempDir())
	cli := CLI{}
	if got := cli.KimiPath(); got != "kimi" {
		t.Fatalf("got %q before install", got)
	}
	path := filepath.Join(home, ".kimi-code", "bin", "kimi")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := cli.KimiPath(); got != path {
		t.Fatalf("got %q, want %q", got, path)
	}
}

func TestMimoPathPrefersConfiguredPath(t *testing.T) {
	t.Setenv("HOME", "/srv/home")
	if got := (CLI{Mimo: "$HOME/custom/mimo"}).MimoPath(); got != "/srv/home/custom/mimo" {
		t.Fatalf("got %q", got)
	}
}

func TestCheckClaudeTokenRejectsOnlyUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "Bearer bad":
			w.WriteHeader(http.StatusUnauthorized)
		case "Bearer narrow":
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer server.Close()
	if err := CheckClaudeToken(context.Background(), server.Client(), server.URL, "bad"); !errors.Is(err, ErrTokenRejected) {
		t.Fatalf("want rejection, got %v", err)
	}
	for _, token := range []string{"good", "narrow"} {
		if err := CheckClaudeToken(context.Background(), server.Client(), server.URL, token); err != nil {
			t.Fatalf("%s: %v", token, err)
		}
	}
}

package agents

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestKimiDeviceLoginFinishesByItself(t *testing.T) {
	cli := fakeCLI(t, "kimi")
	home, err := cli.Prepare("s1")
	if err != nil {
		t.Fatal(err)
	}
	driver, _ := cli.Driver("kimi")
	login, err := driver.StartLogin(home)
	if err != nil {
		t.Fatal(err)
	}
	if login.NeedsCode || login.UserCode != "KO87-A7XC" || !strings.Contains(login.URL, "authorize_device") {
		t.Fatalf("unexpected login %+v", login)
	}
	if err := waitLogin(t, login); err != nil {
		t.Fatal(err)
	}
}

func TestMimoLoginAcceptsCode(t *testing.T) {
	cli := fakeCLI(t, "mimo")
	home, err := cli.Prepare("s1")
	if err != nil {
		t.Fatal(err)
	}
	driver, _ := cli.Driver("mimo")
	login, err := driver.StartLogin(home)
	if err != nil {
		t.Fatal(err)
	}
	if !login.NeedsCode || !strings.Contains(login.URL, "platform.xiaomimimo.com/authorize") {
		t.Fatalf("unexpected login %+v", login)
	}
	if err := login.Submit("good-code"); err != nil {
		t.Fatal(err)
	}
	if err := waitLogin(t, login); err != nil {
		t.Fatal(err)
	}
	account, err := driver.Account(context.Background(), home)
	if err != nil || account.Email != "6890014132" || account.Plan != "MiMo" {
		t.Fatalf("account %+v, %v", account, err)
	}
}

func TestMimoLoginReportsRejectedCode(t *testing.T) {
	cli := fakeCLI(t, "mimo")
	home, _ := cli.Prepare("s1")
	driver, _ := cli.Driver("mimo")
	login, err := driver.StartLogin(home)
	if err != nil {
		t.Fatal(err)
	}
	_ = login.Submit("bad")
	err = waitLogin(t, login)
	if err == nil || !strings.Contains(err.Error(), "wrong code") {
		t.Fatalf("want the CLI message, got %v", err)
	}
}

func TestMimoUsageIsUnsupported(t *testing.T) {
	cli := fakeCLI(t, "mimo")
	home, _ := cli.Prepare("s1")
	driver, _ := cli.Driver("mimo")
	_, err := driver.Usage(context.Background(), home)
	var unavailable UsageUnavailable
	if !errors.As(err, &unavailable) || unavailable.Reason != "unsupported" {
		t.Fatalf("want unsupported, got %v", err)
	}
}

func TestFetchKimiUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/usages" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer good" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, `{"usages":{"limit_5h":{"used_ratio":0.42,"reset_time":"2026-11-19T10:00:00Z"},
			"limit_7d":{"used_ratio":0.1,"reset_time":""},"limit_month_total":{"used_ratio":0.7},"limit_month_code":{"used_ratio":0.5}}}`)
	}))
	defer server.Close()

	windows, err := fetchKimiUsage(context.Background(), server.Client(), server.URL, "good")
	if err != nil {
		t.Fatal(err)
	}
	if len(windows) != 4 {
		t.Fatalf("windows %+v", windows)
	}
	byID := map[string]Window{}
	for _, window := range windows {
		byID[window.ID] = window
	}
	five := byID["limit_5h"]
	if five.UsedPercent != 42 || five.WindowMinutes != 300 || five.Label != "5h" || five.ResetTime == nil {
		t.Fatalf("five hour window %+v", five)
	}
	if byID["limit_month_total"].WindowMinutes != 30*24*60 || byID["limit_7d"].ResetTime != nil {
		t.Fatalf("windows %+v", byID)
	}

	_, err = fetchKimiUsage(context.Background(), server.Client(), server.URL, "stale")
	if !errors.Is(err, ErrKimiUnauthorized) {
		t.Fatalf("want unauthorized, got %v", err)
	}
}

func TestFetchKimiAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer good" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, `{"user_id":"u1","nickname":"dev","email":"dev@example.com","user_level_name":"Allegretto"}`)
	}))
	defer server.Close()
	user, err := fetchKimiAccount(context.Background(), server.Client(), server.URL, "good")
	if err != nil || user.Email != "dev@example.com" || user.UserLevelName != "Allegretto" {
		t.Fatalf("user %+v, %v", user, err)
	}
	if _, err := fetchKimiAccount(context.Background(), server.Client(), server.URL, "stale"); !errors.Is(err, ErrKimiUnauthorized) {
		t.Fatalf("want unauthorized, got %v", err)
	}
}

func TestRefreshKimiToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != "rt-good" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, `{"access_token":"at-new","refresh_token":"rt-new","expires_in":900}`)
	}))
	defer server.Close()
	creds, err := refreshKimiToken(context.Background(), server.Client(), server.URL, "rt-good")
	if err != nil {
		t.Fatal(err)
	}
	if creds.AccessToken != "at-new" || creds.RefreshToken != "rt-new" {
		t.Fatalf("creds %+v", creds)
	}
	if creds.expired(time.Now().Add(800 * time.Second)) {
		t.Fatalf("token must live 900s: %+v", creds)
	}
	if _, err := refreshKimiToken(context.Background(), server.Client(), server.URL, "rt-bad"); !errors.Is(err, ErrKimiUnauthorized) {
		t.Fatalf("want unauthorized, got %v", err)
	}
}

func TestReadKimiConfig(t *testing.T) {
	home := t.TempDir()
	toml := `[providers."managed:kimi-code"]
type = "kimi"
base_url = "https://api.kimi.com/coding/v1"

[providers."managed:kimi-code".oauth]
storage = "file"
key = "oauth/kimi-code-env-abc"
oauth_host = "https://auth.kimi.com"
`
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(toml), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := readKimiConfig(home)
	if cfg.baseURL != "https://api.kimi.com/coding/v1" || cfg.oauthHost != "https://auth.kimi.com" || cfg.oauthKey != "oauth/kimi-code-env-abc" {
		t.Fatalf("config %+v", cfg)
	}
	empty := readKimiConfig(t.TempDir())
	if empty.baseURL != kimiDefaultBase || empty.oauthHost != kimiDefaultOAuth {
		t.Fatalf("defaults %+v", empty)
	}
}

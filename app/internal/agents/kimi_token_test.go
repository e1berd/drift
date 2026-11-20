package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEnsureKimiTokenStoresRefreshedTokenWhereItReadsIt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"access_token":"at-new","refresh_token":"rt-new","expires_in":900}`)
	}))
	defer server.Close()
	home := t.TempDir()
	toml := fmt.Sprintf(`[providers."managed:kimi-code".oauth]
storage = "file"
key = "oauth/kimi-code-env-abc"
oauth_host = %q
`, server.URL)
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(toml), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "credentials"), 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "credentials", "kimi-code-env-abc.json")
	expired := fmt.Sprintf(`{"access_token":"at-old","refresh_token":"rt-old","expires_at":%d,"scope":"kimi-code","token_type":"Bearer"}`, time.Now().Add(-time.Hour).Unix())
	if err := os.WriteFile(path, []byte(expired), 0o600); err != nil {
		t.Fatal(err)
	}

	token, err := ensureKimiToken(context.Background(), home)
	if err != nil {
		t.Fatal(err)
	}
	if token != "at-new" {
		t.Fatalf("token %q", token)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var stored map[string]any
	if err := json.Unmarshal(raw, &stored); err != nil {
		t.Fatal(err)
	}
	if stored["access_token"] != "at-new" || stored["refresh_token"] != "rt-new" || stored["token_type"] != "Bearer" || stored["scope"] != "kimi-code" {
		t.Fatalf("stored %v", stored)
	}
	again, err := ensureKimiToken(context.Background(), home)
	if err != nil || again != "at-new" {
		t.Fatalf("second read %q %v", again, err)
	}
}

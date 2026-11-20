package agents

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeMimoAuth(t *testing.T, content string) string {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".local", "share", "mimocode")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return home
}

func TestMimoAccountReadsXiaomiProviders(t *testing.T) {
	cases := map[string]string{
		"account":    `{"xiaomi":{"metadata":{"uid":"42"}}}`,
		"token plan": `{"xiaomi-token-plan-sgp":{"type":"api","key":"tp-x","metadata":{"uid":"42","base_url":"https://example.com"}}}`,
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			account, err := mimo{}.Account(context.Background(), writeMimoAuth(t, content))
			if err != nil {
				t.Fatal(err)
			}
			if account.Email != "42" || account.Plan != mimoPlan {
				t.Fatalf("unexpected account %+v", account)
			}
		})
	}
}

func TestMimoAccountRejectsOtherProviders(t *testing.T) {
	for _, content := range []string{`{}`, `{"openai":{"metadata":{"uid":"42"}}}`, `{"xiaomi":{"metadata":{}}}`, `not json`} {
		if _, err := (mimo{}).Account(context.Background(), writeMimoAuth(t, content)); err == nil {
			t.Fatalf("expected sign-in error for %s", content)
		}
	}
}

func TestMimoModelUsesSignedInProvider(t *testing.T) {
	cases := []struct {
		name, auth, requested, want string
	}{
		{"token plan default", `{"xiaomi-token-plan-sgp":{"type":"api","key":"k","metadata":{"uid":"42"}}}`, "default", "xiaomi-token-plan-sgp/mimo-v2.6-pro"},
		{"token plan empty", `{"xiaomi-token-plan-sgp":{"metadata":{"uid":"42"}}}`, "", "xiaomi-token-plan-sgp/mimo-v2.6-pro"},
		{"explicit model", `{"xiaomi-token-plan-sgp":{"metadata":{"uid":"42"}}}`, "xiaomi/mimo-v2.6-flash", "xiaomi/mimo-v2.6-flash"},
		{"xiaomi account", `{"xiaomi":{"metadata":{"uid":"42"}}}`, "default", "default"},
		{"not signed in", `not json`, "default", "default"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := MimoModel(writeMimoAuth(t, c.auth), c.requested); got != c.want {
				t.Fatalf("MimoModel = %q, want %q", got, c.want)
			}
		})
	}
}

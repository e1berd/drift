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

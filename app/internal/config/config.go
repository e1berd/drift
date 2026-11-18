package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

// Config holds the process settings read from the environment.
type Config struct {
	Addr            string
	PublicURL       *url.URL
	DatabaseURL     *url.URL
	KratosPublicURL *url.URL
	KratosAdminURL  *url.URL
	ViteDevURL      *url.URL
	WorkspaceRoot   string
	ChatsRoot       string
	AgentsRoot      string
	ClaudeBin       string
	CodexBin        string
	Shell           string
	AntigravityBin  string
	DeliveryRepo    string
	DeliveryAPIURL  *url.URL
}

// Load reads the configuration from environment variables, falling back to local development defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Addr: env("DRIFT_ADDR", ":8080"),
	}
	root, err := filepath.Abs(os.ExpandEnv(env("DRIFT_WORKSPACE_ROOT", "workspace")))
	if err != nil {
		return nil, fmt.Errorf("DRIFT_WORKSPACE_ROOT: %w", err)
	}
	cfg.WorkspaceRoot = root
	chatsRoot, err := filepath.Abs(os.ExpandEnv(env("DRIFT_CHATS_DIR", filepath.Join(root, ".chats"))))
	if err != nil {
		return nil, fmt.Errorf("DRIFT_CHATS_DIR: %w", err)
	}
	cfg.ChatsRoot = chatsRoot
	agentsRoot, err := filepath.Abs(os.ExpandEnv(env("DRIFT_AGENTS_DIR", "data/agents")))
	if err != nil {
		return nil, fmt.Errorf("DRIFT_AGENTS_DIR: %w", err)
	}
	cfg.AgentsRoot = agentsRoot
	cfg.ClaudeBin = env("DRIFT_CLAUDE_BIN", "claude")
	cfg.CodexBin = env("DRIFT_CODEX_BIN", "codex")
	cfg.Shell = os.Getenv("DRIFT_SHELL")
	cfg.AntigravityBin = env("DRIFT_AGY_BIN", "agy")
	cfg.DeliveryRepo = env("DRIFT_DELIVERY_REPO", "e1berd/drift")

	urls := []struct {
		dst      **url.URL
		key, def string
	}{
		{&cfg.PublicURL, "DRIFT_PUBLIC_URL", "http://localhost:8080"},
		{&cfg.DatabaseURL, "DATABASE_URL", ""},
		{&cfg.KratosPublicURL, "KRATOS_PUBLIC_URL", "http://localhost:4433"},
		{&cfg.KratosAdminURL, "KRATOS_ADMIN_URL", "http://localhost:4434"},
		{&cfg.ViteDevURL, "VITE_DEV_URL", "http://localhost:5173"},
		{&cfg.DeliveryAPIURL, "DRIFT_DELIVERY_API_URL", "https://api.github.com"},
	}
	for _, u := range urls {
		raw := os.ExpandEnv(env(u.key, u.def))
		if raw == "" {
			return nil, fmt.Errorf("%s is not set", u.key)
		}
		parsed, err := url.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", u.key, err)
		}
		*u.dst = parsed
	}

	return cfg, nil
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

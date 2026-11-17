package config

import (
	"fmt"
	"net/url"
	"os"
)

// Config holds the process settings read from the environment.
type Config struct {
	Addr            string
	PublicURL       *url.URL
	DatabaseURL     *url.URL
	KratosPublicURL *url.URL
	KratosAdminURL  *url.URL
	ViteDevURL      *url.URL
}

// Load reads the configuration from environment variables, falling back to local development defaults.
func Load() (*Config, error) {
	cfg := &Config{Addr: env("DRIFT_ADDR", ":8080")}

	urls := []struct {
		dst      **url.URL
		key, def string
	}{
		{&cfg.PublicURL, "DRIFT_PUBLIC_URL", "http://localhost:8080"},
		{&cfg.DatabaseURL, "DATABASE_URL", ""},
		{&cfg.KratosPublicURL, "KRATOS_PUBLIC_URL", "http://localhost:4433"},
		{&cfg.KratosAdminURL, "KRATOS_ADMIN_URL", "http://localhost:4434"},
		{&cfg.ViteDevURL, "VITE_DEV_URL", "http://localhost:5173"},
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

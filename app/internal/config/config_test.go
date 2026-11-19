package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAgentBinaryUsesOfficialInstallDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", t.TempDir())
	t.Setenv("DRIFT_KIMI_BIN", "")
	path := filepath.Join(home, ".kimi-code", "bin", "kimi")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := agentBinary("DRIFT_KIMI_BIN", "kimi", ".kimi-code", "bin", "kimi"); got != path {
		t.Fatalf("got %q, want %q", got, path)
	}
}

func TestAgentBinaryPrefersConfiguredPath(t *testing.T) {
	t.Setenv("DRIFT_MIMO_BIN", "$HOME/custom/mimo")
	t.Setenv("HOME", "/srv/home")
	if got := agentBinary("DRIFT_MIMO_BIN", "mimo", ".mimocode", "bin", "mimo"); got != "/srv/home/custom/mimo" {
		t.Fatalf("got %q", got)
	}
}

func TestWorkspaceRootDefaultsToAbsoluteWorkspaceFolder(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/drift")
	t.Setenv("DRIFT_WORKSPACE_ROOT", "")
	t.Chdir(t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(cfg.WorkspaceRoot) || filepath.Base(cfg.WorkspaceRoot) != "workspace" {
		t.Fatalf("got %q", cfg.WorkspaceRoot)
	}
}

func TestWorkspaceRootExpandsEnvironment(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/drift")
	t.Setenv("PROJECTS_HOME", "/srv")
	t.Setenv("DRIFT_WORKSPACE_ROOT", "$PROJECTS_HOME/projects")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WorkspaceRoot != "/srv/projects" {
		t.Fatalf("got %q", cfg.WorkspaceRoot)
	}
}

package config

import (
	"path/filepath"
	"testing"
)

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

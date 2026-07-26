package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultsWithoutFile(t *testing.T) {
	t.Setenv("MODEL_ORCHESTRATOR_CONFIG", filepath.Join(t.TempDir(), "missing.toml"))
	cfg, _, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProfile != "default" {
		t.Fatalf("DefaultProfile = %q", cfg.DefaultProfile)
	}
	if _, err := cfg.Profile(""); err != nil {
		t.Fatal(err)
	}
}

func TestStateDirEnvironmentWins(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "state")
	t.Setenv("MODEL_ORCHESTRATOR_STATE_DIR", dir)
	got, err := StateDir(Default())
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(dir)
	if got != abs {
		t.Fatalf("StateDir() = %q, want %q", got, abs)
	}
}

func TestLoadRejectsUnknownAdapter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	data := []byte("default_profile='default'\n[models.bad]\nadapter='missing'\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(path); err == nil {
		t.Fatal("Load() accepted an unknown adapter")
	}
}

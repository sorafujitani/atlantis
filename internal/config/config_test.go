package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultsWithoutFile(t *testing.T) {
	t.Setenv("ATLANTIS_CONFIG", filepath.Join(t.TempDir(), "missing.toml"))
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
	grok, err := cfg.Profile("grok")
	if err != nil {
		t.Fatal(err)
	}
	if grok.Executor != "grok" || cfg.Models["grok"].Adapter != "grok" {
		t.Fatalf("grok profile = %#v, model = %#v", grok, cfg.Models["grok"])
	}
	omp, err := cfg.Profile("omp")
	if err != nil {
		t.Fatal(err)
	}
	if omp.Executor != "omp" || cfg.Models["omp"].Adapter != "omp" || cfg.Adapters["omp"].Command != "omp" {
		t.Fatalf("omp profile = %#v, model = %#v, adapter = %#v", omp, cfg.Models["omp"], cfg.Adapters["omp"])
	}
}

func TestStateDirEnvironmentWins(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "state")
	t.Setenv("ATLANTIS_STATE_DIR", dir)
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

func TestLoadStateDirIgnoresDormantRoutingValidation(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	path := filepath.Join(dir, "config.toml")
	data := []byte("state_dir = '" + stateDir + "'\ndefault_profile = 'missing'\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := LoadStateDir(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != stateDir {
		t.Fatalf("LoadStateDir() = %q, want %q", got, stateDir)
	}
}

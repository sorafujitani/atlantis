package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sorafujitani/atlantis/internal/orchestration"
	"github.com/sorafujitani/atlantis/internal/state"
	"github.com/spf13/cobra"
)

func TestVersionCommand(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"version"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "atlantis") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestBrainLifecycle(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"brain", "--dir", root, "init"}); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"brain", "--dir", root, "check"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "reachable") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestBrainRejectsRemovedInjectCommand(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"brain", "inject"})
	if err == nil {
		t.Fatal("removed brain inject command was accepted")
	}
}

func TestModelExecutionCommandsAreDisabled(t *testing.T) {
	t.Parallel()
	for _, arguments := range [][]string{
		{"run", "--mode", "single", "ignored"},
		{"resume", "run_ignored"},
		{"eval", "ignored.json"},
	} {
		var stdout, stderr bytes.Buffer
		err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, arguments)
		if err == nil || !strings.Contains(err.Error(), errModelRoutingDisabled.Error()) {
			t.Fatalf("%v error = %v", arguments, err)
		}
	}
}

func TestDormantModelCommandsRemainConstructible(t *testing.T) {
	t.Parallel()
	a := &app{}
	for _, command := range []*cobra.Command{
		a.runCommand(),
		a.resumeCommand(),
		a.evalCommand(),
	} {
		if command == nil {
			t.Fatal("dormant model command is nil")
		}
	}
}

func TestStatusIgnoresDormantRoutingValidation(t *testing.T) {
	t.Parallel()
	stateDir := t.TempDir()
	store, err := state.New(stateDir)
	if err != nil {
		t.Fatal(err)
	}
	runID, err := store.CreateRun(orchestration.Request{Objective: "historical", CWD: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(t.TempDir(), "config.toml")
	data := []byte("state_dir = '" + stateDir + "'\ndefault_profile = 'missing'\n")
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"--config", configPath, "status", runID}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), runID) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestDoctorJSON(t *testing.T) {
	t.Setenv("ATLANTIS_STATE_DIR", filepath.Join(t.TempDir(), "state"))
	t.Setenv("ATLANTIS_CONFIG", filepath.Join(t.TempDir(), "missing.toml"))
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"--output", "json", "doctor"}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout.String()), "[") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

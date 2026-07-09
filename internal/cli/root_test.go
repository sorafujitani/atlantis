package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"version"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "model-orchestrator") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestDoctorJSON(t *testing.T) {
	t.Setenv("MODEL_ORCHESTRATOR_STATE_DIR", filepath.Join(t.TempDir(), "state"))
	t.Setenv("MODEL_ORCHESTRATOR_CONFIG", filepath.Join(t.TempDir(), "missing.toml"))
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"--output", "json", "doctor"}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout.String()), "[") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

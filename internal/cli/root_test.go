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

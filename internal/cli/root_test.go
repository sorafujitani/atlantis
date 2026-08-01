package cli

import (
	"bytes"
	"context"
	"os"
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

func TestBrainContextCommand(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"brain", "--dir", root, "init"}); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"brain", "--dir", root, "context"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "HARD SAFETY:") || !strings.Contains(stdout.String(), "# Brain") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestBrainContextJSONIncludesFingerprint(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"brain", "--dir", root, "init"}); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"-o", "json", "brain", "--dir", root, "context"}); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if !strings.Contains(out, `"fingerprint"`) || !strings.Contains(out, `"context"`) {
		t.Fatalf("stdout = %q", out)
	}
	stdout.Reset()
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"brain", "--dir", root, "context", "--print-fingerprint"}); err != nil {
		t.Fatal(err)
	}
	fp := strings.TrimSpace(stdout.String())
	if len(fp) != 64 {
		t.Fatalf("fingerprint = %q", fp)
	}
}

func TestIntegrationsInstallCommand(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	legacy := filepath.Join(dir, "atlantis-brain.ts")
	if err := os.WriteFile(legacy, []byte("legacy"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"integrations", "install", "pi", "--dir", dir}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "atlantis-brain.js")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy adapter still exists: %v", err)
	}
}

func TestBrainJSONOutput(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"--output", "json", "brain", "--dir", t.TempDir(), "init"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout.String()), "{") {
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

func TestRemovedSupervisorCommandsAreUnknown(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"run", "resume", "eval", "status", "inspect", "cancel", "doctor"} {
		var stdout, stderr bytes.Buffer
		err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{name})
		if err == nil || !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("%s error = %v", name, err)
		}
	}
}

func TestRemovedConfigFlagIsUnknown(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := Execute(context.Background(), strings.NewReader(""), &stdout, &stderr, []string{"--config", "unused.toml", "version"})
	if err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("error = %v", err)
	}
}

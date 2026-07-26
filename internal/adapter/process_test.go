package adapter

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sorafujitani/atlantis/internal/config"
	"github.com/sorafujitani/atlantis/internal/orchestration"
)

func TestGenericAdapterDoesNotUseShell(t *testing.T) {
	t.Parallel()
	runner, err := NewProcessRunner("custom", config.Adapter{Command: "printf", Args: []string{"%s", "{prompt}"}})
	if err != nil {
		t.Fatal(err)
	}
	invocation := Invocation{Assignment: orchestration.Assignment{CWD: t.TempDir(), Permission: orchestration.PermissionRead}, Prompt: "$(touch /tmp/should-not-run)"}
	args, _, err := runner.args(invocation, "schema", "{}")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(args, " ") != "%s $(touch /tmp/should-not-run)" {
		t.Fatalf("args = %#v", args)
	}
}

func TestCapabilitiesRejectCursorWrite(t *testing.T) {
	t.Parallel()
	runner, err := NewProcessRunner("cursor", config.Adapter{Command: "cursor-agent"})
	if err != nil {
		t.Fatal(err)
	}
	if runner.Capabilities().Write {
		t.Fatal("cursor unexpectedly supports write")
	}
}

func TestGrokArguments(t *testing.T) {
	t.Parallel()
	runner, err := NewProcessRunner("grok", config.Adapter{Command: "grok"})
	if err != nil {
		t.Fatal(err)
	}
	cwd := t.TempDir()
	args, stdin, err := runner.args(Invocation{
		Assignment: orchestration.Assignment{CWD: cwd, Permission: orchestration.PermissionRead},
		Prompt:     "inspect",
		Model:      "grok-4.5-build",
	}, "unused.json", `{"type":"object"}`)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{
		"--cwd", cwd,
		"--output-format", "json",
		"--json-schema", `{"type":"object"}`,
		"--permission-mode", "plan",
		"--no-memory", "--no-subagents",
		"--model", "grok-4.5-build",
		"-p", "inspect",
	}, "\x00")
	if got := strings.Join(args, "\x00"); got != want {
		t.Fatalf("args = %#v", args)
	}
	if stdin != "" {
		t.Fatalf("stdin = %q", stdin)
	}
	if !runner.Capabilities().Write || !runner.Capabilities().NativeSchema || !runner.Capabilities().Resume {
		t.Fatalf("capabilities = %#v", runner.Capabilities())
	}
}

func TestGrokResumeUsesWritePermission(t *testing.T) {
	t.Parallel()
	runner, err := NewProcessRunner("grok", config.Adapter{Command: "grok"})
	if err != nil {
		t.Fatal(err)
	}
	args, _, err := runner.args(Invocation{
		Assignment: orchestration.Assignment{CWD: t.TempDir(), Permission: orchestration.PermissionWrite},
		Prompt:     "continue",
		Session:    &orchestration.NativeSessionRef{SessionID: "session-123"},
	}, "unused.json", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"--permission-mode acceptEdits", "--resume session-123", "-p continue"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("args %q do not contain %q", joined, expected)
		}
	}
}

func TestProcessRunnerCancellation(t *testing.T) {
	t.Parallel()
	runner, err := NewProcessRunner("custom", config.Adapter{Command: "sleep", Args: []string{"{prompt}"}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err = runner.Run(ctx, Invocation{
		Assignment: orchestration.Assignment{CWD: t.TempDir(), Permission: orchestration.PermissionRead},
		Prompt:     "5",
	})
	if err == nil || !strings.Contains(err.Error(), "cancelled") {
		t.Fatalf("Run() error = %v", err)
	}
}

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

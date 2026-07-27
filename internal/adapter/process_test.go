package adapter

import (
	"context"
	"os"
	"path/filepath"
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

func TestOMPArgumentsAndCapabilities(t *testing.T) {
	t.Parallel()
	runner, err := NewProcessRunner("omp", config.Adapter{Command: "omp"})
	if err != nil {
		t.Fatal(err)
	}
	caps := runner.Capabilities()
	if !caps.JSONStream || caps.NativeSchema || !caps.Resume || !caps.Usage || !caps.Read || !caps.Write || caps.FinalTextOnly {
		t.Fatalf("capabilities = %#v", caps)
	}

	schemaJSON := `{"type":"object"}`
	cases := []struct {
		name       string
		permission orchestration.Permission
		model      string
		session    *orchestration.NativeSessionRef
		wantParts  []string
		wantTools  string
		wantMode   string
	}{
		{
			name:       "read",
			permission: orchestration.PermissionRead,
			model:      "anthropic/claude-sonnet-4-5",
			wantParts:  []string{"-p", "--mode", "json", "--approval-mode", "always-ask", "--model", "anthropic/claude-sonnet-4-5", "inspect"},
			wantTools:  ompReadTools,
			wantMode:   "always-ask",
		},
		{
			name:       "write-resume",
			permission: orchestration.PermissionWrite,
			session:    &orchestration.NativeSessionRef{SessionID: "omp-session-1"},
			wantParts:  []string{"-p", "--mode", "json", "--approval-mode", "yolo", "--resume", "omp-session-1", "continue"},
			wantTools:  ompWriteTools,
			wantMode:   "yolo",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cwd := t.TempDir()
			prompt := "inspect"
			if tc.permission == orchestration.PermissionWrite {
				prompt = "continue"
			}
			args, stdin, err := runner.args(Invocation{
				Assignment: orchestration.Assignment{CWD: cwd, Permission: tc.permission},
				Prompt:     prompt,
				Model:      tc.model,
				Session:    tc.session,
			}, "unused.json", schemaJSON)
			if err != nil {
				t.Fatal(err)
			}
			if stdin != "" {
				t.Fatalf("stdin = %q", stdin)
			}
			joined := strings.Join(args, "\x00")
			for _, part := range tc.wantParts {
				if !strings.Contains(joined, part) && !containsArg(args, part) {
					t.Fatalf("args %#v missing %q", args, part)
				}
			}
			if !containsPair(args, "--tools", tc.wantTools) {
				t.Fatalf("args %#v missing tools %q", args, tc.wantTools)
			}
			if !containsPair(args, "--approval-mode", tc.wantMode) {
				t.Fatalf("args %#v missing approval-mode %q", args, tc.wantMode)
			}
			if !containsPair(args, "--cwd", cwd) {
				t.Fatalf("args %#v missing cwd", args)
			}
			if !containsPair(args, "--append-system-prompt", ompResultSchemaInstruction(schemaJSON)) {
				t.Fatalf("args %#v missing schema system instruction", args)
			}
			toolSet := "," + tc.wantTools + ","
			for _, excluded := range []string{"task", "hub", "retain", "recall", "reflect", "memory_edit", "learn", "manage_skill"} {
				if strings.Contains(toolSet, ","+excluded+",") {
					t.Fatalf("write/read tools unexpectedly include %q: %s", excluded, tc.wantTools)
				}
			}
			if tc.permission == orchestration.PermissionRead {
				for _, forbidden := range []string{"bash", "edit", "write", "eval"} {
					if strings.Contains(toolSet, ","+forbidden+",") {
						t.Fatalf("read tools include mutating tool %q: %s", forbidden, tc.wantTools)
					}
				}
			}
		})
	}
}

func TestOMPRunnerEndToEndWithFakeCLI(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	argsPath := filepath.Join(dir, "args")
	commandPath := filepath.Join(dir, "omp")
	script := `#!/bin/sh
set -eu
args_path="$1"
shift
printf '%s\n' "$@" > "$args_path"
printf '%s\n' \
  '{"type":"session","id":"omp-session"}' \
  '{"type":"message_end","message":{"role":"assistant","content":[{"type":"text","text":"{\"status\":\"completed\",\"summary\":\"done\",\"output\":\"verified\",\"advice_request\":null,\"tasks\":[]}"}],"usage":{"input":13,"output":5,"cost":{"total":0.01}}}}'
`
	if err := os.WriteFile(commandPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(commandPath, 0o700); err != nil { //nolint:gosec // executable fixture must be runnable by the test process
		t.Fatal(err)
	}
	runner, err := NewProcessRunner("omp", config.Adapter{Command: commandPath, Args: []string{argsPath}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), Invocation{
		Assignment: orchestration.Assignment{CWD: dir, Permission: orchestration.PermissionRead},
		Prompt:     "inspect",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Result.Output != "verified" || result.Session == nil || result.Session.SessionID != "omp-session" {
		t.Fatalf("result = %#v", result)
	}
	if result.Result.Usage.InputTokens != 13 || result.Result.Usage.OutputTokens != 5 || result.Result.Usage.CostUSD != 0.01 {
		t.Fatalf("usage = %#v", result.Result.Usage)
	}
	args, err := os.ReadFile(argsPath) //nolint:gosec // path is created inside this test's private temporary directory
	if err != nil {
		t.Fatal(err)
	}
	argumentLines := strings.Split(strings.TrimSpace(string(args)), "\n")
	if !containsPair(argumentLines, "--tools", ompReadTools) || !containsPair(argumentLines, "--approval-mode", "always-ask") {
		t.Fatalf("args = %#v", argumentLines)
	}
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func containsPair(args []string, flag, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

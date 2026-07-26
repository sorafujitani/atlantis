package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sorafujitani/atlantis/internal/config"
	"github.com/sorafujitani/atlantis/internal/orchestration"
)

type ProcessRunner struct {
	name string
	cfg  config.Adapter
	caps Capabilities
}

func NewProcessRunner(name string, cfg config.Adapter) (*ProcessRunner, error) {
	if strings.TrimSpace(cfg.Command) == "" {
		return nil, errors.New("adapter command is required")
	}
	caps := capabilitiesFor(name, cfg.Args)
	return &ProcessRunner{name: name, cfg: cfg, caps: caps}, nil
}

func (r *ProcessRunner) Name() string               { return r.name }
func (r *ProcessRunner) Capabilities() Capabilities { return r.caps }

func (r *ProcessRunner) Run(ctx context.Context, invocation Invocation) (orchestration.ExecutionResult, error) {
	if invocation.Assignment.Permission == orchestration.PermissionWrite && !r.caps.Write {
		return orchestration.ExecutionResult{}, fmt.Errorf("adapter %q does not support write assignments", r.name)
	}
	if invocation.Session != nil && !r.caps.Resume {
		return orchestration.ExecutionResult{}, fmt.Errorf("adapter %q does not support native resume", r.name)
	}
	schemaFile, schemaJSON, cleanup, err := resultSchemaFile()
	if err != nil {
		return orchestration.ExecutionResult{}, err
	}
	defer cleanup()
	args, stdin, err := r.args(invocation, schemaFile, schemaJSON)
	if err != nil {
		return orchestration.ExecutionResult{}, err
	}
	// The command is a validated local adapter executable and args are passed without a shell.
	command := exec.CommandContext(ctx, r.cfg.Command, args...) //nolint:gosec // user-configured local CLI is the intended trust boundary
	command.Dir = invocation.Assignment.CWD
	command.Stdin = strings.NewReader(stdin)
	command.Env = safeEnvironment(r.cfg.InheritEnv)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.WaitDelay = 2 * time.Second
	command.Cancel = func() error {
		if command.Process == nil {
			return nil
		}
		if err := syscall.Kill(-command.Process.Pid, syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return err
		}
		return nil
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return orchestration.ExecutionResult{}, fmt.Errorf("%s cancelled: %w", r.name, ctx.Err())
		}
		message := strings.TrimSpace(stderr.String())
		if stdoutMessage := strings.TrimSpace(stdout.String()); stdoutMessage != "" {
			if message != "" {
				message += "\n"
			}
			message += stdoutMessage
		}
		if len(message) > 1200 {
			message = message[len(message)-1200:]
		}
		return orchestration.ExecutionResult{}, fmt.Errorf("%s failed: %w: %s", r.name, err, message)
	}
	result, err := ParseOutput(r.name, stdout.Bytes(), r.caps.FinalTextOnly)
	if err != nil {
		return orchestration.ExecutionResult{}, fmt.Errorf("parse %s output: %w", r.name, err)
	}
	if result.Session != nil && result.Session.Adapter == "" {
		result.Session.Adapter = r.name
	}
	return result, nil
}

func (r *ProcessRunner) args(inv Invocation, schemaFile, schemaJSON string) ([]string, string, error) {
	base := append([]string{}, r.cfg.Args...)
	resumeID := ""
	if inv.Session != nil {
		resumeID = inv.Session.SessionID
	}
	switch r.name {
	case "codex":
		if resumeID != "" {
			args := append(base, "exec", "resume", "--json", "--output-schema", schemaFile)
			if inv.Model != "" {
				args = append(args, "--model", inv.Model)
			}
			args = append(args, resumeID, "-")
			return args, inv.Prompt, nil
		}
		args := append(base, "exec", "-C", inv.Assignment.CWD, "--json", "--output-schema", schemaFile)
		if inv.Assignment.Permission == orchestration.PermissionWrite {
			args = append(args, "-s", "workspace-write")
		} else {
			args = append(args, "-s", "read-only")
		}
		if inv.Model != "" {
			args = append(args, "--model", inv.Model)
		}
		args = append(args, "-")
		return args, inv.Prompt, nil
	case "claude":
		mode := "plan"
		if inv.Assignment.Permission == orchestration.PermissionWrite {
			mode = "acceptEdits"
		}
		args := append(base, "-p", "--output-format", "stream-json", "--verbose", "--json-schema", schemaJSON, "--permission-mode", mode)
		if inv.Model != "" {
			args = append(args, "--model", inv.Model)
		}
		if resumeID != "" {
			args = append(args, "--resume", resumeID)
		}
		args = append(args, inv.Prompt)
		return args, "", nil
	case "opencode":
		args := append(base, "run", "--pure", "--format", "json", "--dir", inv.Assignment.CWD)
		if inv.Model != "" {
			args = append(args, "--model", inv.Model)
		}
		if resumeID != "" {
			args = append(args, "--session", resumeID)
		}
		args = append(args, inv.Prompt)
		return args, "", nil
	case "cursor":
		args := append(base, "-p", "--workspace", inv.Assignment.CWD, "--output-format", "stream-json", "--trust", "--mode", "plan")
		if inv.Model != "" {
			args = append(args, "--model", inv.Model)
		}
		if resumeID != "" {
			args = append(args, "--resume", resumeID)
		}
		args = append(args, inv.Prompt)
		return args, "", nil
	case "copilot":
		args := append(base, "-p", inv.Prompt, "-s", "--no-ask-user", "--no-custom-instructions", "--disable-builtin-mcps")
		if inv.Model != "" {
			args = append(args, "--model", inv.Model)
		}
		if resumeID != "" {
			args = append(args, "--resume", resumeID)
		}
		return args, "", nil
	case "grok":
		permissionMode := "plan"
		if inv.Assignment.Permission == orchestration.PermissionWrite {
			permissionMode = "acceptEdits"
		}
		args := append(base,
			"--cwd", inv.Assignment.CWD,
			"--output-format", "json",
			"--json-schema", schemaJSON,
			"--permission-mode", permissionMode,
			"--no-memory",
			"--no-subagents",
		)
		if inv.Model != "" {
			args = append(args, "--model", inv.Model)
		}
		if resumeID != "" {
			args = append(args, "--resume", resumeID)
		}
		args = append(args, "-p", inv.Prompt)
		return args, "", nil
	default:
		args := make([]string, 0, len(base)+1)
		hasPrompt := false
		for _, value := range base {
			replaced := strings.NewReplacer(
				"{prompt}", inv.Prompt,
				"{cwd}", inv.Assignment.CWD,
				"{model}", inv.Model,
				"{session}", resumeID,
			).Replace(value)
			if strings.Contains(value, "{prompt}") {
				hasPrompt = true
			}
			args = append(args, replaced)
		}
		if !hasPrompt {
			args = append(args, inv.Prompt)
		}
		return args, "", nil
	}
}

func capabilitiesFor(name string, args []string) Capabilities {
	switch name {
	case "codex", "claude":
		return Capabilities{JSONStream: true, NativeSchema: true, Resume: true, Usage: true, Read: true, Write: true}
	case "grok":
		return Capabilities{NativeSchema: true, Resume: true, Usage: true, Read: true, Write: true}
	case "opencode":
		return Capabilities{JSONStream: true, Resume: true, Usage: true, Read: true, Write: false}
	case "cursor":
		return Capabilities{JSONStream: true, Resume: true, Read: true, Write: false}
	case "copilot":
		return Capabilities{Resume: true, Read: true, FinalTextOnly: true}
	default:
		joined := strings.Join(args, "\x00")
		return Capabilities{Read: true, Resume: strings.Contains(joined, "{session}"), FinalTextOnly: true}
	}
}

func safeEnvironment(extra []string) []string {
	allowed := map[string]bool{
		"PATH": true, "HOME": true, "USER": true, "LOGNAME": true, "SHELL": true,
		"TMPDIR": true, "LANG": true, "LC_ALL": true, "TERM": true,
		"XDG_CONFIG_HOME": true, "XDG_DATA_HOME": true, "XDG_STATE_HOME": true, "XDG_CACHE_HOME": true,
		"CODEX_HOME": true, "CLAUDE_CONFIG_DIR": true, "ATLANTIS_BRAIN_DIR": true,
	}
	for _, name := range extra {
		allowed[name] = true
	}
	var result []string
	for _, item := range os.Environ() {
		name, _, ok := strings.Cut(item, "=")
		if ok && allowed[name] {
			result = append(result, item)
		}
	}
	return result
}

func resultSchemaFile() (string, string, func(), error) {
	schema, err := json.Marshal(resultSchema())
	if err != nil {
		return "", "", func() {}, err
	}
	dir, err := os.MkdirTemp("", "atlantis-schema-")
	if err != nil {
		return "", "", func() {}, err
	}
	path := filepath.Join(dir, "result.json")
	if err := os.WriteFile(path, schema, 0o600); err != nil {
		_ = os.RemoveAll(dir)
		return "", "", func() {}, err
	}
	return path, string(schema), func() { _ = os.RemoveAll(dir) }, nil
}

func resultSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"status", "summary", "output", "advice_request", "tasks"},
		"properties": map[string]any{
			"status":  map[string]any{"type": "string", "enum": []string{"completed", "needs_advice", "failed"}},
			"summary": map[string]any{"type": "string"},
			"output":  map[string]any{"type": []string{"string", "null"}},
			"advice_request": map[string]any{
				"type":                 []string{"object", "null"},
				"additionalProperties": false,
				"required":             []string{"question", "reason", "facts", "options", "constraints"},
				"properties": map[string]any{
					"question": map[string]any{"type": "string"}, "reason": map[string]any{"type": "string"},
					"facts":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"options":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"constraints": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
			},
			"tasks": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object", "additionalProperties": false,
					"required": []string{"id", "objective", "depends_on", "role", "permission"},
					"properties": map[string]any{
						"id": map[string]any{"type": "string"}, "objective": map[string]any{"type": "string"},
						"depends_on": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"role":       map[string]any{"type": "string", "enum": []string{"worker", "reviewer", "executor"}},
						"permission": map[string]any{"type": "string", "enum": []string{"read", "write"}},
					},
				},
			},
		},
	}
}

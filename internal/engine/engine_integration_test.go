package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sorafujitani/atlantis/internal/config"
	"github.com/sorafujitani/atlantis/internal/orchestration"
	"github.com/sorafujitani/atlantis/internal/state"
)

func TestEngineHybridEndToEndWithLocalProcess(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	agent := filepath.Join(dir, "fake-agent")
	script := `#!/bin/sh
prompt="$*"
case "$prompt" in
  *"Act only as an orchestrator"*)
    printf '%s\n' '{"status":"completed","summary":"planned","tasks":[{"id":"research","objective":"research","role":"worker","permission":"read"},{"id":"apply","objective":"apply","depends_on":["research"],"role":"executor","permission":"read"}]}'
    ;;
  *"Synthesize the worker results"*)
    printf '%s\n' '{"status":"completed","summary":"synthesized","output":"FINAL"}'
    ;;
  *"Advice request"*)
    printf '%s\n' '{"status":"completed","summary":"advised","output":"Choose option A"}'
    ;;
  *"Continue the original assignment"*)
    printf '%s\n' '{"status":"completed","summary":"resumed","output":"APPLIED"}'
    ;;
  *"Advice is available"*)
    printf '%s\n' '{"status":"needs_advice","summary":"uncertain","advice_request":{"question":"A or B?","reason":"consequential"}}'
    ;;
  *)
    printf '%s\n' '{"status":"completed","summary":"done","output":"DONE"}'
    ;;
esac
`
	if err := os.WriteFile(agent, []byte(script), 0o700); err != nil { //nolint:gosec // executable fixture must be runnable
		t.Fatal(err)
	}
	cfg := config.Config{
		DefaultProfile: "test",
		Adapters:       map[string]config.Adapter{"fake": {Command: agent, Args: []string{"{prompt}"}}},
		Models:         map[string]config.Model{"premium": {Adapter: "fake"}, "standard": {Adapter: "fake"}},
		Profiles: map[string]config.Profile{"test": {
			Mode: orchestration.ModeHybrid, Orchestrator: "premium", Executor: "standard", Advisor: "premium", Worker: "standard", Reviewer: "premium",
			MaxWorkers: 2, MaxCalls: 20, MaxAdvisorCalls: 2, MaxRetries: 0, MaxDuration: time.Minute,
		}},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	store, err := state.New(filepath.Join(dir, "state"))
	if err != nil {
		t.Fatal(err)
	}
	orchestrator := New(cfg, store)
	runID, result, err := orchestrator.Start(context.Background(), orchestration.Request{Objective: "test", CWD: dir, Profile: "test", Mode: orchestration.ModeHybrid})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "FINAL" {
		t.Fatalf("result = %#v", result)
	}
	snapshot, err := orchestrator.Snapshot(runID)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Status != "completed" || len(snapshot.CompletedTasks) != 4 {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	for taskID, completed := range snapshot.CompletedTasks {
		if len(completed.Events) != 0 {
			t.Fatalf("task %s persisted raw provider events", taskID)
		}
	}
	resumed, err := orchestrator.Resume(context.Background(), runID)
	if err != nil {
		t.Fatal(err)
	}
	if resumed.Output != "FINAL" {
		t.Fatalf("resumed = %#v", resumed)
	}
}

func TestResumeDoesNotRepeatCompletedSynthesis(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := config.Config{
		DefaultProfile: "test",
		Adapters:       map[string]config.Adapter{"missing": {Command: filepath.Join(dir, "must-not-run")}},
		Models:         map[string]config.Model{"model": {Adapter: "missing"}},
		Profiles: map[string]config.Profile{"test": {
			Mode: orchestration.ModeOrchestrator, Orchestrator: "model", Executor: "model", Advisor: "model", Worker: "model", Reviewer: "model",
			MaxWorkers: 1, MaxCalls: 5, MaxAdvisorCalls: 0, MaxRetries: 0, MaxDuration: time.Minute,
		}},
	}
	store, err := state.New(filepath.Join(dir, "state"))
	if err != nil {
		t.Fatal(err)
	}
	runID, err := store.CreateRun(orchestration.Request{Objective: "test", CWD: dir, Profile: "test", Mode: orchestration.ModeOrchestrator, Permission: orchestration.PermissionRead})
	if err != nil {
		t.Fatal(err)
	}
	final := orchestration.ExecutionResult{Result: orchestration.Result{Status: orchestration.ResultCompleted, Summary: "done", Output: "FINAL"}}
	if err := store.Append(runID, state.EventTaskCompleted, "synthesis", final); err != nil {
		t.Fatal(err)
	}
	result, err := New(cfg, store).Resume(context.Background(), runID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "FINAL" {
		t.Fatalf("result = %#v", result)
	}
}

func TestEngineFallsBackToSecondaryAdapter(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	agent := filepath.Join(dir, "fallback-agent")
	if err := os.WriteFile(agent, []byte("#!/bin/sh\nprintf '%s\\n' '{\"status\":\"completed\",\"summary\":\"fallback\",\"output\":\"FALLBACK_OK\"}'\n"), 0o700); err != nil { //nolint:gosec // executable fixture must be runnable
		t.Fatal(err)
	}
	cfg := config.Config{
		DefaultProfile: "test",
		Adapters: map[string]config.Adapter{
			"missing":  {Command: filepath.Join(dir, "missing-agent")},
			"fallback": {Command: agent, Args: []string{"{prompt}"}},
		},
		Models: map[string]config.Model{
			"primary":   {Adapter: "missing", Fallback: []string{"secondary"}},
			"secondary": {Adapter: "fallback"},
		},
		Profiles: map[string]config.Profile{"test": {
			Mode: orchestration.ModeSingle, Orchestrator: "primary", Executor: "primary", Advisor: "primary", Worker: "primary", Reviewer: "primary",
			MaxWorkers: 1, MaxCalls: 3, MaxAdvisorCalls: 0, MaxRetries: 0, MaxDuration: time.Minute,
		}},
	}
	store, err := state.New(filepath.Join(dir, "state"))
	if err != nil {
		t.Fatal(err)
	}
	runID, result, err := New(cfg, store).Start(context.Background(), orchestration.Request{Objective: "test fallback", CWD: dir, Profile: "test", Mode: orchestration.ModeSingle})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "FALLBACK_OK" {
		t.Fatalf("result = %#v", result)
	}
	events, err := store.LoadEvents(runID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, event := range events {
		if event.Type == state.EventFallbackUsed {
			found = true
		}
	}
	if !found {
		t.Fatal("fallback event was not recorded")
	}
}

func TestValidatePermissionCeilingRejectsWriteTaskInReadOnlyRun(t *testing.T) {
	t.Parallel()
	tasks := []orchestration.Task{{ID: "mutate", Objective: "change a file", Role: orchestration.RoleWorker, Permission: orchestration.PermissionWrite}}
	if err := validatePermissionCeiling(tasks, orchestration.PermissionRead); err == nil {
		t.Fatal("expected read-only ceiling to reject a write task")
	}
	if err := validatePermissionCeiling(tasks, orchestration.PermissionWrite); err != nil {
		t.Fatalf("write ceiling rejected write task: %v", err)
	}
}

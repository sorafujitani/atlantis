package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sorafujitani/atlantis/internal/orchestration"
)

func TestStoreCreateAppendReplay(t *testing.T) {
	t.Parallel()
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runID, err := store.CreateRun(orchestration.Request{Objective: "test", CWD: "/tmp", Profile: "default"})
	if err != nil {
		t.Fatal(err)
	}
	result := orchestration.ExecutionResult{Result: orchestration.Result{Status: orchestration.ResultCompleted, Summary: "done"}}
	if err := store.Append(runID, EventTaskCompleted, "task-1", result); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(runID, EventRunCompleted, "", result.Result); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.Snapshot(runID)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Status != "completed" || snapshot.CompletedTasks["task-1"].Result.Summary != "done" {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	info, err := os.Stat(filepath.Join(store.Root(), "runs", runID, "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("event permissions = %o", info.Mode().Perm())
	}
}

func TestAcquireReclaimsStaleLock(t *testing.T) {
	t.Parallel()
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runID, err := store.CreateRun(orchestration.Request{Objective: "test", CWD: "/tmp", Profile: "default"})
	if err != nil {
		t.Fatal(err)
	}
	lock := filepath.Join(store.Root(), "runs", runID, "lock")
	if err := os.WriteFile(lock, []byte("999999999\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	release, err := store.Acquire(runID)
	if err != nil {
		t.Fatal(err)
	}
	if err := release(); err != nil {
		t.Fatal(err)
	}
}

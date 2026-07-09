package engine

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sorafujitani/model-orchestrator/internal/orchestration"
)

func TestRunGraphSerializesWrites(t *testing.T) {
	t.Parallel()
	tasks := []orchestration.Task{
		{ID: "read-a", Objective: "read", Role: orchestration.RoleWorker, Permission: orchestration.PermissionRead},
		{ID: "read-b", Objective: "read", Role: orchestration.RoleWorker, Permission: orchestration.PermissionRead},
		{ID: "write-a", Objective: "write", Role: orchestration.RoleExecutor, Permission: orchestration.PermissionWrite, DependsOn: []string{"read-a"}},
		{ID: "write-b", Objective: "write", Role: orchestration.RoleExecutor, Permission: orchestration.PermissionWrite, DependsOn: []string{"read-b"}},
	}
	var mu sync.Mutex
	activeWrites, maxWrites := 0, 0
	stream, err := RunGraph(context.Background(), tasks, nil, 2, func(_ context.Context, task orchestration.Task) (TaskOutcome, error) {
		if task.Permission == orchestration.PermissionWrite {
			mu.Lock()
			activeWrites++
			if activeWrites > maxWrites {
				maxWrites = activeWrites
			}
			mu.Unlock()
			time.Sleep(5 * time.Millisecond)
			mu.Lock()
			activeWrites--
			mu.Unlock()
		}
		return TaskOutcome{Result: orchestration.ExecutionResult{Result: orchestration.Result{Status: orchestration.ResultCompleted, Summary: "done"}}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for result := range stream {
		if result.Err != nil {
			t.Fatal(result.Err)
		}
		count++
	}
	if count != len(tasks) {
		t.Fatalf("completed %d tasks, want %d", count, len(tasks))
	}
	if maxWrites != 1 {
		t.Fatalf("max concurrent writes = %d", maxWrites)
	}
}

func TestRunGraphSkipsCompleted(t *testing.T) {
	t.Parallel()
	tasks := []orchestration.Task{{ID: "done", Objective: "done", Role: orchestration.RoleWorker, Permission: orchestration.PermissionRead}, {ID: "next", Objective: "next", Role: orchestration.RoleWorker, Permission: orchestration.PermissionRead, DependsOn: []string{"done"}}}
	completed := map[string]orchestration.ExecutionResult{"done": {Result: orchestration.Result{Status: orchestration.ResultCompleted, Summary: "done"}}}
	stream, err := RunGraph(context.Background(), tasks, completed, 1, func(_ context.Context, task orchestration.Task) (TaskOutcome, error) {
		if task.ID == "done" {
			t.Fatal("completed task reran")
		}
		return TaskOutcome{Result: completed["done"]}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for range stream {
		count++
	}
	if count != 1 {
		t.Fatalf("ran %d tasks", count)
	}
}

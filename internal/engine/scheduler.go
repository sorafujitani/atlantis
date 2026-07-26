package engine

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/sorafujitani/atlantis/internal/orchestration"
)

type TaskOutcome struct {
	Result orchestration.ExecutionResult
	Trace  []TraceEvent
}

type TraceEvent struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type TaskRunner func(context.Context, orchestration.Task) (TaskOutcome, error)

type ScheduledResult struct {
	Task    orchestration.Task
	Outcome TaskOutcome
	Err     error
}

func RunGraph(ctx context.Context, tasks []orchestration.Task, completed map[string]orchestration.ExecutionResult, maxWorkers int, runner TaskRunner) (<-chan ScheduledResult, error) {
	if err := orchestration.ValidateTasks(tasks); err != nil {
		return nil, err
	}
	if maxWorkers <= 0 {
		return nil, fmt.Errorf("maxWorkers must be positive")
	}
	output := make(chan ScheduledResult)
	go func() {
		defer close(output)
		send := func(result ScheduledResult) bool {
			select {
			case output <- result:
				return true
			case <-ctx.Done():
				return false
			}
		}
		done := make(map[string]bool, len(tasks))
		for id := range completed {
			done[id] = true
		}
		remaining := make(map[string]orchestration.Task, len(tasks))
		for _, task := range tasks {
			if !done[task.ID] {
				remaining[task.ID] = task
			}
		}
		for len(remaining) > 0 {
			if ctx.Err() != nil {
				return
			}
			ready := readyTasks(remaining, done)
			if len(ready) == 0 {
				send(ScheduledResult{Err: fmt.Errorf("task graph cannot make progress")})
				return
			}
			var reads []orchestration.Task
			var write *orchestration.Task
			for _, task := range ready {
				if task.Permission == orchestration.PermissionWrite {
					if write == nil {
						copied := task
						write = &copied
					}
				} else {
					reads = append(reads, task)
				}
			}
			if len(reads) > 0 {
				results := runReadWave(ctx, reads, maxWorkers, runner)
				for _, result := range results {
					if !send(result) {
						return
					}
					if result.Err != nil {
						return
					}
					done[result.Task.ID] = true
					delete(remaining, result.Task.ID)
				}
				continue
			}
			if write != nil {
				outcome, err := runner(ctx, *write)
				result := ScheduledResult{Task: *write, Outcome: outcome, Err: err}
				if !send(result) {
					return
				}
				if err != nil {
					return
				}
				done[write.ID] = true
				delete(remaining, write.ID)
			}
		}
	}()
	return output, nil
}

func readyTasks(remaining map[string]orchestration.Task, done map[string]bool) []orchestration.Task {
	var ready []orchestration.Task
	for _, task := range remaining {
		dependenciesDone := true
		for _, dependency := range task.DependsOn {
			if !done[dependency] {
				dependenciesDone = false
				break
			}
		}
		if dependenciesDone {
			ready = append(ready, task)
		}
	}
	sort.Slice(ready, func(i, j int) bool { return ready[i].ID < ready[j].ID })
	return ready
}

func runReadWave(ctx context.Context, tasks []orchestration.Task, maxWorkers int, runner TaskRunner) []ScheduledResult {
	results := make([]ScheduledResult, len(tasks))
	semaphore := make(chan struct{}, maxWorkers)
	var wait sync.WaitGroup
	for index, task := range tasks {
		index, task := index, task
		wait.Add(1)
		go func() {
			defer wait.Done()
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				results[index] = ScheduledResult{Task: task, Err: ctx.Err()}
				return
			}
			defer func() { <-semaphore }()
			outcome, err := runner(ctx, task)
			results[index] = ScheduledResult{Task: task, Outcome: outcome, Err: err}
		}()
	}
	wait.Wait()
	return results
}

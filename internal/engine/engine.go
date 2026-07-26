package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/sorafujitani/model-orchestrator/internal/adapter"
	"github.com/sorafujitani/model-orchestrator/internal/config"
	"github.com/sorafujitani/model-orchestrator/internal/orchestration"
	"github.com/sorafujitani/model-orchestrator/internal/state"
)

const plannerTaskID = "__plan"

type Engine struct {
	cfg     config.Config
	store   *state.Store
	factory *adapter.Factory
}

func New(cfg config.Config, store *state.Store) *Engine {
	return &Engine{cfg: cfg, store: store, factory: adapter.NewFactory(cfg)}
}

func (e *Engine) Start(ctx context.Context, request orchestration.Request) (string, orchestration.Result, error) {
	if err := request.Validate(); err != nil {
		return "", orchestration.Result{}, err
	}
	profile, err := e.cfg.Profile(request.Profile)
	if err != nil {
		return "", orchestration.Result{}, err
	}
	if request.Profile == "" {
		request.Profile = e.cfg.DefaultProfile
	}
	if request.Permission == "" {
		request.Permission = orchestration.PermissionRead
	}
	if request.Mode == "" {
		request.Mode = profile.Mode
	}
	runID, err := e.store.CreateRun(request)
	if err != nil {
		return "", orchestration.Result{}, err
	}
	result, err := e.executeRun(ctx, runID, request, profile, state.Snapshot{CompletedTasks: map[string]orchestration.ExecutionResult{}})
	return runID, result, err
}

func (e *Engine) Resume(ctx context.Context, runID string) (orchestration.Result, error) {
	snapshot, err := e.store.Snapshot(runID)
	if err != nil {
		return orchestration.Result{}, err
	}
	if snapshot.Status == "completed" && snapshot.FinalResult != nil {
		return *snapshot.FinalResult, nil
	}
	if snapshot.Request == nil {
		return orchestration.Result{}, errors.New("run has no request")
	}
	profile, err := e.cfg.Profile(snapshot.Request.Profile)
	if err != nil {
		return orchestration.Result{}, err
	}
	return e.executeRun(ctx, runID, *snapshot.Request, profile, snapshot)
}

func (e *Engine) executeRun(parent context.Context, runID string, request orchestration.Request, profile config.Profile, snapshot state.Snapshot) (result orchestration.Result, returnedErr error) {
	release, err := e.store.Acquire(runID)
	if err != nil {
		return orchestration.Result{}, err
	}
	defer func() { _ = release() }()
	ctx, cancel := context.WithTimeout(parent, profile.MaxDuration)
	defer cancel()
	ledger := NewLedger(profile.MaxCalls, profile.MaxAdvisorCalls)
	defer func() {
		if returnedErr == nil {
			return
		}
		eventType := state.EventRunFailed
		if errors.Is(returnedErr, context.Canceled) || errors.Is(returnedErr, context.DeadlineExceeded) {
			eventType = state.EventRunCancelled
		}
		_ = e.store.Append(runID, eventType, "", map[string]any{"error": returnedErr.Error(), "budget": ledger.Snapshot()})
	}()
	if existing, ok := snapshot.CompletedTasks["root"]; ok && len(snapshot.Plan) == 0 {
		result = existing.Result
		if err := e.store.Append(runID, state.EventRunCompleted, "", result); err != nil {
			return orchestration.Result{}, err
		}
		return result, nil
	}
	switch request.Mode {
	case orchestration.ModeSingle, orchestration.ModeAdvisor:
		outcome, err := e.runAssignment(ctx, runID, "root", request.Objective, orchestration.RoleExecutor, request.Permission, profile.Executor, request.Mode == orchestration.ModeAdvisor, profile, ledger)
		if err != nil {
			return orchestration.Result{}, err
		}
		if err := e.appendOutcome(runID, "root", outcome); err != nil {
			return orchestration.Result{}, err
		}
		result = outcome.Result.Result
	case orchestration.ModeOrchestrator, orchestration.ModeHybrid:
		result, err = e.runOrchestration(ctx, runID, request, profile, snapshot, ledger)
		if err != nil {
			return orchestration.Result{}, err
		}
	default:
		return orchestration.Result{}, fmt.Errorf("unsupported mode %q", request.Mode)
	}
	if err := e.store.Append(runID, state.EventRunCompleted, "", result); err != nil {
		return orchestration.Result{}, err
	}
	return result, nil
}

func (e *Engine) runOrchestration(ctx context.Context, runID string, request orchestration.Request, profile config.Profile, snapshot state.Snapshot, ledger *Ledger) (orchestration.Result, error) {
	if synthesized, exists := snapshot.CompletedTasks["synthesis"]; exists {
		return synthesized.Result, nil
	}
	plan := snapshot.Plan
	plannerResult, plannerExists := snapshot.CompletedTasks[plannerTaskID]
	if len(plan) == 0 && plannerExists && len(plannerResult.Result.Tasks) > 0 {
		plan = plannerResult.Result.Tasks
		if err := e.store.Append(runID, state.EventPlanCreated, "", plan); err != nil {
			return orchestration.Result{}, err
		}
	}
	if len(plan) == 0 {
		outcome, err := e.executeModel(ctx, orchestration.Assignment{RunID: runID, TaskID: plannerTaskID, Objective: request.Objective, Role: orchestration.RoleOrchestrator, Permission: orchestration.PermissionRead, ModelAlias: profile.Orchestrator, CWD: request.CWD}, planningPrompt(request.Objective, request.Mode == orchestration.ModeHybrid, request.Permission), nil, false, profile, ledger)
		if err != nil {
			return orchestration.Result{}, err
		}
		if len(outcome.Result.Result.Tasks) == 0 {
			return orchestration.Result{}, errors.New("orchestrator returned no tasks")
		}
		plannerResult = outcome.Result
		if err := e.appendOutcome(runID, plannerTaskID, outcome); err != nil {
			return orchestration.Result{}, err
		}
		plan = outcome.Result.Result.Tasks
		if err := e.store.Append(runID, state.EventPlanCreated, "", plan); err != nil {
			return orchestration.Result{}, err
		}
		plannerExists = true
	}
	if err := validatePermissionCeiling(plan, request.Permission); err != nil {
		return orchestration.Result{}, err
	}
	completed := snapshot.CompletedTasks
	if completed == nil {
		completed = make(map[string]orchestration.ExecutionResult)
	}
	if plannerExists {
		completed[plannerTaskID] = plannerResult
	}
	stream, err := RunGraph(ctx, plan, completed, profile.MaxWorkers, func(taskCtx context.Context, task orchestration.Task) (TaskOutcome, error) {
		alias := aliasForRole(profile, task.Role)
		return e.runAssignment(taskCtx, runID, task.ID, task.Objective, task.Role, task.Permission, alias, request.Mode == orchestration.ModeHybrid, profile, ledger)
	})
	if err != nil {
		return orchestration.Result{}, err
	}
	for scheduled := range stream {
		if scheduled.Err != nil {
			_ = e.store.Append(runID, state.EventTaskFailed, scheduled.Task.ID, map[string]string{"error": scheduled.Err.Error()})
			return orchestration.Result{}, scheduled.Err
		}
		if err := e.appendOutcome(runID, scheduled.Task.ID, scheduled.Outcome); err != nil {
			return orchestration.Result{}, err
		}
		completed[scheduled.Task.ID] = scheduled.Outcome.Result
	}
	workerResults := make(map[string]orchestration.ExecutionResult, len(plan))
	for _, task := range plan {
		result, exists := completed[task.ID]
		if !exists {
			return orchestration.Result{}, fmt.Errorf("task %q has no result", task.ID)
		}
		workerResults[task.ID] = result
	}
	assignment := orchestration.Assignment{RunID: runID, TaskID: "synthesis", Objective: request.Objective, Role: orchestration.RoleOrchestrator, Permission: orchestration.PermissionRead, ModelAlias: profile.Orchestrator, CWD: request.CWD}
	outcome, err := e.executeModel(ctx, assignment, synthesisPrompt(request.Objective, workerResults), plannerResult.Session, false, profile, ledger)
	if err != nil {
		return orchestration.Result{}, err
	}
	if err := e.appendOutcome(runID, "synthesis", outcome); err != nil {
		return orchestration.Result{}, err
	}
	return outcome.Result.Result, nil
}

func (e *Engine) runAssignment(ctx context.Context, runID, taskID, objective string, role orchestration.Role, permission orchestration.Permission, alias string, allowAdvice bool, profile config.Profile, ledger *Ledger) (TaskOutcome, error) {
	assignment := orchestration.Assignment{RunID: runID, TaskID: taskID, Objective: objective, Role: role, Permission: permission, ModelAlias: alias}
	snapshot, _ := e.store.Snapshot(runID)
	if snapshot.Request != nil {
		assignment.CWD = snapshot.Request.CWD
	}
	outcome, err := e.executeModel(ctx, assignment, executionPrompt(objective, allowAdvice), nil, false, profile, ledger)
	if err != nil {
		return TaskOutcome{}, err
	}
	if outcome.Result.Result.Status != orchestration.ResultNeedsAdvice {
		return outcome, nil
	}
	if !allowAdvice {
		return TaskOutcome{}, errors.New("model requested advice when advice is disabled")
	}
	request := outcome.Result.Result.AdviceRequest
	if request == nil {
		return TaskOutcome{}, errors.New("advice request is missing")
	}
	advisorAssignment := orchestration.Assignment{RunID: runID, TaskID: taskID + ":advice", Objective: request.Question, Role: orchestration.RoleAdvisor, Permission: orchestration.PermissionRead, ModelAlias: profile.Advisor, CWD: assignment.CWD}
	advice, err := e.executeModel(ctx, advisorAssignment, advicePrompt(*request), nil, true, profile, ledger)
	if err != nil {
		return TaskOutcome{}, err
	}
	resumed, err := e.executeModel(ctx, assignment, resumeWithAdvicePrompt(objective, advice.Result.Result), outcome.Result.Session, false, profile, ledger)
	if err != nil {
		return TaskOutcome{}, err
	}
	resumed.Trace = append([]TraceEvent{{Type: string(state.EventAdviceRequested), Data: request}, {Type: string(state.EventAdviceCompleted), Data: advice.Result.Result}}, resumed.Trace...)
	return resumed, nil
}

func (e *Engine) executeModel(ctx context.Context, assignment orchestration.Assignment, prompt string, session *orchestration.NativeSessionRef, advisorCall bool, profile config.Profile, ledger *Ledger) (TaskOutcome, error) {
	chain, err := e.factory.Chain(assignment.ModelAlias)
	if err != nil {
		return TaskOutcome{}, err
	}
	var failures []string
	var trace []TraceEvent
	for chainIndex, alias := range chain {
		runner, model, err := e.factory.Runner(alias)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		for attempt := 0; attempt <= profile.MaxRetries; attempt++ {
			if err := ledger.Reserve(advisorCall); err != nil {
				return TaskOutcome{}, err
			}
			assignment.ModelAlias = alias
			attemptPrompt := prompt
			runSession := session
			if runSession != nil && runSession.Adapter != runner.Name() {
				runSession = nil
				attemptPrompt += "\n\nThe prior native session belongs to a different adapter. Continue from the complete checkpoint context in this prompt."
			}
			if attempt > 0 {
				attemptPrompt += "\n\nThe previous response was invalid or the command failed. Return a valid structured result and do not add prose outside it."
			}
			result, runErr := runner.Run(ctx, adapter.Invocation{Assignment: assignment, Prompt: attemptPrompt, Model: model.Model, Session: runSession})
			if runErr == nil {
				ledger.Settle(result.Result.Usage)
				if err := result.Result.Validate(); err != nil {
					runErr = err
				} else if result.Result.Status == orchestration.ResultFailed {
					runErr = fmt.Errorf("model reported failure: %s", outputOf(result.Result))
				} else {
					if chainIndex > 0 {
						trace = append(trace, TraceEvent{Type: string(state.EventFallbackUsed), Data: map[string]string{"model": alias}})
					}
					return TaskOutcome{Result: result, Trace: trace}, nil
				}
			}
			failures = append(failures, fmt.Sprintf("%s attempt %d: %v", alias, attempt+1, runErr))
			if ctx.Err() != nil {
				return TaskOutcome{}, ctx.Err()
			}
		}
	}
	return TaskOutcome{}, fmt.Errorf("all model attempts failed: %s", strings.Join(failures, "; "))
}

func validatePermissionCeiling(tasks []orchestration.Task, ceiling orchestration.Permission) error {
	if ceiling != orchestration.PermissionRead {
		return nil
	}
	for _, task := range tasks {
		if task.Permission == orchestration.PermissionWrite {
			return fmt.Errorf("task %q requests write permission above the read-only run ceiling", task.ID)
		}
	}
	return nil
}

func (e *Engine) appendOutcome(runID, taskID string, outcome TaskOutcome) error {
	for _, event := range outcome.Trace {
		if err := e.store.Append(runID, state.EventType(event.Type), taskID, event.Data); err != nil {
			return err
		}
	}
	sanitized := outcome.Result
	sanitized.Events = nil
	return e.store.Append(runID, state.EventTaskCompleted, taskID, sanitized)
}

func aliasForRole(profile config.Profile, role orchestration.Role) string {
	switch role {
	case orchestration.RoleReviewer:
		return profile.Reviewer
	case orchestration.RoleExecutor:
		return profile.Executor
	default:
		return profile.Worker
	}
}

func (e *Engine) Snapshot(runID string) (state.Snapshot, error) { return e.store.Snapshot(runID) }

func (e *Engine) Events(runID string) ([]state.Event, error) { return e.store.LoadEvents(runID) }

func (e *Engine) Doctor(ctx context.Context) []adapter.DoctorResult {
	return adapter.Doctor(ctx, e.cfg)
}

func (e *Engine) Cancel(runID string) error {
	pid, err := e.store.RunningPID(runID)
	if err != nil {
		return err
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(syscall.SIGTERM)
}

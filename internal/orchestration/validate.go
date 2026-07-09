package orchestration

import (
	"errors"
	"fmt"
	"strings"
)

// Validate rejects incomplete requests and unsupported execution choices.
func (r Request) Validate() error {
	if strings.TrimSpace(r.Objective) == "" {
		return errors.New("objective is required")
	}
	if strings.TrimSpace(r.CWD) == "" {
		return errors.New("cwd is required")
	}
	if r.Mode != "" && !validMode(r.Mode) {
		return fmt.Errorf("unsupported mode %q", r.Mode)
	}
	if r.Permission != "" && r.Permission != PermissionRead && r.Permission != PermissionWrite {
		return fmt.Errorf("unsupported permission %q", r.Permission)
	}
	return nil
}

// Validate rejects impossible result variants before they enter the engine.
func (r Result) Validate() error {
	if strings.TrimSpace(r.Summary) == "" {
		return errors.New("result summary is required")
	}
	switch r.Status {
	case ResultCompleted:
		if r.AdviceRequest != nil {
			return errors.New("completed result cannot contain advice_request")
		}
	case ResultNeedsAdvice:
		if r.AdviceRequest == nil {
			return errors.New("needs_advice result requires advice_request")
		}
		if err := r.AdviceRequest.Validate(); err != nil {
			return err
		}
	case ResultFailed:
	default:
		return fmt.Errorf("unsupported result status %q", r.Status)
	}
	if len(r.Tasks) > 0 {
		if err := ValidateTasks(r.Tasks); err != nil {
			return err
		}
	}
	return nil
}

// Validate checks that an advice request contains a question and reason.
func (r AdviceRequest) Validate() error {
	if strings.TrimSpace(r.Question) == "" {
		return errors.New("advice question is required")
	}
	if strings.TrimSpace(r.Reason) == "" {
		return errors.New("advice reason is required")
	}
	return nil
}

// ValidateTasks checks task identity, dependencies, roles, permissions, and cycles.
func ValidateTasks(tasks []Task) error {
	byID := make(map[string]Task, len(tasks))
	for _, task := range tasks {
		if strings.TrimSpace(task.ID) == "" || strings.TrimSpace(task.Objective) == "" {
			return errors.New("task id and objective are required")
		}
		if _, exists := byID[task.ID]; exists {
			return fmt.Errorf("duplicate task id %q", task.ID)
		}
		if !validRole(task.Role) {
			return fmt.Errorf("task %q has unsupported role %q", task.ID, task.Role)
		}
		if task.Permission != PermissionRead && task.Permission != PermissionWrite {
			return fmt.Errorf("task %q has unsupported permission %q", task.ID, task.Permission)
		}
		byID[task.ID] = task
	}
	for _, task := range tasks {
		for _, dependency := range task.DependsOn {
			if _, exists := byID[dependency]; !exists {
				return fmt.Errorf("task %q depends on unknown task %q", task.ID, dependency)
			}
		}
	}
	state := make(map[string]uint8, len(tasks))
	var visit func(string) error
	visit = func(id string) error {
		switch state[id] {
		case 1:
			return fmt.Errorf("task graph contains a cycle at %q", id)
		case 2:
			return nil
		}
		state[id] = 1
		for _, dependency := range byID[id].DependsOn {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for id := range byID {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

func validMode(mode ExecutionMode) bool {
	return mode == ModeSingle || mode == ModeAdvisor || mode == ModeOrchestrator || mode == ModeHybrid
}

func validRole(role Role) bool {
	return role == RoleOrchestrator || role == RoleExecutor || role == RoleAdvisor || role == RoleReviewer || role == RoleWorker
}

// Package orchestration defines provider-neutral task and result contracts.
package orchestration

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Role identifies an agent's responsibility in an orchestration run.
type Role string

// Supported agent roles.
const (
	RoleOrchestrator Role = "orchestrator"
	RoleExecutor     Role = "executor"
	RoleAdvisor      Role = "advisor"
	RoleReviewer     Role = "reviewer"
	RoleWorker       Role = "worker"
)

// ExecutionMode selects the orchestration strategy for a run.
type ExecutionMode string

// Supported orchestration strategies.
const (
	ModeSingle       ExecutionMode = "single"
	ModeAdvisor      ExecutionMode = "advisor"
	ModeOrchestrator ExecutionMode = "orchestrator"
	ModeHybrid       ExecutionMode = "hybrid"
)

// Permission declares whether an assignment may mutate its working directory.
type Permission string

// Supported assignment permission levels.
const (
	PermissionRead  Permission = "read"
	PermissionWrite Permission = "write"
)

// ResultStatus describes how an agent yielded control to the supervisor.
type ResultStatus string

// Supported agent result states.
const (
	ResultCompleted   ResultStatus = "completed"
	ResultNeedsAdvice ResultStatus = "needs_advice"
	ResultFailed      ResultStatus = "failed"
)

// Task is a validated node in an orchestration DAG.
type Task struct {
	ID         string     `json:"id"`
	Objective  string     `json:"objective"`
	DependsOn  []string   `json:"depends_on,omitempty"`
	Role       Role       `json:"role"`
	Permission Permission `json:"permission"`
}

// AdviceRequest captures the minimum context needed by an advisor.
type AdviceRequest struct {
	Question    string   `json:"question"`
	Facts       []string `json:"facts,omitempty"`
	Options     []string `json:"options,omitempty"`
	Constraints []string `json:"constraints,omitempty"`
	Reason      string   `json:"reason"`
}

// AdviceResponse is a provider-neutral advisor recommendation.
type AdviceResponse struct {
	Recommendation string   `json:"recommendation"`
	Reasons        []string `json:"reasons,omitempty"`
	Risks          []string `json:"risks,omitempty"`
}

// Usage records normalized model usage when a CLI exposes it.
type Usage struct {
	InputTokens  int64   `json:"input_tokens,omitempty"`
	OutputTokens int64   `json:"output_tokens,omitempty"`
	CostUSD      float64 `json:"cost_usd,omitempty"`
}

// Result is the structured response returned by every agent assignment.
type Result struct {
	Status        ResultStatus   `json:"status"`
	Summary       string         `json:"summary"`
	Output        string         `json:"output,omitempty"`
	AdviceRequest *AdviceRequest `json:"advice_request,omitempty"`
	Tasks         []Task         `json:"tasks,omitempty"`
	Usage         Usage          `json:"usage,omitempty"`
}

// Request describes a new orchestration run.
type Request struct {
	Objective  string        `json:"objective"`
	CWD        string        `json:"cwd"`
	Profile    string        `json:"profile"`
	Mode       ExecutionMode `json:"mode,omitempty"`
	Permission Permission    `json:"permission,omitempty"`
}

// Assignment binds a task and role to a model alias and working directory.
type Assignment struct {
	RunID      string     `json:"run_id"`
	TaskID     string     `json:"task_id"`
	Objective  string     `json:"objective"`
	Role       Role       `json:"role"`
	Permission Permission `json:"permission"`
	ModelAlias string     `json:"model_alias"`
	CWD        string     `json:"cwd"`
}

// NativeSessionRef identifies a resumable session owned by a local agent CLI.
type NativeSessionRef struct {
	Adapter   string `json:"adapter"`
	SessionID string `json:"session_id"`
}

// ExecutionResult combines a normalized result with optional session metadata.
type ExecutionResult struct {
	Result  Result            `json:"result"`
	Session *NativeSessionRef `json:"session,omitempty"`
	Events  []json.RawMessage `json:"events,omitempty"`
}

// NewID returns a cryptographically random identifier with the supplied prefix.
func NewID(prefix string) (string, error) {
	var data [12]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(data[:]), nil
}

// NowUTC returns the current timestamp normalized to UTC.
func NowUTC() time.Time { return time.Now().UTC() }

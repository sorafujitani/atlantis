// Package state persists orchestration runs as append-only events.
package state

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sorafujitani/atlantis/internal/orchestration"
)

// EventType identifies a state transition in a run.
type EventType string

// Supported run event types.
const (
	EventRunStarted      EventType = "run.started"
	EventPlanCreated     EventType = "plan.created"
	EventTaskStarted     EventType = "task.started"
	EventTaskCompleted   EventType = "task.completed"
	EventTaskFailed      EventType = "task.failed"
	EventAdviceRequested EventType = "advice.requested"
	EventAdviceCompleted EventType = "advice.completed"
	EventFallbackUsed    EventType = "fallback.used"
	EventRunCompleted    EventType = "run.completed"
	EventRunFailed       EventType = "run.failed"
	EventRunCancelled    EventType = "run.cancelled"
)

// Event is one durable state transition in sequence order.
type Event struct {
	Sequence int64           `json:"sequence"`
	At       time.Time       `json:"at"`
	Type     EventType       `json:"type"`
	TaskID   string          `json:"task_id,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// Snapshot is the current run state derived from replaying events.
type Snapshot struct {
	RunID          string                                   `json:"run_id"`
	Status         string                                   `json:"status"`
	SupervisorPID  int                                      `json:"supervisor_pid,omitempty"`
	Request        *orchestration.Request                   `json:"request,omitempty"`
	Plan           []orchestration.Task                     `json:"plan,omitempty"`
	CompletedTasks map[string]orchestration.ExecutionResult `json:"completed_tasks,omitempty"`
	FailedTasks    map[string]string                        `json:"failed_tasks,omitempty"`
	FinalResult    *orchestration.Result                    `json:"final_result,omitempty"`
	LastSequence   int64                                    `json:"last_sequence"`
}

// Store manages private local run directories and event logs.
type Store struct{ root string }

// New opens a state store and enforces private directory permissions.
func New(root string) (*Store, error) {
	if root == "" {
		return nil, errors.New("state root is required")
	}
	if err := os.MkdirAll(filepath.Join(root, "runs"), 0o700); err != nil {
		return nil, fmt.Errorf("create state root: %w", err)
	}
	if err := os.Chmod(root, 0o700); err != nil { //nolint:gosec // state directory must be private but traversable
		return nil, fmt.Errorf("secure state root: %w", err)
	}
	return &Store{root: root}, nil
}

// Root returns the configured state root.
func (s *Store) Root() string { return s.root }

// CreateRun allocates a run directory and records its original request.
func (s *Store) CreateRun(request orchestration.Request) (string, error) {
	runID, err := orchestration.NewID("run")
	if err != nil {
		return "", err
	}
	dir := s.runDir(runID)
	if err := os.Mkdir(dir, 0o700); err != nil {
		return "", fmt.Errorf("create run: %w", err)
	}
	if err := s.Append(runID, EventRunStarted, "", request); err != nil {
		return "", err
	}
	return runID, nil
}

// Append durably adds one event to a run.
func (s *Store) Append(runID string, eventType EventType, taskID string, data any) error {
	if err := validateRunID(runID); err != nil {
		return err
	}
	events, err := s.LoadEvents(runID)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode event data: %w", err)
	}
	event := Event{Sequence: int64(len(events) + 1), At: orchestration.NowUTC(), Type: eventType, TaskID: taskID, Data: raw}
	file, err := os.OpenFile(s.eventsPath(runID), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open event store: %w", err)
	}
	defer func() { _ = file.Close() }()
	if err := json.NewEncoder(file).Encode(event); err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync event store: %w", err)
	}
	return nil
}

// LoadEvents reads and validates every event in sequence order.
func (s *Store) LoadEvents(runID string) ([]Event, error) {
	if err := validateRunID(runID); err != nil {
		return nil, err
	}
	file, err := os.Open(s.eventsPath(runID))
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	var events []Event
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("decode event %d: %w", len(events)+1, err)
		}
		if event.Sequence != int64(len(events)+1) {
			return nil, fmt.Errorf("event sequence %d, expected %d", event.Sequence, len(events)+1)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read event store: %w", err)
	}
	return events, nil
}

// Snapshot replays a run's events into its current state.
func (s *Store) Snapshot(runID string) (Snapshot, error) {
	events, err := s.LoadEvents(runID)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot := Snapshot{
		RunID: runID, Status: "unknown",
		CompletedTasks: make(map[string]orchestration.ExecutionResult),
		FailedTasks:    make(map[string]string),
	}
	for _, event := range events {
		snapshot.LastSequence = event.Sequence
		switch event.Type {
		case EventRunStarted:
			var request orchestration.Request
			if err := json.Unmarshal(event.Data, &request); err != nil {
				return Snapshot{}, err
			}
			snapshot.Request = &request
			snapshot.Status = "running"
		case EventPlanCreated:
			if err := json.Unmarshal(event.Data, &snapshot.Plan); err != nil {
				return Snapshot{}, err
			}
		case EventTaskStarted:
			snapshot.Status = "running"
		case EventTaskCompleted:
			var result orchestration.ExecutionResult
			if err := json.Unmarshal(event.Data, &result); err != nil {
				return Snapshot{}, err
			}
			snapshot.CompletedTasks[event.TaskID] = result
			delete(snapshot.FailedTasks, event.TaskID)
		case EventTaskFailed:
			var value struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(event.Data, &value); err != nil {
				return Snapshot{}, err
			}
			snapshot.FailedTasks[event.TaskID] = value.Error
		case EventRunCompleted:
			var result orchestration.Result
			if err := json.Unmarshal(event.Data, &result); err != nil {
				return Snapshot{}, err
			}
			snapshot.FinalResult = &result
			snapshot.Status = "completed"
		case EventRunFailed:
			snapshot.Status = "failed"
		case EventRunCancelled:
			snapshot.Status = "cancelled"
		}
	}
	if snapshot.Status == "running" {
		pid, err := s.RunningPID(runID)
		if err != nil {
			snapshot.Status = "interrupted"
		} else {
			snapshot.SupervisorPID = pid
		}
	}
	return snapshot, nil
}

// Acquire creates an exclusive PID lock and reclaims stale locks.
func (s *Store) Acquire(runID string) (func() error, error) {
	if err := validateRunID(runID); err != nil {
		return nil, err
	}
	path := filepath.Join(s.runDir(runID), "lock")
	for attempt := 0; attempt < 2; attempt++ {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600) //nolint:gosec // path is derived from a validated run ID
		if err == nil {
			if _, err := fmt.Fprintf(file, "%d\n", os.Getpid()); err != nil {
				_ = file.Close()
				return nil, err
			}
			if err := file.Close(); err != nil {
				return nil, err
			}
			return func() error { return os.Remove(path) }, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		data, readErr := os.ReadFile(path) //nolint:gosec // lock path is derived from a validated run ID
		if readErr != nil {
			return nil, readErr
		}
		pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
		if parseErr != nil || !processAlive(pid) {
			if removeErr := os.Remove(path); removeErr != nil {
				return nil, removeErr
			}
			continue
		}
		return nil, fmt.Errorf("run %s is active in process %d", runID, pid)
	}
	return nil, fmt.Errorf("could not acquire run %s", runID)
}

// RunningPID returns the live supervisor PID recorded by a run lock.
func (s *Store) RunningPID(runID string) (int, error) {
	data, err := os.ReadFile(filepath.Join(s.runDir(runID), "lock"))
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid run lock: %w", err)
	}
	if !processAlive(pid) {
		return 0, fmt.Errorf("run process %d is not alive", pid)
	}
	return pid, nil
}

func (s *Store) runDir(runID string) string { return filepath.Join(s.root, "runs", runID) }
func (s *Store) eventsPath(runID string) string {
	return filepath.Join(s.runDir(runID), "events.jsonl")
}

func validateRunID(runID string) error {
	if !strings.HasPrefix(runID, "run_") || strings.ContainsAny(runID, `/\\`) {
		return fmt.Errorf("invalid run id %q", runID)
	}
	return nil
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

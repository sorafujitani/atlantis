// Package evaluation scores versioned orchestration suites.
package evaluation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sorafujitani/atlantis/internal/orchestration"
)

// Suite groups versioned evaluation cases.
type Suite struct {
	Version string `json:"version"`
	Cases   []Case `json:"cases"`
}

// Case defines an objective and its observable output expectation.
type Case struct {
	ID         string                      `json:"id"`
	Objective  string                      `json:"objective"`
	Exact      string                      `json:"exact,omitempty"`
	Contains   string                      `json:"contains,omitempty"`
	Mode       orchestration.ExecutionMode `json:"mode,omitempty"`
	Permission orchestration.Permission    `json:"permission,omitempty"`
}

// CaseResult records one evaluated orchestration run.
type CaseResult struct {
	ID         string              `json:"id"`
	RunID      string              `json:"run_id,omitempty"`
	Passed     bool                `json:"passed"`
	Output     string              `json:"output,omitempty"`
	Error      string              `json:"error,omitempty"`
	DurationMS int64               `json:"duration_ms"`
	Usage      orchestration.Usage `json:"usage,omitempty"`
}

// ProfileResult aggregates all cases executed with one profile.
type ProfileResult struct {
	Profile    string              `json:"profile"`
	Passed     int                 `json:"passed"`
	Total      int                 `json:"total"`
	DurationMS int64               `json:"duration_ms"`
	Usage      orchestration.Usage `json:"usage,omitempty"`
	Cases      []CaseResult        `json:"cases"`
}

// Report compares one evaluation suite across profiles.
type Report struct {
	SuiteVersion string          `json:"suite_version"`
	GeneratedAt  time.Time       `json:"generated_at"`
	Profiles     []ProfileResult `json:"profiles"`
}

// Load parses and validates a versioned evaluation suite.
func Load(path string) (Suite, error) {
	data, err := os.ReadFile(path) //nolint:gosec // eval suite path is explicitly selected by the user
	if err != nil {
		return Suite{}, fmt.Errorf("read eval suite: %w", err)
	}
	var suite Suite
	if err := json.Unmarshal(data, &suite); err != nil {
		return Suite{}, fmt.Errorf("parse eval suite: %w", err)
	}
	if suite.Version == "" || len(suite.Cases) == 0 {
		return Suite{}, fmt.Errorf("eval suite requires version and cases")
	}
	for _, testCase := range suite.Cases {
		if testCase.ID == "" || testCase.Objective == "" {
			return Suite{}, fmt.Errorf("eval case id and objective are required")
		}
		if testCase.Exact == "" && testCase.Contains == "" {
			return Suite{}, fmt.Errorf("eval case %q requires exact or contains", testCase.ID)
		}
	}
	return suite, nil
}

// Score compares model output against a case's exact or substring expectation.
func Score(testCase Case, output string) bool {
	if testCase.Exact != "" {
		return strings.TrimSpace(output) == strings.TrimSpace(testCase.Exact)
	}
	return strings.Contains(output, testCase.Contains)
}

// AddUsage accumulates normalized usage values.
func AddUsage(target *orchestration.Usage, usage orchestration.Usage) {
	target.InputTokens += usage.InputTokens
	target.OutputTokens += usage.OutputTokens
	target.CostUSD += usage.CostUSD
}

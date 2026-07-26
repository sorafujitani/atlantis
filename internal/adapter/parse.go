package adapter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sorafujitani/atlantis/internal/orchestration"
)

func ParseOutput(adapterName string, output []byte, finalTextOnly bool) (orchestration.ExecutionResult, error) {
	var events []json.RawMessage
	var candidates []any
	var sessionID string
	usage := orchestration.Usage{}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var value any
		if err := json.Unmarshal(line, &value); err != nil {
			candidates = append(candidates, string(line))
			continue
		}
		raw := append(json.RawMessage(nil), line...)
		events = append(events, raw)
		candidates = append(candidates, value)
		findSession(value, &sessionID)
		findUsage(value, &usage)
	}
	if err := scanner.Err(); err != nil {
		return orchestration.ExecutionResult{}, err
	}
	for index := len(candidates) - 1; index >= 0; index-- {
		if result, ok := findResult(candidates[index]); ok {
			if err := result.Validate(); err != nil {
				return orchestration.ExecutionResult{}, err
			}
			execution := orchestration.ExecutionResult{Result: result, Events: events}
			execution.Result.Usage = mergeUsage(execution.Result.Usage, usage)
			if sessionID != "" {
				execution.Session = &orchestration.NativeSessionRef{Adapter: adapterName, SessionID: sessionID}
			}
			return execution, nil
		}
	}
	text := strings.TrimSpace(string(output))
	if finalTextOnly && text != "" {
		return orchestration.ExecutionResult{Result: orchestration.Result{Status: orchestration.ResultCompleted, Summary: text, Output: text, Usage: usage}}, nil
	}
	return orchestration.ExecutionResult{}, errors.New("structured result not found")
}

func findResult(value any) (orchestration.Result, bool) {
	switch typed := value.(type) {
	case map[string]any:
		if _, hasStatus := typed["status"]; hasStatus {
			if _, hasSummary := typed["summary"]; hasSummary {
				raw, err := json.Marshal(typed)
				if err == nil {
					var result orchestration.Result
					if json.Unmarshal(raw, &result) == nil {
						return result, true
					}
				}
			}
		}
		for _, key := range []string{"structured_output", "result", "content", "message", "item", "data", "text", "output"} {
			if nested, exists := typed[key]; exists {
				if result, ok := findResult(nested); ok {
					return result, true
				}
			}
		}
		for _, nested := range typed {
			if result, ok := findResult(nested); ok {
				return result, true
			}
		}
	case []any:
		for index := len(typed) - 1; index >= 0; index-- {
			if result, ok := findResult(typed[index]); ok {
				return result, true
			}
		}
	case string:
		text := strings.TrimSpace(typed)
		if strings.HasPrefix(text, "```") {
			text = strings.TrimPrefix(text, "```json")
			text = strings.TrimPrefix(text, "```")
			text = strings.TrimSuffix(text, "```")
			text = strings.TrimSpace(text)
		}
		var nested any
		if json.Unmarshal([]byte(text), &nested) == nil {
			return findResult(nested)
		}
	}
	return orchestration.Result{}, false
}

func findSession(value any, sessionID *string) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"thread_id", "session_id", "sessionID", "chat_id", "chatId"} {
			if candidate, ok := typed[key].(string); ok && candidate != "" {
				*sessionID = candidate
			}
		}
		for _, nested := range typed {
			findSession(nested, sessionID)
		}
	case []any:
		for _, nested := range typed {
			findSession(nested, sessionID)
		}
	}
}

func findUsage(value any, usage *orchestration.Usage) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"input_tokens", "inputTokens", "prompt_tokens"} {
			if candidate := number(typed[key]); candidate > usage.InputTokens {
				usage.InputTokens = candidate
			}
		}
		for _, key := range []string{"output_tokens", "outputTokens", "completion_tokens"} {
			if candidate := number(typed[key]); candidate > usage.OutputTokens {
				usage.OutputTokens = candidate
			}
		}
		for _, key := range []string{"cost_usd", "total_cost_usd"} {
			if candidate, ok := typed[key].(float64); ok && candidate > usage.CostUSD {
				usage.CostUSD = candidate
			}
		}
		for _, nested := range typed {
			findUsage(nested, usage)
		}
	case []any:
		for _, nested := range typed {
			findUsage(nested, usage)
		}
	}
}

func number(value any) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case json.Number:
		value, _ := typed.Int64()
		return value
	default:
		return 0
	}
}

func mergeUsage(left, right orchestration.Usage) orchestration.Usage {
	if right.InputTokens > left.InputTokens {
		left.InputTokens = right.InputTokens
	}
	if right.OutputTokens > left.OutputTokens {
		left.OutputTokens = right.OutputTokens
	}
	if right.CostUSD > left.CostUSD {
		left.CostUSD = right.CostUSD
	}
	return left
}

func FormatResult(result orchestration.Result) string {
	if result.Output != "" {
		return result.Output
	}
	return fmt.Sprintf("%s: %s", result.Status, result.Summary)
}

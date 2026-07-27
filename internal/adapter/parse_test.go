package adapter

import (
	"strings"
	"testing"
)

func TestParseOutputFindsNestedResultAndSession(t *testing.T) {
	t.Parallel()
	output := []byte("{\"type\":\"thread.started\",\"thread_id\":\"abc\"}\n{\"type\":\"result\",\"structured_output\":{\"status\":\"completed\",\"summary\":\"done\",\"output\":\"ok\"},\"usage\":{\"input_tokens\":10,\"output_tokens\":5}}\n")
	result, err := ParseOutput("codex", output, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Session == nil || result.Session.SessionID != "abc" {
		t.Fatalf("session = %#v", result.Session)
	}
	if result.Result.Output != "ok" || result.Result.Usage.InputTokens != 10 {
		t.Fatalf("result = %#v", result.Result)
	}
}

func TestParseOutputFindsGrokBuildResult(t *testing.T) {
	t.Parallel()
	output := []byte(`{
  "text": "{\"status\":\"completed\",\"summary\":\"done\",\"output\":\"ok\"}",
  "sessionId": "grok-session",
  "usage": {"input_tokens": 21, "output_tokens": 8},
  "total_cost_usd": 0.012,
  "structuredOutput": {"status":"completed","summary":"done","output":"ok","advice_request":null,"tasks":[]}
}`)
	result, err := ParseOutput("grok", output, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Session == nil || result.Session.SessionID != "grok-session" {
		t.Fatalf("session = %#v", result.Session)
	}
	if result.Result.Output != "ok" || result.Result.Usage.InputTokens != 21 || result.Result.Usage.CostUSD != 0.012 {
		t.Fatalf("result = %#v", result.Result)
	}
}

func TestParseOutputFinalText(t *testing.T) {
	t.Parallel()
	result, err := ParseOutput("copilot", []byte("plain response\n"), true)
	if err != nil {
		t.Fatal(err)
	}
	if result.Result.Output != "plain response" {
		t.Fatalf("output = %q", result.Result.Output)
	}
}

func TestParseOutputOMPSessionAndUsage(t *testing.T) {
	t.Parallel()
	output := []byte(strings.Join([]string{
		`{"type":"session","id":"omp-abc","timestamp":"2026-01-01T00:00:00Z","cwd":"/tmp"}`,
		`{"type":"message_end","message":{"role":"assistant","content":[{"type":"text","text":"{\"status\":\"completed\",\"summary\":\"done\",\"output\":\"ok\",\"advice_request\":null,\"tasks\":[]}"}],"usage":{"input":11,"output":7,"cost":{"total":0.042}}}}`,
		`{"status":"completed","summary":"done","output":"ok","advice_request":null,"tasks":[]}`,
	}, "\n"))
	result, err := ParseOutput("omp", output, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Session == nil || result.Session.SessionID != "omp-abc" || result.Session.Adapter != "omp" {
		t.Fatalf("session = %#v", result.Session)
	}
	if result.Result.Output != "ok" || result.Result.Usage.InputTokens != 11 || result.Result.Usage.OutputTokens != 7 || result.Result.Usage.CostUSD != 0.042 {
		t.Fatalf("result = %#v", result.Result)
	}
}

package adapter

import "testing"

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

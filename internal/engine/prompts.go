package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sorafujitani/atlantis/internal/orchestration"
)

const resultContract = `Return only the structured result required by the supplied JSON schema.
Use status="completed" when the task is done, status="failed" when it cannot be completed, and status="needs_advice" only when advice is explicitly allowed and a consequential decision cannot be made safely.
Keep summary concise. Put the useful final answer in output. Do not include credentials, access tokens, or environment secrets.`

func executionPrompt(objective string, allowAdvice bool) string {
	advice := "Advice is not available. Do not return needs_advice."
	if allowAdvice {
		advice = "Advice is available. Return needs_advice only when guidance from a stronger model would materially change the result. Include question, known facts, options, constraints, and reason."
	}
	return strings.Join([]string{
		"Act as the executor for this assignment.",
		"Objective: " + objective,
		advice,
		resultContract,
	}, "\n\n")
}

func planningPrompt(objective string, hybrid bool, permission orchestration.Permission) string {
	mode := "Workers cannot ask an advisor."
	if hybrid {
		mode = "Workers may ask an advisor for consequential uncertainty."
	}
	permissionRule := "The run allows write tasks when mutation is necessary."
	if permission == orchestration.PermissionRead {
		permissionRule = "The run is read-only. Every task must use permission=read."
	}
	return strings.Join([]string{
		"Act only as an orchestrator. Plan and delegate; do not execute the task yourself.",
		"Objective: " + objective,
		"Return status=completed and a minimal acyclic tasks array. Each task needs a stable id, objective, role (worker, reviewer, or executor), permission (read or write), and dependencies.",
		"Use read tasks for research, inspection, and review. Use write tasks only for mutations. Prefer independent read tasks when parallel work adds real value. Keep the graph small.",
		permissionRule,
		mode,
		resultContract,
	}, "\n\n")
}

func advicePrompt(request orchestration.AdviceRequest) string {
	raw, _ := json.MarshalIndent(request, "", "  ")
	return strings.Join([]string{
		"Act as a senior advisor. Give a concrete recommendation for the executor. Do not perform the task or delegate it.",
		"Advice request:", string(raw),
		"Return status=completed. Put the recommendation, reasoning, and risks in output.",
		resultContract,
	}, "\n\n")
}

func resumeWithAdvicePrompt(objective string, advice orchestration.Result) string {
	return fmt.Sprintf("Continue the original assignment.\n\nObjective: %s\n\nAdvisor guidance:\n%s\n\nApply the guidance, complete the task, and return status=completed or failed. Do not request further advice.\n\n%s", objective, outputOf(advice), resultContract)
}

func synthesisPrompt(objective string, results map[string]orchestration.ExecutionResult) string {
	raw, _ := json.MarshalIndent(results, "", "  ")
	return strings.Join([]string{
		"Synthesize the worker results into the final answer for the original objective. Do not create another plan.",
		"Original objective: " + objective,
		"Worker results:", string(raw),
		"Return status=completed or failed. Put the complete final answer in output. The tasks array must be empty.",
		resultContract,
	}, "\n\n")
}

func outputOf(result orchestration.Result) string {
	if result.Output != "" {
		return result.Output
	}
	return result.Summary
}

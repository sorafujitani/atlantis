---
name: model-orchestrator
description: Run local multi-model agent workflows through the model-orchestrator CLI. Use when a task should combine stronger and cheaper models through Advisor, Orchestrator, or Hybrid delegation; when the user asks to delegate independent work across models; or when an interrupted orchestration run must be inspected, resumed, or cancelled.
---

# Model Orchestrator

Use the `model-orchestrator` CLI as the only execution interface. Do not reimplement orchestration in the host agent.

## Workflow

1. Check availability with `command -v model-orchestrator` and `model-orchestrator doctor`.
2. Choose the smallest fitting mode:
   - `single`: one model can complete the task.
   - `advisor`: a normal executor should ask a stronger model only for consequential uncertainty.
   - `orchestrator`: a stronger model should plan, delegate independent tasks, and synthesize results.
   - `hybrid`: delegated workers may also escalate bounded advice requests.
3. Default to `--permission read`. Use `--permission write` only when the user requested mutations.
4. Run from the target working directory and request JSON output:

```bash
model-orchestrator --output json run \
  --mode orchestrator \
  --permission read \
  --cwd "$PWD" \
  "<objective>"
```

5. Read `run_id` and `result.output` from the JSON response. Report the final output, not raw provider events.
6. If execution stops, inspect before retrying:

```bash
model-orchestrator --output json status <run-id>
model-orchestrator --output json inspect <run-id>
model-orchestrator --output json resume <run-id>
```

Use `model-orchestrator cancel <run-id>` only for an active run.

## Boundaries

- Never read, copy, or persist provider credentials or native session files.
- Never pass secrets in the objective or tracked config.
- Keep independent research tasks read-only and parallelizable.
- Keep shared-state mutations in the exclusive write lane.
- Do not start a second orchestration for work already owned by an active run.
- Use model aliases and profiles from the local config instead of hardcoding provider-specific behavior in the skill.

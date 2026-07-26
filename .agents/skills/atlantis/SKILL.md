---
name: atlantis
description: Atlantis is a portable agent operating mode backed by the atlantis CLI. Use it for nontrivial engineering work that benefits from principle-grounded playbooks, durable brain context, bounded multi-model delegation, verified execution, or /atlantis.
---

# Atlantis

Atlantis combines three concerns behind one name:

- An operating mode with task-specific playbooks.
- A local multi-model supervisor for bounded delegation.
- A portable brain vault for durable context.

Use the `atlantis` CLI as the execution and memory interface. Do not reimplement orchestration or index management in prompts.

## Start

1. Read the injected brain index and only the notes relevant to the task.
2. For design, review, or refactoring, read `principles.md` and the applicable principle leaves in full.
3. Choose one playbook from `playbooks/`. Keep its named steps visible as the execution checklist.
4. State the data shape or behavior contract before changing code.
5. Verify the result on the real surface before declaring completion.

## Operating Rules

- Prefer deletion and direct designs over compatibility layers or speculative abstractions.
- Observe facts with tools instead of asking the user. Ask only for product intent or irreversible choices that evidence cannot settle.
- Reproduce bugs before fixing them.
- Keep shared writes serialized. Parallelize only independent read work or isolated worktrees.
- Inspect delegated artifacts directly. A delegate's summary is not proof.
- Use reversible actions autonomously. Pause for destructive production actions, force-pushes to shared branches, data deletion, and customer-facing communication.
- Keep prose concise and concrete. Explain user impact, maintainer impact, the decision, and the verification.

## Orchestration

Check availability first:

```bash
command -v atlantis
atlantis doctor
```

Choose the smallest mode that fits:

- `single` uses one executor.
- `advisor` lets an executor escalate one consequential decision.
- `orchestrator` delegates an explicit DAG of independent tasks.
- `hybrid` combines orchestration with bounded worker advice.

Default to read-only. Use write permission only when the user requested mutations.

```bash
atlantis --output json run \
  --mode advisor \
  --permission read \
  --cwd "$PWD" \
  "<objective>"
```

For interrupted work, inspect before resuming:

```bash
atlantis --output json status <run-id>
atlantis --output json inspect <run-id>
atlantis --output json resume <run-id>
```

Never read, copy, or persist provider credentials or native session files. Do not start a second run for work already owned by an active run.

## Brain

The vault defaults to `~/brain` and can be overridden with `ATLANTIS_BRAIN_DIR` or `--dir`.

```bash
atlantis brain init
atlantis brain index
atlantis brain check
atlantis brain inject
atlantis brain plan finish <slug>
```

Persist only verified knowledge that would improve a different task. Merge overlaps and delete stale notes. Plans are transient. After work is verified complete, extract durable lessons and run `atlantis brain plan finish`; never archive completed plans in the vault.

## Conventions Used by Playbooks

- **how** means read-only exploration of the affected subsystem, directly or through `atlantis run --mode orchestrator --permission read`.
- **why** means history inspection with `git log`, `git blame`, issues, and PRs.
- **architect** means comparing concrete designs before implementation, using `advisor` or `orchestrator` mode when independence helps.
- **interrogate** means adversarial review from distinct correctness, simplicity, operational, and security lenses.
- **control** means driving the matching browser, CLI, TUI, server, or simulator rather than relying on proxies.
- **babysit** means owning CI and review follow-up after opening a PR.
- **show-me-your-work** means a compact decision log for long or unattended work.
- **figure-it-out** means the `brain-plan` skill for a temporary phased plan.
- **create-skill** means the installed skill-authoring skill.
- **tdd** means a failing regression test followed by the fix.

## Playbooks

- Investigation: `playbooks/investigation.md`
- Bug fix: `playbooks/bug-fix.md`
- Feature: `playbooks/feature.md`
- Refactoring or rename: `playbooks/refactoring.md`
- Performance issue: `playbooks/perf-issue.md`
- Hillclimb: `playbooks/hillclimb.md`
- Prototype: `playbooks/prototype.md`
- Runtime or trace forensics: `playbooks/runtime-forensics.md`, `playbooks/trace-forensics.md`
- Visual parity: `playbooks/visual-parity.md`
- Skill authoring: `playbooks/authoring-a-skill.md`
- Eval: `playbooks/eval.md`
- Autonomous run: `playbooks/autonomous-run.md`
- Session pickup or pause: `playbooks/session-pickup.md`, `playbooks/pause-safely.md`
- Multi-phase plan: `playbooks/multi-phase-plan.md`
- Opening a PR: `playbooks/opening-a-pr.md`

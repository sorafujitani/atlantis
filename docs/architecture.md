# Architecture

## Invariants

- Brain Markdown is canonical data.
- Harnesses read `index.md` and linked notes through native filesystem APIs.
- The Go CLI owns only derived indexes, validation, and transient plan cleanup.
- The Agent Skill owns playbooks and working conventions.
- The active coding-agent harness owns execution, delegation, progress, inspection, and cancellation.

## Dependency direction

```text
Agent Skill ───────────────> Brain Markdown
                                  ^
Host integration ─────────────────┤
                                  |
CLI ──> Brain maintenance ─────────┘
```

`internal/brain` does not know about host sessions or agent execution. Host integrations may ask the CLI to refresh derived indexes, but they read durable Markdown directly so context remains available when the binary is unavailable.

## Brain lifecycle

1. A harness integration asks Atlantis to refresh derived indexes on an appropriate lifecycle event.
2. The integration reads `brain/index.md` directly.
3. The agent opens only the linked notes relevant to the current task.
4. Verified reusable knowledge is written as Markdown.
5. Completed plans are reduced to reusable knowledge and deleted.

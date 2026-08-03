# Self-improvement loop

Capture lessons close to failure and route them to the right layer.

## Layers
1. **Structural** (preferred): lint, script, test, hook, skill, CI
2. **Atlantis repo** (`brain/`): this protocol, portable `principles/`
3. **Local vault**: `workflow/`, `codebase/`, `env/`
4. **Discard**: one-off noise

Prefer structure over prose (`[[principles/encode-lessons-in-structure]]`).  
Skills: [[protocol/agent-skills]].

## Capture triggers
- User correction
- Verification fails after a success claim
- Non-obvious codebase/env gotcha
- Same friction twice in one session
- Session ends after non-trivial work

Do not wait for the user to say "reflect".

## End-of-session minimum
1. Scan for triggers above
2. None → stop; no empty notes
3. Any → write/update the smallest durable note
4. Route: enforceable → lint/script/CI; procedure → edit or create skill ([[protocol/agent-skills]]); loop-critical → atlantis `brain/` + seed; host/project fact → local; else discard
5. Fix indexes/links if files changed
6. Final reply one-liner: what was captured (incl. skill create/edit), or `no durable lesson`

## Quality bar
Keep only high-signal, high-frequency, or high-impact notes. Update before creating.
Altitude: [[protocol/abstraction-altitude]]. PR review threads: [[protocol/pr-review-threads]].

## Repo-managed vs local
Atlantis repo holds only what a stranger needs for this loop: this protocol and `principles/`. `atlantis brain seed` overwrites those paths. Everything else is local (paths, accounts, one-repo facts, preferences, personal model IDs).

## Cadence
- Hot path: capture at correction time and session end (mandatory)
- Weekly / on request: `meditate`; on request: `ruminate`

Cadence tools maintain the vault; they do not replace the hot path.

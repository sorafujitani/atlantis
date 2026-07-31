# Self-improvement loop

Capture lessons close to failure and route them to the right layer.

## Layers
1. **Structural** (preferred): lint, script, test, hook, skill, CI
2. **Atlantis repo** (`brain/`): this protocol, portable `principles/`
3. **Local vault**: `workflow/`, `codebase/`, `env/`
4. **Discard**: one-off noise

Prefer structure over prose. See `[[principles/encode-lessons-in-structure]]`.

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
4. Route: needed to run the loop → atlantis repo `brain/` + `atlantis brain seed`; host/project/practical → local; skill bug → edit skill
5. Fix indexes/links if files changed
6. Final reply one-liner: what was captured, or `no durable lesson`

## Quality bar
Keep a note only if high-signal, high-frequency, or high-impact. Update existing notes before creating new ones.

## Repo-managed vs local
The atlantis repo holds only what a stranger needs to run this loop: this protocol and `principles/`. `atlantis brain seed` installs them and overwrites local copies.
Everything else is local: captured operational lessons, tool preferences, absolute paths, accounts, one-repo facts, personal model IDs.

## Cadence
- Hot path: capture at correction time and session end (mandatory)
- Weekly / on request: `meditate`
- On request: `ruminate`

Cadence tools maintain the vault; they do not replace the hot path.

# Atlantis CLI verification map

This directory is the maintained source for verifying user-facing Atlantis CLI behavior. Read the index before driving, then use the matching feature file as the recipe.

## Baseline preconditions

- From the Atlantis repo root, run `make build` so `./bin/atlantis` exists.
- Create a disposable brain dir under `.atlantis/verify-brain-$RUN_ID` or `/tmp/atlantis-verify-$RUN_ID`.
- Run `.agents/skills/verify-atlantis/helpers/doctor.sh <dir>` and require exit 0.
- Never point `--dir` / `ATLANTIS_BRAIN_DIR` at `~/brain` or any shared vault.

## Driving conventions

- Start each feature from a fresh disposable dir unless the file says otherwise.
- Prefer `./bin/atlantis … -o json` for stable field checks.
- Treat commands as literal.
- Delete the disposable brain dir after the drive; keep `.atlantis/verify-artifacts/`.

## Proof and skip reporting

- CLI proof includes command, stdout, stderr, and exit code under `.atlantis/verify-artifacts/<feature>/`.
- Mutation proof includes a second read of vault files or a follow-up command.
- Report unreachable paths with the attempted command and unmet precondition.

## Feature entry contract

Each feature file uses H1 + one paragraph, then H2s: `Sub-features`, `How to get to it (user POV)`, `Driving it with atlantis CLI`, `Gotchas`.

## Features

- [Version](./version.md) — version and top-level help.
- [Brain context](./brain-context.md) — disposable vault context output.
- [Brain cache](./brain-cache.md) — fingerprint cache hit/miss and `--force`.
- [Brain index and check](./brain-index-check.md) — index regeneration and validation.
- [Brain seed](./brain-seed.md) — repo-managed seed without destroying local notes.

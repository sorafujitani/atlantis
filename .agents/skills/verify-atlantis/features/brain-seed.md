# Brain seed

Seed installs repo-managed principles and protocol documents into a vault without deleting local-only notes.

## Sub-features

- `seed-principles` writes `principles.md` and `principles/` into the vault.
- `seed-protocol` writes `protocol/` (at least `self-improvement.md`; also altitude, PR-thread, and agent-skills protocols when present in the seed).
- `seed-preserves-local` leaves a pre-existing local note intact.

## How to get to it (user POV)

- Run `atlantis brain seed --dir <path>`.

## Driving it with atlantis CLI

Preconditions:

- Disposable `$DIR`; doctor passed; `brain init --dir "$DIR"` completed.
- Create a local-only note: `mkdir -p "$DIR/workflow" && printf '# Local note\n\nkeep me\n' >"$DIR/workflow/verify-local-keep.md"`.

- **Seed.** Run `./bin/atlantis brain seed --dir "$DIR"`. Exit code `0`.
- **Repo-managed present.** Test `-f "$DIR/principles.md"` and `-f "$DIR/protocol/self-improvement.md"`.
- **Local preserved.** `workflow/verify-local-keep.md` still contains `keep me`.
- **Proof.** Save `find`-style listing and file contents samples under `.atlantis/verify-artifacts/brain-seed/`.

## Gotchas

- Seed replaces repo-managed paths; that is expected. Only local sections (`workflow/`, `codebase/`, `env/`, `plans/`) must survive.
- Running seed against the user vault in verification is forbidden — doctor must reject it.

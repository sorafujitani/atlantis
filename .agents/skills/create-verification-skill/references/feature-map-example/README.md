# Notes CLI verification map

This directory is the maintained source for verifying the user-facing behavior of a Notes CLI. Read the index before driving the app, then use the matching feature file as the recipe.

## Baseline preconditions

- Build once with `make build` so `./bin/notes` exists.
- Set `NOTES_DATA_DIR=/tmp/notes-verify-$RUN_ID` so concurrent runs do not share state.
- Seed notes titled `Quarterly plan` and `Grocery list`.
- Put `./bin/notes` on `PATH` or invoke it by absolute path.
- Run the skill's doctor helper and require the expected binary path and disposable data directory.
- Never drive an instance that shares the user's real data directory.

## Driving conventions

- Start every recipe from the baseline state unless its preconditions say otherwise.
- Treat every command as literal. Keep quoted names and flags unchanged.
- Prefer `--format json` for stable assertions.
- Restore seeded data after a mutation. Do not remove proof artifacts during cleanup.

## Proof and skip reporting

- Capture the user action and the resulting state, not only the final exit code.
- CLI proof includes the command, stdout, stderr, and exit code under `.atlantis/verify-artifacts/<feature>/`.
- Mutation proof includes a read-only second view of the stored value.
- Record the feature ID and entry point used with every artifact.
- Report an unreachable path with the attempted command and the unmet precondition.
- Do not report a skipped entry point as verified through a different path.

## Feature entry contract

Each feature file starts with an H1 title and one paragraph describing the user-visible behavior. It then uses exactly four H2 sections in this order.

1. `Sub-features` lists short IDs with one line for each behavior.
2. `How to get to it (user POV)` lists every user entry point.
3. `Driving it with <harness>` starts with `Preconditions:` and uses labeled bullets that pair each user action with an exact command and observable result.
4. `Gotchas` lists traps that can waste or invalidate a verification run.

Keep implementation details out of the map. Name only user paths, stable handles, required state, commands, and observable proof.

## Features

- [Create a note](./create-note.md) covers CLI creation, cancellation semantics, persistence, and cleanup.
- [Search notes](./search.md) covers matching, empty, and clear states from the terminal.

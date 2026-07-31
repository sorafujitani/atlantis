# Search notes

Search lets a user find notes by title or body text and distinguish no matches from an error.

## Sub-features

- `search-match` returns title and body matches without changing note data.
- `search-empty` returns an empty result for a query with no matches.
- `search-error` fails clearly when the data directory is missing.

## How to get to it (user POV)

- Run `notes search <query>` in a terminal.
- Run `notes search <query> --format json` for machine-readable output.

## Driving it with notes CLI

Preconditions:

- `./bin/notes` is built and doctor reports a disposable `NOTES_DATA_DIR`.
- The disposable data directory contains `Quarterly plan` with body text `Draft budget`.

- **Title match.** Search for `quarterly`. Run `NOTES_DATA_DIR=$DIR ./bin/notes search "quarterly" --format json`. Exit code `0` and stdout contain one object whose title is `Quarterly plan`.
- **Body match.** Search for `budget`. Run `NOTES_DATA_DIR=$DIR ./bin/notes search "budget" --format json`. Exit code `0` and stdout still contain `Quarterly plan`.
- **Empty state.** Search for `volcano`. Run `NOTES_DATA_DIR=$DIR ./bin/notes search "volcano" --format json`. Exit code `0` and stdout are `[]`.
- **Proof.** Save command transcripts under `.atlantis/verify-artifacts/search/`. Confirm search did not mutate `notes list` output before vs after.

## Gotchas

- Human-readable default output is unstable for assertions. Prefer `--format json`.
- Search is read-only. If list order or content changes, the run corrupted state — stop and reset the disposable directory.
- A missing data directory is an error path, not an empty result. Assert the exit code.

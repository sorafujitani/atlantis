# Version

Version lets a user see which Atlantis binary they are running and discover top-level commands from help.

## Sub-features

- `version-print` prints version, commit, and build date.
- `help-top` lists available top-level commands including `brain` and `version`.

## How to get to it (user POV)

- Run `atlantis version`.
- Run `atlantis --help` or `atlantis help`.

## Driving it with atlantis CLI

Preconditions:

- `make build` completed; doctor accepts a disposable brain dir (unused for this feature except as a doctor argument).

- **Print version.** Run `./bin/atlantis version`. Exit code `0`; stdout matches `atlantis ` followed by a non-empty version token.
- **Top-level help.** Run `./bin/atlantis --help`. Exit code `0`; stdout mentions `brain` and `version`.
- **Proof.** Save both transcripts under `.atlantis/verify-artifacts/version/` as `version.stdout.txt`, `help.stdout.txt`, and matching `*.exit` files containing the exit codes.

## Gotchas

- A PATH `atlantis` may be an older install. Always drive `./bin/atlantis` from this checkout for verification.
- Dirty builds append `-dirty` to the version; assert prefix `atlantis `, not an exact release tag.

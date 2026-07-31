# Brain context

Brain context prints refreshed agent context for a vault so a harness can inject the index without reading files by hand.

## Sub-features

- `context-empty` works on a freshly initialized disposable vault.
- `context-prefix` starts with the standard safety prefix about GitHub repo creation and identity.

## How to get to it (user POV)

- Run `atlantis brain context` (default vault).
- Run `atlantis brain context --dir <path>` for an explicit vault.

## Driving it with atlantis CLI

Preconditions:

- Disposable dir `$DIR` created; doctor passed on `$DIR`.
- Run `./bin/atlantis brain init --dir "$DIR"` so the vault exists.

- **Print context.** Run `./bin/atlantis brain context --dir "$DIR"`. Exit code `0`.
- **Assert prefix.** Stdout begins with `Brain vault index.` and mentions never creating a GitHub repo without explicit ask.
- **Proof.** Save stdout/stderr/exit under `.atlantis/verify-artifacts/brain-context/`.

## Gotchas

- `context` refreshes indexes as needed; do not point `--dir` at the user vault during verification.
- Empty vaults still emit the prefix and an index skeleton — absence of local notes is not a failure.

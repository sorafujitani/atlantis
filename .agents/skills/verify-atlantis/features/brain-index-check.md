# Brain index and check

Index regenerates derived indexes; check validates links, reachability, and note size on the vault.

## Sub-features

- `index-write` creates or updates `index.md` in the disposable vault.
- `check-clean` reports reachable notes with zero broken links on a valid vault.
- `check-json` emits structured results with `-o json`.

## How to get to it (user POV)

- Run `atlantis brain index --dir <path>`.
- Run `atlantis brain check --dir <path>`.
- Run `atlantis brain check --dir <path> -o json`.

## Driving it with atlantis CLI

Preconditions:

- Disposable `$DIR`; doctor passed; `brain init` then `brain seed` completed on `$DIR` (seeded vault is a known-good check target).

- **Index.** Run `./bin/atlantis brain index --dir "$DIR"`. Exit code `0`; `$DIR/index.md` exists and is non-empty.
- **Check plain.** Run `./bin/atlantis brain check --dir "$DIR"`. Exit code `0`; stdout reports `0 broken`.
- **Check json.** Run `./bin/atlantis brain check --dir "$DIR" -o json`. Exit code `0`; stdout is JSON containing a broken-link count of zero.
- **Proof.** Copy `index.md` and command transcripts into `.atlantis/verify-artifacts/brain-index-check/`.

## Gotchas

- Oversized notes fail check; keep verification fixtures under the note size limit.
- Seeding before check avoids false failures from an empty uninitialized tree.

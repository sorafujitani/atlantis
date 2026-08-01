# Brain cache

Brain cache skips index rewrites when vault source Markdown is unchanged, and exposes a stable fingerprint for host adapters.

## Sub-features

- `cache-hit` returns identical context without rewriting `index.md` when sources are unchanged.
- `cache-miss` regenerates context after a source note is added.
- `print-fingerprint` prints only the vault source fingerprint.
- `force` rebuilds indexes even when the fingerprint is unchanged.

## How to get to it (user POV)

- Run `atlantis brain context` twice without editing notes.
- Add a note, then run `atlantis brain context` again.
- Run `atlantis brain context --print-fingerprint`.
- Run `atlantis brain context --force` or `atlantis brain index --force`.

## Driving it with atlantis CLI

Preconditions:

- Disposable brain dir under `/tmp` or an absolute path under `<repo>/.atlantis/`.
- Doctor exit 0.
- Start from `./bin/atlantis brain init --dir <dir>`.

- **Warm cache.** Run `./bin/atlantis brain --dir <dir> context` once. Exit `0`. Note `index.md` mtime.
- **Cache hit.** Run context again with no source edits. Exit `0`; stdout equals the first context; `index.md` mtime unchanged; `.atlantis-cache.json` exists.
- **Fingerprint.** Run `./bin/atlantis brain --dir <dir> context --print-fingerprint`. Exit `0`; stdout is one 64-char hex token. `-o json` context includes the same `fingerprint` field.
- **Cache miss.** Write `workflow/cache-probe.md`, run context again. Exit `0`; stdout differs and mentions `[[workflow/cache-probe]]`; fingerprint differs from the previous print.
- **Force.** Run `./bin/atlantis brain --dir <dir> context --force`. Exit `0`; context matches the post-miss output.
- **Proof.** Save transcripts under `.atlantis/verify-artifacts/brain-cache/`.

## Gotchas

- Fingerprint hashes source Markdown only; touching `index.md` alone must not change it.
- Relative `.atlantis/...` paths can fail doctor; use an absolute disposable path or `/tmp/...`.

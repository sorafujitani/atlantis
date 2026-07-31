---
name: verify-atlantis
description: >-
  Drive the Atlantis CLI on a disposable brain vault to prove user-facing
  commands (version, brain context/index/check/seed). Use when verifying
  Atlantis behavior, after CLI changes, or when Atlantis control means this
  skill.
disable-model-invocation: true
---

# Verify Atlantis

Atlantis is a short-lived CLI. There is no long-running server. Launch means build the binary once; each drive uses its own disposable `ATLANTIS_BRAIN_DIR` (or `--dir`) so the user's real vault is never touched.

## Launch

```bash
make build
test -x ./bin/atlantis
```

Ready when `./bin/atlantis version` prints a non-empty version line. Teardown for a drive is deleting that drive's disposable brain directory only — never `rm` the built binary as part of feature cleanup.

## Doctor

```bash
.agents/skills/verify-atlantis/helpers/doctor.sh <disposable-brain-dir>
```

Requires: executable `./bin/atlantis` from the repo root, and `<disposable-brain-dir>` under `/tmp` or the current repo's `.atlantis/` (refuse home `~/brain` and other non-disposable paths). Exit non-zero if either check fails.

## Drive

Invoke `./bin/atlantis` with absolute or `./bin/` path. Prefer:

```bash
export ATLANTIS_BRAIN_DIR="<disposable-brain-dir>"
# or: ./bin/atlantis brain <cmd> --dir "<disposable-brain-dir>"
```

Follow the matching file under [`features/`](features/). Prefer `-o json` when asserting structured fields.

## Evidence

Write proof under `.atlantis/verify-artifacts/<feature>/` (gitignored). Capture command, stdout, stderr, and exit code for each step that asserts behavior. Keep evidence after cleanup.

## Cleanup

Remove only the disposable brain directory created for the run (`rm -rf "$ATLANTIS_BRAIN_DIR"` when it passed doctor). Never delete `.atlantis/verify-artifacts/`. Never kill by process name.

## Helpers

- `helpers/doctor.sh <disposable-brain-dir>` — read-only readiness check (see Doctor).

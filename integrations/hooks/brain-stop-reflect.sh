#!/bin/sh
# Claude Code Stop hook: nudge end-of-session capture into the brain loop.
# Uses additionalContext so the model can still act before the turn fully ends.
set -eu

protocol="${ATLANTIS_BRAIN_DIR:-${HOME:?HOME is required}/brain}/protocol/self-improvement.md"

if [ ! -f "$protocol" ]; then
  exit 0
fi

# Always emit a short mandatory checklist. The agent decides "no durable lesson".
python3 - <<'PY'
import json
msg = """End-of-session brain capture is mandatory for non-trivial turns.

Read [[protocol/self-improvement]] if needed, then either:
1) write/update the smallest durable note (common vs local routing), or
2) finish with the exact line: no durable lesson

Do not ask the user whether to reflect. Capture corrections, failed verifications, gotchas, and repeated friction. Prefer structural enforcement over prose when possible."""
print(json.dumps({
    "hookSpecificOutput": {
        "hookEventName": "Stop",
        "additionalContext": msg,
    }
}, ensure_ascii=False))
PY

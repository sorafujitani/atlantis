#!/bin/sh
# Claude Code Stop hook: nudge end-of-session capture into the brain loop.
# Uses additionalContext so the model can still act before the turn fully ends.
# Stop fires after every assistant turn, not at session end, so the nudge is
# throttled per session: once, then again only after the cooldown elapses.
set -eu

protocol="${ATLANTIS_BRAIN_DIR:-${HOME:?HOME is required}/brain}/protocol/self-improvement.md"

if [ ! -f "$protocol" ]; then
  exit 0
fi

# The heredoc below occupies python's stdin, so the hook payload is read here
# and handed over via the environment.
ATLANTIS_HOOK_PAYLOAD="$(cat 2>/dev/null || true)"
export ATLANTIS_HOOK_PAYLOAD

# A sentinel file keyed by session_id throttles emission. The agent decides
# "no durable lesson".
python3 - <<'PY'
import json
import os
import sys
import tempfile
import time

COOLDOWN_SECONDS = 30 * 60

try:
    payload = json.loads(os.environ.get("ATLANTIS_HOOK_PAYLOAD", ""))
except json.JSONDecodeError:
    payload = {}
session = str(payload.get("session_id") or "unknown")
safe = "".join(c for c in session if c.isalnum() or c in "-_") or "unknown"
sentinel = os.path.join(tempfile.gettempdir(), f"atlantis-stop-reflect-{safe}")
try:
    if time.time() - os.stat(sentinel).st_mtime < COOLDOWN_SECONDS:
        sys.exit(0)
except FileNotFoundError:
    pass
with open(sentinel, "w", encoding="utf-8"):
    pass

msg = """End-of-session brain capture is mandatory for non-trivial turns.

Read [[protocol/self-improvement]] if needed, then either:
1) write/update the smallest durable note (repo-managed vs local routing), or
2) finish with the exact line: no durable lesson

Do not ask the user whether to reflect. Capture corrections, failed verifications, gotchas, and repeated friction. Prefer structural enforcement over prose when possible."""
print(json.dumps({
    "hookSpecificOutput": {
        "hookEventName": "Stop",
        "additionalContext": msg,
    }
}, ensure_ascii=False))
PY

#!/bin/sh
# Cursor sessionStart: sync alwaysApply rule, then emit additional_context JSON.
# Rule sync is the reliable path; JSON injection is best-effort (Agents Window
# has dropped sessionStart additional_context due to a composer timing bug).
set -eu

here="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
"$here/cursor-brain-sync.sh" || true

# Drain stdin so Cursor's hook payload does not block the pipe.
cat >/dev/null 2>&1 || true

python3 - <<'PY'
import json
import subprocess
import sys

try:
    ctx = subprocess.check_output(["atlantis", "brain", "context"], text=True)
except (OSError, subprocess.CalledProcessError):
    sys.exit(0)

print(json.dumps({"additional_context": ctx}, ensure_ascii=False))
PY

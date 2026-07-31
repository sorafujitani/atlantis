#!/bin/sh
# Cursor afterFileEdit / postToolUse: refresh derived indexes and the Cursor rule.
set -eu

here="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

# Drain stdin so Cursor's hook payload does not block the pipe.
cat >/dev/null 2>&1 || true

atlantis brain index >/dev/null
"$here/cursor-brain-sync.sh" || true

#!/bin/sh
# Read-only readiness check for verify-atlantis drives.
# Resolve ./bin/atlantis from the current working directory (repo root), not
# from this script's install path — the skill may live under ~/.agents/skills.
set -eu

brain_dir="${1:-}"
if [ -z "$brain_dir" ]; then
	echo "usage: doctor.sh <disposable-brain-dir>" >&2
	exit 2
fi

find_bin() {
	dir="$(pwd)"
	while [ "$dir" != "/" ]; do
		if [ -x "$dir/bin/atlantis" ]; then
			printf '%s\n' "$dir/bin/atlantis"
			return 0
		fi
		dir="$(dirname "$dir")"
	done
	return 1
}

bin="$(find_bin)" || {
	echo "doctor: missing ./bin/atlantis under $(pwd) or parents (run make build from the atlantis checkout)" >&2
	exit 1
}

case "$brain_dir" in
*/.atlantis/*|/tmp/*|/var/folders/*)
	;;
*)
	echo "doctor: refuse non-disposable brain dir: $brain_dir" >&2
	echo "doctor: use a path under /tmp or <repo>/.atlantis/" >&2
	exit 1
	;;
esac

home_brain="${HOME:?}/brain"
resolved="$(CDPATH= cd -- "$brain_dir" 2>/dev/null && pwd || printf '%s' "$brain_dir")"
if [ "$resolved" = "$home_brain" ]; then
	echo "doctor: refuse user vault $home_brain" >&2
	exit 1
fi

version="$("$bin" version)" || {
	echo "doctor: atlantis version failed" >&2
	exit 1
}
if [ -z "$version" ]; then
	echo "doctor: empty version output" >&2
	exit 1
fi

printf 'doctor: ok bin=%s version=%s brain=%s\n' "$bin" "$version" "$brain_dir"

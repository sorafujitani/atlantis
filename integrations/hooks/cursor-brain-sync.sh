#!/bin/sh
# Materialize brain context into Cursor alwaysApply surfaces.
# Agents Window can drop sessionStart additional_context; these paths are reliable.
set -eu

context="$(atlantis brain context)" || exit 0

rule_dir="${HOME:?HOME is required}/.cursor/rules"
rule_path="${rule_dir}/atlantis-brain.mdc"
agents_path="${HOME}/AGENTS.md"
start_mark='<!-- atlantis-brain:start -->'
end_mark='<!-- atlantis-brain:end -->'

mkdir -p "$rule_dir"
tmp="$(mktemp "${TMPDIR:-/tmp}/atlantis-cursor-brain.XXXXXX")"
trap 'rm -f "$tmp"' EXIT

{
	printf '%s\n' '---'
	printf '%s\n' 'description: Atlantis brain vault index (generated; do not edit)'
	printf '%s\n' 'alwaysApply: true'
	printf '%s\n' '---'
	printf '\n'
	printf '%s\n' "$context"
} >"$tmp"

if [ ! -f "$rule_path" ] || ! cmp -s "$tmp" "$rule_path"; then
	cp "$tmp" "$rule_path"
	chmod 600 "$rule_path"
fi

# ~/AGENTS.md is already always-applied in Cursor; keep a generated section in sync.
if [ -f "$agents_path" ]; then
	python3 - "$agents_path" "$start_mark" "$end_mark" "$context" <<'PY'
import pathlib
import sys

path = pathlib.Path(sys.argv[1])
start, end, context = sys.argv[2], sys.argv[3], sys.argv[4]
block = f"{start}\n{context.rstrip()}\n{end}\n"
text = path.read_text(encoding="utf-8")
if start in text and end in text:
    before, rest = text.split(start, 1)
    _, after = rest.split(end, 1)
    new = before + block + after.lstrip("\n")
else:
    new = text.rstrip() + "\n\n" + block
if new != text:
    path.write_text(new, encoding="utf-8")
PY
fi

#!/bin/sh
set -eu

brain_dir=${ATLANTIS_BRAIN_DIR:-"${HOME:?HOME is required}/brain"}
printf '%s\n\n' 'Brain vault index. Read only the linked notes relevant to the task before acting.'
exec cat "$brain_dir/index.md"

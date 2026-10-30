#!/usr/bin/env sh
# Point git at the tracked hooks in .githooks/.
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

git config core.hooksPath .githooks
printf 'Git hooks enabled: core.hooksPath=.githooks\n'

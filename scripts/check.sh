#!/usr/bin/env sh
# Shared guardrail: formatting, vet, and tests for the Go module in core/,
# plus release-consistency checks (README badge vs CHANGELOG).
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

# Release consistency: the README version badge must match the latest CHANGELOG
# version section, and CHANGELOG version sections must be unique. Catches a
# missed develop back-merge or forgotten badge bump before it ships.
latest_version=$(grep -m1 -E '^## \[[0-9]+\.[0-9]+\.[0-9]+\]' "$repo_root/CHANGELOG.md" | sed -E 's/^## \[([^]]+)\].*/\1/')
if [ -z "$latest_version" ]; then
  echo 'CHANGELOG: no version section found' >&2
  exit 1
fi

duplicates=$(grep -E '^## \[[0-9]+\.[0-9]+\.[0-9]+\]' "$repo_root/CHANGELOG.md" | sed -E 's/^## \[([^]]+)\].*/\1/' | sort | uniq -d)
if [ -n "$duplicates" ]; then
  printf 'CHANGELOG: duplicate version section(s): %s\n' "$duplicates" >&2
  exit 1
fi

for readme in README.md README.en.md; do
  badge=$(grep -m1 -o 'version-[0-9][0-9.]*-blue' "$repo_root/$readme" | sed 's/^version-//;s/-blue$//')
  if [ "$badge" != "$latest_version" ]; then
    printf '%s: badge %s != CHANGELOG latest version %s\n' "$readme" "${badge:-missing}" "$latest_version" >&2
    exit 1
  fi
done

cd "$repo_root/core"

unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  printf 'gofmt needed:\n%s\n' "$unformatted" >&2
  exit 1
fi

go vet ./...
go test ./...

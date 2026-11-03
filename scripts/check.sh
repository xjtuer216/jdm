#!/usr/bin/env sh
# Shared guardrail: formatting, vet, and tests for the Go module in core/.
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root/core"

unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  printf 'gofmt needed:\n%s\n' "$unformatted" >&2
  exit 1
fi

go vet ./...
go test ./...

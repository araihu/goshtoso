#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: scripts/run-focused-e2e.sh <impact.json> [--dry-run]" >&2
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage
  exit 2
fi

impact_file="$1"
dry_run="${2:-}"
if [[ "$dry_run" != "" && "$dry_run" != "--dry-run" ]]; then
  usage
  exit 2
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required to read $impact_file" >&2
  exit 1
fi

mode="$(jq -er '.mode' "$impact_file")"
identity_tags="$(jq -er '.tags | sort | join(",")' "$impact_file")"
case "$mode" in
  full)
    if [[ "$identity_tags" != "full" ]]; then
      echo "full impact must select exactly the full tag" >&2
      exit 1
    fi
    ;;
  focused)
    if [[ -z "$identity_tags" || ",$identity_tags," == *,full,* ]]; then
      echo "focused impact must select at least one non-full identity" >&2
      exit 1
    fi
    ;;
  *)
    echo "unsupported impact mode: $mode" >&2
    exit 1
    ;;
esac

suite_tags="e2e,$identity_tags"
echo "E2E mode: $mode"
echo "E2E tags: $suite_tags"
jq -r '.reasons[] | "- " + .' "$impact_file"

if [[ "$dry_run" == "--dry-run" ]]; then
  echo "go test -tags=e2e,$identity_tags ./site/tests/e2e -count=1 -timeout 15m"
  exit 0
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"
go test -tags="e2e,$identity_tags" ./site/tests/e2e -count=1 -timeout 15m

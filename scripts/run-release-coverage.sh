#!/usr/bin/env bash
set -euo pipefail

if [[ $# -gt 1 || ( $# -eq 1 && "$1" != "--local-dry-run" ) ]]; then
  echo "usage: scripts/run-release-coverage.sh [--local-dry-run]" >&2
  exit 2
fi
if [[ "${1:-}" == "--local-dry-run" ]]; then
  echo "release coverage local dry-run: no external publication will occur"
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
exec "$repo_root/scripts/run-component-coverage.sh" --tags e2e,full --badge

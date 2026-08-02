#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
runner="$repo_root/scripts/run-focused-e2e.sh"
fixture_dir="$(mktemp -d)"
trap 'rm -rf "$fixture_dir"' EXIT

printf '%s\n' '{"mode":"focused","tags":["button","actiongroup"],"reasons":["button changed"]}' > "$fixture_dir/focused.json"
output="$($runner "$fixture_dir/focused.json" --dry-run)"
grep -q 'E2E tags: e2e,actiongroup,button' <<< "$output"
test "$(grep -c '^go test ' <<< "$output")" -eq 1

printf '%s\n' '{"mode":"full","tags":["full"],"reasons":["unsafe"]}' > "$fixture_dir/full.json"
output="$($runner "$fixture_dir/full.json" --dry-run)"
grep -q 'E2E tags: e2e,full' <<< "$output"
test "$(grep -c '^go test ' <<< "$output")" -eq 1

printf '%s\n' '{"mode":"focused","tags":[],"reasons":[]}' > "$fixture_dir/empty.json"
if "$runner" "$fixture_dir/empty.json" --dry-run >/dev/null 2>&1; then
  echo "empty focused selection unexpectedly succeeded" >&2
  exit 1
fi

echo "run-focused-e2e contract: PASS"

#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
actual="$($repo_root/scripts/component-coverpkg.sh)"
expected="$(cd "$repo_root" && go list ./components/... | LC_ALL=C sort | paste -sd, -)"
test -n "$actual"
test "$actual" = "$expected"

grep -q -- '-pkg="$component_coverpkg"' "$repo_root/scripts/run-component-coverage.sh"
grep -q -- 'filter-authored-coverage' "$repo_root/scripts/run-component-coverage.sh"
grep -q -- 'coverage-authored.out' "$repo_root/scripts/run-component-coverage.sh"
grep -q -- 'full-percentage.txt' "$repo_root/scripts/run-component-coverage.sh"
grep -q -- 'authored Go coverage is below 80%' "$repo_root/scripts/run-component-coverage.sh"
grep -q -- '--tags e2e,full' "$repo_root/scripts/run-release-coverage.sh"
grep -q -- 'scripts/run-release-coverage.sh' "$repo_root/justfile"

echo "component coverage contract: PASS"

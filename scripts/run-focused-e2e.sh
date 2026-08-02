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

go_test_args=(
  -tags="e2e,$identity_tags"
  ./site/tests/e2e
  -count=1
)

if [[ "${CI:-}" != "true" ]]; then
  go test "${go_test_args[@]}" -timeout 15m
  exit 0
fi

log_dir="$(mktemp -d)"
trap 'rm -rf "$log_dir"' EXIT
suite_log="$log_dir/full-suite.log"

set +e
go test "${go_test_args[@]}" -timeout 15m 2>&1 | tee "$suite_log"
suite_status="${PIPESTATUS[0]}"
set -e
if [[ "$suite_status" -eq 0 ]]; then
  exit 0
fi

failed_tests=()
while IFS= read -r failed_test; do
  failed_tests+=("$failed_test")
done < <(sed -nE 's/^--- FAIL: (Test[A-Za-z0-9_]+).*/\1/p' "$suite_log" | sort -u)
if [[ "${#failed_tests[@]}" -eq 0 ]]; then
  exit "$suite_status"
fi

# A retry is allowed only when every failed top-level test contains a
# Playwright timeout. Assertion failures and all other failure modes remain
# immediately fatal. This keeps local runs strict and gives overloaded CI
# runners one bounded recovery attempt without weakening test assertions.
if ! awk '
  function finish_test() {
    if (test_name != "" && (trace_count == 0 || trace_count != timeout_count)) {
      non_timeout_failure = 1
    }
  }
  /^--- FAIL: Test[A-Za-z0-9_]+/ {
    finish_test()
    test_name = $3
    trace_count = 0
    timeout_count = 0
    next
  }
  test_name != "" && /Error Trace:/ {
    trace_count++
  }
  test_name != "" && /playwright: timeout: Timeout [0-9]+ms exceeded/ {
    timeout_count++
  }
  END {
    finish_test()
    exit non_timeout_failure
  }
' "$suite_log"; then
  exit "$suite_status"
fi

retry_pattern="$(IFS='|'; echo "${failed_tests[*]}")"
echo "Retrying timeout-only E2E failures once: ${failed_tests[*]}"
go test "${go_test_args[@]}" -run "^(${retry_pattern})$" -timeout 5m

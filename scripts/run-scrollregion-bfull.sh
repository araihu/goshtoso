#!/usr/bin/env bash
# Required-CI selection contract for the literal T-GS-011 B-FULL browser run.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
archive_path="${GOSHTOSO_AXE_CORE_TGZ:-}"
artifact_dir="${GOSHTOSO_SCROLLREGION_BFULL_ARTIFACT_DIR:-$repo_root/site/tests/e2e/test-results/scrollregion-bfull}"
suite_tags="e2e,scrollregion,bfull,axe"
selected_tests=()
while IFS= read -r test_name; do
  [[ -n "$test_name" ]] && selected_tests+=("$test_name")
done < <(cd "$repo_root/site" && go run ./cmd/e2econstraints --print-specialized-tests=scrollregion_bfull)
if [[ "${#selected_tests[@]}" -eq 0 ]]; then
  echo "identities.json declares no required Scroll Region B-FULL tests" >&2
  exit 1
fi

if [[ ! -r "$archive_path" ]]; then
  echo "GOSHTOSO_AXE_CORE_TGZ must name the readable authenticated axe-core archive" >&2
  exit 1
fi

if [[ -n "${GOSHTOSO_SCROLLREGION_BFULL_DIAGNOSTIC:-}" || -n "${GOSHTOSO_SCROLLREGION_BFULL_DIAGNOSTIC_MAX_CELLS_PER_ZOOM:-}" ]]; then
  echo "Required B-FULL CI rejects diagnostic caps; run the literal full matrix only" >&2
  exit 1
fi

cd "$repo_root"
pattern="^($(IFS='|'; echo "${selected_tests[*]}"))$"
echo "Required T-GS-011 B-FULL tags: $suite_tags"
echo "Required T-GS-011 B-FULL tests: ${selected_tests[*]}"
listing="$(GOSHTOSO_E2E_LIST_ONLY=1 go test -tags=e2e,scrollregion,bfull,axe ./site/tests/e2e -list "$pattern")"
for test_name in "${selected_tests[@]}"; do
  if [[ "$(grep -c "^${test_name}$" <<< "$listing")" -ne 1 ]]; then
    echo "B-FULL selection must contain exactly one $test_name" >&2
    exit 1
  fi
done

if [[ "${GOSHTOSO_SCROLLREGION_BFULL_LIST_ONLY:-}" == "1" ]]; then
  # Required selection preflight: output canonical manifest order after Go has
  # proved every declared test is compiled exactly once. No browser run here.
  printf '%s\n' "${selected_tests[@]}"
  exit 0
fi

printf '%s\n' "$listing"

mkdir -p "$artifact_dir"
go_test=(
  go test
  -tags=e2e,scrollregion,bfull,axe
  ./site/tests/e2e
  -count=1
  -timeout=60m
  -run="$pattern"
)

if [[ "${GOSHTOSO_SCROLLREGION_BFULL_XVFB:-auto}" != "off" && "$(uname -s)" == "Linux" ]]; then
  command -v xvfb-run >/dev/null 2>&1 || {
    echo "xvfb-run is required for the real-UA-zoom browser contract on Linux" >&2
    exit 1
  }
  GOSHTOSO_SCROLLREGION_BFULL_ARTIFACT_DIR="$artifact_dir" xvfb-run -a "${go_test[@]}"
else
  GOSHTOSO_SCROLLREGION_BFULL_ARTIFACT_DIR="$artifact_dir" "${go_test[@]}"
fi

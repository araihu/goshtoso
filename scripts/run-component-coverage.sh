#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage: scripts/run-component-coverage.sh [--phase all|units|e2e-merge] (--tags e2e,... | --impact file.json) [--badge]
EOF
}

phase="all"
tags=""
impact_file=""
write_badge=false
while [[ $# -gt 0 ]]; do
  case "$1" in
    --phase) phase="${2:-}"; shift 2 ;;
    --tags) tags="${2:-}"; shift 2 ;;
    --impact) impact_file="${2:-}"; shift 2 ;;
    --badge) write_badge=true; shift ;;
    *) usage; exit 2 ;;
  esac
done
if [[ "$phase" != "all" && "$phase" != "units" && "$phase" != "e2e-merge" ]]; then
  usage
  exit 2
fi
if [[ "$phase" != "units" && -z "$tags" && -z "$impact_file" ]]; then
  usage
  exit 2
fi
if [[ -n "$tags" && -n "$impact_file" ]]; then
  usage
  exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"
component_coverpkg="$(scripts/component-coverpkg.sh)"
if [[ -z "$component_coverpkg" ]]; then
  echo "component coverage package set is empty" >&2
  exit 1
fi

if [[ ! -f go.work ]]; then
  go work init . ./site
fi

if [[ "$phase" == "all" || "$phase" == "units" ]]; then
  rm -rf "$repo_root/.coverage"
  mkdir -p .coverage/unit-root .coverage/unit-site .coverage/e2e .coverage/merged
  printf '%s\n' "$component_coverpkg" | tr ',' '\n' > .coverage/component-packages.txt

  root_pkgs="$(go list ./...)"
  go test -cover -coverpkg="$component_coverpkg" $root_pkgs -count=1 \
    -args -test.gocoverdir="$repo_root/.coverage/unit-root"

  (
    cd site
    site_pkgs="$(go list ./... | grep -v '/tests/e2e')"
    go test -cover -coverpkg="$component_coverpkg" $site_pkgs -count=1 \
      -args -test.gocoverdir="$repo_root/.coverage/unit-site"
  )
fi

if [[ "$phase" == "units" ]]; then
  echo "component unit coverage phase: PASS"
  exit 0
fi

if [[ ! -s .coverage/component-packages.txt ]]; then
  echo "unit coverage phase is missing .coverage/component-packages.txt" >&2
  exit 1
fi

export GOSHTOSO_E2E_COVERDIR="$repo_root/.coverage/e2e"
export GOSHTOSO_E2E_COVERPKG="$component_coverpkg"
if [[ -n "$impact_file" ]]; then
  scripts/run-focused-e2e.sh "$impact_file"
else
  if [[ "$tags" != e2e,* ]]; then
    echo "coverage E2E tags must begin with e2e," >&2
    exit 1
  fi
  identity_tags="${tags#e2e,}"
  go test -tags="e2e,$identity_tags" ./site/tests/e2e -count=1 -timeout 15m
fi

rm -rf "$repo_root/.coverage/merged"
mkdir -p "$repo_root/.coverage/merged"
go tool covdata merge \
  -i=.coverage/unit-root,.coverage/unit-site,.coverage/e2e \
  -o=.coverage/merged
go tool covdata percent \
  -i=.coverage/merged -pkg="$component_coverpkg" \
  > .coverage/coverage-percent.txt
go tool covdata textfmt \
  -i=.coverage/merged -pkg="$component_coverpkg" \
  -o=.coverage/coverage.out
scripts/filter-authored-coverage \
  .coverage/coverage.out .coverage/coverage-authored.out
go tool cover -func=.coverage/coverage.out > .coverage/coverage-func.txt
go tool cover -func=.coverage/coverage-authored.out > .coverage/coverage-authored-func.txt
go tool cover -html=.coverage/coverage.out -o .coverage/coverage.html
go tool cover -html=.coverage/coverage-authored.out -o .coverage/coverage-authored.html

full_total_line="$(grep '^total:' .coverage/coverage-func.txt)"
full_total_percent="$(printf '%s\n' "$full_total_line" | sed -E 's/.*[[:space:]]([0-9.]+)%.*/\1/')"
printf '%s\n' "$full_total_percent" > .coverage/full-percentage.txt

authored_total_line="$(grep '^total:' .coverage/coverage-authored-func.txt)"
total_percent="$(printf '%s\n' "$authored_total_line" | sed -E 's/.*[[:space:]]([0-9.]+)%.*/\1/')"
printf '%s\n' "$total_percent" > .coverage/percentage.txt
if ! awk "BEGIN { exit !($total_percent >= 80) }"; then
  echo "authored Go coverage is below 80%: $total_percent%" >&2
  exit 1
fi

color="red"
awk "BEGIN { exit !($total_percent >= 80) }" && color="brightgreen" || true
awk "BEGIN { exit !($total_percent >= 60 && $total_percent < 80) }" && color="green" || true
awk "BEGIN { exit !($total_percent >= 40 && $total_percent < 60) }" && color="yellow" || true
printf '%s\n' "$color" > .coverage/color.txt

if [[ "$write_badge" == true ]]; then
  scripts/coveragebadge "$total_percent"
fi

echo "full generated-inclusive coverage: $full_total_line"
echo "authored Go coverage: $authored_total_line"
echo "component package list: .coverage/component-packages.txt"
echo "authored coverage percentage: $total_percent"

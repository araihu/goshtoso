#!/usr/bin/env bash
set -euo pipefail

required=(
  MODULE_CANDIDATE_REPOSITORY MODULE_PATH MODULE_CANDIDATE_COMMIT
  MODULE_CANDIDATE_TREE MODULE_CANDIDATE_SUBDIR MODULE_CANDIDATE_OUTPUT
  MODULE_CANDIDATE_DEPENDENCY_PROXY MODULE_CANDIDATE_RECEIPT DEPLOY_TARGET
)
for name in "${required[@]}"; do
  if [[ ! -v "$name" ]]; then
    echo "preflight-public-deploy: required environment is absent: $name" >&2
    exit 2
  fi
done
if [[ -z "$MODULE_CANDIDATE_REPOSITORY" || -z "$MODULE_PATH" || -z "$MODULE_CANDIDATE_COMMIT" || -z "$MODULE_CANDIDATE_TREE" || -z "$MODULE_CANDIDATE_OUTPUT" || -z "$MODULE_CANDIDATE_DEPENDENCY_PROXY" || -z "$MODULE_CANDIDATE_RECEIPT" || -z "$DEPLOY_TARGET" ]]; then
  echo "preflight-public-deploy: required environment must be nonempty except MODULE_CANDIDATE_SUBDIR" >&2
  exit 2
fi
if [[ "$MODULE_CANDIDATE_SUBDIR" != "" ]]; then
  echo "preflight-public-deploy: MODULE_CANDIDATE_SUBDIR must be empty" >&2
  exit 1
fi
case "$MODULE_CANDIDATE_DEPENDENCY_PROXY" in
  file://*) ;;
  *) echo "preflight-public-deploy: dependency proxy must use file:// with no fallback" >&2; exit 1 ;;
esac
if [[ ! -f "$MODULE_CANDIDATE_RECEIPT" ]]; then
  echo "preflight-public-deploy: candidate receipt is missing" >&2
  exit 1
fi
receipt_value() {
  local key=$1
  awk -F= -v key="$key" '$1 == key { print substr($0, length(key) + 2) }' "$MODULE_CANDIDATE_RECEIPT"
}
for key in MODULE_PATH MODULE_CANDIDATE_COMMIT MODULE_CANDIDATE_TREE MODULE_CANDIDATE_SUBDIR DEPLOY_TARGET; do
  got=$(receipt_value "$key")
  expected=${!key}
  if [[ "$got" != "$expected" ]]; then
    echo "preflight-public-deploy: receipt mismatch for $key" >&2
    exit 1
  fi
done

repo_root=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_root"
exec env GOWORK=off GOPROXY=off GOSUMDB=off go run ./cmd/module-candidate-proxy \
  -repository "$MODULE_CANDIDATE_REPOSITORY" \
  -module-path "$MODULE_PATH" \
  -commit "$MODULE_CANDIDATE_COMMIT" \
  -tree "$MODULE_CANDIDATE_TREE" \
  -subdir "$MODULE_CANDIDATE_SUBDIR" \
  -output "$MODULE_CANDIDATE_OUTPUT" \
  -dependency-proxy "$MODULE_CANDIDATE_DEPENDENCY_PROXY"

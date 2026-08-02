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

fake_bin="$fixture_dir/bin"
fake_go="$fake_bin/go"
fake_state="$fixture_dir/go-state"
fake_args="$fixture_dir/go-args"
mkdir -p "$fake_bin"
printf '%s\n' \
  '#!/usr/bin/env bash' \
  'set -euo pipefail' \
  'count=0' \
  '[[ ! -f "$FAKE_GO_STATE" ]] || count="$(<"$FAKE_GO_STATE")"' \
  'count=$((count + 1))' \
  'printf "%s\n" "$count" > "$FAKE_GO_STATE"' \
  'printf "%s\n" "$*" >> "$FAKE_GO_ARGS"' \
  'if [[ "$count" -eq 1 ]]; then' \
  '  echo "--- FAIL: TestTimeoutVictim (5.01s)"' \
  '  echo "    Error Trace: fake_test.go:10"' \
  '  if [[ "${FAKE_GO_MODE:-timeout}" != "assertion" ]]; then' \
  '    echo "    timeout:Timeout 5000.00ms exceeded."' \
  '    if [[ "${FAKE_GO_MODE:-timeout}" == "mixed" ]]; then' \
  '      echo "    Error Trace: fake_test.go:11"' \
  '      echo "    Error: Not equal"' \
  '    fi' \
  '    echo "--- FAIL: TestSecondTimeoutVictim (3.01s)"' \
  '    echo "    Error Trace: fake_test.go:20"' \
  '    echo "    playwright: timeout: Timeout 3000ms exceeded."' \
  '  else' \
  '    echo "    Error: Not equal"' \
  '  fi' \
  '  exit 1' \
  'fi' \
  'exit 0' \
  > "$fake_go"
chmod +x "$fake_go"

output="$(CI=true PATH="$fake_bin:$PATH" FAKE_GO_STATE="$fake_state" FAKE_GO_ARGS="$fake_args" "$runner" "$fixture_dir/full.json")"
grep -q 'Retrying timeout-only E2E failures once: TestSecondTimeoutVictim TestTimeoutVictim' <<< "$output"
test "$(<"$fake_state")" -eq 2
grep -q -- '-run ^(TestSecondTimeoutVictim|TestTimeoutVictim)$ -timeout 5m' "$fake_args"

rm -f "$fake_state" "$fake_args"
if CI=true PATH="$fake_bin:$PATH" FAKE_GO_MODE=assertion FAKE_GO_STATE="$fake_state" FAKE_GO_ARGS="$fake_args" \
  "$runner" "$fixture_dir/full.json" >/dev/null 2>&1; then
  echo "non-timeout failure unexpectedly retried and passed" >&2
  exit 1
fi
test "$(<"$fake_state")" -eq 1

rm -f "$fake_state" "$fake_args"
if CI=true PATH="$fake_bin:$PATH" FAKE_GO_MODE=mixed FAKE_GO_STATE="$fake_state" FAKE_GO_ARGS="$fake_args" \
  "$runner" "$fixture_dir/full.json" >/dev/null 2>&1; then
  echo "mixed timeout and assertion failure unexpectedly retried and passed" >&2
  exit 1
fi
test "$(<"$fake_state")" -eq 1

printf '%s\n' '{"mode":"focused","tags":[],"reasons":[]}' > "$fixture_dir/empty.json"
if "$runner" "$fixture_dir/empty.json" --dry-run >/dev/null 2>&1; then
  echo "empty focused selection unexpectedly succeeded" >&2
  exit 1
fi

echo "run-focused-e2e contract: PASS"

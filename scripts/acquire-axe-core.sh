#!/usr/bin/env bash
# Acquire the exact maintained axe-core tarball used by the browser contract.
# A cache hit is never trusted implicitly: it must pass the archive and payload
# hashes before use. A bad cached artifact fails closed instead of redownloading.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
lock_file="$repo_root/scripts/axe-core.lock"
cache_dir="$repo_root/.cache/axe-core"
output_path=""

usage() {
  cat >&2 <<'EOF'
usage: scripts/acquire-axe-core.sh [--cache-dir DIR] [--output FILE]
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --cache-dir)
      [[ $# -ge 2 ]] || { usage; exit 2; }
      cache_dir="$2"
      shift 2
      ;;
    --output)
      [[ $# -ge 2 ]] || { usage; exit 2; }
      output_path="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

fail() {
  echo "axe-core acquisition: $*" >&2
  exit 1
}

lock_value() {
  local key="$1"
  local value
  value="$(awk -F= -v key="$key" '
    $1 == key { count++; print substr($0, length(key) + 2) }
    END { if (count != 1) exit 1 }
  ' "$lock_file")" || fail "lock must contain exactly one $key"
  [[ -n "$value" ]] || fail "lock value $key must not be empty"
  printf '%s' "$value"
}

sha256_file() {
  shasum -a 256 "$1" | awk '{print $1}'
}

[[ -f "$lock_file" ]] || fail "missing lock file $lock_file"
version="$(lock_value version)"
url="$(lock_value url)"
archive_sha256="$(lock_value archive_sha256)"
script_path="$(lock_value script_path)"
script_sha256="$(lock_value script_sha256)"

[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || fail "invalid locked version"
[[ "$url" == "https://registry.npmjs.org/axe-core/-/axe-core-${version}.tgz" ]] || fail "locked URL does not bind the locked version"
[[ "$archive_sha256" =~ ^[0-9a-f]{64}$ ]] || fail "invalid archive SHA-256"
[[ "$script_sha256" =~ ^[0-9a-f]{64}$ ]] || fail "invalid script SHA-256"
[[ "$script_path" == "package/axe.min.js" ]] || fail "unexpected injected script path"

verify_archive() {
  local archive="$1"
  [[ -f "$archive" ]] || fail "archive is absent: $archive"
  [[ "$(sha256_file "$archive")" == "$archive_sha256" ]] || fail "archive SHA-256 mismatch: $archive"
  [[ "$(tar -xOf "$archive" "$script_path" | shasum -a 256 | awk '{print $1}')" == "$script_sha256" ]] || fail "injected axe-core script SHA-256 mismatch: $archive"
}

mkdir -p "$cache_dir"
archive_path="$cache_dir/axe-core-${version}.tgz"
if [[ -e "$archive_path" ]]; then
  verify_archive "$archive_path"
else
  temporary_path="$(mktemp "$cache_dir/.axe-core-${version}.XXXXXX")"
  cleanup_temporary() { rm -f -- "$temporary_path"; }
  trap cleanup_temporary EXIT
  curl --fail --location --proto '=https' --tlsv1.2 --retry 3 --output "$temporary_path" "$url"
  verify_archive "$temporary_path"
  mv -- "$temporary_path" "$archive_path"
  trap - EXIT
  verify_archive "$archive_path"
fi

if [[ -n "$output_path" ]]; then
  mkdir -p "$(dirname "$output_path")"
  install -m 0644 "$archive_path" "$output_path"
  verify_archive "$output_path"
  archive_path="$output_path"
fi

printf 'authenticated axe-core %s archive: %s\n' "$version" "$archive_path"

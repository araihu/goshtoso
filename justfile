# Goshtoso justfile — see Makefile for a full target list

# Generate templ files
gp-generate:
    templ generate

# Run dev server (builds CSS first)
gp-dev: css
    go run ./site/cmd/server

# Build assets/styles.css with the PINNED Tailwind (assets/tailwind.version),
# regenerating the embeddable theme source first. Fetches the standalone binary
# on demand into .tools/ (gitignored) — no binary is committed.
css:
    #!/usr/bin/env bash
    set -euo pipefail
    go run ./scripts/themegen
    ver="$(cat assets/tailwind.version)"
    bin=".tools/tailwindcss-${ver}"
    if [ ! -x "$bin" ]; then
      mkdir -p .tools
      _uname="$(uname -s | tr '[:upper:]' '[:lower:]')"
      case "$_uname" in darwin) os=macos;; linux) os=linux;; *) os="$_uname";; esac
      arch="$(uname -m)"; case "$arch" in arm64|aarch64) arch=arm64;; x86_64) arch=x64;; *) echo "unsupported arch: $arch" >&2; exit 1;; esac
      echo "fetching tailwindcss v${ver} (${os}-${arch})..."
      tmp="$(mktemp .tools/tailwindcss.XXXXXX)"
      curl -fsSL -o "$tmp" \
        "https://github.com/tailwindlabs/tailwindcss/releases/download/v${ver}/tailwindcss-${os}-${arch}"
      chmod +x "$tmp"
      mv "$tmp" "$bin"
    fi
    "$bin" -i css/main.css -o assets/styles.css
    echo "css: built assets/styles.css with tailwindcss v${ver}"

# Download the PINNED vendored JS (assets/js/runtime/versions.json) into
# versioned dirs and regenerate the URL constants. Mirrors `just css`.
vendor-js:
    go run ./scripts/vendorgen -download

# Goshtoso justfile — see Makefile for a full target list

# Generate templ files
gp-generate:
    templ generate

# Run dev server (builds CSS first)
gp-dev: css
    go run cmd/server/main.go

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
      arch="$(uname -m)"; case "$arch" in arm64|aarch64) arch=arm64;; x86_64) arch=x64;; esac
      echo "fetching tailwindcss v${ver} (${os}-${arch})..."
      curl -fsSL -o "$bin" \
        "https://github.com/tailwindlabs/tailwindcss/releases/download/v${ver}/tailwindcss-${os}-${arch}"
      chmod +x "$bin"
    fi
    "$bin" -i css/main.css -o assets/styles.css
    echo "css: built assets/styles.css with tailwindcss v${ver}"

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
    go run ./cmd/themegen
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
    go run ./cmd/vendorgen -download

# Build tracked library and demo-site JavaScript from their owned source roots.
js:
    go run ./cmd/jsbuild

# Parse authored JavaScript, report structural similarity, enforce the inline
# extraction baseline, and verify generated artifacts have no drift.
js-check:
    go run ./cmd/jslint
    go run ./cmd/jsbuild -check

# Test the nested site against the root library in this checkout through a
# throwaway workspace. This is the current-source integration contract.
site-current-source-integration:
    ./scripts/check-site-module current-source

# Test the nested site exactly as a standalone consumer of site/go.mod. This is
# the pinned-dependency deployability contract and always forces GOWORK=off.
site-pinned-dependency-deployability:
    ./scripts/check-site-module pinned-dependency

site-module-contracts: site-current-source-integration site-pinned-dependency-deployability

# Run E2E identities impacted by committed changes since the supplied base.
test-e2e-focused base="origin/main":
    go run ./cmd/e2eimpact --base "{{base}}" --head HEAD > .e2e-impact.json
    scripts/run-focused-e2e.sh .e2e-impact.json

# Run the authoritative release-equivalent full coverage pipeline locally.
coverage:
    scripts/run-release-coverage.sh

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

# Run root unit tests, site unit tests, and E2E tests, then merge component
# coverage data into .coverage/coverage.out and .coverage/coverage.html.
coverage:
    #!/usr/bin/env bash
    set -euo pipefail
    root="$PWD"
    component_coverpkg="$(go list ./components/... | paste -sd, -)"

    if [ ! -f go.work ]; then
      go work init . ./site
    fi

    rm -rf .coverage
    mkdir -p .coverage/unit-root .coverage/unit-site .coverage/e2e .coverage/merged

    root_pkgs="$(go list ./... | grep -v '^github.com/araihu/goshtoso/site/')"
    go test -cover -coverpkg="$component_coverpkg" $root_pkgs -count=1 \
      -args -test.gocoverdir="$root/.coverage/unit-root"

    (
      cd site
      site_pkgs="$(go list ./... | grep -v '/tests/e2e')"
      site_component_coverpkg="$(go list -deps $site_pkgs | grep '^github.com/araihu/goshtoso/components/' | sort -u | paste -sd, -)"
      go test -cover -coverpkg="$site_component_coverpkg" $site_pkgs -count=1 \
        -args -test.gocoverdir="$root/.coverage/unit-site"

      GOSHTOSO_E2E_COVERDIR="$root/.coverage/e2e" \
      GOSHTOSO_E2E_COVERPKG="$component_coverpkg" \
        go test ./tests/e2e/... -count=1 -timeout 15m
    )

    go tool covdata merge \
      -i=.coverage/unit-root,.coverage/unit-site,.coverage/e2e \
      -o=.coverage/merged
    go tool covdata percent -i=.coverage/merged -pkg="$component_coverpkg" > .coverage/coverage-percent.txt
    go tool covdata textfmt -i=.coverage/merged -pkg="$component_coverpkg" -o=.coverage/coverage.out
    go tool cover -func=.coverage/coverage.out > .coverage/coverage-func.txt
    go tool cover -html=.coverage/coverage.out -o .coverage/coverage.html

    total_line="$(grep '^total:' .coverage/coverage-func.txt)"
    total_percent="$(printf '%s\n' "$total_line" | sed -E 's/.*\t([0-9.]+)%.*/\1/')"
    scripts/coveragebadge "$total_percent"

    echo "$total_line"
    echo "coverage artifacts: .coverage/coverage.out .coverage/coverage-func.txt .coverage/coverage-percent.txt .coverage/coverage.html badges/coverage.svg"

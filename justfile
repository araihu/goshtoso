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

# Run root unit tests, site unit tests, and E2E tests, then merge all Go
# coverage data into .coverage/coverage.out and .coverage/coverage.html.
coverage:
    #!/usr/bin/env bash
    set -euo pipefail
    root="$PWD"
    root_coverpkg="github.com/araihu/goshtoso/..."
    all_coverpkg="${root_coverpkg},github.com/araihu/goshtoso/site/..."

    if [ ! -f go.work ]; then
      go work init . ./site
    fi

    rm -rf .coverage
    mkdir -p .coverage/unit-root .coverage/unit-site .coverage/e2e .coverage/merged

    go test -cover -coverpkg="$root_coverpkg" ./... -count=1 \
      -args -test.gocoverdir="$root/.coverage/unit-root"

    (
      cd site
      site_pkgs="$(go list ./... | grep -v '/tests/e2e')"
      go test -cover -coverpkg="$all_coverpkg" $site_pkgs -count=1 \
        -args -test.gocoverdir="$root/.coverage/unit-site"

      GOSHTOSO_E2E_COVERDIR="$root/.coverage/e2e" \
      GOSHTOSO_E2E_COVERPKG="$all_coverpkg" \
        go test ./tests/e2e/... -count=1 -timeout 15m
    )

    go tool covdata merge \
      -i=.coverage/unit-root,.coverage/unit-site,.coverage/e2e \
      -o=.coverage/merged
    go tool covdata percent -i=.coverage/merged > .coverage/coverage-percent.txt
    go tool covdata textfmt -i=.coverage/merged -o=.coverage/coverage.out
    go tool cover -func=.coverage/coverage.out > .coverage/coverage-func.txt
    go tool cover -html=.coverage/coverage.out -o .coverage/coverage.html

    total_line="$(grep '^total:' .coverage/coverage-func.txt)"
    total_percent="$(printf '%s\n' "$total_line" | sed -E 's/.*\t([0-9.]+)%.*/\1/')"
    scripts/coveragebadge "$total_percent"

    echo "$total_line"
    echo "coverage artifacts: .coverage/coverage.out .coverage/coverage-func.txt .coverage/coverage-percent.txt .coverage/coverage.html badges/coverage.svg"

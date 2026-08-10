# Goshtoso task runner. Project workflows live here.

# Build CSS and the demo server.
all: build

build: css
    go build -o bin/server ./site/cmd/server

# Run the development server after rebuilding CSS.
dev: css
    go run ./site/cmd/server

# Show the two-terminal development workflow.
dev-watch: css
    @echo "Starting dev server with CSS watch..."
    @echo "Run these in separate terminals:"
    @echo "  Terminal 1: just css-watch"
    @echo "  Terminal 2: just dev"

# Run with Air for live reloading.
dev-air: css
    @command -v air >/dev/null || { echo "Air not installed. Run: just install-air"; exit 1; }
    air

# Install all local development dependencies.
install: install-templ install-playwright install-air
    go mod download

install-air:
    go install github.com/air-verse/air@latest

install-templ:
    go install github.com/a-h/templ/cmd/templ@latest

install-playwright:
    #!/usr/bin/env bash
    set -euo pipefail
    go install github.com/mxschmitt/playwright-go/cmd/playwright@v0.6100.0
    playwright_bin="$(go env GOBIN)"
    if [ -z "$playwright_bin" ]; then playwright_bin="$(go env GOPATH)/bin"; fi
    "$playwright_bin/playwright" install chromium

# Build assets/styles.css with the Tailwind version and platform binary locked
# by Muamba. The executable cache remains untracked.
css:
    @echo "Building Tailwind CSS..."
    go tool muamba sync --strict tailwindcss/cli
    go run ./cmd/themegen
    .tools/tailwindcss -i css/main.css -o assets/styles.css

# Watch Tailwind CSS for changes.
css-watch:
    @echo "Watching CSS for changes..."
    go tool muamba sync --strict tailwindcss/cli
    go run ./cmd/themegen
    .tools/tailwindcss -i css/main.css -o assets/styles.css --watch

# Materialize locked runtime inputs and regenerate their Go/metadata consumers.
vendor-js:
    go tool muamba sync --strict
    go tool muamba generate-go --strict --dir assets --output muamba_gen.go
    go run ./cmd/runtimegen

# Verify local acquisition bytes and every generated consumer without network.
vendor-js-verify:
    go tool muamba verify --strict
    go tool muamba generate-go --strict --check --dir assets --output muamba_gen.go
    go run ./cmd/runtimegen -check

# Build tracked library and demo-site JavaScript from their owned source roots.
js:
    go run ./cmd/jsbuild

# Parse authored JavaScript, report structural similarity, enforce the inline
# extraction baseline, and verify generated artifacts have no drift.
js-check:
    go run ./cmd/jslint
    go run ./cmd/jsbuild -check

# Generate templ files.
generate:
    templ generate

# Run root and site unit tests.
test:
    go test ./...
    cd site && go test $(go list ./... | grep -v /tests/e2e)

# Run the full Playwright suite.
test-e2e: css
    cd site && go test -tags=e2e,full ./tests/e2e/... -v

# Run E2E identities impacted by committed changes since the supplied base.
test-e2e-focused base="origin/main": css
    go run ./cmd/e2eimpact --base "{{base}}" --head HEAD > .e2e-impact.json
    scripts/run-focused-e2e.sh .e2e-impact.json

# Run one named E2E test.
test-e2e-one test: css
    cd site && go test -tags=e2e,full ./tests/e2e/... -v -run "{{test}}"

# List and run the root-catalog browser agreement in a temporary current-source
# workspace. This test is intentionally absent from pinned v0.1.12 E2E builds.
test-e2e-theme-catalog-current-source:
    scripts/run-focused-e2e.sh --current-source-theme-catalog

# Format root and site Go code.
fmt:
    go fmt ./...
    cd site && go fmt ./...

# Run root and site vet checks.
lint:
    go vet ./...
    cd site && go vet ./...

# Remove local build and E2E artifacts.
clean:
    rm -rf bin/
    rm -rf assets/styles.css
    rm -rf site/tests/e2e/test-results/

# Print available tasks. `just --list` remains the source of task descriptions.
help:
    just --list

# Test the nested site against the root library in this checkout through a
# throwaway workspace. This is the current-source integration contract.
site-current-source-integration:
    ./scripts/check-site-module current-source

# Test the nested site exactly as a standalone consumer of site/go.mod. This is
# the pinned-dependency deployability contract and always forces GOWORK=off.
site-pinned-dependency-deployability:
    ./scripts/check-site-module pinned-dependency

site-module-contracts: site-current-source-integration site-pinned-dependency-deployability

# Run the authoritative release-equivalent full coverage pipeline locally.
coverage:
    scripts/run-release-coverage.sh

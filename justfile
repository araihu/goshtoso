# Goshtoso justfile — see Makefile for a full target list

# Generate templ files
gp-generate:
    templ generate

# Run dev server (builds CSS first)
gp-dev: css
    go run ./site/cmd/server

# Build assets/styles.css with the Tailwind version and platform binary locked
# by Muamba. The executable cache remains untracked.
css:
    go tool muamba sync --strict tailwindcss/cli
    go run ./cmd/themegen
    .tools/tailwindcss -i css/main.css -o assets/styles.css

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

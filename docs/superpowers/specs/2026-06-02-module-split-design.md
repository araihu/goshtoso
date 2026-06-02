# Split Goshtoso into library + site modules

**Date:** 2026-06-02
**Status:** Approved design, pending implementation plan

## Goal

Split the repo into two Go modules in one monorepo:

- **Library** — the publishable UI component library, kept as the *main module* at
  the repo root (`github.com/araihu/goshtoso`). Slim dependency set; what
  consumers `go get`.
- **Site + examples** — the demo website and runnable example apps, moved into a
  nested module at `site/` (`github.com/araihu/goshtoso/site`). Not published;
  carries the heavy/dev-only dependencies.

The motivation: a consumer importing a component should not pull Playwright, the
WebSocket server, or the demo-server's transitive deps.

## Module boundary

### Root module — `github.com/araihu/goshtoso` (library)

Stays at repo root, module path unchanged. Contents:

```
components/                 # UI components (public API)
assets/                     # embedded assets + StylesCSS() accessor, cmd-facing
cmd/goshtoso/               # asset/version extraction CLI (library distribution tool)
scripts/skillgen/           # reads components/ → regenerates using-goshtoso skill ref
scripts/themegen/           # writes assets/goshtoso-theme.css (theme source)
css/  all-themes.css  style.css
go.mod  go.sum
```

Direct dependencies: `github.com/a-h/templ`, `github.com/alecthomas/chroma/v2`
(used by `components/codeblock`), `github.com/stretchr/testify` (test-only).

**Drops** from the library's dependency graph: `playwright-community/playwright-go`,
`coder/websocket`.

The library must build and test standalone with `GOWORK=off` — it never imports
anything under `site/`.

### Site module — `github.com/araihu/goshtoso/site` (demo + examples)

New nested module under `site/`. Go automatically excludes a subtree containing
its own `go.mod` from the parent module, so the root module will not see `site/`.

```
site/
├── go.mod  go.sum
├── cmd/server/             # demo HTTP server (was cmd/server/)
├── internal/
│   ├── server/             # route + HTMX handlers
│   ├── pages/demo/         # demo + docs pages
│   └── examples/           # example app domain logic (chat, logs, profile, ticker, todo)
└── tests/e2e/              # Playwright E2E suite
```

Direct dependencies: the library (`github.com/araihu/goshtoso`), `playwright-go`,
`coder/websocket`, `templ`, `chroma` (transitively via codeblock; may be direct).

Site imports the library through its normal published module path — these import
paths are **unchanged** by the move:
`github.com/araihu/goshtoso/components/...`, `github.com/araihu/goshtoso/assets`.

## Dependency wiring

`go.work` is **local-dev only and gitignored** (it is already in `.gitignore`).
It is never committed and never used by CI.

- **Fresh clone / standalone site build:** resolved from the module proxy via
  `site/go.mod`'s `require github.com/araihu/goshtoso <version>`. No `replace`
  directive.
- **Local development against working-tree library changes:** the developer runs
  `go work init . ./site` once (documented in CLAUDE.md). The resulting
  `go.work` overlays the working tree so edits to `components/`/`assets/` are
  picked up by the site without a release.
- **Initial pin:** `site/go.mod` requires a pseudo-version of the current
  pre-split main HEAD (`39b8ee3`), obtained via `go get github.com/araihu/goshtoso@39b8ee3`.
  This is valid because the split does not modify any library package — the
  library tree at `39b8ee3` is byte-identical to the post-split library API.
  (Note: the existing `v0.0.1` tag is on an orphan commit `1bc96e6`, not an
  ancestor of HEAD, so it is **not** a usable pin.)

## Import rewrite

The only import paths that change are the site's own internal packages:

- `github.com/araihu/goshtoso/internal/*` → `github.com/araihu/goshtoso/site/internal/*`

This touches ~59 `.go` files (the site sources + `cmd/server`). Library import
paths inside those files are left untouched. Mechanics: move the trees, rewrite
with `gofmt -r` / `sed`, then `templ generate` to regenerate `*_templ.go`.

`tests/e2e` imports **no** internal packages (it is black-box: it builds and
hits the server over HTTP), so the only e2e change is the build target path
`./cmd/server` → `./site/cmd/server` and the test directory location.

## Runtime path resolution

`cmd/server/main.go` has `resolveProjectRoot()` used for dev-time CSS/asset path
resolution. After the server moves down one directory level
(`site/cmd/server/`), this resolver must still locate the correct root (the repo
root for dev asset serving, or rely on embedded assets from the library
package). **Implementation must verify and, if needed, fix this resolver** — it
is the one piece of logic the move can silently break.

## Build / CI / tooling changes

| File | Change |
|------|--------|
| `Makefile` | `cmd/server` → `site/cmd/server`; `tests/e2e` → `site/tests/e2e`. `css`/`generate`/`themegen` targets stay at root (library). |
| `justfile` | `gp-dev` runs `site/cmd/server`. `css` target unchanged (themegen + tailwind are library). |
| `Dockerfile` | Copy both module roots and `go.sum`s; build `./site/cmd/server`. Build standalone (no `go.work`) so it resolves the pinned library version, OR copy the whole repo and build the site against the in-repo library via a build-time `go.work` created in the image. Decide at impl time; prefer pinned-version build for reproducibility. fly.toml deploys the site. |
| `.github/workflows/ci.yml` | Run lint + unit tests **per module** (golangci-lint resolves one module per invocation): once at root, once in `site/`. Build `./site/cmd/server`. Run E2E from `site/tests/e2e`. Day-to-day: build the site against `$GITHUB_SHA` ephemerally (`cd site && go get github.com/araihu/goshtoso@$GITHUB_SHA && go mod tidy` in-runner, no commit) so the site is verified against the just-merged library. |
| `.github/workflows/release.yml` | On `v*` tag, after the library release is created, bump and **commit** the site pin: `cd site && go get github.com/araihu/goshtoso@<tag> && go mod tidy`, commit `site/go.mod`+`site/go.sum` back to main with `[skip ci]`. The site tracks released library tags, not every SHA. |
| pre-commit hook (`.githooks`) | `go fix ./...` must run per module (root and `site/`). `skillgen` still reads `components/` at root — unchanged. |

## Documentation

`CLAUDE.md` (and its `AGENTS.md` symlink): update the Repository Structure
section, Quick Commands (paths), and add a short "Two modules / go.work" note
explaining the boundary and the `go work init . ./site` local-dev step.

## Testing / verification (acceptance)

1. `GOWORK=off go build ./... && GOWORK=off go test ./...` at root — library
   builds and component tests pass with no site leakage.
2. In `site/` (with go.work or pinned version): `templ generate` clean,
   `go build ./cmd/server`, `go test ./...` minus e2e green.
3. Full E2E: `go test ./site/tests/e2e/... -count=1 -timeout 15m` — all
   previously-passing tests still pass (381 baseline, no new skips).
4. `golangci-lint run` clean in both module directories.
5. `go run ./scripts/skillgen` produces no diff (skill ref still in sync).
6. Docker image builds and the server boots.

## Out of scope

- No behavioral changes to any component, page, example, or handler.
- No new components or examples.
- No dependency upgrades beyond what `go mod tidy` requires for the split.
- No change to the theme/CSS build pipeline beyond path references.

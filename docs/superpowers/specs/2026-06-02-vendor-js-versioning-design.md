# Design — Versioned vendor JS paths + stated provenance

**Date:** 2026-06-02
**Status:** Approved (design)
**Author:** consumer-friction follow-up session

## Problem

Vendored third-party JS (Alpine, Alpine plugins, HTMX, HTMX extensions) lives
flat at `assets/js/vendor/*.min.js` with **no machine-readable version record**.
The versions are buried inside the minified blobs and a couple of `curl
unpkg@X` lines in old plan docs. This is out of step with the Tailwind
provenance work (PR #17), which pinned the CSS toolchain version in
`assets/tailwind.version` and exposed it via `assets.TailwindVersion()`.

External consumers cannot tell which Alpine/HTMX version Goshtoso ships, and the
asset URLs carry no provenance. The Tailwind side already solved this; the JS
side should match.

## Goal

1. **Versioned paths**: serve each external dep at
   `/assets/js/vendor/<module>/<version>/<file>.js` — the version is in the URL.
2. **Clearly-stated provenance**: a single machine-readable manifest is the
   source of truth for every external dep's name, version, file, and origin,
   exposed through Go accessors parallel to `assets.TailwindVersion()`.

Mirrors the `tailwind.version` + `just css` pattern.

## Decisions (locked during brainstorming)

- **Sync mechanism: codegen.** A manifest is the SSOT; a generator emits Go URL
  constants that `head.templ` references; CI fails if regeneration produces a
  diff. Matches the repo's existing `themegen`/`skillgen` codegen precedent;
  zero drift.
- **Module dir names = npm package names**: `alpinejs`, `alpinejs-collapse`,
  `alpinejs-focus`, `htmx.org`, `htmx-ext-sse`, `htmx-ext-ws`.
- **First-party scripts stay unversioned** at `/assets/js/` (`darkmode.js`,
  `combobox.js`) — they are not external deps.
- **Hard cut, no aliases.** Only versioned paths exist; old flat paths 404. All
  internal refs move to versioned. Consumers on `head.Dependencies()` are
  unaffected; anyone who hardcoded flat paths must switch to `Dependencies()`
  (already the documented path in `USAGE.md`).
- **Vendoring automated** via a `just vendor-js` recipe mirroring `just css`.

## Current versions (extracted / from vendoring docs)

| Module | Version | File | Source URL template |
|--------|---------|------|---------------------|
| alpinejs | 3.14.9 | alpine.min.js | `https://unpkg.com/alpinejs@{v}/dist/cdn.min.js` |
| alpinejs-collapse | 3.14.9 | alpine-collapse.min.js | `https://unpkg.com/@alpinejs/collapse@{v}/dist/cdn.min.js` |
| alpinejs-focus | 3.14.9 | alpine-focus.min.js | `https://unpkg.com/@alpinejs/focus@{v}/dist/cdn.min.js` |
| htmx.org | 2.0.8 | htmx.min.js | `https://unpkg.com/htmx.org@{v}/dist/htmx.min.js` |
| htmx-ext-sse | 2.2.3 | htmx-ext-sse.min.js | `https://unpkg.com/htmx-ext-sse@{v}/dist/sse.min.js` |
| htmx-ext-ws | 2.0.3 | htmx-ext-ws.js | `https://unpkg.com/htmx-ext-ws@{v}/ws.js` |

Alpine core + collapse + focus share one version (single npm monorepo). The
recipe MUST verify each downloaded file against the version it claims (see
Verification) so a wrong/renamed upstream artifact fails loudly rather than
silently shipping the wrong bytes.

## Architecture

### 1. Manifest — `assets/js/vendor/versions.json` (SSOT)

```json
{
  "alpinejs":          { "version": "3.14.9", "file": "alpine.min.js",          "url": "https://unpkg.com/alpinejs@{v}/dist/cdn.min.js" },
  "alpinejs-collapse": { "version": "3.14.9", "file": "alpine-collapse.min.js", "url": "https://unpkg.com/@alpinejs/collapse@{v}/dist/cdn.min.js" },
  "alpinejs-focus":    { "version": "3.14.9", "file": "alpine-focus.min.js",    "url": "https://unpkg.com/@alpinejs/focus@{v}/dist/cdn.min.js" },
  "htmx.org":          { "version": "2.0.8",  "file": "htmx.min.js",            "url": "https://unpkg.com/htmx.org@{v}/dist/htmx.min.js" },
  "htmx-ext-sse":      { "version": "2.2.3",  "file": "htmx-ext-sse.min.js",    "url": "https://unpkg.com/htmx-ext-sse@{v}/dist/sse.min.js" },
  "htmx-ext-ws":       { "version": "2.0.3",  "file": "htmx-ext-ws.js",         "url": "https://unpkg.com/htmx-ext-ws@{v}/ws.js" }
}
```

`{v}` is substituted with `version` by the recipe. The destination path is
derived: `js/vendor/<key>/<version>/<file>`.

### 2. Disk layout

```
assets/js/vendor/
  versions.json
  alpinejs/3.14.9/alpine.min.js
  alpinejs-collapse/3.14.9/alpine-collapse.min.js
  alpinejs-focus/3.14.9/alpine-focus.min.js
  htmx.org/2.0.8/htmx.min.js
  htmx-ext-sse/2.2.3/htmx-ext-sse.min.js
  htmx-ext-ws/2.0.3/htmx-ext-ws.js
```

`assets/embed.go`'s `//go:embed ... js ...` already recurses, so the versioned
subdirs and `versions.json` are embedded with no directive change. Migration is
`git mv` of the six existing files; flat copies deleted.

### 3. Generator — `scripts/vendorgen`

Sibling to `scripts/themegen` and `scripts/skillgen`. Reads `versions.json`,
then:

- Writes `assets/vendor_gen.go` (generated, lint-excluded) with one exported
  URL-path constant per module, plus a sorted slice for tests:
  ```go
  // Code generated by vendorgen — DO NOT EDIT.
  package assets

  const (
      AlpineJSURL        = "/assets/js/vendor/alpinejs/3.14.9/alpine.min.js"
      AlpineCollapseURL  = "/assets/js/vendor/alpinejs-collapse/3.14.9/alpine-collapse.min.js"
      AlpineFocusURL     = "/assets/js/vendor/alpinejs-focus/3.14.9/alpine-focus.min.js"
      HTMXURL            = "/assets/js/vendor/htmx.org/2.0.8/htmx.min.js"
      HTMXExtSSEURL      = "/assets/js/vendor/htmx-ext-sse/2.2.3/htmx-ext-sse.min.js"
      HTMXExtWSURL       = "/assets/js/vendor/htmx-ext-ws/2.0.3/htmx-ext-ws.js"
  )
  ```
  Constant naming is a fixed map from manifest key → Go identifier (kept in the
  generator; new deps require adding the mapping, which is intentional — a new
  external dep is a deliberate change).
- **Verifies on-disk presence**: every declared `js/vendor/<key>/<version>/<file>`
  must exist; missing → generate fails (vendoring not run).
- `-check` mode: regenerate to a temp buffer and compare against the committed
  `vendor_gen.go`; nonzero exit on mismatch. Used by CI.

### 4. `head.templ` consumes the constants

`components/head` imports `github.com/araihu/goshtoso/assets` (no cycle — assets
does not import head). Both entry points switch to attribute expressions:

```go
templ Dependencies() {
    <link rel="stylesheet" href="/assets/styles.css"/>
    <script defer src={ assets.AlpineCollapseURL }></script>
    <script defer src={ assets.AlpineFocusURL }></script>
    <script defer src={ assets.AlpineJSURL }></script>
    <script src={ assets.HTMXURL }></script>
    <script defer src="/assets/js/combobox.js"></script>
}
```

`DependenciesMinimal()` likewise (alpine core + htmx + combobox).

### 5. `embed.go` version accessors (provenance API)

Parallel to `TailwindVersion()`. A single `vendorManifest()` helper reads and
caches `versions.json`; thin accessors expose versions:

```go
func AlpineVersion() string       // "3.14.9"
func HTMXVersion() string         // "2.0.8"
func HTMXExtSSEVersion() string   // "2.2.3"
func HTMXExtWSVersion() string    // "2.0.3"
```

(Alpine plugins share `AlpineVersion()`.) Doc comment lists the deps and points
at `versions.json` as SSOT.

### 6. Vendoring recipe — `just vendor-js`

Mirrors `just css`. For each manifest entry: substitute `{v}`, `curl -fsSL` the
url into a temp file, **verify** (see below), then move to
`assets/js/vendor/<key>/<version>/<file>`. Prune any `vendor/<key>/<otherver>/`
dirs not in the manifest. After download, runs `go run ./scripts/vendorgen` to
regenerate constants.

### 7. Reference updates (hard cut)

- `components/head/head.templ` — constants (§4).
- `internal/pages/demo/layout.templ` — alpine/collapse/focus/htmx + the two
  htmx extensions → `assets.*URL`.
- `internal/pages/demo/components/landing.templ` — alpine + htmx → `assets.*URL`.
- `docs/USAGE.md` — the explicit `<script>` list: state versions, show versioned
  paths; note `head.Dependencies()` is the recommended escape from hardcoding.
- `assets/embed.go` — doc comment paths.
- `CLAUDE.md` — the "bundled locally under `assets/js/vendor/`" note → mention
  versioned subdirs + `versions.json`.

## Verification

- **Recipe download integrity**: after `curl`, assert the file is non-empty and
  (where the artifact embeds a version string — Alpine `version:"X"`, htmx
  `version:"X"`) grep-confirm it matches the manifest version. SSE/WS lack a
  reliable embedded version → assert non-empty + a known token (`htmx:sse` /
  `htmx-ext-ws` marker) instead. A mismatch aborts the recipe.
- **Tests** (`assets/version_test.go`, extended):
  - `versions.json` parses; every entry has version/file/url.
  - every declared file is present in the embedded FS at its versioned path.
  - each accessor returns the manifest version.
  - new `head` test: render `Dependencies()` + `DependenciesMinimal()`, assert
    output contains the versioned paths and **no** flat `vendor/*.min.js` path.
- **CI**: add `go run ./scripts/vendorgen -check` to the staleness job next to
  skillgen; `git diff --exit-code` after generate. Existing E2E suite
  (381 tests) must stay green — the demo site loads the new paths, so a broken
  path 404s and fails the no-console-errors assertions.
- **Behavioral**: rebuild + `go run` the getting-started example and the demo
  server; `curl -I` a versioned path (200) and an old flat path (404); browser
  console clean.

## Out of scope

- Subresource Integrity (SRI) hashes — possible future provenance layer; not now.
- Switching demo `layout.templ`/`landing.templ` to `head.Dependencies()` — they
  have distinct needs (ws/sse); keep hand-rolled, just version the paths.
- Versioning first-party `darkmode.js`/`combobox.js`.

## Risks

- **Upstream URL drift** (unpkg path shape changes per package) — handled by the
  per-entry `url` template in the manifest and the recipe's post-download verify.
- **Constant-name map** must be updated when a dep is added — intentional friction;
  documented in the generator.

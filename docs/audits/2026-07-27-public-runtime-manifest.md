# Public Runtime Manifest and Library Version Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `superpowers:executing-plans` to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish a deterministic, caller-owned Goshtoso runtime/fallback
manifest plus fail-closed build-version identity, and make `head.Dependencies`
render from that contract.

**Architecture:** Package `assets` owns public immutable value types and builds
a fresh default manifest on every call from the same generated URL and SRI
constants used by embedded files. Package `head` copies that manifest into its
per-render config, applies existing functional options to the copy, and derives
both loader JSON and local-only tags from the ordered dependency slice. Version
identity reads Go build information, exposes an exact version only for an
unreplaced module, and reports development, replacement, or unavailable states
without pretending they are a release.

**Tech Stack:** Go 1.26.5, `runtime/debug`, `net/http/httptest`, templ,
Playwright, generated vendorgen constants.

## Global Constraints

- Frozen base: `7576710844cf2e60a5e928c837df534ef5878c04` for both `HEAD`
  and fetched `origin/main` before edits.
- Worktree: isolated Codex worktree on `codex/public-runtime-manifest`;
  primary checkout and other worktrees are read-only and out of scope.
- No Manja-specific API, Manja edits, module-cache scraping, runtime network
  access, exposed `embed.FS`, mutable global manifest, CSP/SRI weakening, or
  future release version hardcode.
- Runtime dependency order: compiled CSS; Alpine Collapse; Alpine Focus;
  Alpine Mask; Alpine core; HTMX; combobox. The default bootstrap loader is
  separately explicit because it is rendered by CDN-first mode but is not a
  runtime dependency to execute twice.
- All declared local URLs must resolve through `assets.Handler()`; every
  non-empty SRI must match the exact served bytes.
- Caller mutation must not affect later manifest calls, functional options, or
  rendering.
- Brand SVGs at `araihu/assets@81300f5`, Goshtoso logos, favicons, themes, and
  unrelated site recipe work remain untouched.
- No push, PR, merge, tag, release, deployment, or Manja mutation.

## Pre-code audit and decisions

- [x] Read repository `AGENTS.md` completely.
- [x] Fetched `origin`; verified clean detached base and exact frozen SHA;
  created only `codex/public-runtime-manifest` in the supplied worktree.
- [x] Read approved Manja design lines 57-90 and 190-240 plus complete rollout
  and acceptance sections. Contract requires public version identity, public
  ordered fallback data, same-version `assets.Handler()` bytes, and fail-closed
  handling when identity is not exact.
- [x] Read `docs/audits/2026-07-26-head-dependency-fallback.md`. Existing
  v0.0.13 loader is sequential, creates a fresh fallback script, waits for
  `window.load` before dynamic HTMX insertion, and keeps combobox first-party.
- [x] Inspected `assets`, `components/head`, vendorgen inputs/outputs, loader,
  unit/E2E tests, `docs/USAGE.md`, `docs/COMPONENT_MODEL.md`, release buildinfo,
  external examples, and generated using-goshtoso references.
- [x] Existing code duplicates stylesheet/loader/combobox paths and the five
  dependency records inside private `head.config`; consumers cannot access the
  complete ordered contract without reading private/generated code.
- [x] Existing public URL/SRI constants cover versioned third-party files, but
  compiled CSS, loader, and combobox URLs have no public constants.
- [x] Existing `site/internal/buildinfo` is demo-only ldflag metadata and cannot
  identify the library bytes linked into an arbitrary consumer.
- [x] API decision: `assets.DefaultRuntimeManifest() RuntimeManifest` returns
  value fields plus a fresh `Dependencies` slice. `Stylesheet` and `Loader` are
  explicit top-level assets; ordered runtime dependencies remain one slice so a
  caller never mistakes the bootstrap loader for an additional runtime script.
- [x] API decision: `assets.GoshtosoVersion() VersionInfo` returns
  `VersionExact`, `VersionDevelopment`, `VersionReplaced`, or
  `VersionUnavailable`. `VersionInfo.Version` stays empty unless status is
  exact. Replacement request and actual replacement metadata use separate
  fields, preventing a replaced `v0.0.13` request from being reported as the
  bytes of release `v0.0.13`.

---

### Task 1: Lock public manifest and handler identity with RED tests

**Files:**
- Create: `assets/runtime_manifest_test.go`
- Test: `assets/runtime_manifest_test.go`

**Interfaces:**
- Consumes: existing public URL/SRI constants and `assets.Handler()`.
- Produces expected API: `DefaultRuntimeManifest() RuntimeManifest`,
  `RuntimeAsset`, `RuntimeAssetKind`, and `RuntimeAssetRole`.

- [x] **Step 1: Write failing API/order test**

  Assert literal roles in order: collapse, focus, mask, Alpine, HTMX,
  combobox. Assert stylesheet and loader roles/kinds/URLs, runtime primary and
  local URLs, SRI, enabled, defer, and readiness flags.

- [x] **Step 2: Run test and preserve RED**

  Run: `go test ./assets -run 'TestDefaultRuntimeManifest' -count=1`

  Expected: compile failure because public manifest types/functions do not
  exist.

- [x] **Step 3: Add mutation isolation and Handler/SRI tests while still RED**

  Mutate every field and the dependency slice returned by one call, then assert
  a later call remains literal-default. Serve every local URL through
  `httptest.NewServer(Handler())`; assert HTTP 200 and verify SHA-384 for every
  non-empty integrity value against response bytes.

### Task 2: Lock fail-closed version behavior with RED tests

**Files:**
- Create: `assets/version.go`
- Create: `assets/library_version_test.go`

**Interfaces:**
- Produces: `GoshtosoVersion() VersionInfo`, `VersionStatus`,
  `VersionReference`, and status constants.
- Internal test seam: `resolveGoshtosoVersion(*debug.BuildInfo, bool)`.

- [x] **Step 1: Write synthetic build-info table test**

  Cover released dependency `v0.0.13`, exact main-module version, main-module
  development, absent module, unavailable build info, local-path replacement,
  and versioned replacement. Only unreplaced exact cases may populate
  `VersionInfo.Version`.

- [x] **Step 2: Run test and preserve RED**

  Run: `go test ./assets -run 'TestResolveGoshtosoVersion' -count=1`

  Expected: compile failure because version identity API does not exist.

### Task 3: Implement smallest public assets contract

**Files:**
- Create: `assets/runtime_manifest.go`
- Create: `assets/version.go`
- Modify: `assets/embed.go`

**Interfaces:**
- `func DefaultRuntimeManifest() RuntimeManifest`
- `func GoshtosoVersion() VersionInfo`
- Public constants for Goshtoso stylesheet, dependency loader, and combobox URLs.

- [x] **Step 1: Implement public value types and literal fresh manifest**

  Use no package-level slice/map. Construct a new dependency slice per call
  from vendorgen constants. Document order, ownership, mount behavior, SRI,
  defer, loader readiness, and local URL semantics.

- [x] **Step 2: Implement build-info resolver**

  Find `github.com/araihu/goshtoso` as main module or dependency. Return
  replacement status before interpreting requested version. Treat empty,
  `(devel)`, and `devel` as development. Keep exact version empty for every
  non-exact status.

- [x] **Step 3: Run focused assets tests GREEN**

  Run: `go test ./assets -count=1`

### Task 4: Make all head render paths derive from public manifest

**Files:**
- Modify: `components/head/types.go`
- Modify: `components/head/head.templ`
- Modify: `components/head/head_test.go`
- Modify: `components/head/head_coverage_test.go`
- Generate: `components/head/head_templ.go`

**Interfaces:**
- Consumes: `assets.DefaultRuntimeManifest()` and its value-owned fields.
- Preserves: every existing `head.Option`, dependency constant, nonce behavior,
  loader JSON/event names, local-only tags, zero-value instance behavior, and
  full/minimal public constructors.

- [x] **Step 1: Write failing manifest/render parity tests**

  Assert default loader JSON matches manifest runtime order and values;
  local-only markup matches manifest local URLs and defer order; minimal omits
  only the three plugin roles; functional options alter only their manifest
  copy; mutation of a public result cannot alter rendering.

- [x] **Step 2: Run focused RED**

  Run: `go test ./components/head -run 'TestDependencies.*Manifest' -count=1`

  Expected: failure because existing private config is not driven by the public
  manifest and does not expose parity helpers.

- [x] **Step 3: Refactor config and templates minimally**

  Store one `assets.RuntimeManifest` per config. Map existing public dependency
  identifiers to roles, apply options to copied asset entries, filter minimal
  roles from the same slice, and range local scripts from that slice.

- [x] **Step 4: Generate and run GREEN**

  Run: `templ generate && go test ./components/head -count=1`

### Task 5: Update public consumer guidance and generated references

**Files:**
- Modify: `docs/USAGE.md`
- Modify: `.agents/skills/using-goshtoso/SKILL.md`
- Generate if changed: `.agents/skills/using-goshtoso/references/components-reference.md`
- Generate if changed: `.claude/skills/using-goshtoso/components-reference.md`
- Modify: this audit record

**Interfaces:**
- Documents exact API, fail-closed status check, immutable result ownership,
  direct `/assets/` mount, loader separation, and replacement behavior.

- [x] **Step 1: Add concise consumer examples and caveats**

  Show caching every declared local URL, checking `VersionExact`, and rejecting
  development/replaced/unavailable identity for same-version offline readiness.

- [x] **Step 2: Run `go run ./cmd/skillgen`**

  Preserve generated reference parity. Record zero drift if new API does not
  change component package declarations.

- [x] **Step 3: Record all snags and API tradeoffs**

  Append exact source dives, loader-order distinction, build-info replacement
  hazard, generator behavior, and any deferred gate receipt to this document.

### Task 6: Progressive assurance and delivery

**Files:**
- Verify all owned paths only.
- Commit all intended changes locally.

- [x] **Step 1: Focused and generator gates**

  Run assets/head tests, `templ generate`, `go run ./cmd/vendorgen -check`,
  `go run ./cmd/skillgen`, and a second full generation/check pass. Require zero
  second-pass drift.

- [x] **Step 2: Standalone external consumer gate**

  Create a temporary module outside the repository, require Goshtoso, replace
  it with this worktree, and run a real compile/test under `GOWORK=off`. Assert
  manifest use and `VersionReplaced` behavior.

- [x] **Step 3: Root and site gates**

  Run root and site `go test`, `go vet`, build, `go fix`, and `golangci-lint`
  against the in-repo workspace. Confirm `go fix` causes no unexplained drift.

- [x] **Step 4: Browser fallback gate**

  Run
  `go test ./site/tests/e2e/... -count=1 -timeout 5m -run TestDependenciesCDNFailureLoadsOrderedLocalFallback`
  and require forced primary failures, ordered local readiness, and all runtime
  behaviors.

- [x] **Step 5: Review, determinism, ownership, and commit**

  Review public API and security/current-path behavior; correct confirmed
  findings; rerun affected gates. Run `git diff --check`, inspect literal diff
  and changed paths, commit locally, then prove clean worktree and exact base
  ancestry. No push.

## Evidence log

RED/GREEN outputs, exact gate results, review findings, commit SHA, snags, and
any permitted assurance receipts are appended here during execution.

- RED 1, missing public manifest API:
  `go test ./assets -run TestDefaultRuntimeManifestHasCompleteOrderedContract -count=1`
  failed to compile with `undefined: DefaultRuntimeManifest`,
  `undefined: RuntimeAsset`, and missing public role/kind/URL constants.
- GREEN 1: the same focused order/API test passed after the smallest public
  value contract was added.
- RED 2, Handler/SRI identity:
  `go test ./assets -run 'TestDefaultRuntimeManifest(IsCallerOwned|LocalURLsMatchHandlerBytesAndSRI)' -count=1`
  reached every declared local URL through `Handler()` but failed with
  `manifest SRI count = 0, want 5 version-matched third-party dependencies`.
  Caller-mutation isolation already passed because the minimal constructor
  returns a fresh value-owned slice rather than a package global.
- GREEN 2: all declared local URLs returned HTTP 200; all five public SRI
  values matched SHA-384 of exact Handler response bytes; mutation isolation
  remained green.
- RED 3, fail-closed library version identity:
  `go test ./assets -run TestResolveGoshtosoVersionDistinguishesExactDevelopmentUnavailableAndReplaced -count=1`
  failed to compile with missing `VersionInfo`, `GoshtosoModulePath`, and
  status constants. The table already covered released dependency,
  development, unavailable, local replacement, and versioned replacement.
- GREEN 3: all version-resolution table cases passed. Non-exact cases exposed
  no `Version` or `Sum`; replacement request and selected replacement metadata
  remained separate.
- RED 4, rendered source parity and copy isolation:
  `go test ./components/head -run 'TestDependencies(RenderFromPublicManifestCopy|MinimalFiltersPublicManifestOrder|LocalRuntimeUsesPublicManifestOrderAndDefer)' -count=1`
  failed to compile with `undefined: newConfigFromManifest`, proving the
  existing private config had no path from the public contract.
- GREEN 4: injected-manifest tests passed for CDN-first loader JSON, minimal
  filtering, option application, local direct-tag order/defer, and mutation
  after config construction. Full `go test ./components/head -count=1` passed.
- External-consumer first run exposed a test-harness snag:
  `GoshtosoVersion status = "unavailable", want "replaced"`. `go version -m`
  confirmed the test binary omitted test-only dependency records, while a
  normal `cmd/probe` binary included `dep github.com/araihu/goshtoso v0.0.13`
  and `=> ../../.. (devel)` and returned `VersionReplaced`.
- Primary CodeRabbit review (`0.7.0`, uncommitted scope) reported four items.
  Accepted missing `fmt` import in the consumer snippet and local-only
  stylesheet localization. Rejected the empty-integrity guard because empty is
  the documented/tested SRI-disable escape hatch. Rejected arbitrary manifest
  reorder handling because no public head API accepts a caller manifest;
  `DefaultRuntimeManifest` order and rendered parity are already locked.
- RED 5, review correction:
  `go test ./components/head -run TestDependenciesLocalRuntimeUsesPublicManifestOrderAndDefer -count=1`
  rendered `href="https://primary.example/styles.css"` in local-only mode
  instead of the manifest local URL.
- GREEN 5: local-only stylesheet regression and full head/docs suites passed;
  scoped CodeRabbit re-review returned zero findings.
- RED 6, literal self-review correction:
  `go test ./assets ./components/head -run 'Test(DefaultRuntimeManifestHasCompleteOrderedContract|DependenciesMinimalFiltersPublicManifestOrder)' -count=1`
  failed to compile because `RuntimeAsset.IncludeInMinimal` did not exist.
  Private role filtering meant the public manifest was not yet the sole source
  of truth for `DependenciesMinimal` membership.
- GREEN 6: public manifest now owns minimal-set membership; custom-manifest
  parity and full assets/head suites passed with no private role switch.
- Generator assurance: two consecutive runs of `templ generate`,
  `go run ./cmd/vendorgen`, `go run ./cmd/vendorgen -check`, and
  `go run ./cmd/skillgen` preserved identical SHA-256 values for generated
  templ, vendorgen, and both skill reference outputs. Final generator pass also
  reported zero templ updates. CSS inputs/utilities did not change, so
  `just css` was not applicable.
- External consumer: `GOWORK=off go test ./... -count=1` passed in
  `tests/external/runtime-manifest`. Its normal executable reported requested
  `v0.0.13`, local replacement metadata, empty exact version, and
  `VersionReplaced`; public manifest and Handler use compiled without private
  imports.
- Root and site: `go fix ./...`, `go test`, `go vet`, `go build`, and
  `golangci-lint run` all passed against the in-repo workspace; both lints
  reported `0 issues`.
- Browser: representative
  `TestDependenciesCDNFailureLoadsOrderedLocalFallback` passed after final
  corrections. Full E2E also passed: `ok .../site/tests/e2e 342.020s`.
- Review: primary CodeRabbit review produced two accepted and two rejected
  findings as recorded above; correction tests passed; scoped re-review
  returned zero findings. Literal self-review then moved minimal membership
  into the public manifest and reran all affected gates.
- `git diff --check` passed. No deferrable gate was skipped; no quality-debt
  receipt is required.
- Staged-scope CodeRabbit closure reviewed all 15 owned paths, including new
  assets API and external fixture files. Accepted two audit-only findings:
  remove machine-specific worktree layout and synchronize completed
  checkboxes. Rejected mandatory-checksum findings because Go build information
  can omit `Sum` for an exact versioned main module and the contract requires an
  exact unreplaced version; `Sum` remains supplementary when present. Rejected
  extra external loader-body/Content-Type assertions because Handler body/SRI
  identity is already exhaustive in assets unit tests and those loader headers
  are not public contract. No further assurance-only review wave is scheduled.

## Snags and API tradeoffs

- Rendered CDN-first HTML contains stylesheet plus one bootstrap loader tag,
  while local-only HTML contains stylesheet plus the ordered runtime scripts.
  A flat list would invite consumers to execute the loader and scripts twice.
  `RuntimeManifest` therefore exposes `Stylesheet`, `Loader`, and one ordered
  `Dependencies` slice with explicit roles.
- Go build information records the requested dependency version even when a
  `replace` directive selects different bytes. Reporting that requested value
  as the library version would violate same-version cache identity. Replaced
  builds therefore expose separate request/replacement records and leave the
  exact `Version` empty.
- Stylesheet, loader, and combobox paths were private string literals in
  `components/head`. They now have public assets constants and appear once in
  `DefaultRuntimeManifest`; head options mutate only a per-render copy.
- Vendorgen remains the canonical generator for third-party CDN/local URL and
  SRI constants from `versions.json` plus embedded bytes. The public manifest
  references those constants instead of reproducing their values.
- `skillgen` indexes component packages, not the `assets` package. The public
  assets API therefore needs explicit consumer-guide and using-goshtoso skill
  guidance; generated component references should remain unchanged.
- Previous manual-tag documentation required source-diving and copied versioned
  paths. It remains a low-level escape hatch; consumers needing an inventory
  now use the public manifest instead.
- Go test binaries omit module records for dependencies imported only by
  external `_test` packages. The first external gate therefore reported
  `VersionUnavailable` even though a normal consumer binary built from the
  same module reports the expected replacement metadata. The durable fixture
  runs a normal `cmd/probe` executable under `GOWORK=off`; unit tests continue
  to cover synthetic unavailable metadata as a valid fail-closed state.
- First root lint pass found the now-unused private `config.nonceAttributes`
  helper after local rendering moved to the manifest loop. Removed the dead
  helper; nonce behavior remains covered on loader and every local script.

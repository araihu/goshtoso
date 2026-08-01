# Focused E2E Selection and Demo Package Segmentation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep every Go unit test unconditional while selecting only the
Playwright E2E identities affected by ordinary changes, with the authoritative
full E2E and merged-coverage suite required before releases.

**Architecture:** Preserve the single E2E package, server, and Chromium
lifecycle. Give each component and runnable example a leaf demo-page package,
derive reverse consumers with the Go package graph, and compile one E2E binary
from boolean build constraints. Unknown, shared, destructive, or
history-ambiguous changes fail safe to `full`.

**Tech Stack:** Go 1.26.5, `go/build/constraint`, `go list -json`, templ,
Tailwind CSS v4, Playwright Go, GitHub Actions, `go tool covdata`.

## Global Constraints

- Work only in `/private/tmp/gs-segment-demo-pages` on branch
  `refactor/segment-demo-page-packages`.
- Keep all root and site unit-test packages unconditional in every workflow.
- Keep `site/tests/e2e` as one Go package with one `TestMain`, one demo server,
  and one shared Chromium process.
- Never hand-edit `*_templ.go`, `assets/styles.css`, generated theme CSS, or
  generated runtime constants.
- Run `templ generate` after moving or editing `.templ` sources. Run `just css`
  only when authored markup changes Tailwind utilities.
- Preserve every route, title, active-navigation key, description, fragment
  response, and Alpine/HTMX behavior during package moves.
- Explicit imports construct registries. Do not use registration through
  `init` functions.
- A focused identity must compile independently and list at least one test.
- Rename, deletion, force-push, missing-history, unknown-path, and shared-runtime
  cases select `full`.
- Each task ends with the stated focused verification and a dedicated commit.

---

## Locked Identities and Page Batches

Component E2E identities match the publishable package names:

```text
accordion actiongroup alert appshell avatar badge banner breadcrumbs button
card carousel chatbubble checkbox codeblock combobox drawer dropdown emptystate
fileinput form head icon kbd link modal navbar pageheader
pagination palette panel radio range rating schemaform search select sidebar
skeleton spinner steps structuredinput table tabs tagslist textarea textinput
toast toggle toolbar tooltip
```

Runnable example identities are:

```text
example_chat example_expense example_logs example_profile example_ticker
example_todo example_wizard
```

Non-identity pages are grouped as follows and always participate only through
`full` or through an impacted component/example leaf that imports them:

| Package | Existing pages |
| --- | --- |
| `contentpages/docs` | agents, application patterns, component model, theme |
| `contentpages/legal` | attributions, license, privacy |
| `contentpages/modules` | module index, charts showcase |
| `contentpages/start` | landing, getting started |
| `examplepages/index` | examples index |

---

### Task 1: Introduce neutral page definitions and a validated registry

**Files:**

- Modify: `site/internal/pages/demo/component_demo.templ`
- Create: `site/internal/pages/demo/page_definition.go`
- Create: `site/internal/pages/demo/registry/registry.go`
- Create: `site/internal/pages/demo/registry/registry_test.go`
- Modify: `site/internal/pages/demo/components/registry.go`

**Interfaces:**

- Produce `demo.PageDefinition` with `Key`, `Title`, `Active`, `Description`,
  `Type`, and `Content func() templ.Component`.
- Produce immutable registry construction plus `Lookup`, `MetaForKey`, and
  `AllPublicMeta` operations.
- Temporarily preserve current `components.LookupDemo` callers through a thin
  compatibility facade.

- [ ] **Step 1: Write failing registry validation tests**

Cover valid construction, duplicate key, empty key, nil content factory,
leading-slash/malformed key, and missing component-catalog entry. Assert returned
metadata slices cannot mutate registry state.

- [ ] **Step 2: Prove the API is absent**

Run:

```bash
(cd site && go test ./internal/pages/demo/registry -count=1)
```

Expected: compile failure because registry construction does not exist.

- [ ] **Step 3: Implement `PageDefinition` and registry validation**

Keep the registry independent of leaf packages in this step by accepting a
slice of definitions in its constructor. Use catalog data supplied by the
caller to validate public component coverage without importing the old
`components` page package.

- [ ] **Step 4: Add the temporary facade and regenerate templ output**

Run:

```bash
templ generate
(cd site && go test ./internal/pages/demo/... -count=1)
```

Expected: all demo-page tests pass with existing callers unchanged.

- [ ] **Step 5: Commit**

```bash
git add site/internal/pages/demo
git commit -m "refactor(site): add validated demo page registry"
```

---

### Task 2: Prove the leaf-package model with Button and ActionGroup

**Files:**

- Move: `site/internal/pages/demo/components/button.templ` to
  `site/internal/pages/demo/componentpages/button/button.templ`
- Move: `site/internal/pages/demo/components/actiongroup.templ` to
  `site/internal/pages/demo/componentpages/actiongroup/actiongroup.templ`
- Create: `site/internal/pages/demo/componentpages/button/definition.go`
- Create: `site/internal/pages/demo/componentpages/actiongroup/definition.go`
- Modify: `site/internal/pages/demo/components/registry.go`
- Create: `site/internal/pages/demo/registry/dependency_smoke_test.go`
- Regenerate: the corresponding `*_templ.go` files

- [ ] **Step 1: Add failing definition and registry tests**

Assert both definitions preserve their exact keys and metadata, registry lookup
renders non-empty HTML, and the public catalog remains covered.

- [ ] **Step 2: Move authored templates and implement definitions**

The package names are `buttonpage` and `actiongrouppage` so their library
imports retain the short names `button` and `actiongroup`. Delete old generated
files and use `templ generate` to create new generated files.

- [ ] **Step 3: Prove the dependency boundary**

Run:

```bash
go list -deps ./site/internal/pages/demo/componentpages/button
go list -deps ./site/internal/pages/demo/componentpages/actiongroup
go list -json ./site/internal/pages/demo/componentpages/button
```

Expected: Button's leaf imports `components/button`; ActionGroup's leaf imports
both its direct component dependencies; neither leaf imports the registry.

- [ ] **Step 4: Verify pilot behavior**

```bash
templ generate
(cd site && go test ./internal/pages/demo/... ./internal/server/... -count=1)
go build -o bin/server ./site/cmd/server
```

- [ ] **Step 5: Commit**

```bash
git add site/internal/pages/demo
git commit -m "refactor(site): segment button demo page packages"
```

---

### Task 3: Move primitive and display component pages

**Files:**

- Create leaf directories under
  `site/internal/pages/demo/componentpages/` for:
  `accordion`, `alert`, `avatar`, `badge`, `banner`, `breadcrumbs`, `card`,
  `chatbubble`, `codeblock`, `emptystate`, `head`, `icon`,
  `kbd`, `link`, `palette`, `panel`, `skeleton`, `spinner`, `steps`, `toolbar`,
  and `tooltip`.
- Move the matching authored `.templ` files out of
  `site/internal/pages/demo/components/`.
- Create `definition.go` in every public component identity leaf. The existing
  `dependencies.templ` source moves to the `head` leaf because the catalog
  explicitly maps that route to the publishable `components/head` package.
- Modify: `site/internal/pages/demo/components/registry.go`
- Regenerate: all moved `*_templ.go` files

- [ ] **Step 1: Add a table-driven registry parity test**

Lock each moved route key, title, active key, type, and non-nil content factory
before moving files.

- [ ] **Step 2: Move the authored sources in four small groups**

After each group run:

```bash
templ generate
(cd site && go test ./internal/pages/demo/... -count=1)
```

The groups are: feedback/status; content/display; structure; utility.

- [ ] **Step 3: Check leaf dependency direction**

```bash
go list -deps ./site/internal/pages/demo/componentpages/... >/tmp/gs-componentpage-deps.txt
! rg 'site/internal/pages/demo/registry' /tmp/gs-componentpage-deps.txt
```

- [ ] **Step 4: Commit**

```bash
git add site/internal/pages/demo
git commit -m "refactor(site): segment primitive demo pages"
```

---

### Task 4: Move navigation, form, and composite component pages

**Files:**

- Create leaf directories and definitions for:
  `appshell`, `carousel`, `checkbox`, `combobox`, `drawer`, `dropdown`,
  `fileinput`, `form`, `modal`, `navbar`, `pageheader`, `pagination`, `radio`,
  `range`, `rating`, `schemaform`, `search`, `select`, `sidebar`,
  `structuredinput`, `table`, `tabs`, `tagslist`, `textarea`, `textinput`,
  `toast`, and `toggle`.
- Move the matching `.templ` sources plus tightly owned non-generated helpers,
  including form validation and table-specific page helpers.
- Move package-local tests alongside their owned source when they test only
  that leaf.
- Modify server handler imports for exact fragment owners where necessary.
- Modify: `site/internal/pages/demo/components/registry.go`
- Regenerate: all moved `*_templ.go` files

- [ ] **Step 1: Extend registry parity tests for the remaining identities**

Expected failure: leaf definitions do not exist.

- [ ] **Step 2: Move pages in dependency-aware batches**

Use these batches and run demo/server tests after each:

```text
navigation: appshell drawer dropdown modal navbar pageheader pagination sidebar tabs
forms: checkbox combobox fileinput form radio range select textarea textinput toggle
composites: carousel rating schemaform search structuredinput table tagslist toast
```

Verification per batch:

```bash
templ generate
(cd site && go test ./internal/pages/demo/... ./internal/server/... -count=1)
```

- [ ] **Step 3: Run complete root and site unit suites**

```bash
go test ./... -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
```

- [ ] **Step 4: Commit**

```bash
git add site/internal/pages/demo site/internal/server
git commit -m "refactor(site): segment composite demo pages"
```

---

### Task 5: Segment runnable examples and remaining content pages

**Files:**

- Move each runnable example from `site/internal/pages/demo/examples/` into
  `site/internal/pages/demo/examplepages/{chat,expense,logs,profile,ticker,todo,wizard}/`.
- Create one definition per example using tags `example_<name>`.
- Move the examples index to `examplepages/index`.
- Move remaining authored pages to `contentpages/{docs,legal,modules,start}`
  according to the locked table.
- Modify exact handlers in `site/internal/server/{chat,expense,logs,profile,ticker,todo,wizard}_handler.go`.
- Create: `site/internal/pages/demo/registry/default.go`
- Modify all callers of the old `components.LookupDemo`, `MetaForKey`, and
  `AllPublicMeta` APIs.
- Delete: `site/internal/pages/demo/components/registry.go` after cutover.
- Regenerate: all moved `*_templ.go` files

- [ ] **Step 1: Lock example and content route parity in tests**

Assert all old route keys resolve through the new default registry and catalog
coverage is exact—no missing or extra component entries.

- [ ] **Step 2: Move examples and cut handlers to exact leaves**

Run after each example pair:

```bash
templ generate
(cd site && go test ./internal/pages/demo/... ./internal/server/... -count=1)
```

- [ ] **Step 3: Move grouped content and remove the compatibility facade**

```bash
rg -n 'pages/demo/components|components\.(LookupDemo|MetaForKey|AllPublicMeta)' site --glob '*.go' --glob '*.templ'
```

Expected: no old page-package imports or registry calls.

- [ ] **Step 4: Verify the full untagged E2E suite before tag work**

```bash
templ generate
go build -o bin/server ./site/cmd/server
go test ./... -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
go test ./site/tests/e2e/... -count=1 -timeout 15m
```

- [ ] **Step 5: Commit**

```bash
git add site/internal/pages/demo site/internal/server
git commit -m "refactor(site): complete demo page package segmentation"
```

---

### Task 6: Extract every cross-file E2E declaration

**Files:**

- Create support files under `site/tests/e2e/` grouped by responsibility:
  `page_support_test.go`, `alpine_support_test.go`, `attribute_support_test.go`,
  `assertion_support_test.go`, and `fixture_support_test.go` as required by the
  type-aware inventory.
- Modify all E2E files that currently declare helpers consumed elsewhere.
- Create: `site/cmd/e2econstraints/main.go`
- Create: `site/internal/e2econstraints/inventory.go`
- Create: `site/internal/e2econstraints/inventory_test.go`
- Modify: `site/go.mod`, `site/go.sum` to add the tool-only
  `golang.org/x/tools/go/packages` dependency without growing the publishable
  root module.

- [ ] **Step 1: Write an AST/type-check inventory test**

Load `site/tests/e2e` with `go/packages`, record each package-level declaration
and every source file that refers to it, and fail when a declaration in a
runnable identity file has cross-file consumers. Exempt declarations in named
support files and `Test*` functions.

- [ ] **Step 2: Confirm the inventory finds known blockers**

```bash
(cd site && go test ./internal/e2econstraints -run TestInventory -count=1)
```

Expected initial findings include `newIsolatedPage` from
`todo_example_test.go`, `waitForAlpine` from `modal_test.go`, and
`mustAttribute` from `card_test.go`.

- [ ] **Step 3: Move all reported declarations into support files**

Do not add build constraints yet. Preserve helper signatures and behavior.

- [ ] **Step 4: Prove the real full suite still compiles and passes**

```bash
(cd site && go test ./internal/e2econstraints ./cmd/e2econstraints -count=1)
(cd site && go test -c -o /tmp/gs-e2e-untagged.test ./tests/e2e)
go test ./site/tests/e2e/... -count=1 -timeout 15m
```

- [ ] **Step 5: Commit**

```bash
git add site/cmd/e2econstraints site/internal/e2econstraints site/go.mod site/go.sum site/tests/e2e
git commit -m "refactor(e2e): extract shared test support"
```

---

### Task 7: Atomically gate E2E and migrate every active caller

**Files:**

- Modify every `.go` file in `site/tests/e2e/`.
- Modify: `site/tests/e2e/e2e_test.go`
- Create: `site/internal/e2econstraints/callers.go`
- Create: `site/internal/e2econstraints/callers_test.go`
- Modify: `Makefile`, `Justfile`, `site/tests/e2e/Dockerfile`
- Modify: `.github/workflows/ci.yml`, `.github/workflows/release.yml`
- Modify: `AGENTS.md`, `README.md`, `CONTRIBUTING.md`,
  `.github/pull_request_template.md`, `docs/RELEASE_CHECKLIST.md`, and
  `site/tests/e2e/README.md`

- [ ] **Step 1: Add a failing active-caller scan**

Reject executable or currently documented commands that target
`site/tests/e2e` or `./tests/e2e` without `-tags=e2e,...`. Exclude only
historical audit records and `docs/plans`.

- [ ] **Step 2: Add list-only lifecycle coverage**

Test that `GOSHTOSO_E2E_LIST_ONLY=1` makes `TestMain` call `m.Run()` without
building the server or launching Chromium.

- [ ] **Step 3: Apply the suite gate and migrate callers in one diff**

Every E2E Go source gets `//go:build e2e`. Every existing full-suite caller
gets `-tags=e2e,full`. `test-e2e-one` retains `-run` but is still a full build.

- [ ] **Step 4: Prove positive and negative package discovery**

```bash
(cd site && go test -c -tags=e2e,full -o /tmp/gs-e2e-full.test ./tests/e2e)
(cd site && GOSHTOSO_E2E_LIST_ONLY=1 go test -tags=e2e,full -list '^Test' ./tests/e2e)
(cd site && ! go test -c -o /tmp/gs-e2e-bare.test ./tests/e2e)
(cd site && go test ./internal/e2econstraints ./cmd/e2econstraints -count=1)
```

Expected: tagged compile/list succeeds; bare E2E package selection fails; the
caller scan passes.

- [ ] **Step 5: Commit**

```bash
git add .github AGENTS.md CONTRIBUTING.md Justfile Makefile README.md docs/RELEASE_CHECKLIST.md site/cmd/e2econstraints site/internal/e2econstraints site/tests/e2e
git commit -m "test(e2e): require explicit suite build tags"
```

---

### Task 8: Assign identities and prove every focused binary inventory

**Files:**

- Modify runnable test files under `site/tests/e2e/` to use expressions such as
  `//go:build e2e && (full || button || actiongroup)`.
- Keep harness/support sources at `//go:build e2e`.
- Create: `site/tests/e2e/identities.json`
- Extend: `site/internal/e2econstraints` and `site/cmd/e2econstraints` with
  build-expression parsing, compile-matrix, and list-inventory validation.

- [ ] **Step 1: Write constraint-parser tests**

Use `go/build/constraint`—never a handwritten parser. Test single identities,
multi-identity expressions, suite-only support, missing `full`, unknown tags,
duplicate identities, and malformed constraints.

- [ ] **Step 2: Add the complete identity manifest**

For every identity record the files and expected `Test*` names. Global docs,
theme, security, page-failure, shutdown, and release-contract tests are
`full`-only.

- [ ] **Step 3: Apply constraints to runnable files**

Cross-component files list every component they intentionally exercise. Example
files use only their `example_*` identity unless they are explicitly composite.

- [ ] **Step 4: Run real compile and list matrices**

For every identity, the validator creates a temporary binary directory and
executes the equivalent of this loop:

```bash
matrix_dir=$(mktemp -d)
for identity in accordion actiongroup alert appshell avatar badge banner breadcrumbs button card carousel chatbubble checkbox codeblock combobox drawer dropdown emptystate fileinput form head icon kbd link modal navbar pageheader pagination palette panel radio range rating schemaform search select sidebar skeleton spinner steps structuredinput table tabs tagslist textarea textinput toast toggle toolbar tooltip example_chat example_expense example_logs example_profile example_ticker example_todo example_wizard; do
  (cd site && go test -c -tags="e2e,$identity" -o "$matrix_dir/$identity.test" ./tests/e2e)
  (cd site && GOSHTOSO_E2E_LIST_ONLY=1 go test -tags="e2e,$identity" -list '^Test' ./tests/e2e)
done
(cd site && go test -c -tags=e2e,full -o "$matrix_dir/full.test" ./tests/e2e)
(cd site && GOSHTOSO_E2E_LIST_ONLY=1 go test -tags=e2e,full -list '^Test' ./tests/e2e)
```

Also run `e2e,full`. The validator must assert that the actual list equals the
tests predicted from parsed constraints, each identity lists at least one test,
and `full` is the exact union of all runnable tests plus full-only tests.

- [ ] **Step 5: Run one real focused Playwright smoke**

```bash
go test -tags=e2e,button ./site/tests/e2e -count=1 -timeout 5m
```

Expected: Button plus intentional composite consumers run; unrelated tests do
not appear.

- [ ] **Step 6: Commit**

```bash
git add site/cmd/e2econstraints site/internal/e2econstraints site/tests/e2e
git commit -m "test(e2e): add component identity constraints"
```

---

### Task 9: Implement conservative Go-graph impact selection

**Files:**

- Create: `cmd/e2eimpact/main.go`
- Create: `internal/e2eimpact/{gitdiff,graph,classifier,handlers,result}.go`
- Create matching `_test.go` files and `testdata/` fixtures.
- Refactor mixed component endpoint implementations out of
  `site/internal/server/server.go` into identity-owned handler files.

**CLI contract:**

```bash
go run ./cmd/e2eimpact --base "$BASE_SHA" --head "$HEAD_SHA"
```

Emit stable JSON with `mode` (`focused` or `full`), sorted `tags`, sorted
`changed_paths`, and sorted human-readable `reasons`.

- [ ] **Step 1: Write Git range fixtures**

Cover modify, add, delete, rename detected by `-M`, multi-commit push,
branch-behind PR merge-base, all-zero `before`, force/non-ancestor update,
missing base, and shallow/missing history. Delete/rename/unsafe history must
return `full` and report old/new paths.

- [ ] **Step 2: Implement NUL-safe Git parsing**

Execute `git diff --name-status -z -M <start> <head>` and parse NUL records.
PR callers supply a fetched merge base. Main callers supply event `before` and
`github.sha`.

- [ ] **Step 3: Write graph-selection tests**

Load both root and site graphs using `go list -json`. Starting from changed
component packages, traverse reverse imports to leaf page packages, convert
only leaf identities to E2E tags, and ignore registry/server aggregators as
identities.

- [ ] **Step 4: Implement authored-versus-derived classification**

When authored component `.templ` or Go sources explain regenerated
`*_templ.go` and `assets/styles.css`, ignore those known derived paths. A
generated-only change, direct global CSS/theme/runtime/shared JS/assets-handler
change, or unexplained path selects `full`.

- [ ] **Step 5: Add checked server-handler ownership**

Map exact handler files to one or more identities. Move table/search/toast and
example endpoint bodies out of mixed `server.go` where needed. Shared routing,
middleware, lifecycle, storage, or unclassified server files select `full`.

- [ ] **Step 6: Prove Button reverse consumers locally**

```bash
go run ./cmd/e2eimpact --base origin/main --head HEAD
go test ./internal/e2eimpact/... ./cmd/e2eimpact/... -count=1
```

Use fixture commits for the focused assertion so the design/plan changes on
the branch do not force this live branch invocation to be focused.

- [ ] **Step 7: Commit**

```bash
git add cmd/e2eimpact internal/e2eimpact site/internal/server
git commit -m "ci: select impacted e2e identities"
```

---

### Task 10: Wire unconditional unit tests and focused E2E in CI

**Files:**

- Create: `scripts/run-focused-e2e.sh`
- Modify: `.github/workflows/ci.yml`
- Modify: `Makefile`, `Justfile`, `site/tests/e2e/README.md`

- [ ] **Step 1: Add script contract tests**

Test no-tag/full/focused JSON results, sorted comma-joined tag construction,
and preservation of one Go invocation. Reject an empty focused set.

- [ ] **Step 2: Implement the focused runner**

It receives impact JSON, chooses `e2e,full` or
`e2e,<tag1>,<tag2>`, prints the decision, and runs exactly one `go test`
invocation with the selected tags.

- [ ] **Step 3: Wire PR and main range semantics**

Fetch sufficient history. PR uses `git merge-base` with the fetched base. Main
uses `github.event.before` and `github.sha`; all-zero/force/missing cases become
full through the selector. Upload the decision JSON as a diagnostic artifact.

- [ ] **Step 4: Keep unit suites unconditional**

Verify the workflow always runs:

```bash
go test ./... -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
```

Partial E2E coverage may be uploaded as diagnostics but never updates the
coverage badge.

- [ ] **Step 5: Validate workflow and local focused command**

```bash
go test ./internal/e2eimpact/... ./cmd/e2eimpact/... -count=1
(cd site && go test ./internal/e2econstraints/... ./cmd/e2econstraints/... -count=1)
shellcheck scripts/run-focused-e2e.sh
go test -tags=e2e,button ./site/tests/e2e -count=1 -timeout 5m
```

- [ ] **Step 6: Commit**

```bash
git add .github/workflows/ci.yml Justfile Makefile scripts site/tests/e2e/README.md
git commit -m "ci: run focused playwright identities"
```

---

### Task 11: Make release coverage authoritative and reproducible

**Files:**

- Create: `scripts/component-coverpkg.sh`
- Create: `scripts/run-release-coverage.sh`
- Create matching shell contract tests under `scripts/testdata/` or Go tests
  under `internal/e2eimpact` that execute the scripts.
- Modify: `Justfile`
- Modify: `.github/workflows/release.yml`
- Modify: `docs/RELEASE_CHECKLIST.md`

- [ ] **Step 1: Lock the coverage denominator**

The publishable denominator is the exact sorted result of
`go list ./components/...`. Join it as a comma-separated `-pkg` filter. Site,
command, and server packages may be instrumented for E2E execution but never
enter the badge denominator.

- [ ] **Step 2: Create one reusable coverage pipeline**

The script runs root and site coverage, builds/runs the instrumented full E2E
suite with `-tags=e2e,full`, merges covdata, and invokes both
`go tool covdata percent -pkg="$component_coverpkg"` and
`go tool covdata textfmt -pkg="$component_coverpkg"`. The final percentage is
read from `go tool cover -func` over the filtered profile.

- [ ] **Step 3: Complete the pre-release job**

Install Go, Node, templ, Tailwind, Playwright, and Chromium; create the
root/site workspace; run regeneration/drift checks; run both site module
contracts; run the reusable full coverage script; upload covdata/profile; and
publish the badge only after every gate passes.

- [ ] **Step 4: Add manual dry-run behavior**

`workflow_dispatch` runs the identical full gate but skips release publication
and badge mutation. Release events never call the focused selector.

- [ ] **Step 5: Assert local/release parity**

On the same commit, the local Just target and reusable release script must emit
the exact same sorted component package list and percentage string.

- [ ] **Step 6: Commit**

```bash
git add .github/workflows/release.yml Justfile docs/RELEASE_CHECKLIST.md scripts
git commit -m "ci(release): require authoritative full e2e coverage"
```

---

### Task 12: Run final safety, behavior, and documentation gates

**Files:**

- Modify any affected current documentation or tests discovered by the gates.
- Do not modify historical audit records solely to rewrite old commands.

- [ ] **Step 1: Regeneration and dirty-output checks**

```bash
templ generate
just css
git diff --exit-code -- '*_templ.go' assets/styles.css
```

If authored markup did not change utilities, `just css` must leave
`assets/styles.css` unchanged.

- [ ] **Step 2: Static and unit gates**

```bash
(cd site && go run ./cmd/e2econstraints)
go test ./... -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
golangci-lint run
(cd site && golangci-lint run)
go build -o bin/server ./site/cmd/server
```

- [ ] **Step 3: Site module contracts**

```bash
just site-current-source-integration
just site-pinned-dependency-deployability
```

- [ ] **Step 4: Focused matrix and full E2E**

```bash
(cd site && go run ./cmd/e2econstraints --compile-matrix --list-matrix)
go test -tags=e2e,button ./site/tests/e2e -count=1 -timeout 5m
go test -tags=e2e,full ./site/tests/e2e -count=1 -timeout 15m
```

- [ ] **Step 5: Release dry run and coverage parity**

```bash
just coverage
scripts/run-release-coverage.sh --local-dry-run
```

Record the component package list and exact percentage in the verification
notes. Confirm the two invocations match.

- [ ] **Step 6: Review the complete diff**

```bash
git status --short
git diff --check origin/main...HEAD
git diff --stat origin/main...HEAD
```

Confirm unit tests are never path-conditional, all active E2E callers include
tags, focused runs use one invocation, release always uses `full`, no generated
file was hand-edited, and no temporary binaries are tracked.

- [ ] **Step 7: Commit final corrections**

```bash
git add -A
git commit -m "test: verify focused e2e selection pipeline"
```

Skip this commit when the final gates require no corrections.

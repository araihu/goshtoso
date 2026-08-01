# Focused E2E Selection and Demo Package Segmentation Design

**Status:** Approved direction; adversarial corrections incorporated; awaiting
re-review before implementation planning
**Date:** 2026-08-01
**Scope:** Demo-site package boundaries, E2E build constraints, dependency
impact analysis, CI selection, and release coverage

## Context

Goshtoso currently runs all root unit tests, site unit tests, and Playwright
E2E tests for each code change. Unit tests remain fast enough to run in full.
The E2E package is the expensive layer.

The E2E harness already has the correct process boundary: one `TestMain`
builds and starts the real demo server on a random port, launches one shared
Chromium instance, runs the selected tests, and shuts both down. Splitting E2E
tests into separate Go packages would repeat that setup and make the full suite
slower. The E2E package therefore remains unified.

The library already has useful package boundaries under `components/<name>`.
The demo site does not. Component pages share
`site/internal/pages/demo/components`, and example pages share
`site/internal/pages/demo/examples`. Go package analysis can report that those
large packages import Button, but cannot identify which component or example
page owns each use. Better leaf packages make the standard Go dependency graph
precise enough to drive focused E2E selection.

Boolean Go build constraints were smoke-tested independently. A focused
`button` build included Button and a composite Tooltip test, excluded an
unrelated release-only test, and retained the shared harness. This proved build
constraint mechanics only. It did not prove that the current Goshtoso E2E
package can compile under focused constraints.

## Goals

- Continue running every root and site unit test for ordinary changes.
- Run only impacted Playwright tests for classifiable pull-request and main
  changes.
- Include reverse component dependencies and real demo/example consumers.
- Preserve one E2E package, one server, and one shared browser per invocation.
- Give `go list` precise component-page and example-page package boundaries.
- Make focused test binaries independently compilable and inventory-checked.
- Distinguish authored changes from generator-owned outputs before choosing a
  test scope.
- Run the complete E2E and merged-coverage pipeline before publishing a
  release.
- Fall back to the full E2E suite for shared, unknown, or unsafe-to-classify
  changes.

## Non-goals

- Skip or focus root and site unit tests.
- Split `site/tests/e2e` into multiple Go packages.
- Infer impact from changed exported symbols. Package behavior can change
  without an exported declaration changing.
- Use `gopls` as a required CI dependency.
- Preserve the mutable exported `components.Demos` registry.
- Treat a synthetic build-tag fixture as acceptance evidence for the real E2E
  package.
- Guarantee a focused run for shared layout, global theme, runtime, or server
  infrastructure changes.

## Target Package Architecture

Component and example pages become leaf packages:

```text
site/internal/pages/demo/
  componentpages/
    button/
    actiongroup/
    table/
    ...
  examplepages/
    chat/
    todo/
    expense/
    ...
  registry/
  existing shared demo primitives
```

One public component demo lives in one `componentpages/<identity>` package.
One runnable example page lives in one `examplepages/<identity>` package.
Package identities match E2E identities, using library package names such as
`button`, `actiongroup`, `textinput`, and `schemaform`. Example tags are
prefixed, such as `example_chat`, to avoid collisions.

Documentation, legal, landing, and module pages are grouped separately after
component extraction. They do not need one package per page because they do
not improve component dependency precision.

The component library remains under `components/<name>`. The E2E suite remains
one package under `site/tests/e2e`.

### Dependency Direction

A neutral page definition lives in the existing shared `demo` package:

```go
type PageDefinition struct {
	Key         string
	Title       string
	Active      string
	Description string
	Type        string
	Content     func() templ.Component
}
```

Each leaf exports its definition. A neutral registry imports leaf packages
explicitly and builds the route map. Leaf packages never import the registry.
Shared rendering primitives remain below both. This prevents cycles.

Registration uses explicit imports, not `init` side effects. Registry
construction rejects duplicate keys, empty keys, empty content factories,
malformed paths, and incomplete component-catalog coverage. `Lookup`,
`MetaForKey`, and `AllPublicMeta` remain stable registry operations. Component
descriptions remain authoritative in the component catalog.

Component-specific HTTP handlers import the exact leaf package needed for a
fragment. The full-page router uses the registry. The registry and server are
aggregators in the Go graph, not E2E identities.

## Migration Strategy

Migration is incremental so every batch compiles and renders correctly.

1. Add `PageDefinition`, registry validation, and a temporary compatibility
   facade from the current `components.LookupDemo` API.
2. Move Button and ActionGroup as a pilot. Prove that `go list` reports the
   library and page reverse dependencies expected for a Button change.
3. Move remaining component pages in dependency-aware batches: primitives,
   navigation/display components, forms, then composites.
4. Move each example page and update its handlers to import the exact example
   page package. Existing domain packages under `site/internal/examples`
   remain unchanged.
5. Move remaining non-component pages into grouped page packages, cut callers
   over to the neutral registry, and remove the compatibility facade.
6. Extract every cross-file E2E declaration into dedicated support files while
   the suite is still untagged. The existing full E2E command must compile and
   pass after extraction.
7. Perform one atomic suite-gate cutover: add `e2e` constraints, add the
   list-only `TestMain` path, and migrate every active E2E caller to pass
   `-tags=e2e,full`. The repository may not contain a state where constraints
   have landed but active commands cannot find the package.
8. Replace runnable-file suite-only constraints with identity expressions and
   keep support files suite-only.
9. Run the per-identity compile and list inventories. Fix missing shared
   declarations and incomplete identity expressions until every identity and
   `full` pass.
10. Add dependency impact selection and focused CI wiring only after the real
    compile/list matrix is green.

Only `.templ` sources and ordinary Go sources are moved manually. Generated
`*_templ.go` files are removed at their old location and recreated with
`templ generate`; generated code is never hand-edited.

Every batch preserves route keys, metadata, public HTML behavior, handler
responses, HTMX fragment contracts, and Alpine initialization.

## E2E Build Constraints

Shared harness and helper files use only the suite gate:

```go
//go:build e2e
```

Identity test files include their focused identity and `full`:

```go
//go:build e2e && (full || button)
```

Cross-component tests declare every identity they intentionally cover:

```go
//go:build e2e && (full || button || actiongroup)
```

Example tests use prefixed identities:

```go
//go:build e2e && (full || example_chat)
```

Global all-page, full theme-matrix, and release-only smoke files use:

```go
//go:build e2e && full
```

Every Go file under `site/tests/e2e` has a build expression that implies
`e2e`. Support files use only `e2e` and contain no runnable tests. Runnable
identity files include `full` plus one or more focused identities. Validation
rejects untagged files and expressions that can become true without `e2e`.

### Active Caller Migration

Adding the suite gate changes the package's public command contract. Every
active caller is migrated in the same atomic cutover:

- `Makefile` targets `test-e2e` and `test-e2e-one`;
- the Justfile coverage target and its extracted reusable coverage script;
- `site/tests/e2e/Dockerfile`;
- Code CI and release workflows;
- `AGENTS.md`, `README.md`, `CONTRIBUTING.md`, the pull-request template,
  `docs/RELEASE_CHECKLIST.md`, and `site/tests/e2e/README.md`.

Full-suite and single-test documentation uses `-tags=e2e,full`; focused
automation uses `-tags=e2e,<calculated identities>`. A repository validation
scan rejects executable or current documented `go test` commands that target
`site/tests/e2e` or `tests/e2e` without an `e2e` tag. Historical records under
`docs/audits` and `docs/plans` remain immutable and are explicitly excluded
from this active-caller scan.

### Shared Helper Extraction

Focused constraints compile files independently. A helper declared in one
identity file is unavailable when another identity selects only its own file.
The current suite already violates this requirement. Known examples are:

- `newIsolatedPage`, declared in `todo_example_test.go` and used across twenty
  files;
- `waitForAlpine`, declared in `modal_test.go` and used across twenty-nine
  files;
- `mustAttribute`, declared in `card_test.go` and used across fourteen files.

Every cross-file helper, type, variable, and constant first moves into a
dedicated, temporarily untagged support file containing no runnable `Test...`
function. The atomic suite-gate cutover then adds `//go:build e2e` to those
support files and the `TestMain` file before identity constraints are applied.

Validation performs a type-aware cross-file declaration/use inventory. A
declaration used from another identity file is rejected unless its declaration
lives in an `e2e`-only support file. The three known helpers are mandatory test
fixtures for this validator; the implementation may not stop after moving only
those three.

## Dependency Impact Analysis

`cmd/e2eimpact` accepts Git base and head revisions:

```bash
go run ./cmd/e2eimpact --base <sha> --head <sha>
```

It emits machine-readable mode and tags plus human-readable reasons:

```text
mode=focused
tags=e2e,button,actiongroup,example_chat
```

```text
button: changed directly
actiongroup: imports button
example_chat: imports button
```

The command:

1. reads changed paths from the Git diff;
2. classifies authored and generator-owned files;
3. maps authored component/page/example paths to graph roots;
4. loads direct imports with `go list -json`;
5. computes the reverse import closure;
6. emits tags only for component-page and example-page identity packages;
7. ignores registry and server aggregators as identities;
8. emits `e2e,full` when classification or graph construction is unsafe.

Production imports drive impact. Test-only imports do not expand the graph
because all unit tests already run. Cross-component E2E contracts remain
explicit in their build expressions.

### Git Range and History Semantics

Changed paths come from NUL-delimited
`git diff --name-status -z -M <start> <head>`. Status and paths are parsed
without line-oriented shell splitting.

For pull requests, CI fetches both the event's base SHA and head SHA with enough
history to compute their merge base. The start revision is
`git merge-base <base-sha> <head-sha>`, so a branch behind `main` does not treat
unrelated base-branch commits as the pull request's work. Missing commits or an
unavailable merge base select `full` with an explicit history reason.

For a main push, the start revision is `github.event.before` and the head is
`github.sha`; this covers every commit delivered by a multi-commit push. An
all-zero `before`, missing revision, or non-ancestor/force-push relationship
selects `full`. Release tag runs do not calculate impact because releases are
always full.

The head checkout supplies the package graph for added and modified files.
Rename records classify both old and new paths for reporting, but any rename or
deletion selects `full`; this avoids silently losing a removed package or old
reverse edge from a head-only graph. Add records classify the new path.
This policy deliberately chooses a full fallback instead of maintaining a
second base-side Go workspace.

Impact tests cover modified and added files, deleted files, renamed files,
branch-behind pull requests, multi-commit pushes, an all-zero first push,
force-push ancestry failure, and missing/shallow history.

`gopls references` remains an optional diagnostic for investigating exact
symbol uses. It is not part of CI correctness.

## Authored and Derived Change Classification

Impact is determined from authored source first. Generated outputs must not
turn an otherwise focused component change into a full-suite run.

The impact command owns a tested list of generator-produced paths, including
templ output, compiled Tailwind/theme output, first-party JavaScript bundles,
and other checked generator artifacts. Classification follows these rules:

1. Classify every authored path.
2. Compute identities from authored paths and package dependencies.
3. Ignore a recognized derived output for scope selection only when the same
   diff contains classifiable authored input for its generator family.
4. Treat a generated-only diff as unexplained and run `e2e,full`.
5. Treat unknown generated paths as full-suite impact.

Example: changing `components/button/button.templ` may regenerate
`components/button/button_templ.go` and `assets/styles.css`. The authored
Button source selects Button and its consumers; the known derived files do not
force `full`.

Direct authored changes to global CSS, theme definitions or generators,
vendored runtime versions/assets, shared JavaScript, the asset handler, or
Tailwind configuration select `full`. Component-scoped JavaScript sources may
select their matching component identity. Drift checks still verify that every
derived artifact matches its authored input; ignoring a derived path for scope
does not ignore drift.

## Server Handler Classification

`site/internal/server` remains one Go package, so package-level analysis alone
is too coarse. Existing and newly extracted handler files follow an identity
naming convention and a checked mapping:

- `table_handler.go` selects `table`;
- `search_handler.go` selects `search`;
- `toast_handler.go` selects `toast`;
- example handler files select their `example_<name>` identity;
- mixed component handlers currently in `server.go` move to identity-specific
  files during page segmentation.

The impact command validates a checked-in Go mapping from each non-shared
handler file to one or more known component/example identities. For example,
form-validation behavior may select both `form` and `textinput`. A classifiable
handler change selects every mapped identity and reverse consumer. Validation
fails when a new handler file is neither mapped nor explicitly marked shared.

`server.go`, shared middleware, storage consent, server lifecycle, route
aggregation, and unclassified handler files select `full`. Adding a new route
through the central aggregator therefore runs the full suite once. This is an
accepted safety fallback, not a silent omission.

## Real Focused-Build Verification

Build-expression parsing alone cannot prove the package compiles. The real E2E
package must pass a per-identity compile matrix.

For every identity reported by the registry and impact tool:

```bash
cd site
go test -c \
  -tags=e2e,<identity> \
  -o "$temporary_directory/<identity>.test" \
  ./tests/e2e
```

The matrix also compiles `e2e,full`. Any undefined helper, excluded support
type, duplicate declaration, invalid import, or empty focused package fails CI.
Compiled binaries live in a temporary directory and are not committed.

Inventory validation uses the Go tool, not only AST prediction:

```bash
GOSHTOSO_E2E_LIST_ONLY=1 \
  go test -tags=e2e,<identity> -list '^Test' ./tests/e2e
```

`TestMain` recognizes `GOSHTOSO_E2E_LIST_ONLY=1` and calls `m.Run` without
building the server or launching Playwright. The validator parses each test
file's build expression with `go/build/constraint`, predicts the tests for each
identity, and compares that exact set with `go test -list`. Every identity must
list at least one test. The `full` inventory must equal the union of all focused
and full-only tests. Unexpected or missing tests fail CI.

This compile/list matrix is required locally for the pilot and in CI whenever
E2E test files, constraints, shared helpers, the impact tool, or page identities
change.

## Pull-Request and Main CI

Ordinary code CI performs these steps:

1. check out the change and configure Go/Playwright caches;
2. install pinned templ, Tailwind, Playwright, and Chromium tooling;
3. regenerate templ, CSS/theme, and JavaScript artifacts and run drift checks;
4. create the root/site workspace;
5. run all root unit tests with component coverage;
6. run all site non-E2E tests with component coverage;
7. calculate E2E impact from committed authored changes;
8. run one Playwright invocation with all selected tags;
9. merge unit coverage with the selected E2E coverage;
10. label the result as partial and publish selected identities and reasons in
    the job summary.

Partial PR/main coverage never updates the global coverage badge. Failures
upload current E2E artifacts. A shared, unknown, or analysis-failure result runs
`e2e,full`.

## Complete Pre-release Coverage Gate

A tag push must not create a GitHub release until a complete pre-release job
passes. Adding only `go test -tags=e2e,full` is insufficient.

The pre-release job runs before the publishing job and includes:

1. checkout of the exact tag;
2. pinned Go and Node setup;
3. pinned templ and Tailwind installation;
4. pinned Playwright CLI installation and `playwright install --with-deps
   chromium`;
5. `go work init . ./site` for current-source integration;
6. templ, theme, CSS, and JavaScript regeneration plus zero-drift checks;
7. both site-module contracts;
8. root unit coverage over all root packages;
9. site unit coverage over all non-E2E site packages;
10. a coverage-instrumented demo server driven by
    `go test -tags=e2e,full`, using `GOSHTOSO_E2E_COVERDIR` and
    `GOSHTOSO_E2E_COVERPKG`;
11. `go tool covdata merge`, percent, text, function, and HTML output;
12. upload of merged coverage and E2E failure artifacts;
13. coverage percentage/color outputs for badge publication.

### Authoritative Coverage Denominator

The published badge denominator is exactly the publishable component package
set returned by:

```bash
component_coverpkg="$(go list ./components/... | sort | paste -sd, -)"
```

Root tests, site tests, and the E2E server instrument component packages as
needed. `site/cmd/server` is instrumentation-only so the E2E shutdown path can
flush coverage; it is never part of the published denominator.

After merging native coverage data, both conversion commands use the same
component filter:

```bash
go tool covdata percent \
  -i=.coverage/merged \
  -pkg="$component_coverpkg"
go tool covdata textfmt \
  -i=.coverage/merged \
  -pkg="$component_coverpkg" \
  -o=.coverage/coverage.out
```

The badge percentage is the `total:` value from `go tool cover
-func=.coverage/coverage.out`. PR focused runs use the same package denominator
but are labelled partial because selected E2E execution may not exercise every
component path. Only release-full output updates the badge.

The existing coverage shell logic moves into one reusable repository script so
PR partial coverage, local full coverage, and release full coverage cannot
silently diverge. The script receives the E2E tag set explicitly.

The reusable script records its exact component package list and percentage.
On the same commit, local `e2e,full` execution and release manual dry-run must
produce the same package list and numeric percentage; mismatch fails acceptance.

The release-publishing job has `needs: pre-release`. Only after the full gate
passes may it create the GitHub release, update the coverage badge from the
full merged result, and update the release badge. The coverage badge therefore
means coverage from the most recent successful release, not a partial main
run.

`release.yml` also exposes an explicit manual dry-run mode. Dry-run executes
the complete pre-release job but skips GitHub release creation and both badge
updates. This provides workflow-level acceptance evidence without publishing a
tag or mutating release state.

## Failure and Fallback Policy

Focused selection must fail safe.

- Unknown authored path: `full`.
- Shared layout/runtime/theme/server infrastructure: `full`.
- Impact command error or incomplete graph: `full`, with reason.
- Recognized generated output without authored cause: `full`.
- Missing identity-to-test mapping: CI failure, not a skipped test.
- Focused compile or inventory mismatch: CI failure.
- Release full coverage failure: no GitHub release or badge update.

The impact summary always states why the run is focused or full. Silent empty
selection is forbidden.

## Acceptance Gates

Package segmentation is accepted only when:

- registry route keys, titles, active states, descriptions, and catalog
  coverage match the pre-migration behavior;
- templ regeneration has zero uncommitted drift;
- root and site unit suites pass;
- current-source and pinned-dependency site contracts pass;
- server build and lint pass;
- direct-load and HTMX fragment behavior pass for each moved batch;
- every E2E identity compiles independently;
- every identity's `go test -list` inventory matches its build constraints;
- `full` lists and runs the complete suite;
- every executable/current documented E2E caller passes `-tags=e2e,...`, while
  historical audit and plan records remain unchanged;
- authored Button plus derived templ/CSS output remains focused;
- direct global CSS/runtime/theme and generated-only fixtures select `full`;
- impact fixtures prove merge-base pull-request selection, exact main-push
  before/after selection, add/modify handling, and full fallback for
  delete/rename/force-push/all-zero/missing-history cases;
- classifiable handler fixtures select their identities and shared server
  fixtures select `full`;
- local and release-full coverage use the exact `components/...` denominator,
  exclude `site/cmd/server`, and report the same percentage on the same commit;
- the complete pre-release coverage job succeeds through the manual dry-run
  path without publishing or updating badges;
- the latest adversarial review finds no implementation blocker.

The broad package migration itself receives one final full E2E run because it
changes routing and page ownership across the site. Focused selection becomes
the default only after that migration and all compile/inventory gates pass.

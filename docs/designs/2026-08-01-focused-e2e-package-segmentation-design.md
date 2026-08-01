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
6. Add E2E build constraints, impact analysis, and CI selection only after the
   real E2E package passes the focused compile and inventory gates below.

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

### Shared Helper Extraction

Focused constraints compile files independently. A helper declared in one
identity file is unavailable when another identity selects only its own file.
The current suite already violates this requirement. Known examples are:

- `newIsolatedPage`, declared in `todo_example_test.go` and used across twenty
  files;
- `waitForAlpine`, declared in `modal_test.go` and used across twenty-nine
  files;
- `mustAttribute`, declared in `card_test.go` and used across fourteen files.

Before adding identity constraints, every cross-file helper, type, variable,
and constant moves into a support file that has `//go:build e2e` and contains
no runnable `Test...` function. `TestMain` also remains in an `e2e`-only shared
file.

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

The existing coverage shell logic moves into one reusable repository script so
PR partial coverage, local full coverage, and release full coverage cannot
silently diverge. The script receives the E2E tag set explicitly.

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
- authored Button plus derived templ/CSS output remains focused;
- direct global CSS/runtime/theme and generated-only fixtures select `full`;
- classifiable handler fixtures select their identities and shared server
  fixtures select `full`;
- the complete pre-release coverage job succeeds through the manual dry-run
  path without publishing or updating badges;
- a second adversarial review finds no implementation blocker.

The broad package migration itself receives one final full E2E run because it
changes routing and page ownership across the site. Focused selection becomes
the default only after that migration and all compile/inventory gates pass.

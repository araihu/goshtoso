# Pair A1 modernization results

Baseline: `0ff8117c`; runtimes: htmx 4.0.0, Alpine 3.17.2 and official hx-alpine-compat. A03 requires paired A07 synchronization before integration. Shared minified bundles are regenerated locally for testing and finalized by the coordinator.

| Task | Outcome | Evidence |
|---|---|---|
| A01 | Removed ineffective watcher disposer storage/destroy. Retained explicit bridge after native and normalized modelable trials failed initialization parity. | `TestSelectBindingDefaultsAndRepeatedMounts`, existing dependent binding/external draft/keyboard tests; prototype matrix below. |
| A02 | Client factory restores storage during Alpine init. Alpine owns persisted-pageshow listener. Removed document-ready/global pageshow scans. | Repeated real fragment entry, two roots, malformed/unavailable storage, opt-out, ordinary versus persisted pageshow, existing BFCache and no-HTTP tests. |
| A03 | Server toggle/clear return one full Combobox with outerMorph. Search remains options-only replacement. External OOB helpers retain their separate public contract. | Repeated rendered morph covers hidden values, label/ARIA, removed option, open state, dirty search/caret. Handler and lazy/cascade/error tests. A2 covers races. |
| A04 | Retained replacement for changed Select/Search/StructuredInput configuration. Added ownership comments and rendered replacement regression. | Actual rendered morph below shows stale options/cache/indexed defaults; replacement resets all three. |
| A05 | Static JSON hx-vals carries kind/field identity; removed wrapper name. | Missing/empty and quotes/newline/HTML-special names round-trip through rendered HTML/JSON. Native change and dependent slug tests. |
| A06 | Dependent response renders copied FieldGroup with OOB set; reusable state is never mutated. | Two dependents, initial true/false OOB, writer failure, cancellation, observation during render and reuse. Dependent/value browser tests. |

## Modelable experiment

A browser fixture used the actual Select factory/vendored Alpine, options a/b and server default a. The native prototype removed paired watches and exposed selectedValue via selectedOption getter/syncFromInput setter, with x-modelable and parent x-model. A second trial normalized falsey initial parent to the default before entanglement. No browser errors occurred.

| Initial parent | Baseline parent / selection | Native parent / selection | Normalized parent / selection |
|---|---|---|---|
| empty | empty / a | empty / empty | a / a |
| undefined | undefined / a | undefined / empty | a / a |
| missing | empty / empty | missing / empty | missing / empty |
| b | b / b | b / b | b / b |

Normalization changes dependent enablement by writing a into an initially empty parent. Native entanglement also does not publish invalid-value normalization during initialization. Preserving both requires special initialization gates/additional write-back ownership alongside the getter/setter bridge. Keeping the small explicit bridge is clearer. The regression covers dotted parent expressions, the matrix, repeated mounts, removal and parent mutation after removal.

Tagged [modelable](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/directives/x-modelable.js) owns entanglement cleanup; [entangle](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/entangle.js) initially assigns the outer value inward. This predates 3.17.2.

## Morph ownership experiments

Real Go-rendered server Combobox fixtures shared one ID. Before morph the menu was open with a selected, focused dirty search and caret2. New state removed a, selected/relabeled b and supplied another search default. Native outerMorph produced:

```text
same root: true; open: true; focusIndex: -1
same focused search: true; value: typed; caret: 2
label: Changed Beta; hidden: [b]; option ARIA: [[b,true]]
```

Committed regression repeats selected/cleared/reselected states. The focused search draft is user-owned while mounted; an authoritative reset must replace the root. Options, selection, labels and option ARIA remain server-owned. No client-mode morph conversion.

Another real rendered fixture dirtied StructuredInput rows and seeded Search cache, then innerMorphed the host with new Select options, columns/defaults and Search URL:

```text
Select live label: Alpha; new dataset label: Beta
Structured rows: [[dirty]]; defaults: [initial]; new column: other
Search items: [{title: old cache}]; loaded: true; new URL: /new-source
browser errors: []
```

This rejects a proposed boundary; the shipped replacement path was not broken. Indexed arrays cannot safely become another schema, and old cached items cannot become another URL's results. The replacement regression covers new/removed options, changed columns/defaults/rows and reset Search query/cache. No reconciliation observer or forced initialization was added. The [compat extension](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-alpine-compat.js) preserves state but does not own application reconciliation.

## Verification and paired review

- `templ generate` and `go run ./cmd/jsbuild`: generated local test artifacts.
- `go run ./cmd/jslint`: passed (existing similarity/baseline output only).
- `go test ./components/select ./components/combobox ./components/form/... -count=1`: passed.
- `/tmp/gs-golangci/golangci-lint run`: zero issues.
- `go test -tags=e2e,full ./site/tests/e2e/... -count=1 -timeout 8m -run 'Test(Select|Combobox|InputReplacement|FormValidation)'`: passed in47.370s.

The first browser run caught a helper manually reading removed wrapper name; native change already passed. The helper now consumes static hx-vals, and the complete focused rerun passed. Unit markup expectations were updated with an additional full-response regression.

Pair A2 reciprocal review/combined race verification pending at this implementation checkpoint. No public API fields or dependency pins changed here. Coordinator owns combined full suite, public pins/module contracts and shared generated outputs.

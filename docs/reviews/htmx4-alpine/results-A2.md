# Pair A: request modernization results

Driver A2 implemented A07–A12 and integrated A1's A01–A06 for reciprocal regression testing. The implementation targets the pinned htmx 4.0.0 and Alpine 3.17.2; it introduces no htmx 2 compatibility. Canonical completion tracking remains in [TASKS.md](TASKS.md).

## Decisions by task

| Task | Result | Evidence and retained boundaries |
|---|---|---|
| A07 | Implemented | Independent field validation uses latest-result cancellation. Form submission aborts outstanding validation and prevents validation from starting until the actual submit context finishes, even with a separate loading indicator. Combobox search has its own read owner; mutations use a stable widget `drop` queue, cancel pending reads and temporarily make search read-only to preserve its focused draft/caret. Mutation buttons exclude pre-disabled controls from native disabling. |
| A08 | Implemented | Replaced custom sentinel WeakMap, observer creation, imperative ajax, scans and dependency-ready restart with native intersection, explicit root/margin, row replacement and a manual retry button. 4xx/5xx preserve the sentinel. Native htmx owns observer teardown and late core processing. The grouped `closest :is(.overflow-y-auto,[style*=overflow-y])` selector prevents a comma from becoming an unrelated global selector. Nearest matching contained scroller wins; otherwise the viewport is used. Browsers must support IntersectionObserver, consistent with the tested browser baseline. |
| A09 | Implemented with small retained normalizer | Filter controls are named and included by htmx. A custom native request trigger replaces imperative ajax; a component-scoped event replaces the per-table document listener. Server Go builds fixed query parameters. The scoped normalizer remains necessary because omitting an empty successful control does not delete an old filter from a server-generated sort/page URL. It clears those stale keys, preserves the empty/false omission contract and carries current server sort. Extra `per_page` retains its override precedence. Toggle defaults now normalize `true`/`false` to actual booleans; string `false` cannot accidentally become a checked native input. |
| A10 | Implemented with upstream-defect guard | Filter, sort and pagination share a stable per-table queue owner in a `display: contents` wrapper. Explicit request policies do not attach replacement cancellation to nested row mutations. The pinned native queue needs the context guard described below to guarantee third-request freshness. Two table instances remain independent. |
| A11 | Implemented | Form exposes `RootAttrs`, `HTMX.Sync` and `HTMX.Disable`. The real validation demo drops duplicate submits, disables only eligible submit buttons and preserves the form on unexpected 5xx. Expected 422 response HTML still renders. Combobox provider 502 remains an intentional retargeted retry fragment. |
| A12 | Implemented/reviewed | Reviewed A1's watcher cleanup, mount restoration, full-root server morph, static field identity, copied dependent-response routing and explicit replacement ownership. Preserved nested row-control guards and existing pagination/OOB hosts. Combined request tests cover actual full-root mutation responses, not just direct morph primitives. |

## New retained workaround: htmx 4 request-queue cancellation

The [tagged htmx 4.0.0 source](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/htmx.js) has two relevant behaviors:

- `RequestQueue.admit('replace', ...)` aborts A and stores B as current, but A's eventual `#issueRequest` finally unconditionally calls `RequestQueue.continue()`. That clears B's slot. A third request can consequently overlap B, and `htmx:abort` no longer finds B.
- `#cleanup` disconnects observers and removes listeners/timers, but does not abort in-flight fetch requests.

These were reproduced with the exact vendored runtime: first search aborted, second search started, then a mutation began; native abort left the second search running in both Form and Combobox fixtures. A third table request also exercises the lost-slot condition. This is a discovered upstream limitation, not an assumed need for legacy compatibility.

`assets/js/src/components/requests.js` retains the real context abort handle per owner in a WeakMap. Before a new read it aborts the previous context. Cleanup and finally handlers compare context identity before clearing it, so an older completion cannot clear a newer handle. Cleanup listens on the actual source and ignores bubbling descendant cleanup. Native admission remains responsible for mutation `drop` policies. Read-only input ownership restores the prior property after its own mutation finishes. Remove this helper only after the pinned runtime changes and the delayed three-request, mutation and removal regressions pass without it.

## Additional contract correction

`closest [data-combobox] input[type=hidden]` was not a chained htmx selector: v4 passed the whole CSS expression to `closest`, which could not match the search input or option item. `closest [data-combobox]` now includes that widget's successful named controls, including selected hidden values and query; explicit dependency selectors remain appended. Actual two-toggle and clear requests assert the prior selections and sibling dependency, and exclude an unrelated external control.

## Reciprocal review

A1's review identified four concrete defects in the initial A2 patch, all resolved:

1. Loading CSS is not submission ownership when a consumer provides a separate indicator. Submission blocking now uses actual tracked request context; the test uses an external indicator.
2. Native filter URL construction initially dropped an extra-query `per_page`. Corrected precedence and exact payload assertion preserve it.
3. A source cleanup listener initially reacted to bubbling child cleanup. It now checks the event target; tests prove descendant cleanup preserves the request and source removal aborts it.
4. Disabling a focused search input would blur it before full-root morph. Search is now temporarily read-only; an actual mutation response preserves focus, value and caret and restores editability.

A1 also challenged the obsolete cleanup-aborts-search assumption and the new shared helper's context identity handling. A2 reviewed A1's committed implementation and found no remaining source issues beyond the request/focus interactions closed above. Final integration review is coordinated by the root agent.

## Verification

`site/tests/e2e/request_modernization_test.go` contains rendered-component regressions with a controllable fetch transport. Requests are genuinely held pending and resolved in opposing orders; this is not an attribute-only test:

- Three table intents, out-of-order stale responses, separate widget independence, named payloads, numeric zero, unusual keys, false defaults, stale-key clearing and nested consumer form isolation.
- Child versus source cleanup cancellation; unrelated scroller protection; contained and viewport sentinel activation, 403/500 retry, page chains and exactly one successful request per page.
- Per-field replacement cancellation, submit during pending validation, blocked validation during submit, external indicator, synthetic duplicate submit, pre-disabled buttons, expected 422 and unexpected 500.
- Actual server Combobox full-root mutation responses with pending searches, focus/caret preservation, two successive selection payloads, clear payload and intentional provider 502 retry.

The full existing docs-page infinite-scroll append test caught a selector regression missed by the initial isolated fixture. The final grouped selector passes that actual demo test. Existing input, form and table suites provide the keyboard, row-control, history and rendering coverage around these targeted request tests. Focused package tests and root lint were rerun after implementation; combined browser results are recorded in the final execution update below. Shared generated bundles, CSS, skill references and public dependency pins are finalized by the coordinator.

### Final execution update

- Combined browser run `TestModern|TestSelect|TestCombobox|TestFormValidation|TestTable`: passed, **81.012s**. Log: `/tmp/modern-a2-combined-final.log`.
- After extracting the three long event hooks into named runtime methods, `TestModern|TestTableCoverageDemo`: passed, **8.187s**. Log: `/tmp/modern-a2-extracted-final.log`.
- Table, Form, validation, Combobox, component-runtime and JavaScript-tooling package tests: all passed. Log: `/tmp/modern-a2-unit-final.log`.
- Root Go lint: **0 issues**. Authored JavaScript lint: **no new inline findings** (19 existing baselined candidates); no extraction-policy waiver was added. Logs: `/tmp/modern-a2-lint-final.log`, `/tmp/modern-a2-jslint-final.log`.

The first broad run's real-demo sentinel failure was fixed by the grouped selector and then verified in the successful combined run. No test was weakened to accept the unrelated-scroller behavior. The first native-sync probe's cancellation failures were fixed with the context guard and retained as regression coverage. Full repository and public-pinned integration gates remain the coordinator's final responsibility.

### Integrated landing dependency correction

The coordinator's full suite found that `/playground/theme` still declared only htmx and Alpine manually, without the component bundle. Its lazy table therefore invoked an undefined request helper. The homepage separately omitted Alpine compatibility from its manual chain. Both standalone documents now use the same manifest-backed `head.DependenciesMinimal(head.WithLocalRuntime())` contract as library consumers. Their rendered controls do not require focus/collapse/mask plugins.

`TestLandingPlaygroundWaitsForComponentRuntime` holds the actual page's component bundle download, lets the document finish parsing, and asserts that no table request runs early. Releasing the bundle then allows guarded lazy loading and sorting without console/page errors. This fixes the missing dependency rather than silently skipping synchronization. It changes only site code; it needs no new library runtime version.

Validation: `TestLanding|TestModern|TestHead` passed **32.839s**, and head/start-page units passed. Logs: `/tmp/modern-a2-landingfix.log`, `/tmp/modern-a2-landingunit.log`.

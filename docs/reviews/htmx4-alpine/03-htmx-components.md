# Pair B1 — htmx component contracts

Report only; reviewed 2026-09-10 at Goshtoso `707126f3cb4d1efda821c047905d64768abefff3`. No implementation changes or whole-suite rerun. Prior migration test results are context, not evidence that the proposals below work.

The best removal candidates are the duplicate lazy-tab request path and table sentinel observer machinery. Table request serialization and payload ownership deserve a separate, behavior-driven change. Do not mechanically change replacement swaps to morphs: several current responses intentionally replace server-owned state.

## Evidence and coverage

Read repository instructions, htmx and Alpine skills and their migration/integration/pattern/gotcha references. Traced authored templates, Go config/attribute helpers, component JavaScript, the relevant demo server responses and existing browser assertions. Generated templ files were used only as supporting rendered-markup evidence. Scope is htmx request/response contracts; Pair A owns local widget state and accessibility, and B2 owns runtime/streaming/app-shell lifecycle.

| Surface | Coverage and disposition |
| --- | --- |
| Form / validation | Form root verbs, field wrapper triggers, explicit field identity, include behavior, dependency OOB response, full submit response; B1-04/05/06. |
| Table / pagination | Sort URLs and head response, filter values, pagination host including empty pages, lazy tbody, custom sentinel and both scrolling layouts, row links and nested-control guards; B1-02/03/06/07. |
| Combobox | Server options/toggle/clear and provider error routing, body/label OOB, includes/dependencies, search debounce, client/server mode distinction; B1-03/05/06. |
| Tabs / carousel | Dual lazy-tab request path versus single native carousel loader; B1-01; retain carousel loader. Accordion has local expand/collapse, no built-in remote loader to prune. |
| Button / radio / actiongroup / splitbutton | Verbs, loading/disable behavior, include/values, action adapters and passthroughs; B1-05/07. |
| Alert / banner / dropdown / modal / toast | Action config emission, dismiss/action local state, toast append-OOB; B1-05/07, otherwise retain. |
| Sidebar / navbar / link / breadcrumbs | Passthrough/sanitization and native navigation; navbar secondary-link allowlist is deliberately narrower than root hooks; do not broadly relax it just to expose new htmx attributes. |
| Drawer / panel / appshell | Stable content slots are swap destinations, not duplicate request engines; retain. |
| Textinput / textarea / fileinput / range / chatbubble | Attribute escape hatches accept htmx without an owned network algorithm; no dedicated rewrite needed. |
| Search / palette | Search's optional JSON corpus fetch, normalization, cache and abort are a different contract from HTML swaps; palette is local color selection. Keep those mechanisms; do not substitute htmx HTML requests for JSON merely to reduce `fetch` occurrences. |
| Head / scrollregion / codeblock | Identified runtime ordering and process/cleanup hooks, handed lifecycle to B2; no additional request-contract proposal. |
| Remaining directory inventory | Avatar, badge, card, checkbox, emptystate, icon, inlinecode, kbd, pageheader, popover, rating, schemaform, schematree, skeleton, spinner, select, steps, structuredinput, tagslist, toggle, toolbar, tooltip: no independent htmx request engine found to replace; nested/passthrough interactions belong to the contracts above and Pair A. |

All 55 component directory identities were included in this classification. This is not a claim of line-by-line review of every styling helper or every test.

Primary implementation reference: [tagged htmx 4 source](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/htmx.js). Specific mechanisms below were checked against that source, downloaded to a temporary file, rather than inferred from legacy htmx documentation. The documentation attribute URLs attempted during review did not resolve through the browser tool; source is authoritative for those checks. Alpine preservation is verified against the [tagged compatibility extension](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-alpine-compat.js) and [official morph guide](https://four.htmx.org/docs/morphing-swaps-guide).

## B1-01 — Remove the second lazy-tab request engine

**Replace/remove; P1; high confidence in duplicate mechanism, medium confidence in final replacement semantics. Canonical cross-pair finding: A2-01.**

[Panel template, lines 112–128](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/tabs/tabs.templ#L112) both sets `dataset.loaded` and calls `htmx.ajax` from `x-effect`, and independently declares `hx-get` with `intersect once`. The marker guards only the effect. The observer never reads it.

Temporary Playwright reproduction copied these attributes, included `[x-cloak]{display:none!important}`, and loaded the exact vendored htmx/compat/Alpine files. An initially hidden panel produced **0** requests before selection and **2** after first selection. The endpoint was intercepted and returned a small HTML fragment; this proves the duplicated mechanism in an equivalent fixture, not an exact whole-demo request trace. A2 independently reproduced the same result.

Reproduction recipe: serve an empty same-origin document, intercept `/data` and increment a counter while returning `<p>loaded</p>`, insert the following fixture, then load the three vendored scripts in order: htmx core, hx-alpine-compat, Alpine core. Wait 200ms, read counter, click Next, wait 500ms, read counter. Observed `0 -> 2`.

```html
<style>[x-cloak]{display:none!important}</style>
<section x-data="{selectedTab:'first'}">
  <button @click="selectedTab='second'">Next</button>
  <div x-cloak id="panel" x-show="selectedTab==='second'"
    x-effect="if(selectedTab==='second'&&!$el.dataset.loaded&&window.htmx){$el.dataset.loaded='true';htmx.ajax('GET','/data',{target:'#panel',swap:'innerHTML'})}"
    hx-get="/data" hx-trigger="intersect once">loading</div>
</section>
```

Keep one declarative request path. Prefer a named selection event dispatched by Alpine and handled by the panel if preserving today's eager load on selection even when offscreen is required. `intersect once` alone is shorter but changes that contract to viewport loading. Remove manual URL interpolation, imperative ajax, and its independent loaded flag only after choosing that behavior. Native `once` consumes an attempt, not successful response completion: decide whether failure exposes explicit retry or rearms the loader.

Existing tests: `TestTabs_HTMXLazyPanelLoadsOnlyAfterSelection`, `TestTabs_HTMXFragmentLifecycle`, spinner reduced-motion coverage. The first checks resulting text, not request count. Regression plan: count exactly one request per panel across selection, reselection and fragment reinsertion; test offscreen selection, indicator visibility and a failed first response followed by retry. Source: htmx `#onTrigger` observer and once handling. Do not count this independently from A2-01 in the action backlog.

## B1-02 — Replace table sentinel JavaScript with native intersection triggers

**Replace; P2; high confidence in native capability, medium in behavior parity.**

[Sentinel runtime, line 132](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/table.js#L132) owns a WeakMap, IntersectionObserver creation, root discovery, request helper, subtree scans, cleanup hooks and dependency-ready restart. [Sentinel markup, line 662](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/table/table.templ#L662) already exposes a request URL in `data-hx-get` but relies on that runtime for viewport activation and outer replacement.

Native v4 `intersect` supports `root`, `rootMargin` and `threshold`, with htmx-owned observer lifecycle. Emit a normal native trigger, explicit target/swap (`this`, `outerHTML settle:200ms`) and the same 400px prefetch margin; render or otherwise supply the appropriate root selector for contained versus page scroll. This can remove the observer registry and its process/cleanup/restart plumbing, while retaining row-link behavior in the same file.

Do not substitute bare `revealed`: the existing algorithm uses a selected scrolling root and prefetch margin. Current fallback without IntersectionObserver immediately requests; native source constructs IntersectionObserver directly, so browser-support policy must explicitly allow dropping that fallback. Current code disconnects on request admission rather than success; native once has a similar attempt-oriented concern, not automatic robust retries. A failed sentinel needs an intentional retry contract.

Existing tests: `TestTable_InfiniteScroll`, `TestTableHTMX_BrowserInfiniteScroll`, table pagination-nav coverage and `components/table/table_coverage_test.go` sentinel rendering checks. Add contained scroller, page scroller, sentinel initially visible, append chain, network failure, rapid fragment removal, late dependency initialization and exact one request per page assertions. Preserve valid `<tr>` response context. Primary source: tagged htmx `#onTrigger`, `#cleanupTrigger` and `#insertContent`; verify generated root syntax in a browser before deleting code.

## B1-03 — Give shared-result requests an explicit synchronization contract

**Replace/investigate; P1; high confidence in distinct queues, medium in user-visible race severity.**

[Table filter request, line 69](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/table.js#L69), [sortable header, line 623](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/table/table.templ#L623), pagination links, and [combobox search/options](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/combobox/combobox.templ#L91) can issue different requests that update overlapping UI. No owned `hx-sync` policy is emitted. In v4 the default is `queue first` **per source**, not per destination. Imperative ajax with a target and no source chooses the target as source; thus filters normally queue on tbody, while headers and pagination have different queues. This is not body-global unless no target/source is supplied.

Use a stable per-widget synchronization owner with explicitly inherited `hx-sync` where descendants should share it. Candidate read-only search/filter policy is replace or queue-last, based on intended response freshness and server cancellation behavior. Give imperative calls an explicit source if they remain. This makes independent widgets independent while coordinating requests that overwrite the same rows/options. Debounce alone does not solve overlapping in-flight requests.

Do **not** use replace for mutations by default. Combobox multi-toggle derives the next selection from submitted hidden values; canceling a browser request does not undo a server mutation, and queued requests may already contain stale snapshots. Probe rapid two-option toggles and clear/search overlap before choosing queue, disabled controls, or a revised idempotent selection payload. Native sync is a building block, not a complete selection protocol.

Existing tests: `TestTableFilter_SortPersistsAcrossFilterAndPagination`, `TestTableHTMX_FilterSortPaginate`, `TestComboboxCoverage_ServerLazySearchAndToggle`, cascading provider test. Add deferred responses delivered out of order; type three distinct queries during one slow request; sort while filter pending; interact with two widgets simultaneously; rapid multi-selection and clear. These are missing focused race assertions, not a claim an existing test fails. Primary source: tagged `ajax`, `RequestQueue.admit`, `#determineSyncStrategy`, `#getRequestQueue` and request-body collection preceding admission.

## B1-04 — Reduce table filter payload reconstruction with named native inputs

**Investigate/replace; P2; medium confidence.**

[Filter state/runtime, lines 8–93](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/table.js#L8) reconstructs URLs, removes pagination, copies extra query settings, reads current sort from thead and installs one document request listener per table. [Filter controls, line 210](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/table/table.templ#L210) store keys/defaults in data attributes and Alpine state instead of exposing the whole filter payload as successful named controls.

A native filter form or explicitly included control region can own named search/select/toggle fields, while a shared request-source policy and narrowly inherited `hx-include` carry them to sort/pagination. Hidden inputs can represent server-owned sort and fixed parameters. Keep Alpine for expansion/instant presentation only. This may delete the global `config:request` listener, manual FormData writes and duplicate URL construction.

This is a contract refactor, not a one-line attribute change: false/empty currently omitted, unchecked checkbox serialization differs, extra parameters can overlap filter keys, filtering resets page but sorting/pagination preserve filters, and nested consumer forms cannot accept a nested `<form>`. Consider a non-form included region in that case. Do not add `changed` to a wrapper/form: v4 compares source `.value`, which containers do not supply. Retain current `hx-preserve` until tests prove controls are outside replacement targets or morph ownership is deliberately defined.

Existing tests: table filter, inline filter and sort/filter/pagination suites; `TestSecurityAttackSurfaceTableFilterTargetDoesNotExecuteScript` and filter-key equivalent; table coverage render tests. Add exact query payload comparisons for empty/false/default values, repeated keys, hostile labels/keys, two independent tables and consumer form embedding. Primary source: tagged `#collectFormData`, `#collectInputValues`, `#attributeValue` and changed trigger guard.

## B1-05 — Expose deliberate status and request-state policies without expanding every config indiscriminately

**Replace/extend; P2; high confidence in capability, medium on API shape.**

[Form HTMX config](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/form/types.go#L53) exposes verbs/target/swap/encoding but no root attribute hook for status, sync or disable. Many action wrappers expose still smaller typed configs; button/radio and primitive inputs already have escape hatches. [Button loading behavior](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/button/button.templ#L43) is already declarative `hx-disable` plus CSS state: keep it.

Native `hx-status` can route validation/error HTML and deliberately avoid replacing a form with an unexpected error body. Add a narrowly designed root hook or selected typed fields first where a real consumer needs them; reuse shared action emission only when configurations genuinely share semantics. The form that owns the request must own `hx-disable` for submit buttons: putting it on an ordinary child submit button does not move request ownership in explicit-inheritance v4.

Do not globally suppress all 5xx: [combobox provider error](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/combobox/handler.go#L144) deliberately returns 502 HTML and `HX-Retarget` to render a retry UI. A blanket noSwap rule would break that feature. Likewise modal/dropdown close-on-click may be product policy, not an obsolete network workaround.

Existing tests: form submit/field validation, button loading and action variants, combobox provider-error handler tests. Add slow submit duplicate-click protection with pre-disabled controls, abort/failure recovery, expected 422 fragments, unexpected 500 body retention and expected 502 combobox retry rendering. Primary source: tagged `#handleStatusCodes`, `#disableElements` and `#attributeValue`; inspect request ownership, not just rendered attribute presence.

## B1-06 — Keep replacement/OOB ownership; trial whole-root morph only on an appropriate component

**Retain/investigate; P2; high confidence in retained reasons, medium in simplification candidate.**

[Combobox shell](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/combobox/combobox.templ#L245) owns ephemeral open/focus state, while [body and label OOB](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/combobox/oob.templ#L3) update server-owned selection. With compat, a server-mode whole-root `outerMorph` is a plausible trial to remove the split body/label response and duplicated render paths while keeping isOpen. It is not yet an approved replacement: prove hidden selected values, option aria state, search text, focus, trigger styling and intentional resets all reconcile.

Do not extrapolate to Select or arbitrary Alpine providers. A1 independently probed same-root morph of Select: dataset options changed while the visible old label remained, because provider config was read once and reactive state survived. Native morph preservation does not refresh application-owned cached arrays. Also field validation can auto-generate dependent slug input values; preserving dirty input values through morph could hide those server changes. Keep explicit primary/dependent replacement semantics unless ownership is redesigned.

The table migration already uses `<hx-partial>` for its head; retain it. Stable pagination hosts and toast append-OOB remain valid and useful. Native partial syntax is not a reason to rewrite every working OOB fragment. In particular do not remove empty pagination hosts or first-render/update distinctions.

Existing tests: `TestCombobox_Toggle_PreservesIsOpen`, server toggle/cascade, `TestFormValidation_Dependency_SlugAutoUpdates`, `TestFormValidation_ValuePreservation`, table pagination transitions and `TestHTMX4AlpineSwapLifecycle`. Add whole-root repeated morph tests with hidden values and focus, provider-config change tests, and explicit reset-versus-preserve assertions. Sources: tagged compat hooks and [morph guide](https://four.htmx.org/docs/morphing-swaps-guide); the generic lifecycle test demonstrates the primitive, not every component's suitability.

## B1-07 — Retain explicit component boundaries and useful wrappers

**Retain; P3; high confidence.**

[Row link guard](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/table.js#L98) prevents nested anchors/buttons/inputs from activating the surrounding row. htmx inheritance changes do not replace event-propagation or keyboard semantics. Keep it and its `TestTableLinkedRowNestedControls` coverage. Header-suffix `click.stop` and `keydown.stop` serve the same boundary; do not delete them alongside obsolete inheritance workarounds.

[Pagination attributes](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/pagination/pagination.templ#L49) could be made inherited on a common ancestor, but explicit leaf attributes support independently rendered fragments and consumer overrides. Only consolidate when the ancestor is guaranteed to survive every response. Actiongroup/splitbutton adapters preserve their own disabled/link/action precedence; shared rendering must retain those distinctions. Native links remain valuable for modifier keys and fallback navigation.

[Field identity](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/form/form.templ#L326) and `validation.Handle`'s `X-Goshtoso-Field` form value are a real application contract after removal of HX-Trigger-Name. Do not “simplify” by guessing field identity from target IDs or reintroducing htmx 2 headers. `TestFormValidation_NativeChangeCarriesFieldIdentity` is the relevant guard. Primary source: tagged htmx core request headers and explicit inheritance lookup.

## Suggested implementation order and unresolved gates

1. A2-01/B1-01: single tab request path with request-count and retry behavior tests.
2. B1-02: native sentinel trial with both scroll-root layouts and request counts.
3. B1-03/04 together: request ownership, payload and ordering before deleting filter runtime.
4. B1-05 only for concrete consumers; avoid API churn across every action config.
5. B1-06 as an isolated server-combobox experiment, with A1's configuration/state warning as a gate.

The temporary tabs probe is `/tmp/gs-b1-probe.cjs`; it is intentionally outside tracked files. No proposed production table, sync, status or morph change was implemented during this report. A narrow native sentinel capability probe was added during pair review, described below; it does not establish full component parity. Pair B2 challenged sentinel lifecycle parity, synchronization assumptions and the whole-root combobox experiment; the completed cross-review is recorded below.


## Pair B cross-review addendum

Shared B1-02/03/06 with B2 for independent challenge. Independently inspected native cleanup and late-load initialization: htmx cleans powered descendants, disconnects each trigger observer and clears its timers; a late-loaded core initializes the body. This supports removing the custom dependency-ready sentinel restart in ordinary loading, but the fallback-loader browser scenario remains an implementation gate. `intersect once` removes its event listener on attempt but retains its observer until cleanup. Configured `revealed` can also carry root/rootMargin and disconnects on intersection; the warning against **bare** revealed is about losing the existing root/margin policy, not an absence of those capabilities.

A narrow exact-vendor browser capability probe after the initial report used a 100px-tall `.scroller`, a 900px spacer, then a table row with:

```html
<tr id="sentinel" hx-get="/data" hx-target="this" hx-swap="outerHTML"
    hx-trigger='intersect once root:"closest .scroller" rootMargin:"400px 0px"'>
  <td>loading</td>
</tr>
```

The same-origin intercepted response was `<tr id="done"><td>loaded</td></tr>`. After late-loading the vendored core, request count was zero; setting the scroller's `scrollTop` to 500 produced exactly one request and one `#done` row. This validates parsing of the quoted extended root selector, 400px margin, and row replacement in an isolated fixture. It does not test chained pages, failures, dependency fallback, Alpine removal or real layout geometry. Probe: `/tmp/gs-b1-sentinel.cjs`.


B2 independently accepted the quoted-root/margin capability, native late initialization, explicit-target ajax source ownership and shared-queue reasoning. B2 agreed that whole-root combobox morph must remain an experiment with provider-config ownership gates. In reciprocal review, B1 checked B2's proposed removal of `siteRuntimeScripts`' v0.1.0 manifest fallbacks and the sidebar's manual `Alpine.initTree`/`htmx.process` calls: both are credible pruning candidates in the modern-only deployment contract. The latter still needs OOB sidebar first-init/navigation/focus coverage; redundant calls do not by themselves prove duplicate Alpine initialization because framework guards can make them idempotent. B2 owns those findings and the componentdocshell/site TOC overlap; this report does not create duplicate action items.

# Modernization task list

Canonical scope, decisions and acceptance criteria: [MODERNIZATION.md](MODERNIZATION.md). All 24 tasks are resolved; full integration gates below remain authoritative for release readiness.

| Task | Driver | Status | Resolution |
|---|---|---|---|
| A01 | A1 | Implemented + tested retention | Removed dead watcher teardown; retained explicit Select binding because native/normalized modelable changes empty/invalid parent semantics. |
| A02 | A1 | Implemented | Client Combobox restores persistence once per Alpine mount and on persisted pageshow; fragment entry no longer misses restoration. |
| A03 | A1 | Implemented | Server toggle/clear return one full-root outerMorph response; options-only search remains narrow. Real mutation tests preserve focused search/caret and cumulative selection. |
| A04 | A1 | Tested retention | Select/Search/StructuredInput configuration changes require replacement; probes demonstrated stale options, cache and indexed defaults under morph. |
| A05 | A1 | Implemented | Validation field identity is static escaped JSON hx-vals; removed indirect wrapper-name lookup. |
| A06 | A1 | Implemented | Dependent validation renders a copied OOB config; shared FieldGroup state is never temporarily mutated, even on errors/cancellation. |
| A07 | A2 | Implemented | Read/submission coordination uses scoped context tracking, native admission and readonly search locking; a minimal guard covers the pinned upstream queue bug. |
| A08 | A2 | Implemented | Native sentinel intersection with explicit contained/page root, margin and retry button replaces observer registry and restart plumbing. |
| A09 | A2 | Implemented | Named included filters replace imperative AJAX/global request listener; scoped normalization preserves empty/false/query precedence. |
| A10 | A2 | Implemented | A stable per-table request owner coordinates filter/sort/pagination; exact-source cleanup avoids cancelling unrelated descendant work. |
| A11 | A2 | Implemented | Form root attributes, Sync/Disable and validation policy support deliberate status and submission behavior; pre-disabled controls stay disabled. |
| A12 | A2 | Reviewed and verified | Pair A closed indicator ownership, extra per_page, descendant cleanup and real mutation-focus findings. Native include parsing now preserves prior selections and dependencies. |
| B01 | B1 | Implemented | One declarative selected-panel load owner; failures clear pending state and retry on reselection, including offscreen activation. |
| B02 | B1 | Implemented | Navbar uses shared popover state/focus ownership; keyboard opening exposes correct expanded state. |
| B03 | B1 | Implemented | ActionGroup has idempotent directive/htmx cleanup for observers, frames and pending font callbacks. |
| B04 | B1 | Implemented | Shared disposable Toast provider waits for public x-show hide completion; scoped timers, idempotent dismiss and monotonic IDs handle bursts. |
| B05 | B1 | Implemented | Generated server identities namespace implicit accordion roots/items; explicit caller IDs remain unchanged. |
| B06 | B1 | Implemented | ScrollRegion uses processed-subtree enhancement and cleanup under htmx/Alpine removal; departed children are unobserved. |
| B07 | B2 | Implemented | Removed historical manifest-role fallbacks while retaining declared loading order and real network fallback. |
| B08 | B2 | Implemented | Active docs shell exclusively owns TOC/scroll; legacy layout remains supported and redundant OOB initialization is removed. |
| B09 | B2 | Implemented | Removed no-op watch bookkeeping/deep option; Alpine markup owns ticker/log event listeners with original pause and insertion semantics. |
| B10 | B2 | Tested retention | Rejected native drawer trap after rapid open/close left Tab blocked on a closed drawer. Restored existing trap and added a permanent regression. |
| B11 | B2 | Tested retention | Changed Tabs/Carousel/Tooltip configuration requires replacement; retained transition-aware popover focus safeguards. |
| B12 | B2 | Reviewed and verified | Retained SSE cancel-before-abort and processing of Alpine-created connectors; both pair programmers reviewed the final shell/lifecycle decisions. |

## Combined integration gates

- [x] All 24 tasks resolved with implementation or tested retention rationale.
- [x] Pair A reciprocal code review closed; four concrete findings fixed and independently retested.
- [x] Pair B reciprocal code review closed; native-trap blocker reproduced red, restored implementation tested green.
- [x] Generated templ/JS/CSS/skill and runtime integrity checks clean.
- [x] Root unit tests and root/site lint pass.
- [x] Current-source and public-pinned site contracts pass.
- [x] Full Goshtoso browser suite passes.
- [x] Full app-shells unit/browser suite passes against the modernized public library pin.
- [x] Final commits and public dependency pins persisted on feature branches.

Evidence archives: [A1](results-A1.md), [A2](results-A2.md), [B1](results-B1.md), [B2](results-B2.md). These preserve experiment and reciprocal-review details; the consolidated report above owns the final decisions.

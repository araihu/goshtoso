# Consolidated htmx 4 / Alpine modernization report

This is the canonical implementation report, consolidating the four reviewer reports at Goshtoso `e5ecabb2` (implementation baseline `707126f3`) and app-shells `7b3d1f0`. Target runtimes remain htmx 4.0.0, Alpine/core plugins 3.17.2 and official hx-alpine-compat. htmx 2 compatibility is not required. The original reports remain as evidence archives; this report and [TASKS.md](TASKS.md) own decisions and execution status.

## Findings and approach

The sweep covered all 55 component directories, 20 library scripts, 14 site scripts and all three shell families plus ComponentPage. Narrow browser fixtures confirmed duplicate lazy-tab requests, incorrect navbar keyboard ARIA, missed client Combobox restoration on fragment entry, ineffective watcher bookkeeping, competing TOC observers, and stale Select configuration during a proposed morph. These findings are not all production traces; rendered-component regressions must accompany fixes. Native table intersection root/margin and row replacement were capability-probed. Other refactors are source-backed candidates requiring parity checks.

Modernization has three purposes: remove duplicate request/lifecycle owners, use declarative framework behavior where it preserves the component contract, and make server data versus retained local state explicit. Native htmx morph/partial responses are new opportunities; Alpine modelable, scoped IDs, watcher cleanup and provider lifecycle are established APIs. A simpler implementation is accepted only after its original behavior is covered. A failed experiment is resolved by retaining the current mechanism and recording concrete evidence, not by forcing a rewrite.

## Two equal worklists and reciprocal review

Pair A owns tasks A01–A12 (inputs, forms, table requests). Pair B owns B01–B12 (overlays, lifecycle, site and shells). Each pair has two programmers with disjoint initial file ownership; each reviews the other's patch, challenges tests and fixes findings before integration. Four subagents are scheduled within the available three-worker concurrency limit. Each has an isolated worktree; the coordinator integrates reviewed commits, regenerates shared assets and runs combined gates.

Within Pair A, programmer A1 drives A01–A06 and A2 drives A07–A12. Within Pair B, B1 drives B01–B06 and B2 drives B07–B12. Review tasks include verification work and do not require gratuitous production edits. Shared Combobox/form files are handed off explicitly; shared generated assets and dependency pins are finalized by the coordinator.

## Consolidated task specifications

### A01 — Select binding

Driver: A1; original findings: A1-01/02.

Remove ineffective watcher disposer bookkeeping; prototype native x-modelable while preserving default/empty/invalid parent precedence, hidden-input restore and focus.

Acceptance: Select dependent binding, external restore, defaults, repeated mounts and keyboard tests.

### A02 — Client Combobox mount persistence

Driver: A1; original findings: A1-03.

Restore persisted client selection once per newly mounted root, including fragment entry; preserve opt-out, pageshow and live edits.

Acceptance: Direct/fragment parity, two instances, malformed/disabled storage, BFCache and no-HTTP behavior.

### A03 — Server Combobox morph

Driver: A1; original findings: A1-04/B1-06.

Prototype one full-root server response using native outerMorph; adopt only if server selection/ARIA/labels reconcile while open/focus state remains correct.

Acceptance: Toggle/search/cascade, provider error, hidden values, caret, removed options and repeated morphs; record a concrete rejection if parity fails.

### A04 — Input morph ownership

Driver: A1; original findings: A1-05.

Define server-data versus user-state ownership for Select, Search and StructuredInput. Add targeted reconciliation only where morph is adopted; retain replacement elsewhere.

Acceptance: Changed/reordered/removed config, dirty values and intentional reset checks. No global morph conversion.

### A05 — Static field identity

Driver: A1; original findings: A1-06.

Encode known validation metadata as static hx-vals; retain explicit server field identity and safe escaping.

Acceptance: Native change, dependent validation, missing metadata and unusual field names.

### A06 — Dependent validation partials

Driver: A1; original findings: A1-08.

Remove temporary mutation of reusable FieldGroup OOB state using native partials or a copied configuration if that is clearer.

Acceptance: Multiple dependents, render errors/cancellation, reuse after success/failure, slug updates and value preservation.

### A07 — Form and Combobox synchronization

Driver: A2; original findings: A1-07/B1-03.

Define read-only latest-result and validation/submit coordination; avoid cancellation policies that lose mutation intent. Coordinate Combobox markup with A1.

Acceptance: Delayed/out-of-order responses, independent fields/widgets, search/toggle and submit during validation.

### A08 — Native table sentinel

Driver: A2; original findings: B1-02.

Replace custom sentinel observer/request/restart plumbing with native intersection trigger and explicit root/margin/swap.

Acceptance: Contained/page scrolling, late load, append chains, one request per page, failure/retry and removal.

### A09 — Native table filter payloads

Driver: A2; original findings: B1-04.

Use named included controls and deliberate parameters to remove document request listener and duplicate URL reconstruction where parity is proven.

Acceptance: False/empty/default, sort persistence, page reset, extra parameters, two tables, nested consumer forms and escaping.

### A10 — Shared table request queue

Driver: A2; original findings: B1-03.

Coordinate filter/sort/pagination on stable per-table request ownership so stale requests cannot overwrite newer intent.

Acceptance: Slow filter plus sort/pagination, out-of-order responses, query bursts and widget independence.

### A11 — Form status and disable policy

Driver: A2; original findings: B1-05.

Add narrowly useful root hooks/fields for status, synchronization and request-owned disabling; exercise a real demo consumer.

Acceptance: Double submit, pre-disabled controls, expected422 versus unexpected500, and preserve Combobox502 retry UI.

### A12 — Request boundary regression review

Driver: A2; original findings: B1-07.

Review A1 implementation; preserve field identity, nested row/action guards and intentional OOB/pagination hosts; close Pair A regressions.

Acceptance: Run focused combined input/form/table suites and document reciprocal review findings and resolutions.

### B01 — Single lazy Tabs request

Driver: B1; original findings: A2-01/B1-01.

Replace dual effect/intersection request engines with one selected-panel request owner and explicit retry semantics.

Acceptance: Exactly one request, initial/offscreen/hash selection, repeated selection, failure/retry, indicator and fragment entry.

### B02 — Navbar menu ownership

Driver: B1; original findings: A2-02.

Fix keyboard expanded state and consolidate avatar menu into shared popover where its focus/slot contract fits.

Acceptance: Enter/Space/ArrowDown, Escape/outside click, focus ownership, mobile, reduced motion and repeated swaps.

### B03 — ActionGroup resource ownership

Driver: B1; original findings: A2-03.

Make observers, frames and pending font callbacks disposable under htmx and Alpine removal; retain actual width allocation.

Acceptance: Connect/disconnect counts, x-if/removal, partial collapse, keyboard action focus and responsive layouts.

### B04 — Toast timer lifecycle

Driver: B1; original findings: A2-04.

Centralize disposable/idempotent timer ownership and replace fixed removal delays with actual transition completion plus safe cancellation.

Acceptance: Auto/persistent, hover, rapid dismiss, parent removal, reduced motion, notification cap and multiple containers.

### B05 — Accordion identity

Driver: B1; original findings: A2-05.

Namespace implicit root/item identities without changing explicit caller IDs or server-addressable selectors.

Acceptance: Repeated default/configured-root accordions, ARIA relationships before JS and after replacement, explicit IDs.

### B06 — ScrollRegion lifecycle

Driver: B1; original findings: A2-08/B2-02.

Trim redundant after-swap enhancement scans after proving after-process coverage; release observers/frames including Alpine removal and removed children.

Acceptance: Primary/secondary partials, standalone loading, changing children, x-if teardown and scroll geometry.

### B07 — Runtime manifest pruning

Driver: B2; original findings: B2-01.

Remove old site-manifest role compatibility while retaining ordered core/compat/plugins/providers/Alpine and real network fallback.

Acceptance: Manifest/head tests, local/fallback startup and both site module contracts.

### B08 — One active shell controller

Driver: B2; original findings: B2-02/03.

Remove redundant OOB init calls after lifecycle proof; give current shell exclusive TOC/scroll ownership while retaining legacy layout behavior.

Acceptance: Actual-site observer counts, OOB init, deep links/history/focus, reduced motion and legacy frame.

### B09 — Site Alpine lifecycle

Driver: B2; original findings: B2-04.

Remove no-op watch disposer/deep option and replace manually bound stream listeners with Alpine-owned declarative listeners when timing matches.

Acceptance: First message, pause before connection, cancellation, fragment teardown and theme changes.

### B10 — Docs drawer native focus trap

Driver: B2; original findings: B2-05.

Prototype bundled x-trap with explicit close-reason focus policy; remove manual trap only if all accessibility behavior remains correct.

Acceptance: Tab/ShiftTab/outside focus, nested overlays, breakpoint, Escape/backdrop, navigation/history and no-JS fallback.

### B11 — Overlay morph and focus boundaries

Driver: B2; original findings: A2-06/07.

Review B1 lifecycle work and establish replacement/morph ownership for Tabs/Carousel/Tooltip. Keep transition-safe popover focus safeguards unless a proven smaller equivalent exists.

Acceptance: Changed config, removed active items, closing/reopening, external focus and resource counters; document retained boundaries.

### B12 — Streaming and shell integration guard

Driver: B2; original findings: B2-06.

Preserve SSE cancel-order repair and explicit htmx.process for Alpine-created connectors; review Pair B integration and all shell families.

Acceptance: SSE cleanup, pause/resume connection counts, WS chat, shell history/focus and full shell suite.

## Retained boundaries

Keep transition-aware popover focus ownership, safe encoding/nil normalization, native control semantics, search JSON API and cancellation, server reset behavior, stable OOB/pagination destinations, nested action guards, IndexedDB/object-URL cleanup, ordered runtime acquisition and immutable upstream bytes. SSE reader cancellation must precede abort until upstream cleanup is fixed. Alpine-created htmx connectors still need explicit processing. The compatibility extension does not own application timers, caches or data reconciliation. A03/A04/B10/B11 are controlled experiments and contract work, not authorization for blanket morph/default changes.

## Validation and completion

Every task requires a specific implementation or evidence-backed retained decision, focused validation and partner review. Run regeneration, JS generation/integrity, root and site unit tests/lint, both current-source and public-pinned module contracts, the full Goshtoso browser suite and the full app-shells browser suite after integration. If new root APIs are adopted, publish a reachable feature commit then update public pseudo-version pins in sequence; never hide the pinned contract with a committed replace. Record failures and fixes rather than treating prior migration success as evidence for this new implementation.

## Primary sources and archived evidence

- [htmx 4 release](https://four.htmx.org/announcements/2026-08-28-htmx-4.0.0-is-released), [tagged core](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/htmx.js), [tagged compat](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-alpine-compat.js), [tagged SSE](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-sse.js).
- [Alpine 3.17.2](https://github.com/alpinejs/alpine/releases/tag/v3.17.2), [modelable](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/directives/x-modelable.js), [watch](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/magics/%24watch.js), [Focus](https://alpinejs.dev/plugins/focus).
- Archived [A1](01-alpine-inputs.md), [A1 probes](01-alpine-inputs-probes.md), [A2](02-alpine-overlays.md), [B1](03-htmx-components.md), [B2](04-htmx-lifecycle.md) preserve exact source anchors, coverage ledgers, reproduction recipes and original candidate confidence.

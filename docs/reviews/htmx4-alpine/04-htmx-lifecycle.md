# Pair B2 — Runtime, site and app-shell lifecycle

Report only, 2026-09-10. Goshtoso baseline `707126f3cb4d1efda821c047905d64768abefff3`; app-shells baseline `7b3d1f0c7aaccd29138cceafb7d28a516ce1293c`. Runtime: htmx 4.0.0, official hx-alpine-compat, Alpine/core plugins 3.17.2. No implementation, generated assets, dependency pins or tests changed. Prior migration suite success is not validation of these proposed deletions.

The best small removals are obsolete site-manifest compatibility branches and ineffective watcher-disposal bookkeeping. The larger opportunity is making one shell own navigation/TOC work. Keep the SSE cancellation repair, manual processing of genuinely Alpine-inserted connectors, and the ordered runtime loader.

## B2-01 — Remove obsolete site runtime-manifest compatibility branches

**Remove; P2; high confidence, static contract review.**

[Runtime selection](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/site/internal/pages/demo/runtime_dependencies.go#L35) builds `declared`/`seen` maps, supplies a missing dark-mode role, and appends fallback SSE/WS roles explicitly described as support for v0.1.0. Both current-source and publicly pinned site contracts now depend on the migrated manifest with these roles. User explicitly permits breaking compatibility. Remove those old-manifest paths and keep the manifest-driven order, site enablement map, and site-provider insertion before Alpine.

This removes multiple descriptions of optional roles and dead branches; it does not remove the CDN-to-local network fallback. No new upstream API is needed: this is pruning made possible by the new minimum dependency baseline. Evidence is the checked-in `site/go.mod`, `assets/runtime.overlay.yaml` and generated manifest, rather than a claim Alpine 3.17.2 introduced a loader feature.

Validate `site/internal/pages/demo/runtime_dependencies_test.go`, `components/head/runtime_manifest_test.go`, `TestDependenciesLocalRuntimeBootsReusableComponentBundle`, then both documented site module contracts. Preserve exactly one compat script after htmx and before Alpine, plugin-before-core order, first-party provider availability and optional SSE/WS inclusion. B1 independently confirmed these branches are historical compatibility, not required fallback resilience.

## B2-02 — Let native swap processing own OOB initialization; trim duplicate enhancement scans

**Remove candidate; P2; high confidence in duplicated responsibility, medium in removal parity.**

[Legacy demo navigation handler](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/site/assets/js/src/demo-layout.js#L105) manually runs `Alpine.initTree(sidebarContent)` and `htmx.process(sidebarContent)` following every main-content swap. The sidebar comes from [the OOB response](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/site/internal/pages/demo/fragment.templ#L15), not external DOM insertion. On the newer shell this legacy sidebar ID is absent, so the calls are normally no-ops. On the legacy frame they duplicate core/compat ownership. Remove them after an OOB lifecycle regression test; do not assert that two calls necessarily initialize Alpine twice, since Alpine markers make many repeats idempotent.

Similarly, [ScrollRegion](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/scroll-region.js#L88) scans on both `after:process` and `after:swap`, with a context-target fallback. Prefer the processed subtree (`event.target`) for htmx enhancement and retain document-ready/load support for standalone runtime use. Candidate removal is the redundant after-swap scan, not its ResizeObserver/MutationObserver/scroll measurement or cleanup.

New integration support: [tagged hx-alpine-compat](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-alpine-compat.js) defers and flushes mutations around swaps; [tagged htmx `process` and swap implementation](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/htmx.js) processes inserted content and emits `after:process` on that subtree. Secondary partials require this coverage. Avoid replacing these hooks with `after:init`, which only concerns powered elements.

Existing tests: `TestHTMX4AlpineSwapLifecycle`, `TestHTMX4PartialEnhancesCodeBlock`, ScrollRegion E2E coverage and sidebar fragment-navigation tests. Add an OOB sidebar with init/destroy counters, nested htmx controls and keyboard navigation; test primary and secondary partials, replacement, morph, manual insertion and standalone runtime loading. Prove exactly-once enhancement and zero retained observers after repeated removal. B1 accepts the responsibility overlap but correctly requires runtime evidence before deletion. No removal was browser-tested here.

## B2-03 — Give the active shell exclusive TOC and scroll ownership

**Consolidate; P2; high confidence in duplicate owners, browser fixture confirmed.**

[Site TOC builder](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/site/assets/js/src/demo-layout.js#L157) and [componentdocshell TOC builder](https://github.com/araihu/goshtoso-app-shells/blob/7b3d1f0c7aaccd29138cceafb7d28a516ce1293c/componentdocshell/assets/shell.js#L181) both replace the same link list and observe the same headings when hosted by the site. [Site configuration](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/site/internal/pages/demo/componentdocshell.go#L49) supplies the legacy `toc-rail`/`toc-list` IDs and the shared site bundle. Each runtime also remembers sidebar scroll and handles main swaps. The TOC observers have different root margins and link/active styles; the site implementation hardcodes smooth scrolling while the reusable shell respects reduced motion.

Keep the legacy layout's TOC where that frame is rendered, but gate it to that layout or isolate its bundle. Let componentdocshell own its TOC, history focus and scroll work. Do not delete the entire site demo bundle: it registers streaming providers, storage consent and the SSE repair. This is ownership consolidation enabled by the completed shell migration, not a new Alpine directive.

Actual temporary Chromium probe: construct `#page-scroll > #main-content` with two `[data-toc-heading][id]` headings and `#toc-rail[data-componentdocshell-toc][data-enabled=true] > #toc-list[data-componentdocshell-toc-list]`; instrument IntersectionObserver to record creation/disconnection; load the exact authored componentdocshell shell script then `demo-layout.js`; dispatch DOMContentLoaded. The fixture is already document-ready when scripts are injected: demo-layout builds immediately, while the shell registers its DOMContentLoaded callback. The synthetic event triggers that shell callback; it does not install a second demo-layout callback. Instrumented disconnect marks each replaced observer inactive. Result: **two live observers**, margins `0px 0px -75% 0px` and `0px 0px -70%`, one link list. This proves duplicated ownership in an equivalent fixture, not a production trace or a complete visual regression. Script: `/tmp/gs-b2-toc.cjs` (temporary only).

Validate `TestPagination_DeepLinkKeepsTOCRailAttachedToContent`, `TestLogFeed_SidebarScrollPreservedDuringStream`, shell `TestFamilyNavigationHTMXHistoryAndFocus` and component-doc smoke tests. Add actual-site observer counts, heading active state after primary/OOB navigation, hash deep links, back/forward, reduced motion and legacy-layout rendering. Keep the site-specific family mapping in `site-navigation.js`; the shell cannot infer those routes. Native history refetch does not generate a TOC or define accessibility focus policy.

## B2-04 — Remove ineffective watcher disposal and use Alpine-owned event listeners

**Remove/replace; P2; high confidence in watcher no-op, medium in event refactor parity.**

[demoLayout watcher bookkeeping](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/site/assets/js/src/demo-layout.js#L65) stores `_stopThemeWatch`, checks whether it is a function in `destroy`, then resets it. Tagged Alpine [the `$watch` magic](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/magics/$watch.js) registers cleanup internally and returns no disposer. Keep the watch and remove the ineffective stored field/destroy wrapper. This corroborates A1's Select finding; it is the same cleanup class, a separate site location. `theme-page.js` additionally passes `{deep:true}` as a third `$watch` argument, which this two-argument magic ignores; deep observation is already native. Removing that option is clarity-only.

[Ticker state](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/site/assets/js/src/ticker-pane.js#L12) stores a connector and two bound listener references solely to listen for before-message/error and remove them later. Candidate replacement: `x-on:htmx:sse:before:message` and `x-on:htmx:sse:error` on the existing connector, preserving `connected = true` and conditional `$event.preventDefault()` while paused. That can remove connect/disconnect callbacks, fields and `$nextTick` setup. [Tagged `x-on`](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/directives/x-on.js) owns listener teardown. These are established Alpine APIs, not features first introduced in 3.17.2.

Existing tests: `TestTicker_ProviderBootstrapsFirstPaint`, `TestTicker_PauseStopsUpdates`, `TestTicker_FragmentNavNoErrors`, theme switching/fragment tests. Add first-message timing under local and delayed loading, repeated replacement, pause before first message and before-message cancellation checks. Ticker pause deliberately leaves transport open and suppresses updates; do not accidentally adopt the Logs disconnect-on-pause contract. LogFeed's bound after-swap listener could similarly become `x-on` calling its existing `onSwap($event)`, but keep event scoping, row cap and animation-frame scrolling.

## B2-05 — Trial the bundled Focus plugin for the docs drawer's manual focus trap

**Replace candidate; P2; medium confidence, static only.**

[componentdocshell focus runtime](https://github.com/araihu/goshtoso-app-shells/blob/7b3d1f0c7aaccd29138cceafb7d28a516ce1293c/componentdocshell/assets/shell.js#L7) implements its own focusable-element selector, Tab wrapping, focusin containment, initial focus watchers and document listener cleanup. The shell already loads full Goshtoso dependencies including Focus. Trial `x-trap` gated by `sidebarOpen && !sidebarPersistent` on the sidebar to remove the manual trap machinery. Keep responsive matchMedia state, closed-sidebar inertness and explicit close reasons.

[Official Focus documentation](https://alpinejs.dev/plugins/focus) supports expression-controlled traps, nesting and return-focus modifiers. This predates 3.17.2. Default focus restoration is not equivalent to current navigation policy: an htmx navigation should finish focused on main content, while Escape/backdrop should return to the trigger. Consider `.noreturn` with explicit close restoration, and prove activation occurs after drawer visibility settles. Keep the existing implementation until this is demonstrated.

Validate shell `TestFamilyNavigationVisualMatrix`, `TestFamilyNavigationHTMXHistoryAndFocus`, `TestFamilyNavigationWithoutJavaScript`, with first/last Tab and Shift-Tab, programmatic outside focus, nested modal/select, breakpoint changes while open, Escape, backdrop, navigation and history. ConsoleShell uses a different drawer implementation; do not assume docs-shell behavior applies automatically. LandingShell already composes the library Drawer and retains its native no-JS details fallback; no parallel trap rewrite needed there.

## B2-06 — Retain the SSE repair and genuinely manual insertion processing

**Retain; P1 if removal is proposed; high confidence from tagged source and existing regression coverage.**

[demo-layout SSE guard](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/site/assets/js/src/demo-layout.js#L3) wraps the exposed connection's abort method to cancel its reader first. [Shipped upstream SSE cleanup](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-sse.js#L345) still aborts then invokes reader cancellation without observing the promise. The migration's unhandled AbortError regression is therefore not made obsolete by merely having an official extension. Keep the narrowly filtered AbortError handling; reevaluate against an upstream fix and rerun `TestHTMX4SSECleanup` before removing it. A DOM before-cleanup listener is not an equivalent replacement for extension-level ordering.

[Logs connector](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/site/internal/pages/demo/examplepages/logs/logs.templ#L127) is inserted by Alpine `x-if`. Its `connect()` waits for Alpine and explicitly processes the connected node. Compat handles htmx-originated changes; it does not make arbitrary Alpine insertion htmx-owned. Keep that call, the connection guard and pause-removes-connector design. `TestLogFeed_PauseStopsAndResumes` and `TestLogFeed_FragmentNavNoErrors` are relevant; add repeated pause/resume connection counts and no post-removal messages. This report did not rerun the streaming suite or reproduce the upstream rejection anew.

## Coverage ledger and deliberate no-change decisions

All 14 authored site JavaScript files were read, including non-htmx behavior. The following ledger covers lifecycle decisions; it is not a general security/performance audit of each example domain.

| Site source | Review/disposition |
| --- | --- |
| `action-group.js` | Small registered demo provider; retain. Registration guard supports shared bundle use. |
| `avatar-showcase.js` | Local selection/computed CSS mappings; retain. No htmx workaround. |
| `charts-showcase.js` | Isolated iframe load-time fallback reveal; retain. Do not change to fragment hook without changing that hosting contract. |
| `icon-catalog.js` | Code encoder, mutually related local inputs and focus restoration; retain source encoding. Overlay consolidation belongs with A2 focus work, not morph. |
| `profile-images.js` | IndexedDB transaction close, revision guards, destroyed flag and object-URL revocation protect asynchronous completion; retain. Compat does not cancel database promises or revoke URLs. |
| `select-demo.js` | Programmatic value assignment plus native change event is an intentional demo contract; retain. |
| `site-bootstrap.js` | Synchronous storage consent and first-paint theme setup; retain. Alpine deferred initialization cannot replace the pre-paint responsibility. |
| `site-navigation.js` | Site route-family mapping on shell navigated event; retain. |
| `theme-page.js` | Timers, root attribute observer, JSON/page-data parsing, CSS export, color calculations and storage reviewed; B2-04 minor watch-option cleanup. Retain observer until theme ownership is unified; shell/external root attribute changes are real inputs. One-time page-data parsing is a morph-adoption gate like A1 Select. |
| `chat.js` | Scoped-by-ID global outgoing composer clearing and incoming scrolling; retain current small hooks. Any future multiple-chat-instance support should scope to event source. New WS extension already owns transport queues/reconnect; no parallel transport needed. |
| `demo-layout.js` | B2-02/03/04/06; separate legacy-frame behavior from shared providers/repair. |
| `landing-playground.js` | Origin+source-checked iframe sizing, ResizeObserver and post-settle height report; retain. Cross-document resize is outside htmx/Alpine compatibility. |
| `log-feed.js` | B2-04/06; retain cap/filter/scroll and explicit insertion bridge. |
| `ticker-pane.js` | B2-04 declarative event candidate; keep pause semantics. |

| Runtime / integration | Review/disposition |
| --- | --- |
| `muamba.yaml`, `assets/runtime.overlay.yaml`, head templates/config and manifest consumers | Retain exact acquisition/integrity/order and CDN fallback. Head local tags need synchronous htmx/compat before shell extra scripts; dynamic loader awaits each script. B2-01 removes only old site-manifest roles. |
| `assets/js/src/dependency-loader.js` | Read full implementation: sequential acquisition, same-version fallback, nonce/integrity, ready/error events, window-load gate. No supported native replacement identified. |
| `assets/js/src/combobox.js` | Read full delegated stateless keyboard runtime. Fresh DOM queries cover fragment replacements. Keep keyboard guards; native htmx morph does not provide listbox navigation. |
| Shared enhancements | Read ScrollRegion and cross-checked action-group/code-block/table hook ownership with A2/B1; B2-02. Keep document-ready enhancement for progressive/standalone use. |
| Client combobox restore | Independently confirmed restore scans only ready/pageshow in `components/combobox-client.js`; A1-03 owns fragment restoration finding. Add processed-root support with once-per-new-root semantics, not indiscriminate reapplication over live edits. |
| `ticker_handler.go`, `logs_handler.go` and streaming templates | Read complete handlers: explicit ticker partials, linewise SSE framing, context cancellation/unsubscribe, bounded log streaming; retain. Native extension does not replace server framing, broker or domain filtering. |
| `chat_handler.go` and chat templates | Traced WS accept/read/write teardown, server-rendered frames, rename hidden-input OOB and composer send; retain server cancellation, identity/payload validation and HTML rendering. No transport wrapper to prune. |
| Vendored compat/SSE/WS | Inspected compatibility hooks and transport initialization/cleanup. Keep upstream bytes unchanged, including upstream's own legacy aliases; user authorization to drop app htmx2 support is not a reason to fork immutable vendor files. |
| App-shells ComponentDocShell | Read shell JS, head/layout, fragment and family/mobile utility templates; B2-03/05. Keep main/body navigation guard, family Select reconciliation and focus policies pending A1 model contract work. Morph is not a automatic replacement for cached Select configuration. |
| App-shells ConsoleShell | Read shell JS, layout and fragment including optional OOB navigation; keep after-settle target guard, main lookup, active-nav reconciliation and drawer-close reason. Test Alpine-generated focus targets before changing lifecycle timing. |
| App-shells LandingShell | Read shell JS/layout and mobile navigation; retain composition with library Drawer, no-JS details, theme bootstrap and local state. No htmx-specific workaround to prune. |
| App-shells ComponentPage | Read page/section composition and CodeBlock integration; static semantic headings/preview/code slots need no htmx conversion. |

The complete 20-library-JS inventory is distributed across reports: A1 owns `components/{combobox-client,data,palette,search,select,structured-input}.js`; A2 owns `action-group.js`, `darkmode.js`, `components/{carousel,dropdown,navigation,popover,sidebar,tabs,tooltip,code-block,scroll-region}.js`; B1 owns `components/table.js`; this report explicitly covers `dependency-loader.js` and standalone `combobox.js`, with shared lifecycle overlap. This attribution is coverage ownership, not a claim B2 reread every component algorithm.

## Pair review and validation limits

B1 and B2 exchanged source challenges. B2 independently checked native intersection options: both `intersect` and `revealed` accept root/rootMargin; `revealed` disconnects its observer immediately, while `intersect once` removes admission listeners and leaves observer cleanup to teardown. B1 added a late-loaded native contained-scroller browser probe. This supports B1-02 but not blanket sentinel deletion before full widget parity. Native body initialization can replace the custom dependency-ready rescan for declarative sentinels.

B2 agrees with B1-03's source-based queues and explicit-target ajax ownership; synchronization must respect mutation semantics. B1-06 remains an isolated server-combobox experiment, not global morph adoption. A1's stale Select dataset/state result remains a prerequisite warning. Tabs A2-01/B1-01 is one finding, not two. B1 independently accepted B2-01 and the guarded B2-02 removal proposal. B2's TOC fixture is independent evidence, not B1 corroboration.

Only the TOC ownership fixture was executed by B2. Source inspection, existing test inspection and partner probe review support the other findings. No full suite was rerun for this report-only change. Start future implementation with B2-01 and watcher bookkeeping, then single-shell TOC ownership and OOB enhancement tests; leave focus-trap substitution and morph experiments as separate behavior changes.

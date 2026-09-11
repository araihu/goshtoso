# Pair A / reviewer 2 — Alpine overlays, navigation, layout and display

Review baseline: Goshtoso `707126f3cb4d1efda821c047905d64768abefff3`, htmx 4.0.0, Alpine/core/focus/collapse/mask 3.17.2. Report only, 2026-09-10. No implementation, test, generated asset, dependency, commit or push changes by this reviewer.

Two candidates already reproduce in isolated browser fixtures: a lazy tab makes two requests on first selection; keyboard opening the navbar avatar menu leaves `aria-expanded=false`. Resource ownership and implicit IDs are the next useful cleanup work. Broadly replacing components with morph swaps is premature: preserving client state also preserves stale server configuration and captured DOM references.

## Evidence and version boundary

Reviewed authored component templates/configuration, the corresponding first-party runtime, and relevant test implementations/coverage names. Generated templ output was not treated as a separate implementation. Browser probes used the exact vendored three runtime files, Playwright Chromium, a mocked same-origin endpoint and component-equivalent markup; they are not full rendered-component regression tests. `/tmp/a2-probe.cjs` produced:

```json
{"lazyRequestsBefore":0,"lazyRequestsAfter":2,"avatarExpanded":"false","menuVisible":true}
```

No unit/full browser suites were run for this documentation-only sweep. Existing tests below describe checked-in coverage, not freshly passing execution.

The new integration capability is htmx 4 built-in morphing plus [tagged hx-alpine-compat](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-alpine-compat.js): it carries Alpine scope through morphing, handles reactive-ID matching and teleported nodes, and defers mutation processing during swaps. It does not supply arbitrary resource cleanup, application data reconciliation or transition-completion callbacks.

`x-id`, provider `destroy()`, declarative events and focus helpers are established Alpine facilities, not inventions of 3.17.2: compare [3.14.9 x-id](https://github.com/alpinejs/alpine/blob/v3.14.9/packages/alpinejs/src/directives/x-id.js) and [3.14.9 x-data](https://github.com/alpinejs/alpine/blob/v3.14.9/packages/alpinejs/src/directives/x-data.js). Tagged [3.17.2 x-data](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/directives/x-data.js) includes expression-replacement reconciliation; do not confuse that with automatic refresh of data attributes read once by a factory. Tagged [x-on](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/directives/x-on.js) owns listener disposal. Tagged [x-if](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/directives/x-if.js) destroys removed subtrees; it is not a leave-animation completion API.

## Findings

### A2-01 — Replace two lazy-tab request owners with one

**P1 · high confidence · replace · browser reproduced.** [tabs.templ:116](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/tabs/tabs.templ#L116).

The panel has an Alpine `x-effect` that marks `dataset.loaded` then calls `htmx.ajax`, alongside `hx-get` and `hx-trigger="intersect once"`. The imperative path loads on selection; the second path independently loads when the selected panel intersects. Marking the dataset does not suppress that trigger. The isolated fixture issues zero requests while hidden and two on first selection. B1 independently reproduced the same count using its own fixture.

Replace the generated URL/target/swap JavaScript and dataset sentinel with one component selection event and a declarative htmx trigger. Dispatch after Alpine updates visibility, with the event arriving after htmx processing. Keep eager-on-selection semantics if the selected panel is offscreen; merely retaining `intersect once` changes that contract. Decide whether once means one attempt or one successful load, and retain deliberate retry behavior after failures. This is simplification enabled by the current native stack, not a newly introduced Alpine directive.

**Benefit:** one request, one indicator/target configuration, fewer interpolated expressions. **Invariants:** no fetch for unselected panels, selected offscreen behavior, initial hash selection, indicator behavior, safe escaped URLs, no duplicate requests after fragment navigation. **Coverage:** `TestTabs_HTMXLazyPanelLoadsOnlyAfterSelection` in `site/tests/e2e/tabs_test.go:49` verifies content after activation but not request count; `TestTabs_HTMXFragmentLifecycle` covers navigation. **Validation:** route-count assertions for first/repeated selection, initially selected tab, hidden/offscreen panel, failed first response then retry, full and fragment loads.

### A2-02 — Consolidate navbar avatar behavior into the shared popover

**P1 · high confidence · replace · browser reproduced.** [navbar.templ:145](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/navbar/navbar.templ#L145).

Navbar retains its own `userDropDownIsOpen`/`openWithKeyboard` state, trap and arrow handlers. Enter/Space/ArrowDown set only the second variable, while `aria-expanded` reads only the first. Probe: the menu is visible after ArrowDown while expanded is false. The existing shared popover already synchronizes its two opening modes, trigger ARIA and close/focus handling.

Render navbar content through the shared popover/dropdown lifecycle, preserving avatar/header/menu-item styling and public slots. At minimum correct the expanded expression before a larger consolidation. This is shared-component reuse and established Alpine focus behavior, not a 3.17-only feature.

**Benefit:** remove a parallel menu implementation and accessibility divergence. **Invariants:** first-item focus, disabled/hidden-item exclusion, no focus theft on Escape from another control, outside click, menu role and accessible name, mobile presentation. **Coverage:** `TestNavbar_AvatarDropdown`, `TestNavbarUserMenuReducedMotionShowsWithoutVisualTransition`, `TestDropdownEscapeFocusRestorationIsReopenSafe`, `TestDropdownEscapeFromExternalFocusPreservesExternalFocus`. **Validation:** keyboard Enter/Space/ArrowDown must expose true expanded state and focus the first eligible item; Escape and repeated swaps must restore only owned focus; exercise light/dark/Minimal and mobile navbar separately.

### A2-03 — Give ActionGroup a disposable root lifecycle

**P2 · high confidence missing teardown; medium confidence retained-resource impact · replace.** [action-group.js:85](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/action-group.js#L85), [observer setup:186](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/action-group.js#L186).

The progressive enhancer owns a ResizeObserver, animation-frame scheduling and `document.fonts.ready` callback, stores an observer on the element, and has no matching disposal path. `isConnected` avoids layout work after removal but does not explicitly release resources. The factory captures primary/secondary/overflow elements and counts once; stale references after a future identity-preserving child update are an **opt-in morph risk**, not a demonstrated current default-swap regression.

Use an Alpine directive with registered `cleanup`, or a shared root lifecycle helper with explicit htmx cleanup and an Alpine-removal path. Track/cancel the frame and invalidate pending font callbacks. Keep resize measurements: Alpine has no native width-aware overflow allocator. Re-query/reconcile children when supporting morph updates, or document replacement-only configuration.

**Benefit:** deterministic disposal and a clear future morph contract. **Invariants:** progressive server-only output, action order, partial collapse, hidden dropdown closure and focus transfer. **Coverage:** `TestActionGroupResponsiveTransformAndAccessibility`, `TestActionGroupPartialCollapseKeyboardNavigation`, `TestActionGroupFragmentNavigationUsesBundledProvider`; these do not prove observer disposal. **Validation:** instrument observer connect/disconnect and pending frames over repeated htmx replacement and Alpine `x-if` removal. Separately morph an added/deleted secondary action and verify measured live nodes and counts before enabling that path.

### A2-04 — Centralize toast timers and replace fixed removal delays

**P2 · high confidence · replace.** [toast/types.go:195](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/toast/types.go#L195), [toast.templ:82](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/toast/toast.templ#L82), [server dismissal:233](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/toast/toast.templ#L233).

Client variants repeat timer expressions without `destroy`; the container schedules another untracked 400ms removal. Message dismissal can first wait 400ms and then invoke the second delay. Server toasts immediately remove the node on `toast-dismiss`, bypassing their leave transition. These paths implement auto-dismiss and leave timing independently.

Use one registered toast provider with timer handles and idempotent close/destroy methods. Clear auto-dismiss/removal timers on destroy and before replacement. Complete removal from an explicit, public animation-completion mechanism with a zero/reduced-motion branch and cancellation fallback, rather than a hardcoded 400ms or private Alpine `_x_*` hooks. Do not simply switch to `x-if`: removal is not coordinated with leave transitions. Scope pause/resume to the owning container if multiple containers are supported.

**Benefit:** fewer duplicated handlers, no stale timeout callbacks after replacement, consistent dismissal timing. **Invariants:** persistent server toasts, live-region announcement, hover pause, maximum notification count, repeated dismiss idempotency, reduced motion. **Coverage:** `TestToastCoverageDemo`, `TestToastReducedMotionShowsAndDismissesWithoutVisualTransition`, `TestToastStaticExamplesStayVisible`. **Validation:** normal/reduced-motion close completion, immediate parent removal before timeout, rapid notify/dismiss, mouse pause/resume, two containers, and queued notification delivery during fragment replacement. The existing 400ms value is a local workaround, not an htmx compatibility requirement.

### A2-05 — Fix implicit accordion ID collisions; consider scoped Alpine IDs selectively

**P2 · high confidence source finding · replace.** [accordion.templ:16](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/accordion/accordion.templ#L16), [implicit item IDs:60](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/components/accordion/accordion.templ#L60).

The default root ID is `accordion`, and missing item IDs become `accordion-item-N`; controls/content then share those values across repeated components. Even distinct configured root IDs do not namespace implicit items: `ContainerID` is passed to item data but not used for these IDs.

Prefer the smallest server-side namespace correction using a unique component identity, preserving caller-specified IDs. For purely internal interactive IDs, evaluate `x-id` plus `$id` so ARIA references resolve within a component. htmx compatibility now handles Alpine-bound IDs during morph matching, reducing one integration concern; `x-id` itself is long-existing. Do not replace addressable IDs used in htmx selectors or deep links with client-only IDs.

**Benefit:** reliable control-to-region association with repeated components. **Invariants:** explicit public IDs, server-rendered accessibility, static/no-JS output, deterministic server targets. **Coverage:** `TestAccordion_Accessibility` checks individual relationships; `TestCoverageRenderDefaultAccordion` checks the default rendering. **Validation:** two accordions without item IDs, including with distinct configured root IDs; verify document-wide uniqueness and each aria-controls/labelledby pair before/after morph and replacement. A1 agreed that server namespacing is probably the narrower change and that explicit IDs must remain stable.

### A2-06 — Make server configuration refresh explicit before adopting morph widely

**P2 · high confidence architectural blocker; per-component behavior needs probes · investigate.** [tabs.js:6](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/tabs.js#L6), [carousel.js:7](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/carousel.js#L7), [tooltip.js:241](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/tooltip.js#L241).

Tabs captures config/valid IDs once. Carousel seeds slides/autoplay configuration once. Tooltip portal closes over a panel and skips repeat setup via a dataset flag. Morphing the same root can preserve desired local state alongside stale server-derived arrays or replaced descendant references. A1 independently reproduced a stale label in Select after morph; that is corroboration of the class of risk, not proof of every component listed here.

Define UI-owned state (selected tab, current slide, open state) separately from server-owned data (available tabs/slides/content). Add deliberate reconciliation and invalid-selection fallback, or retain replacement swaps. Use htmx4 morph only for specific stable-identity refresh contracts; do not bulk-convert every `innerHTML`.

**Benefit:** safe use of the genuinely new integration capability without silent stale content. **Invariants:** removed selected item falls back predictably, refreshed slide content actually appears, no duplicate autoplay/listeners, changed tooltip trigger retains correct accessible description. **Coverage:** `TestHTMX4AlpineSwapLifecycle` proves generic state transfer; `TestCarousel_HTMXLoadedSlidesRemainInteractive` and `TestTabs_HTMXFragmentLifecycle` prove fragment interaction, not changed-config morph reconciliation. **Validation:** morph same root with changed/removed/reordered configuration, retained edited child controls, active traps, `x-for`/`x-if`, and provider init/destroy counters. Test teleports only where introduced; no owned overlay here currently uses `x-teleport`.

### A2-07 — Retain popover focus-ownership safeguards; do not replace them blindly with `$focus`

**P2 regression guard · high confidence · retain, with a bounded simplification experiment.** [popover.js:62](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/popover.js#L62), [menu filtering:225](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/popover.js#L225).

The MutationObserver/fallback frame loop waits for the menu to become hidden, protects against reopen/destroy and avoids moving focus from unrelated controls. Menu navigation filters actual visible, enabled menu items. These are not obsolete Alpine-startup workarounds.

The [tagged focus plugin](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/focus/src/index.js) enumerates with `displayCheck: 'none'`; trap deactivation follows reactive state rather than waiting for leave completion. Replacing the custom list with unfiltered `$focus.wrap()` or removing `.noreturn` would change those guarantees. A narrow experiment can pass the already filtered item array to `$focus.within(...)`, but it is minor deduplication and needs focus timing parity. No verified 3.17.2 public API automatically replaces the full close-and-focus state machine.

**Coverage and validation:** retain `TestDropdownEscapeFocusRestorationIsReopenSafe`, `TestDropdownEscapeFromExternalFocusPreservesExternalFocus`, both `TestDropdownEscapeWithoutMutationObserver...` cases, and `TestReviewDropdownAlpineDestroyInvalidatesQueuedFocusRestore`. Add root morph while closing and while externally focused before considering deletion. A1 agreed explicit resource cleanup remains necessary despite Alpine's automatic `$watch` cleanup.

### A2-08 — Keep scroll measurement; improve ownership across non-htmx removal

**P3 · medium confidence gap · investigate/retain.** [scroll-region.js:12](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/scroll-region.js#L12), [cleanup:103](https://github.com/araihu/goshtoso/blob/707126f3cb4d1efda821c047905d64768abefff3/assets/js/src/components/scroll-region.js#L103).

ScrollRegion already disconnects observers/listener/frame during htmx cleanup. Retain resize/mutation observers: they measure overflow and changing content, not Alpine initialization. Its disposal entrypoint is tied to htmx; Alpine-owned `x-if` or direct DOM removal deserves explicit coverage. `observeContent` also observes new children but does not unobserve removed children until whole-region disposal.

Consider the same declarative cleanup owner as A2-03; track/unobserve departed children if the viewport is long-lived. Adding Alpine Intersect solely to replace geometry is not a clear improvement and introduces an unbundled plugin. **Coverage:** `assets/js/src/components/runtime_test.go:195` checks cleanup source markers, and log sidebar-scroll E2E checks behavior, neither proves per-child/resource disposal. **Validation:** repeated child replacements and region removal under both htmx and Alpine; observer counters, scroll indicators after resize/font change, pending frame cancellation. Do not label current behavior a demonstrated leak without that probe.

## Coverage ledger and deliberate retention

All 38 remaining component directories were scanned for authored Alpine/htmx behavior, configuration and related runtime (input ownership belongs to A1). No-change means no migration-specific simplification justified by this sweep, not exhaustive accessibility certification.

| Directories | Disposition |
|---|---|
| accordion | A2-05; retain `x-collapse`, disabled semantics and reduced-motion classes. |
| actiongroup | A2-03; retain real width allocation and action-focus transfer. |
| alert, banner | Small `x-show` dismissal state already idiomatic; no added provider/morph default justified. Keep transition/reduced-motion behavior and caller actions. |
| avatar | Retain image `complete`/`naturalWidth` startup checks and source-expression guards; network completion can precede Alpine listeners regardless of compatibility extension. Do not prune as an old Alpine bug. |
| appshell | Layout slots/classes, no owned Alpine state; no change. External app-shells reviewed by Pair B. |
| badge, breadcrumbs, card, chatbubble, emptystate, icon, inlinecode, kbd, link, pageheader, panel, skeleton, spinner, steps, toolbar | Server-rendered display/composition; no owned Alpine lifecycle to simplify. Preserve semantics, reduced motion and safe URL/rendering paths. |
| button, splitbutton, pagination | Declarative action/HTMX configuration or composition; no extra Alpine layer justified. Request API opportunities belong to B1. |
| carousel | A2-06; existing interval/media-query `destroy` retained. Keep reduced-motion changes and safe-navigation validation; compatibility cannot cancel browser timers for the application. |
| codeblock | Delegated clipboard enhancement remains reasonable; no need to introduce Alpine for a single button. Optional cleanup of the short reset timeout on removed buttons is lower value than A2-03/04. |
| drawer, modal | Keep event-ID routing, close-request contract, focus trapping and reduced motion. Drawer deliberately omits `.inert` for sibling dialog interaction. Native Dialog `.noreturn` avoids double focus restoration; don't strip scroll lock because the dialog is native. `TestContentDialogAboveDrawer` is a key invariant. |
| dropdown, popover | A2-07 and reuse in A2-02. The historical `goshtosoDropdown` factory alias is component API compatibility, not htmx 2 support; removing it needs an explicit component API decision. |
| head | Current core/compat/Alpine order retained; loader/fallback review belongs to B2. |
| navbar | A2-02; mobile menu is a separate layout/focus behavior and needs its own regression checks during consolidation. |
| scrollregion | A2-08. |
| sidebar | Keep per-instance state, event-ID/ARIA relationships and disclosure behavior. `closeAndFocus` is much simpler than popover; add external-focus/reopen tests before unifying implementations. |
| table | Local checkAll/openRows/filter state: define preservation/reset on data refresh before morphing. Retain `x-collapse` and suffix propagation guards; server request/OOB review belongs to B1. |
| tabs | A2-01, A2-06; keep roving tabindex/manual selection and hash opt-in. |
| toast | A2-04. |
| tooltip | A2-06; retain custom-trigger attribute ownership restoration and idempotent keyboard listeners. The browser top-layer portal solves clipping; replacing it with `x-teleport` would also require positioning/focus/event-forwarding design. |

## Pair review and implementation order

A1 challenged broad morph adoption with its reproduced stale Select configuration. I accepted that constraint and classified A2-06 as a prerequisite/probe, not an immediate default-swap conversion. A1 corroborated the implicit-ID issue and recommended the smaller server namespace fix. I challenged deletion of explicit timer/listener teardown on the basis of Alpine owning watchers: only watcher bookkeeping is redundant. Its Select `$watch` return-value cleanup candidate is independent and worth doing; none of my providers contained that same dead bookkeeping.

B1 independently reproduced A2-01 with a separate fixture. Canonical duplication finding is A2-01; the server/component report should cross-reference rather than count a second issue.

Suggested later implementation sequence: A2-01/02 with browser regressions; A2-03/04 resource ownership; A2-05 implicit IDs; then A2-06 reconciliation/morph experiments. Retained safeguards A2-07/08 should constrain cleanup work. Full suites should run after actual code changes; this report makes no claim to have rerun or certified them.

## Reproduce the isolated probes

Use a blank same-origin Playwright page, route `/lazy` to return `<p>Loaded</p>` with `Content-Type: text/html`, and increment a request counter in that route. Before loading the runtime, set this body (the fixture deliberately excludes styling and unrelated template slots):

```html
<style>[x-cloak]{display:none!important}</style>
<div x-data="{selectedTab:'first'}">
  <button id="choose" @click="selectedTab='details'">Details</button>
  <div x-cloak id="panel" x-show="selectedTab==='details'"
    x-effect="if(selectedTab==='details'&&!$el.dataset.loaded&&window.htmx){$el.dataset.loaded='true';htmx.ajax('GET','/lazy',{target:'#panel',swap:'innerHTML'});}"
    hx-get="/lazy" hx-trigger="intersect once" hx-swap="innerHTML">Loading</div>
</div>
<div x-data="{userDropDownIsOpen:false,openWithKeyboard:false}">
  <button id="avatar" :aria-expanded="userDropDownIsOpen"
    @keydown.down.prevent="openWithKeyboard=true">User</button>
  <ul id="menu" x-show="userDropDownIsOpen||openWithKeyboard"><li>A</li></ul>
</div>
```

Load, in sequence, `assets/js/runtime/htmx.org/4.0.0/htmx.min.js`, `assets/js/runtime/htmx.org/4.0.0/hx-alpine-compat.js`, and `assets/js/runtime/alpinejs/3.17.2/alpine.min.js` via Playwright `addScriptTag({path: ...})`. Wait 100ms for initialization and record the route count (0). Click `#choose`, wait 500ms and record the count (2). Focus `#avatar`, press ArrowDown and wait 50ms; read `#avatar`'s live `aria-expanded` (false) and `#menu` visibility (true). These waits are probe instrumentation, not proposed permanent test synchronization. The navbar probe omits the Focus plugin because it tests only the independently incorrect expanded-state expression; full component regressions must load the normal manifest and assert focus behavior too.

Coordinator verification: repeated the isolated probe with the `x-cloak` guard shown above, matching initial hidden-panel behavior and avoiding initialization timing ambiguity; observed the same `0 -> 2` request count and incorrect navbar expanded state. This does not replace a rendered-component regression test.

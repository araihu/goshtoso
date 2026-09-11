# Pair B programmer 2 — site and shell lifecycle results

B2 drives B07–B12 at plan `0ff8117c`, with B1 `86659f1` applied for integration and reciprocal review. Coordinator owns shared generated bundles and public dependency pins.

## B07 — manifest pruning implemented

Removed declared/seen maps and historical missing-role fallbacks. Runtime order comes only from the locked manifest; the site bundle still precedes Alpine. Core/compat ordering and real CDN/local acquisition fallback remain. Head and site runtime manifest tests pass, as do demo page units. Public-pinned and combined contracts remain coordinator gates.

## B08 — one shell controller implemented

Legacy site TOC and scroll handlers now defer to the component-doc-shell root marker, including layouts with a disabled TOC. Legacy frames retain their TOC with reduced-motion-aware scrolling. Removed blanket Alpine.initTree/htmx.process for the htmx-owned secondary sidebar. Providers, storage consent and the SSE repair remain available to both frames.

`TestModernSiteHasOneTOCOwner` instruments IntersectionObserver on the actual Pagination page, then performs Tabs fragment navigation; exactly one live heading observer remains each time. `TestModernSecondaryPartialOwnsInitialization` verifies repeated secondary sidebar partial swaps initialize once per new Alpine scope and destroy the replaced scope once. Existing htmx partial CodeBlock enhancement and shell navigation/history suites provide additional coverage.

## B09 — Alpine lifecycle implemented

Removed demoLayout's ineffective watcher disposer and theme-page's ignored deep option. Ticker before-message/error and Logs after-swap listeners now live in Alpine-owned markup, removing bound callback storage and manual listener teardown. Ticker pause still cancels message application while keeping transport alive; Logs still removes its connector on pause.

`TestModernTickerDeclarativeCancellation` proves a paused cancelable message is prevented, a resumed message is admitted, and error resets connection status. Updated first-paint coverage to wait for a real message setting connected=true instead of inspecting the deleted connect method. Existing stream-update, pause/resume, fragment and cleanup tests exercise real transports.

## B10 — tested retention after reciprocal review

The native `x-trap.noreturn` prototype passed the original full app-shells browser suite (41.590s) and a nested-trap probe (1.318s), but those checks missed rapid open/close. B1 supplied a blocking red/green regression: open the drawer, let Alpine's effect run, and close before Focus 3.17.2's deferred 15ms activation. The delayed activation survives closing, so the next Tab is prevented behind the closed, inert sidebar.

`TestDocsDrawerRapidCloseReleasesTab` failed against the native prototype (0.640s). App-shells correction `a813133f` restores the prior manual focus lifecycle and template and removes the native-only nested-trap test. The regression and docs visual matrix then pass (45.362s); shell units pass. B2 inspected the permanent regression and correction and accepts this blocker and resolution. The manual listeners check current open/persistent state for every event, and the queued initial-focus callback rechecks state before moving focus.

No additional 15ms scheduling workaround or vendored Focus patch was introduced. That would couple shell behavior to upstream timer internals instead of simplifying ownership. Native adoption awaits an upstream activation-cancellation fix and rerunning the permanent regression. Existing manual behavior retains its separate nested-overlay limitations; this task does not claim to resolve those. ConsoleShell and LandingShell trapping remain unchanged.

## B11 — replacement boundaries retained with browser evidence

Prototyped outerMorph against actual rendered Tabs and Carousel using valid base64 configuration and bundled htmx4/compat/Alpine. Tabs root identity survived and its dataset became `{default: "new-server-tab", ids: ["new-server-tab"]}`, but selectedTab remained `groups`. Carousel dataset became an empty slide array while Alpine retained 3 slides. A custom Tooltip trigger fixture changed its content ID from old-content to new-tooltip-content while its aria-describedby retained old-content. These are opt-in morph blockers, not defects in the shipped replacement paths. No blanket morph default or speculative reconciliation was introduced.

`TestModernOverlayConfigurationReplacement` proves root replacement consumes changed Tabs/Carousel configuration and resets ownership correctly. Changed active-item identity requires replacement today. Existing Popover/Dropdown external-focus, reopen, no-MutationObserver fallback and destroy regressions validate retained transition-aware ownership; replacing that mechanism with immediate unconditional return would violate those contracts. Drawer containment is a different interaction contract from menu return-focus behavior.

## B12 — streaming guards retained

SSE4 still aborts before reader cancellation; retain the existing cancel-order wrapper and its narrow AbortError handling. Alpine-created Logs connectors still explicitly call htmx.process on the next tick after insertion. Removing either is not justified by compat, which owns htmx-originated mutations. Existing SSE cleanup, real Logs pause/resume and WS Chat coverage remain required coordinator gates. The initial three-family suite passed on the prototype; final shell verification must use the B10 correction, not treat the prototype run as complete parity evidence.

## Critical reciprocal review of B1

Reviewed the full B1 patch and focused regressions, especially initial/hash Tabs processing before event dispatch, finally-request error/retry admission, Toast close/destroy idempotence and timer container scope, enhancer dual cleanup paths, Accordion pre-JS IDs and Navbar shared focus state. No confirmed blocker found. Independently ran B1's new Modern regressions with streaming/site tests. Challenged the assumption that state-preserving morph automatically consumes new server configuration; the B11 probes confirm replacement must remain the default. No request to delete valid observer/focus safeguards.

B1's reciprocal review approves the remaining B2 changes after applying the B10 correction. B2 also reviewed B1's follow-up `1b7c40dd`: the named Toast helpers preserve container-specific hover events and per-notification completion, while server Toast defaults to root removal. No blocker found. B1 reports Toast/runtime units, JS extraction checks and the focused Toast browser suite passing (9.463s). The coordinator's shell lint cleanup was reviewed by B1. Pair B mutual review is closed with the native-trap blocker resolved by tested retention; final integration gates remain coordinator-owned.

## Validation notes

Head/Tabs/Toast and all demo page units pass; app-shells unit suite passes. Site demo lint reports zero issues. Initial app-shells lint reported six preexisting findings; the coordinator subsequently fixed them and B1 reviewed the cleanup. Generated templ output and authored diff checks pass. Full combined generation, CSS, root/site lint and public-pinned gates are coordinator-owned.

During test authoring, corrected Playwright's required map[string]any argument and used the actual base64 data contract for configuration replacement. These fixture errors did not justify changing runtime behavior. Temporary morph probes were removed after execution; the recorded state/config outcomes above preserve their evidence.

Final focused combined browser run passed (106.174s): Ticker, LogFeed, HTMX4, all Modern regressions, Pagination deep link, Dropdown Escape/focus, Popover, Tooltip and Carousel. This includes B1's modernization regressions and B2's corrected configuration/secondary-partial tests. No skips were intentionally requested. Initial app-shells prototype checks and final B10 correction checks are distinguished above.

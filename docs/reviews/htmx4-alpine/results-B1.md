# Pair B programmer 1 — component modernization results

Driver worktree: `feat/modern-b1`; baseline is the consolidated plan at `0ff8117c`. Partner review is recorded below when completed. Shared minified artifacts and public pins are integrated by the coordinator.

## B01 — one selected-panel Tabs request

Removed the second intersection request engine and imperative `htmx.ajax` options. The selected panel dispatches one custom event to its own declarative `hx-get`, target, swap and indicator attributes. A pending guard prevents concurrent duplicate loads; htmx 4 `finally:request` clears it even after network failure. Only a 2xx/3xx response marks it loaded. Selecting another tab and returning retries failures. Selected offscreen panels remain eager. The explicit subtree `htmx.process(panel)` before initial dispatch covers Alpine's initial/hash selection before htmx's first scan; it does not reinitialize Alpine.

Evidence: `TestModernTabsSingleRequestAndRetry` exercises real rendered Tabs, offscreen selection, HTTP 503, a rejected fetch, success and repeated selection, asserting exactly three requests total. Existing hash, fragment lifecycle, spinner and keyboard tests remain applicable. Custom swap responses remain caller-owned; broad morph conversion was not adopted.

## B02 — shared Navbar menu state

The existing avatar/header/UL markup now uses the shared `goshtosoPopover` factory, including keyboard opening, ARIA synchronization, menu-item filtering and owned-focus restoration. Removed Navbar's separate boolean and unfiltered `$focus` navigation. Preserved its original placement, motion styling and public slots. The mobile menu remains a separate layout controller.

Evidence: `TestModernNavbarKeyboardOwnsARIAAndFocus` verifies Enter, Space and ArrowDown expose expanded=true, focus a menu item and restore the trigger on Escape. Existing desktop/mobile and reduced-motion Navbar tests and the shared popover's external-focus/reopen guards pass. No shared popover safeguards were deleted.

## B03 — ActionGroup resources

Added an Alpine custom directive cleanup owner alongside progressive DOM/htmx enhancement. It disconnects ResizeObserver, cancels the queued frame and invalidates the pending font callback. Initialization remains idempotent. An empty `x-init` enables Alpine discovery even without an enclosing x-data scope; caller x-init remains authoritative. htmx cleanup also disposes the resource for the progressive path. Width measurement, partial overflow allocation and focus transfer are retained.

Evidence: `TestModernEnhancerRemovalDisposesObservers` instruments real ResizeObservers on rendered ActionGroup, removes the root, remounts its markup through x-if and removes it again, requiring every owned observer to release its targets. Existing responsive/partial keyboard/fragment tests pass. Configuration-changing root morph is deliberately not enabled; captured child configuration requires replacement.

## B04 — Toast lifecycle

Added one shared `goshtosoToast` provider for client and server notifications. It owns the auto-dismiss timer, idempotent close and destroy cleanup. Removal observes x-show's public DOM result (`display:none`) after leave completion, including reduced motion; it does not depend on private Alpine fields or a fixed 400ms delay. Parent removal cancels both timer and completion observer. Client container removal is immediate only after that completion. Container-scoped pause/resume prevents hover in one container pausing another. Monotonic per-container notification IDs replace millisecond timestamps, preventing duplicate keys during a synchronous burst; the existing 20-notification cap remains.

Evidence: `TestModernToastBurstAndIdempotentDismiss` dispatches 25 synchronous notifications, checks the cap and repeated close under normal/reduced motion. `TestModernToastTimersAreContainerScopedAndDisposable` instruments timer ownership for two real rendered containers, verifies pause isolation and disposal on parent removal. Existing static persistent, OOB/action, stacking and reduced-motion tests pass. Server dismiss events now request close instead of bypassing the leave transition.

## B05 — Accordion IDs

Explicit caller root/item IDs remain unchanged. Implicit items are now namespaced by their configured root. Missing root IDs receive a concurrency-safe server render identity, preserving complete ARIA links before JavaScript runs. Consumers requiring stable HTMX selectors across independent renders should configure ID.

Evidence: `TestRepeatedAccordionIdentityRelationships` renders four accordions (two default, two configured) into one document and checks global ID uniqueness and every reciprocal control/region relationship. Existing explicit-ID rendering and browser accessibility behavior are retained.

## B06 — ScrollRegion lifecycle

Removed the redundant after-swap scan; after-process scans the processed subtree directly, including partial targets. Added Alpine directive teardown for direct DOM and x-if removal while retaining htmx cleanup and initial progressive scanning. Departed viewport children are now explicitly unobserved. Scroll/mutation/resize measurements remain because they implement the component's overflow semantics.

Evidence: the shared instrumented observer test covers rendered ScrollRegion direct removal and x-if creation/removal. Existing source checks continue to require htmx subtree disposal. Root/configuration morph remains replacement-only; no stale captured sentinel configuration is introduced by a new morph default.

## Verification and review

- Targeted units: Tabs, Navbar, Accordion, Toast, ActionGroup, ScrollRegion, authored runtime parsing and JS tooling pass.
- Existing focused browser suite passed (48.664s) before additional regressions.
- New modernization browser regressions passed (9.984s), including both motion modes, x-if ownership and network retry.
- `templ generate`, `cmd/jsbuild`, focused root lint and `git diff --check` run locally; the coordinator owns final generated assets and the combined full suites.
- Combined focused browser suite (existing and new regressions) passed: 58.573s. Tagged E2E lint reports only two preexisting unchecked Body.Close calls in iconpack_consumer_test.go:75,112; none in the owned regression file.
- The pre-commit hook failed while staging the ignored legacy skill reference. Its go-fix steps were run explicitly; generated references remain for coordinator integration.
- Pair B reciprocal review: pending programmer B2 checkpoint; no claim of review completion yet.

## B2 reciprocal review and extraction follow-up

Reviewed Goshtoso `85f5396615ae81c5b441c3a823c00c79b6e1da2b` and app-shells `72e64eda1608cf4f5880ab000c4789ffb48fe752`, including new tests, manifest order, shell ownership guards, secondary partial initialization, log/ticker declarative listeners and native drawer trapping. Confirmed that native trap focus filtering uses the trap's full display check; the `$focus` helper's displayCheck:none warning does not apply to this adoption. Explicit `.noreturn`, responsive inertness and navigation focus ownership remain intact. Streaming cancel-order and manual insertion guards are retained. No confirmed source blocker was found. Requested one integration check of rapid open/close against Alpine's delayed trap activation; any confirmed outcome should be recorded by the coordinator before closing that acceptance gate.

Moved the four new inline extraction candidates into named toast helpers, without raising the baseline: container hover dispatch and client completion now live in the shared runtime; server removal is its default completion. `go run ./cmd/jslint` reports 14 baselined candidates and **no new findings**. Toast/runtime units pass; the focused Toast and ModernToast browser suite passes in **9.463s**, including burst cap, repeated dismissal in both motion modes, scoped hover timers and parent removal. B2 reviewed the B1 implementation and reported no blockers.

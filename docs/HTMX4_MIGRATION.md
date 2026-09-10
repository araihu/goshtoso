# htmx 4 and Alpine runtime migration

Goshtoso now bundles htmx 4.0.0, Alpine 3.17.2, matching collapse/focus/mask
plugins, and htmx's `hx-alpine-compat` extension. SSE and WebSocket extensions
come from the same htmx 4.0.0 distribution. This is a breaking runtime upgrade;
htmx 2 event names, extension APIs and configuration are not supported.

## Consumer changes

- Use colon-separated events, such as `htmx:after:swap`. Request and swap
  handlers read `event.detail.ctx`; settle handlers read `event.detail.task`.
  Enhance inserted subtrees on `htmx:after:process` using `event.target`, so
  secondary partial updates also initialize their component behavior.
- Shared ancestor attributes need `:inherited`. The default remains explicit
  inheritance; Goshtoso does not enable a compatibility bridge.
- `hx-disable` now disables controls during requests. Use `hx-ignore` to exclude
  a subtree from htmx processing.
- Full versus fragment responses must follow `HX-Request-Type` (`full` or
  `partial`), including history navigation. `HX-Target` contains `tag#id`.
- Field validation submits `X-Goshtoso-Field` as a form value. Custom clients
  must send it along with `X-Goshtoso-Validation=field`; `HX-Trigger-Name` is gone.
- All responses except 204/304 swap by default. Review validation/error responses.
  OOB-only responses leave the main target unchanged by default.
- Load htmx core and `hx-alpine-compat` before Alpine. Goshtoso's direct local
  tags load both synchronously; Alpine and its plugins remain deferred. The
  dynamic dependency loader executes the manifest sequentially, without the
  former window-load wait on htmx. Custom manifests must preserve this order.
- `head.DependencyHTMXAlpineCompat` supports URL, integrity and omission options
  like other runtime dependencies. Keep its version matched to htmx core.
- Replace `hx-ext`, `sse-connect`, `sse-swap`, `ws-connect`, and `ws-send` with
  the htmx 4 extension contracts. Unnamed SSE events swap HTML; named events
  dispatch DOM events. Ticker updates now use explicit HTML partials.

Alpine-created htmx elements still require processing when htmx has not inserted
or scanned them. The log example processes each connector from its Alpine
initialization hook, including resume and fragment navigation. Do not add blanket
`Alpine.initTree()` calls after swaps; compatibility manages Alpine's mutation
processing through settlement.

## Regression findings

The browser suite exposed extension execution before core, stale event payloads,
SSE connector initialization after fragment navigation, and full-page history
requests receiving fragments. The expense filter used `changed` on a form,
which has no input value; it now debounces bubbled input events directly.
Table headers use explicit `hx-partial` responses because an OOB element inside
an ordinary template is not discovered by htmx 4.

The official hx-sse 4.0.0 cleanup aborts a stream before calling `reader.cancel()`
and leaves that rejected promise unhandled. The demo navigation runtime wraps
the connection's abort method to cancel the reader first; other errors remain
visible. Upstream vendored bytes are unchanged. Recheck this workaround when
upgrading hx-sse. The browser regression checks navigation away from both ticker
and logs, aborted connections, released readers, and unhandled rejections.

A rejected expense submission also replaced the live form despite the returned
`hx-preserve` marker in the tested 4.0.0 flow. The handler now returns only OOB
updates on rejection, preserving the input without relying on that path. This
observation is not a confirmed upstream defect report.

## Local validation (2026-09-10)

- Root, site and app-shells unit suites pass; root and site Go lint report zero
  issues. Vendored integrity, runtime generation, JavaScript checks and both
  updated skill validators pass.
- All 426 top-level browser tests were exercised in two complete-suite
  partitions. Two fixture races failed initially: log auto-scroll asserted
  before its animation frame, and a sidebar fixture replaced the landing
  document while a lazy request was outstanding. Both fixtures were corrected
  and passed 10 consecutive reruns. One conditional pagination subtest skips
  when its selected fixture has no next-page control; dedicated sorting and
  pagination tests pass.
- Added browser regressions cover Alpine morph identity/state, replacement
  initialization/cleanup, secondary-partial enhancement, and SSE teardown.
  App-shells browser checks cover console and component-docs history/focus,
  plus drawer behavior in light/dark and Goshtoso/Minimal themes.
- Current-source site integration passes. Standalone pinned-dependency
  deployability fails until the coordinated releases below provide the new
  runtime API and shell lifecycle through public module versions.

## Coordinated release

The migration spans Goshtoso and goshtoso-app-shells. Local validation uses a
workspace containing the root library, site and app-shells checkout. No local
replacement is committed to a module manifest.

Release the root Goshtoso runtime first, pin that reachable version in
app-shells and release its htmx 4 lifecycle changes, then pin both reachable
versions in `site/go.mod`. The standalone pinned-site gate cannot validate the
new runtime while that file still points to a pre-migration release. Run both
site module contracts after updating those public pins.

Sources: [release](https://four.htmx.org/announcements/2026-08-28-htmx-4.0.0-is-released),
[migration guide](https://four.htmx.org/docs/whats-new-in-htmx-4),
[Alpine compatibility](https://four.htmx.org/extensions/hx-alpine-compat),
[Alpine release](https://github.com/alpinejs/alpine/releases/tag/v3.17.2).

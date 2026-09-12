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
- One uninterrupted full browser run passes all 426 top-level tests (1,167
  including subtests), with zero failures and zero skips, in 12m55s. The
  pagination test now uses the actual next-page link and asserts page 2 loads.
  The auto-scroll test waits for the animation-frame result, and the sidebar
  fixture loads as its own document instead of replacing a busy landing page.
- App-shells' complete browser suite passes all 9 top-level tests (112 including
  subtests), with zero failures and zero skips, against the pinned public root
  module with `GOWORK=off`.
- Added browser regressions cover Alpine morph identity/state, replacement
  initialization/cleanup, secondary-partial enhancement, and SSE teardown.
  App-shells browser checks cover console and component-docs history/focus,
  plus drawer behavior in light/dark and Goshtoso/Minimal themes.
- Current-source integration and standalone pinned-dependency deployability
  both pass. A standalone browser smoke run also passes with `GOWORK=off`.

## Coordinated dependency pins

These pins record the migration at the time. The site now uses the library in
the same checkout; see [Site and Published-Package Contracts](SITE_MODULE_CONTRACTS.md).

The migration spans Goshtoso and goshtoso-app-shells. The root runtime commit
was pushed first, then app-shells pinned that reachable revision and published
its migration commit. The site now pins both exact public pseudo-versions:

- Goshtoso: `v0.2.11-0.20260910201950-724e5ba38a60`
- goshtoso-app-shells: `v0.1.9-0.20260910202148-7b3d1f0c7aac`

These versions resolve without local replacements. The feature branches are
`feat/htmx4-alpine` in Goshtoso and `feat/htmx4-events` in app-shells; they are
not stable release tags. Development validation also uses an ignored workspace
containing the root, site and app-shells checkouts. Both site module contracts
must remain green when replacing these pins with future release tags.

Sources: [release](https://four.htmx.org/announcements/2026-08-28-htmx-4.0.0-is-released),
[migration guide](https://four.htmx.org/docs/whats-new-in-htmx-4),
[Alpine compatibility](https://four.htmx.org/extensions/hx-alpine-compat),
[Alpine release](https://github.com/alpinejs/alpine/releases/tag/v3.17.2).

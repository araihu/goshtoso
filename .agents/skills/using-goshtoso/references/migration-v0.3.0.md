# Migrating consumers to Goshtoso v0.3.0

Verified against the public Goshtoso v0.3.0 tag and its `muamba.yaml` on
2026-09-11. This is a breaking runtime upgrade from the 0.2.x line.

## Dependencies and delivery

Pin `github.com/araihu/goshtoso@v0.3.0`. Its supported runtime is htmx 4.0.0,
Alpine.js 3.17.2 and matching Collapse, Focus, and Mask plugins. Keep
`head.Dependencies()` and `assets.Handler()` together: the loader includes
`hx-alpine-compat` and same-version local fallbacks. Remove app-owned duplicate
runtime tags. Use `head.WithLocalRuntime()` for an explicit no-CDN requirement.

For App Shells, the compatible version used by the released site's integration
is `v0.1.9-0.20260910224508-5b2222e54637`. Pin that commit when adding shells
unless a newer release has been verified with HTMX 4. The older `v0.1.8` tag
still uses HTMX 2 events. Inspect the selected public APIs with `go doc`; do not
copy shell internals into the consumer.

## Consumer-owned markup and handlers

- Use colon-separated lifecycle events: `htmx:before:request`,
  `htmx:after:swap`, `htmx:after:settle`, and `htmx:config:request`. Audit custom
  listeners and Alpine `x-on` handlers for old camel-case event names.
- Request event data is in `event.detail.ctx`; the transport uses Fetch.
  Rewrite XHR-dependent handlers against the selected event's documented
  context rather than blindly renaming fields.
- Attribute inheritance is explicit. If child requests rely on a parent target,
  declare `hx-target:inherited="#content"`; test the actual child request.
- Use `hx-disable` for request-time disabling, including
  `hx-disable="find button[type='submit']"` on a form-owned request.
- All response statuses except 204/304 swap by default. Return useful validation
  and recovery HTML for 4xx/5xx responses; deliberately cancel unwanted swaps
  with `htmx:before:swap` instead of preserving a blanket HTMX 2 error workaround.
- Extensions activate when loaded. Remove `hx-ext` and use the HTMX 4 SSE/WS
  extensions supplied by the manifest; old extension packages are incompatible.
- Use native `innerMorph` or `outerMorph` when preserving local state is desired.
  `hx-alpine-compat` coordinates Alpine during settlement and morphing. A
  replacement swap remains appropriate when state should reset. Do not add
  blanket Alpine reinitialization or a separate Alpine Morph plugin.

Keep changes scoped to the consumer's actual integration. Server-owned business
rules, validation, and authorization remain server-owned.

## Validation before shipping the consumer

Regenerate the app's templ files, tidy its modules, and run its Go tests. In a
browser, exercise direct entry, fragment navigation, Back/Forward, repeated
swaps, edited inputs, focus, and any open overlays. Hold a submission pending
to check disabling and deduplication; exercise validation and server-error
responses. Check console errors and CDN failure/local fallback. Test SSE/WS
only when the app uses them. Do not treat a static screenshot as lifecycle
coverage.

## Sources

- [Goshtoso v0.3.0](https://github.com/araihu/goshtoso/releases/tag/v0.3.0)
- [Locked runtime inputs](https://github.com/araihu/goshtoso/blob/v0.3.0/muamba.yaml)
- [Runtime ordering](https://github.com/araihu/goshtoso/blob/v0.3.0/assets/runtime.overlay.yaml)
- [HTMX 4 reference](https://four.htmx.org/reference)
- [Alpine compatibility extension](https://four.htmx.org/extensions/hx-alpine-compat)

# Alpine 3.17.2 with htmx 4.0.0

Verified 2026-09-10. Sources: [compatibility docs](https://four.htmx.org/extensions/hx-alpine-compat), [tagged extension source](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-alpine-compat.js), [Alpine release](https://github.com/alpinejs/alpine/releases/tag/v3.17.2), [installation](https://alpinejs.dev/essentials/installation).

## Installation contract

Use `htmx.org@4.0.0/dist/ext/hx-alpine-compat.js`. This is an htmx extension, not an Alpine plugin. No `hx-ext` attribute is needed in htmx 4.

Illustrative static script order (Goshtoso must express it in its runtime manifest/loader):

```html
<script src="https://cdn.jsdelivr.net/npm/htmx.org@4.0.0/dist/htmx.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/htmx.org@4.0.0/dist/ext/hx-alpine-compat.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/@alpinejs/collapse@3.17.2/dist/cdn.min.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/@alpinejs/focus@3.17.2/dist/cdn.min.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/@alpinejs/mask@3.17.2/dist/cdn.min.js"></script>
<!-- Register first-party alpine:init listeners before Alpine core executes. -->
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.17.2/dist/cdn.min.js"></script>
```

Plugins precede Alpine core. Register providers before initialization; call `Alpine.start()` only once for a manually bundled setup (the CDN build starts itself). Keep versions and local fallbacks pinned together.

## What compatibility handles

It holds Alpine mutation processing through the htmx swap/settle interval, transfers reactive scope during built-in `innerMorph`/`outerMorph`, and accounts for Alpine-bound IDs during morph matching. Those htmx swaps no longer need the separate Alpine Morph plugin. Replacement swaps still represent replacement: compatibility does not promise state preservation for destroyed components.

With optional `hx-history-cache`, it also coordinates Alpine state snapshots/restoration. That extension is a separate choice; test state and cleanup across back/forward if enabled.

## Application responsibilities

- Make component providers available before first initialization, including after fragment navigation. Prefer registration in the shared first-party bundle.
- Let the extension coordinate swaps; avoid redundant blanket `Alpine.initTree()` or mutation-observer workarounds. Remove existing workarounds only after testing their original use cases.
- Use htmx 4 names in Alpine handlers, e.g. `x-on:htmx:after:swap`. Wait for settlement and `$nextTick` when assertions require final DOM state.
- Keep `htmx.process()` for genuinely manual/Alpine insertion of htmx markup when needed.
- Clean up timers, observers and global listeners in component `destroy()`.
- Test focus, open dropdowns, edited inputs, reactive IDs, `x-for`, `x-if`, and teleported content through repeated swaps. Stable identity is a prerequisite for predictable morph preservation.

Read the [Goshtoso migration plan](../../htmx/reference/migration.md) for the actual loader changes and known component/server contracts.

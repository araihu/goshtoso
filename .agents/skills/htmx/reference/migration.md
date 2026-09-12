# Goshtoso htmx 4 migration plan

Research date: 2026-09-10. This records the migration baseline and worklist. The implementation now locks the target versions; use the manifests and tests for current state.

## Verified targets and primary sources

| Dependency | Previous Goshtoso lock | Target |
|---|---|---|
| htmx | 2.0.8 | 4.0.0 |
| Alpine.js | 3.14.9 | 3.17.2 |
| Alpine collapse / focus / mask | 3.14.9 | 3.17.2 each |
| Alpine compatibility | absent | `htmx.org@4.0.0/dist/ext/hx-alpine-compat.js` |

- [htmx 4.0.0 announcement, August 28, 2026](https://four.htmx.org/announcements/2026-08-28-htmx-4.0.0-is-released)
- [htmx tagged release](https://github.com/bigskysoftware/htmx/releases/tag/v4.0.0)
- [What's new / migration guide](https://four.htmx.org/docs/whats-new-in-htmx-4)
- [Official versioned upgrade skill](https://raw.githubusercontent.com/bigskysoftware/htmx/v4.0.0/dist/skills/htmx-upgrade-from-htmx2.md)
- [Official versioned development guidance](https://raw.githubusercontent.com/bigskysoftware/htmx/v4.0.0/dist/skills/htmx-guidance.md)
- [Extension migration guide](https://four.htmx.org/docs/extension-htmx-4-migration-guide)
- [Alpine compatibility documentation](https://four.htmx.org/extensions/hx-alpine-compat) and [tagged implementation](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-alpine-compat.js)
- [Alpine 3.17.2 release](https://github.com/alpinejs/alpine/releases/tag/v3.17.2)
- [Alpine installation](https://alpinejs.dev/essentials/installation)

Registry checks confirmed htmx `next=4.0.0`, `latest=2.0.10`; Alpine core and collapse/focus/mask `latest=3.17.2`. Recheck tags for future upgrade tasks; pin reviewed versions, not moving aliases.

## 1. Inventory and baseline

Start a dedicated worktree from freshly fetched `origin/main` per AGENTS.md. Run the official read-only checker with Go/templ extensions:

```bash
python3 /path/to/htmx.org/dist/scripts/upgrade-check.py --ext .templ --ext .go components assets/js/src site/internal
```

Treat checker findings as candidates, not a complete migration proof. Search tests, Go attribute builders, layout helpers, response handlers, runtime configuration and standalone JS too. Avoid editing generated templ or minified outputs.

## 2. Upgrade acquisition and loading together

Change `muamba.yaml` to the verified versions, acquiring compatibility from the htmx package and locking its bytes with SHA-384. Preserve package/license provenance. Update `assets/runtime.overlay.yaml` to execute htmx core and compatibility before Alpine core, while keeping Alpine plugins and first-party `alpine:init` registrations before Alpine startup.

The pre-migration overlay loaded Alpine before htmx and marked htmx `wait_for_window_loaded: true`. Review that loader gate as part of reordering; a script list alone does not prove execution order. Test standard, minimal, CDN and local-fallback paths, including any Alpine-only/htmx-disabled configuration. Compatibility depends on htmx and must not load without it.

Run `just vendor-js` and `just vendor-js-verify`; regenerate consumers through repository commands. Do not hand-edit hashes or generated runtime constants. Migrate optional SSE/WS packages to htmx 4's `hx-sse`/`hx-ws` extensions, using their current docs to update markup, events and streaming behavior.

## 3. Migrate component and server contracts

Observed areas in origin/main at research time:

- `assets/js/src/components/table.js`: request configuration listener and request payload handling; init/cleanup events.
- `assets/js/src/components/scroll-region.js`, `code-block.js`, `assets/js/src/action-group.js`: old lifecycle/swap events. Search the remaining first-party runtimes and regenerate built bundles.
- `components/button/button.templ`: `hx-disabled-elt`; update public comments/examples and relevant assertions too.
- `components/combobox/combobox.templ`: Alpine `x-on:htmx:after-swap` needs the 4.x event name.
- `components/form/validation/handle.go`: reads removed `HX-Trigger-Name`. Establish an explicit field-name contract or tested source-to-field mapping; `HX-Source` is an element selector, not a form field name.
- `site/internal/server/charts.go`: reads `HX-Target` as a target identifier; account for the new `tag#id` format.
- `site/internal/server/server.go` and example handlers: review full/fragment branching against `HX-Request-Type`, body targets, selection and history.
- Chat/logs/ticker templates: migrate `hx-ext`, legacy SSE/WS attributes and connection events alongside their runtimes.
- Table pagination/filter/OOB paths: verify main-before-OOB ordering, OOB-only behavior, GET/DELETE form inclusion, inherited targets/headers, and validation/error swaps.

Use explicit inheritance and native 4.x events. The migration intentionally drops htmx 2 compatibility; do not add a legacy bridge.

## 4. Validate behavior and release sequencing

Add focused browser coverage for one-time Alpine initialization, repeated replacement and morph swaps, preserved versus reset state, `x-id`/reactive IDs, teleports, focus traps, masks/collapse, and cleanup after removal. Exercise direct loads, sidebar fragment navigation, back/forward, console errors, table sorting/filter/pagination/OOB, validation field selection, loading/disabled controls and streaming reconnects.

Run relevant unit tests, full E2E tests and regeneration gates from AGENTS.md for the actual migration. Verify both `just site-current-source-integration` and `just published-consumer`; keep new root APIs and their site examples in the same PR, as described in `docs/SITE_MODULE_CONTRACTS.md`. Keep runtime, event/markup changes and public compatibility documentation coordinated in release planning.

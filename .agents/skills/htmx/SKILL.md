---
name: htmx
description: Write, debug, and review htmx attributes, HTML responses, request headers, events, JavaScript APIs, and extensions. Covers htmx 4 and migrating Goshtoso from htmx 2, including Alpine.js integration.
---

# htmx

## Version and scope

Target **htmx 4.0.0**, verified 2026-09-10 against the official release and npm registry. npm `next` is 4.0.0 while `latest` remains 2.0.10: pin the exact 4.x version. Start with [sources and migration plan](reference/migration.md) for upgrade work.

Goshtoso locks **htmx 4.0.0 / Alpine 3.17.2** in `muamba.yaml`. This is a breaking migration: author native htmx 4 behavior without htmx 2 compatibility bridges.

Return server-rendered HTML for swaps. Use Alpine for local state and htmx for server interactions. Keep the project's templ and runtime acquisition conventions.

## htmx 4 essentials

- `hx-get/post/put/patch/delete` issue requests; `hx-target` selects the destination and `hx-swap` selects the operation. Default swap is `innerHTML`.
- Inputs/selects/textareas default to `change`, forms to `submit`, other elements to `click`. Use `hx-trigger="input changed delay:300ms"` for active search.
- Parent attributes need explicit inheritance, e.g. `hx-target:inherited="#content"`. Apply it where descendants actually depend on the parent.
- Events use colons: `htmx:after:swap`, `htmx:config:request`. Request data lives in `event.detail.ctx`; XHR assumptions no longer apply.
- Responses other than 204/304 swap by default, including 4xx/5xx. Set deliberate error handling with `hx-status` or `config.noSwap`.
- `innerMorph` / `outerMorph` are built in. Use replacement swaps when state should reset.
- Extensions register when their scripts load; htmx 4 has no `hx-ext` activation attribute. htmx 2 extension packages are not drop-in replacements.

```html
<section hx-target:inherited="#result">
  <button hx-get="/items" hx-swap="innerMorph">Refresh</button>
</section>
<div id="result"></div>
```

## Alpine integration

Use htmx 4's **`dist/ext/hx-alpine-compat.js`**, loaded after htmx and before Alpine starts. It coordinates Alpine initialization with settling, preserves reactive state during morphs, and handles reactive IDs. It is distinct from htmx 2's `alpine-morph` extension; htmx morph swaps do not require `@alpinejs/morph` with this integration.

Read [Alpine compatibility](../alpinejs/reference/htmx4.md) for loading and lifecycle details. Register component providers before initialization; do not add unconditional `Alpine.initTree()` calls after every swap.

## References

- [Migration and sources](reference/migration.md): verified versions, breaking changes, Goshtoso worklist and validation.
- [API quick reference](reference/cheatsheet.md): htmx 4 attributes, headers, events and configuration.
- [Patterns](reference/patterns.md): search, validation, multi-region updates and debugging.
- [Gotchas](reference/gotchas.md): templ, request values, OOB ordering, history and security.

Use the linked official htmx 4 documentation for less common APIs; the legacy `htmx.org/docs` site documents the 2.x line.

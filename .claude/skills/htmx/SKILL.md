---
name: htmx
description: Use when writing, debugging, or reviewing htmx markup — any hx-* attribute (hx-get/post/put/patch/delete, hx-swap, hx-target, hx-trigger, hx-vals, hx-boost, hx-swap-oob, hx-include, hx-sync, hx-indicator), htmx request/response headers (HX-Trigger, HX-Redirect, HX-Retarget, HX-Reswap, HX-Location, HX-Push-Url), the htmx JS API (htmx.ajax/process/on/swap/trigger), htmx events, polling, infinite scroll, lazy load, inline edit, active search, out-of-band swaps, or htmx interacting with Alpine.js / templ attribute escaping. Server returns HTML fragments, not JSON.
---

# htmx

## Overview

htmx lets any HTML element issue AJAX requests via `hx-*` attributes. **The server returns HTML fragments, not JSON.** htmx swaps the returned HTML into the DOM. The mental model: AJAX, CSS transitions, WebSockets and SSE directly in HTML, driven by attributes instead of JavaScript.

Reference docs crawled from htmx.org/docs and htmx.org/reference (htmx v2.0.x).

## When to use

- Writing/editing any element with `hx-get`, `hx-post`, `hx-swap`, `hx-target`, `hx-trigger`, etc.
- Building server-driven UI patterns: active search, infinite scroll, lazy load, inline edit, delete-row, polling, click-to-load, indicators.
- Returning HTML from a handler and deciding swap strategy / OOB swaps / response headers.
- Debugging why an htmx interaction silently fails (often attribute escaping — see gotchas).

**This repo (Goshtoso) uses htmx v2.0.8 + templ + Alpine.js.** Read `reference/gotchas.md` whenever htmx markup is generated through templ or combined with Alpine `x-data` — templ attribute escaping silently breaks both.

## The 5-attribute core

```html
<button hx-post="/messages" hx-target="#result" hx-swap="innerHTML">Send</button>
<div id="result"></div>
```

| Attribute | Does |
|-----------|------|
| `hx-get` / `hx-post` / `hx-put` / `hx-patch` / `hx-delete` | Issue request to URL |
| `hx-trigger` | Which event fires the request (override default) |
| `hx-target` | Where the response HTML goes |
| `hx-swap` | How the response replaces the target |
| `hx-boost` | Turn normal links/forms into AJAX (progressive enhancement) |

**Default triggers:** `input`/`textarea`/`select` → `change`; `form` → `submit`; everything else → `click`.

## Quick reference

| Need | Use |
|------|-----|
| Search-as-you-type | `hx-trigger="keyup changed delay:500ms"` |
| Poll every 2s | `hx-trigger="every 2s"` (respond `286` to stop) |
| Load when scrolled into view | `hx-trigger="revealed"` |
| Run on load | `hx-trigger="load"` |
| Target a relative element | `hx-target="closest tr"` / `next` / `previous` / `find` / `this` |
| Update several regions in one response | `hx-swap-oob="true"` on extra fragments |
| Send extra values | `hx-vals='{"k":"v"}'` / `hx-include="#el"` |
| Loading spinner | `class="htmx-indicator"` + optional `hx-indicator="#spinner"` |
| Confirm before request | `hx-confirm="Sure?"` |
| Disable button during request | `hx-disabled-elt="this"` |
| Cancel overlapping requests | `hx-sync="closest form:abort"` |
| Update URL bar / history | `hx-push-url="true"` |
| File upload | `hx-encoding="multipart/form-data"` |

`hx-swap` values: `innerHTML` (default) · `outerHTML` · `beforebegin` · `afterbegin` · `beforeend` · `afterend` · `delete` · `none`. Modifiers (space/colon): `transition:true`, `swap:100ms`, `settle:50ms`, `scroll:top`, `show:bottom`, `ignoreTitle:true`, `focus-scroll:true`.

## Server side — drive the client from response headers

Return plain HTML, plus optional `HX-*` response headers to control the client:

- `HX-Trigger: myEvent` — fire a client event after the response.
- `HX-Redirect` / `HX-Location` — client-side navigate (Location = no full reload).
- `HX-Retarget` / `HX-Reswap` / `HX-Reselect` — override target/swap/selection from the server.
- `HX-Push-Url` / `HX-Replace-Url` — change the URL bar.
- `HX-Refresh: true` — force full page reload.

Inbound, htmx sends `HX-Request: true`, `HX-Trigger`, `HX-Target`, `HX-Current-URL`, `HX-Boosted`, `HX-Prompt`. Branch on `HX-Request` to return a fragment vs. a full page.

## Updating more than the target — pick the lightest of three

When one action must refresh several regions, escalate only as far as you need:

1. **Expand the selection (simplest).** Make `hx-target` a common ancestor and return the whole block in one fragment. No coordination, no ids to keep in sync. Reach for this first; only move on when the regions are too far apart to share a sensible parent.
2. **Out-of-band swaps.** Keep the primary target small, append extra fragments tagged `hx-swap-oob="true"` (matched by `id`) for the stragglers — counters, badges, toasts. One response, several disjoint updates. See `reference/patterns.md`.
3. **Event-driven (most decoupled).** Respond with `HX-Trigger: {"thing-updated":{}}`; unrelated elements listen via `hx-trigger="thing-updated from:body"` and re-fetch themselves. The handler never names them — cleanest for many independent consumers, but the indirection is harder to trace. Use when OOB coupling gets unwieldy.

## One handler, both full page and fragment — `hx-select` / `HX-Reselect`

Let a handler always render the **full page**, and have htmx pluck out the part it needs: `hx-get="/page" hx-select="#content" hx-target="#content"`. The same URL works for a normal navigation (no-JS / refresh / paste) and for an htmx swap — no `HX-Request` branching, no duplicate fragment template. Server-side override: `HX-Reselect: #content`. Trade-off: the server renders (and ships) the whole page each time, so skip it on hot paths where a dedicated fragment is cheaper.

## Detailed references (load when needed)

- **`reference/cheatsheet.md`** — every attribute, all HX request/response headers, all events, the full JS API (`htmx.ajax`, `htmx.process`, `htmx.on`, `htmx.swap`, ...), CSS classes, and every `htmx.config.*` option. Go here for exact names/signatures.
- **`reference/patterns.md`** — copy-paste recipes: active search, infinite scroll, lazy load, click-to-load, inline edit, delete row, polling, indicators, OOB swaps, `hx-select` (full-page→fragment), CSRF headers, auth headers, custom confirm, **debugging a silent interaction** (`htmx.logAll`, `monitorEvents`, `configRequest`/`beforeSwap`).
- **`reference/gotchas.md`** — attribute escaping (templ/Alpine), security (`selfRequestsOnly`, `hx-disable`, XSS), attribute inheritance, settling/transitions, common silent failures.

## Common mistakes

- **Returning JSON.** htmx swaps HTML. Return rendered HTML fragments. (JSON only via the client-side-templates extension.)
- **Forgetting the default trigger.** A `<form>` triggers on `submit`, not `click`; a non-input triggers on `click`. Set `hx-trigger` explicitly when unsure.
- **`hx-swap-oob` element not found.** OOB fragments swap by matching `id` already in the DOM; if no match, htmx logs `htmx:oobErrorNoTarget`. `<tr>`/`<td>` OOB fragments must be wrapped in `<template>`.
- **Escaped attributes via templ/Alpine** — silent failure, no console error. See `reference/gotchas.md`.
- **Inline `delay` confusion:** `delay` resets the timer on every event (debounce); `throttle` ignores events during the window. Active search wants `delay`.

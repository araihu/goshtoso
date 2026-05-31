# htmx Gotchas & Security

## Attribute escaping — templ + Alpine (this repo's #1 bug)

This repo (Goshtoso) generates htmx markup through **templ** and pairs it with **Alpine.js**. templ's `EscapeString` rewrites `"` → `&quot;`, `'` → `&#39;`, `&` → `&amp;` inside HTML attributes. This silently breaks any attribute that holds JS/JSON:

- **Alpine `x-data`** — Alpine swallows parse failures with no console error; the component just doesn't work.
- **htmx `hx-vals='{"k":"v"}'`, `hx-headers`, `hx-on:*`** — same failure mode if you build them with `json.Marshal` (double-quoted) and let templ escape them.

Symptoms: feature silently dead, unit tests pass, browser fails, no console error. **Check rendered HTML in devtools for `&quot;` inside the attribute.**

Fixes (from CLAUDE.md):
- `hx-vals` / `hx-headers`: build JSON with **single-quoted** JS string builders, never `json.Marshal` into an attribute. `hx-vals='{"k":"v"}'` works because templ leaves the inner double-quotes — but only if YOU write the literal, not marshal it.
- Complex JS: register via `<script>` + `templ.Raw()` + `Alpine.data()` / event listeners instead of inline attributes.
- Guard null arrays: `json.Marshal([]string(nil))` → `null`, and Alpine `.includes()` throws on null. Coerce `"null"` → `"[]"`.

## hx-swap-oob element not found

OOB fragments swap by matching an `id` **already present** in the DOM. No match → htmx logs `htmx:oobErrorNoTarget` and drops it. Also: the browser's HTML parser strips bare `<tr>`, `<td>`, `<option>`, `<li>` outside their parent — wrap OOB versions of these in `<template>`.

## Default trigger surprises

- `<form>` → `submit` (not click). A submit button inside it triggers the form, not itself.
- `<input>/<select>/<textarea>` → `change`. For keystroke search you must set `hx-trigger="input ..."` or `keyup`.
- Everything else → `click`.

## delay vs throttle

- `delay:Xms` = **debounce**: timer resets on every event; fires X after the last one. Use for search.
- `throttle:Xms` = fire immediately, then ignore events for X. Use for rate-limited live updates.

## htmx only processes content it swapped

If you set `innerHTML` / `fetch()` yourself, htmx won't wire up `hx-*` in that new content. Call `htmx.process(el)` after injecting. (Swaps htmx performs are processed automatically.)

## Polling never stops

`hx-trigger="every Ns"` polls forever until the element leaves the DOM or the server returns **HTTP 286**. Don't rely on the client to stop it.

## Boost gotchas

`hx-boost="true"` AJAX-ifies descendant `<a>`/`<form>`, swapping `<body>` by default. Only same-origin, non-`target=_blank`, GET-or-form requests are boosted. Inline `<script>` in boosted bodies re-runs per navigation; third-party widgets may double-init — re-init in an `htmx:load` handler. Anchors with `href="#..."`, `hx-*`, or `download` are left alone.

## Response must be 2xx/3xx to swap

By default non-2xx/3xx responses fire `htmx:responseError` and do **not** swap. To swap error bodies (e.g. validation HTML on 422), either use the `response-targets` extension (`hx-target-error`, `hx-target-422`) or tweak `htmx.config.responseHandling`.

## Title & history

A `<title>` anywhere in a swapped response updates `document.title` unless `hx-swap="… ignoreTitle:true"` or `htmx.config.ignoreTitle`. URLs pushed via `hx-push-url`/`hx-boost` must be **fully navigable** (refresh/paste must work) — the server has to render a full page for that URL.

## Security

- **Escape all user content.** htmx swaps raw HTML; unescaped user data = XSS. Rely on server-side template auto-escaping (templ escapes by default — don't bypass with `templ.Raw` on user data).
- `htmx.config.selfRequestsOnly` defaults to **`true`** in v2 — htmx blocks cross-origin requests. To call another host, allowlist it via the `htmx:validateUrl` event, not by globally disabling it:
  ```javascript
  document.body.addEventListener("htmx:validateUrl", e => {
    if (!e.detail.sameHost && e.detail.url.hostname !== "api.example.com") e.preventDefault();
  });
  ```
- `hx-disable` on a wrapper turns OFF all htmx processing for that subtree and **cannot** be re-enabled by injected descendants — wrap any region that renders untrusted HTML.
- Lock down eval-based features if you don't need them: `htmx.config.allowEval=false` (kills `js:` prefixes, event filters), `htmx.config.allowScriptTags=false` (don't run `<script>` in swaps). Set CSP nonces via `inlineScriptNonce`/`inlineStyleNonce`.
- Add a CSP header (`default-src 'self'`) as defense-in-depth.

## Events fire camelCase AND kebab-case

`htmx:afterSwap` also dispatches as `htmx:after-swap`. Alpine's `x-on:htmx:after-swap` works because of the kebab alias. Pick one consistently.

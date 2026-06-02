# htmx Patterns — copy-paste recipes

Server returns **HTML fragments**. Examples show the markup; a Go/templ handler returns the matching fragment.

## Active search (search-as-you-type)

```html
<input type="search" name="q"
       hx-get="/search"
       hx-trigger="input changed delay:500ms, search"
       hx-target="#results"
       hx-indicator="#spinner">
<span id="spinner" class="htmx-indicator">Searching…</span>
<div id="results"></div>
```

`delay` debounces; `changed` skips no-op keystrokes; the extra `search` event catches the clear-button. Server returns just the `#results` inner HTML.

## Infinite scroll

Last row carries the trigger; the response replaces it (`outerHTML`) with more rows + a new sentinel.

```html
<tr hx-get="/rows?page=2" hx-trigger="revealed" hx-swap="outerHTML">
  <td>last visible row…</td>
</tr>
```

## Lazy load

```html
<div hx-get="/graph" hx-trigger="load" hx-swap="outerHTML">
  <img class="htmx-indicator" src="/spinner.svg">
</div>
```

## Click to load (button → append)

```html
<tbody id="rows">
  <!-- rows … -->
  <tr id="load-more">
    <td colspan="3">
      <button hx-get="/rows?page=2" hx-target="#load-more" hx-swap="outerHTML">Load more</button>
    </td>
  </tr>
</tbody>
```

## Polling

```html
<div hx-get="/job/status" hx-trigger="every 2s" hx-swap="innerHTML">Pending…</div>
```

Stop server-side by responding **HTTP 286** (htmx halts polling on that status). Or load-poll once: `hx-trigger="load delay:1s"`.

## Delete a table row

```html
<tr>
  <td>Row data</td>
  <td><button hx-delete="/contacts/42"
              hx-target="closest tr"
              hx-swap="outerHTML swap:500ms"
              hx-confirm="Delete this contact?">Delete</button></td>
</tr>
```

`swap:500ms` lets a `tr.htmx-swapping{opacity:0;transition:.5s}` fade out first. Server returns empty body.

## Inline edit (click to edit)

```html
<!-- view fragment -->
<div hx-target="this" hx-swap="outerHTML">
  <p>Name: Joe</p>
  <button hx-get="/contact/1/edit">Edit</button>
</div>
```

`/contact/1/edit` returns a form fragment with `hx-put="/contact/1" hx-target="this" hx-swap="outerHTML"`; PUT returns the view fragment again.

## Out-of-band swaps (update multiple regions in one response)

Request targets `#main`; the response also updates an unrelated counter:

```html
<!-- response body -->
<div id="main">…primary swap (uses hx-target)…</div>
<span id="cart-count" hx-swap-oob="true">3</span>
```

`hx-swap-oob="true"` matches the existing element by `id` and swaps it `outerHTML`. Use `hx-swap-oob="innerHTML"` or `hx-swap-oob="beforeend:#log"` for other strategies. **Wrap OOB `<tr>`/`<td>`/`<option>` in `<template>`** so the browser doesn't strip them.

## Select a fragment from a full-page response (`hx-select`)

One handler renders the whole page; htmx swaps in only the part it asked for. The same URL serves a real navigation *and* an htmx swap — no `HX-Request` branching, one template.

```html
<a hx-get="/dashboard" hx-select="#content" hx-target="#content" hx-swap="outerHTML">Refresh</a>
```

Server can override per-response with `HX-Reselect: #content`. Cost: the full page is rendered/shipped each time — prefer a dedicated fragment on hot paths.

## Server-triggered client events

Response header:

```
HX-Trigger: {"showToast":{"level":"success","msg":"Saved"}}
```

```javascript
document.body.addEventListener("showToast", e => toast(e.detail.level, e.detail.msg));
```

Or listen for an event to drive a request: `hx-trigger="showToast from:body"`.

## Indicators

```css
/* default: htmx injects .htmx-indicator{opacity:0;transition:opacity .2s} and
   .htmx-request .htmx-indicator{opacity:1}. Override if you prefer display: */
.htmx-indicator { display: none; }
.htmx-request .htmx-indicator { display: inline-block; }
.htmx-request.htmx-indicator { display: inline-block; } /* when the indicator IS the requesting elt */
```

## CSRF / auth headers globally

```html
<body hx-headers='{"X-CSRF-Token": "TOKEN_HERE"}'>
```

Or dynamically, for bearer tokens:

```javascript
document.body.addEventListener("htmx:configRequest", e => {
  e.detail.headers["Authorization"] = "Bearer " + getToken();
});
```

## Custom confirm dialog (async)

```javascript
document.addEventListener("htmx:confirm", function(e) {
  if (!e.target.hasAttribute("hx-confirm")) return;
  e.preventDefault();
  myAsyncConfirm(e.detail.question).then(ok => { if (ok) e.detail.issueRequest(true); });
});
```

## Abort an in-flight request

```html
<button id="job" hx-post="/start">Start</button>
<button onclick="htmx.trigger('#job','htmx:abort')">Cancel</button>
```

Or coordinate automatically: `hx-sync="closest form:abort"` on a field cancels its request when the form submits. Other strategies: `:drop` (default — ignore new while one runs), `:replace`, `:queue first|last|all`.

## Programmatic request after you inject HTML

```javascript
const html = await (await fetch("/frag")).text();
container.innerHTML = html;
htmx.process(container);   // REQUIRED: htmx only scans content it swapped itself
```

## Debugging an htmx interaction (when a request silently does nothing)

Work outward from "did it even fire":

```javascript
htmx.logAll();                          // log every htmx event to the console
monitorEvents(document.getElementById("thing"));  // Chrome DevTools: every DOM event on an element
```

- **No request fired?** Check the *default trigger* (a `<form>` fires on `submit`, an `<input>` on `change`, everything else on `click`) and confirm the element was actually processed by htmx (it must have been swapped in by htmx or passed to `htmx.process` — see below).
- **Request fired, wrong params/headers?** Listen on `htmx:configRequest` and inspect/patch `e.detail.parameters`, `e.detail.headers`, `e.detail.target`, `e.detail.verb` before it sends.
- **Response came back but didn't swap?** Non-2xx/3xx won't swap by default (see gotchas). Inspect/override the swap in `htmx:beforeSwap` — `e.detail.shouldSwap`, `e.detail.target`, `e.detail.serverResponse`:
  ```javascript
  document.body.addEventListener("htmx:beforeSwap", e => {
    if (e.detail.xhr.status === 422) { e.detail.shouldSwap = true; e.detail.isError = false; }
  });
  ```
- **Attribute looks right in source but dead in browser?** Inspect the *rendered* DOM in devtools for `&quot;`/`&#39;` — templ escaped it. See `gotchas.md`.

## Server: detect htmx vs full-page request (Go)

```go
func handler(w http.ResponseWriter, r *http.Request) {
    if r.Header.Get("HX-Request") == "true" {
        renderFragment(w)   // return just the partial
        return
    }
    renderFullPage(w)       // first load / refresh / no-JS
}
```

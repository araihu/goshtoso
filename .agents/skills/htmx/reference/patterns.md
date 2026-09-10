# htmx 4 patterns

Sources: [official patterns](https://four.htmx.org/patterns), [versioned guidance](https://raw.githubusercontent.com/bigskysoftware/htmx/v4.0.0/dist/skills/htmx-guidance.md).

## Search and validation

```html
<input name="search" hx-get="/search" hx-trigger="input changed delay:300ms"
       hx-target="#results" hx-indicator="#busy">
<span id="busy" class="htmx-indicator">Searching…</span>
<div id="results"></div>

<form hx-post="/save" hx-target="#result"
      hx-status:422="target:#errors" hx-status:5xx="swap:none">
  <input name="name" required>
  <button hx-disable="this">Save</button>
</form>
<div id="errors"></div>
<div id="result"></div>
```

Search returns result HTML. Save returns success HTML or a 422 validation fragment. Choose explicitly whether server-error bodies should replace UI.

## Multiple regions

Prefer a common ancestor target when regions naturally belong together. For disjoint updates, existing OOB markup remains available; htmx 4 also supports explicit partials:

```html
<hx-partial hx-target="#messages" hx-swap="beforeend">
  <div>New message</div>
</hx-partial>
<hx-partial hx-target="#count"><span>5</span></hx-partial>
```

Main content swaps before OOB/partials. A response containing only those updates leaves the main target alone by default. Use `swapEmpty:true` only when clearing it is intended.

For independently refreshed consumers, return `HX-Trigger: items-updated`; listeners can use `hx-trigger="items-updated from:body"` to fetch their own fragments.

## Navigation and state

`hx-select="#content" hx-target="#content"` can select a fragment from a full document. Ensure the server returns that document for full requests and supports direct navigation, refresh, and history. Read `HX-Request-Type` when distinguishing full and partial responses.

Use `innerMorph`/`outerMorph` with `hx-alpine-compat` when Alpine state and input state should survive. Use `innerHTML`/`outerHTML` when replacing or resetting the component is the intended behavior.

## Debugging

Enable `htmx.config.logAll = true` temporarily. Check the network request's form values, headers, status, response HTML and destination. Inspect live `getAttribute()` values for templ expressions. Listen to `htmx:error` and `htmx:response:error`; inspect event-specific `detail` rather than a presumed XHR object. For lifecycle checks, wait through `htmx:after:settle` and Alpine's DOM updates.

For lazy loading use `hx-trigger="load"`; for viewport loading use `revealed` or `intersect`; for polling use `every 2s`. Keep inserted content valid for its HTML context (especially table rows).

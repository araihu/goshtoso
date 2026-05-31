# htmx Cheatsheet — full reference

htmx v2.0.x. Source: htmx.org/reference. Every attribute prefix also works as `data-hx-*`.

## Core attributes

| Attribute | Description |
|-----------|-------------|
| `hx-get` | issue a `GET` to the URL |
| `hx-post` | issue a `POST` to the URL |
| `hx-put` | issue a `PUT` to the URL |
| `hx-patch` | issue a `PATCH` to the URL |
| `hx-delete` | issue a `DELETE` to the URL |
| `hx-on:<event>` | handle any event with an inline script (e.g. `hx-on:click`, `hx-on::before-request`) |
| `hx-push-url` | push URL into browser location bar / history |
| `hx-select` | select a subset of the response to swap in (CSS selector) |
| `hx-select-oob` | select content from the response to swap out of band |
| `hx-swap` | how content is swapped in (see below) |
| `hx-swap-oob` | mark this element to be swapped out of band by id |
| `hx-target` | which element to swap the response into |
| `hx-trigger` | which event triggers the request (see below) |
| `hx-vals` | extra values to submit, JSON (`hx-vals='{"a":1}'`; prefix `js:` to eval) |

## Additional attributes

| Attribute | Description |
|-----------|-------------|
| `hx-boost` | progressively enhance anchors/forms into AJAX requests |
| `hx-confirm` | `confirm()` dialog text before the request |
| `hx-disable` | disable htmx processing for this node + children (cannot be re-enabled by descendants) |
| `hx-disabled-elt` | add `disabled` attr to matched elements during the request |
| `hx-disinherit` | stop named attributes from being inherited by children |
| `hx-encoding` | request encoding; set `multipart/form-data` for file upload |
| `hx-ext` | enable htmx extension(s) on this element/subtree |
| `hx-headers` | extra request headers, JSON (prefix `js:` to eval) |
| `hx-history` | `false` to omit this page from the history snapshot (sensitive data) |
| `hx-history-elt` | element to snapshot/restore on history navigation |
| `hx-include` | include values of additional elements (CSS selector) |
| `hx-indicator` | element(s) to receive `htmx-request` class during the request |
| `hx-inherit` | re-enable inheritance of named attributes when globally disabled |
| `hx-params` | filter params: `*` / `none` / `not a,b` / `a,b` |
| `hx-preserve` | keep this element unchanged across swaps (matched by id) |
| `hx-prompt` | `prompt()` text; answer sent as `HX-Prompt` header |
| `hx-replace-url` | replace (not push) the URL in the location bar |
| `hx-request` | configure request: `timeout`, `credentials`, `noHeaders` (JSON) |
| `hx-sync` | coordinate requests across elements (see strategies) |
| `hx-validate` | force HTML5 validation before request (for non-form elements) |
| `hx-vars` | **deprecated** — use `hx-vals` |

## hx-swap

Strategies: `innerHTML` (default), `outerHTML`, `beforebegin`, `afterbegin`, `beforeend`, `afterend`, `delete`, `none`.

Modifiers (append after the strategy):

| Modifier | Effect |
|----------|--------|
| `transition:true` | wrap the swap in the View Transitions API |
| `swap:<time>` | delay between receiving content and swapping (e.g. `swap:200ms`) |
| `settle:<time>` | delay between swap and settle (default 20ms) |
| `ignoreTitle:true` | don't update document title from a `<title>` in the response |
| `scroll:top` / `scroll:bottom` | scroll target to top/bottom after swap (can target `scroll:#id:top`) |
| `show:top` / `show:bottom` | scroll so target's top/bottom is visible (can target `show:#id:top`, `show:window:top`) |
| `focus-scroll:true|false` | scroll the focused element into view after swap |

## hx-trigger

Syntax: `event[ modifier ...][, event2 ...]`. An event may be filtered: `hx-trigger="click[ctrlKey]"`.

| Modifier | Effect |
|----------|--------|
| `once` | fire at most one time |
| `changed` | only fire if the element's value changed since last event |
| `delay:<time>` | debounce — wait, resetting on each new event |
| `throttle:<time>` | rate-limit — fire then ignore events during the window |
| `from:<selector>` | listen for the event on a different element (e.g. `from:body`, `from:closest form`) |
| `target:<selector>` | only fire if the event originated from a matching element |
| `consume` | stop the event bubbling to parents |
| `queue:<first|last|all|none>` | how to queue events that arrive during a request (default `last`) |

Special triggers:

| Trigger | Fires |
|---------|-------|
| `load` | once when the element is loaded into the DOM |
| `revealed` | when the element first scrolls into the viewport |
| `intersect` | on intersection; options `root:<sel>` and `threshold:<0..1>` |
| `every <time>` | polling, e.g. `every 2s` (server returns HTTP `286` to stop polling) |

## CSS classes (applied automatically)

| Class | When |
|-------|------|
| `htmx-request` | on the element (or `hx-indicator` target) for the duration of the request |
| `htmx-indicator` | base style for indicators; `opacity:0` until an ancestor has `htmx-request` |
| `htmx-added` | on new content just before swap, removed after settle |
| `htmx-swapping` | on the target during the swap phase |
| `htmx-settling` | on the target during the settle phase (enables CSS transitions) |

## Request headers (htmx → server)

| Header | Value |
|--------|-------|
| `HX-Request` | always `true` |
| `HX-Boosted` | `true` when via `hx-boost` |
| `HX-Current-URL` | the browser's current URL |
| `HX-History-Restore-Request` | `true` when restoring from history (cache miss) |
| `HX-Prompt` | the user's `hx-prompt` answer |
| `HX-Target` | id of the target element (if any) |
| `HX-Trigger` | id of the triggering element (if any) |
| `HX-Trigger-Name` | name of the triggering element (if any) |

## Response headers (server → htmx)

| Header | Effect |
|--------|--------|
| `HX-Location` | client-side redirect (htmx AJAX nav, no full page reload) — value may be JSON for target/swap |
| `HX-Push-Url` | push this URL into history (or `false` to disable) |
| `HX-Redirect` | client-side redirect to a new location (full `window.location`) |
| `HX-Refresh` | `true` → full page refresh |
| `HX-Replace-Url` | replace the current URL in the location bar |
| `HX-Reswap` | override the `hx-swap` value for this response |
| `HX-Retarget` | CSS selector overriding the target for this response |
| `HX-Reselect` | CSS selector choosing which part of the response to swap |
| `HX-Trigger` | trigger client events (string event name, or JSON `{"evt":detail}`) |
| `HX-Trigger-After-Settle` | like `HX-Trigger` but after the settle step |
| `HX-Trigger-After-Swap` | like `HX-Trigger` but after the swap step |

## Events

Every event fires in both `htmx:camelCase` and `htmx:kebab-case` form (kebab aids Alpine.js).

Lifecycle: `htmx:confirm`, `htmx:configRequest`, `htmx:beforeRequest`, `htmx:beforeSend`, `htmx:xhr:loadstart`, `htmx:xhr:progress`, `htmx:xhr:loadend`, `htmx:beforeOnLoad`, `htmx:afterOnLoad`, `htmx:beforeSwap`, `htmx:afterSwap`, `htmx:afterSettle`, `htmx:afterRequest`, `htmx:load`.

Node processing: `htmx:beforeProcessNode`, `htmx:afterProcessNode`, `htmx:beforeCleanupElement`.

Errors: `htmx:responseError` (non-2xx/3xx), `htmx:sendError` (network), `htmx:timeout`, `htmx:swapError`, `htmx:targetError`, `htmx:onLoadError`, `htmx:abort`, `htmx:sendAbort`, `htmx:xhr:abort`.

OOB: `htmx:oobBeforeSwap`, `htmx:oobAfterSwap`, `htmx:oobErrorNoTarget`.

History: `htmx:beforeHistorySave`, `htmx:pushedIntoHistory`, `htmx:replacedInHistory`, `htmx:historyRestore`, `htmx:historyCacheHit`, `htmx:historyCacheMiss`, `htmx:historyCacheMissLoad`, `htmx:historyCacheMissLoadError`, `htmx:historyCacheError`.

Validation: `htmx:validation:validate`, `htmx:validation:failed`, `htmx:validation:halted`.

SSE/WS (extensions): `htmx:sseOpen`, `htmx:sseError`, `htmx:noSSESourceError`.

Transitions: `htmx:beforeTransition`. Prompt: `htmx:prompt`.

## JavaScript API

| Method | Signature / note |
|--------|------------------|
| `htmx.ajax(verb, path, ctx)` | issue an htmx request programmatically; `ctx` = target selector or `{target, swap, values, headers}` |
| `htmx.find(selector)` / `htmx.find(elt, selector)` | first matching element |
| `htmx.findAll(selector)` / `htmx.findAll(elt, selector)` | all matching elements |
| `htmx.closest(elt, selector)` | nearest matching ancestor |
| `htmx.on(target, event, handler)` / `htmx.on(event, handler)` | add listener, returns it |
| `htmx.off(...)` | remove listener |
| `htmx.onLoad(handler)` | shortcut for `htmx:load` (handler gets the new element) |
| `htmx.trigger(elt, eventName, detail)` | dispatch an event |
| `htmx.process(elt)` | scan an element for htmx attributes (use after injecting HTML yourself) |
| `htmx.swap(target, content, swapSpec)` | perform a swap + settle programmatically |
| `htmx.values(elt, verb?)` | collect the input values an element would submit |
| `htmx.addClass/removeClass/toggleClass(elt, class)` | class helpers |
| `htmx.takeClass(elt, class)` | add class to elt, remove from siblings |
| `htmx.remove(elt, delay?)` | remove an element |
| `htmx.defineExtension(name, ext)` / `htmx.removeExtension(name)` | extension registration |
| `htmx.parseInterval("1s")` | parse to milliseconds |
| `htmx.logAll()` | log every htmx event (debugging) |
| `htmx.logger` | settable logger function |
| `htmx.config` | live config object (see below) |
| `htmx.createEventSource` / `htmx.createWebSocket` | overridable factories (auth, etc.) |

## htmx.config.* options

Set via `htmx.config.x = ...` or `<meta name="htmx-config" content='{"x":...}'>`.

| Option | Default | Meaning |
|--------|---------|---------|
| `defaultSwapStyle` | `innerHTML` | swap when `hx-swap` omitted |
| `defaultSwapDelay` | `0` | ms before swap |
| `defaultSettleDelay` | `20` | ms before settle |
| `timeout` | `0` | request timeout (ms) |
| `historyEnabled` | `true` | enable history |
| `historyCacheSize` | `10` | snapshots cached (set `0` to disable caching) |
| `refreshOnHistoryMiss` | `false` | full reload instead of AJAX on history miss |
| `includeIndicatorStyles` | `true` | inject default `.htmx-indicator` CSS |
| `indicatorClass` / `requestClass` / `addedClass` / `settlingClass` / `swappingClass` | `htmx-*` | class names |
| `allowEval` | `true` | allow eval-based features (`js:`, event filters) |
| `allowScriptTags` | `true` | execute `<script>` in swapped content |
| `inlineScriptNonce` / `inlineStyleNonce` | `''` | CSP nonce |
| `attributesToSettle` | `["class","style","width","height"]` | attrs animated during settle |
| `withCredentials` | `false` | send cookies/auth on cross-site requests |
| `selfRequestsOnly` | `true` | only allow same-origin requests (security) |
| `scrollBehavior` | `instant` | `instant`/`smooth`/`auto` for boost scroll |
| `defaultFocusScroll` | `false` | scroll focused elt into view after swap |
| `getCacheBusterParam` | `false` | append cache-buster to GETs |
| `globalViewTransitions` | `false` | use View Transitions for all swaps |
| `methodsThatUseUrlParams` | `["get","delete"]` | encode params in URL not body |
| `ignoreTitle` | `false` | never update `document.title` |
| `scrollIntoViewOnBoost` | `true` | scroll boosted target into view |
| `allowNestedOobSwaps` | `true` | process OOB in nested elements |
| `responseHandling` | (array) | rules mapping status codes to swap/error behavior |
| `wsReconnectDelay` | `full-jitter` | WS reconnect strategy |
| `wsBinaryType` | `blob` | WS binary type |

## Extensions (htmx team)

Enable with `hx-ext="name"`. Core ones: `head-support`, `idiomorph` (morph swap), `preload`, `response-targets` (`hx-target-404`, `hx-target-5xx`...), `sse`, `ws`, `htmx-1-compat`. Install via CDN script or npm package `htmx-ext-<name>`.

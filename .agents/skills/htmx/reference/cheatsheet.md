# htmx 4 quick reference

Baseline: 4.0.0, checked 2026-09-10. Sources: [reference](https://four.htmx.org/reference), [versioned upstream guidance](https://raw.githubusercontent.com/bigskysoftware/htmx/v4.0.0/dist/skills/htmx-guidance.md), [events guide](https://four.htmx.org/docs/htmx-events-guide).

## Attributes

| Need | Attribute |
|---|---|
| Request | `hx-get`, `hx-post`, `hx-put`, `hx-patch`, `hx-delete`, `hx-query` |
| Event / destination / response selection | `hx-trigger`, `hx-target`, `hx-select` |
| Shared ancestor settings | `hx-target:inherited`, `hx-headers:inherited`, etc. |
| Extra request data | `hx-vals`, `hx-include` |
| Request headers / fetch configuration | `hx-headers`, `hx-config` |
| Link/form enhancement for descendants | `hx-boost:inherited="true"` |
| Disable controls during request | `hx-disable="this"` |
| Exclude subtree from htmx | `hx-ignore` |
| Status-specific behavior | `hx-status:422`, `hx-status:5xx` |
| Loading / confirmation / synchronization | `hx-indicator`, `hx-confirm`, `hx-sync` |
| URL history | `hx-push-url`, `hx-replace-url` |
| Multiple response regions | `hx-swap-oob` or `<hx-partial>` |

`hx-*` and `data-hx-*` both work by default. `hx-vals`, `hx-headers`, and `hx-config` accept HCON (including JSON); dynamic values use `js:` where supported.

Swaps: `innerHTML`, `outerHTML`, `outerSync`, `innerMorph`, `outerMorph`, `textContent`, `beforebegin`/`before`, `afterbegin`/`prepend`, `beforeend`/`append`, `afterend`/`after`, `delete`, `none`.

Selected modifiers: `swap:100ms`, `settle:50ms`, `transition:true`, `ignoreTitle:true`, `scroll:top`, `show:bottom`, `focusScroll:true`, `strip:true`, `swapEmpty:true`.

## Headers

Requests include `HX-Request: true`, `HX-Current-URL`, `HX-Source` and `HX-Target` (the latter two use `tag#id`), and `HX-Request-Type` (`full` or `partial`). Boosted/history requests have their corresponding headers. Do not equate every htmx request with a fragment request: body targets and `hx-select` can require full documents.

Responses support `HX-Trigger`, `HX-Redirect`, `HX-Location`, `HX-Refresh`, `HX-Retarget`, `HX-Reswap`, `HX-Reselect`, `HX-Push-Url`, `HX-Replace-Url`. `HX-Trigger-After-Swap` and `HX-Trigger-After-Settle` are removed. Request `HX-Trigger-Name` is also removed; response `HX-Trigger` remains supported.

## Events and JavaScript

```js
document.addEventListener('htmx:config:request', event => {
  const request = event.detail.ctx.request
  request.headers['X-Example'] = 'value'
  request.body.set('extra', 'value') // FormData
})
htmx.ajax('GET', '/items', { target: '#result' }) // Promise
htmx.process(element) // manually inserted subtree
htmx.config.logAll = true // development debugging
```

Lifecycle: `htmx:before:init`, `htmx:after:init`, `htmx:before:process`, `htmx:after:process`, `htmx:before:cleanup`, `htmx:after:cleanup`.
Requests: `htmx:config:request`, `htmx:before:request`, `htmx:before:response`, `htmx:after:request`, `htmx:finally:request`.
Swaps: `htmx:before:swap`, `htmx:after:swap`, `htmx:finally:swap`, `htmx:before:settle`, `htmx:after:settle`.
Errors: `htmx:error`, `htmx:response:error`. Use the event-specific detail schema; not every lifecycle event carries request context. Morph-node hooks are extension-only, not DOM events.

`htmx.swap(ctx)` takes a context object, unlike the 2.x multi-argument signature. `htmx.registerExtension(name, hooks)` replaces `defineExtension`; use native DOM methods for removed class/removal/listener helpers. Consult the official API before converting call arguments.

## Configuration

| Setting | Default / purpose |
|---|---|
| `defaultSwap` | `innerHTML` |
| `defaultTimeout` | 60000 ms |
| `defaultSettleDelay` | 1 ms |
| `implicitInheritance` | false |
| `noSwap` | `[204, 304]` |
| `allowEmptySwapAfterOOB` | false |
| `mode` | `same-origin` |
| `history` | true |
| `transitions` | false |
| `includeIndicatorCSS` | true |
| `extensions` | Empty permits all; otherwise registration-name allowlist |

Do not reuse 2.x `responseHandling`, `allowEval`, `allowScriptTags`, or `selfRequestsOnly` settings; consult the migration guide for removed configuration.

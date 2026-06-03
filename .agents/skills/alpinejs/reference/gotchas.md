# Alpine.js Gotchas (Goshtoso: Alpine + templ + htmx)

The failure mode that wastes the most time: **Alpine swallows expression parse failures with no console error.** A component just silently does nothing. Always inspect the *rendered* HTML in devtools before debugging logic.

## 1. templ attribute escaping — the #1 bug

templ's `EscapeString` converts `"` → `&quot;`, `'` → `&#39;`, `&` → `&amp;` inside HTML attributes. That breaks the JS Alpine parses out of `x-data`/`x-bind`/`@`.

**Symptom:** dropdown/combobox options missing, toggles dead — no console error, unit tests pass, browser fails. Look for `&quot;` inside `x-data` in devtools.

**Simple `x-data` (data only):** use unquoted keys, avoid quoted string literals.
```go
return fmt.Sprintf(`{ opened: [false,false], count: 0 }`) // GOOD — nothing to escape
```

**Complex Alpine (functions/strings):** register via `<script>` + `Alpine.data()`, reference by name.
```go
func myAlpineScript(cfg Config) string {
    return fmt.Sprintf(`document.addEventListener('alpine:init', () => {
        Alpine.data('myComponent', () => ({
            value: '%s',
            doThing() { htmx.ajax('GET', '/api/data', {target: '#target'}); }
        }));
    });`, cfg.DefaultValue)
}
```
```templ
templ myScript(cfg Config) {
    @templ.Raw("<script>" + myAlpineScript(cfg) + "</script>")
}
// then: <div x-data="myComponent">
```

## 2. NEVER json.Marshal into an HTML attribute

`json.Marshal` emits double-quoted strings → templ escapes the quotes → Alpine sees broken syntax. Build single-quoted JS instead.
```go
func optionsToJS(options []Option) string {
    result := "["
    for i, opt := range options {
        if i > 0 { result += "," }
        result += fmt.Sprintf("{value:'%s',label:'%s'}",
            jsEscapeSingle(opt.Value), jsEscapeSingle(opt.Label))
    }
    return result + "]"
}
```

## 3. null arrays crash Alpine

`json.Marshal([]string(nil))` → `null`, not `[]`. Alpine `selectedValues.includes(...)` throws on null. Guard:
```go
if string(selectedJSON) == "null" {
    selectedJSON = []byte("[]")
}
```

## 4. Fragment-nav: register Alpine.data immediately, not only on alpine:init

Alpine is already running when a page/fragment arrives via an htmx (sidebar) swap. A component registered ONLY inside an `alpine:init` listener is **undefined** for the swapped-in node. Register immediately if Alpine is already up:
```js
function register() { Alpine.data('logFeed', () => ({ /* … */ })) }
if (window.Alpine) register()                                  // fragment-nav path
else document.addEventListener('alpine:init', register)        // first paint
```

## 5. htmx-inserted nodes aren't auto-processed by Alpine

When Alpine (or your own JS) inserts nodes, htmx won't bind them automatically — call `htmx.process(el)` on the new subtree. (Conversely, htmx-swapped HTML containing Alpine directives is initialized by Alpine's mutation observer.)

## 6. hx-swap-oob on first paint → htmx:oobErrorNoTarget

An element carrying `hx-swap-oob` on its *first* render makes htmx attempt an OOB swap (with no target) when it arrives via fragment nav. Gate the attribute to update-only (an `oob bool` that's false on first paint).

## 7. x-cloak FOUC

Initially-hidden (`x-show="false"`) elements flash before Alpine boots unless they carry `x-cloak` AND the page has `[x-cloak]{display:none!important}`.

## 8. E2E: Alpine state in Playwright

- `GetAttribute("aria-expanded")` returns the *static* HTML attribute, not the Alpine-bound live value. Use `Evaluate("el => el.getAttribute('aria-expanded')", nil)`.
- Wait for Alpine: `WaitForFunction("() => typeof Alpine !== 'undefined'")`.
- `Locator.Fill()` does NOT fire a native `input` event → `x-model` won't update. Dispatch `input` manually (the `fillSearchInput` helper).

## 9. $dispatch reaches ancestors, not siblings

Custom events bubble up. To catch a sibling component's event, listen at window: `@my-event.window="..."`.

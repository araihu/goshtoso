# agents.md - Process & Learnings for AI Agents

This file documents the development process, known pitfalls, and lessons learned
while building Goshtoso PenguinUI, intended to help AI agents (and humans) avoid
repeating past mistakes.

## Development Process

### Component Porting Workflow

1. **Study the original** — Open the PenguinUI reference page for the component.
   Inspect the HTML structure, CSS classes, Alpine.js behavior, and responsive design.
2. **Create types.go** — Define `Config` struct with all variants, options, and
   CSS class builder methods.
3. **Create the .templ file** — Translate the HTML into templ syntax. Pay special
   attention to Alpine.js `x-data` attributes (see pitfalls below).
4. **Create demo page** — Add a demo in `internal/pages/demo/components/` that shows
   all variants side by side.
5. **Register route + sidebar** — Add handler in `server.go`, add item to
   `getSidebarItems()` in `layout.templ`.
6. **Generate & test** — `make generate && make dev`, visually compare with original.
7. **Write E2E tests** — Cover interactive behavior, not just rendering.

### Build Pipeline

```
.templ file edited → make generate → _templ.go created → make css → go build
```

Always run `make generate` after editing `.templ` files. The `*_templ.go` files are
generated and must never be edited by hand.

## Known Pitfalls & Lessons Learned

### 1. Templ HTML Escaping Breaks Alpine.js (Combobox Bug #1)

**Severity:** Critical — components silently fail to render data

**Problem:** `json.Marshal()` produces double-quoted strings. When these are placed
inside templ HTML attributes (like `x-data`), templ's `EscapeString` converts `"` to
`&quot;`. Alpine.js then receives malformed JS object literals and fails silently.

**Symptoms:**
- Dropdown/combobox options don't appear
- No console errors (Alpine silently swallows parse failures)
- Works in unit tests but fails in browser

**Root cause:**
```go
// json.Marshal produces: [{"value":"tech","label":"Tech"}]
// templ escapes to:      [{&quot;value&quot;:&quot;tech&quot;,...}]
// Alpine.js sees:         broken syntax
optsJSON, _ := json.Marshal(cfg.Options)
```

**Fix:** Build JS literals using single-quoted strings:
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

**Rule:** Never use `json.Marshal` for data that ends up inside HTML attributes
via templ. Always use single-quoted JS string builders.

### 2. Null Arrays Crash Alpine.js (Combobox Bug #2)

**Severity:** High — runtime errors on empty selections

**Problem:** Go's `json.Marshal([]string(nil))` produces `null`, not `[]`.
Alpine.js code like `selectedValues.includes(...)` throws on null.

**Fix:** Always guard against null arrays:
```go
if string(selectedJSON) == "null" {
    selectedJSON = []byte("[]")
}
```

**Rule:** Never pass null to Alpine.js where an array is expected. Always
default to `[]`.

### 3. Layout: Avoid Duplicate Headers/Branding

**Problem:** The sidebar component has its own logo section. When used inside a
layout that already has a header with branding, this creates a "two disjointed
containers" appearance — the sidebar looks like a separate app from the header.

**Fix:** The sidebar logo section is conditional — only renders when `Logo` or
`LogoText` is set. In the demo layout, omit `LogoText` since the page header
already shows "Goshtoso PenguinUI".

Also: don't duplicate positioning CSS. The layout wrapper handles `fixed`/`static`
responsive positioning; the sidebar component itself should only handle its own
styling (borders, background, flex layout).

### 4. Mobile Sidebar Positioning

**Problem:** Using `fixed inset-y-0` on the mobile sidebar makes it span the full
viewport height, overlapping the sticky header.

**Fix:** Use `fixed top-16 bottom-0` so the sidebar starts below the 4rem header.

## Debugging Tips

### Alpine.js Issues
- Check rendered HTML in browser devtools — look for `&quot;` inside `x-data`
- Alpine fails silently on malformed `x-data` — the component just won't work
- Test with `console.log` inside `x-init` to verify data parsing

### Templ Issues
- Always check the generated `_templ.go` file if behavior is unexpected
- Templ escapes all string values in attributes — plan for this
- Use `templ.Attributes` for dynamic attribute maps

### Visual Parity
- Compare in both light and dark modes
- Test with multiple themes (especially Minimal which has no border-radius)
- Check responsive breakpoints — mobile sidebar behavior differs significantly

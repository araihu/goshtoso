# AGENT.md - Goshtoso PenguinUI

## Project Overview

**Goshtoso** (Go + Alpine.js + Tailwind CSS + HTMX + Templ) is a UI component library that replicates [PenguinUI](https://penguinui.com) components using Go's templating system.

## Completed Components (22)

Accordion, Alert, Avatar, Badge, Banner, Button, Card, Checkbox, Combobox, Dropdown, Modal, Pagination, Select, Sidebar, Spinner, Table, Tabs, Textarea, Text Input, Toast, Toggle, Tooltip

## Critical Gotchas

### 1. Templ HTML escaping breaks Alpine.js
Templ escapes `"` → `&quot;`, `'` → `&#39;`, `&` → `&amp;` inside HTML attributes. This silently breaks Alpine.js `x-data` objects.

**Solution for complex Alpine data**: Use `templ.Raw()` with a `<script>` tag that registers an `Alpine.data()` component. Then reference it with `x-data="componentName"`.

```go
// types.go — generates unescaped JS inside <script>
func myScript(cfg Config) string {
    return `document.addEventListener('alpine:init', () => {
        Alpine.data('myComp', () => ({ count: 0 }));
    });`
}

// component.templ — render via templ.Raw
templ scriptTag(cfg Config) {
    @templ.Raw("<script>" + myScript(cfg) + "</script>")
}
// then: <div x-data="myComp">
```

**For simple data** (no functions, no strings with quotes): inline x-data with unquoted keys works:
```go
`{ opened: [false,false], count: 0 }`
```

### 2. Templ generate sometimes reports "0 updates"
Force regeneration: `rm components/<name>/<name>_templ.go && templ generate`

### 3. Alpine.js attributes need JS Evaluate in tests
`GetAttribute("aria-expanded")` returns the static HTML attribute, NOT the Alpine.js-bound live value. Use:
```go
expanded, _ := locator.Evaluate("el => el.getAttribute('aria-expanded')", nil)
```

### 4. Port 8090 is reserved for manual dev
E2E tests use a random free port via `freePort()` in `e2e_test.go`. Never hardcode 8090 in tests.

### 5. Generated files in merge conflicts
Never try to manually resolve conflicts in `*_templ.go` or `assets/styles.css`. Resolve the `.templ` source, then `templ generate`.

## E2E Test Architecture

- **Single shared browser** via `TestMain` (1 Chromium launch, not per-test)
- **`newPage(t, browser)`** creates a tab with 2s timeout + auto-cleanup
- **`setupServer(t)` / `setupPlaywright(t)`** are no-ops (TestMain handles everything)
- **TestMain** builds server binary, picks random port, starts server, launches Playwright

### Running tests
```bash
go test ./tests/e2e/... -count=1 -timeout 15m       # full suite (~2.5min)
go test ./tests/e2e/... -count=1 -run TestSidebar    # specific test
```

## Component Development Pattern

Each component is `components/<name>/` with:
- `types.go` — Config struct, variant constants, CSS class methods
- `<name>.templ` — Template with public entry point + private helpers
- Demo page: `internal/pages/demo/components/<name>.templ`
- Route: case in `internal/server/server.go:handleComponent()`
- Sidebar entry: `internal/pages/demo/layout.templ:getSidebarItems()`

## Table Component (Most Complex)

Features: sorting (neutral→asc→desc→neutral cycle), pagination with OOB swap, infinite scroll, lazy load, filter bar (in progress).

HTMX endpoint: `/api/components/table/rows`
Query params: `order_by`, `order_dir`, `page`, `per_page`, `search`, `membership`, `variant`

### In-Progress: Filter Bar
Types defined (`FilterSearch`, `FilterSelect`, `FilterToggle`), templ rendering works, Alpine.js integration via `<script>` + `Alpine.data()`. Blocker: Alpine.js CDN scripts load via `defer` causing timing issues in headless Chromium E2E tests. See `SESSION_STATE.md` for details.

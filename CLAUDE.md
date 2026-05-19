# CLAUDE.md — Goshtoso

This is the single source of truth for AI agents (Claude, Copilot, Codex, etc.)
working in this repo. `AGENTS.md` is a symlink to this file.

## Project Overview

**Goshtoso** (Go + Alpine.js + Tailwind CSS + HTMX + Templ) is a UI component
library that replicates [PenguinUI](https://penguinui.com) components using Go's
templating system. Hard fork of PenguinUI targeting 99.99% visual parity.

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.26+ | Backend server and templating |
| templ | v0.3.x | HTML template generation (.templ → .go) |
| Tailwind CSS | v4 | Utility-first styling |
| HTMX | v2.0.8 | Dynamic content loading |
| Alpine.js | v3.x | Reactive UI components |
| Playwright | v0.5700.1 | E2E testing |

Alpine.js + HTMX are bundled locally under `assets/js/vendor/` (no CDN at
runtime). Page loads are deterministic in E2E.

## Quick Commands

```bash
# Generate templ files (REQUIRED after editing .templ files)
templ generate
# or: just gp-generate

# Build Tailwind CSS (REQUIRED after editing CSS)
tailwindcss -i css/main.css -o assets/styles.css

# Run dev server (default port 8090)
go run cmd/server/main.go
# or: just gp-dev

# Run E2E tests (full suite ~2.5 min)
go test ./tests/e2e/... -count=1 -timeout 15m

# Run specific E2E test
go test ./tests/e2e/... -count=1 -timeout 5m -run TestDropdown

# Build server binary
go build -o bin/server ./cmd/server
```

## Repository Structure

```
goshtoso/
├── cmd/server/main.go          # Server entry point
├── components/                 # Reusable UI components (22 total)
│   └── <name>/
│       ├── <name>.templ        # Component template
│       ├── types.go            # Config types and variant classes
│       └── <name>_templ.go     # Generated (DO NOT EDIT)
├── internal/
│   ├── server/
│   │   ├── server.go           # Route handlers
│   │   └── table_handler.go    # Table HTMX endpoint (/api/components/table/rows)
│   └── pages/demo/
│       ├── layout.templ        # Main layout, sidebar, theme selector
│       └── components/         # Demo pages per component
├── css/main.css                # Tailwind source + theme imports
├── all-themes.css              # 13 theme definitions
├── assets/
│   ├── embed.go                # Embedded assets + StylesCSS() accessor
│   ├── js/vendor/              # Bundled Alpine.js + HTMX
│   └── styles.css              # Generated CSS (DO NOT EDIT)
├── tests/e2e/                  # Playwright E2E tests
│   ├── e2e_test.go             # TestMain, shared browser, helpers
│   ├── visual_helpers.go       # Screenshot comparison utilities
│   ├── class_verifier.go       # CSS class extraction/comparison
│   ├── table_htmx_test.go      # Table API-level tests (pagination, sort, filter)
│   ├── table_pagination_nav_test.go  # Browser paginator style tests
│   ├── table_filter_test.go    # Browser filter interaction tests (bar + inline variants)
│   └── sidebar_test.go         # All-components-present test
└── <component-name>/           # Original PenguinUI HTML (for reference/parity)
```

## Component Development Workflow

1. **Analyze reference** — Read PenguinUI HTML in `/<component-name>/`
2. **Create component** — `components/<name>/types.go` + `<name>.templ`
3. **Create demo page** — `internal/pages/demo/components/<name>.templ`
4. **Register route** — Add case in `internal/server/server.go:handleComponent()`
5. **Add to sidebar** — Add entry in `internal/pages/demo/layout.templ:getSidebarItems()`
6. **Write E2E tests** — `tests/e2e/<name>_test.go`
7. **Build & verify** — `templ generate && go build -o bin/server ./cmd/server`

Each component is `components/<name>/` with:
- `types.go` — Config struct, variant constants, CSS class methods
- `<name>.templ` — Template with public entry point + private helpers

## Critical Rules

### Templ escaping — the #1 source of bugs

Templ's `EscapeString` converts `"` → `&quot;`, `'` → `&#39;`, `&` → `&amp;` in
HTML attributes. This silently breaks Alpine.js — Alpine swallows parse failures
without console errors, so components just don't work.

**Symptoms**: dropdown/combobox options missing, no console errors, unit tests
pass but browser fails. Check rendered HTML in devtools — look for `&quot;`
inside `x-data`.

**For simple Alpine.js x-data** (no JS code, just data):
Use unquoted JS property names and avoid string literals with quotes:
```go
// GOOD — unquoted keys, no quotes to escape
return fmt.Sprintf(`{ opened: [false,false], count: 0 }`)
```

**For complex Alpine.js with functions/strings**:
Use `templ.Raw()` inside a `<script>` tag + `Alpine.data()` registration:
```go
// In types.go
func myAlpineScript(cfg Config) string {
    return fmt.Sprintf(`document.addEventListener('alpine:init', () => {
        Alpine.data('myComponent', () => ({
            value: '%s',
            doThing() { htmx.ajax('GET', '/api/data', {target: '#target'}); }
        }));
    });`, cfg.DefaultValue)
}

// In .templ
templ myScript(cfg Config) {
    @templ.Raw("<script>" + myAlpineScript(cfg) + "</script>")
}
// Then reference with: <div x-data="myComponent">
```

**NEVER use `json.Marshal` for data that ends up inside HTML attributes via
templ.** `json.Marshal` produces double-quoted strings; templ escapes the
quotes; Alpine sees broken syntax. Use single-quoted JS string builders:

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

### Null arrays crash Alpine.js

Go's `json.Marshal([]string(nil))` produces `null`, not `[]`. Alpine code like
`selectedValues.includes(...)` throws on null. Always guard:
```go
if string(selectedJSON) == "null" {
    selectedJSON = []byte("[]")
}
```

### Generated files — NEVER edit manually
- `*_templ.go` — regenerated by `templ generate`
- `assets/styles.css` — regenerated by Tailwind
- When resolving merge conflicts, resolve `.templ` source files then `templ generate`. Never hand-resolve generated files.

### Templ regeneration quirks
- `templ generate` sometimes reports "0 updates" even when source changed
- Force regeneration: `rm components/<name>/<name>_templ.go && templ generate`
- Always check the generated `_templ.go` if rendering looks wrong
- Use `templ.Attributes` for dynamic attribute maps

### Dark mode — always use both light and dark variants
```css
bg-surface text-on-surface dark:bg-surface-dark dark:text-on-surface-dark
```
Test in both light and dark, across multiple themes (especially Minimal which
has no border-radius). Check responsive breakpoints — mobile sidebar behavior
differs significantly.

### Port 8090 reserved for manual dev
E2E tests use a random free port (`freePort()` in `e2e_test.go`). Never hardcode
8090 in tests.

### Layout: avoid duplicate headers/branding
Sidebar has its own logo section, conditional on `Logo`/`LogoText` being set. In
the demo layout, omit `LogoText` since the page header already shows "Goshtoso
PenguinUI". Don't duplicate positioning CSS — the layout wrapper handles
`fixed`/`static` responsive positioning; the sidebar handles its own styling
(borders, background, flex).

### Mobile sidebar positioning
`fixed inset-y-0` makes the mobile sidebar overlap the sticky header. Use
`fixed top-16 bottom-0` so the sidebar starts below the 4rem header.

## E2E Testing Architecture

```
TestMain (e2e_test.go)
├── Builds server binary (always rebuilds to avoid stale binary)
├── Finds free port via freePort(), starts server
├── Launches single shared Chromium (Playwright) — 1 launch, not per-test
└── Runs all tests with shared browser

Each test:
├── Gets shared browser via setupPlaywright()
├── Creates new page/tab via newPage(t, browser)
├── Page auto-closes via t.Cleanup()
└── Uses 2s element timeout / 3s navigation timeout
```

Key helpers:
- `newPage(t, browser, ...opts)` — creates tab with tight timeouts + auto-close
- `setupServer(t)` / `setupPlaywright(t)` — no-ops (backward compat, TestMain handles everything)
- `takeScreenshot(t, page, name)` — saves debug screenshots
- `fillSearchInput()` — fills + dispatches `input` event for Alpine `x-model`
- `CompareScreenshots()` / `AssertClassParity()` — visual + class parity helpers (`visual_helpers.go`, `class_verifier.go`)

### Alpine.js in tests
- `GetAttribute("aria-expanded")` returns the static HTML attribute, NOT the
  Alpine-bound live value. Use `Evaluate("el => el.getAttribute('aria-expanded')", nil)` instead.
- Wait for Alpine: `WaitForFunction("() => typeof Alpine !== 'undefined'")`.
- Playwright `Locator.Fill()` does NOT fire a native `input` event — Alpine
  `x-model` won't update. Tests using debounced inputs must dispatch `input`
  manually (`fillSearchInput` helper).

## Table Component

The table is the most complex component. Key features:
- Static, striped, checkbox variants
- Sortable columns (sort cycles: neutral → asc → desc → neutral); `SortURL`
  omits sort params when direction is `SortNone`
- Pagination with HTMX OOB swap (paginator updates active state, "Page X of Y"
  text, Prev/Next disabled states)
- Sort headers update via HTMX OOB swap on click (`TheadID()`, `TableHeadOOB`)
- Infinite scroll with sentinel row
- Lazy loading
- Filter bar (`bar` + `inline` variants); search/select/toggle filter types;
  `FilterConfig.HxTarget` overrides default tbody target
- Filter Alpine component registered via `<script>` + `Alpine.data()` to avoid
  templ escaping; `htmx:configRequest` listener appends filter params to all
  HTMX requests

HTMX endpoint: `/api/components/table/rows`
Query params: `order_by`, `order_dir`, `page`, `per_page`, `search`,
`membership`, `variant`

## Theme System

- Themes defined in `all-themes.css` using `[data-theme="name"]` selectors (13 themes)
- Dark mode uses `.dark` class on `<html>` via Alpine.js store
- Default theme: **Minimal** (black/white, no border radius)

## Current Status

**Completed (22):** Accordion, Alert, Avatar, Badge, Banner, Button, Card,
Checkbox, Combobox, Dropdown, Modal, Pagination, Select, Sidebar, Spinner,
Table, Tabs, Textarea, Text Input, Toast, Toggle, Tooltip.

381 E2E tests passing, full suite ~2.5 minutes, no skipped tests.

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

Alpine.js, HTMX, and the htmx SSE extension (`htmx-ext-sse.min.js`) are bundled
locally under `assets/js/vendor/` (no CDN at runtime). Page loads are
deterministic in E2E.

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

# Lint (gates PRs in CI — must be clean before merge)
golangci-lint run

# Apply go fix modernizations (also runs automatically via the pre-commit hook)
go fix ./...
```

### Linting & modernization

- **golangci-lint** gates every PR (`lint-build` job in `.github/workflows/ci.yml`).
  Config is `.golangci.yml`: the standard linters (errcheck, govet, ineffassign,
  staticcheck, unused) plus **cyclop** with a hard **cyclomatic-complexity ceiling
  of 20**. Keep new functions under 20 — extract helpers instead of suppressing.
- Generated `*_templ.go` files are excluded from linting (strict generated header).
- **`go fix`** ([go.dev/blog/gofix](https://go.dev/blog/gofix)) applies AST-safe
  modernizations (e.g. `slices.Contains`, `min`, `for i := range n`). It runs via
  the committed **pre-commit hook** — enable it once per clone:
  ```bash
  git config core.hooksPath .githooks
  ```
  The hook applies `go fix ./...`, never touches generated files, and re-stages
  any modernized staged files (review with `git diff --cached`, then commit again).
  It is intentionally NOT a CI gate — run `go fix ./...` locally before committing.

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
   (MUST follow the docs-page pattern — see below)
4. **Register route** — Add case in `internal/server/server.go:handleComponent()`
5. **Add to sidebar** — Add entry in `internal/pages/demo/layout.templ:getSidebarItems()`
6. **Write E2E tests** — `tests/e2e/<name>_test.go`
7. **Build & verify** — `templ generate && go build -o bin/server ./cmd/server`
8. **Sync the usage skill** — `go run ./scripts/skillgen` regenerates
   `.claude/skills/using-goshtoso/components-reference.md` from the component
   source. The pre-commit hook does this automatically when `components/**.go`
   is staged; CI fails if it is stale. Run it after any change to a component's
   `types.go` or entry points. Never hand-edit the generated reference.

Each component is `components/<name>/` with:
- `types.go` — Config struct, variant constants, CSS class methods
- `<name>.templ` — Template with public entry point + private helpers

### Demo/docs page pattern (mandatory for ALL components)

Every demo page MUST follow the shared docs pattern, mirroring penguinui.com:
**one preview box + one code box per variant** (`demo.ComponentDemo` for the
first/primary variant, `demo.DemoSection` for each additional one), an
**`demo.APIReference` table at the bottom** (outside the `#<name>-fragment`
e2e anchor), and the **right-rail "On this page" TOC** which auto-builds from
the `data-toc-heading` headings the helpers emit. Give each variant container a
unique ID (`<name>-default`, `<name>-split`, …). Rebuild Tailwind + `go build`
after introducing any new utility class (CSS is embedded).

All 36 component demo pages follow this pattern (API reference via a bottom
`demo.APIReference` call, or the inline `Props:` field on `ComponentDemoProps` —
both render the same table). The only exempt pages are the non-component
specials: **getting-started**, **landing**, and **theme** (no variants/API).

**Full reference + skeleton + pitfalls:** the **`component-docs` skill**
(`.claude/skills/component-docs/SKILL.md`). Canonical example:
`internal/pages/demo/components/accordion.templ`. Invoke the skill before
creating or restructuring any demo page.

## Codex Integration

The **Codex CLI** (OpenAI) is wired into this repo via the `codex` plugin as a
second engine for review and parallel work. Auth is ChatGPT login; the shared
runtime starts on demand on the first review/task command.

- **Stop-time review gate is ON.** A fresh Codex review is required before a
  Claude session can stop. Don't disable it casually.
- **Use Codex for code review.** After a non-trivial change — and always before
  finishing a branch/PR — hand the diff to Codex for an independent second
  opinion (`codex:rescue` skill, or the review gate). Treat its findings as a
  reviewer's, not gospel: confirm against the code before acting.
- **Use Codex for parallel work when it fits.** When a task splits into
  independent slices (e.g. several components, a broad migration, a sweep), or
  when a hard problem benefits from a second implementation/diagnosis pass,
  delegate a slice to Codex via the `codex:codex-rescue` agent / `codex:rescue`
  skill and work the rest in parallel. Reserve for genuinely independent or
  stuck work — not routine single-file edits.
- Setup/health check: `/codex:setup`. Manual login if auth lapses: `!codex login`.

## Critical Rules

### Frontend interactivity hierarchy: htmx → Alpine.js → vanilla JS

When adding any client behavior, pick the **highest** tier that solves the
problem, never a lower one for convenience:

1. **Server-side rendering driven by htmx (default, idiomatic).** Prefer a
   server round-trip that returns rendered HTML (templ) swapped by htmx over
   holding state or building markup on the client. Reach first for `hx-get`/
   `hx-post`, swaps, OOB swaps, and `HX-*` response headers. Most "interactive"
   needs (filtering, pagination, inline edit, lazy load, validation, toasts)
   are SSR + htmx, not JavaScript.
2. **Alpine.js — only for genuinely client-side interactivity.** Use Alpine
   when the behavior is **not achievable via SSR, or a server round-trip would
   be too expensive / too laggy** for the interaction: purely local UI state
   (open/closed, tabs, hover, focus traps), instant input feedback, and
   transitions where a network hop would feel wrong. Don't use Alpine to do
   what htmx already does.
3. **Vanilla JS — last resort.** Only when neither htmx nor Alpine fits (a niche
   browser API, a one-off integration). Justify it; prefer wrapping it in an
   `Alpine.directive`/`Alpine.magic` so it composes with the rest.

**Solutions must play nicely with htmx first.** Anything you add has to survive
htmx swaps and fragment-nav: register `Alpine.data()` immediately (not only on
`alpine:init`) so swapped-in nodes can find it, call `htmx.process()` on nodes
you insert yourself, and gate `hx-swap-oob` to update-only so first paint
doesn't error. See the **`htmx`** and **`alpinejs`** skills (and
`alpinejs/reference/gotchas.md`) for the escaping and swap pitfalls.

### Always work in a git worktree

Make changes in an isolated git worktree, never directly on a shared branch in
the main checkout. Use the `superpowers:using-git-worktrees` skill to create one
before starting feature work, bugfixes, or any non-trivial edit. This keeps the
main checkout clean and lets multiple changes proceed without collision.

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

### HTMX swap rebind race — clicking a just-swapped control

When a click triggers an HTMX swap that **replaces the control itself**
(`hx-swap="outerHTML"` on a wrapper, or a thead/paginator OOB swap), htmx
re-binds the swapped-in element a beat *after* it appears in the DOM. A test
that clicks the new control in that window fires **no request** — the state
never advances and the click is lost forever. This is load-sensitive: it passes
in isolation but fails under full-suite browser pressure (it sank `TestSteps`
and `TestTableHTMX_SortCycling`).

Don't wait on a proxy attribute then click — use the **`clickUntil(t, page,
loc, jsCondition)`** helper (`e2e_test.go`). It clicks and waits for the
condition, re-firing only when state hasn't moved (a lost click fires no
request, so it never double-advances a stateful control).

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

## Example Apps (`/examples/*`)

Full, runnable apps that showcase the components in real use (distinct from the
per-component docs pages). They flow through the **same demo registry +
`renderDemo`** (so theme, dark mode, and fragment-swap nav come for free) and get
their own collapsible **"Examples"** sidebar section.

- **Stateless by design:** per-user state is serialized into a cookie, never held
  in server memory. The todo reference app uses cookie `gt_todo`
  (base64url-JSON), capped by an encoded-size budget (`maxCookieBytes`) so the
  browser never silently drops it.
- **Layering:** pure domain logic in `internal/examples/<app>/` (HTTP-free,
  unit-tested) → exported templ in `internal/pages/demo/examples/` (rendered by
  both the page shell *and* the HTMX handlers) → thin handlers in
  `internal/server/` (read cookie → mutate → write cookie → render fragment).
- **HTMX pattern:** every membership-changing mutation re-renders the whole
  `#todo-list` (`outerHTML`) plus an OOB count badge (and a toast where apt), so
  empty-state and filter membership stay correct. Reorder is via up/down buttons
  (`/move`); there is intentionally no drag-and-drop (native HTML5 DnD was
  removed as unreliable).
- **Components used (vs hand-rolled):** the filter is a segmented `radio` group
  (native `:checked` is the active highlight, and it lives outside `#todo-list`
  so it survives every swap); Clear-completed is a `button.Button` wrapped in a
  `#todo-clear` OOB div so its disabled state can update; the undo-delete bar is
  an `alert.Alert` whose primary action POSTs to `/restore`. The row's done
  checkbox and the icon-only ↑/↓/✕ stay raw — the `checkbox` component has no
  HTMX hook and `button` has no aria-label/attrs hook for icon-only buttons.
- **OOB + fragment-nav gotcha:** an element that carries `hx-swap-oob` on first
  paint will make htmx try to OOB-swap it (no target) when the page arrives via
  a sidebar fragment swap → `htmx:oobErrorNoTarget`. Gate the attribute to
  update-only (see `CountBadge`/`ClearButton` `oob bool`). Likewise, register
  any `Alpine.data()` immediately when Alpine is already running, not only on
  `alpine:init`, or it is undefined on fragment nav.

- **Live Log Feed** (`/examples/logs`) is the first SSE / server-push example.
  Transport: the vendored htmx SSE extension (`htmx-ext-sse`); markup uses
  `hx-ext="sse"` + `sse-connect`. The server endpoint
  (`/api/examples/logs/stream`, query params `interval` + `max`) emits one
  server-rendered `LogRow` per tick and returns on `ctx.Done()` — no shared
  state, only the per-connection goroutine. Division of labour: htmx owns row
  *insertion* (append into `#log-feed`); Alpine (`x-data="logFeed"`, registered
  via `<script>` + `Alpine.data()`, immediately if Alpine is already running so
  fragment-nav works) owns row cap (last 100), the min-severity filter (pure
  scoped CSS), auto-scroll, pause/resume, and connection status. Pause removes
  the `sse-connect` *connector* element (which closes the `EventSource`) while
  the persistent `#log-feed` retains its rows because the connector targets it
  via `hx-target`; on resume the re-added connector is processed with
  `htmx.process()` (htmx does not auto-process Alpine-inserted nodes).
  Components showcased: Badge (levels + status), Button (pause/clear), Select
  (min-level filter), Spinner (connecting), Toggle (auto-scroll), Tooltip.

**To add an example:** domain pkg under `internal/examples/` + templ in
`internal/pages/demo/examples/` + a `Demos` registry entry + a sidebar item +
E2E (cover the **sidebar fragment-nav** path and assert no console errors, not
just direct loads). Reference app: `internal/pages/demo/examples/todo.templ`.
Endpoints: `/api/examples/todo/{add,toggle,delete,edit,filter,move,clear-completed,restore}`.

## Current Status

**Completed (22):** Accordion, Alert, Avatar, Badge, Banner, Button, Card,
Checkbox, Combobox, Dropdown, Modal, Pagination, Select, Sidebar, Spinner,
Table, Tabs, Textarea, Text Input, Toast, Toggle, Tooltip.

381 E2E tests passing, full suite ~2.5 minutes, no skipped tests.

**Example apps:** Todo List (`/examples/todo`) — cookie-backed, HTMX-driven; the
first entry in the `/examples/*` family. Live Log Feed (`/examples/logs`) —
SSE-streamed synthetic logs over htmx-ext-sse + Alpine; the first server-push
example.

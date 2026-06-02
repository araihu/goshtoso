# Homepage Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the stack-first landing page with a "show, don't tell" homepage that leads with what Goshtoso does — a live theme-switching component playground and an example-apps gallery — and demotes the stack to a single strip.

**Architecture:** One standalone templ document (`landing.templ`, package `components`, rendered at `server.go:99`). It is NOT part of the `renderDemo`/fragment-nav shell, so no OOB/fragment-nav gotchas apply. The page loads its own Alpine + (newly) htmx vendor scripts. Theme switching is a local Alpine inline expression that sets `data-theme` on `<html>` and persists to `localStorage` (mirroring `layout.templ`'s `setTheme`). The live table is HTMX-driven against the existing `/api/components/table/rows` endpoint. Real component entry points are reused — nothing is mocked or screenshotted.

**Tech Stack:** Go 1.26, templ v0.3, Tailwind v4, Alpine.js v3, HTMX v2, Playwright (playwright-go) E2E.

---

## File Structure

- **Modify:** `internal/pages/demo/components/landing.templ` — the whole rewrite lives here. New imports for the reused component packages; new data slices (`exampleApp`, `themeChip`); `landingContent()` rebuilt section by section; `<head>` gains an htmx `<script>`. `cookieConsent()` is unchanged.
- **Regenerated:** `internal/pages/demo/components/landing_templ.go` — via `templ generate`. NEVER hand-edited.
- **Create:** `tests/e2e/landing_test.go` — E2E coverage for the new page (currently no landing test exists).
- **Regenerated:** `assets/styles.css` — via Tailwind, only if a new utility class is introduced.

No other files change. No new components, no new server endpoints.

---

## Conventions used throughout

- **Component invocation** (confirmed from existing demos):
  - `@button.Button(button.Config{Variant: button.Primary, Type: "button"}) { Label }`
  - `@badge.Badge(badge.Config{Text: "Active", Variant: badge.Success})`
  - `@toggle.Toggle(toggle.Config{ID: "demoToggle", Label: "Notifications", Checked: true})`
  - Live table (lazy variant — fetches its own rows on load, so no initial `Rows` needed):
    `@table.Table(table.Config{ID: "home-table", HTMXEndpoint: "/api/components/table/rows?variant=lazy", LazyLoad: true, Columns: []table.Column{...}})`
- **Imports** use module root `github.com/araihu/goshtoso/...`.
- **Templ escaping rule (verified):** static single-quoted JS inside `@click`/`x-init` is emitted verbatim; only `{ }` interpolation is HTML-escaped. So dynamic values (theme key) go in a plain attribute (`data-theme-key={ t.Key }`) and the static `@click` reads `$el.dataset.themeKey`.
- **Dark mode:** every color utility needs both light and `dark:` variants.
- **After any `.templ` edit:** run `templ generate`, then `go build -o bin/server ./cmd/server`. If a new utility class was added, also `tailwindcss -i css/main.css -o assets/styles.css`.

---

### Task 1: E2E test scaffold — page loads, hero present (failing first)

**Files:**
- Create: `tests/e2e/landing_test.go`

- [ ] **Step 1: Write the failing test**

```go
package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestLanding_HeroAndStructure loads the homepage ("/") and asserts the
// redesigned hero — headline + primary CTA — is present. The page is a
// standalone document (not the demo shell).
func TestLanding_HeroAndStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("HeroHeadline", func(t *testing.T) {
		h1 := page.Locator("#hero h1")
		txt, err := h1.InnerText()
		require.NoError(t, err)
		require.Contains(t, txt, "Build interactive UIs in Go")
	})

	t.Run("BrowseComponentsCTA", func(t *testing.T) {
		cta := page.Locator("#hero a[href='/components/button']")
		visible, err := cta.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "Browse components CTA should be visible")
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestLanding_HeroAndStructure`
Expected: FAIL — there is no `#hero` element / headline text yet (current hero has no `id` and different copy).

- [ ] **Step 3: No implementation yet** — proceed to Task 2 (the page rewrite makes this pass). Do not commit a failing test alone; Task 2 commits test + impl together.

---

### Task 2: Rewrite head + hero section

**Files:**
- Modify: `internal/pages/demo/components/landing.templ` (head: add htmx; replace hero block; update imports)

- [ ] **Step 1: Add imports + example/theme data slices**

Replace the top of the file (the `stackInfo`/`getStackItems` block, lines 1–18) with:

```go
package components

import (
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/components/toggle"
)

type exampleApp struct {
	Title string
	Desc  string
	URL   string
}

func getExampleApps() []exampleApp {
	return []exampleApp{
		{Title: "Todo List", Desc: "Cookie-backed, fully HTMX-driven CRUD — no server-side session state.", URL: "/examples/todo"},
		{Title: "Chat", Desc: "Real-time messaging over a WebSocket, server-rendered bubbles.", URL: "/examples/chat"},
		{Title: "Live Log Feed", Desc: "Server-push log stream over SSE (htmx-ext-sse) with client-side filtering.", URL: "/examples/logs"},
		{Title: "Profile", Desc: "An account settings screen composed from form components.", URL: "/examples/profile"},
		{Title: "Live Ticker", Desc: "A streaming price table that updates rows in place.", URL: "/examples/ticker"},
	}
}

type themeChip struct {
	Key   string
	Label string
}

// curated subset of the 15 themes — visually distinct, not the whole set
func getThemeChips() []themeChip {
	return []themeChip{
		{Key: "minimal", Label: "Minimal"},
		{Key: "modern", Label: "Modern"},
		{Key: "dracula", Label: "Dracula"},
		{Key: "90s", Label: "90s"},
		{Key: "pastel", Label: "Pastel"},
		{Key: "neo-brutalism", Label: "Neo Brutalism"},
	}
}
```

- [ ] **Step 2: Add htmx to the `<head>`**

In the `<head>` (after the existing alpine `<script defer ...>` at landing.templ:35), add htmx so the live table works on this standalone page:

```html
<script src="/assets/js/vendor/htmx.min.js"></script>
```

- [ ] **Step 3: Replace the hero block**

Replace the current hero `<div class="flex flex-col items-center text-center"> ... </div>` (landing.templ:48–80) with:

```html
<!-- Hero -->
<div id="hero" class="flex flex-col items-center text-center">
	<img
		src="/assets/images/goshtoso-art.png"
		alt="Goshtoso mascot — a blue Go gopher in sunglasses and a Hawaiian shirt at Copacabana"
		class="w-32 h-auto rounded-radius mb-6"
	/>
	<h1 class="text-4xl md:text-5xl font-bold font-title text-on-surface-strong dark:text-on-surface-dark-strong">
		Build interactive UIs in Go.
	</h1>
	<p class="mt-3 text-lg md:text-xl text-on-surface dark:text-on-surface-dark">
		No JavaScript build step.
	</p>
	<p class="mt-4 max-w-2xl text-on-surface dark:text-on-surface-dark">
		Goshtoso is a server-rendered UI component library for Go. Templates render HTML,
		HTMX swaps fragments, and Alpine.js adds local interactivity — all shipped as a single binary.
	</p>
	<div class="mt-8 flex flex-wrap gap-4 justify-center">
		<a href="/components/button" class="px-6 py-2.5 rounded-radius bg-primary text-on-primary dark:bg-primary-dark dark:text-on-primary-dark font-medium text-sm hover:opacity-90 transition-opacity">
			Browse components
		</a>
		<a href="https://github.com/araihu/goshtoso" target="_blank" class="px-6 py-2.5 rounded-radius border border-outline dark:border-outline-dark text-on-surface dark:text-on-surface-dark font-medium text-sm hover:bg-surface-alt dark:hover:bg-surface-dark-alt transition-colors">
			GitHub
		</a>
		<a href="/getting-started" class="px-6 py-2.5 rounded-radius border border-outline dark:border-outline-dark text-on-surface-strong dark:text-on-surface-dark-strong font-medium text-sm hover:bg-surface-alt dark:hover:bg-surface-dark-alt transition-colors">
			Get started
		</a>
	</div>
</div>
```

- [ ] **Step 4: Regenerate + build**

Run:
```bash
templ generate && go build -o bin/server ./cmd/server
```
Expected: no errors. (If `templ generate` reports "0 updates", force: `rm internal/pages/demo/components/landing_templ.go && templ generate`.)

- [ ] **Step 5: Run the Task 1 test → passes**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestLanding_HeroAndStructure`
Expected: PASS (both subtests).

- [ ] **Step 6: Commit**

```bash
git add internal/pages/demo/components/landing.templ internal/pages/demo/components/landing_templ.go tests/e2e/landing_test.go
git commit -m "feat(landing): show-don't-tell hero + htmx on homepage"
```

---

### Task 3: Live playground section (theme chips + components + lazy HTMX table)

**Files:**
- Modify: `internal/pages/demo/components/landing.templ` (replace the old "The Stack" section)
- Modify: `tests/e2e/landing_test.go` (add playground subtests)

- [ ] **Step 1: Write the failing tests** — append two subtests inside `TestLanding_HeroAndStructure` (after the `BrowseComponentsCTA` subtest):

```go
	t.Run("ThemeChipSwitchesTheme", func(t *testing.T) {
		_, err := page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
		require.NoError(t, err)
		// click the Dracula chip
		require.NoError(t, page.Locator("#playground button[data-theme-key='dracula']").Click())
		got, err := page.Evaluate("() => document.documentElement.getAttribute('data-theme')", nil)
		require.NoError(t, err)
		require.Equal(t, "dracula", got, "clicking a chip should set data-theme on <html>")
	})

	t.Run("LiveTableLoadsRows", func(t *testing.T) {
		// lazy table fetches rows via HTMX on load
		_, err := page.WaitForFunction(
			"() => document.querySelectorAll('#home-table tbody tr').length > 0", nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)},
		)
		require.NoError(t, err, "live HTMX table should populate rows")
	})
```

- [ ] **Step 2: Run to verify failure**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestLanding_HeroAndStructure`
Expected: FAIL — `#playground` and `#home-table` do not exist yet.

- [ ] **Step 3: Implement the playground section**

Replace the entire "The Stack" block (current landing.templ:82–97, the `<div>` with `<h2>The Stack</h2>` and the `getStackItems()` loop) with:

```html
<!-- Live playground -->
<div id="playground">
	<h2 class="text-2xl font-bold font-title text-on-surface-strong dark:text-on-surface-dark-strong mb-2 text-center">See it live</h2>
	<p class="text-center text-on-surface dark:text-on-surface-dark mb-6">
		Pick a theme — every component below recolors instantly. These are real Goshtoso components, not screenshots.
	</p>
	<div class="flex flex-wrap gap-2 justify-center mb-8">
		for _, c := range getThemeChips() {
			<button
				type="button"
				data-theme-key={ c.Key }
				@click="document.documentElement.setAttribute('data-theme', $el.dataset.themeKey); localStorage.setItem('theme', $el.dataset.themeKey)"
				class="px-4 py-1.5 rounded-radius border border-outline dark:border-outline-dark text-sm text-on-surface dark:text-on-surface-dark hover:border-primary dark:hover:border-primary-dark hover:text-primary dark:hover:text-primary-dark transition-colors"
			>
				{ c.Label }
			</button>
		}
		<a href="/docs/theme" class="px-4 py-1.5 rounded-radius text-sm text-primary dark:text-primary-dark underline underline-offset-2 self-center">
			All 15 themes →
		</a>
	</div>
	<div class="border border-outline dark:border-outline-dark rounded-radius p-6 bg-surface dark:bg-surface-dark space-y-6">
		<div class="flex flex-wrap items-center gap-4">
			@button.Button(button.Config{Variant: button.Primary, Type: "button"}) { Primary }
			@button.Button(button.Config{Variant: button.Secondary, Type: "button"}) { Secondary }
			@badge.Badge(badge.Config{Text: "Active", Variant: badge.Success})
			@badge.Badge(badge.Config{Text: "Pending", Variant: badge.Warning})
			@toggle.Toggle(toggle.Config{ID: "homeToggle", Label: "Notifications", Checked: true})
		</div>
		@table.Table(table.Config{
			ID:           "home-table",
			HTMXEndpoint: "/api/components/table/rows?variant=lazy",
			LazyLoad:     true,
			Columns: []table.Column{
				{Key: "id", Label: "CustomerID", Sortable: true},
				{Key: "name", Label: "Name", Sortable: true},
				{Key: "email", Label: "Email"},
				{Key: "membership", Label: "Membership", Sortable: true},
			},
		})
	</div>
</div>
```

> Note: verify `badge.Warning` and `badge.Success` constants exist (`grep -nE 'Warning|Success' components/badge/types.go`). If a constant name differs, use the actual one (e.g. `badge.Primary`). Do not invent names.

- [ ] **Step 4: Regenerate + build**

Run:
```bash
templ generate && go build -o bin/server ./cmd/server
```
Expected: compiles. If it fails on an unknown badge/toggle field or constant, fix to the real identifier (confirm via `grep` in the component's `types.go`), then rebuild.

- [ ] **Step 5: Run tests → pass**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestLanding_HeroAndStructure`
Expected: PASS, including `ThemeChipSwitchesTheme` and `LiveTableLoadsRows`.

- [ ] **Step 6: Commit**

```bash
git add internal/pages/demo/components/landing.templ internal/pages/demo/components/landing_templ.go tests/e2e/landing_test.go
git commit -m "feat(landing): live theme-switching component playground"
```

---

### Task 4: Example apps gallery

**Files:**
- Modify: `internal/pages/demo/components/landing.templ` (replace the "Why?" section with the gallery)
- Modify: `tests/e2e/landing_test.go` (add gallery subtest)

- [ ] **Step 1: Write the failing test** — append inside `TestLanding_HeroAndStructure`:

```go
	t.Run("ExampleGalleryLinks", func(t *testing.T) {
		for _, route := range []string{
			"/examples/todo", "/examples/chat", "/examples/logs",
			"/examples/profile", "/examples/ticker",
		} {
			loc := page.Locator("#examples a[href='" + route + "']")
			cnt, err := loc.Count()
			require.NoError(t, err)
			require.GreaterOrEqual(t, cnt, 1, "gallery should link to "+route)
		}
	})
```

- [ ] **Step 2: Run to verify failure**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestLanding_HeroAndStructure`
Expected: FAIL — `#examples` does not exist.

- [ ] **Step 3: Implement the gallery**

Replace the entire "Why?" block (current landing.templ:99–117, the `<div class="max-w-3xl mx-auto">` with `<h2>Why?</h2>`) with:

```html
<!-- Example apps gallery -->
<div id="examples">
	<h2 class="text-2xl font-bold font-title text-on-surface-strong dark:text-on-surface-dark-strong mb-2 text-center">Built with Goshtoso</h2>
	<p class="text-center text-on-surface dark:text-on-surface-dark mb-6">
		Full, runnable apps composed entirely from these components.
	</p>
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
		for _, a := range getExampleApps() {
			<a href={ templ.SafeURL(a.URL) } class="block border border-outline dark:border-outline-dark rounded-radius p-6 bg-surface dark:bg-surface-dark hover:border-primary dark:hover:border-primary-dark transition-colors group">
				<h3 class="text-lg font-bold font-title text-on-surface-strong dark:text-on-surface-dark-strong mb-2 group-hover:text-primary dark:group-hover:text-primary-dark transition-colors">
					{ a.Title }
				</h3>
				<p class="text-sm text-on-surface dark:text-on-surface-dark leading-relaxed">
					{ a.Desc }
				</p>
			</a>
		}
	</div>
</div>
```

- [ ] **Step 4: Regenerate + build**

Run: `templ generate && go build -o bin/server ./cmd/server`
Expected: compiles.

- [ ] **Step 5: Run tests → pass**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestLanding_HeroAndStructure`
Expected: PASS including `ExampleGalleryLinks`.

- [ ] **Step 6: Commit**

```bash
git add internal/pages/demo/components/landing.templ internal/pages/demo/components/landing_templ.go tests/e2e/landing_test.go
git commit -m "feat(landing): example apps gallery"
```

---

### Task 5: How-it-works + condensed stack strip + footer

**Files:**
- Modify: `internal/pages/demo/components/landing.templ` (add how-it-works + stack strip before footer; footer copy stays)
- Modify: `tests/e2e/landing_test.go` (add stack-strip subtest)

- [ ] **Step 1: Write the failing test** — append inside `TestLanding_HeroAndStructure`:

```go
	t.Run("StackStripCondensed", func(t *testing.T) {
		strip := page.Locator("#stack-strip")
		visible, err := strip.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "condensed stack strip should be present")
		// it is a single strip, not six cards: at most a handful of links
		links := page.Locator("#stack-strip a")
		cnt, err := links.Count()
		require.NoError(t, err)
		require.LessOrEqual(t, cnt, 6, "stack strip should be condensed, not a card grid")
	})
```

- [ ] **Step 2: Run to verify failure**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestLanding_HeroAndStructure`
Expected: FAIL — `#stack-strip` does not exist.

- [ ] **Step 3: Implement how-it-works + stack strip**

Immediately BEFORE the existing `<!-- Footer -->` block (current landing.templ:119), insert:

```html
<!-- How it works -->
<div class="max-w-3xl mx-auto">
	<h2 class="text-2xl font-bold font-title text-on-surface-strong dark:text-on-surface-dark-strong mb-6 text-center">How it works</h2>
	<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
		<div class="text-center">
			<div class="text-2xl font-bold text-primary dark:text-primary-dark mb-2">1</div>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">Write a typed <code class="font-mono text-on-surface-strong dark:text-on-surface-dark-strong">.templ</code> component — HTML that compiles to Go.</p>
		</div>
		<div class="text-center">
			<div class="text-2xl font-bold text-primary dark:text-primary-dark mb-2">2</div>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">Add an <code class="font-mono text-on-surface-strong dark:text-on-surface-dark-strong">hx-get</code> and the server swaps a rendered fragment — no fetch, no JSON.</p>
		</div>
		<div class="text-center">
			<div class="text-2xl font-bold text-primary dark:text-primary-dark mb-2">3</div>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">Build one binary. No bundler, no <code class="font-mono text-on-surface-strong dark:text-on-surface-dark-strong">node_modules</code>, no client routing.</p>
		</div>
	</div>
</div>

<!-- Stack strip -->
<div id="stack-strip" class="text-center">
	<p class="text-sm text-on-surface-muted dark:text-on-surface-dark-muted mb-3">Built on</p>
	<div class="flex flex-wrap gap-x-6 gap-y-2 justify-center text-sm">
		<a href="https://go.dev/" target="_blank" class="font-medium text-on-surface dark:text-on-surface-dark hover:text-primary dark:hover:text-primary-dark transition-colors">Go</a>
		<a href="https://templ.guide/" target="_blank" class="font-medium text-on-surface dark:text-on-surface-dark hover:text-primary dark:hover:text-primary-dark transition-colors">Templ</a>
		<a href="https://tailwindcss.com/" target="_blank" class="font-medium text-on-surface dark:text-on-surface-dark hover:text-primary dark:hover:text-primary-dark transition-colors">Tailwind CSS</a>
		<a href="https://alpinejs.dev/" target="_blank" class="font-medium text-on-surface dark:text-on-surface-dark hover:text-primary dark:hover:text-primary-dark transition-colors">Alpine.js</a>
		<a href="https://htmx.org/" target="_blank" class="font-medium text-on-surface dark:text-on-surface-dark hover:text-primary dark:hover:text-primary-dark transition-colors">HTMX</a>
		<a href="https://penguinui.com/" target="_blank" class="font-medium text-on-surface dark:text-on-surface-dark hover:text-primary dark:hover:text-primary-dark transition-colors">PenguinUI</a>
	</div>
</div>
```

The existing Footer block stays as-is.

- [ ] **Step 4: Regenerate + build**

Run: `templ generate && go build -o bin/server ./cmd/server`
Expected: compiles.

- [ ] **Step 5: Run tests → pass**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestLanding_HeroAndStructure`
Expected: PASS including `StackStripCondensed`.

- [ ] **Step 6: Commit**

```bash
git add internal/pages/demo/components/landing.templ internal/pages/demo/components/landing_templ.go tests/e2e/landing_test.go
git commit -m "feat(landing): how-it-works steps + condensed stack strip"
```

---

### Task 6: No-console-errors guard + title + final verification

**Files:**
- Modify: `tests/e2e/landing_test.go` (add a standalone no-console-errors test + title check)

- [ ] **Step 1: Write the test**

Append a new top-level test to `landing_test.go`:

```go
// TestLanding_NoConsoleErrors loads the homepage and asserts no JS console or
// page errors — the silent-Alpine-failure guard (broken x-data fails quietly).
func TestLanding_NoConsoleErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	title, err := page.Title()
	require.NoError(t, err)
	require.Contains(t, title, "Goshtoso")

	require.Empty(t, jsErrors, "no JS console/page errors on homepage: %v", jsErrors)
}
```

- [ ] **Step 2: Run the full landing test file**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestLanding`
Expected: PASS — both `TestLanding_HeroAndStructure` and `TestLanding_NoConsoleErrors`.

- [ ] **Step 3: Lint + check for accidental new utility classes**

Run:
```bash
golangci-lint run
```
Expected: clean. If any new Tailwind utility class was introduced that isn't already compiled, run `tailwindcss -i css/main.css -o assets/styles.css`, rebuild, and re-run the landing tests.

- [ ] **Step 4: Visual sanity check in both modes**

Run the dev server (`go run cmd/server/main.go`), open `http://localhost:8090/`, and verify: hero, theme chips recolor the cluster, the table loads rows, the gallery links work, dark mode toggles cleanly, and the Minimal theme (no border-radius) still looks right. (Manual — no automated assertion.)

- [ ] **Step 5: Full E2E suite (regression)**

Run: `go test ./tests/e2e/... -count=1 -timeout 15m`
Expected: all pass (previously 381+; now +new landing tests). If `TestLanding_*` flakes under full-suite load on the table, switch the table-row wait to `clickUntil`-style polling (see `e2e-suite-flaky-full-run` memory) — but the lazy table loads on page load (no click), so a `WaitForFunction` on row count is the correct guard and should be stable.

- [ ] **Step 6: Commit**

```bash
git add tests/e2e/landing_test.go
git commit -m "test(landing): no-console-errors guard + title check"
```

---

## Self-Review

**Spec coverage:**
- Hero (headline, CTAs, shrunk mascot) → Task 2 ✓
- Live playground (theme chips, real components, HTMX table) → Task 3 ✓
- Example apps gallery (5 routes) → Task 4 ✓
- How it works (3 steps) → Task 5 ✓
- Condensed stack strip (replaces 6-card section) → Task 5 ✓
- Footer unchanged + cookieConsent unchanged → preserved (only the sections between hero and footer are replaced) ✓
- htmx-first live table, Alpine-only theme switch, no vanilla JS → Tasks 2–3 ✓
- templ-escaping-safe theme switch (data-attr + static `@click`) → Task 3 ✓
- both light/dark variants → every class block includes `dark:` ✓
- `templ generate` + Tailwind rebuild after edits → every task Step 4 / Task 6 Step 3 ✓
- E2E: sections render, CTA links, theme-switch changes `data-theme` (via `Evaluate`), table loads via HTMX, gallery links, no console errors → Tasks 1–6 ✓

**Placeholder scan:** none — all steps carry concrete code/commands. The one conditional ("verify badge constant names") instructs grepping the real source rather than inventing identifiers, with a concrete fallback.

**Type consistency:** `exampleApp`/`getExampleApps`, `themeChip`/`getThemeChips`, element IDs (`#hero`, `#playground`, `#home-table`, `#examples`, `#stack-strip`) are used identically across tasks and tests. `home-table` matches between the `table.Config{ID: "home-table"}` and the `#home-table tbody tr` selector. Routes in the gallery data match the routes asserted in Task 4.

**Out of scope (unchanged):** demo layout, sidebar, component packages, example apps, server routes.

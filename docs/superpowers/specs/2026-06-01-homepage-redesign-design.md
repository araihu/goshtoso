# Homepage Redesign — Design

**Date:** 2026-06-01
**Status:** Approved (design), pending spec review
**File touched:** `internal/pages/demo/components/landing.templ` (+ generated `landing_templ.go`)

## Problem

The current landing page (`landing.templ`) leads with **"The Stack"** — six cards
naming Go, Templ, Tailwind, Alpine, HTMX, PenguinUI — followed by a prose "Why?"
section. A first-time visitor learns *what Goshtoso is built from* before they
learn *what it does or why they'd use it*. The page tells; it doesn't show. For a
library that ships a live theme switcher and five runnable example apps, that's a
missed opportunity: the homepage should demonstrate the product, not describe its
ingredients.

## Direction (approved)

**"Show, don't tell."** Lead with what Goshtoso does, prove it with live
components on the page itself, and demote the stack to a single condensed strip.

## Page structure (top → bottom)

1. **Hero** — value-prop headline + CTAs; mascot shrunk to a supporting role.
2. **Live playground** (the centerpiece) — a theme-switcher chip row driving a
   cluster of *real* components + one live HTMX table.
3. **Example apps gallery** — Todo, Chat, Live Logs, Profile, Ticker.
4. **How it works** — three short steps + a small code peek.
5. **Stack strip** — Go · Templ · Tailwind · Alpine · HTMX as one condensed
   logo/link row (replaces the old six-card section).
6. **Footer** — stats + get-started + license + privacy (mostly as today).

## Section detail

### 1. Hero
- Headline: **"Build interactive UIs in Go."** Subhead: **"No JavaScript build
  step."** + one supporting sentence: server-rendered components, HTMX + Alpine
  for interactivity, ships as one binary.
- CTAs: **Browse components** (primary → `/components/button`), **GitHub**
  (secondary → repo). Keep a tertiary **Get started** link.
- Mascot (`goshtoso-art.png`): moved from large/centered to a smaller supporting
  element beside or above the headline (≈ half its current 256px width). It stays
  — it's brand identity — but no longer dominates.

### 2. Live playground (centerpiece)
- A horizontal, wrapping row of **theme chips** (subset of the 15 themes, e.g.
  minimal, modern, dracula, 90s, pastel, neo-brutalism + a "…all themes" link to
  `/docs/theme`). Clicking a chip calls the **same mechanism the demo layout
  already uses**: `document.documentElement.setAttribute('data-theme', name)` and
  persists to `localStorage` (mirror `setTheme` in `layout.templ:12`). Because
  themes are CSS-variable driven, the whole cluster recolors instantly with no
  re-render.
- The cluster below the chips renders **real Goshtoso components** (not images):
  a Button + Badge, a Toggle and/or Tabs, and a **live sortable/paginated table**
  backed by the existing endpoint `/api/components/table/rows` (HTMX). This is the
  literal proof: the homepage uses the library it documents.
- Reuse component entry points directly (`button.Button(cfg)`, `badge.…`,
  `toggle.…`, `tabs.…`, the table component). Pull the theme list from a small
  local slice mirroring `getThemeOptions()` (layout.templ:647) — keep it a curated
  subset for the hero, not all 15, to avoid a wall of chips.

### 3. Example apps gallery
- A responsive grid of cards, one per example app, linking to its route:
  `/examples/todo`, `/examples/chat`, `/examples/logs`, `/examples/profile`,
  `/examples/ticker`. Each card: title + one-line description + the components/
  pattern it showcases (e.g. Logs → "SSE server-push over htmx-ext-sse"). A
  "view all examples" affordance is optional.
- Purpose: proves the components compose into real apps, not just isolated demos.

### 4. How it works
- Three compact steps: ① write a `.templ` component → ② `hx-get` swaps a
  server-rendered fragment → ③ build one binary, no bundler. Optionally a tiny
  syntax-highlight-free code peek (a few lines of templ + an `hx-get`).
- This is the *only* "tell" section and stays short. (User kept it in; can be
  trimmed later if the page feels long.)

### 5. Stack strip
- Replaces the old "The Stack" six-card grid. One horizontal row: Go · Templ ·
  Tailwind CSS · Alpine.js · HTMX, each linking out to its site. PenguinUI
  attribution can live here or in the footer. Compact, single line on desktop.

### 6. Footer
- Keep the existing stat line ("23 components · 15 themes · 386 E2E tests · Get
  started") and the existing license/privacy links pattern. Update counts to
  current values at build time. The `cookieConsent()` template is **unchanged**.

## Technical constraints (carry over from CLAUDE.md)

- **Templ escaping:** any Alpine `x-data`/`@click` must use unquoted keys / single
  quotes — never `json.Marshal` into an attribute. The theme switcher uses simple
  inline expressions like the existing `x-init`; keep them quote-safe.
- **htmx first:** the live table is HTMX (`hx-get` → `/api/components/table/rows`),
  not client JS. The theme switch is local UI state → Alpine inline expression is
  correct (a server round-trip would feel wrong). No vanilla JS.
- **Fragment-nav / OOB:** the landing page is a standalone full HTML document
  (its own `<html>`, not the demo `renderDemo` shell), so the fragment-nav OOB
  gotcha does not apply here. Keep it a standalone doc as today.
- **Both light & dark variants** on every utility class; test across themes
  (especially Minimal — no border radius).
- Run `templ generate` + rebuild Tailwind (`tailwindcss -i css/main.css -o
  assets/styles.css`) after edits — CSS is embedded; any new utility class must be
  compiled in.

## Testing

- Extend/replace the landing E2E coverage: assert the new sections render, the
  hero CTAs link correctly, **clicking a theme chip changes
  `documentElement[data-theme]`** (read via `Evaluate`, not the static attr — per
  the Alpine testing note), the live table loads rows via HTMX, and the example
  gallery links point at the five `/examples/*` routes. Assert **no console
  errors** (the silent-Alpine-failure guard).
- Use `clickUntil` for any click that triggers an HTMX swap on the live table
  (rebind race — see `e2e-suite-flaky-full-run` memory).

## Out of scope

- No changes to the demo layout, sidebar, component packages, or example apps
  themselves — this is the landing page only.
- No new components. The playground reuses existing ones.
- No drag/drop, no new server endpoints (the table endpoint already exists).

## Open questions (resolve during implementation, low-risk)

- Exact curated theme subset for the hero chips (pick ~6 visually distinct ones).
- Whether the code peek in "How it works" is worth the vertical space — drop if
  the page runs long.

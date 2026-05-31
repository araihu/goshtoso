---
target: running app page — /docs/theme (Theme Settings)
total_score: 27
p0_count: 0
p1_count: 1
timestamp: 2026-05-31T02-17-19Z
slug: internal-pages-demo-components-theme-templ
---
# Impeccable Critique — Theme Settings page (`/docs/theme`)

**Register:** Product (this is a tool: a live theme customizer for developers).
**Evidence:** desktop + mobile headless screenshots (1440px / 390px), playwright-go layout probe, source review of `theme.templ` + `layout.templ`. Deterministic detector unavailable (bundled `detector/` missing) — manual + browser review only.

## Design Health Score

| # | Heuristic | Score | Key Issue |
|---|-----------|-------|-----------|
| 1 | Visibility of System Status | 3 | Strong: live preview, "Link Copied!", active-theme ring. No feedback on reset. |
| 2 | Match System / Real World | 3 | Token names ("On Surface Dark Strong", "Outline Dark") are system-speak; OK for devs. |
| 3 | User Control and Freedom | 2 | "Reset all" wipes every customization, no confirm, no undo. |
| 4 | Consistency and Standards | 3 | Reused Select shell + uniform section headers. Solid. |
| 5 | Error Prevention | 2 | Destructive reset is a one-click text link, no guard. |
| 6 | Recognition Rather Than Recall | 3 | Swatches + labels + live preview + contrast table. Wall of similar rows taxes it slightly. |
| 7 | Flexibility and Efficiency | 3 | Share-by-URL, single/multi CSS export, current/all filter, font lazy-load. |
| 8 | Aesthetic and Minimalist Design | 3 | Clean, but one very dense page; nested cards in showcase. |
| 9 | Error Recovery | 3 | Contrast table flags failing ratios in danger color. Few real error flows. |
| 10 | Help and Documentation | 2 | Intro + a couple hints, but no explanation of what each token actually controls. |
| **Total** | | **27/40** | **Solid, with clear gaps** |

## Anti-Patterns Verdict

**Does it look AI-generated? Mostly no.** This page shows real craft and resists the slop tells:
- Theme tiles use mini-skeleton previews (color dots + fake text bars + accent bar), not generic swatch squares.
- The border-radius picker draws a custom corner SVG per radius with hover tooltips.
- The two-column editor + sticky live showcase is a deliberate, non-template layout.

Two genuine anti-pattern hits:
- **Nested cards** (shared design ban). `themePreviewSection` is a bordered `bg-surface` card containing two more bordered `bg-surface-alt` cards (`theme.templ:1352` → `1355`/`1360`). Card-in-card.
- **Density approaching identical-grid fatigue.** The Colors section stacks ~10 near-identical dropdown rows; the page also carries 15 theme tiles + contrast table + CSS export blocks. A lot of same-shaped repetition on one scroll.

**Deterministic scan:** unavailable this run (detector bundle missing). Treat the above as LLM-only; no automated overlay was produced.

## Overall Impression

This is a competent, genuinely useful theme editor that mostly avoids the AI look. It is let down by one real bug and one structural problem: it **breaks at mobile width**, and it asks the user to absorb everything at once. Biggest single opportunity: fix the responsive clipping (it's a layout-wide bug, not page-local), then thin the cognitive load of the Colors section.

## What's Working

- **Live, sticky showcase.** Editing on the left, a pinned preview on the right (`lg:sticky lg:top-6`) that reacts as you change tokens. This is the right affordance and it's executed well.
- **Theme tiles with real micro-previews.** Each tile previews the theme's palette and shape, not just a name. Discoverable and on-brand.
- **Power-user export.** Single vs multiple themes, current vs all, `@layer base` wrapping, share-by-URL. Real depth for the developer persona.

## Priority Issues

### [P1] Mobile: page content is clipped, not scrollable, below ~430px
**Why it matters:** At 390px the document lays out **496px wide** (probe: `docW 496` vs `winW 390`). Because `<body>` is `overflow-hidden` (`layout.templ:98`), the extra ~106px is **cut off with no horizontal scroll to reach it** — the intro copy's right edge, the theme grid's entire second column, and the header controls are silently truncated. Driver is the **global header**, not this page: brand "Goshtoso PenguinUI" + theme-selector button + dark toggle in a `justify-between` row (`layout.templ:101–162`) don't fit and don't wrap/collapse/truncate, forcing min document width up. This hits **every page in the app**, not just the theme page.
**Fix:** Give the header room to collapse on small screens: hide the "PenguinUI" sub-label below `sm`, let the theme-selector button shrink (`min-w-0` + truncate the label, or icon-only under `sm`), and/or allow the controls row to wrap. Verify `document.documentElement.scrollWidth === innerWidth` at 360/390/414. Command: `adapt`.

### [P2] "Reset all customizations" is a one-click, unguarded, low-visibility destructive action
**Why it matters:** It wipes every font/radius/color override (and localStorage) instantly. It's rendered as small primary-colored text under the intro (`theme.templ:1144`), easy to hit, with no confirm and no undo. Violates Error Prevention + User Control.
**Fix:** Require confirmation (inline "Reset? Yes/Cancel", not a modal), and/or offer per-section reset. Consider a brief "Customizations cleared" status. Command: `harden`.

### [P2] Nested cards + single-page density in the showcase and Colors sections
**Why it matters:** Card-in-card reads as visual noise (banned), and the Colors section is a tall stack of ~10 visually identical dropdown rows that's hard to scan. Aesthetic/Minimalist + Recognition both pay for it.
**Fix:** Flatten the showcase to one frame with dividers/spacing instead of nested bordered cards. In Colors, group rows visually (surface vs primary vs feedback) or use a tighter two-column grid so the eye has anchors. Command: `distill`, then `layout`.

### [P2] Token jargon is unexplained
**Why it matters:** "On Surface Dark Strong", "On Primary Dark", "Outline Dark Strong" are Tailwind role names with no inline hint of what surface they paint. Fine for someone who already knows the system; opaque for a first-timer evaluating the library. Help/Documentation + Match-Real-World.
**Fix:** Add a one-line description or tooltip per role (or a small "what is this?" legend mapping roles to the showcase elements they affect). Command: `clarify`.

## Persona Red Flags

**Dana (developer evaluating the library):** Lands expecting to judge polish. Resizes the browser narrow (devs do) and watches the right column of themes and the header get chopped with no scrollbar. First impression of a "99.99% parity" UI library is a responsive bug. Also can't tell what "On Surface Dark All" maps to without scrolling to the showcase and guessing.

**Sam (on a phone):** Opens `/docs/theme` on a 390px device. Header brand/controls are clipped, the second theme column is unreachable, the localStorage toast overflows the right edge. Can browse the left column of themes but half the picker is invisible. High bounce risk.

**Priya (cautious customizer):** Spends ten minutes hand-tuning colors, then clicks the small "Reset all customizations" link thinking it resets one thing. Everything's gone, no confirm, no undo. Trust hit.

## Minor Observations

- Theme tiles double-signal color (primary+secondary dots *and* a primary accent bar); one is redundant.
- The contrast table is correctly scroll-contained (`overflow-x-auto`) — good, and explicitly *not* the overflow culprit.
- Longer theme names ("Neo Brutalism") in the header selector will widen the overflow further; whatever fix you pick must handle the longest label.
- `x-cloak` is used on the editor's dark-mode group — good; make sure the header dropdown and any other `x-show` blocks are equally cloaked to avoid first-paint flashes.

## Questions to Consider

- The page tries to be picker + editor + contrast auditor + code exporter at once. Would a tab or step split (Pick → Tune → Export) lower the load without losing power?
- What's the confident, one-glance version of the Colors section that doesn't read as 10 identical dropdowns?
- Is the showcase's nested framing earning its borders, or would whitespace alone separate the groups?

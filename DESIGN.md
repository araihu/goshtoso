---
name: Goshtoso
description: Go-native server-rendered UI components with local assets, themeable states, and practical interactivity.
colors:
  surface: "oklch(97% 0 0)"
  surface-alt: "oklch(92.2% 0 0)"
  surface-dark: "oklch(26.9% 0 0)"
  surface-dark-alt: "oklch(20.5% 0 0)"
  ink: "oklch(26.9% 0 0)"
  ink-strong: "#000"
  ink-muted: "oklch(37.4% 0 0)"
  primary: "oklch(62.7% 0.265 303.9)"
  primary-ink: "#fff"
  secondary: "oklch(68.5% 0.169 237.323)"
  secondary-ink: "#fff"
  outline: "oklch(87% 0 0)"
  outline-strong: "oklch(26.9% 0 0)"
  info: "oklch(71.5% 0.143 215.221)"
  success: "oklch(77.7% 0.152 181.912)"
  warning: "oklch(90.5% 0.182 98.111)"
  danger: "oklch(65.6% 0.241 354.308)"
typography:
  display:
    fontFamily: "var(--font-title)"
    fontSize: "clamp(2.25rem, 5vw, 3rem)"
    fontWeight: 700
    lineHeight: 1.1
    letterSpacing: "0"
  headline:
    fontFamily: "var(--font-title)"
    fontSize: "1.5rem"
    fontWeight: 700
    lineHeight: 1.2
    letterSpacing: "0"
  title:
    fontFamily: "var(--font-title)"
    fontSize: "1.125rem"
    fontWeight: 700
    lineHeight: 1.35
    letterSpacing: "0"
  body:
    fontFamily: "var(--font-body)"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "0"
  label:
    fontFamily: "var(--font-body)"
    fontSize: "0.875rem"
    fontWeight: 500
    lineHeight: 1.25
    letterSpacing: "0.025em"
rounded:
  none: "0"
  sm: "0.125rem"
  md: "0.375rem"
  lg: "0.5rem"
  xl: "0.75rem"
  theme: "0.75rem"
  button: "1rem"
spacing:
  xs: "0.25rem"
  sm: "0.5rem"
  md: "1rem"
  lg: "1.5rem"
  xl: "2rem"
  page-x: "1.5rem"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-ink}"
    rounded: "{rounded.button}"
    padding: "0.5rem 1rem"
    typography: "{typography.label}"
  button-secondary:
    backgroundColor: "{colors.secondary}"
    textColor: "{colors.secondary-ink}"
    rounded: "{rounded.button}"
    padding: "0.5rem 1rem"
    typography: "{typography.label}"
  card-surface:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.theme}"
    padding: "1.5rem"
  input-default:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.theme}"
    padding: "0.5rem 0.75rem"
    typography: "{typography.body}"
  badge-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-ink}"
    rounded: "{rounded.theme}"
    padding: "0.25rem 0.5rem"
    typography: "{typography.label}"
---

# Design System: Goshtoso

## 1. Overview

**Creative North Star: "The Go Workbench at the Beach"**

Goshtoso should feel like a real workbench for Go developers: sturdy, typed,
inspectable, and ready for application UI. The beach signal comes from the
Copacabana mascot and the warmer theme personality, but it should appear as
confidence and memorability, not as decoration pasted over technical docs.

The system is theme-first. Surfaces, accents, text, borders, feedback states, and
radius all flow through CSS custom properties so demos, examples, and consumer
apps can change clothes without changing component contracts. The default stance
is practical density: predictable navigation, visible examples, compact API
tables, and clear component states.

This design explicitly rejects generic SaaS landing pages, CDN-snippet galleries,
interchangeable icon-card catalogs, glossy framework marketing, and AI-looking
patterns unrelated to the Go-first product.

**Key Characteristics:**

- Token-driven themes with light and dark parity.
- Server-rendered examples that use the real component library.
- Rounded, friendly controls balanced by dense documentation layouts.
- Primary accent as an interaction signal, not a decorative wash.
- Accessibility visible through focus states, contrast checks, and reduced-motion
  fallbacks.

## 2. Colors

The palette is a live theme surface, not a fixed poster palette. Documented color
roles resolve through `var(--color-*)` tokens, with the default Goshtoso theme
using purple primary, sky secondary, neutral surfaces, and explicit semantic
feedback colors.

### Primary

- **Goshtoso Signal** (`primary`): The main action and selected-state color. Use
  it for primary buttons, active navigation, search highlights, and a small number
  of decisive links.

### Secondary

- **Server Sky** (`secondary`): Supporting accent for secondary actions and
  complementary emphasis. It should not compete with the primary action on the
  same surface.

### Tertiary

- **Feedback Set** (`info`, `success`, `warning`, `danger`): State colors for
  alerts, badges, validation, and status affordances. Each state has an explicit
  foreground token; never assume white text on saturated colors.

### Neutral

- **Component Surface** (`surface`): The default canvas for pages, cards, inputs,
  and tables.
- **Workbench Surface** (`surface-alt`): The second tonal layer for hover states,
  sidebars, panels, and grouped controls.
- **Readable Ink** (`ink`, `ink-strong`, `ink-muted`): Text hierarchy tokens. The
  muted role is derived to preserve contrast across shipped themes.
- **Tool Outline** (`outline`, `outline-strong`): Borders, dividers, active row
  markers, and control strokes.

### Named Rules

**The Token First Rule.** Use semantic color roles in components. Do not hardcode
one theme's palette into reusable UI.

**The Accent Scarcity Rule.** Primary is for action, active state, and orientation.
If every heading is primary, no element is primary.

**The Contrast Is Product Rule.** State and muted text must clear AA contrast
across themes; opacity-as-hierarchy is forbidden for body copy.

## 3. Typography

**Display Font:** `var(--font-title)` with theme-selected sans or display face.
**Body Font:** `var(--font-body)` with system sans fallback.
**Label/Mono Font:** system mono only inside code, snippets, and literal API
values.

**Character:** The type system is plainspoken and component-oriented. Theme
authors can change the voice, but Goshtoso pages should preserve clear scale,
short labels, and readable documentation before expressive flourish.

### Hierarchy

- **Display** (700, fluid 2.25rem to 3rem, 1.1 line-height): Homepage and major
  docs entry headings only.
- **Headline** (700, 1.5rem, 1.2 line-height): Page sections, component demo
  headings, and example app region titles.
- **Title** (700, 1.125rem, 1.35 line-height): Cards, table captions, and
  component subsection headings.
- **Body** (400, 1rem, 1.5 line-height): Documentation prose and form helper
  text. Keep long prose to roughly 65 to 75 characters.
- **Label** (500, 0.875rem, slight tracking): Button labels, field labels,
  compact metadata, and badge text.

### Named Rules

**The Literal Label Rule.** Buttons and links say what happens: "Save changes",
"Browse components", "View docs". Avoid clever CTA language.

**The Code Is Evidence Rule.** Use mono for code identifiers and snippets, not as
lazy technical atmosphere.

## 4. Elevation

Goshtoso is flat by default. Depth is conveyed primarily through tonal layers,
borders, active inset markers, and state changes. Shadows exist only when a
component needs a small physical cue, such as a popover, selected palette swatch,
or transient overlay.

### Shadow Vocabulary

- **State Inset** (`inset 2px 0 0 var(--color-outline-strong)`): Active search
  and navigation rows. It marks selection without adding a side-stripe accent.
- **Tiny Surface Shadow** (`var(--shadow-sm)` or equivalent): Small swatches and
  floating UI details only. Do not pair soft broad shadows with decorative card
  borders.

### Named Rules

**The Flat By Default Rule.** Cards, docs panels, and component demos rest on
borders and tonal contrast. If a shadow is visible at rest, it must explain a
layering relationship.

## 5. Components

### Buttons

- **Shape:** Friendly pill-leaning corners (1rem) with a 1px border.
- **Primary:** `primary` background, `primary-ink` text, medium label weight,
  compact padding (`0.5rem 1rem`), and theme-aware dark tokens.
- **Hover / Focus:** Hover lowers opacity for quick feedback. Focus uses a
  visible 2px outline with offset; active removes the offset. Disabled reduces
  opacity and blocks pointer intent.
- **Secondary / Alternate / Semantic:** Secondary uses the `secondary` role.
  Alternate uses the alternate surface for quieter actions. Info, success,
  warning, and danger map directly to semantic tokens.

### Chips

- **Style:** Badges and segmented controls use theme radius, compact padding, and
  either solid semantic color or soft tonal backgrounds.
- **State:** Selected state must change both color and structural affordance. Dots
  and icons are supporting signals, not the only signal.

### Cards / Containers

- **Corner Style:** Theme radius (`var(--radius-radius)`), usually 0 to 0.75rem
  depending on active theme.
- **Background:** Default surface with alternate surface for grouped or preview
  regions.
- **Shadow Strategy:** Flat at rest. Use borders and hover border color before
  reaching for shadow.
- **Border:** `outline` for passive separation, `primary` for actionable hover or
  selected state.
- **Internal Padding:** Common card padding is `1.5rem`; dense docs rows and
  controls may use `1rem`.

### Inputs / Fields

- **Style:** Theme surface, readable ink, outline border, theme radius, and
  compact vertical rhythm.
- **Focus:** Focus must be visible through border or outline changes and cannot
  rely on placeholder text disappearing.
- **Error / Disabled:** Error and success use semantic tokens with icons where
  helpful. Disabled fields reduce opacity but must remain legible.

### Navigation

- **Style:** Sticky header plus persistent sidebar on desktop, mobile sidebar
  below the header. Navigation copy is plain and searchable.
- **States:** Active rows use tonal background and a structural inset marker.
  Hover states use `surface-alt`; current route must remain clear in light and
  dark mode.
- **Mobile Treatment:** Layout owns responsive positioning; sidebar owns borders,
  background, scrolling, and flex structure.

### Search

Search is a command surface. Result rows should be dense, keyboardable, and
neutral by default. Matched text uses `search-highlight` and active rows use
outline-based structure rather than decorative saturated backgrounds.

## 6. Do's and Don'ts

### Do:

- **Do** compose public pages from real Goshtoso components whenever possible.
- **Do** keep every component themeable through semantic `var(--color-*)` and
  `var(--radius-radius)` roles.
- **Do** verify light mode, dark mode, the Goshtoso theme, and the Minimal theme
  before shipping a visual change.
- **Do** use examples, fragments, and live component states to explain the
  server-rendered model.
- **Do** preserve WCAG AA contrast, visible focus, keyboard paths, and
  reduced-motion fallbacks.

### Don't:

- **Don't** make Goshtoso look like a generic SaaS landing page.
- **Don't** make it feel like a CDN-snippet gallery or a component catalog made
  from interchangeable icon cards.
- **Don't** use glossy framework marketing, vague productivity claims, or
  decorative AI-looking patterns unrelated to the Go-first product.
- **Don't** let the Copacabana gopher personality become a gimmick that overwhelms
  technical trust.
- **Don't** hand-code one theme's values into reusable components; token drift is
  a product bug.

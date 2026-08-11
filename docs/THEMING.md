# Goshtoso Theming Reference

Goshtoso uses CSS custom properties (design tokens) for all colors, fonts, and border radii. Themes override these tokens via `[data-theme="name"]` selectors. All components reference these tokens — never hardcoded color values.

## Applying a Theme

Set the `data-theme` attribute on `<html>`:

```html
<html data-theme="modern">
```

Or switch at runtime:

```javascript
document.documentElement.setAttribute('data-theme', 'modern');
```

## Dark Mode

Add the `dark` class to `<html>`:

```html
<html data-theme="modern" class="dark">
```

Toggle at runtime:

```javascript
document.documentElement.classList.toggle('dark');
```

Components use both light and dark variants in their Tailwind classes:

```
bg-surface text-on-surface dark:bg-surface-dark dark:text-on-surface-dark
```

Goshtoso includes an Alpine.js dark mode store at `assets/js/darkmode.js` that persists the preference to `localStorage`.

## Token Reference

### Surface Colors

| Token | Purpose | Usage in Tailwind |
|-------|---------|-------------------|
| `--color-surface` | Main background | `bg-surface` |
| `--color-surface-alt` | Secondary background (headers, cards, alternating rows) | `bg-surface-alt` |
| `--color-surface-dark` | Main background (dark mode) | `dark:bg-surface-dark` |
| `--color-surface-dark-alt` | Secondary background (dark mode) | `dark:bg-surface-dark-alt` |

### Text Colors

| Token | Purpose | Usage in Tailwind |
|-------|---------|-------------------|
| `--color-on-surface` | Body text on surface backgrounds | `text-on-surface` |
| `--color-on-surface-strong` | Headings, emphasized text | `text-on-surface-strong` |
| `--color-on-surface-dark` | Body text (dark mode) | `dark:text-on-surface-dark` |
| `--color-on-surface-dark-strong` | Headings (dark mode) | `dark:text-on-surface-dark-strong` |

### Brand Colors

| Token | Purpose | Usage in Tailwind |
|-------|---------|-------------------|
| `--color-primary` | Primary action, links, active states | `bg-primary`, `text-primary` |
| `--color-on-primary` | Text on primary backgrounds | `text-on-primary` |
| `--color-secondary` | Secondary action, accents | `bg-secondary`, `text-secondary` |
| `--color-on-secondary` | Text on secondary backgrounds | `text-on-secondary` |
| `--color-primary-dark` | Primary (dark mode) | `dark:bg-primary-dark` |
| `--color-on-primary-dark` | Text on primary (dark mode) | `dark:text-on-primary-dark` |
| `--color-secondary-dark` | Secondary (dark mode) | `dark:bg-secondary-dark` |
| `--color-on-secondary-dark` | Text on secondary (dark mode) | `dark:text-on-secondary-dark` |

### Semantic / Status Colors

Base status colors are shared across modes and are intended for fills, borders,
and icons. For labels, helper/error copy, and other small text on a surface, use
the derived `*-text` pair; it mixes the tone into each mode's strongest readable
surface foreground instead of assuming one status color contrasts on both
backgrounds.

| Token | Purpose | Usage in Tailwind |
|-------|---------|-------------------|
| `--color-info` | Informational states | `bg-info`, `text-info` |
| `--color-on-info` | Text on info backgrounds | `text-on-info` |
| `--color-success` | Success states | `bg-success`, `text-success` |
| `--color-on-success` | Text on success backgrounds | `text-on-success` |
| `--color-warning` | Warning states | `bg-warning`, `text-warning` |
| `--color-on-warning` | Text on warning backgrounds | `text-on-warning` |
| `--color-danger` | Error/destructive states | `bg-danger`, `text-danger` |
| `--color-on-danger` | Text on danger backgrounds | `text-on-danger` |
| `--color-info-text` / `--color-info-text-dark` | Informational text on surfaces | `text-info-text dark:text-info-text-dark` |
| `--color-success-text` / `--color-success-text-dark` | Success text on surfaces | `text-success-text dark:text-success-text-dark` |
| `--color-warning-text` / `--color-warning-text-dark` | Warning text on surfaces | `text-warning-text dark:text-warning-text-dark` |
| `--color-danger-text` / `--color-danger-text-dark` | Error/destructive text on surfaces | `text-danger-text dark:text-danger-text-dark` |

Filled status actions use the derived `--color-*-action` /
`--color-on-*-action` pairs (plus `*-action-dark`) so their text, hover, and
focus states remain readable. Prefer `button.WithTone(button.ToneDanger)` over
assembling `bg-danger text-on-danger` yourself.

### Border & Outline

| Token | Purpose | Usage in Tailwind |
|-------|---------|-------------------|
| `--color-outline` | Default borders, dividers | `border-outline` |
| `--color-outline-strong` | Emphasized borders | `border-outline-strong` |
| `--color-outline-dark` | Default borders (dark mode) | `dark:border-outline-dark` |
| `--color-outline-dark-strong` | Emphasized borders (dark mode) | `dark:border-outline-dark-strong` |

### Typography

| Token | Purpose |
|-------|---------|
| `--font-title` | Headings, brand text |
| `--font-paragraph` / `--font-body` | Body text, UI labels |

### Layout

| Token | Purpose | Usage in Tailwind |
|-------|---------|-------------------|
| `--radius-radius` | Global border radius for all components | `rounded-radius` |

Tooltip panels use `rounded-radius` unless their public contract documents a
component-specific semantic radius token. Tooltip currently has no radius
exception: default, rich, and persistent Tooltip panels all resolve through
`--radius-radius`. The built-in Minimal theme maps that token to
`var(--radius-none)`, so Tooltip panels compute to exactly `0px` in both light
and dark mode.

### Backdrop and Elevation

Backdrop and shadow classes use semantic layering roles instead of palette
colors or Tailwind's generic shadow scale. These roles describe visual layer
provenance; they do not claim text or non-text contrast conformance.

| Token | Purpose | Usage in Tailwind |
|-------|---------|-------------------|
| `--color-backdrop` | Neutral overlay base; components retain their established alpha strength | `bg-backdrop/20`, `bg-backdrop/40`, `bg-backdrop/50`, `dark:bg-backdrop/60` |
| `--color-backdrop-surface` | Surface-derived overlay base used by Search in light mode | `bg-backdrop-surface/55` |
| `--shadow-elevation-control` / `--shadow-elevation-control-dark` | Compact control separation | `shadow-elevation-control dark:shadow-elevation-control-dark` |
| `--shadow-elevation-raised` | Raised panel separation | `shadow-elevation-raised` |
| `--shadow-elevation-overlay` | Highest overlay-panel separation | `shadow-elevation-overlay` |

Drawer, Modal, Sidebar, and Search backdrops resolve through the backdrop
roles. Drawer and Search panels use distinct raised/overlay elevation tiers;
Kbd uses the control tier. The default mappings preserve their previous layer
geometry, alpha, and light/dark separation while allowing custom themes to
override either base backdrop color or a complete elevation role.

## Built-in Theme Catalog

The root module publishes stable built-in keys and canonical design-system
labels through `github.com/araihu/goshtoso/themes`:

```go
import (
    "fmt"

    "github.com/araihu/goshtoso/themes"
)

for _, theme := range themes.BuiltIn() {
    fmt.Printf("%s: %s\n", theme.Key, theme.Label)
}
```

`themes.BuiltIn()` returns caller-owned values in deterministic key order. That
traversal order supports reproducible serialization; it is not a required
presentation order and does not select a default. Consumers own selector
presentation order, defaults, and custom themes. A shell may render different
presentation copy for a label or omit a built-in option without changing this
design-system catalog.

The catalog intentionally contains keys and labels only. CSS tokens, fonts,
colors, radii, and other styling details remain CSS and consumer concerns.
The built-in Arai Hû theme uses Modern typography and corner geometry (`Lato`
and `--radius-sm`) with the Arai Hû semantic palette. Its fallback copy is
byte-identical to `araihu/assets` release `v0.2.1` at commit
`fdfb1c2aad8fa61779e7b8c6f208e52a6cf825ce`.

| Key | Canonical label |
|-----|-----------------|
| `90s` | 90s |
| `araihu` | Arai Hû |
| `arctic` | Arctic |
| `christmas` | Christmas |
| `dracula` | Dracula |
| `goshtoso` | Goshtoso |
| `halloween` | Halloween |
| `high-contrast` | High Contrast |
| `industrial` | Industrial |
| `minimal` | Minimal |
| `modern` | Modern |
| `neo-brutalism` | Neo Brutalism |
| `news` | News |
| `pastel` | Pastel |
| `prototype` | Prototype |
| `zombie` | Zombie |

## Creating a Custom Theme

Add a new `[data-theme="your-theme"]` block in your CSS that overrides the tokens:

```css
@layer base {
    [data-theme="your-theme"] {
        --font-body: 'Your Font', sans-serif;
        --font-title: 'Your Font', sans-serif;

        /* Light */
        --color-surface: var(--color-white);
        --color-surface-alt: var(--color-gray-100);
        --color-on-surface: var(--color-gray-700);
        --color-on-surface-strong: var(--color-black);
        --color-primary: var(--color-blue-600);
        --color-on-primary: var(--color-white);
        --color-secondary: var(--color-indigo-600);
        --color-on-secondary: var(--color-white);
        --color-outline: var(--color-gray-300);
        --color-outline-strong: var(--color-gray-800);

        /* Dark */
        --color-surface-dark: var(--color-gray-900);
        --color-surface-dark-alt: var(--color-gray-800);
        --color-on-surface-dark: var(--color-gray-300);
        --color-on-surface-dark-strong: var(--color-white);
        --color-primary-dark: var(--color-blue-400);
        --color-on-primary-dark: var(--color-black);
        --color-secondary-dark: var(--color-indigo-400);
        --color-on-secondary-dark: var(--color-black);
        --color-outline-dark: var(--color-gray-700);
        --color-outline-dark-strong: var(--color-gray-300);

        /* Status (shared light/dark) */
        --color-info: var(--color-sky-500);
        --color-on-info: var(--color-white);
        --color-success: var(--color-green-500);
        --color-on-success: var(--color-white);
        --color-warning: var(--color-amber-500);
        --color-on-warning: var(--color-black);
        --color-danger: var(--color-red-500);
        --color-on-danger: var(--color-white);

        /* Layout */
        --radius-radius: var(--radius-md);
    }
}
```

Every token must be defined — components reference all of them. Use Tailwind's built-in color palette variables (e.g., `var(--color-blue-600)`) or raw hex values.

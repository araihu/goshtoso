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

Shared across light and dark modes (no `-dark` variants):

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

## Available Themes

| Theme | Font | Primary | Radius | Description |
|-------|------|---------|--------|-------------|
| *(default)* | Poppins / Inter | Purple | xl | Default PenguinUI theme |
| `arctic` | Inter | Blue | lg | Cool blue professional |
| `minimal` | Montserrat | Black | none | Clean, no rounded corners |
| `modern` | Lato | Black | sm | Subtle, professional |
| `high-contrast` | Inter | Dark Sky | sm | Maximum readability |
| `neo-brutalism` | Space Mono / Montserrat | Violet | none | Bold, graphic style |
| `halloween` | Poppins / Denk One | Orange | xl | Orange & purple festive |
| `zombie` | Montserrat / Denk One | Orange | xl | Violet-tinted Halloween variant |
| `pastel` | Playpen Sans | Rose | xl | Soft, warm tones |
| `90s` | Poppins / Oswald | Purple | xl | Retro feel |
| `christmas` | Lato / Jost | Red | md | Red & green festive |
| `prototype` | Playpen Sans | Black | none | Wireframe/sketch style |
| `news` | Inter / Merriweather | Sky | sm | Editorial, serif headings |
| `industrial` | Poppins / Oswald | Amber | none | Bold, utilitarian |
| `dracula` | Fira Code | Purple (#bd93f9) | md | Developer-focused dark theme |

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

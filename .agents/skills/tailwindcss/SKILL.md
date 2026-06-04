---
name: tailwindcss
description: Use when writing, debugging, or reviewing Tailwind CSS utilities, generated CSS drift, theme tokens, responsive variants, dark mode variants, or Goshtoso Tailwind v4 build behavior.
---

# Tailwind CSS In Goshtoso

Use this skill whenever a change touches Tailwind utility classes, `css/main.css`,
theme tokens, responsive behavior, dark mode styles, generated CSS, or the
Tailwind build.

## Build Contract

- Tailwind CSS is v4 and uses the pinned standalone binary version from
  `assets/tailwind.version`.
- Run `just css` after editing CSS, theme sources, or introducing new utility
  classes in templ/Go code.
- `just css` regenerates the theme source and `assets/styles.css`.
- `assets/styles.css` is generated. Never hand-edit it.

## Theme And Dark Mode

- Use semantic theme tokens instead of one-off colors when possible.
- Include both light and dark variants for visible surfaces and text:
  `bg-surface text-on-surface dark:bg-surface-dark dark:text-on-surface-dark`.
- Test multiple themes for UI changes, especially Minimal because it removes
  border radius.

## Utility Authoring

- Prefer existing component class helpers and local patterns before introducing
  new utility combinations.
- Keep state classes explicit and scan-friendly.
- For fixed-format controls, give stable dimensions with explicit sizing,
  aspect ratios, grid tracks, or min/max constraints so hover, labels, icons,
  and dynamic content do not shift layout.
- Do not scale font size with viewport width.
- Keep letter spacing at `0` unless the existing component pattern requires
  otherwise.

## Verification

For Tailwind-affecting changes, run:

```bash
just css
go build -o bin/server ./site/cmd/server
```

Use focused tests or screenshots when the change affects layout, themes,
responsive behavior, or component states.

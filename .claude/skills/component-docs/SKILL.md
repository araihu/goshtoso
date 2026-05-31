---
name: component-docs
description: Use when creating or editing a Goshtoso component demo/documentation page under internal/pages/demo/components/ — adding a new component page, adding or restructuring variant examples, adding an API reference, or wiring the right-rail "On this page" table of contents. Covers demo.ComponentDemo, demo.DemoSection, demo.APIReference, the data-toc-heading anchor contract, and the #<component>-fragment e2e anchor.
---

# Goshtoso Component Docs Page Pattern

## Overview

Every component demo page mirrors [penguinui.com](https://penguinui.com) layout: a header, then **one preview box + one code box per variant**, an **API reference table at the bottom**, and a **right-rail "On this page" TOC** auto-built from section headings. This pattern is mandatory for all 22+ components so docs stay consistent.

Helpers live in `internal/pages/demo/component_demo.templ`. The page file lives in `internal/pages/demo/components/<name>.templ`.

## The Rule

One variant = one `DemoSection` (its own preview frame + its own code block). Never stack multiple variants inside a single preview with one combined code blob. The first/primary variant uses `ComponentDemo`; every additional variant uses `DemoSection`; the API table is `APIReference` at the very bottom.

## Page Skeleton

```go
templ <Name>DemoPage() {
    @demo.Layout("<Title>", "<active-key>", <name>DemoContent())
}

templ <name>DemoContent() {
    <div id="<name>-fragment">
        @demo.ComponentDemo(
            demo.ComponentDemoProps{Title: "<Title>", Description: "..."},
            <name>DefaultPreview(),
            `// Go usage example for the default variant`,
        )
        @demo.DemoSection(
            demo.DemoSectionProps{Title: "<Variant>", Description: "..."},
            <name><Variant>Preview(),
            `// Go usage for this variant`,
        )
        // ...one DemoSection per remaining variant...
    </div>
    // API reference goes OUTSIDE the fragment wrapper (see Common Mistakes).
    @demo.APIReference([]demo.PropDoc{
        {Name: "...", Type: "...", Default: "...", Description: "..."},
    })
}
```

Each `*Preview()` templ wraps its component in a centering container, e.g. `<div class="w-full max-w-2xl mx-auto">`, and gives every variant container a **stable, unique ID** (`<name>-default`, `<name>-split`, ...) so e2e tests target the right box. Never leave the default `id="<component>"` on multiple instances (duplicate IDs).

## Right-Rail TOC (automatic — usually nothing to do)

The rail in `layout.templ` builds itself from `[data-toc-heading]` elements inside `#main-content`. `ComponentDemo`'s `<h1>`, `DemoSection`'s `<h2>`, and `APIReference`'s `<h2>` already carry `data-toc-heading` + a `slugify(Title)` id + `scroll-mt-8`. So: use the helpers and headings/anchors/scroll-spy just work. Deep links like `/components/<name>#<variant-slug>` resolve automatically. The rail hides below `xl` and when a page has <2 headings.

## Workflow Checklist

1. Create `internal/pages/demo/components/<name>.templ` using the skeleton.
2. One `DemoSection` per variant; give each a unique container ID.
3. `APIReference` at the bottom, **outside** `#<name>-fragment`.
4. Register route (`server.go`) + sidebar entry (`layout.templ`) per CLAUDE.md.
5. `templ generate` (force-regen if "0 updates": `rm <name>_templ.go && templ generate`).
6. **Rebuild Tailwind** if you introduced any new utility class: `tailwindcss -i css/main.css -o assets/styles.css` — CSS is embedded, so also `go build`. Forgetting this is why a class like `xl:block` silently does nothing.
7. E2E: target `#<name>-fragment` and the per-variant container IDs; `templ generate && go build && go test ./tests/e2e/... -run Test<Name>`.

## Common Mistakes

| Mistake | Consequence | Fix |
|---|---|---|
| API table inside `#<name>-fragment` | Its `<tbody class="divide-y">` pollutes `.divide-y` selectors; `.Last()` grabs the table, not a variant | Put `@demo.APIReference(...)` outside the fragment wrapper |
| All variants in one preview + one code blob | Breaks the per-variant pattern; no deep links | One `DemoSection` per variant |
| New Tailwind class without rebuilding CSS | Class missing from generated/embedded `assets/styles.css`; rail/styles silently absent | Rebuild Tailwind + `go build` |
| Reusing default `id` on multiple component instances | Duplicate DOM IDs; fragile tests | Unique `<name>-<variant>` IDs |
| Hand-editing `*_templ.go` or `assets/styles.css` | Overwritten on regen | Edit `.templ`/`css/main.css`, regenerate |

## Reference

- Helpers: `internal/pages/demo/component_demo.templ` (`ComponentDemo`, `DemoSection`, `APIReference`, `slugify`).
- Rail + scroll-spy: `internal/pages/demo/layout.templ` (`#toc-rail`, `window.buildTOC`, `data-toc-heading`).
- Canonical example: `internal/pages/demo/components/accordion.templ`.

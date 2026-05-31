---
name: using-goshtoso
description: Use when consuming Goshtoso UI components in Go/templ code — importing a component, building its config struct, rendering it in a page, wiring HTMX/Alpine, or looking up a component's props/entry point. Covers the component package layout, the config contract, the generated per-component reference, head Dependencies, and the escaping pitfalls that silently break Alpine.
---

# Using Goshtoso Components

## Overview

Goshtoso (Go + templ + Tailwind v4 + HTMX + Alpine) is a typed UI component
library. **The shape is consistent:** each component is a Go package under
`components/<name>/` exporting one (or a few) `templ` entry function(s) that take
a config struct defined in that package's `types.go`. The exact entry name,
config type name, and fields vary per component — see the per-component
reference below; don't assume from one component to the next.

Module path: `github.com/araihu/goshtoso`. Components are pure templ — they
render HTML, they do not run a server. You compose them inside your own
`.templ` pages.

**For building a component's demo/docs page, use the `component-docs` skill
instead. This skill is about consuming components in application code.**

## The contract

1. Import the component package by its directory path.
2. Build its config struct.
3. Render with `@pkg.Entry(pkg.ConfigType{...})`. The zero value is a usable
   default for most components; required data (options, items, rows) is explicit.

```go
import (
    "github.com/araihu/goshtoso/components/button"
    "github.com/araihu/goshtoso/components/accordion"
)

templ MyPage() {
    @button.Button(button.Config{
        Variant: button.Primary,
        Size:    button.SizeMedium,
        Type:    "submit",
    }) {
        Save changes        // children → button label
    }

    @accordion.Accordion(accordion.AccordionConfig{
        Items: []accordion.AccordionItem{
            {ID: "a1", Title: "Question?", Content: myContent()},
        },
    })
}
```

### Two facts that are NOT uniform — always check the reference

- **Config type name varies.** Most packages call it `Config` (`button.Config`,
  `alert.Config`), but several use a prefixed name (`accordion.AccordionConfig`,
  `checkbox.GroupConfig`). Some packages export several entry points with
  different config types.
- **Package name can differ from the directory.** Notably
  `components/select` declares `package selectfield`, so you import the `select`
  path but write `@selectfield.Select(...)`.

### Common (but not guaranteed) field conventions

When present, these fields behave consistently — but **not every component has
them, and HTMX/Alpine wiring differs per component:**

| Field | Meaning |
|-------|---------|
| `Variant` | typed string constant (`button.Primary`, `accordion.Split`). Empty = default. |
| `Size` | typed constant (`SizeSmall`/`SizeMedium`/`SizeLarge`/`SizeXLarge`) where offered. |
| `ID` | element id (accessibility wiring, HTMX targets). |
| `Class` | extra Tailwind classes appended to the root. |
| `Content` / item fields | `templ.Component`, not string — pass a `templ` func or `templ.Raw(...)`. |

**HTMX/Alpine wiring has no single shape.** Different components expose it
differently — e.g. `button.Config` takes pointer `HTMX *HTMXConfig` /
`Alpine *AlpineConfig`; the `select` package's `Config` instead takes string
fields (`AlpineModel`, `AlpineBindDisabled`) plus an `Attrs templ.Attributes`
escape hatch for arbitrary `hx-*`. Some components have no HTMX/Alpine fields.
**Do not assume — look the component up.**

**The authoritative, always-current per-component API (entry points, config
types, every field, enum values) is `components-reference.md` in this skill
directory.** It is generated from source — trust it over memory. To find raw
fields directly, read `components/<dir>/types.go`; every field is doc-commented.

## Page setup — load dependencies once

Put the asset tags in your `<head>` via the `head` package:

```go
import "github.com/araihu/goshtoso/components/head"

templ Layout() {
    <html>
        <head>
            @head.Dependencies()        // Alpine + collapse/focus plugins + HTMX + Tailwind
            // or @head.DependenciesMinimal() — Alpine core + HTMX + Tailwind only
        </head>
        ...
    </html>
}
```

`Dependencies()` uses CDN tags. The Goshtoso repo itself bundles Alpine/HTMX
locally under `assets/js/vendor/` and ships a prebuilt `assets/styles.css`; in
your own app, serve those or use the CDN tags above.

## Build steps (after editing .templ or CSS)

```bash
templ generate                                    # .templ → _templ.go (REQUIRED)
tailwindcss -i css/main.css -o assets/styles.css  # only if you added utility classes
```

Never edit `*_templ.go` or `assets/styles.css` by hand — both are generated.

## Critical pitfalls — Alpine + templ escaping

These break components **silently** (no console error, unit tests pass, browser
fails). For depth, use the `templ` and `htmx` skills. The rules:

- **Never `json.Marshal` data that lands in an HTML attribute.** templ escapes
  `"` → `&quot;`; Alpine then sees broken syntax. Build JS with single-quoted
  string builders instead.
- **Guard null arrays.** `json.Marshal([]T(nil))` → `null`, and Alpine
  `.includes()` throws on null. Force `[]` when the marshalled value is `"null"`.
- **Complex Alpine** (functions/strings) → register via `<script>` +
  `Alpine.data()` with `templ.Raw`, not inline `x-data`.
- **Simple Alpine** → unquoted keys, no string literals: `{ open: false, n: 0 }`.
- Symptom of a hit: rendered HTML shows `&quot;` inside `x-data`/`hx-vals`.

## Component catalog

**The full catalog — every package, entry point, config struct, field, and enum
— lives in [`components-reference.md`](components-reference.md), generated from
the `components/` source.** Read it to look up any component. It covers every
public package under `components/`, including nested ones (e.g. `combobox/v2`,
whose package is `v2` → `@v2.Combobox(...)`).

The **table** is the most complex: sortable columns, HTMX OOB-swap pagination,
filter bar/inline variants, infinite scroll. Its HTMX endpoint contract is
`/api/components/table/rows` with params `order_by`, `order_dir`, `page`,
`per_page`, `search`, plus filter keys. See `components/table/types.go` and
`internal/server/table_handler.go`.

## Keeping this skill accurate

`components-reference.md` is **generated** — never hand-edit it. Regenerate
after any change to a component's `types.go` or entry points:

```bash
go run ./scripts/skillgen
```

The repo enforces this: the pre-commit hook regenerates and re-stages it when a
component's Go files are staged, and CI fails if it is stale (the "using-goshtoso
skill reference drift check" step in the `lint-build` job). If you add, rename,
or change a component, run the generator and commit the updated reference
alongside.

## Theming

Themes are `[data-theme="name"]` selectors in `all-themes.css` (13 themes);
default is **Minimal** (black/white, no border radius). Dark mode toggles the
`.dark` class on `<html>`. Always pair light + dark utilities on custom classes:
`bg-surface dark:bg-surface-dark text-on-surface dark:text-on-surface-dark`.

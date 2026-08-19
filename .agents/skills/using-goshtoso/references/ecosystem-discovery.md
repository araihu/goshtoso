# Goshtoso Ecosystem Discovery

Run this pass before implementing custom HTML, CSS, JavaScript, or a new templ
primitive. Search the consumer's selected versions, not an unrelated checkout
or an unreleased `main` branch.

## Contents

- [Verified public baseline](#verified-public-baseline)
- [Evidence order](#evidence-order)
- [Goshtoso components and patterns](#1-goshtoso-components-and-patterns)
- [Goshtoso Charts](#2-goshtoso-charts)
- [Margo](#3-margo)
- [Goshtoso App Shells](#4-goshtoso-app-shells)
- [Icons and Iconpack](#5-icons-and-iconpack)
- [Fallback boundary](#fallback-boundary)

## Verified Public Baseline

This reference was reconciled on 2026-08-19 against Goshtoso `v0.2.5`,
Goshtoso App Shells `v0.1.6`, Margo `v0.0.6`, and Goshtoso Charts `v0.0.2`
module tag. These versions describe the review baseline, not a command to
downgrade or upgrade. Recheck public versions and honor the consumer's `go.mod`
before using an API.

## Evidence Order

1. Existing consumer code and `go.mod` version selections.
2. Public docs and APIs in the selected module versions.
3. Current public releases when adding a new dependency.
4. Repository examples only as usage evidence, never as copyable public API.

Start with the consumer's module graph:

```bash
go list -m all | rg '^github.com/araihu/(goshtoso|goshtoso-charts|goshtoso-app-shells|margo)( |$)'
go list -m -versions github.com/araihu/goshtoso
go list -m -versions github.com/araihu/goshtoso-charts
go list -m -versions github.com/araihu/goshtoso-app-shells
go list -m -versions github.com/araihu/margo
```

After a module is selected, inspect its installed public surface:

```bash
go doc github.com/araihu/goshtoso/components/button
go doc github.com/araihu/goshtoso-charts/components/line
go doc github.com/araihu/goshtoso-app-shells/consoleshell
go doc github.com/araihu/margo
```

Do not run `go get` merely to browse. Add a dependency only after choosing its
surface, then retain the selected version in `go.mod` and test that version.

Record the decision before implementation:

```text
Need:
Selected module versions:
Searched public surfaces:
Closest supported surface:
Decision: reuse | compose | gap
Custom HTML, CSS, or JavaScript needed because:
```

An empty or generic "not suitable" reason is not a gap. Name the missing
behavior or contract.

## 1. Goshtoso Components and Patterns

Search sibling `components-reference.md` for exact component packages, entry
points, options, config fields, and enums. Read sibling
`application-patterns.md` for page composition. Use public packages under
`github.com/araihu/goshtoso/components/...`; never import `site/` or `internal/`.

Check core surfaces before creating equivalents:

- structure and navigation: `appshell`, `navbar`, `sidebar`, `breadcrumbs`,
  `pageheader`, `tabs`, `steps`, and `pagination`;
- actions and overlays: `actiongroup`, `button`, `link`, `dropdown`, `popover`,
  `splitbutton`, `drawer`, `modal`, `toast`, and `tooltip`;
- data and state: `table`, `toolbar`, `search`, `emptystate`, `skeleton`,
  `spinner`, `badge`, `alert`, and `panel`;
- forms: `form`, `textinput`, `textarea`, `select`, `combobox`, `checkbox`,
  `radio`, `range`, `fileinput`, `schemaform`, and `structuredinput`;
- content: `card`, `codeblock`, `inlinecode`, `icon`, and `scrollregion`.

Prefer supported options, attrs, slots, and app-owned wrappers over selectors
against private DOM. Compose components when no single component owns the full
surface.

## 2. Goshtoso Charts

Use `github.com/araihu/goshtoso-charts` instead of hand-building SVG, canvas,
chart controls, export behavior, or renderer JavaScript.

Prefer static/vector packages when exploration is unnecessary: `line`, `bar`,
`pie`, `scatter`, `radar`, `candlestick`, `funnel`, and `heatmap`. Static charts
render accessible server-side SVG and still need adjacent exact data when users
must read precise values.

Use the selected release's interactive API only for real browser exploration,
large interactive surfaces, 3D, maps, graphs, or live data. Charts `v0.0.2`
exposes convenience constructors in `components/interactive` and focused APIs
under `components/interactive/<type>`. Inspect the selected release before
importing. Mount the Charts asset handler separately from Goshtoso. Interactive
charts also render `components/dependencies.Dependencies()`; local vendored
runtime is the default and CDN delivery is explicit.

Use `chartcontrol` and `charttheme` instead of custom expand, fullscreen,
export, or palette code. Keep renderer types out of application contracts.

## 3. Margo

Use `github.com/araihu/margo` for Markdown-backed content rather than writing a
Markdown renderer, document shell, PDF pipeline, slide runtime, or static-site
link validator.

- root `margo`: compile, check, render, and produce standalone HTML;
- `margo/site`: deterministic linked multi-page sites;
- `margo/ssg`: layout-neutral frame, shell, binding, and resource contracts for
  extensible static sites;
- `margo/deck`: accessible HTML presentation decks;
- `margo/pdf` plus an explicit engine: PDF contracts and export;
- `margo/charts`: optional static Goshtoso chart fences;
- `cmd/margo`: CLI workflows for `check`, `html`, `site`, `pdf`, and `deck`.

Margo owns document rendering. Applications retain URLs, navigation, storage,
authorization, and deployment. Supported packages belong to the root Margo
module; do not add historical nested-module requirements.

## 4. Goshtoso App Shells

Use `github.com/araihu/goshtoso-app-shells` before creating page-frame HTML,
responsive navigation JavaScript, theme bootstrap, HTMX shell lifecycle, or
duplicated shell asset CSS.

- `landingshell`: product and organization landing pages;
- `consoleshell`: operations consoles and server-rendered HTMX products;
- `componentdocshell`: component docs, API references, and design systems;
- `componentpage`: repeated component-reference page structure inside a shell.

Mount each shell's public asset handler at its documented prefix. Let the shell
own its head/frame, first-paint color mode, responsive navigation, and fragment
lifecycle. Supply product content, navigation data, domain actions, metadata,
and art direction through public configs and slots. Never copy shell assets or
recreate their runtime in consumer code.

## 5. Icons and Iconpack

Use bundled typed Heroicons first. For product-specific or third-party icons,
use `github.com/araihu/goshtoso/cmd/iconpack`; do not paste raw SVG or add names
to Goshtoso's embedded package.

Prefer `.iconpack.yaml` for GitHub trees, single SVG files, or multiple source
packs. First trust requires `-trust`; review and commit `.iconpack.lock.yaml`,
generated manifest, provenance, and licenses. Later generation omits `-trust`
and verifies identical source bytes. Serve the generated sprite at the exact
configured URL and render through the generated typed helper, which delegates
to `components/icon`.

Use Arai Hû Assets release mode only with exact release/archive identities from
an accepted public release. Keep legacy source-manifest and `iconcatalog` flows
for existing migrations; they are not the default for new icon packs.

## Fallback Boundary

When discovery finds no complete solution:

1. Compose supported components, chart primitives, or shell slots.
2. Add small semantic app-owned markup for domain content.
3. Use HTMX for server-owned interaction and fragments.
4. Use Alpine.js for local state and instant feedback.
5. Write vanilla JavaScript only for a named behavior absent from those APIs.

Keep the custom layer narrow. Do not fork public component markup, target
private DOM, duplicate shell assets, reimplement chart runtimes, or copy demo
site code. Add a focused test that proves the ecosystem gap and the fallback
behavior.

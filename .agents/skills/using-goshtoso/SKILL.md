---
name: using-goshtoso
description: Use when integrating or updating Goshtoso in an external Go/templ application, including installing github.com/araihu/goshtoso, serving bundled assets, wiring head.Dependencies, importing components, choosing a Tailwind CSS strategy, or debugging missing Goshtoso styles, Alpine.js, HTMX, combobox behavior, or component config fields.
---

# Using Goshtoso

## Overview

Goshtoso is a Go UI component library for server-rendered templ apps. Treat it
as a library dependency: consumers import pre-generated components, serve the
embedded assets, and run `templ generate` only for their own `.templ` files.
This skill is for integrating Goshtoso into external applications, not for
modifying the Goshtoso component library or demo site itself.

## Default Integration

Use this path unless the app deliberately owns a custom Tailwind build.

```bash
go get github.com/araihu/goshtoso@latest
go get github.com/a-h/templ
go install github.com/a-h/templ/cmd/templ@latest
templ generate
```

Mount Goshtoso assets directly at `/assets/` with a method-qualified Go
`ServeMux` pattern:

```go
import "github.com/araihu/goshtoso/assets"

mux := http.NewServeMux()
mux.Handle("GET /assets/", assets.Handler())
```

Do not wrap `assets.Handler()` in `http.StripPrefix`; it already strips
`/assets/`.

In the page shell, prefer the `head` component so CSS and JavaScript paths stay
in sync with the library version:

```templ
import "github.com/araihu/goshtoso/components/head"

templ Layout() {
	<html>
		<head>
			@head.Dependencies()
		</head>
		<body>{ children... }</body>
	</html>
}
```

`head.Dependencies()` emits Goshtoso CSS, Alpine.js plus the bundled collapse
and focus plugins, HTMX, and first-party component scripts such as
`combobox.js`.

## Rendering Components

Import from `github.com/araihu/goshtoso/components/<name>` and call the exported
templ component from your own `.templ` files:

```templ
import "github.com/araihu/goshtoso/components/button"

templ Example() {
	@button.Button(
		button.WithTone(button.TonePrimary),
		button.WithType("button"),
	) {
		Save changes
	}
}
```

Read `references/components-reference.md` when you need exact component import
paths, package names, entry points, config fields, enum values, or package names
that differ from the directory name.

Read the public
[component model](https://github.com/araihu/goshtoso/blob/main/docs/COMPONENT_MODEL.md)
before choosing between constructors or configuration fields. It documents the
common component interface, concrete return values, constructor styles, stable
Kind identity, and rendered defaults.

## From First Component to Application

Do not invent the page around isolated components. Choose the closest task
pattern and read `references/application-patterns.md` before composing it:

- **App Shell** for persistent top navigation, sidebar, global search, and one
  main scroll region.
- **Operations List** for page tools, filters, Table, and loading, empty, error,
  and success states.
- **Detail Workspace** for identity, status, tabs, actions, and a secondary
  detail rail.
- **Multi-step Workflow** for Steps, Form, server validation, review, and safe
  submission.

Keep domain vocabulary, authorization, data priority, and workflow rules in the
application. Goshtoso supplies the component vocabulary and supported layout
contract, not the product decisions.

Before declaring the surface finished, apply
`references/visual-acceptance.md`. It requires 390 px and 1440 px checks,
Goshtoso and Minimal themes, light and dark modes, the full state matrix,
keyboard use, console checks, and accessibility scanning.

## CSS Strategy

Prefer the embedded stylesheet served by `assets.Handler()` for components and
official recipes. Stock CDN Tailwind does not include Goshtoso's theme tokens or
component classes.

The embedded stylesheet is not a general Tailwind compiler for consumer markup.
Hooks such as `RootClass` accept class names, but an arbitrary Tailwind utility
can fail silently when that selector was not emitted. The official application
patterns list the small guaranteed layout set. If the app needs other Tailwind
utilities, build an app stylesheet and load it after `/assets/styles.css`.

For apps with their own Tailwind build, keep Goshtoso CSS as a separate
stylesheet unless a unified build is truly needed:

```html
<link rel="stylesheet" href="/assets/styles.css">
<link rel="stylesheet" href="/css/app.css">
```

If a unified build is required, use the `goshtoso` CLI to extract theme source
and print the component source path for Tailwind scanning:

```bash
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -theme -out=css/goshtoso-theme.css
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -source-path
```

## Integration Checks

- Missing styling usually means `/assets/styles.css` is not served or
  `head.Dependencies()` is absent.
- Dead combobox keyboard navigation usually means `/assets/js/combobox.js` is
  missing; prefer `head.Dependencies()` instead of hand-written script tags.
- Alpine plugins must load before Alpine core. Avoid manual tags unless the app
  has a strong reason.
- HTMX handlers should return rendered HTML fragments, not JSON.
- Goshtoso's own components are pre-generated. Do not run `templ generate`
  against the module cache or vendor copy; run it for the consumer app's files.
- `.templ` files can use `templ.URL`, `templ.KV`, and other templ helpers without
  explicitly importing `github.com/a-h/templ`; the generator injects that import
  and an explicit duplicate can fail generation.
- Do not import `site/`; it is the Goshtoso demo app, not public library API.

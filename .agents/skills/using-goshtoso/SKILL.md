---
name: using-goshtoso
description: Use when building, designing, redesigning, integrating, or updating an external Go/templ application with Goshtoso, including application shells, dashboards, operations lists, detail pages, settings, onboarding, workflows, public content, component selection, visual direction, state design, installing github.com/araihu/goshtoso, serving bundled assets, wiring head.Dependencies, Tailwind CSS strategy, or debugging Goshtoso styles, Alpine.js, HTMX, and component APIs.
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
Goshtoso requires **Go 1.26.5 or newer**.

```bash
go get github.com/araihu/goshtoso@latest
go get github.com/a-h/templ
go install github.com/a-h/templ/cmd/templ@latest
templ generate
go mod tidy
go test ./...
```

Create the consumer's `.templ` file before `templ generate`. Run
`templ generate` before `go mod tidy`: generated Go is what makes the templ
runtime import visible, while tidying an empty pre-generation module can remove
the dependency you are about to need.

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

If the consumer sets Content Security Policy, test the rendered application
under that exact policy. The bundled standard Alpine runtime requires dynamic
function evaluation, and Alpine/component state writes inline style
attributes. A policy that allows only `script-src 'self'` can load every local
file and still leave Alpine behavior dead. Allow the required local-runtime
behavior (`'unsafe-eval'` for scripts and inline style mutation), or deliberately
replace the default head/runtime wiring with a CSP-compatible stack and verify
every Alpine-backed component. Do not weaken unrelated CSP directives.

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

Do not invent the page around isolated components. For any build or redesign,
read `references/design-intelligence.md` first. Write its compact surface brief,
use the existing identity when present, route by the user's task archetype, and
choose a deliberate visual direction before selecting components. Do not ask
for an aesthetic preference when a reversible, context-backed choice is
available.

Then choose the closest task pattern and read
`references/application-patterns.md` before composing it:

- **App Shell** starts with `appshell.AppShell`, then supplies navigation,
  sidebar, global search, and route content.
- **Operations List** combines `pageheader.PageHeader`, `toolbar.Toolbar`,
  `table.Table`, `skeleton.Skeleton`, and `emptystate.EmptyState`.
- **Detail Workspace** combines `pageheader.PageHeader`, identity, status,
  route-local views, actions, neutral panels, and a secondary detail rail.
- **Multi-step Workflow** combines `pageheader.PageHeader`, Steps, Form,
  server validation, review, and safe submission.

Keep domain vocabulary, authorization, data priority, and workflow rules in the
application. Goshtoso supplies the component vocabulary and supported layout
contract, not the product decisions.

Use `panel.Panel` for neutral full-width application regions and `card.Card`
only for genuinely card-like content. For invalid forms, give controls stable
IDs, set `FormErrorItem.TargetID` from `FieldGroupConfig.FocusTargetID()` after
validation binding, and let `FieldGroup` connect field errors, hints, and
accessible required state to built-in controls. Composite values still need
server validation.

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
- HTMX does not swap 4xx/5xx responses by default. For expected validation or
  recovery fragments, either return a swappable 2xx response while preserving
  application status in a header, or deliberately opt in from
  `htmx:beforeSwap`. Do not let a correct server fragment fail silently.
- Move focus after a fragment is settled, not merely swapped. Use
  `htmx:afterSettle` to focus `[data-autofocus]`, `FormErrors`, or the next
  task target only when the response deliberately marks one; filter and search
  swaps must keep focus and caret in the initiating control. Never install a
  global fallback that focuses the page title after every swap.
- Before implementing a consequential POST, write its allowed state-transition
  table. Terminal decisions stay terminal unless the product explicitly models
  reversal; stale or partial evidence must have an explicit action policy; and
  conflict/retry UI must not bypass idempotency or reapply a side effect. Test
  the recovery action itself with two stale tabs and a repeated final request.
- Keep native constraint attributes (`required`, `min`, `max`, `pattern`) on
  app-owned native controls when the rule is known, while retaining identical
  server validation. Composite controls still require server validation.
- Render recovery inside the application shell for unknown IDs and routes.
  Preserve selected record, filters, theme, and color mode across error,
  conflict, retry, and success links.
- Use `link.Link(..., link.WithAppearance(link.AppearanceButton))` for GET
  navigation that should look like a button. Keep `button.Button` for form
  submission and mutations; visual appearance must not erase native semantics.
- Goshtoso's own components are pre-generated. Do not run `templ generate`
  against the module cache or vendor copy; run it for the consumer app's files.
- `.templ` files can use `templ.URL`, `templ.KV`, and other templ helpers without
  explicitly importing `github.com/a-h/templ`; the generator injects that import
  and an explicit duplicate can fail generation.
- Templ component calls can consume adjacent text as child content. After
  `@Component()`, wrap intended sibling or adjacent text in an explicit element
  such as `<span>` so it cannot disappear into the component's child slot.
- Do not import `site/`; it is the Goshtoso demo app, not public library API.

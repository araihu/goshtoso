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

`head.Dependencies()` emits Goshtoso CSS and an ordered loader for Alpine.js,
its collapse/focus/mask plugins, HTMX, and the first-party
`/assets/js/goshtoso.min.js` bundle. Third-party dependencies try
version-pinned CDN URLs first and create a fresh script for the exact embedded
version when a CDN download fails. Keep `assets.Handler()` mounted even when
the third-party CDN normally succeeds.

Choose deliberately:

- Use `head.Dependencies()` for the resilient CDN-first default.
- Use `head.Dependencies(head.WithLocalRuntime())` when the app must make no
  runtime CDN requests, such as an offline PWA, desktop/mobile WebView,
  air-gapped deployment, or explicit network policy. Do not select it merely to
  make a probe easier; exercise the default fallback contract.
- Override one pair with `WithDependencyCDNURL` and
  `WithDependencyLocalURL`, plus `WithDependencyIntegrity` when the bytes
  change. One integrity value applies to both sources, so primary and fallback
  bytes must match.
- Use `WithRuntimeManifest(assets.DefaultRuntimeManifest())` when the consumer
  must enable, omit, add, or reorder the complete typed dependency set. The
  option snapshots once; other options apply afterward regardless of argument
  position. Invalid manifests fail before writing HTML.
- Use `WithoutLocalFallback()` only when failure is preferable to local retry.
- Await `window.goshtosoDependencies.ready` before application JavaScript that
  requires every dynamically loaded dependency.

For offline inventories or build descriptors, use
`assets.DefaultRuntimeManifest()` instead of copying versioned paths from
generated files or the module cache. Its caller-owned dependency slice is in
execution order and includes roles, CDN primary, Handler-served local URL, SRI,
enabled, minimal-set membership, defer, and loader-readiness semantics. Cache
the separate stylesheet and bootstrap loader local URLs too; execute only
enabled dependencies. The combined first-party bundle is enabled by default;
dark mode, HTMX SSE/WS, Combobox, and ActionGroup are disabled inventory. Legacy
Combobox/ActionGroup overrides replace the bundle with both standalone helpers.
Do not execute the loader and direct local scripts together.

Custom manifests may contain unique safe custom script roles. Preserve Alpine
plugin/first-party/dark-mode-before-Alpine and HTMX-before-SSE/WS ordering.
`WithLocalRuntime` is only for the default manifest; custom local-only loading
sets every desired `PrimaryURL` to its `LocalURL`, keeps the loader local, and
uses `WithoutLocalFallback()`. Loader `LocalURL` is inventory, not bootstrap
fallback. Dependency `Defer` describes direct tags, not custom-loader order.
At the top level, only `Loader.Defer` is supported; stylesheet `Defer` and both
top-level `WaitForWindowLoaded` values are rejected before rendering.

Configuration freedom is not a compatibility guarantee. Goshtoso tests the
pinned combination: Alpine 3.14.9, HTMX 2.0.8, SSE 2.2.3, and WS 2.0.3.

Bind same-version caches only when `assets.GoshtosoVersion().Status` is
`assets.VersionExact`. Development, replaced, and unavailable builds leave the
exact `Version` empty. Replacement request/target metadata is diagnostic only:
the requested release version does not identify replacement bytes.

If the consumer sets Content Security Policy, test the rendered application
under that exact policy. The bundled standard Alpine runtime requires dynamic
function evaluation, and Alpine/component state writes inline style
attributes. A policy that allows only `script-src 'self'` can load every local
file and still leave Alpine behavior dead. Allow the configured CDN origin or
use `WithLocalRuntime()`, plus the required Alpine behavior (`'unsafe-eval'` for
scripts and inline style mutation). `templ.WithNonce` is propagated to the
loader and its child scripts for nonce/`strict-dynamic` policies. Otherwise,
deliberately replace the default head/runtime wiring with a CSP-compatible stack
and verify every Alpine-backed component. Do not weaken unrelated CSP
directives.

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

For short install and command snippets, use the component-owned compact
density. Keep multiline manifests and source examples at default density:

```templ
@codeblock.CodeBlock(codeblock.Config{
	Language: "bash",
	Label: "Install",
	Code: "go get github.com/araihu/goshtoso@v0.1.5",
	Density: codeblock.DensityCompact,
})

@codeblock.CodeBlock(codeblock.Config{
	Language: "yaml",
	Label: "manifest.yaml",
	Code: manifestYAML,
})
```

Do not rebuild the header/copy control or override internal padding in consumer
CSS. Compact density is for short snippets; default density preserves scanning
rhythm for multiline code.

## Sprite Icons

Use `components/icon` for accessible SVG sprite symbols. External mode is the
zero-value default: give it a relative, same-origin `SpriteURL` whenever
possible. The bundled Heroicons package supplies that URL and typed constants:

```templ
import (
    "github.com/araihu/goshtoso/components/icon"
    "github.com/araihu/goshtoso/components/icon/heroicons"
)

templ StatusMark() {
    @icon.Icon(icon.Config{
        SpriteURL: heroicons.SpriteURL,
        Symbol:    heroicons.Icon16SolidCheck,
        Label:     "Saved",
    })
}
```

Use a nonblank `Label` for a labelled image. Empty label or `Decorative: true`
means decorative; do not present a label and decorative intent as separate
meanings. `RootClass` can use `text-*` utilities because compatible sprites
inherit `currentColor`; do not add root fill or stroke overrides.

Use `Mode: icon.ModeInline` with no `SpriteURL` only when the `<symbol>` already
exists in the current document. Cross-origin external `<use>` references vary
by browser and CORS policy; an HTTPS page may block an HTTP sprite. Keep
same-origin relative sprite paths as the normal deployment choice.

Generate project-local typed bindings for a schema-v1 catalog instead of
putting project-specific names in Goshtoso:

```bash
go run github.com/araihu/goshtoso/cmd/iconcatalog@latest \
  -catalog ./assets/icons/catalog.json \
  -namespace ui -product application -sprite-url /assets/icons/app.svg \
  -package appicons -const-prefix Icon -out ./internal/appicons/names_gen.go
```

Run it with `-check` in CI. The generator rejects unsupported schemas,
duplicate or malformed names/symbols, identifier collisions, non-sprite SVG
inputs, and incompatible color behavior.

## From First Component to Application

Do not invent the page around isolated components. For any build or redesign,
read `references/design-intelligence.md` first. Write its compact surface brief,
use the existing identity when present, route by the user's task archetype, and
choose a deliberate visual direction before selecting components. Do not ask
for an aesthetic preference when a reversible, context-backed choice is
available.

For a public organization, product, portfolio, or publication with static
content, begin with `examples/brand-site`, not App Shell. It is a complete,
copyable Go/templ static generator and explicitly separates Goshtoso tokens
from product-owned visual direction. Create it in an empty target with:

```bash
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -init-brand-site=./my-site
```

Then replace its placeholder content and `brand.css`; do not turn the starter
into generic hero/features/testimonials/CTA sections. Add Goshtoso controls
only when the site has real navigation, feedback, or form behavior needing
them.

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

### Reusable application shells

For a public product landing page that should follow the Goshtoso site frame,
use `github.com/araihu/goshtoso-app-shells/landingshell` instead of copying the
demo landing page's utility classes. The shell owns the document head,
first-paint and interactive color mode, responsive brand/navigation header,
version badge, icon-only mode and repository controls, hero boundary, content
container, and structured linked footer. The consumer owns the product's hero
copy, content sections, calls to action, code examples, and art direction.

Install the first public release containing this package explicitly:

```bash
go get github.com/araihu/goshtoso-app-shells/landingshell@v0.1.2
```

Configure `landingshell.Config` with `landingshell.Brand`,
`[]landingshell.Link`, `landingshell.AppearanceConfig`, and
`landingshell.Footer`, then supply `landingshell.Page.Hero` and
`landingshell.Page.Content` components. Keep footer identity structured:
product logo/name, concise metadata, linked organization, and typed links.

Static generators must import the assets package and extract the exact
content-versioned files from its handler:

```go
import landingassets "github.com/araihu/goshtoso-app-shells/landingshell/assets"

handler := landingassets.Handler()
stylesheetURL := landingassets.StylesheetURL("")
scriptURL := landingassets.ScriptURL("")
```

Do not copy those files or recreate the dark-mode runtime in the consumer.

For a documentation site that should follow the Goshtoso demo frame, use the
public `github.com/araihu/goshtoso-app-shells/componentdocshell` module instead
of copying the demo site's layout. It owns the full-width brand header,
search-first fixed desktop sidebar and mobile drawer, theme and dark-mode
controls, optional table of contents, and HTMX main-content navigation. The
consumer still supplies brand assets, navigation/search data, page content,
runtime slots, and application-specific metadata.

Configure the shell through `componentdocshell.Config` and its nested config
structs. Set the default theme, available theme list, theme selector ID,
dark-mode binding, TOC hooks, and whether appearance controls render. The shell
includes all Goshtoso themes by default and can run with one fixed theme and no
selector when a product requires that behavior.

Use `github.com/araihu/goshtoso-app-shells/componentpage` for the repeated
structure inside component reference pages, including semantic sections and
preview/code examples. It does not own the site frame; the catalog shell is a
separate pattern. Do not treat a catalog grid and a component documentation
page as the same shell.

Before implementing a consequential action, stale-data recovery, or
interruptible workflow, read `references/adversarial-acceptance.md`. Copy its
invariant ledger into the consumer repository and derive the HTTP and browser
tests from every state/action row. A rendered state gallery or builder-authored
happy-path harness is not sufficient evidence.

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

## Public Surface Guardrails

- For a one-line install command, use the standard `codeblock.CodeBlock` inside
  an app-owned wrapper that sets only width and spacing, as on the Goshtoso
  landing page. Do not turn the component header and code body into a custom
  grid or style its private DOM structure. For a dense list of short commands,
  use a documented compact API when the pinned release provides one; otherwise
  prefer prose or inline code instead of several bespoke mini code blocks.
- Match app-shell color-mode controls with an icon button whose accessible label
  changes between `Switch to dark mode` and `Switch to light mode`. Use
  `toggle.Toggle` only when a labelled switch is the intended interface. Keep
  the control inside a live Alpine `x-data` scope and bind it to the same state
  that applies the document's `dark` class.
- Treat unavailable browser storage as a normal runtime mode. A consent or
  capability probe alone is not a fallback: guard every `localStorage`
  read/write/remove operation, while the in-memory store continues to update the
  document class for the session. Browser-test the actual visible control under
  light and dark system preferences and with `localStorage` throwing. Assert the
  document class, accessible control state, store, and persisted value remain
  synchronized with no page errors.

## Integration Checks

- Missing styling usually means `/assets/styles.css` is not served or
  `head.Dependencies()` is absent.
- Dead combobox keyboard navigation usually means `/assets/js/goshtoso.min.js` is
  missing. An unformatted `x-mask` input means the Mask plugin or an Alpine
  `x-data` root is missing. Prefer `head.Dependencies()` instead of hand-written
  script tags.
- Alpine plugins must load before Alpine core. Avoid manual tags unless the app
  has a strong reason.
- HTMX handlers should return rendered HTML fragments, not JSON.
- `button.WithLoadingText` follows either a Button-owned request or an ancestor
  HTMX form. For a form-owned mutation, add
  `hx-disabled-elt="find button[type='submit']"` to the form so pending copy and
  duplicate prevention are both real; hold the request in a browser test and
  assert the label, disabled state, and final response.
- HTMX does not swap 4xx/5xx responses by default. For expected validation or
  recovery fragments, either return a swappable 2xx response while preserving
  application status in a header, or deliberately opt in from
  `htmx:beforeSwap`. Do not let a correct server fragment fail silently.
- Move focus after a fragment is settled, not merely swapped. Use
  `htmx:afterSettle` to focus `[data-autofocus]`, `FormErrors`, or the next
  task target only when the response deliberately marks one; filter and search
  swaps must keep focus and caret in the initiating control. Never install a
  global fallback that focuses the page title after every swap.
- A queue/detail swap must leave one selected identity everywhere. After settle,
  assert that the URL, rendered detail key, focused target, visual selected row,
  and collection semantics (`aria-current` or `aria-selected`) all name the same
  record. Prefer rerendering the collection from server truth; if the collection
  stays outside the swap, update every one of those representations explicitly.
- Render exactly one primary mobile navigation trigger. Navbar right actions can
  create a Navbar mobile menu; do not place that hamburger beside a Sidebar
  overlay trigger unless the two controls expose clearly different destinations.
  Exercise Escape and assert the trigger name and `aria-expanded` after closing.
  Keep the overlay viewport-owned (`fixed top-16 bottom-0`, adjusted to the real
  header height), not `absolute top-full` inside a header child; after opening at
  390 px, assert the panel intersects the viewport and has positive height.
- Before implementing a consequential POST, write its allowed state-transition
  table. Terminal decisions stay terminal unless the product explicitly models
  reversal; stale or partial evidence must have an explicit action policy; and
  conflict/retry UI must not bypass idempotency or reapply a side effect. Test
  the recovery action itself with two stale tabs and a repeated final request.
- Treat `references/adversarial-acceptance.md` as the executable contract for
  those transitions. Test forged or hidden actions against server truth, real
  transport failure, retained drafts, loading deduplication, final URL or
  receipt identity, focus, and the exact side-effect count. No consequential
  ledger row may remain untested.
- Keep native constraint attributes (`required`, `min`, `max`, `pattern`) on
  app-owned native controls when the rule is known, while retaining identical
  server validation. Composite controls still require server validation.
- Render recovery inside the application shell for unknown IDs and routes.
  Preserve selected record, filters, theme, and color mode across error,
  conflict, retry, and success links.
- After PRG, Back/Forward must not present stale authoritative status from the
  browser cache. Compare the restored revision/status with a fresh server read;
  use `Cache-Control: no-store` for sensitive task pages or a `pageshow` refresh
  when a persisted document cannot be trusted.
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

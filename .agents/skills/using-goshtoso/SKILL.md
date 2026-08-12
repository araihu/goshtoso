---
name: using-goshtoso
description: Use when building, designing, redesigning, integrating, or updating an external Go/templ application in the Goshtoso ecosystem. Triggers for discovery and reuse of Goshtoso components, Goshtoso Charts, Margo Markdown/document rendering, Goshtoso App Shells, consumer icon packs, application shells, dashboards, operations lists, detail pages, settings, onboarding, workflows, public content, metadata, social previews, component selection, visual direction, state design, asset serving, Tailwind CSS, Alpine.js, HTMX, or component API debugging. Requires checking supported ecosystem solutions before writing custom HTML, CSS, or JavaScript.
---

# Using Goshtoso

## Overview

Goshtoso is a Go UI component library for server-rendered templ apps. Treat it
as a library dependency: consumers import pre-generated components, serve the
embedded assets, and run `templ generate` only for their own `.templ` files.
This skill is for integrating Goshtoso into external applications, not for
modifying the Goshtoso component library or demo site itself.

## When to Use

- Build or redesign a Go/templ product with Goshtoso.
- Select components, charts, document rendering, page frames, or icons.
- Integrate Goshtoso assets, metadata, CSS, HTMX, or Alpine.js.
- Diagnose a consumer-side Goshtoso rendering or interaction problem.

## When NOT to Use

- Modify Goshtoso or a related module itself; follow that repository's rules.
- Build a non-Go or non-templ UI that cannot consume server-rendered components.
- Perform release, publication, deployment, or repository-maintainer work.

## Required Discovery Pass

Before writing custom HTML, CSS, JavaScript, or a new templ primitive, inspect
the consumer's selected module versions and public packages:

```bash
go list -m all | rg '^github.com/araihu/(goshtoso|goshtoso-charts|goshtoso-app-shells|margo)( |$)'
for module in goshtoso goshtoso-charts goshtoso-app-shells margo; do
  go list -m -versions "github.com/araihu/$module"
done
```

For modules already selected in `go.mod`, inspect exact public packages with
`go doc`, including Goshtoso `components`, Charts `components/line`, App Shells
`consoleshell`, and Margo's root package. For an absent module, inspect the
chosen release tag's docs and public source without adding it merely to browse.
Absence from `go.mod` is not proof of a gap.

Route the need through these surfaces in order:

1. Goshtoso components and documented application patterns for UI and layout.
2. Goshtoso Charts for static/vector or interactive data visualization.
3. Margo for Markdown, standalone HTML, linked sites, PDF, or slide decks.
4. Goshtoso App Shells for landing, console, component-doc, or component-page
   frames.
5. Bundled Heroicons or `cmd/iconpack` for icons.

Record the closest supported surface and choose `reuse`, `compose`, or `gap`.
Custom HTML, CSS, or JavaScript requires a concrete gap. Prefer app-owned
semantic content inside supported component or shell slots; never copy
demo-site or `internal/` implementation markup.

When installed references are available, read `references/ecosystem-discovery.md`
for module-specific searches and public surfaces. The discovery decision above
remains complete when a streaming `skills use` client provides only this file.

## Default Integration

Use this path unless the app deliberately owns a custom Tailwind build.
Goshtoso requires **Go 1.26.5 or newer**.

```bash
GOSHTOSO_VERSION="${GOSHTOSO_VERSION:?set a release such as v0.1.13}"
go get github.com/araihu/goshtoso@"$GOSHTOSO_VERSION"
go get github.com/a-h/templ
TEMPL_VERSION="$(go list -m -f '{{.Version}}' github.com/a-h/templ)"
go install github.com/a-h/templ/cmd/templ@"$TEMPL_VERSION"
templ generate
go mod tidy
go test ./...
```

Create the consumer's `.templ` file before `templ generate`. Run
`templ generate` before `go mod tidy`: generated Go is what makes the templ
runtime import visible, while tidying an empty pre-generation module can remove
the dependency required by the generated source.

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
			@head.Metadata(head.MetadataConfig{
				Title:        "Page title",
				Description:  "Route-specific description",
				CanonicalURL: "https://example.com/route",
				SiteName:     "Product name",
				Image: head.SocialImage{
					URL:      "https://example.com/og.jpg",
					MIMEType: "image/jpeg",
					Width:    1280,
					Height:   640,
					Alt:      "Useful description of the preview",
				},
			})
			@head.Dependencies()
		</head>
		<body>{ children... }</body>
	</html>
}
```

`head.Metadata()` emits the document title and description, canonical URL,
required Open Graph properties, structured image metadata and alt text, plus
explicit X/Twitter Card tags. Title, description, canonical URL, image URL,
parameter-free RFC 6838 image MIME type, positive dimensions, and image alt
text are required. Both URLs must be absolute HTTPS. `Render` validates the
whole configuration before writing and returns an error instead of emitting
partial or mixed metadata.
Open Graph type defaults to `website`; X/Twitter Card defaults to
`summary_large_image`. Give every indexable route distinct values and do not
reuse homepage metadata for subpages. A 1280x640 JPEG or PNG under 1 MB is the
safe default social image. Render metadata in initial HTML, not through client
JavaScript or HTMX fragments.

`head.Dependencies()` emits Goshtoso CSS and an ordered loader for Alpine.js,
its collapse/focus/mask plugins, HTMX, and the first-party
`/assets/js/goshtoso.min.js` bundle. Third-party dependencies try
version-pinned CDN URLs first and create a fresh script for the exact embedded
version when a CDN download fails. Keep `assets.Handler()` mounted even when
the third-party CDN normally succeeds.

Use `head.Dependencies()` for the resilient CDN-first default and
`head.Dependencies(head.WithLocalRuntime())` for explicit offline or no-CDN
policy. Await `window.goshtosoDependencies.ready` before app JavaScript needing
the complete runtime. Read `references/runtime-integration.md` before changing
the runtime manifest, integrity, fallback, cache identity, or CSP behavior.

## Rendering Components

Import from `github.com/araihu/goshtoso/components/<name>` and call the exported
templ component from application-owned `.templ` files:

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

Search `references/components-reference.md` with `rg` for the package or symbol;
do not load its full generated catalog. It contains exact import paths, entry
points, config fields, enum values, and differing package names.

Read the selected Goshtoso tag's `docs/COMPONENT_MODEL.md` before choosing
between constructors or configuration fields. It documents the common
component interface, concrete return values, constructor styles, stable Kind
identity, and rendered defaults.

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

For icons outside bundled Heroicons, use the consumer-local `iconpack` flow.
Declare sources in `.iconpack.yaml`, then explicitly establish first trust:

```bash
GOSHTOSO_VERSION="${GOSHTOSO_VERSION:?set v0.1.12 or newer}"
go run github.com/araihu/goshtoso/cmd/iconpack@"$GOSHTOSO_VERSION" \
  -config ./.iconpack.yaml -trust \
  -out ./internal/appicons -package appicons \
  -const-prefix Icon -sprite-url /assets/icons/appicons/sprite.svg
```

Review and commit `.iconpack.lock.yaml`, generated manifest, provenance, and
licenses. Rerun without `-trust`; changed or unavailable source bytes must fail
instead of silently re-trusting. Serve the generated sprite at the exact
`-sprite-url`, then render through the generated typed `Icon` helper. Read the
selected tag's `docs/ICONPACK.md` and run the selected command with `-help` for
arbitrary GitHub trees/files, multiple packs, Arai Hû Assets release archives,
CI checks, and migration from legacy source manifests. The `.iconpack.yaml`
flow requires Goshtoso `v0.1.12` or newer; retain an older app dependency only
when the generated package remains compatible with that app.

## From First Component to Application

Do not invent the page around isolated components. Complete the Required
Discovery Pass, then read `references/design-intelligence.md`. Write its compact
surface brief, use the existing identity when present, route by the user's task
archetype, and choose a deliberate visual direction before selecting
components. Do not ask for an aesthetic preference when a reversible,
context-backed choice is available.

For a public product or organization landing page, try App Shells
`landingshell` first. It already owns the responsive frame, metadata,
first-paint color mode, navigation, hero boundary, and structured footer while
leaving content and art direction to the app.

Choose the selected Goshtoso tag's `examples/brand-site` starter only when the
surface needs a fully app-owned static generator or content structure that
`landingshell` slots cannot express. Record that concrete gap first. Create the
starter in an empty target with:

```bash
GOSHTOSO_VERSION="${GOSHTOSO_VERSION:?set a release}"
go run github.com/araihu/goshtoso/cmd/goshtoso@"$GOSHTOSO_VERSION" -init-brand-site=./my-site
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

### Related ecosystem modules

Read `references/ecosystem-discovery.md` before implementing any page frame,
chart, document renderer, or icon acquisition. It records current public
surfaces for `goshtoso-app-shells`, `goshtoso-charts`, Margo, and Iconpack.
Use the consumer's selected release and its public packages; do not infer API
availability from a repository's `main` branch.

App Shells supplies `landingshell`, `consoleshell`, `componentdocshell`, and
`componentpage`. Charts supplies typed static/vector and opt-in interactive
components. Margo owns Markdown-to-HTML/site/PDF/deck workflows. These are
supported solutions, not examples to copy into app-owned HTML or JavaScript.

Use `github.com/araihu/goshtoso-app-shells/componentdocshell` for the full
documentation or API-reference frame. Use
`github.com/araihu/goshtoso-app-shells/componentpage` only for repeated
component-reference page structure inside that frame; the catalog shell is a
separate pattern. Keep `consoleshell` for operations products and
`landingshell` for public product or organization pages.

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
GOSHTOSO_VERSION="${GOSHTOSO_VERSION:?set a release}"
go run github.com/araihu/goshtoso/cmd/goshtoso@"$GOSHTOSO_VERSION" -theme -out=css/goshtoso-theme.css
go run github.com/araihu/goshtoso/cmd/goshtoso@"$GOSHTOSO_VERSION" -source-path
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

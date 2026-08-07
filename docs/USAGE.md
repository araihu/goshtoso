# Goshtoso Consumer Integration Guide

Use this guide to add Goshtoso to a Go web application and render a first
component. Complete installation, asset, and head setup in order; then choose a
CSS strategy and component API for the application you are building.

For component constructors and options, read the
[Goshtoso Component Model](COMPONENT_MODEL.md). It documents the common
`components.Component` interface, concrete return values, stable `Kind`
identity, constructor styles, and rendered defaults.

## Installation

Goshtoso requires **Go 1.26.5 or newer**. Use the same or a newer Go toolchain
for the consumer module so module resolution does not fail late in setup.

### 1. Add the dependency

```bash
go get github.com/araihu/goshtoso@latest
```

Goshtoso's own components ship **pre-generated** — you never run `templ generate`
against the library. But your *own* `.templ` pages do need the templ toolchain:

```bash
go get github.com/a-h/templ                       # runtime (your generated code imports it)
go install github.com/a-h/templ/cmd/templ@latest  # the CLI, if not already installed
templ generate                                    # YOUR .templ → _templ.go
go mod tidy                                       # run after generation sees templ imports
```

Create at least one consumer `.templ` file before generation. Generate before
tidying: without generated Go imports, `go mod tidy` can correctly remove the
templ runtime that the next step needs.

### 2. Serve the bundled assets (required)

Serve Goshtoso's embedded assets and let `head.Dependencies()` link them. The
handler supplies the compiled theme CSS, the dependency loader, first-party
helpers, and version-matched local fallbacks. No Tailwind build or extraction
is required.

```go
// main.go
import "github.com/araihu/goshtoso/assets"

mux := http.NewServeMux()
mux.Handle("GET /assets/", assets.Handler()) // styles.css + js/ + fonts/ + images/
                                             // self-strips /assets/; do not use StripPrefix
```

```go
// page.templ — pinned CDN first, matching /assets/* fallback
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
            @head.Dependencies()        // CSS + Alpine + plugins + HTMX + first-party helpers
            // or @head.DependenciesMinimal() — no Alpine plugins
        </head>
        ...
    </html>
}
```

`head.Metadata()` emits route-specific document, canonical, Open Graph, and
X/Twitter Card metadata in initial server-rendered HTML. Title, description,
canonical URL, image URL, parameter-free RFC 6838 image MIME type, positive
dimensions, and image alt text are required; both URLs must be absolute HTTPS.
`Render` validates the whole configuration before writing and returns an error
rather than partial metadata. Open Graph type defaults to `website`; X/Twitter
Card defaults to `summary_large_image`. Give each indexable route distinct values;
1280x640 JPEG or PNG under 1 MB is the safe default. Do not inject share
metadata through JavaScript or an HTMX fragment because link-preview crawlers
may not execute either.

The served `styles.css` already carries every component style + the theme system
(16 themes). **Stock CDN Tailwind will not work** — the theme tokens
(`bg-primary`, `text-on-surface`, …) live only in this compiled CSS.

The default loader tries exact-version unpkg URLs in dependency order. If a
download fails, it creates a new script element for the matching embedded file;
it does not mutate and retry a script the browser has already abandoned. Await
`window.goshtosoDependencies.ready` when application code must run only after
every dependency is available. The loader also emits
`goshtoso:dependency-fallback`, `goshtoso:dependencies-ready`, and
`goshtoso:dependency-error` events on `window`.

### 3. Render a component

Import a component into one of your own `.templ` files. Library components are
already generated; run `templ generate` after changing your file so the
application gets its generated Go code.

```templ
import "github.com/araihu/goshtoso/components/button"

templ SaveAction() {
	@button.Button(
		button.WithTone(button.TonePrimary),
		button.WithType("button"),
	) {
		Save changes
	}
}
```

Run `templ generate`, then build or test the application. Add HTMX attributes
when the action needs a server-rendered fragment swap; use Alpine for local
state or immediate client feedback.

#### Dependency loading options

Use local-only mode when an offline PWA, desktop/mobile WebView, air-gapped
deployment, or explicit network policy must forbid runtime CDN requests:

```templ
@head.Dependencies(head.WithLocalRuntime())
```

Override one CDN and its matching fallback without taking ownership of the
rest of the stack:

```templ
@head.Dependencies(
    head.WithDependencyCDNURL(head.DependencyHTMX, "https://cdn.example.com/custom-htmx.min.js"),
    head.WithDependencyLocalURL(head.DependencyHTMX, "/static/runtime/custom-htmx.min.js"),
    head.WithDependencyIntegrity(head.DependencyHTMX, "sha384-..."),
)
```

Other controls are `WithoutLocalFallback()`, `WithoutDependency(...)`,
`WithStylesheetURL(...)`, `WithComboboxURL(...)`,
`WithActionGroupURL(...)`, and `WithLoaderURL(...)`.
`WithoutLocalFallback()` keeps the configured CDN primaries but turns a failed
download into `goshtoso:dependency-error`. `WithoutDependency(...)` is for an
application that deliberately loads that runtime itself. Empty override URLs
leave the strong default unchanged.

Default third-party scripts carry canonical SHA-384 Subresource Integrity from
`muamba.yaml`. Offline verification checks every embedded byte; the explicit
`just vendor-js` acquisition step fetches and materializes locked inputs. When a
custom CDN or local URL changes the bytes, pass its
matching hash with `WithDependencyIntegrity`; pass an empty hash only when the
application deliberately disables SRI for that dependency. A dependency has one
integrity value for both its primary and fallback URL, so those two sources must
serve byte-identical content when SRI is enabled.

The exact tested pins, paths, loading order, and defaults are generated in
[`RUNTIME_DEPENDENCIES.md`](RUNTIME_DEPENDENCIES.md) from `muamba.yaml` and
`assets/runtime.overlay.yaml`. Changing a URL or manifest configures how the browser loads
the application-owned stack; it does not guarantee that arbitrary dependency
versions are compatible with Goshtoso or each other.

#### Runtime manifest and exact library identity

Consumers that cache Goshtoso for offline use, publish an asset inventory, or
need to validate their rendered shell can use the public assets contract instead
of copying paths from generated files or the Go module cache:

```go
import (
    "fmt"

    "github.com/araihu/goshtoso/assets"
)

identity := assets.GoshtosoVersion()
if identity.Status != assets.VersionExact {
    // Fail closed: development, replaced, and unavailable builds do not expose
    // identity.Version as if it were the bytes of a released dependency.
    return fmt.Errorf("goshtoso identity is %s", identity.Status)
}

runtime := assets.DefaultRuntimeManifest()
// runtime.Stylesheet.LocalURL and runtime.Loader.LocalURL are served by Handler.
// runtime.Dependencies is caller-owned and ordered for execution.
for _, dependency := range runtime.Dependencies {
	if !dependency.Enabled {
		continue
	}
	hash, ok := assets.RuntimeHash(dependency.Role)
	if !ok {
		continue
	}
	cache(dependency.Role, dependency.LocalURL, dependency.Integrity, hash)
}
```

`RuntimeManifest.Stylesheet` is the compiled CSS. `Loader` is the external
bootstrap rendered by the CDN-first default. `Dependencies` contains Alpine
Collapse, Alpine Focus, Alpine Mask, the first-party bundle, dark-mode store,
Alpine core, HTMX, HTMX SSE, HTMX WebSocket, Combobox, and ActionGroup in exact
declared order. Dark mode, SSE, WS, Combobox, and ActionGroup are disabled
inventory by default. Every entry exposes primary and local URLs, SRI, enabled,
minimal-set membership, direct-tag defer, and loader-readiness semantics.
Direct local rendering uses the dependency slice and does not execute the
loader again.

Each `DefaultRuntimeManifest` call owns its value and dependency slice; caller
mutation cannot change later results or `head.Dependencies()` rendering. Every
declared local URL uses the `/assets/` mount and is served by `assets.Handler()`.
`RuntimeHash` returns normalized `sha384:<hex>` content hashes for vendored
roles; `MuambaHash` and `MuambaResources` expose the lower-level acquisition
records, retained licenses, and provenance files.

For full dependency ownership, pass a customized snapshot to
`head.WithRuntimeManifest`. The loader emits enabled scripts in exact slice
order, accepts unique custom roles, and filters `DependenciesMinimal` with
`IncludeInMinimal`:

```go
func applicationRuntime() assets.RuntimeManifest {
    runtime := assets.DefaultRuntimeManifest()
    for i := range runtime.Dependencies {
        dependency := &runtime.Dependencies[i]
        switch dependency.Role {
        case assets.RuntimeRoleDarkMode,
            assets.RuntimeRoleHTMXExtSSE,
            assets.RuntimeRoleHTMXExtWS:
            dependency.Enabled = true
        }
    }
    runtime.Dependencies = append(runtime.Dependencies, assets.RuntimeAsset{
        Role:             "application-runtime",
        Kind:             assets.RuntimeAssetScript,
        PrimaryURL:       "/static/application-runtime.js",
        LocalURL:         "/static/application-runtime.js",
        Enabled:          true,
        IncludeInMinimal: true,
    })
    return runtime
}
```

```templ
@head.Dependencies(head.WithRuntimeManifest(applicationRuntime()))
```

`WithRuntimeManifest` snapshots its argument immediately and can appear once.
Other options apply afterward regardless of the manifest option's argument
position. Under a custom manifest, an option targeting an absent built-in role
is a render error instead of a silent no-op. Validation happens before HTML is
written: roles must be unique and safe; dependencies must be scripts; URLs must
be HTTP(S) or relative; enabled dependencies need the selected loader; Alpine
plugins, the first-party bundle, and dark mode must precede Alpine; HTMX must
precede SSE/WS; and the combined first-party bundle cannot run with either
standalone compatibility helper.

`Stylesheet.Enabled`, `Loader.Enabled`, and both top-level
`IncludeInMinimal` values control whether those tags render. Their `Integrity`
values render SRI attributes. `Loader.Defer` controls the loader tag.
`Stylesheet.Defer`, `Stylesheet.WaitForWindowLoaded`, and
`Loader.WaitForWindowLoaded` are unsupported and rejected before HTML is
written. A loader's `LocalURL` is asset inventory, not an automatic fallback
for the loader itself; only dependency entries use `LocalURL` as loader
fallback. Likewise, a custom stylesheet's `LocalURL` is not an automatic CSS
fallback. On dependency entries, `WaitForWindowLoaded` is honored by the loader
and `Defer` describes direct local script tags; neither field guarantees
compatibility or custom-loader execution order.

`WithLocalRuntime` deliberately rejects a custom manifest because a mixed set
of direct deferred and non-deferred tags cannot guarantee declared execution
order. For a custom local-only stack, use the ordered loader itself from a local
URL, copy each desired `LocalURL` into its `PrimaryURL`, and disable fallback:

```go
func localApplicationRuntime() assets.RuntimeManifest {
    runtime := applicationRuntime()
    runtime.Stylesheet.PrimaryURL = runtime.Stylesheet.LocalURL
    runtime.Loader.PrimaryURL = runtime.Loader.LocalURL
    for i := range runtime.Dependencies {
        runtime.Dependencies[i].PrimaryURL = runtime.Dependencies[i].LocalURL
    }
    return runtime
}
```

```templ
@head.Dependencies(
    head.WithRuntimeManifest(localApplicationRuntime()),
    head.WithoutLocalFallback(),
)
```

Legacy `WithComboboxURL` and `WithActionGroupURL` remain compatibility options.
Either replaces the combined first-party bundle with both standalone helpers;
the bundle and standalones cannot be enabled together.

`GoshtosoVersion` reads Go build information. `VersionExact` covers an
unreplaced versioned module, including immutable pseudo-versions. A local or
module replacement reports `VersionReplaced`; its requested and replacement
records remain available as metadata, but `VersionInfo.Version` stays empty so
the requested release cannot be mistaken for replacement bytes. Unversioned
main-module builds report `VersionDevelopment`; missing build metadata reports
`VersionUnavailable`.

Skip the rest of this section unless you maintain your own Tailwind build.

### Optional: extract Goshtoso CSS for a custom Tailwind build

Goshtoso ships a CLI that extracts the pre-built CSS from embedded assets. Register it as a Go tool for version-pinned reproducibility:

```bash
# Add to go.mod (alongside your other tools)
# tool github.com/araihu/goshtoso/cmd/goshtoso
go mod tidy

# Extract CSS
go tool goshtoso -out=css/goshtoso-base.css
```

Or use `go run` for one-off extraction:

```bash
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -out=css/goshtoso-base.css
```

Then import it in your Tailwind entry point:

```css
/* your-project/css/main.css */
@import "tailwindcss";
@import "./goshtoso-base.css";
```

The extracted CSS includes all Goshtoso component styles, the theme system (16 themes), and base utilities. Add it to `.gitignore` since it's a build artifact.

### 4. Required JavaScript

If you use `@head.Dependencies()` (section 2), the JS is already wired — skip
this. Only hand-roll the tags if you are not using the `head` package. In that
case preserve the manifest order, including plugins **before** Alpine core.

Use `assets.DefaultRuntimeManifest()` as the programmatic inventory instead of
copying versioned paths into application source. The generated
[`RUNTIME_DEPENDENCIES.md`](RUNTIME_DEPENDENCIES.md) table is the exact human
readable view. Paths change when a dependency is upgraded; prefer
`@head.Dependencies()` so the application never hardcodes them. Do not omit the
first-party bundle: it contains only reusable component
behavior, including combobox/ActionGroup helpers and Alpine factories. Demo-site
providers are not part of this consumer contract. The legacy
`/assets/js/combobox.js` and `/assets/js/action-group.js` URLs remain available
for consumers using `WithComboboxURL` or `WithActionGroupURL`, but new
integrations should load the bundle once.

### Content Security Policy

Local assets do not automatically make a strict CSP compatible. Goshtoso's
bundled standard Alpine runtime uses dynamic function evaluation, and
Alpine-backed components update inline style attributes. With the default
runtime, allow the configured CDN origin (unpkg by default) or select
`WithLocalRuntime()`, permit `'unsafe-eval'` in `script-src`, and permit the
required inline style mutation while retaining restrictive `default-src`,
`connect-src`, `base-uri`, `form-action`, and `frame-ancestors` directives. A
nonce set with `templ.WithNonce` is propagated to the loader and every script it
creates, which supports nonce-based policies and `strict-dynamic`.

If the application cannot permit those runtime capabilities, do not use
`head.Dependencies()` unchanged. Supply and test a CSP-compatible Alpine stack
with the required plugins and component scripts. Verify behavior in the browser:
a self-hosted file may return 200 while CSP still prevents Alpine from starting.

## Sprite icons

`icon.Icon` renders one accessible SVG sprite symbol. Prefer a relative,
same-origin `SpriteURL`: it works with the same asset handler and deployment
origin as the rest of the application. The bundled Heroicons defaults require
no custom sprite hosting when `/assets/` is mounted:

```templ
import (
    "github.com/araihu/goshtoso/components/icon"
    "github.com/araihu/goshtoso/components/icon/heroicons"
)

templ StatusIcon() {
    @icon.Icon(icon.Config{
        SpriteURL: heroicons.SpriteURL,
        Symbol:    heroicons.Icon16SolidCheckCircle,
        Label:     "Approved",
        Size:      icon.SizeLG,
        RootClass: "text-success",
    })
}
```

Use a nonblank `Label` for an image announced to assistive technology. An empty
label is decorative; `Decorative: true` also makes the icon decorative and wins
even if a label was provided. Do not set both in a way that suggests two
meanings. `RootClass` can set `text-*` color utilities because compatible sprite
paths inherit `currentColor`; Goshtoso does not force root `fill` or `stroke`.

For a symbol already embedded in the same document, omit `SpriteURL` and use
inline mode:

```templ
@icon.Icon(icon.Config{
    Mode:   icon.ModeInline,
    Symbol: "app-mark",
    Label:  "Application mark",
})
```

Cross-origin external `<use>` references remain browser- and CORS-dependent.
Do not rely on an HTTP sprite from an HTTPS page: mixed-content protections can
block it. Keep same-origin relative URLs as the deployment default.

### Generate project-local typed bindings

`iconcatalog` generates an enumerable `Glyphs` list and typed `icon.Symbol`
constants from a schema-v1 asset catalog. Keep the catalog and generated Go file
in the consuming project; the generic Goshtoso package contains no project or
brand-specific names.

```bash
go run github.com/araihu/goshtoso/cmd/iconcatalog@latest \
  -catalog ./assets/icons/catalog.json \
  -namespace ui -product application -sprite-url /assets/icons/app.svg \
  -package appicons -const-prefix Icon \
  -out ./internal/appicons/names_gen.go
```

Run the same command with `-check` in CI. The generator selects matching records
that declare a sprite symbol and ignores other matching release artifacts. It
validates every selected symbol, rejecting unsupported schema versions,
malformed or duplicate canonical names and sprite symbols, identifier
collisions, non-SVG assets, and invalid color behavior rather than emitting
unsafe or ambiguous bindings. `monochrome` and `tintable` symbols can inherit
`currentColor`; `protected` brand symbols keep their intrinsic fills and should
not be presented as recolorable by an icon component.

### Default Heroicons provenance

Goshtoso's bundled default sprite contains third-party
[Heroicons v2.2.0](https://github.com/tailwindlabs/heroicons/tree/0435d4ca364a608cc75e2f8683d374e55abbae26)
from upstream commit `0435d4ca364a608cc75e2f8683d374e55abbae26`, under the
MIT license. The immutable Arai Hu Assets v0.1.1 catalog at
`https://araihu.com/assets/releases/v0.1.1/catalog.json` records its names and
symbols for generator compatibility; Arai Hu Assets does
not author Heroicons. Goshtoso retains the bundled MIT notice at
`/assets/icons/HEROICONS_LICENSE.txt`.

## Using your own Tailwind build

`goshtoso -version` prints the Tailwind version Goshtoso's CSS was built with
(also in [`VERSIONS.md`](../VERSIONS.md)). Match your own Tailwind to it.

### Path A — two stylesheets (recommended, low coupling)

Serve Goshtoso's prebuilt CSS and run your own Tailwind into a *separate* file.
No recompiling Goshtoso.

```html
<link rel="stylesheet" href="/assets/styles.css"/>  <!-- Goshtoso, via assets.Handler() -->
<link rel="stylesheet" href="/css/app.css"/>          <!-- your own Tailwind output -->
```

```css
/* your app.css — your own tokens/classes only */
@import "tailwindcss";
@theme { --color-brand: oklch(0.7 0.15 250); }
```

The embedded stylesheet is a contract for Goshtoso components and the official
application recipes, not a general Tailwind compiler for consumer markup.
`RootClass` and attribute hooks accept arbitrary class names, but a valid
Tailwind utility can fail silently if its selector was not emitted.

The recipe contract explicitly guarantees these audited layout utilities:
`max-w-7xl`, `xl:grid-cols-4`, `lg:col-span-2`, `min-h-64`, `sm:text-4xl`,
`first:pt-0`, `last:pb-0`, `min-w-[220px]`, and `sm:col-span-2`. Give the
application its own Tailwind build for any broader utility vocabulary and load
its stylesheet after `/assets/styles.css`.

### Path B — unified build (one tree-shaken stylesheet)

Compile Goshtoso's theme source together with your own. Requires your Tailwind
to match `goshtoso -version`.

```bash
# 1. extract the theme SOURCE next to your CSS
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -theme -out=css/goshtoso-theme.css

# 2. discover the components dir Tailwind must scan
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -source-path
# -> /…/go/pkg/mod/github.com/araihu/goshtoso@vX.Y.Z/components
```

```css
/* your main.css */
@import "tailwindcss";
@import "./goshtoso-theme.css";                 /* tokens + selectors + themes */
@source "/…/goshtoso@vX.Y.Z/components";        /* emit Goshtoso's classes (path from -source-path) */
@theme { --color-brand: oklch(0.7 0.15 250); }  /* your own tokens too */
```

Goshtoso's fonts/images are still served by `assets.Handler()` at `/assets/`,
so mount it regardless of which path you choose.

## From a component to an application

After the first component renders, choose a task-oriented pattern instead of
inventing a page from isolated demos:

| Task | Pattern | Main packages |
|---|---|---|
| Persistent product navigation | App Shell | `appshell`, `navbar`, `sidebar`, `search` |
| Search, filter, compare, act | Operations List | `pageheader`, `toolbar`, `table`, `emptystate`, `skeleton` |
| Inspect and change one resource | Detail Workspace | `pageheader`, `breadcrumbs`, `badge`, `tabs`, `button` |
| Complete a long or risky task | Multi-step Workflow | `pageheader`, `steps`, `form`, `alert`, `button` |

The installable skill includes four progressive references:

- [design intelligence](../.agents/skills/using-goshtoso/references/design-intelligence.md)
  turns the task, operating context, register, archetype, identity, states,
  density, and visual direction into a compact surface brief without using
  category-to-style presets;
- [application patterns](../.agents/skills/using-goshtoso/references/application-patterns.md)
  defines anatomy, state matrices, responsive behavior, accessibility, app
  boundaries, and completion checks for all four patterns;
- [adversarial acceptance](../.agents/skills/using-goshtoso/references/adversarial-acceptance.md)
  turns consequential state/action rules into an invariant ledger whose rows
  drive HTTP and browser tests, including denied transitions, retained drafts,
  transport failure, final identity, focus, and side-effect counts;
- [visual acceptance](../.agents/skills/using-goshtoso/references/visual-acceptance.md)
  requires 390 px and 1440 px, Goshtoso and Minimal, light and dark, keyboard,
  console, accessibility, and screenshot checks.

## Reusable documentation shells

Documentation applications can reuse
[`componentdocshell`](https://github.com/araihu/goshtoso-app-shells/tree/main/componentdocshell)
for the Goshtoso demo frame: brand header, search-first fixed navigation,
responsive drawer, appearance controls, optional table of contents, and HTMX
main-content swaps. Its config structs let the application choose the default
theme, available themes, dark-mode integration, control visibility, IDs, slots,
and navigation/search data while keeping product metadata and content local.

Use
[`componentpage`](https://github.com/araihu/goshtoso-app-shells/tree/main/componentpage)
for the semantic sections and preview/code examples repeated inside component
reference pages. It is content structure, not the outer site shell. Catalog
surfaces remain a separate application-shell pattern.

Model loading, empty, error, and success before polishing the happy path. Keep
domain vocabulary, information priority, authorization, and workflow rules in
the application. Goshtoso supplies a consistent component vocabulary, not the
product decisions.

For queue/detail fragments, treat selection as one invariant rather than a CSS
highlight: after each swap, the URL, detail identity, focus, selected-row style,
and `aria-current` or `aria-selected` must all name the same record. Rerender the
collection from server truth when practical; otherwise synchronize every
representation together on `htmx:afterSettle` and test Back/Forward as well.

`button.WithLoadingText` works when HTMX lives on the Button or on an ancestor
form. For form-owned mutations, also set
`hx-disabled-elt="find button[type='submit']"` on the form and prove the pending
label and disabled state by holding a real request. After PRG, compare the DOM
restored by Back with current server truth; use `Cache-Control: no-store` or a
`pageshow` refresh when a persisted task document may be stale.

Mobile Sidebar overlays should be viewport-owned, typically with
`PanelPositionClass: "fixed top-16 bottom-0 left-0"` and matching backdrop
geometry. Do not use `absolute top-full` inside a nested header container. At
390 px, open the drawer and assert its bounding box intersects the viewport with
positive height; `aria-expanded=true` is not sufficient visual evidence.

For operation tables, `Row.Link` and `Row.Actions` are safe to combine: the
link moves into the first data cell and the row retains native table semantics,
so trailing buttons are never nested inside a clickable row. Avoid adding
`OnClick` or row-level `HTMX` as competing navigation paths when `Link` is set.

## Component Catalog

All components are imported from `github.com/araihu/goshtoso/components/<name>`.
The public surface has 54 public component packages and 83 renderable primitives;
the demo catalog has 50 documentation pages.
Run the demo server (`go run ./site/cmd/server`) or visit
[goshtoso.araihu.com](https://goshtoso.araihu.com/) for interactive examples,
configuration previews, and API tables.

| Component | Import | Description |
|-----------|--------|-------------|
| `accordion` | `components/accordion` | Collapsible sections with default, plain, and split appearances |
| `actiongroup` | `components/actiongroup` | Responsive primary, secondary, stacked, and flat overflow actions |
| `alert` | `components/alert` | Dismissable alert banners with info, success, warning, and danger tones |
| `appshell` | `components/appshell` | Application frame with skip link, persistent regions, and one scrollable main surface |
| `avatar` | `components/avatar` | User avatar with image, initials fallback, status indicator |
| `badge` | `components/badge` | Inline status badges with independent tone, appearance, and size dimensions |
| `banner` | `components/banner` | Full-width notifications and consent dialogs as separate `Banner` and `CookieBanner` primitives |
| `breadcrumbs` | `components/breadcrumbs` | Navigation breadcrumb trail with custom separators |
| `button` | `components/button` | Buttons with tone and size options plus HTMX and Alpine.js integration |
| `card` | `components/card` | Article-like content cards with image, title, description, arbitrary body/footer content, and vertical or horizontal layout |
| `carousel` | `components/carousel` | Image carousel with autoplay, navigation, and HTMX lazy loading |
| `chatbubble` | `components/chatbubble` | Chat/message bubbles with sender alignment and avatar support |
| `checkbox` | `components/checkbox` | Checkboxes with semantic tones, group layout, and indeterminate state |
| `codeblock` | `components/codeblock` | Code display block with copy button, compact density, and max-height scrolling |
| `inlinecode` | `components/inlinecode` | Semantic inline code fragments for prose and documentation |
| `combobox` | `components/combobox` | Searchable dropdown with single/multi-select, HTMX server search |
| `drawer` | `components/drawer` | Slide-over drawers for navigation and contextual panels |
| `dropdown` | `components/dropdown` | Context menus, action menus with icons, shortcuts, sections |
| `emptystate` | `components/emptystate` | Instructive empty surfaces with optional icon and next action |
| `fileinput` | `components/fileinput` | File input controls with labels, helper text, and validation states |
| `form` | `components/form` | Form orchestrator: Section, FlipSection, CollapsibleSection, FieldGroup |
| `kbd` | `components/kbd` | Semantic keyboard shortcut and user input hints |
| `link` | `components/link` | Styled link primitives with external-link and navigation affordances |
| `modal` | `components/modal` | General and confirmation dialogs as separate `Modal` and `AlertDialog` primitives; `Tone` belongs to `AlertDialog` |
| `navbar` | `components/navbar` | Top navigation bar with links, user profile dropdown, action items |
| `pageheader` | `components/pageheader` | Page identity, breadcrumbs, description, and task-level actions |
| `panel` | `components/panel` | Neutral full-width application surface with arbitrary header, actions, body, and footer regions |
| `pagination` | `components/pagination` | Page navigation with HTMX, ellipsis, prev/next buttons |
| `palette` | `components/palette` | Color palette and swatch utilities for theme demos and pickers |
| `radio` | `components/radio` | Radio inputs and groups with validation and semantic tones |
| `range` | `components/range` | Range sliders with labels, helper text, and icon slots |
| `rating` | `components/rating` | Rating controls and display states |
| `schemaform` | `components/schemaform` | Schema Form: generate form controls from JSON Schema, defaults, current values, and allow-list rules |
| `search` | `components/search` | Search input and command-palette style result lists |
| `select` | `components/select` | HTML select dropdown with validation states, readonly mode |
| `sidebar` | `components/sidebar` | Collapsible sidebar with sections, nested items, badges |
| `skeleton` | `components/skeleton` | Accessible loading placeholders for text, rectangles, and circles |
| `spinner` | `components/spinner` | Loading spinner with independent size and tone dimensions |
| `steps` | `components/steps` | Stepper/progress navigation for multi-step flows |
| `structuredinput` | `components/structuredinput` | Repeatable structured row editor (for labels, taints, rules) |
| `table` | `components/table` | Data table with sorting, pagination, infinite scroll, filters, row links |
| `tabs` | `components/tabs` | Tab navigation with badges, HTMX lazy content loading |
| `tagslist` | `components/tagslist` | Dynamic tag list editor (add/remove string tags) |
| `textarea` | `components/textarea` | Multi-line text input with validation states |
| `textinput` | `components/textinput` | Text input with types (text, email, password, number), validation |
| `toast` | `components/toast` | Notifications as separate `Toast` and `MessageToast` primitives; sender and avatar content belongs to `MessageToast` |
| `toolbar` | `components/toolbar` | Accessible search, filter, and action regions with responsive wrapping |
| `toggle` | `components/toggle` | Toggle switch with semantic tones |
| `tooltip` | `components/tooltip` | Hover tooltips with position options, rich content support |

## Basic Component Pattern

Import the component package you need and follow its constructor contract.
Atomic primitives such as Button use functional options; provide child content
when the component accepts children:

```go
import "github.com/araihu/goshtoso/components/button"

templ Actions() {
    @button.Button(
        button.WithTone(button.TonePrimary),
        button.WithType("submit"),
    ) {
        Save changes
    }
}
```

For component-specific fields, prefer the generated Go documentation and the
demo site's API tables:

- [Go package reference](https://pkg.go.dev/github.com/araihu/goshtoso)
- [Live component docs](https://goshtoso.araihu.com/components/button)

Public config fields follow
[`docs/COMPONENT_API_NAMING.md`](COMPONENT_API_NAMING.md). Shared extension
points generally use target-specific names such as `RootClass`, `InputAttrs`,
`HTMX`, and `Alpine`.

## Theming

### Available Themes

Goshtoso ships 16 built-in themes. The default theme is `goshtoso`; the Minimal
theme is useful for checking no-radius edge cases.

### Switching Themes

Set the theme via data attribute on `<html>`:

```html
<html data-theme="modern">
```

Or with JavaScript:

```javascript
document.documentElement.setAttribute('data-theme', 'modern');
```

### Dark Mode

Add/remove the `dark` class on `<html>`:

```javascript
document.documentElement.classList.toggle('dark');
```

## Best Practices

### 1. Prefer templ components for rich content

Create separate templ components for accordion content to keep code clean:

```go
templ SettingsContent() {
	@textinput.TextInput(textinput.Config{
		ID: "settings-name", Name: "name", Label: "Name",
	})
	@textinput.TextInput(textinput.Config{
		ID: "settings-email", Name: "email", Label: "Email",
		Type: textinput.TypeEmail,
	})
}

// Use it
@accordion.Accordion(accordion.AccordionConfig{
    Items: []accordion.AccordionItem{
        {Title: "Settings", Content: SettingsContent()},
    },
})
```

### 2. Pass icons and custom slots as templ components

Many components accept icons, actions, details, or custom bodies as
`templ.Component`. Keep those as normal templ functions when possible. Use
`templ.Raw` only for trusted, static HTML or scripts you fully control.

### 3. HTMX Integration

Components work seamlessly with HTMX for dynamic updates:

```go
// Initial render
@accordion.Accordion(accordion.AccordionConfig{
    ID: "cluster-accordion",
    Items: []accordion.AccordionItem{
        {
            ID:      "node-pools",
            Title:   "Node Pools",
            Content: NodePoolsTable(cluster.NodePools),
        },
    },
})

// Update fragment via HTMX
func HandleNodePoolsUpdate(w http.ResponseWriter, r *http.Request) {
    clusterID := r.URL.Query().Get("cluster_id")
    nodePools := fetchNodePools(clusterID)
    
    // Render just the content that changed
    accordion.AccordionItem{
        ID:      "node-pools",
        Title:   "Node Pools",
        Content: NodePoolsTable(nodePools),
    }.Render(r.Context(), w)
}
```

### 4. Testing

For application tests, render your own templ pages and assert on the generated
HTML, then cover important browser behavior with Playwright or your preferred
E2E tool. The Goshtoso repository's `components/*/*_test.go` and
`site/tests/e2e/*_test.go` files are useful examples.

For HTMX validation or mutation fragments, restore focus on
`htmx:afterSettle`, after the replacement has become reliably focusable. Prefer
an explicit `[data-autofocus]` target or the rendered `FormErrors` summary, and
assert the live `document.activeElement`; `htmx:afterSwap` can be too early.

## Known Pitfalls

### HTMX History Cache vs Alpine.js State

When using HTMX SPA navigation (`hx-get` + `hx-target="#main-content-area"` + `hx-push-url`), HTMX caches the raw `document.body.innerHTML` for back-button history restore. The problem: Alpine-generated DOM nodes (from `x-for`, `x-text`, etc.) are saved in the cache, but Alpine scope objects are lost. On back-button restore, the page shows stale Alpine-generated elements with no reactivity — combobox dropdowns with blank items, broken toggles, etc.

**Recommended approaches (pick one per use case):**

1. **`LinkMode: LinkBoost`** on table rows - swaps the full `<body>` via `hx-select="body"` + `hx-target="body"`. Back-button re-fetches from server, so Alpine re-initializes cleanly. No stale cache.

2. **`LinkMode: LinkFull`** on table rows - plain `window.location.href` navigation. Simplest, safest. Use when the target page has complex Alpine state.

3. **`hx-history="false"`** on a container - tells HTMX not to cache this page. Back-button will fetch from server. Useful when you can't control the navigation source.

4. **Alpine re-init on history restore** - listen for `htmx:historyRestore` and call `Alpine.initTree(document.body)`. Works in theory but is fragile: HTMX strips `<script>` tags from cached HTML, so Alpine data registrations may be missing.

```go
// Example: table rows with boost mode (recommended for lists → detail navigation)
row := table.Row{
    ID:       "cluster-1",
    Link:     "/clusters/abc-123",
    LinkMode: table.LinkBoost,
    Cells:    cells,
}
```

### IntersectionObserver in Nested Scroll Containers

HTMX's `intersect` and `revealed` triggers use `IntersectionObserver` with the **viewport** as root. If the table is inside a container with `overflow-y-auto` (e.g., a scrollable main content area), the sentinel element may already be in the viewport even though it's scrolled out of view within its parent. The observer fires immediately or never fires on scroll.

Goshtoso's table infinite scroll sentinel includes a built-in scroll-listener fallback that attaches to the nearest `.overflow-y-auto` ancestor. This handles the nested-scroll case automatically.

If you're building custom infinite scroll outside the table component, use this pattern:

```html
<tr id="sentinel"
    hx-get="/next-page"
    hx-trigger="intersect once"
    hx-swap="outerHTML">
</tr>
<script>
// Fallback for nested scroll containers
(function() {
    var sentinel = document.getElementById('sentinel');
    if (!sentinel) return;
    var container = sentinel.closest('.overflow-y-auto');
    if (!container) return;
    function check() {
        var rect = sentinel.getBoundingClientRect();
        var cRect = container.getBoundingClientRect();
        if (rect.top < cRect.bottom + 200) {
            container.removeEventListener('scroll', check);
            htmx.trigger(sentinel, 'intersect');
        }
    }
    container.addEventListener('scroll', check);
    check(); // check immediately in case already visible
})();
</script>
```

## Troubleshooting

### Component not styled correctly?

1. Confirm `/assets/styles.css` is being served by `assets.Handler()`.
2. Confirm `@head.Dependencies()` or an equivalent `<link>` tag is present.
3. If you run a custom Tailwind build, match the version in `VERSIONS.md` and
   include Goshtoso's theme source and component `@source` path.
4. Ensure the `data-theme` attribute is set on `<html>`.

### Alpine.js not working?

1. Prefer `@head.Dependencies()` so Alpine core and plugins are loaded in the
   supported order.
2. Check browser console for Alpine errors.
3. Avoid embedding marshaled JSON directly in Alpine attributes; templ escapes
   quotes and Alpine can fail silently. Register complex behavior with
   `Alpine.data()` instead.
4. For collapse animations, ensure the collapse plugin is loaded before Alpine
   core.

### Dark mode not working?

1. Add the `dark` class to `<html>`.
2. Verify Goshtoso's CSS is loaded before app-specific overrides.
3. Check that theme state is applied before first paint if your app persists
   theme preferences.

## Examples and References

See the `/components` directory for component implementations and the demo site
for complete examples of each component.

Run the demo server:

```bash
cd /path/to/goshtoso
go run ./site/cmd/server -port 8090
```

Then visit:
- http://localhost:8090/components/accordion
- http://localhost:8090/components/table
- http://localhost:8090/examples/todo

The public documentation site is available at
[https://goshtoso.araihu.com/](https://goshtoso.araihu.com/).

## Contributing

For contribution workflow, generated-file rules, and local quality gates, see
[`CONTRIBUTING.md`](../CONTRIBUTING.md). For release expectations, see
[`docs/RELEASE_CHECKLIST.md`](RELEASE_CHECKLIST.md).

## License

MIT. See [`LICENSE`](../LICENSE).

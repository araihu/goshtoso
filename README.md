# Goshtoso

<p align="center">
  <img src="assets/images/goshtoso-art.png" alt="Goshtoso mascot" width="320" />
</p>

<p align="center">
  <a href="https://github.com/araihu/goshtoso/actions/workflows/ci.yml"><img src="https://github.com/araihu/goshtoso/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="https://app.codecov.io/gh/araihu/goshtoso"><img src="https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/guilycst/fb3843c3a13793eb6cc0af638bc00ad4/raw/coverage.json" alt="Authored Go coverage" /></a>
  <a href="https://pkg.go.dev/github.com/araihu/goshtoso"><img src="https://pkg.go.dev/badge/github.com/araihu/goshtoso.svg" alt="Go Reference" /></a>
  <a href="https://goreportcard.com/report/github.com/araihu/goshtoso"><img src="https://goreportcard.com/badge/github.com/araihu/goshtoso" alt="Go Report Card" /></a>
  <a href="https://github.com/araihu/goshtoso/releases/latest"><img src="https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/guilycst/fb3843c3a13793eb6cc0af638bc00ad4/raw/release.json" alt="Latest release" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT" /></a>
</p>

**Goshtoso** is a Go UI component library for server-rendered web apps. Add
pre-generated [templ](https://templ.guide/) components, serve the bundled
assets, and use HTMX or Alpine.js only where the interface needs interaction.

The project began as a hard fork of [PenguinUI](https://www.penguinui.com). It
now provides a Go-first component system for applications that own versioned
dependencies and render most HTML on the server.

> Goshtoso is actively evolving. The components are usable, but the API surface
> is still being refined as the library moves toward a stable public release.
> See [ROADMAP.md](ROADMAP.md) for the alpha stability policy and release path.

## Highlights

- **53 public component packages** documented across **50 documentation pages**,
	  exposing **83 renderable primitives** for composition, forms, navigation, overlays, data
  display, feedback, layout, and richer inputs.
- **Server-rendered by default** with HTMX-friendly markup and Alpine.js where
  instant local interaction makes sense.
- **Bundled assets** for Tailwind CSS, Alpine.js, HTMX, htmx extensions, fonts,
  and images. No runtime CDN dependency is required.
- **Theme system included** with light/dark support and 16 built-in themes.
- **Two-module repository**: a slim publishable library at the repo root and a
  demo/test site under `site/`.
- **Go-native examples and tests** using templ generation and Playwright-backed
  E2E coverage.

## Quick Start

Goshtoso requires **Go 1.26.5 or newer**.

Install the library:

```bash
go get github.com/araihu/goshtoso@latest
```

Mount the embedded assets in your server:

```go
package main

import (
    "net/http"

    "github.com/araihu/goshtoso/assets"
)

func main() {
    http.Handle("/assets/", assets.Handler())

    http.ListenAndServe(":8080", nil)
}
```

Include Goshtoso's CSS and JavaScript in your page shell:

```go
import "github.com/araihu/goshtoso/components/head"

templ Layout() {
    <html>
        <head>
            @head.Dependencies()
        </head>
        <body>
            { children... }
        </body>
    </html>
}
```

`head.Dependencies()` uses version-pinned unpkg URLs first and automatically
retries the matching embedded JavaScript when a CDN download fails. The CSS,
loader, first-party helpers, and fallback files still come from
`assets.Handler()`; third-party bytes are protected by generated SHA-384 SRI.
Consumers that need content-addressed cache keys can use
`assets.RuntimeHash(role)` or inspect `assets.MuambaResources()` without
copying generated paths or hashes.
For an offline PWA or desktop/mobile WebView, an air-gapped
deployment, or another application that must never request a CDN, use:

```templ
@head.Dependencies(head.WithLocalRuntime())
```

Functional options can replace the CDN and local URL of each dependency,
disable fallback, omit an application-owned runtime, or move the stylesheet,
loader, and combobox helper. See [docs/USAGE.md](docs/USAGE.md#dependency-loading-options).

Render components from their packages:

```go
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

Goshtoso components ship pre-generated, so consumers do not run
`templ generate` on the library itself. You still run `templ generate` for your
own `.templ` files.

For a task-by-task integration guide, including your first component, custom
Tailwind builds, and manual asset wiring, see [docs/USAGE.md](docs/USAGE.md). The
[Goshtoso Component Model](docs/COMPONENT_MODEL.md) documents the common
component interface, concrete return values, constructor styles, stable `Kind`
identity, and rendered defaults.
Release changes are recorded in the [changelog](CHANGELOG.md); applications
upgrading from `v0.0.11` should follow the
[component API migration guide](docs/MIGRATING_COMPONENT_API.md).
Maintainers updating repository-owned brand, theme, or UI-icon fallbacks should
follow the [immutable Arai Hu asset update contract](docs/ARAIHU_ASSETS.md).

## AI Agent Skill

Goshtoso ships an installable skill for AI coding agents that need to use the
library inside consumer applications. It teaches agents how to install the Go
module, serve bundled assets, wire `head.Dependencies()`, import components,
choose a CSS strategy, write a low-interaction surface brief, route the real
task into supported application patterns, reject generic design reflexes, and
verify the result in the browser. The composition contracts remain App Shell,
Operations List, Detail Workspace, and Multi-step Workflow.

Install the skill into a project or agent workspace:

```bash
npx skills add araihu/goshtoso --skill using-goshtoso
```

For Codex:

```bash
npx skills add araihu/goshtoso --skill using-goshtoso --agent codex
```

Use it without installing files:

```bash
npx skills use araihu/goshtoso --skill using-goshtoso
```

The skill is intentionally consumer-focused. It does not cover maintaining
Goshtoso itself, editing component internals, or running releases. The public
docs site includes an AI Agents page at `/docs/agents`. The installed skill also
ships a [design-intelligence reference](.agents/skills/using-goshtoso/references/design-intelligence.md),
an [application patterns reference](.agents/skills/using-goshtoso/references/application-patterns.md),
and a [visual acceptance checklist](.agents/skills/using-goshtoso/references/visual-acceptance.md).

For a public organization, product, or publication site, start with the
copyable [`examples/brand-site`](examples/brand-site) fixture rather than an
application shell. It generates static HTML and makes the product-owned
typography, art direction, and content hierarchy explicit. Create a fresh copy
with `go run github.com/araihu/goshtoso/cmd/goshtoso@latest -init-brand-site=./my-site`.

## Component Catalog

All components are imported from:

```text
github.com/araihu/goshtoso/components/<name>
```

Current components:

```text
accordion        actiongroup  alert        appshell     avatar       badge
banner           breadcrumbs  button       card         carousel     chatbubble
checkbox         codeblock    combobox     drawer       dropdown     emptystate
fileinput        form         head         kbd          link         modal
navbar           pageheader   pagination   palette      panel        radio
range            rating       schemaform   search       select       sidebar
skeleton         spinner      steps        structuredinput table     tabs
tagslist         textarea     textinput    toast        toolbar      toggle
tooltip
```

Run the demo site to explore component options, API tables, HTMX
behavior, Alpine.js states, themes, and example apps:

```bash
go run ./site/cmd/server
```

Then open:

- <http://localhost:8090/getting-started>
- <http://localhost:8090/components/accordion>
- <http://localhost:8090/examples/todo>
- <http://localhost:8090/examples/logs>

The public documentation site is available at
[https://goshtoso.araihu.com/](https://goshtoso.araihu.com/).

## Using Assets

The recommended path is to serve Goshtoso's embedded assets:

```go
mux := http.NewServeMux()
mux.Handle("GET /assets/", assets.Handler())
```

and let `@head.Dependencies()` emit the matching stylesheet and script tags.
The default loader tries version-pinned CDN URLs for Alpine.js and HTMX, then
falls back to the same versions under `/assets/js/runtime/`. Use
`head.WithLocalRuntime()` when an offline application such as a PWA or native
WebView must be fully local.

`assets.RuntimeHash(role)` exposes normalized SHA-384 hashes for vendored
runtime roles. `assets.MuambaHash(resource, download)` and
`assets.MuambaResources()` expose the complete embedded acquisition inventory.

If you maintain a custom Tailwind build, Goshtoso also ships a CLI that extracts
the compiled CSS or theme source:

```bash
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -out=css/goshtoso-base.css
```

See [docs/USAGE.md](docs/USAGE.md) for the full asset strategy.
Release maintainers should also use
[docs/RELEASE_CHECKLIST.md](docs/RELEASE_CHECKLIST.md) before tagging.

## Sprite Icons

`components/icon` renders accessible SVG `<use>` references. The bundled
`components/icon/heroicons` package provides typed symbols and a same-origin
default sprite URL:

```templ
import (
    "github.com/araihu/goshtoso/components/icon"
    "github.com/araihu/goshtoso/components/icon/heroicons"
)

templ SaveIcon() {
    @icon.Icon(icon.Config{
        SpriteURL: heroicons.SpriteURL,
        Symbol:    heroicons.Icon16SolidCheck,
        Label:     "Saved",
    })
}
```

Use a relative, same-origin sprite URL by default. `ModeInline` resolves an
already-present symbol from the current document; cross-origin external sprites
depend on browser support and CORS, and HTTPS pages should not reference an HTTP
sprite. A blank label and `Decorative: true` both produce a decorative icon.
See [docs/USAGE.md](docs/USAGE.md#sprite-icons) for generator and deployment
details.

## Repository Layout

```text
goshtoso/
├── cmd/                     # Thin command entry points
├── components/              # Publishable component library
├── assets/                  # Embedded CSS, JS, fonts, and images
├── css/                     # Tailwind source
├── docs/                    # Consumer and project documentation
├── examples/                # Standalone examples
├── internal/                # Generator internals used by cmd/* tools
└── site/                    # Demo site, example app pages, server, E2E tests
```

The root module is `github.com/araihu/goshtoso`. The `site/` directory is a
separate module for the demo website and test harness.

For local development, create a Go workspace once per clone so the site imports
your working-tree copy of the library:

```bash
go work init . ./site
```

## Development

Useful commands from the repo root:

```bash
# Generate *_templ.go files after editing .templ sources
templ generate
# or
just gp-generate

# Rebuild the embedded Tailwind CSS after editing CSS/theme sources
just css

# Run the demo server on :8090
go run ./site/cmd/server
# or
just gp-dev

# Build the demo server
go build -o bin/server ./site/cmd/server
```

Run tests:

```bash
# Library tests
go test ./...

# Site tests
cd site && go test ./...

# Full Playwright E2E suite
go test -tags=e2e,full ./site/tests/e2e/... -count=1 -timeout 15m

# Release-equivalent unit + Playwright coverage
just coverage
```

Release coverage keeps two reports. When `CODECOV_TOKEN` is configured, the
public [Codecov report](https://app.codecov.io/gh/araihu/goshtoso) measures
authored Go source and excludes only generated `*_templ.go` files. Release
artifacts always retain the full generated-inclusive profile and HTML report.
Both reports come from the same root, site, and real-browser test run.

Run lint checks per module:

```bash
golangci-lint run
cd site && golangci-lint run
```

Generated files are part of the repo, but should not be edited by hand:

- `*_templ.go` is generated by `templ generate`
- `assets/styles.css` is generated by `just css`

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines and
[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community expectations.

When adding or changing components, keep the component source, demo page, E2E
coverage, generated templ output, CSS output, and usage reference in sync.

## Credits

Goshtoso began as a hard fork of [PenguinUI](https://www.penguinui.com) by
[Salar Houshvand](https://x.com/salar_houshvand), transformed from static
HTML/Alpine.js examples into an importable Go component library.

## License

MIT. See [LICENSE](LICENSE).

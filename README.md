# Goshtoso

<p align="center">
  <img src="assets/images/goshtoso-art.png" alt="Goshtoso mascot" width="320" />
</p>

<p align="center">
  <a href="https://github.com/araihu/goshtoso/actions/workflows/ci.yml"><img src="https://github.com/araihu/goshtoso/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="https://pkg.go.dev/github.com/araihu/goshtoso"><img src="https://pkg.go.dev/badge/github.com/araihu/goshtoso.svg" alt="Go Reference" /></a>
  <a href="https://goreportcard.com/report/github.com/araihu/goshtoso"><img src="https://goreportcard.com/badge/github.com/araihu/goshtoso" alt="Go Report Card" /></a>
  <a href="https://github.com/araihu/goshtoso/releases"><img src="https://img.shields.io/github/v/tag/araihu/goshtoso?label=release&sort=semver" alt="Latest tag" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT" /></a>
</p>

⚠️ Work-In-Progress

There is still lots of rough edeges to iron out, most related to preserving Alpine.js state when a wired component is swapped in by HTMX.

**Goshtoso**: Go + Templ + Tailwind CSS + HTMX + Alpine.js

## About This Fork

This is a hard fork of [Penguin UI](https://www.penguinui.com) by Salar Houshvand ([source here](https://github.com/SalarHoushvand/penguinui-components)), transformed from static HTML/Alpine.js components into a complete Go web component library.


### What's Changed?

| Original | Goshtoso Fork |
|----------|-------------|
| Static HTML | Go + Templ templates |
| CDN assets | Configurable (CDN/Embedded/Custom) |
| Copy-paste | `go get` importable |
| Alpine.js only | HTMX + Alpine.js + Go backend |

## Credits

- **Original**: [Penguin UI](https://www.penguinui.com) by [Salar Houshvand](https://x.com/salar_houshvand)
- **License**: MIT (preserved from original)

## CSS Integration

Goshtoso ships a CLI tool that extracts the pre-built Tailwind CSS from the embedded assets. Client applications use this instead of manually copying CSS files.

```bash
# Via go tool (recommended — version-pinned in go.mod)
go tool goshtoso -out=css/goshtoso-base.css

# Or via go run (for one-off use)
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -out=goshtoso-base.css
```

Then import it in your Tailwind entry point:

```css
@import "tailwindcss";
@import "./goshtoso-base.css";
```

See [docs/USAGE.md](docs/USAGE.md) for full setup instructions.

## Project Structure

Two Go modules in one repo (a workspace): the library at the root and the demo
site under `site/`.

```
goshtoso/                    # ROOT MODULE — library (github.com/araihu/goshtoso)
├── cmd/goshtoso/            # CSS extraction CLI tool
├── components/              # Goshtoso component library (32 components)
│   └── badge/
│       ├── types.go         # Configuration types
│       └── badge.templ      # Templ component
├── assets/
│   ├── embed.go             # Embedded assets + StylesCSS() accessor
│   ├── styles.css           # Compiled Tailwind CSS
│   ├── js/                  # Alpine.js, HTMX, plugins
│   └── fonts/               # TOTVS brand fonts
├── docs/                    # Integration guides
└── site/                    # SITE MODULE — demo + examples (…/goshtoso/site)
    ├── cmd/server/          # Demo server
    ├── internal/pages/demo/ # Demo pages
    └── tests/e2e/           # Playwright E2E tests
```

## Running the Demo

### Quick Start (Recommended: Air with Live Reload)

```bash
# Install dependencies (including Air)
make install
make install-air

# Run with live reload (auto-rebuilds on file changes)
make dev-air

# Server will start on http://localhost:8090
# Accordion Demo: http://localhost:8090/components/accordion
# Button Demo: http://localhost:8090/components/button
```

### Standard Development

```bash
# Install dependencies
make install

# Run the demo server (run `go work init . ./site` once per clone for local dev)
make dev
# or
go run ./site/cmd/server

# Server will start on http://localhost:8090
# - Original PenguinUI: http://localhost:8090/original/
# - Goshtoso Components: http://localhost:8090/gottha/
```

## Running E2E Tests

E2E tests use [playwright-go](https://github.com/playwright-community/playwright-go) (following the tks-console pattern) for Go-based browser automation.

```bash
# Using just (from repo root)
just gp-test-e2e                    # Run all E2E tests
just gp-test-e2e-one TestButton     # Run specific test

# Or directly
go test ./site/tests/e2e/... -v

# First time setup - install Playwright browsers
just gp-install-playwright
# or
go install github.com/playwright-community/playwright-go/cmd/playwright@v0.5700.1
playwright install chromium
```

### Test Results

Tests automatically:
- Start the demo server
- Run browser automation tests
- Capture screenshots on failures to `test-results/screenshots/`
- Verify both Original PenguinUI and Goshtoso component rendering

### Current Test Coverage

- **Button Component**: Verifies all 8 variants render correctly, HTMX attributes, Alpine.js integration
- **Screenshots**: Auto-captured for visual debugging

## Component Usage

### Button Component

```go
import "github.com/araihu/goshtoso/components/button"

// Basic button
@button.Button(button.Config{
    Variant: button.Primary,
    Type:    "button",
}) {
    Click Me
}

// With HTMX
@button.Button(button.Config{
    Variant: button.Primary,
    HTMX: &button.HTMXConfig{
        Post:   "/api/action",
        Target: "#result",
        Swap:   "innerHTML",
    },
}) {
    Submit
}

// With Alpine.js
@button.Button(button.Config{
    Variant: button.Primary,
    Alpine: &button.AlpineConfig{
        OnClick: "modalIsOpen = true",
    },
}) {
    Open Modal
}
```

## Development

### Building

```bash
make build
```

### Generating Templ Files

```bash
make generate
```

### Testing

```bash
# Go tests
make test

# E2E tests
make test-e2e
```

## License

MIT License - See original [Penguin UI](https://www.penguinui.com/docs/license) for details.

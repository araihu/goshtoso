# Contributing to Goshtoso

Thanks for your interest! Goshtoso is an alpha-stage, MIT-licensed UI component
library (Go + templ + Tailwind CSS v4 + HTMX + Alpine.js) — a hard fork of
[PenguinUI](https://penguinui.com) targeting 99.99% visual parity.

Contributions of all kinds are welcome: bug reports, new components, parity
fixes, docs, and tests.

## Code of Conduct

This project ships a [Code of Conduct](CODE_OF_CONDUCT.md). By participating you
agree to uphold it.

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.26+ | https://go.dev/dl |
| templ | v0.3.x | `go install github.com/a-h/templ/cmd/templ@latest` |
| Tailwind CSS | v4 | standalone CLI or `npm i` |
| golangci-lint | latest | https://golangci-lint.run |
| Playwright (E2E) | v0.5700.1 | resolved by `go test ./tests/e2e/...` |

## Getting started

```bash
git clone https://github.com/araihu/goshtoso
cd goshtoso

# enable the pre-commit hook (runs go fix + regenerates skill reference)
git config core.hooksPath .githooks

# run the dev server (port 8090)
go run cmd/server/main.go
```

## Development workflow

Generated artifacts are checked in but **must never be hand-edited**
(`*_templ.go`, `assets/styles.css`). Regenerate them:

```bash
# after editing any .templ file
templ generate

# after editing CSS / introducing a new Tailwind utility class
tailwindcss -i css/main.css -o assets/styles.css

go build -o bin/server ./cmd/server
```

### Adding a component

See `CLAUDE.md` (the full agent/contributor guide) and the `component-docs`
convention. Every component lives in `components/<name>/` (`types.go` +
`<name>.templ`) and ships a demo page that follows the docs-page pattern
(one preview + code box per variant, an API reference table, the right-rail
TOC). After changing a component, regenerate the usage reference:

```bash
go run ./scripts/skillgen
```

### Frontend interactivity hierarchy

Pick the **highest** tier that solves the problem: **htmx (SSR) → Alpine.js →
vanilla JS**. Most "interactive" needs (filtering, pagination, inline edit,
lazy load, validation, toasts) are SSR + htmx, not JavaScript. Anything you add
must survive HTMX swaps and fragment navigation.

> ⚠️ **Templ + Alpine escaping is the #1 source of bugs.** Never use
> `json.Marshal` for data that lands inside an HTML attribute — templ escapes
> the quotes and Alpine silently fails. See `CLAUDE.md` for the safe patterns.

## Before you open a PR

Run the full gate locally — CI enforces all of it:

```bash
templ generate
golangci-lint run                                   # cyclomatic ceiling: 20
go fix ./...                                         # also runs via pre-commit hook
go build -o bin/server ./cmd/server
go test ./... -count=1                               # unit tests
go test ./tests/e2e/... -count=1 -timeout 15m        # full E2E (~2.5 min)
```

Test components in **both light and dark mode** across themes (especially
**Minimal**, which has no border-radius).

## Pull requests

1. Fork and branch from `main` (`fix/...`, `feat/...`).
2. Keep PRs focused; one logical change per PR.
3. Fill out the PR template checklist.
4. A Codex review and CI (`lint-build` + E2E) must pass before merge.

## Reporting bugs / requesting features

Use the [issue templates](.github/ISSUE_TEMPLATE). For questions, open a
[Discussion](https://github.com/araihu/goshtoso/discussions). For security
issues, see [SECURITY.md](SECURITY.md) — do **not** file a public issue.

## License

By contributing, you agree your contributions are licensed under the
[MIT License](LICENSE).

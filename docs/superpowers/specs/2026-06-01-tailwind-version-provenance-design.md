# Tailwind version provenance + consumer Tailwind integration — Design

**Date:** 2026-06-01
**Status:** Approved (brainstorm), pending implementation plan
**Branch context:** follows PR #17 (`fix(head): serve local /assets/* instead of dead CDN tags`)

## Problem

A consumer who serves Goshtoso's prebuilt `assets/styles.css` (PR #17) is fully
covered. But a consumer who wants to run **their own Tailwind build** alongside
Goshtoso — to add their own utility classes, tokens, or purge — hits friction:

1. **No discoverable Tailwind version.** Nothing in the module, CLI, or docs
   tells the consumer which Tailwind version Goshtoso's CSS was built with, so
   they can't keep their own build on a compatible v4 line. The only marker that
   ever existed (`head.TailwindVersion = "4.2.2"`) was hand-typed, **drifted**
   (CI actually builds with `v4.3.0`), fed a dead CDN URL, and was deleted in
   PR #17. The lesson: any version marker must be *derived from the build*, not
   hand-maintained.

2. **No exposed token source.** The embedded assets ship only the *compiled*
   `styles.css`. A consumer doing a unified Tailwind build needs Goshtoso's
   **source** theme layer (`@theme` tokens + `@custom-variant dark` + the 13
   `[data-theme]` blocks) to `@import`, plus a way to point Tailwind's `@source`
   at Goshtoso's templates so the JIT emits Goshtoso's classes.

3. **No release process.** `v0.0.1` was hand-tagged. Nothing records the
   `goshtoso tag → tailwind version` mapping or ships the CSS as a release asset.

4. **Unpinned local build.** CI pins Tailwind (`TAILWIND_VERSION: v4.3.0`,
   duplicated across 3 jobs) and guards drift via rebuild + `git diff
   --exit-code`. But the local dev build (per CLAUDE.md) uses *whatever*
   `tailwindcss` the dev has installed — an unpinned surface that can silently
   produce drift only caught later in CI.

## What already exists (do not rebuild)

- **CI pin + drift guard.** `.github/workflows/ci.yml` curls the exact standalone
  Tailwind binary at `TAILWIND_VERSION`, rebuilds `assets/styles.css`, and fails
  on `git diff`. This pattern is reused, not replaced — it is just rewired to a
  single source of truth and DRY'd.
- **CSS source layout.** `all-themes.css` is already Tailwind-free (`@theme {…}`
  at line 1 + `@layer base { [data-theme=…] {…} }`). `css/main.css` =
  `@custom-variant dark` + `@import "tailwindcss"` + repo `@source` globs +
  `@import "../all-themes.css"` + `@import "./codeblock.css"` + a `@theme` block.
  The consumer theme file is therefore `main.css` **minus** the tailwind import
  and the repo-relative `@source` lines.
- **Extractor CLI.** `cmd/goshtoso` already extracts `assets.StylesCSS()` to disk
  via `-out`. Extended here, not rewritten.
- **`assets` embed + accessors.** `assets/embed.go` embeds `styles.css js fonts
  images` and exposes `Handler()` + `StylesCSS()`. Extended here.

## Goals / non-goals

**Goals**
- One drift-proof source of truth for the Tailwind version, surfaced to Go code,
  the CLI, consumers, CI, and releases.
- Expose Goshtoso's token *source* so a consumer can run a unified Tailwind build.
- Document both consumer integration paths with their trade-offs.
- A tag-driven GitHub release that ships the CSS assets and records the version
  mapping.
- A reproducible *local* build pinned to the same version as CI.

**Non-goals**
- Supporting Tailwind v3 or non-v4 builds.
- Vendoring the Tailwind standalone binary in the tree (explicitly rejected —
  platform-specific, bloat, churn; fetch-on-demand keyed by the pin instead).
- Changing how PR #17's `head.Dependencies()` / `assets.Handler()` work.
- A JS/npm Tailwind toolchain (the repo uses the standalone CLI; keep it).

## Design

### A. Single source of truth — `.tailwind-version`

A plain-text file at repo root containing exactly the version, no `v` prefix:

```
4.3.0
```

Everything reads this one file:
- CI installs `tailwindcss-linux-x64` at `v$(cat .tailwind-version)` (replacing the
  three duplicated `TAILWIND_VERSION` env entries).
- The local `just css` target reads it.
- `assets` embeds it.
- The release workflow reads it.

The pin *causes* every build, so the reported number cannot disagree with the
artifact.

### B. Surface the version

- **Go accessor.** `assets/embed.go` adds the file to `//go:embed` and exposes:

  ```go
  // TailwindVersion returns the Tailwind CSS version that styles.css was
  // compiled with (the .tailwind-version pin), e.g. "4.3.0".
  func TailwindVersion() string // strings.TrimSpace of embedded .tailwind-version
  ```

  This redeems the deleted `head.TailwindVersion` const — now derived from the
  pin, not hand-typed.

- **CLI `-version`.** `cmd/goshtoso -version` prints the Goshtoso module version
  (from `runtime/debug.ReadBuildInfo`, falling back to `"(devel)"`) and
  `assets.TailwindVersion()`:

  ```
  goshtoso v0.0.1 (tailwindcss 4.3.0)
  ```

- **`VERSIONS.md`.** A generated table at repo root mapping each released
  Goshtoso tag to its Tailwind version, appended by the release workflow (F):

  ```
  | Goshtoso | Tailwind CSS |
  |----------|--------------|
  | v0.0.1   | 4.3.0        |
  ```

### C. Path B token source (unified consumer build)

- **New file `css/goshtoso-theme.css`** — the consumer-importable theme layer.
  It is `main.css` with the Tailwind import and repo `@source` lines removed:

  ```css
  /* Goshtoso theme layer — import into your OWN Tailwind v4 build.
     Do NOT @import "tailwindcss" here; your build already does. */
  @custom-variant dark (&:where(.dark, .dark *));
  @import "./all-themes.css";   /* @theme tokens + 13 [data-theme] blocks */
  @import "./codeblock.css";
  ```

  `all-themes.css` and `codeblock.css` are referenced as siblings, so the
  extracted bundle must inline them (the CLI extractor concatenates, see below)
  — the consumer receives a single self-contained file.

- **Embed + accessor.** `assets/embed.go` embeds the theme source and exposes:

  ```go
  // ThemeCSS returns the Goshtoso theme layer source (tokens + variants + the
  // 13 themes) for importing into a consumer's own Tailwind v4 build. Unlike
  // StylesCSS (compiled output), this is source to be compiled by the consumer.
  func ThemeCSS() ([]byte, error)
  ```

  Because Tailwind's standalone CLI does not resolve `@import` of arbitrary files
  the same way across setups, `ThemeCSS()` returns a **single concatenated**
  document (theme header + inlined `all-themes.css` + inlined `codeblock.css`),
  produced at build time and embedded. Implementation detail for the plan:
  generate `assets/goshtoso-theme.css` from the three sources via a small
  `go:generate` step or the `just css` target, embed that.

- **CLI `-theme`.** `goshtoso -theme -out=css/goshtoso-theme.css` writes
  `assets.ThemeCSS()` to disk (parallel to the existing CSS extraction).

- **`-source-path` helper.** The consumer's Tailwind must scan Goshtoso's
  templates so the JIT emits Goshtoso's classes. `goshtoso -source-path` prints
  the absolute path to the installed module's `components` dir (resolved via
  `go list -m -f '{{.Dir}}' github.com/araihu/goshtoso` semantics; in practice
  the CLI runs from the consumer's module and shells `go list`). Consumer wiring:

  ```css
  /* consumer main.css */
  @import "tailwindcss";
  @import "./goshtoso-theme.css";              /* from goshtoso -theme */
  @source "<output of: goshtoso -source-path>"; /* emits goshtoso classes */
  @theme { --color-brand: oklch(…); }           /* consumer's own tokens */
  ```

### D. Reproducible local build — `just css`

A new justfile target replaces the unpinned manual `tailwindcss …` invocation in
CLAUDE.md's quick-commands:

```
css:
    #!/usr/bin/env bash
    set -euo pipefail
    ver="$(cat .tailwind-version)"
    bin=".tools/tailwindcss-${ver}"
    if [ ! -x "$bin" ]; then
      mkdir -p .tools
      os="$(uname -s | tr '[:upper:]' '[:lower:]')"   # darwin|linux
      arch="$(uname -m)"; case "$arch" in arm64|aarch64) arch=arm64;; x86_64) arch=x64;; esac
      curl -fsSL -o "$bin" \
        "https://github.com/tailwindlabs/tailwindcss/releases/download/v${ver}/tailwindcss-${os}-${arch}"
      chmod +x "$bin"
    fi
    "$bin" -i css/main.css -o assets/styles.css
```

- `.tools/` is gitignored. **No binary committed.**
- Same version as CI → local rebuild matches CI rebuild → drift guard never
  surprises a contributor.
- CLAUDE.md's "Build Tailwind CSS" command updates to `just css`.

The theme-source generation (C) hangs off the same target (regenerate
`assets/goshtoso-theme.css` from the three source files before/alongside the
compile) so the embedded theme bundle stays in sync.

### E. Drift guard (DRY the existing check)

- CI's three duplicated `TAILWIND_VERSION` env entries collapse to reading
  `.tailwind-version` (single source). The existing rebuild + `git diff
  --exit-code -- assets/styles.css` step is kept.
- Add a guard for the generated theme bundle:
  `git diff --exit-code -- assets/goshtoso-theme.css`.
- Optional belt: stamp a `/* tailwindcss <pin> */` header comment on the compiled
  CSS and assert it equals the pin, catching a mismatched local build before the
  full diff. (Plan may drop this if the diff guard is deemed sufficient.)

### F. Release workflow — `.github/workflows/release.yml`

Triggered on tag push matching `v*`:

1. Checkout, set up Go, install Tailwind at `v$(cat .tailwind-version)`.
2. `templ generate`, `just css` (rebuild styles.css + theme bundle).
3. Verify clean working tree (`git diff --exit-code`) — fail if the tag's
   committed artifacts are stale.
4. Append/refresh the `tag → tailwind` row in `VERSIONS.md` (commit back to the
   default branch or include in release notes; plan decides — default: include
   in release notes, and a separate maintenance commit updates the table).
5. Create the GitHub release with:
   - Release notes including a line: `Built with Tailwind CSS <pin>.`
   - Attached assets: `assets/styles.css`, `assets/goshtoso-theme.css`.

### G. Docs

Update `docs/USAGE.md` and `.claude/skills/using-goshtoso/SKILL.md` with a
"Using your own Tailwind" section presenting both paths:

- **Path A — two stylesheets (low coupling, recommended for most).** Serve
  Goshtoso's prebuilt `styles.css` via `assets.Handler()` (PR #17) and run your
  own Tailwind for your own classes into a *separate* output. No recompiling
  Goshtoso. Keep your Tailwind on the version from `goshtoso -version`.
- **Path B — unified build (single tree-shaken CSS).** Import
  `goshtoso-theme.css` + `@source` Goshtoso's templates into your own Tailwind
  build (section C wiring). Requires matching Tailwind v4 version.

Both reference version discovery (`goshtoso -version`, `VERSIONS.md`) and the
`@source` module-cache trick. Cross-link from the PR #17 "Page setup" section.

## Affected units

| Unit | Change |
|------|--------|
| `.tailwind-version` (new) | Source-of-truth pin file. |
| `assets/embed.go` | Embed pin + theme bundle; add `TailwindVersion()`, `ThemeCSS()`. |
| `assets/goshtoso-theme.css` (generated) | Concatenated theme source, embedded. |
| `css/goshtoso-theme.css` (new source) | Tailwind-free theme entry (header for the bundle). |
| `cmd/goshtoso/main.go` | Add `-version`, `-theme`, `-source-path` modes. |
| `justfile` | New `css` target (pinned fetch + compile + theme regen). |
| `.github/workflows/ci.yml` | Read pin file (DRY); add theme-bundle drift guard. |
| `.github/workflows/release.yml` (new) | Tag-driven release with CSS assets + version notes. |
| `VERSIONS.md` (new) | `goshtoso → tailwind` mapping table. |
| `docs/USAGE.md`, `using-goshtoso/SKILL.md` | Path A / Path B integration docs. |
| `CLAUDE.md` | "Build Tailwind CSS" command → `just css`; note `.tailwind-version`. |

## Risks / open questions

- **`@source` ergonomics (Path B).** Pointing Tailwind at the module cache is
  inherently awkward; `-source-path` mitigates but does not eliminate it. Path A
  remains the recommended default for most consumers.
- **`ThemeCSS()` concatenation vs `@import`.** Concatenating at build time avoids
  the consumer's Tailwind needing to resolve relative `@import`s of files it
  doesn't have. Confirm the standalone CLI accepts the concatenated document with
  the `@theme`/`@custom-variant`/`@layer` ordering preserved.
- **VERSIONS.md write-back.** Whether the release workflow commits to the default
  branch (needs a token + protected-branch consideration) or only writes release
  notes. Default to release-notes-only + manual table maintenance to avoid CI
  write permissions; revisit if it becomes a chore.
- **Scope.** ~10 touched units in one spec. They are tightly coupled around the
  pin, so this is treated as one implementation plan; the plan may sequence it as
  (1) pin + surface, (2) theme source + CLI, (3) build + CI, (4) release, (5)
  docs.

## Success criteria

- `goshtoso -version` prints the Goshtoso + Tailwind versions; both derive from
  build/embed, never hand-typed.
- `.tailwind-version` is the only place the number is written; CI, local build,
  Go, and release all read it.
- A consumer can produce a working unified Tailwind build following the Path B
  docs, and a working two-stylesheet setup following Path A.
- CI fails if `assets/styles.css` or `assets/goshtoso-theme.css` is stale vs a
  pinned rebuild.
- Tagging `vX.Y.Z` produces a GitHub release with the CSS assets attached and the
  Tailwind version recorded.

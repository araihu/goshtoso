# Tailwind Version Provenance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give Goshtoso a single drift-proof Tailwind version pin, surface it to Go/CLI/consumers/CI/releases, expose the theme source for consumers running their own Tailwind, and ship a tag-driven release.

**Architecture:** One embeddable pin file `assets/tailwind.version` is the single source of truth — CI, the local `just css` build, the Go `assets` package, and the release workflow all read it. A code generator (`scripts/themegen`) produces an embeddable, self-contained theme-source CSS (`assets/goshtoso-theme.css`) by stripping the Tailwind import + repo scan globs from `css/main.css` and inlining its two relative `@import`s. The `goshtoso` CLI gains `-version`, `-theme`, and `-source-path`. CI's existing rebuild+diff drift guard is DRY'd to read the pin and extended to the new generated artifacts. A new release workflow rebuilds with the pin and attaches the CSS assets.

**Tech Stack:** Go 1.26, templ, Tailwind CSS v4 standalone CLI, `just`, GitHub Actions, `go:embed`.

**Spec:** `docs/superpowers/specs/2026-06-01-tailwind-version-provenance-design.md`

**Refinements from spec (intentional):**
- Pin lives at `assets/tailwind.version` (not repo root) so `//go:embed` reaches it directly — single source, no copy step. `just`/CI `cat assets/tailwind.version`.
- The theme bundle is produced by a Go generator (`scripts/themegen`, mirroring `scripts/skillgen`) rather than a hand-maintained `css/goshtoso-theme.css`, avoiding two-file drift. The generated file `assets/goshtoso-theme.css` is committed and CI-diff-guarded.

---

## File Structure

| File | Responsibility |
|------|----------------|
| `assets/tailwind.version` (new) | Single source-of-truth pin, e.g. `4.3.0`. Embedded + read by `just`/CI/release. |
| `assets/embed.go` (modify) | Embed pin + theme bundle; add `TailwindVersion()`, `ThemeCSS()`. |
| `assets/version_test.go` (new) | Test `TailwindVersion()`, `ThemeCSS()`. |
| `scripts/themegen/main.go` (new) | Generate `assets/goshtoso-theme.css` from `css/main.css` + sources. |
| `scripts/themegen/gen.go` (new) | Pure `generateTheme()` transform (testable). |
| `scripts/themegen/gen_test.go` (new) | Test the transform. |
| `assets/goshtoso-theme.css` (new, generated) | Self-contained theme source for consumer Tailwind builds. |
| `cmd/goshtoso/main.go` (modify) | Add `-version`, `-theme`, `-source-path`; factor testable helpers. |
| `cmd/goshtoso/main_test.go` (new) | Test `versionString()`, `sourcePath()`. |
| `justfile` (modify) | `css` target: pinned fetch + theme regen + compile. |
| `.gitignore` (modify) | Ignore `.tools/`. |
| `.github/workflows/ci.yml` (modify) | Read pin (DRY); regen theme; diff-guard new artifacts. |
| `.github/workflows/release.yml` (new) | Tag-driven release with CSS assets + version notes. |
| `VERSIONS.md` (new) | `goshtoso tag → tailwind version` table. |
| `docs/USAGE.md` (modify) | Path A / Path B consumer Tailwind docs. |
| `.claude/skills/using-goshtoso/SKILL.md` (modify) | Same, condensed. |
| `CLAUDE.md` (modify) | Build command → `just css`; pin note. |

**Module path:** `github.com/araihu/goshtoso`. Run all commands from repo root.

---

### Task 1: Pin file + `assets.TailwindVersion()`

**Files:**
- Create: `assets/tailwind.version`
- Modify: `assets/embed.go`
- Test: `assets/version_test.go`

- [ ] **Step 1: Create the pin file**

The value must match what CI currently installs (`ci.yml` env `TAILWIND_VERSION: 'v4.3.0'`), without the `v`:

```bash
printf '4.3.0\n' > assets/tailwind.version
```

- [ ] **Step 2: Write the failing test**

Create `assets/version_test.go`:

```go
package assets

import (
	"strings"
	"testing"
)

func TestTailwindVersion(t *testing.T) {
	got := TailwindVersion()
	if got != "4.3.0" {
		t.Fatalf("TailwindVersion() = %q, want %q", got, "4.3.0")
	}
	if strings.HasPrefix(got, "v") {
		t.Fatalf("TailwindVersion() must not include a leading v: %q", got)
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./assets/ -run TestTailwindVersion`
Expected: FAIL — `undefined: TailwindVersion`.

- [ ] **Step 4: Implement**

In `assets/embed.go`, extend the embed directive and add the accessor. Change:

```go
//go:embed styles.css js fonts images
var files embed.FS
```

to:

```go
//go:embed styles.css goshtoso-theme.css tailwind.version js fonts images
var files embed.FS
```

Add after `StylesCSS`:

```go
// TailwindVersion returns the Tailwind CSS version that styles.css and
// goshtoso-theme.css were built with — the single-source pin in
// assets/tailwind.version (e.g. "4.3.0", no leading "v"). Match your own
// Tailwind build to this when compiling Goshtoso's theme source yourself.
func TailwindVersion() string {
	b, err := files.ReadFile("tailwind.version")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
```

Add `"strings"` to the import block.

> NOTE: This task references `goshtoso-theme.css` in the embed directive, created in Task 2. To keep Task 1 compiling on its own, create a placeholder now:
> ```bash
> printf '/* generated in Task 2 */\n' > assets/goshtoso-theme.css
> ```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./assets/ -run TestTailwindVersion`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add assets/tailwind.version assets/embed.go assets/version_test.go assets/goshtoso-theme.css
git commit -m "feat(assets): add tailwind.version pin + TailwindVersion() accessor"
```

---

### Task 2: Theme generator + `assets.ThemeCSS()`

**Files:**
- Create: `scripts/themegen/gen.go`, `scripts/themegen/gen_test.go`, `scripts/themegen/main.go`
- Modify: `assets/embed.go`, `assets/version_test.go`
- Generated: `assets/goshtoso-theme.css`

- [ ] **Step 1: Write the failing transform test**

Create `scripts/themegen/gen_test.go`:

```go
package main

import (
	"strings"
	"testing"
)

func TestGenerateTheme(t *testing.T) {
	mainCSS := `@custom-variant dark (&:where(.dark, .dark *));
@import "tailwindcss";
@source "../components/**/*.templ";
@source inline("md:w-64");
@import "../all-themes.css";
@import "./codeblock.css";
@theme { --font-body: x; }`

	imports := map[string]string{
		"all-themes.css": "@theme { --color-primary: red; }",
		"codeblock.css":  ".ch-x { color: red; }",
	}
	out := generateTheme(mainCSS, imports)

	if strings.Contains(out, `@import "tailwindcss"`) {
		t.Error("tailwind import must be stripped")
	}
	if strings.Contains(out, `@source "../components`) {
		t.Error("repo path @source globs must be stripped")
	}
	if !strings.Contains(out, `@source inline("md:w-64")`) {
		t.Error("@source inline safelists must be kept")
	}
	if !strings.Contains(out, "--color-primary: red") {
		t.Error("all-themes.css must be inlined")
	}
	if !strings.Contains(out, ".ch-x") {
		t.Error("codeblock.css must be inlined")
	}
	if !strings.Contains(out, "@custom-variant dark") {
		t.Error("@custom-variant must be preserved")
	}
	if !strings.Contains(out, "--font-body: x") {
		t.Error("@theme blocks must be preserved")
	}
	if strings.Contains(out, `@import "../all-themes.css"`) || strings.Contains(out, `@import "./codeblock.css"`) {
		t.Error("relative @imports must be replaced, not left in place")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./scripts/themegen/ -run TestGenerateTheme`
Expected: FAIL — `undefined: generateTheme`.

- [ ] **Step 3: Implement the transform**

Create `scripts/themegen/gen.go`:

```go
package main

import "strings"

// generateTheme produces the consumer-importable Goshtoso theme source from the
// contents of css/main.css. It removes the Tailwind import (the consumer's own
// build provides it) and the repo-relative @source scan globs (meaningless
// outside this repo), and inlines the two relative @imports from imports[name]
// so the output is a single self-contained file. @source inline(...) safelists,
// @custom-variant, @theme, @layer, and @font-face are preserved verbatim.
func generateTheme(mainCSS string, imports map[string]string) string {
	var b strings.Builder
	b.WriteString("/* GENERATED by scripts/themegen — DO NOT EDIT.\n")
	b.WriteString("   Goshtoso theme source for your OWN Tailwind v4 build.\n")
	b.WriteString("   Import AFTER `@import \"tailwindcss\";` in your CSS. */\n\n")

	for _, line := range strings.Split(mainCSS, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == `@import "tailwindcss";`:
			continue
		case strings.HasPrefix(trimmed, `@source "`):
			// path glob like @source "../components/**" — drop.
			// (@source inline(...) starts with `@source inline(`, not `@source "`.)
			continue
		case trimmed == `@import "../all-themes.css";`:
			b.WriteString(imports["all-themes.css"])
			b.WriteByte('\n')
		case trimmed == `@import "./codeblock.css";`:
			b.WriteString(imports["codeblock.css"])
			b.WriteByte('\n')
		default:
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return b.String()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./scripts/themegen/ -run TestGenerateTheme`
Expected: PASS.

- [ ] **Step 5: Implement the generator main**

Create `scripts/themegen/main.go`:

```go
// Command themegen generates assets/goshtoso-theme.css — the self-contained
// Goshtoso theme source a consumer imports into their own Tailwind v4 build.
// Run from the repo root: go run ./scripts/themegen
package main

import (
	"fmt"
	"os"
)

func main() {
	mainCSS, err := os.ReadFile("css/main.css")
	must(err)
	allThemes, err := os.ReadFile("all-themes.css")
	must(err)
	codeblock, err := os.ReadFile("css/codeblock.css")
	must(err)

	out := generateTheme(string(mainCSS), map[string]string{
		"all-themes.css": string(allThemes),
		"codeblock.css":  string(codeblock),
	})

	must(os.WriteFile("assets/goshtoso-theme.css", []byte(out), 0644))
	fmt.Printf("themegen: wrote assets/goshtoso-theme.css (%d bytes)\n", len(out))
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "themegen: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 6: Generate the real theme bundle**

Run: `go run ./scripts/themegen`
Expected: `themegen: wrote assets/goshtoso-theme.css (N bytes)` (N in the tens of thousands). This overwrites the Task 1 placeholder.

- [ ] **Step 7: Add the `ThemeCSS()` accessor + test**

In `assets/embed.go`, add after `TailwindVersion`:

```go
// ThemeCSS returns the Goshtoso theme SOURCE (tokens, @custom-variant, the 13
// [data-theme] blocks, base + utility layers) for importing into your OWN
// Tailwind v4 build. Unlike StylesCSS (compiled output you serve directly),
// this is source your Tailwind compiles. Pair it with a @source pointing at
// Goshtoso's components dir (see `goshtoso -source-path`).
func ThemeCSS() ([]byte, error) {
	return files.ReadFile("goshtoso-theme.css")
}
```

Append to `assets/version_test.go`:

```go
func TestThemeCSS(t *testing.T) {
	b, err := ThemeCSS()
	if err != nil {
		t.Fatalf("ThemeCSS() error: %v", err)
	}
	s := string(b)
	for _, want := range []string{"@custom-variant dark", "[data-theme=minimal]", "@theme"} {
		if !strings.Contains(s, want) {
			t.Errorf("ThemeCSS() missing %q", want)
		}
	}
	if strings.Contains(s, `@import "tailwindcss"`) {
		t.Error("ThemeCSS() must not contain the tailwind import")
	}
}
```

- [ ] **Step 8: Run tests**

Run: `go test ./assets/ ./scripts/themegen/`
Expected: PASS (all).

- [ ] **Step 9: Commit**

```bash
git add scripts/themegen assets/embed.go assets/version_test.go assets/goshtoso-theme.css
git commit -m "feat(assets): generate + expose goshtoso-theme.css for consumer tailwind builds"
```

---

### Task 3: CLI `-version`, `-theme`, `-source-path`

**Files:**
- Modify: `cmd/goshtoso/main.go`
- Test: `cmd/goshtoso/main_test.go`

- [ ] **Step 1: Write the failing tests**

Create `cmd/goshtoso/main_test.go`:

```go
package main

import (
	"runtime/debug"
	"testing"
)

func TestVersionString(t *testing.T) {
	info := &debug.BuildInfo{Main: debug.Module{Version: "v0.0.1"}}
	if got := versionString(info, "4.3.0"); got != "goshtoso v0.0.1 (tailwindcss 4.3.0)" {
		t.Fatalf("got %q", got)
	}
	if got := versionString(nil, "4.3.0"); got != "goshtoso (devel) (tailwindcss 4.3.0)" {
		t.Fatalf("nil buildinfo: got %q", got)
	}
}

func TestSourcePath(t *testing.T) {
	if got := sourcePath("/mod/cache/goshtoso@v1"); got != "/mod/cache/goshtoso@v1/components" {
		t.Fatalf("got %q", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/goshtoso/`
Expected: FAIL — `undefined: versionString`, `undefined: sourcePath`.

- [ ] **Step 3: Implement**

Rewrite `cmd/goshtoso/main.go` (preserves existing `-out` CSS extraction; adds modes):

```go
// Command goshtoso extracts embedded Goshtoso assets and reports versions.
//
// Usage:
//
//	goshtoso -out=css/goshtoso-base.css   # extract compiled styles.css
//	goshtoso -theme -out=goshtoso-theme.css  # extract theme SOURCE for your own tailwind build
//	goshtoso -version                     # print goshtoso + tailwind versions
//	goshtoso -source-path                 # print the components dir to @source
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/araihu/goshtoso/assets"
)

func versionString(info *debug.BuildInfo, tailwind string) string {
	v := "(devel)"
	if info != nil && info.Main.Version != "" {
		v = info.Main.Version
	}
	return fmt.Sprintf("goshtoso %s (tailwindcss %s)", v, tailwind)
}

func sourcePath(moduleDir string) string {
	return filepath.Join(moduleDir, "components")
}

// moduleDir resolves the installed goshtoso module directory via `go list`.
func moduleDir() (string, error) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/araihu/goshtoso").Output()
	if err != nil {
		return "", fmt.Errorf("go list (run from a module that requires goshtoso): %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func main() {
	out := flag.String("out", "goshtoso-base.css", "output path for extracted CSS")
	theme := flag.Bool("theme", false, "extract the theme SOURCE (for your own Tailwind build) instead of compiled CSS")
	version := flag.Bool("version", false, "print goshtoso and tailwind versions, then exit")
	srcPath := flag.Bool("source-path", false, "print the components dir to feed Tailwind @source, then exit")
	flag.Parse()

	if *version {
		info, _ := debug.ReadBuildInfo()
		fmt.Println(versionString(info, assets.TailwindVersion()))
		return
	}

	if *srcPath {
		dir, err := moduleDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "goshtoso: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(sourcePath(dir))
		return
	}

	var (
		data []byte
		err  error
	)
	if *theme {
		data, err = assets.ThemeCSS()
	} else {
		data, err = assets.StylesCSS()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "goshtoso: %v\n", err)
		os.Exit(1)
	}

	if dir := filepath.Dir(*out); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "goshtoso: mkdir %s: %v\n", dir, err)
			os.Exit(1)
		}
	}
	if err := os.WriteFile(*out, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "goshtoso: write %s: %v\n", *out, err)
		os.Exit(1)
	}
	fmt.Printf("goshtoso: wrote %s (%d bytes)\n", *out, len(data))
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/goshtoso/`
Expected: PASS.

- [ ] **Step 5: Smoke-test the binary**

Run:
```bash
go run ./cmd/goshtoso -version
go run ./cmd/goshtoso -theme -out=/tmp/gt-theme.css && head -3 /tmp/gt-theme.css
```
Expected: first prints `goshtoso (devel) (tailwindcss 4.3.0)`; second writes the theme file whose first line is the generated header.

- [ ] **Step 6: Commit**

```bash
git add cmd/goshtoso/main.go cmd/goshtoso/main_test.go
git commit -m "feat(cli): add -version, -theme, -source-path to goshtoso command"
```

---

### Task 4: Pinned local build (`just css`) + gitignore + CLAUDE.md

**Files:**
- Modify: `justfile`, `.gitignore`, `CLAUDE.md`

- [ ] **Step 1: Add the `css` recipe to `justfile`**

Append:

```just
# Build assets/styles.css with the PINNED Tailwind (assets/tailwind.version),
# regenerating the embeddable theme source first. Fetches the standalone binary
# on demand into .tools/ (gitignored) — no binary is committed.
css:
    #!/usr/bin/env bash
    set -euo pipefail
    go run ./scripts/themegen
    ver="$(cat assets/tailwind.version)"
    bin=".tools/tailwindcss-${ver}"
    if [ ! -x "$bin" ]; then
      mkdir -p .tools
      os="$(uname -s | tr '[:upper:]' '[:lower:]')"
      arch="$(uname -m)"; case "$arch" in arm64|aarch64) arch=arm64;; x86_64) arch=x64;; esac
      echo "fetching tailwindcss v${ver} (${os}-${arch})..."
      curl -fsSL -o "$bin" \
        "https://github.com/tailwindlabs/tailwindcss/releases/download/v${ver}/tailwindcss-${os}-${arch}"
      chmod +x "$bin"
    fi
    "$bin" -i css/main.css -o assets/styles.css
    echo "css: built assets/styles.css with tailwindcss v${ver}"
```

- [ ] **Step 2: Gitignore the tool cache**

Append to `.gitignore`:

```
# Pinned Tailwind standalone binary cache (fetched by `just css`)
.tools/
```

- [ ] **Step 3: Run the recipe and verify a clean rebuild**

Run:
```bash
just css
git diff --stat -- assets/styles.css assets/goshtoso-theme.css
```
Expected: `just css` succeeds; `git diff` shows **no changes** (the committed artifacts already match a pinned rebuild). If `styles.css` changes, the previously committed CSS was built with a different Tailwind — commit the rebuilt version as part of Step 5.

- [ ] **Step 4: Update CLAUDE.md build command**

In `CLAUDE.md`, replace the Tailwind build command line:

```bash
# Build Tailwind CSS (REQUIRED after editing CSS)
tailwindcss -i css/main.css -o assets/styles.css
```

with:

```bash
# Build Tailwind CSS with the pinned version (REQUIRED after editing CSS)
# Reads assets/tailwind.version; regenerates the theme source; no global tailwind needed.
just css
```

- [ ] **Step 5: Commit**

```bash
git add justfile .gitignore CLAUDE.md assets/styles.css assets/goshtoso-theme.css
git commit -m "build: pin local tailwind via 'just css' (assets/tailwind.version)"
```

---

### Task 5: DRY the CI pin + drift-guard the new artifacts

**Files:**
- Modify: `.github/workflows/ci.yml`

> Context: `ci.yml` has three jobs that each (a) define `TAILWIND_VERSION` via the top-level `env:` and curl the binary, and (b) one job rebuilds `styles.css` and runs `git diff --exit-code`. We make the version come from the pin file and add the new generated artifacts to the rebuild + diff.

- [ ] **Step 1: Make the install step read the pin**

In each "install tailwind" step (3 occurrences), replace the hardcoded `${TAILWIND_VERSION}` curl with a pin read. Change each:

```yaml
      - name: Install Tailwind CSS
        run: |
          curl -fsSL -o /tmp/tailwindcss "https://github.com/tailwindlabs/tailwindcss/releases/download/${TAILWIND_VERSION}/tailwindcss-linux-x64"
          chmod +x /tmp/tailwindcss
          sudo mv /tmp/tailwindcss /usr/local/bin/tailwindcss
```

to:

```yaml
      - name: Install Tailwind CSS (pinned)
        run: |
          ver="$(cat assets/tailwind.version)"
          curl -fsSL -o /tmp/tailwindcss "https://github.com/tailwindlabs/tailwindcss/releases/download/v${ver}/tailwindcss-linux-x64"
          chmod +x /tmp/tailwindcss
          sudo mv /tmp/tailwindcss /usr/local/bin/tailwindcss
```

- [ ] **Step 2: Remove the now-unused env var**

Delete the `TAILWIND_VERSION: 'v4.3.0'` line from the top-level `env:` block (the pin file is now the source of truth). Leave `GO_VERSION`, `TEMPL_VERSION`, `PLAYWRIGHT_VERSION`.

- [ ] **Step 3: Regenerate the theme bundle before the drift check**

In the "Tailwind CSS drift check" step (the `lint-build` job), regenerate the theme source first and extend the diff. Replace:

```yaml
      - name: Tailwind CSS drift check
        run: |
          tailwindcss -i css/main.css -o assets/styles.css
          git diff --exit-code -- assets/styles.css \
            || { echo "::error::tailwind drift — run 'tailwindcss -i css/main.css -o assets/styles.css' and commit"; exit 1; }
```

with:

```yaml
      - name: Tailwind CSS + theme drift check
        run: |
          go run ./scripts/themegen
          tailwindcss -i css/main.css -o assets/styles.css
          git diff --exit-code -- assets/styles.css assets/goshtoso-theme.css \
            || { echo "::error::css/theme drift — run 'just css' and commit assets/styles.css + assets/goshtoso-theme.css"; exit 1; }
```

- [ ] **Step 4: Validate the workflow locally**

Run: `go run ./scripts/themegen && git diff --exit-code -- assets/goshtoso-theme.css && echo OK`
Expected: `OK` (regeneration is deterministic, no diff). If `actionlint` is available, run it: `actionlint .github/workflows/ci.yml`.

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: read tailwind pin from assets/tailwind.version; drift-guard theme bundle"
```

---

### Task 6: Release workflow + VERSIONS.md

**Files:**
- Create: `.github/workflows/release.yml`, `VERSIONS.md`

- [ ] **Step 1: Seed the version map**

Create `VERSIONS.md`:

```markdown
# Version Compatibility

Each released Goshtoso tag and the Tailwind CSS version its CSS was built with.
The source of truth is `assets/tailwind.version`; match your own Tailwind build
to the row for the Goshtoso version you depend on (see `goshtoso -version`).

| Goshtoso | Tailwind CSS |
|----------|--------------|
| v0.0.1   | 4.3.0        |
```

- [ ] **Step 2: Create the release workflow**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.26.x'

      - name: Install templ
        run: go install github.com/a-h/templ/cmd/templ@v0.3.1020

      - name: Install Tailwind CSS (pinned)
        run: |
          ver="$(cat assets/tailwind.version)"
          curl -fsSL -o /tmp/tailwindcss "https://github.com/tailwindlabs/tailwindcss/releases/download/v${ver}/tailwindcss-linux-x64"
          chmod +x /tmp/tailwindcss
          sudo mv /tmp/tailwindcss /usr/local/bin/tailwindcss

      - name: Build assets (pinned, reproducible)
        run: |
          templ generate
          go run ./scripts/themegen
          tailwindcss -i css/main.css -o assets/styles.css

      - name: Verify committed assets match a pinned rebuild
        run: |
          git diff --exit-code -- assets/styles.css assets/goshtoso-theme.css \
            || { echo "::error::tag has stale assets — run 'just css' and re-tag"; exit 1; }

      - name: Release notes
        run: |
          ver="$(cat assets/tailwind.version)"
          {
            echo "Built with Tailwind CSS ${ver}."
            echo
            echo "Assets attached: \`styles.css\` (serve directly) and \`goshtoso-theme.css\` (import into your own Tailwind v4 build)."
          } > release-notes.md

      - name: Create GitHub release
        uses: softprops/action-gh-release@v2
        with:
          body_path: release-notes.md
          files: |
            assets/styles.css
            assets/goshtoso-theme.css
```

- [ ] **Step 3: Validate**

If `actionlint` is available: `actionlint .github/workflows/release.yml`. Otherwise review by eye against `ci.yml` step shapes.

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/release.yml VERSIONS.md
git commit -m "ci(release): tag-driven release shipping styles.css + goshtoso-theme.css"
```

---

### Task 7: Consumer Tailwind docs (Path A / Path B)

**Files:**
- Modify: `docs/USAGE.md`, `.claude/skills/using-goshtoso/SKILL.md`

- [ ] **Step 1: Add the integration section to `docs/USAGE.md`**

After the existing "### 3. Required JavaScript" section, insert:

````markdown
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
@import "./goshtoso-theme.css";                 /* tokens + variants + themes */
@source "/…/goshtoso@vX.Y.Z/components";        /* emit Goshtoso's classes (path from -source-path) */
@theme { --color-brand: oklch(0.7 0.15 250); }  /* your own tokens too */
```

Goshtoso's fonts/images are still served by `assets.Handler()` at `/assets/`,
so mount it regardless of which path you choose.
````

- [ ] **Step 2: Add a condensed pointer to the skill**

In `.claude/skills/using-goshtoso/SKILL.md`, under the "Theming" section, append:

```markdown
### Running your own Tailwind

Two paths (full detail in `docs/USAGE.md` → "Using your own Tailwind build"):

- **Path A (recommended):** serve our prebuilt `styles.css` via `assets.Handler()`
  and run your own Tailwind into a *separate* file. No coupling.
- **Path B (unified):** `goshtoso -theme -out=…` extracts the theme source to
  `@import`, and `goshtoso -source-path` prints the components dir to `@source`.
  Your Tailwind must match `goshtoso -version` (also in `VERSIONS.md`).
```

- [ ] **Step 3: Verify the skill reference is not stale**

Run: `go run ./scripts/skillgen && git diff --exit-code -- .claude/skills/using-goshtoso/components-reference.md && echo OK`
Expected: `OK` (this change does not touch component `types.go`, so the generated reference is unchanged).

- [ ] **Step 4: Commit**

```bash
git add docs/USAGE.md .claude/skills/using-goshtoso/SKILL.md
git commit -m "docs: document Path A / Path B consumer Tailwind integration"
```

---

## Final verification

- [ ] **Build + test the whole module**

Run: `go build ./... && go test ./assets/ ./scripts/themegen/ ./cmd/goshtoso/`
Expected: build clean; all tests PASS.

- [ ] **Lint**

Run: `golangci-lint run`
Expected: `0 issues.`

- [ ] **Reproducible-build guard**

Run: `just css && git diff --exit-code -- assets/styles.css assets/goshtoso-theme.css && echo CLEAN`
Expected: `CLEAN`.

- [ ] **End-to-end consumer smoke (Path B)**

In a throwaway module that requires this branch (via `replace`):
```bash
go run github.com/araihu/goshtoso/cmd/goshtoso -version       # prints versions
go run github.com/araihu/goshtoso/cmd/goshtoso -source-path   # prints components dir
go run github.com/araihu/goshtoso/cmd/goshtoso -theme -out=/tmp/t.css && grep -q '@custom-variant' /tmp/t.css && echo THEME_OK
```
Expected: versions print; source-path prints a real dir; `THEME_OK`.

---

## Notes for the implementer

- **Run everything from the repo root** (`css/main.css`, `all-themes.css` paths are root-relative).
- **`assets/goshtoso-theme.css` and `assets/styles.css` are generated** — never hand-edit; regenerate via `just css` (which runs `themegen` then the pinned Tailwind).
- **The pin `assets/tailwind.version` is the only place the version is written.** If you bump Tailwind, change that file, run `just css`, commit the rebuilt assets, and add a `VERSIONS.md` row.
- **`@font-face` url() paths** in `css/main.css` carry through into `goshtoso-theme.css` verbatim; they resolve against the consumer's `/assets/fonts/` served by `assets.Handler()`. If a consumer build warns about missing fonts, that mount is the fix — document only if it surfaces.
- Follow existing repo conventions: generators live in `scripts/<name>/` (mirror `scripts/skillgen`); CI drift checks use `git diff --exit-code`.
```

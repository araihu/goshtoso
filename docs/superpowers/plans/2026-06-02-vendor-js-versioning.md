# Versioned Vendor JS + Stated Provenance — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Serve every vendored third-party JS dep at a versioned path
`/assets/js/vendor/<module>/<version>/<file>` driven by a single
`versions.json` manifest, with the versions exposed through Go accessors —
mirroring the Tailwind `tailwind.version` provenance pattern.

**Architecture:** `assets/js/vendor/versions.json` is the source of truth. A
`scripts/vendorgen` Go tool reads it, verifies the on-disk files exist, and
generates `assets/vendor_gen.go` (URL-path constants). `head.templ` references
those constants; `embed.go` exposes version accessors that read the manifest.
A `-download` mode + `just vendor-js` recipe fetches pinned versions; a `-check`
mode + CI step enforce no drift. Hard cut: old flat paths are removed.

**Tech Stack:** Go 1.26, templ v0.3.1020, `embed.FS`, `just`, GitHub Actions.

**Design doc:** `docs/superpowers/specs/2026-06-02-vendor-js-versioning-design.md`

**Reference versions:** alpinejs 3.14.9 (core+collapse+focus), htmx.org 2.0.8,
htmx-ext-sse 2.2.3, htmx-ext-ws 2.0.3.

---

### Task 1: Create the version manifest

**Files:**
- Create: `assets/js/vendor/versions.json`

- [ ] **Step 1: Write the manifest**

Create `assets/js/vendor/versions.json` (two-space indent, trailing newline):

```json
{
  "alpinejs":          { "version": "3.14.9", "file": "alpine.min.js",          "url": "https://unpkg.com/alpinejs@{v}/dist/cdn.min.js" },
  "alpinejs-collapse": { "version": "3.14.9", "file": "alpine-collapse.min.js", "url": "https://unpkg.com/@alpinejs/collapse@{v}/dist/cdn.min.js" },
  "alpinejs-focus":    { "version": "3.14.9", "file": "alpine-focus.min.js",    "url": "https://unpkg.com/@alpinejs/focus@{v}/dist/cdn.min.js" },
  "htmx.org":          { "version": "2.0.8",  "file": "htmx.min.js",            "url": "https://unpkg.com/htmx.org@{v}/dist/htmx.min.js" },
  "htmx-ext-sse":      { "version": "2.2.3",  "file": "htmx-ext-sse.min.js",    "url": "https://unpkg.com/htmx-ext-sse@{v}/dist/sse.min.js" },
  "htmx-ext-ws":       { "version": "2.0.3",  "file": "htmx-ext-ws.js",         "url": "https://unpkg.com/htmx-ext-ws@{v}/ws.js" }
}
```

- [ ] **Step 2: Validate it parses**

Run: `python3 -c "import json;json.load(open('assets/js/vendor/versions.json'))" && echo OK`
Expected: `OK`

- [ ] **Step 3: Commit**

```bash
git add assets/js/vendor/versions.json
git commit -m "feat(assets): add vendor JS version manifest (SSOT)"
```

---

### Task 2: Migrate vendored files into versioned directories

**Files:**
- Move: `assets/js/vendor/{alpine.min.js,alpine-collapse.min.js,alpine-focus.min.js,htmx.min.js,htmx-ext-sse.min.js,htmx-ext-ws.js}`
  → `assets/js/vendor/<module>/<version>/<file>`

Reuse the existing vetted bytes (do NOT re-download — avoids any minification
diff). `//go:embed js` recurses, so no embed directive change is needed.

- [ ] **Step 1: git mv each file into its versioned dir**

```bash
cd assets/js/vendor
mkdir -p alpinejs/3.14.9 alpinejs-collapse/3.14.9 alpinejs-focus/3.14.9 \
         htmx.org/2.0.8 htmx-ext-sse/2.2.3 htmx-ext-ws/2.0.3
git mv alpine.min.js          alpinejs/3.14.9/alpine.min.js
git mv alpine-collapse.min.js alpinejs-collapse/3.14.9/alpine-collapse.min.js
git mv alpine-focus.min.js    alpinejs-focus/3.14.9/alpine-focus.min.js
git mv htmx.min.js            htmx.org/2.0.8/htmx.min.js
git mv htmx-ext-sse.min.js    htmx-ext-sse/2.2.3/htmx-ext-sse.min.js
git mv htmx-ext-ws.js         htmx-ext-ws/2.0.3/htmx-ext-ws.js
cd -
```

- [ ] **Step 2: Verify the new tree (no flat .js left under vendor/)**

Run: `find assets/js/vendor -name '*.js' | sort`
Expected (exactly these six):
```
assets/js/vendor/alpinejs-collapse/3.14.9/alpine-collapse.min.js
assets/js/vendor/alpinejs-focus/3.14.9/alpine-focus.min.js
assets/js/vendor/alpinejs/3.14.9/alpine.min.js
assets/js/vendor/htmx-ext-sse/2.2.3/htmx-ext-sse.min.js
assets/js/vendor/htmx-ext-ws/2.0.3/htmx-ext-ws.js
assets/js/vendor/htmx.org/2.0.8/htmx.min.js
```

- [ ] **Step 3: Verify Alpine version sanity in the moved file**

Run: `grep -o 'version:"3.14.9"' assets/js/vendor/alpinejs/3.14.9/alpine.min.js | head -1`
Expected: `version:"3.14.9"`

- [ ] **Step 4: Commit**

```bash
git add -A assets/js/vendor
git commit -m "refactor(assets): move vendored JS into versioned dirs"
```

---

### Task 3: Write the vendorgen generator + generated constants

**Files:**
- Create: `scripts/vendorgen/main.go`
- Create (generated): `assets/vendor_gen.go`
- Test: `scripts/vendorgen/main_test.go`

The generator has three modes: default (generate + verify presence),
`-check` (drift check for CI), `-download` (fetch from manifest — added in
Task 8). This task implements default + `-check` + the test.

- [ ] **Step 1: Write the failing test**

Create `scripts/vendorgen/main_test.go`:

```go
package main

import "testing"

func TestURLPath(t *testing.T) {
	got := urlPath("alpinejs", dep{Version: "3.14.9", File: "alpine.min.js"})
	want := "/assets/js/vendor/alpinejs/3.14.9/alpine.min.js"
	if got != want {
		t.Fatalf("urlPath = %q, want %q", got, want)
	}
}

func TestConstNameMapComplete(t *testing.T) {
	// Every module in the canonical list must have a Go constant name.
	for _, k := range []string{
		"alpinejs", "alpinejs-collapse", "alpinejs-focus",
		"htmx.org", "htmx-ext-sse", "htmx-ext-ws",
	} {
		if constName[k] == "" {
			t.Errorf("missing constName for %q", k)
		}
	}
}

func TestRenderDeterministic(t *testing.T) {
	deps := map[string]dep{
		"htmx.org": {Version: "2.0.8", File: "htmx.min.js"},
		"alpinejs": {Version: "3.14.9", File: "alpine.min.js"},
	}
	a := render(deps)
	b := render(deps)
	if a != b {
		t.Fatal("render not deterministic")
	}
	// Constants emitted in sorted-by-Go-name order: AlpineJSURL before HTMXURL.
	if ai, hi := indexOf(a, "AlpineJSURL"), indexOf(a, "HTMXURL"); ai < 0 || hi < 0 || ai > hi {
		t.Fatalf("ordering wrong: AlpineJSURL@%d HTMXURL@%d", ai, hi)
	}
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./scripts/vendorgen/ 2>&1 | tail -5`
Expected: FAIL — `undefined: urlPath`, `undefined: dep`, `undefined: constName`, `undefined: render`.

- [ ] **Step 3: Write the generator**

Create `scripts/vendorgen/main.go`:

```go
// Command vendorgen generates assets/vendor_gen.go from the vendor version
// manifest (assets/js/vendor/versions.json) — the single source of truth for
// every vendored third-party JS dependency's version, file, and origin.
//
// Modes:
//
//	go run ./scripts/vendorgen            // generate vendor_gen.go + verify files exist
//	go run ./scripts/vendorgen -check     // CI: fail if regeneration would change the file
//	go run ./scripts/vendorgen -download  // fetch pinned versions from the manifest, then generate
//
// Run from the repo root. CI fails if the committed vendor_gen.go is stale.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
)

const (
	manifestPath = "assets/js/vendor/versions.json"
	outPath      = "assets/vendor_gen.go"
	vendorRoot   = "assets/js/vendor"
)

// dep is one entry in versions.json.
type dep struct {
	Version string `json:"version"`
	File    string `json:"file"`
	URL     string `json:"url"`
}

// constName maps a manifest module key to its exported Go constant name.
// Adding a new dep requires adding it here — intentional friction.
var constName = map[string]string{
	"alpinejs":          "AlpineJSURL",
	"alpinejs-collapse": "AlpineCollapseURL",
	"alpinejs-focus":    "AlpineFocusURL",
	"htmx.org":          "HTMXURL",
	"htmx-ext-sse":      "HTMXExtSSEURL",
	"htmx-ext-ws":       "HTMXExtWSURL",
}

func urlPath(module string, d dep) string {
	return "/assets/js/vendor/" + module + "/" + d.Version + "/" + d.File
}

func diskPath(module string, d dep) string {
	return filepath.Join(vendorRoot, module, d.Version, d.File)
}

func loadManifest() (map[string]dep, error) {
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var m map[string]dep
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", manifestPath, err)
	}
	return m, nil
}

// render produces the formatted vendor_gen.go source for the given manifest.
func render(deps map[string]dep) string {
	type kv struct {
		name, url string
	}
	var rows []kv
	for module, d := range deps {
		name := constName[module]
		if name == "" {
			panic("vendorgen: no constName for module " + module)
		}
		rows = append(rows, kv{name, urlPath(module, d)})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].name < rows[j].name })

	var buf bytes.Buffer
	buf.WriteString("// Code generated by vendorgen. DO NOT EDIT.\n")
	buf.WriteString("// Source of truth: assets/js/vendor/versions.json\n\n")
	buf.WriteString("package assets\n\n")
	buf.WriteString("const (\n")
	for _, r := range rows {
		fmt.Fprintf(&buf, "\t%s = %q\n", r.name, r.url)
	}
	buf.WriteString(")\n")

	out, err := format.Source(buf.Bytes())
	if err != nil {
		panic(fmt.Sprintf("vendorgen: gofmt failed: %v\n%s", err, buf.String()))
	}
	return string(out)
}

func verifyFilesExist(deps map[string]dep) error {
	for module, d := range deps {
		p := diskPath(module, d)
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("vendored file missing: %s (run `just vendor-js`): %w", p, err)
		}
	}
	return nil
}

func main() {
	check := flag.Bool("check", false, "fail if vendor_gen.go would change")
	download := flag.Bool("download", false, "download pinned versions from the manifest")
	flag.Parse()

	deps, err := loadManifest()
	if err != nil {
		fmt.Fprintln(os.Stderr, "vendorgen:", err)
		os.Exit(1)
	}

	if *download {
		if err := downloadAll(deps); err != nil { // defined in download.go (Task 8)
			fmt.Fprintln(os.Stderr, "vendorgen download:", err)
			os.Exit(1)
		}
	}

	if err := verifyFilesExist(deps); err != nil {
		fmt.Fprintln(os.Stderr, "vendorgen:", err)
		os.Exit(1)
	}

	gen := render(deps)

	if *check {
		existing, err := os.ReadFile(outPath)
		if err != nil || string(existing) != gen {
			fmt.Fprintf(os.Stderr, "::error::%s is stale — run `go run ./scripts/vendorgen` and commit\n", outPath)
			os.Exit(1)
		}
		return
	}

	if err := os.WriteFile(outPath, []byte(gen), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "vendorgen:", err)
		os.Exit(1)
	}
	fmt.Printf("vendorgen: wrote %s (%d deps)\n", outPath, len(deps))
}
```

Note: `downloadAll` is referenced but defined in Task 8 (`download.go`). To keep
this task compiling on its own, add a temporary stub at the bottom of `main.go`
now; Task 8 moves it to `download.go`:

```go
// Temporary stub — replaced by download.go in Task 8.
func downloadAll(map[string]dep) error { return fmt.Errorf("-download not yet implemented") }
```

- [ ] **Step 4: Run the generator test**

Run: `go test ./scripts/vendorgen/ -v 2>&1 | tail -12`
Expected: PASS (TestURLPath, TestConstNameMapComplete, TestRenderDeterministic).

- [ ] **Step 5: Generate vendor_gen.go**

Run: `go run ./scripts/vendorgen && cat assets/vendor_gen.go`
Expected: file written; contains the six constants in sorted-by-name order, e.g.
`AlpineCollapseURL = "/assets/js/vendor/alpinejs-collapse/3.14.9/alpine-collapse.min.js"`.

- [ ] **Step 6: Verify it builds**

Run: `go build ./assets/ && echo OK`
Expected: `OK`

- [ ] **Step 7: Commit**

```bash
git add scripts/vendorgen assets/vendor_gen.go
git commit -m "feat(assets): vendorgen generates versioned vendor URL constants"
```

---

### Task 4: Add version accessors to embed.go

**Files:**
- Modify: `assets/embed.go` (add manifest reader + accessors)
- Test: `assets/version_test.go` (extend)

- [ ] **Step 1: Write the failing test**

Append to `assets/version_test.go`:

```go
func TestVendorVersions(t *testing.T) {
	cases := map[string]string{
		"Alpine":     AlpineVersion(),
		"HTMX":       HTMXVersion(),
		"HTMXExtSSE": HTMXExtSSEVersion(),
		"HTMXExtWS":  HTMXExtWSVersion(),
	}
	want := map[string]string{
		"Alpine": "3.14.9", "HTMX": "2.0.8", "HTMXExtSSE": "2.2.3", "HTMXExtWS": "2.0.3",
	}
	for k, got := range cases {
		if got != want[k] {
			t.Errorf("%sVersion() = %q, want %q", k, got, want[k])
		}
	}
}

func TestVendorFilesEmbedded(t *testing.T) {
	// Every file declared in the manifest must be present in the embedded FS
	// at its versioned path (the path the generated constants point at).
	for _, p := range []string{
		"js/vendor/alpinejs/3.14.9/alpine.min.js",
		"js/vendor/alpinejs-collapse/3.14.9/alpine-collapse.min.js",
		"js/vendor/alpinejs-focus/3.14.9/alpine-focus.min.js",
		"js/vendor/htmx.org/2.0.8/htmx.min.js",
		"js/vendor/htmx-ext-sse/2.2.3/htmx-ext-sse.min.js",
		"js/vendor/htmx-ext-ws/2.0.3/htmx-ext-ws.js",
	} {
		if _, err := files.ReadFile(p); err != nil {
			t.Errorf("embedded file missing: %s: %v", p, err)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./assets/ -run TestVendor 2>&1 | tail -5`
Expected: FAIL — `undefined: AlpineVersion` (and siblings).

- [ ] **Step 3: Add the accessors to embed.go**

Add `"encoding/json"` to the import block, then append to `assets/embed.go`:

```go
// vendorDep mirrors one entry in js/vendor/versions.json.
type vendorDep struct {
	Version string `json:"version"`
	File    string `json:"file"`
	URL     string `json:"url"`
}

// vendorVersion returns the pinned version of a vendored module from the
// embedded manifest (js/vendor/versions.json), or "" if absent.
func vendorVersion(module string) string {
	b, err := files.ReadFile("js/vendor/versions.json")
	if err != nil {
		return ""
	}
	var m map[string]vendorDep
	if json.Unmarshal(b, &m) != nil {
		return ""
	}
	return m[module].Version
}

// AlpineVersion returns the Alpine.js version Goshtoso vendors (core, collapse,
// and focus share this version). Pinned in js/vendor/versions.json.
func AlpineVersion() string { return vendorVersion("alpinejs") }

// HTMXVersion returns the vendored HTMX version (js/vendor/versions.json).
func HTMXVersion() string { return vendorVersion("htmx.org") }

// HTMXExtSSEVersion returns the vendored htmx-ext-sse version.
func HTMXExtSSEVersion() string { return vendorVersion("htmx-ext-sse") }

// HTMXExtWSVersion returns the vendored htmx-ext-ws version.
func HTMXExtWSVersion() string { return vendorVersion("htmx-ext-ws") }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./assets/ -run TestVendor -v 2>&1 | tail -8`
Expected: PASS (TestVendorVersions, TestVendorFilesEmbedded).

- [ ] **Step 5: Commit**

```bash
git add assets/embed.go assets/version_test.go
git commit -m "feat(assets): expose vendored JS version accessors"
```

---

### Task 5: Switch head.templ to the generated constants

**Files:**
- Modify: `components/head/head.templ`
- Regenerate: `components/head/head_templ.go` (via `templ generate`)
- Test: `components/head/head_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `components/head/head_test.go` (`&strings.Builder` is an `io.Writer`, which
is exactly templ's `Component.Render(context.Context, io.Writer) error` signature):

```go
package head

import (
	"context"
	"io"
	"strings"
	"testing"
)

func render(t *testing.T, c interface {
	Render(context.Context, io.Writer) error
}) string {
	t.Helper()
	var sb strings.Builder
	if err := c.Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

func TestDependenciesUsesVersionedPaths(t *testing.T) {
	out := render(t, Dependencies())
	for _, want := range []string{
		"/assets/js/vendor/alpinejs/3.14.9/alpine.min.js",
		"/assets/js/vendor/alpinejs-collapse/3.14.9/alpine-collapse.min.js",
		"/assets/js/vendor/alpinejs-focus/3.14.9/alpine-focus.min.js",
		"/assets/js/vendor/htmx.org/2.0.8/htmx.min.js",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Dependencies() missing versioned path %q\n%s", want, out)
		}
	}
	// Hard cut: no flat vendored paths.
	for _, bad := range []string{`vendor/alpine.min.js`, `vendor/htmx.min.js`} {
		if strings.Contains(out, bad) {
			t.Errorf("Dependencies() still emits flat path %q", bad)
		}
	}
}

func TestDependenciesMinimalUsesVersionedPaths(t *testing.T) {
	out := render(t, DependenciesMinimal())
	for _, want := range []string{
		"/assets/js/vendor/alpinejs/3.14.9/alpine.min.js",
		"/assets/js/vendor/htmx.org/2.0.8/htmx.min.js",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("DependenciesMinimal() missing %q\n%s", want, out)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./components/head/ 2>&1 | tail -6`
Expected: FAIL — current `head.templ` still emits flat `vendor/alpine.min.js`, so
the versioned-path assertions fail (the test compiles; `Dependencies()` exists).

- [ ] **Step 3: Update head.templ to reference the constants**

Replace the two `templ` blocks in `components/head/head.templ` (keep the doc
comments above them). Add the import.

```go
package head

import "github.com/araihu/goshtoso/assets"
```

`Dependencies()`:

```go
templ Dependencies() {
	<link rel="stylesheet" href="/assets/styles.css"/>
	<!-- Alpine.js Collapse Plugin (must load before Alpine core) -->
	<script defer src={ assets.AlpineCollapseURL }></script>
	<!-- Alpine.js Focus Plugin -->
	<script defer src={ assets.AlpineFocusURL }></script>
	<!-- Alpine.js Core -->
	<script defer src={ assets.AlpineJSURL }></script>
	<!-- HTMX -->
	<script src={ assets.HTMXURL }></script>
	<!-- Combobox keyboard nav -->
	<script defer src="/assets/js/combobox.js"></script>
}
```

`DependenciesMinimal()`:

```go
templ DependenciesMinimal() {
	<link rel="stylesheet" href="/assets/styles.css"/>
	<script defer src={ assets.AlpineJSURL }></script>
	<script src={ assets.HTMXURL }></script>
	<!-- Combobox keyboard nav -->
	<script defer src="/assets/js/combobox.js"></script>
}
```

- [ ] **Step 4: Regenerate templ + run the test (now passes)**

Run:
```bash
templ generate
go test ./components/head/ -v 2>&1 | tail -12
```
Expected: PASS. If templ reports "0 updates", force:
`rm components/head/head_templ.go && templ generate`.

- [ ] **Step 5: Verify whole module builds**

Run: `go build ./... && echo OK`
Expected: `OK`

- [ ] **Step 6: Commit**

```bash
git add components/head/head.templ components/head/head_templ.go components/head/head_test.go
git commit -m "feat(head): emit versioned vendor JS paths via generated constants"
```

---

### Task 6: Update the demo site hand-rolled heads

> **NOTE (repo split #21):** the demo site lives in the separate `site/` module
> (`site/go.mod`). It imports the root library's `assets` package, so the
> generated `assets.*URL` constants are available cross-module. Site Go commands
> must run from `site/` (`cd site`); `templ generate` from the repo root is
> file-based and regenerates both modules.

**Files:**
- Modify: `site/internal/pages/demo/layout.templ` (the vendored `<script>` lines)
- Modify: `site/internal/pages/demo/components/landing.templ` (the alpine + htmx lines)
- Regenerate: the corresponding `*_templ.go`

These hand-roll their heads (layout adds htmx-ext-ws + htmx-ext-sse; landing is
minimal). Point them at the generated constants.

- [ ] **Step 1: Update layout.templ**

In `site/internal/pages/demo/layout.templ`, ensure `assets` is imported
(`"github.com/araihu/goshtoso/assets"`), then replace the six vendored `<script>`
lines (currently `/assets/js/vendor/alpine-focus.min.js` … `htmx-ext-sse.min.js`)
with:

```go
			<script defer src={ assets.AlpineFocusURL }></script>
			<script defer src={ assets.AlpineCollapseURL }></script>
			<script defer src={ assets.AlpineJSURL }></script>
			<script src={ assets.HTMXURL }></script>
			<script src={ assets.HTMXExtWSURL }></script>
			<script src={ assets.HTMXExtSSEURL }></script>
```

- [ ] **Step 2: Update landing.templ**

In `site/internal/pages/demo/components/landing.templ`, ensure `assets` is
imported, then replace the two vendored lines (`alpine.min.js` + `htmx.min.js`):

```go
			<script defer src={ assets.AlpineJSURL }></script>
			<script src={ assets.HTMXURL }></script>
```

- [ ] **Step 3: Regenerate + build (both modules)**

Run:
```bash
templ generate
go build ./... && (cd site && go build ./...) && echo OK
```
Expected: `OK`. If templ reports "0 updates" for the site files, force-regen the
two: `rm site/internal/pages/demo/layout_templ.go site/internal/pages/demo/components/landing_templ.go && templ generate`.

- [ ] **Step 4: Confirm no flat vendor paths remain in any source**

Run: `grep -rn 'vendor/alpine\.min\|vendor/htmx\.min\|vendor/alpine-\|vendor/htmx-ext' --include='*.templ' .`
Expected: no matches (all moved to `assets.*URL`).

- [ ] **Step 5: Commit**

```bash
git add site/internal/pages/demo/layout.templ site/internal/pages/demo/layout_templ.go \
        site/internal/pages/demo/components/landing.templ site/internal/pages/demo/components/landing_templ.go
git commit -m "refactor(demo): use versioned vendor JS constants in hand-rolled heads"
```

---

### Task 7: Update docs (USAGE.md, embed.go comment, CLAUDE.md)

**Files:**
- Modify: `docs/USAGE.md:93-97`
- Modify: `assets/embed.go` (doc comment, lines ~11-15)
- Modify: `CLAUDE.md:24-26`

- [ ] **Step 1: USAGE.md — state versions + versioned paths**

In `docs/USAGE.md`, replace the explicit `<script>` block (lines 93-97) with the
versioned paths and add a provenance sentence. New block:

```html
<!-- Pinned versions live in assets/js/vendor/versions.json (see assets.AlpineVersion(), assets.HTMXVersion()). -->
<script defer src="/assets/js/vendor/alpinejs-collapse/3.14.9/alpine-collapse.min.js"></script>
<script defer src="/assets/js/vendor/alpinejs-focus/3.14.9/alpine-focus.min.js"></script>
<script defer src="/assets/js/vendor/alpinejs/3.14.9/alpine.min.js"></script>
<script src="/assets/js/vendor/htmx.org/2.0.8/htmx.min.js"></script>
```

Add immediately after the block: a sentence reading
"These versioned paths change when you upgrade a dep; prefer `@head.Dependencies()`
so you never hardcode them."

- [ ] **Step 2: embed.go doc comment**

In `assets/embed.go`, update the bullet list (lines ~11-15) to the versioned
paths, e.g.:

```go
//   - /assets/js/vendor/alpinejs/3.14.9/alpine.min.js — Alpine.js
//   - /assets/js/vendor/htmx.org/2.0.8/htmx.min.js — HTMX
//   - /assets/js/vendor/htmx-ext-sse/2.2.3/htmx-ext-sse.min.js — HTMX SSE extension
//   - /assets/js/vendor/alpinejs-collapse/3.14.9/alpine-collapse.min.js — Alpine collapse plugin
//   - /assets/js/vendor/alpinejs-focus/3.14.9/alpine-focus.min.js — Alpine focus plugin
//   - versions pinned in js/vendor/versions.json (see AlpineVersion()/HTMXVersion())
```

- [ ] **Step 3: CLAUDE.md note**

In `CLAUDE.md` (lines ~24-26), change the "bundled locally under
`assets/js/vendor/`" sentence to note the versioned layout:
"…bundled locally under `assets/js/vendor/<module>/<version>/` (pinned in
`assets/js/vendor/versions.json`; no CDN at runtime)."

- [ ] **Step 4: Verify no stale flat paths in docs**

Run: `grep -rn 'js/vendor/alpine\.min\|js/vendor/htmx\.min' docs/USAGE.md assets/embed.go CLAUDE.md`
Expected: no matches.

- [ ] **Step 5: Commit**

```bash
git add docs/USAGE.md assets/embed.go CLAUDE.md
git commit -m "docs: state vendored JS versions + versioned asset paths"
```

---

### Task 8: Vendoring automation (-download mode + just vendor-js)

**Files:**
- Create: `scripts/vendorgen/download.go`
- Modify: `scripts/vendorgen/main.go` (remove the temporary `downloadAll` stub from Task 3)
- Modify: `justfile` (add `vendor-js` recipe)

- [ ] **Step 1: Remove the stub from main.go**

Delete the temporary stub added in Task 3:

```go
// Temporary stub — replaced by download.go in Task 8.
func downloadAll(map[string]dep) error { return fmt.Errorf("-download not yet implemented") }
```

- [ ] **Step 2: Write download.go**

Create `scripts/vendorgen/download.go`:

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// downloadAll fetches every dep in the manifest into its versioned dir,
// verifies the bytes, and prunes stale version dirs. Run via `just vendor-js`.
func downloadAll(deps map[string]dep) error {
	for module, d := range deps {
		url := strings.ReplaceAll(d.URL, "{v}", d.Version)
		body, err := fetch(url)
		if err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
		if err := verifyBytes(module, d, body); err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
		dst := diskPath(module, d)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, body, 0o644); err != nil {
			return err
		}
		fmt.Printf("vendorgen: fetched %s@%s -> %s\n", module, d.Version, dst)
		if err := pruneStale(module, d.Version); err != nil {
			return err
		}
	}
	return nil
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec // url is from the committed manifest
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// verifyBytes guards against a wrong/renamed upstream artifact. Files that
// embed a version string (Alpine, HTMX) must contain the manifest version;
// the htmx extensions instead must contain a known marker token.
func verifyBytes(module string, d dep, body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("empty download")
	}
	s := string(body)
	switch module {
	case "alpinejs", "alpinejs-collapse", "alpinejs-focus", "htmx.org":
		if !strings.Contains(s, `version:"`+d.Version+`"`) {
			return fmt.Errorf("version %q not found in downloaded bytes", d.Version)
		}
	case "htmx-ext-sse":
		// htmx extensions register via htmx.defineExtension; sse also fires htmx:sse* events.
		if !strings.Contains(s, "defineExtension") || !strings.Contains(s, "sse") {
			return fmt.Errorf("htmx-ext-sse markers not found")
		}
	case "htmx-ext-ws":
		if !strings.Contains(s, "defineExtension") || !strings.Contains(s, "ws") {
			return fmt.Errorf("htmx-ext-ws markers not found")
		}
	}
	return nil
}

// pruneStale removes assets/js/vendor/<module>/<otherVersion> dirs that are not
// the pinned version, so only the current version ships.
func pruneStale(module, keep string) error {
	dir := filepath.Join(vendorRoot, module)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() != keep {
			if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
				return err
			}
			fmt.Printf("vendorgen: pruned stale %s/%s\n", module, e.Name())
		}
	}
	return nil
}
```

- [ ] **Step 3: Add the just recipe**

In `justfile`, add after the `css:` recipe:

```makefile
# Download the PINNED vendored JS (assets/js/vendor/versions.json) into
# versioned dirs and regenerate the URL constants. Mirrors `just css`.
vendor-js:
    go run ./scripts/vendorgen -download
```

- [ ] **Step 4: Verify the tool still builds + tests pass**

Run:
```bash
go build ./scripts/vendorgen/ && go test ./scripts/vendorgen/ 2>&1 | tail -3
```
Expected: build OK; tests PASS.

- [ ] **Step 5: Dry-run the recipe (network) and confirm idempotence**

Run: `just vendor-js && git status --short assets/js/vendor`
Expected: re-downloads the pinned versions; `git status` shows **no changes**
(same bytes already committed) OR only expected updates. If unexpected diffs
appear, inspect — upstream artifact changed.

- [ ] **Step 6: Commit**

```bash
git add scripts/vendorgen/download.go scripts/vendorgen/main.go justfile
git commit -m "feat(assets): just vendor-js downloads + verifies pinned vendor JS"
```

---

### Task 9: CI staleness gate

**Files:**
- Modify: `.github/workflows/ci.yml` (add a vendorgen drift step after the skillgen step, ~line 64)

- [ ] **Step 1: Add the drift-check step**

In `.github/workflows/ci.yml`, after the "using-goshtoso skill reference drift
check" step, add:

```yaml
      - name: vendor JS constants drift check
        run: |
          go run ./scripts/vendorgen -check \
            || { echo "::error::assets/vendor_gen.go is stale — run 'go run ./scripts/vendorgen' and commit"; exit 1; }
```

- [ ] **Step 2: Validate the workflow YAML parses**

Run: `python3 -c "import yaml;yaml.safe_load(open('.github/workflows/ci.yml'))" && echo OK`
Expected: `OK`

- [ ] **Step 3: Locally simulate the CI check (must pass on clean tree)**

Run: `go run ./scripts/vendorgen -check && echo CLEAN`
Expected: `CLEAN`

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: fail on stale vendor_gen.go (vendorgen -check)"
```

---

### Task 10: Full verification (hard-cut behavioral proof)

**Files:** none (verification only)

> **NOTE (repo split #21):** library commands (`assets`, `components/head`,
> `scripts/vendorgen`) run from the repo root; site commands (server build, e2e)
> run from `site/` (`cd site`).

- [ ] **Step 1: Lint + full build (both modules)**

Run: `golangci-lint run && (cd site && "$(go env GOPATH)/bin/golangci-lint" run --timeout 5m) && go build ./... && (cd site && go build ./...) && echo OK`
Expected: `0 issues.` (×2) then `OK`. (Generated `vendor_gen.go` is excluded by the
generated-header rule; confirm no lint error references it.)

- [ ] **Step 2: Unit tests (library)**

Run: `go test ./assets/ ./components/head/ ./scripts/vendorgen/ 2>&1 | tail -8`
Expected: all PASS. (These three packages are all in the root library module.)

- [ ] **Step 3: Run the demo server, assert versioned 200 / flat 404**

```bash
(cd site && go build -o /tmp/gs ./cmd/server) && /tmp/gs -port 8099 &
sleep 1.3
curl -s -o /dev/null -w "versioned alpine: %{http_code}\n" http://localhost:8099/assets/js/vendor/alpinejs/3.14.9/alpine.min.js
curl -s -o /dev/null -w "versioned htmx:   %{http_code}\n" http://localhost:8099/assets/js/vendor/htmx.org/2.0.8/htmx.min.js
curl -s -o /dev/null -w "flat alpine(404): %{http_code}\n" http://localhost:8099/assets/js/vendor/alpine.min.js
pkill -f '/tmp/gs'
```
Expected: versioned → `200`, `200`; flat → `404`.

- [ ] **Step 4: Demo page references resolve (no flat paths in served HTML)**

```bash
/tmp/gs -port 8099 & sleep 1.3
curl -s http://localhost:8099/components/accordion | grep -oE '/assets/js/vendor/[^"]+' | sort -u
pkill -f '/tmp/gs'
```
Expected: only versioned `…/<module>/<version>/…` paths; no flat `vendor/*.min.js`.

- [ ] **Step 5: getting-started example still works on the new paths**

```bash
cd examples/getting-started && go build . && ./getting-started & sleep 1.3
curl -s http://localhost:3000/ | grep -oE '/assets/js/vendor/[^"]+' | sort -u
curl -s -o /dev/null -w "styles:%{http_code}\n" http://localhost:3000/assets/styles.css
pkill -f getting-started; cd -
```
Expected: served head shows versioned paths (it uses `@head.Dependencies()`);
styles 200.

- [ ] **Step 6: E2E smoke (full suite is ~2.5min; run a representative subset)**

Run: `cd site && go test ./tests/e2e/... -count=1 -timeout 5m -run 'TestSidebar|TestDropdown|TestTableHTMX' 2>&1 | tail -15`
Expected: PASS (these load the demo layout's vendored JS; a broken path would
404 and fail the no-console-error / interaction assertions).

- [ ] **Step 7: Final full E2E (gate before finishing the branch)**

Run: `cd site && go test ./tests/e2e/... -count=1 -timeout 15m 2>&1 | tail -15`
Expected: all PASS, no skips.

---

## Notes for the executor

- Tasks 1→9 are mostly sequential (each builds on the prior). Task 3's stub keeps
  the generator compiling before Task 8 adds the real `downloadAll`.
- Do NOT hand-edit `assets/vendor_gen.go` — regenerate via `go run ./scripts/vendorgen`.
- Resolve `.templ` then `templ generate`; never hand-edit `*_templ.go`.
- If `templ generate` reports "0 updates" after a `.templ` edit, force-regenerate:
  `rm <file>_templ.go && templ generate`.

# Color Rows → Components Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the theme page's bespoke color-override rows from reusable Goshtoso components — a "shell mode" on the existing Select plus a new generic Palette color picker.

**Architecture:** Extend `components/select` with a `Shell` mode (slim Alpine x-data, leading slot, value expression, children-as-dropdown-body, `select-close` event). Add `components/palette` (Tailwind hue×shade grid + white/black + Reset + optional hex). Theme page composes `Select(shell){ Palette }`; theme keeps its `pickColor/classLabel/resolved` logic and wires side effects via the palette's `OnSelectExpr`.

**Tech Stack:** Go 1.26, templ v0.3, Alpine.js v3, Tailwind v4, Playwright (Go) E2E. Working dir: worktree `.claude/worktrees/color-rows` on branch `feat/color-rows`.

**Escaping rule (verified):** Dynamic templ attribute values via `{ goExpr }` are written **raw** (single quotes preserved; see existing `:style={ "'background-color:'..." }`). Build all Alpine expressions with **single quotes only** — never double quotes, never `json.Marshal`. After any `.templ` edit run `templ generate`; never edit `*_templ.go`.

---

## File Structure

- `components/select/types.go` — add `Shell`, `TriggerLeading`, `ValueExpr` fields; `shellData()` JS builder.
- `components/select/select.templ` — branch trigger + dropdown body on `Shell`; `select-close` listener.
- `components/select/select_shell_test.go` — new render test (Go).
- `components/palette/types.go` — `Config`, `DefaultHues`, `DefaultShades`, `pickExpr`, `ContainerClasses`.
- `components/palette/palette.templ` — the grid component.
- `components/palette/palette_test.go` — new render test (Go).
- `internal/pages/demo/components/palette.templ` — demo page (`paletteDemoContent`).
- `internal/pages/demo/components/registry.go` — register `components/palette`.
- `internal/pages/demo/layout.templ` — sidebar group item + flat list entry.
- `internal/pages/demo/components/theme.templ` — rewrite `colorRow`, add `colorSwatch`, update `pickColor`, delete `tailwindPalette` + `tailwindHues` + `tailwindShades` + `togglePalette` + `openPalette`.
- `tests/e2e/sidebar_test.go` — add Palette to expected components.
- `tests/e2e/theme_page_test.go` — update color-row interactions.
- `tests/e2e/palette_test.go` — new E2E behavior test.

---

## Task 1: Select shell mode

**Files:**
- Modify: `components/select/types.go`
- Modify: `components/select/select.templ`
- Test: `components/select/select_shell_test.go`

- [ ] **Step 1: Write the failing render test**

Create `components/select/select_shell_test.go`:

```go
package selectfield

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderSelect(t *testing.T, cfg Config, children templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	ctx := context.Background()
	if children != nil {
		ctx = templ.WithChildren(ctx, children)
	}
	require.NoError(t, Select(cfg).Render(ctx, &buf))
	return buf.String()
}

func TestSelect_ShellMode_RendersValueExprAndChildren(t *testing.T) {
	body := templ.Raw(`<div data-test-body>palette here</div>`)
	html := renderSelect(t, Config{
		ID:        "color-surface",
		Shell:     true,
		ValueExpr: "classLabel('surface')",
	}, body)

	// slim shell x-data, not the option-based selectData
	assert.Contains(t, html, `x-data="{ isOpen: false, openedWithKeyboard: false }"`)
	assert.NotContains(t, html, "allOptions")
	// value expression drives the trigger text
	assert.Contains(t, html, `x-text="classLabel('surface')"`)
	// listens for the palette close event
	assert.Contains(t, html, "select-close")
	// hosts arbitrary children as the dropdown body, no option <ul>
	assert.Contains(t, html, "data-test-body")
	assert.Equal(t, 0, strings.Count(html, `role="option"`))
	// no hidden form input in shell mode
	assert.Equal(t, 0, strings.Count(html, `type="text"`))
}

func TestSelect_DataMode_Unchanged(t *testing.T) {
	html := renderSelect(t, Config{
		ID:      "fruit",
		Options: []Option{{Value: "a", Label: "Apple"}},
	}, nil)
	assert.Contains(t, html, "allOptions")
	assert.Contains(t, html, `role="listbox"`)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./components/select/ -run TestSelect_ShellMode -v`
Expected: FAIL — `Config` has no field `Shell` / `ValueExpr` (compile error).

- [ ] **Step 3: Add Config fields**

In `components/select/types.go`, inside the `Config` struct (after the `Attrs` field), add:

```go
	// Shell enables "shell mode": the Select renders its trigger + dropdown
	// chrome but hosts arbitrary templ children as the dropdown body instead
	// of an option list. Used to wrap custom pickers (e.g. a color palette).
	Shell bool
	// TriggerLeading is optional content rendered at the start of the trigger
	// in shell mode (e.g. a color swatch). Ignored when Shell is false.
	TriggerLeading templ.Component
	// ValueExpr is an Alpine expression (x-text) for the trigger's value text
	// in shell mode. Resolves against the host page's x-data scope.
	ValueExpr string
```

- [ ] **Step 4: Add the shell x-data builder**

In `components/select/types.go`, add at the end of the file:

```go
// shellData returns the slim Alpine x-data for shell mode: only open state,
// since the hosted content owns the value.
func shellData() string {
	return `{ isOpen: false, openedWithKeyboard: false }`
}
```

- [ ] **Step 5: Branch the template on Shell**

In `components/select/select.templ`, replace the inner state `<div>` opening tag (the one currently `x-data={ selectData(cfg) }` with the esc handler) with a Shell branch:

```templ
		<div
			if cfg.Shell {
				x-data={ shellData() }
				x-on:select-close="isOpen = false; openedWithKeyboard = false"
			} else {
				x-data={ selectData(cfg) }
			}
			x-on:keydown.esc.window="isOpen = false; openedWithKeyboard = false"
			class="relative"
		>
```

In the trigger `<button>`, replace the `x-bind:aria-label` line and the value `<span>` with branched versions. The button keeps its existing click/keydown handlers, `disabled?=`, `AlpineBindDisabled` conditional, and `class={ cfg.TriggerClasses() }`. Change the aria-label binding to:

```templ
				if cfg.Shell {
					x-bind:aria-label={ cfg.ValueExpr }
				} else {
					x-bind:aria-label="selectedOption ? selectedOption.label : placeholder"
				}
```

And replace the value `<span>` (currently `<span class="truncate text-sm font-normal" x-text="selectedOption ? selectedOption.label : placeholder"></span>`) with:

```templ
				if cfg.Shell {
					if cfg.TriggerLeading != nil {
						@cfg.TriggerLeading
					}
					<span class="truncate text-sm font-normal" x-text={ cfg.ValueExpr }></span>
				} else {
					<span class="truncate text-sm font-normal" x-text="selectedOption ? selectedOption.label : placeholder"></span>
				}
```

Wrap the hidden `<input>` (the `x-ref="hiddenInput"` one) so it only renders in data mode:

```templ
			if !cfg.Shell {
				<input
					id={ cfg.ID }
					name={ cfg.Name }
					type="text"
					hidden
					x-ref="hiddenInput"
					x-bind:value="selectedOption ? selectedOption.value : ''"
					if cfg.Autocomplete != "" {
						autocomplete={ cfg.Autocomplete }
					}
				/>
			}
```

In the dropdown panel `<div>` (the `role="listbox"` container with `x-show`), branch its body — render children in shell mode, the existing `<ul>...</ul>` otherwise:

```templ
				if cfg.Shell {
					{ children... }
				} else {
					<ul class="flex max-h-52 flex-col overflow-y-auto py-1" role="listbox">
						<!-- existing template x-for ... unchanged ... -->
					</ul>
				}
```

(Keep the `<ul>` content exactly as-is; only wrap it in the `else`.)

- [ ] **Step 6: Regenerate**

Run: `templ generate`
Expected: completes; `components/select/select_templ.go` updated.

- [ ] **Step 7: Run tests to verify they pass**

Run: `go test ./components/select/ -v`
Expected: PASS (both shell and data-mode tests).

- [ ] **Step 8: Commit**

```bash
git add components/select/
git commit -m "feat(select): add shell mode (leading slot, value expr, children body)"
```

---

## Task 2: Palette component

**Files:**
- Create: `components/palette/types.go`
- Create: `components/palette/palette.templ`
- Test: `components/palette/palette_test.go`

- [ ] **Step 1: Write the failing render test**

Create `components/palette/palette_test.go`:

```go
package palette

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func render(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Palette(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestPalette_GridCountAndPickExpr(t *testing.T) {
	html := render(t, Config{
		ID:           "palette-surface",
		OnSelectExpr: "pickColor('surface', $value)",
	})
	// one swatch button per hue×shade
	want := len(DefaultHues) * len(DefaultShades)
	assert.Equal(t, want, strings.Count(html, `data-cls=`), "one button per hue×shade")
	// pick wires the host expr with the concrete value + closes the shell
	assert.Contains(t, html, `@click="pickColor('surface', 'blue-700'); $dispatch('select-close')"`)
	// white/black + reset present by default
	assert.Contains(t, html, `@click="pickColor('surface', 'white'); $dispatch('select-close')"`)
	assert.Contains(t, html, `@click="pickColor('surface', ''); $dispatch('select-close')"`)
	// no hex section unless requested
	assert.NotContains(t, html, `type="color"`)
}

func TestPalette_AlpineModelAndHex(t *testing.T) {
	html := render(t, Config{
		ID:          "p",
		AlpineModel: "myColor",
		ShowHex:     true,
	})
	assert.Contains(t, html, `@click="myColor = 'blue-700'; $dispatch('select-close')"`)
	assert.Contains(t, html, `type="color"`)
	assert.Contains(t, html, `@change="myColor = $event.target.value; $dispatch('select-close')"`)
}

func TestPalette_HideFlags(t *testing.T) {
	html := render(t, Config{ID: "p", HideNeutral: true, HideReset: true})
	assert.NotContains(t, html, "Reset")
	assert.NotContains(t, html, `bg-white`)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./components/palette/ -v`
Expected: FAIL — package/`Config` does not exist (compile error).

- [ ] **Step 3: Create types.go**

Create `components/palette/types.go`:

```go
// Package palette renders a generic color picker grid: a Tailwind hue×shade
// matrix plus optional white/black swatches, a Reset action, and a hex input.
// It has no trigger of its own — wrap it in a Select shell (Shell: true) to get
// a dropdown. On pick it sets AlpineModel and/or runs OnSelectExpr (with $value
// substituted) and dispatches a `select-close` event so a hosting shell closes.
package palette

import (
	"fmt"
	"strings"
)

// DefaultHues lists Tailwind v4's named hue families in display order.
var DefaultHues = []string{
	"red", "orange", "amber", "yellow", "lime",
	"green", "emerald", "teal", "cyan", "sky",
	"blue", "indigo", "violet", "purple", "fuchsia",
	"pink", "rose", "slate", "gray", "zinc",
	"neutral", "stone",
}

// DefaultShades lists the shade steps for each hue.
var DefaultShades = []string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900", "950"}

// Config configures a Palette.
type Config struct {
	// ID is the wrapper element id.
	ID string
	// AlpineModel, when set, is assigned the picked value (a JS lvalue).
	AlpineModel string
	// OnSelectExpr, when set, is an Alpine expression run on pick. The literal
	// substring "$value" is replaced with the chosen value.
	OnSelectExpr string
	// Hues / Shades override the default Tailwind sets.
	Hues   []string
	Shades []string
	// HideNeutral hides the white/black quick swatches (shown by default).
	HideNeutral bool
	// HideReset hides the Reset action (shown by default).
	HideReset bool
	// ShowHex adds a native color input + hex text field (off by default).
	ShowHex bool
	// Class appends classes to the wrapper.
	Class string
}

func (c Config) hues() []string {
	if len(c.Hues) > 0 {
		return c.Hues
	}
	return DefaultHues
}

func (c Config) shades() []string {
	if len(c.Shades) > 0 {
		return c.Shades
	}
	return DefaultShades
}

// ContainerClasses returns wrapper classes.
func (c Config) ContainerClasses() string {
	base := "p-2 space-y-2"
	if c.Class != "" {
		base += " " + c.Class
	}
	return base
}

// pickExpr builds the Alpine @click/@change expression run when a value is
// chosen. valueJS is a JS expression evaluating to the chosen value string
// (e.g. "'blue-700'" or "$event.target.value"). Single quotes only — safe for
// raw templ attribute rendering.
func (c Config) pickExpr(valueJS string) string {
	parts := make([]string, 0, 3)
	if c.AlpineModel != "" {
		parts = append(parts, fmt.Sprintf("%s = %s", c.AlpineModel, valueJS))
	}
	if c.OnSelectExpr != "" {
		parts = append(parts, strings.ReplaceAll(c.OnSelectExpr, "$value", valueJS))
	}
	parts = append(parts, "$dispatch('select-close')")
	return strings.Join(parts, "; ")
}

// classValueJS returns the single-quoted JS literal for a hue-shade class.
func classValueJS(hue, shade string) string { return "'" + hue + "-" + shade + "'" }
```

- [ ] **Step 4: Create palette.templ**

Create `components/palette/palette.templ`:

```templ
package palette

templ Palette(cfg Config) {
	<div x-data="{ hovered: '' }" data-palette id={ cfg.ID } class={ cfg.ContainerClasses() }>
		if !cfg.HideNeutral || !cfg.HideReset {
			<div class="flex items-center gap-1">
				if !cfg.HideNeutral {
					<button
						type="button"
						data-cls="white"
						@click={ cfg.pickExpr("'white'") }
						@mouseenter="hovered = 'white'"
						@mouseleave="hovered = ''"
						class="h-5 w-5 rounded border border-outline bg-white hover:ring-2 hover:ring-primary dark:border-outline-dark dark:hover:ring-primary-dark"
						title="white"
					></button>
					<button
						type="button"
						data-cls="black"
						@click={ cfg.pickExpr("'black'") }
						@mouseenter="hovered = 'black'"
						@mouseleave="hovered = ''"
						class="h-5 w-5 rounded border border-outline bg-black hover:ring-2 hover:ring-primary dark:border-outline-dark dark:hover:ring-primary-dark"
						title="black"
					></button>
				}
				<span class="ml-2 text-[11px] font-mono text-on-surface-muted dark:text-on-surface-dark-muted truncate" x-text="hovered || 'Pick a color'"></span>
				if !cfg.HideReset {
					<button
						type="button"
						@click={ cfg.pickExpr("''") }
						class="ml-auto text-[10px] font-medium text-on-surface-muted dark:text-on-surface-dark-muted hover:text-primary dark:hover:text-primary-dark"
					>Reset</button>
				}
			</div>
		}
		<div class="grid w-full gap-1 max-h-44 overflow-y-auto pr-1" style="grid-template-columns: repeat(11, minmax(0, 1fr));">
			for _, hue := range cfg.hues() {
				for _, shade := range cfg.shades() {
					<button
						type="button"
						data-cls={ hue + "-" + shade }
						@click={ cfg.pickExpr(classValueJS(hue, shade)) }
						@mouseenter={ "hovered = '" + hue + "-" + shade + "'" }
						@mouseleave="hovered = ''"
						@focus={ "hovered = '" + hue + "-" + shade + "'" }
						@blur="hovered = ''"
						class="h-5 w-full rounded-sm border border-outline/30 dark:border-outline-dark/30 transition-transform hover:scale-125 hover:ring-2 hover:ring-primary focus:scale-125 dark:hover:ring-primary-dark"
						style={ "background-color: var(--color-" + hue + "-" + shade + ")" }
						title={ hue + "-" + shade }
					></button>
				}
			}
		</div>
		if cfg.ShowHex {
			<div class="flex items-center gap-2 pt-1">
				<input
					type="color"
					@change={ cfg.pickExpr("$event.target.value") }
					class="size-7 rounded border border-outline dark:border-outline-dark cursor-pointer"
					title="Custom color"
				/>
				<input
					type="text"
					placeholder="#000000"
					@change={ cfg.pickExpr("$event.target.value") }
					class="w-24 rounded-radius border border-outline bg-surface px-2 py-1 text-xs font-mono text-on-surface dark:border-outline-dark dark:bg-surface-dark dark:text-on-surface-dark"
				/>
			</div>
		}
	</div>
}
```

- [ ] **Step 5: Regenerate**

Run: `templ generate`
Expected: creates `components/palette/palette_templ.go`.

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./components/palette/ -v`
Expected: PASS (all three tests).

- [ ] **Step 7: Verify no escaping corruption**

The `TestPalette_GridCountAndPickExpr` assertion in Step 1 already proves the
`@click` renders unescaped (it asserts the exact literal
`@click="pickColor('surface', 'blue-700'); $dispatch('select-close')"`). As a
belt-and-suspenders grep against the generated file:

Run: `grep -c '&#39;\|&quot;' components/palette/palette_templ.go || true`
Expected: the only matches (if any) are inside Go string-literal helper code,
not inside emitted `@click`/`@change` attribute strings. The Step-1 PASS is the
authoritative check.

- [ ] **Step 8: Commit**

```bash
git add components/palette/
git commit -m "feat(palette): generic color picker grid (tailwind + hex)"
```

---

## Task 3: Palette demo page + registration

**Files:**
- Create: `internal/pages/demo/components/palette.templ`
- Modify: `internal/pages/demo/components/registry.go`
- Modify: `internal/pages/demo/layout.templ`
- Modify: `tests/e2e/sidebar_test.go`

- [ ] **Step 1: Add the sidebar test expectation (failing first)**

In `tests/e2e/sidebar_test.go`, add to the `expectedComponents` slice (keep alphabetical-ish with the Form group; place after the Pagination entry):

```go
		{"/components/palette", "Palette"},
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./tests/e2e/ -run TestSidebar_AllComponentsPresent -count=1 -timeout 3m`
Expected: FAIL — no sidebar link for `/components/palette`.

- [ ] **Step 3: Create the demo page**

Create `internal/pages/demo/components/palette.templ`:

```templ
package components

import (
	palettecomp "github.com/araihu/goshtoso/components/palette"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/internal/pages/demo"
)

func paletteDemoContent() templ.Component { return paletteDemo() }

templ paletteDemo() {
	@demo.DemoPage(demo.PageConfig{Title: "Palette", ActiveComponent: "palette"}) {
		<div class="space-y-8" x-data="{ picked: '', shellPicked: '' }">
			<div>
				<h2 class="text-2xl font-bold font-title text-on-surface-strong dark:text-on-surface-dark-strong mb-2">Palette</h2>
				<p class="text-sm text-on-surface dark:text-on-surface-dark mb-4">
					A generic color picker grid (Tailwind hues + white/black + hex). Pick a swatch:
				</p>
				<div class="max-w-md rounded-radius border border-outline dark:border-outline-dark">
					@palettecomp.Palette(palettecomp.Config{ID: "demo-palette", AlpineModel: "picked", ShowHex: true})
				</div>
				<p class="mt-2 text-sm font-mono text-on-surface dark:text-on-surface-dark">Selected: <span x-text="picked || '—'"></span></p>
			</div>
			<div>
				<h3 class="text-lg font-semibold text-on-surface-strong dark:text-on-surface-dark-strong mb-2">Inside a Select shell</h3>
				<div class="max-w-xs">
					@selectfield.Select(selectfield.Config{
						ID:        "demo-shell",
						Shell:     true,
						ValueExpr: "shellPicked || 'Pick a color'",
					}) {
						@palettecomp.Palette(palettecomp.Config{ID: "demo-shell-palette", AlpineModel: "shellPicked", ShowHex: true})
					}
				</div>
			</div>
		</div>
	}
}
```

Note: if `demo.DemoPage`/`demo.PageConfig` do not match the existing demo-page wrapper used by sibling files, open an existing demo (e.g. `internal/pages/demo/components/select.templ`) and copy its exact page-wrapper invocation and import alias instead. The component-call blocks above stay the same.

- [ ] **Step 4: Register the route**

In `internal/pages/demo/components/registry.go`, add to the `Demos` map (after the `components/pagination` line):

```go
	"components/palette":         {"Palette", "palette", paletteDemoContent},
```

- [ ] **Step 5: Add sidebar entries**

In `internal/pages/demo/layout.templ`, in the Form group (after the `radio` line, before `select`):

```templ
				sItem("palette", "Palette", "/components/palette", activeComponent),
```

And in the flat component list (after the `Select` entry):

```templ
	{"Palette", "/components/palette"},
```

- [ ] **Step 6: Regenerate + build**

Run: `templ generate && go build ./...`
Expected: builds clean.

- [ ] **Step 7: Run sidebar + a smoke nav test**

Run: `go test ./tests/e2e/ -run TestSidebar_AllComponentsPresent -count=1 -timeout 3m`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/pages/demo/ tests/e2e/sidebar_test.go
git commit -m "feat(demo): add Palette demo page + sidebar/registry entries"
```

---

## Task 4: Theme page integration

**Files:**
- Modify: `internal/pages/demo/components/theme.templ`

- [ ] **Step 1: Make pickColor handle reset**

In `internal/pages/demo/components/theme.templ`, find the `pickColor(token, cls)` method (in the Alpine script, ~line 905):

```js
    pickColor(token, cls) {
      this.setColor(token, cls);
      this.openPalette = '';
      requestAnimationFrame(() => this.refreshResolved());
    },
```

Replace with (drop `openPalette`, delegate empty to clearColor):

```js
    pickColor(token, cls) {
      if (!cls) { this.clearColor(token); }
      else { this.setColor(token, cls); }
      requestAnimationFrame(() => this.refreshResolved());
    },
```

- [ ] **Step 2: Add the import + colorSwatch + rewrite colorRow**

In `internal/pages/demo/components/theme.templ`, add to the import block:

```go
	palettecomp "github.com/araihu/goshtoso/components/palette"
```

(`selectfield` is already imported from the earlier filter-select migration.)

Replace the entire `templ colorRow(t colorToken) { ... }` block (currently the button + `tailwindPalette` dropdown) with:

```templ
templ colorSwatch(token string) {
	<span
		class="size-7 shrink-0 rounded border border-outline dark:border-outline-dark"
		:style={ "'background-color:' + (resolved['" + token + "'] || '#888')" }
	></span>
}

templ colorRow(t colorToken) {
	@selectfield.Select(selectfield.Config{
		ID:             "color-" + t.Key,
		Shell:          true,
		TriggerLeading: colorSwatch(t.Key),
		ValueExpr:      "classLabel('" + t.Key + "')",
	}) {
		@palettecomp.Palette(palettecomp.Config{
			ID:           "palette-" + t.Key,
			OnSelectExpr: "pickColor('" + t.Key + "', $value)",
			ShowHex:      true,
		})
	}
}
```

- [ ] **Step 3: Delete the now-dead helpers**

In the same file delete:
- `templ tailwindPalette(token string) { ... }` (entire block).
- `var tailwindHues = []string{ ... }` and `var tailwindShades = []string{ ... }`.

Then verify nothing else references `togglePalette` or `openPalette`. Run:

Run: `grep -n "tailwindHues\|tailwindShades\|tailwindPalette\|togglePalette\|openPalette" internal/pages/demo/components/theme.templ`
Expected: no matches. If `openPalette`/`togglePalette` remain (e.g. an `openPalette: ''` field or `togglePalette()` method in the Alpine script), delete those too — they are unused after the rewrite.

- [ ] **Step 4: Regenerate + build**

Run: `templ generate && go build ./...`
Expected: builds clean (no unused-import / undefined errors).

- [ ] **Step 5: Commit**

```bash
git add internal/pages/demo/components/theme.templ
git commit -m "refactor(theme): rebuild color rows on Select shell + Palette"
```

---

## Task 5: Update existing theme E2E tests

**Files:**
- Modify: `tests/e2e/theme_page_test.go`

- [ ] **Step 1: Find color-row interactions to fix**

Run: `grep -n "openPalette\|togglePalette\|tailwindPalette\|data-token\|colorRow\|pickColor\|-trigger\|data-cls" tests/e2e/theme_page_test.go`
Expected: a list of selectors that drove the old markup.

- [ ] **Step 2: Rewrite each to the new structure**

For each color-row interaction, the new structure is: trigger button `#color-<token>-trigger`, dropdown swatch buttons `#palette-<token> button[data-cls="<hue>-<shade>"]`. A pick updates `document.documentElement.style.getPropertyValue('--color-<token>')`.

Replace an old "open palette + pick a color" interaction with this pattern (example for the first color token used by the existing test — keep the token the test already targets):

```go
	// Open the color picker for the token and choose blue-700.
	require.NoError(t, page.Locator("#color-surface-dark-trigger").Click())
	require.NoError(t, page.Locator(`#palette-surface-dark button[data-cls="blue-700"]`).Click())
	_, err = page.WaitForFunction(
		`() => getComputedStyle(document.documentElement).getPropertyValue('--color-surface-dark').trim() !== ''`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)
```

(Use the exact token key the existing test referenced; if it used `data-token="..."`, the same key becomes the `#color-<key>-trigger` / `#palette-<key>` id.)

- [ ] **Step 3: Run the theme tests**

Run: `go test ./tests/e2e/ -run TestThemePage -count=1 -timeout 8m`
Expected: PASS. If a test asserted on `openPalette` Alpine state, replace with a DOM/`--color-*` assertion as above.

- [ ] **Step 4: Commit**

```bash
git add tests/e2e/theme_page_test.go
git commit -m "test(theme): update color-row E2E to shell+palette structure"
```

---

## Task 6: Palette behavior E2E

**Files:**
- Create: `tests/e2e/palette_test.go`

- [ ] **Step 1: Write the behavior test**

Create `tests/e2e/palette_test.go`:

```go
package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestPalette_PickSetsModel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/palette", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	// Standalone palette → AlpineModel "picked".
	require.NoError(t, page.Locator(`#demo-palette button[data-cls="blue-700"]`).Click())
	require.NoError(t, page.Locator("text=Selected: blue-700").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(2000),
	}))
}

func TestPalette_InShell_OpensPicksCloses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/palette", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	swatch := page.Locator(`#demo-shell-palette button[data-cls="red-500"]`)
	// Closed initially.
	hidden, err := swatch.IsHidden()
	require.NoError(t, err)
	require.True(t, hidden)
	// Open via the shell trigger.
	require.NoError(t, page.Locator("#demo-shell-trigger").Click())
	require.NoError(t, swatch.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(2000),
	}))
	// Pick → trigger reflects value AND panel closes (select-close).
	require.NoError(t, swatch.Click())
	require.NoError(t, page.Locator("#demo-shell-trigger").GetByText("red-500").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(2000),
	}))
	require.NoError(t, swatch.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden, Timeout: playwright.Float(2000),
	}))
}
```

- [ ] **Step 2: Run the test**

Run: `go test ./tests/e2e/ -run TestPalette -count=1 -timeout 5m`
Expected: PASS. If `setupPlaywright`/`baseURL`/`newPage` signatures differ, mirror an existing test in the same package (e.g. `theme_page_test.go`).

- [ ] **Step 3: Commit**

```bash
git add tests/e2e/palette_test.go
git commit -m "test(palette): E2E pick-sets-model + shell open/pick/close"
```

---

## Task 7: Full verification + assets

**Files:** none new (verification + optional asset rebuild).

- [ ] **Step 1: Rebuild Tailwind (new utility classes may have appeared)**

Run: `tailwindcss -i css/main.css -o assets/styles.css`
Expected: writes `assets/styles.css`. The palette reuses classes already present from the old `tailwindPalette`, so the diff should be minimal/empty.

- [ ] **Step 2: Build the server binary**

Run: `go build -o bin/server ./cmd/server`
Expected: succeeds.

- [ ] **Step 3: Run the affected E2E suites**

Run: `go test ./tests/e2e/ -run 'TestThemePage|TestSelect|TestPalette|TestSidebar' -count=1 -timeout 12m`
Expected: PASS, 0 failures.

- [ ] **Step 4: Run the component unit tests**

Run: `go test ./components/... -count=1`
Expected: PASS.

- [ ] **Step 5: Commit any asset/diff**

```bash
git add -A
git commit -m "chore: rebuild styles.css for palette/select-shell" || echo "nothing to commit"
```

- [ ] **Step 6: Manual verification**

Run the dev server (`go run cmd/server/main.go`, port 8090) and confirm at `/docs/theme`: each color row opens a palette on click, picking a swatch updates the row swatch + value and closes the dropdown, Reset clears the override, hex input applies a custom color. Confirm `/components/palette` demo works in both light and dark, Minimal + one other theme.

---

## Notes / Out of scope

- `contrastBase` inline select and the font/`cssFilter` selects are untouched (the latter were migrated earlier).
- Shell mode deliberately ignores `AlpineBindDisabled` (color rows never disable).
- If `templ generate` reports "0 updates" after an edit, force it: `rm <component>_templ.go && templ generate`.

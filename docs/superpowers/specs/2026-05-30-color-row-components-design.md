# Design: Color Rows → Goshtoso Components

**Date:** 2026-05-30
**Status:** Approved (design phase)

## Problem

The theme page (`internal/pages/demo/components/theme.templ`) renders ~24 color
override rows via bespoke `colorRow(t)` + `tailwindPalette(token)` templ helpers.
These are real UI patterns (a swatch-select trigger + a color-picker dropdown)
but they are hand-rolled in the demo package, not built from Goshtoso
components. Goal: dogfood. Rebuild the rows from reusable components, extending
components where they fall short.

## Decisions (from brainstorm)

1. **Two components**: reuse the existing **Select** component as the dropdown
   *shell* (trigger + open/close + panel chrome), and add a new generic
   **Palette** component for the color grid. (Not a single monolithic
   ColorPicker; not in-page-only helpers.)
2. **Generic palette**: the Palette supports arbitrary **hex** values (native
   `<input type="color">` + hex text) in addition to the Tailwind class grid.

## Architecture

Three deliverables: extend Select, add Palette, rewire the theme page.

### A. `components/select` — add "shell mode"

Existing data-driven Select is unchanged when `Shell == false` (default). When
`Shell == true`:

- **x-data** is slim: `{ isOpen: false, openedWithKeyboard: false }` — no
  `allOptions` / `selectedValues` / `selectOption`. (Implemented as a separate
  `shellData()` JS builder so `selectData()` stays untouched.)
- **Trigger** renders, left→right: `TriggerLeading` (optional `templ.Component`,
  e.g. a swatch) → `<span x-text={ ValueExpr }>` → the existing chevron svg.
  Reuses `TriggerClasses()` so visual parity with the normal Select holds.
- **Dropdown panel** renders templ `{ children... }` instead of the option
  `<ul>`. Same panel container classes, same `x-show` / `x-transition` /
  `x-on:click.outside` / esc-to-close behavior.
- **Close-on-pick**: the panel listens `x-on:select-close="isOpen=false;
  openedWithKeyboard=false"` so hosted content (Palette) can request a close by
  dispatching a `select-close` event. (`.stop` so it doesn't escape the widget.)

New Config fields:

```go
Shell          bool             // enable shell mode
TriggerLeading templ.Component  // optional leading content in the trigger
ValueExpr      string           // Alpine x-text expression for the trigger value text
```

Notes:
- In shell mode the option-list keyboard helpers (`x-trap`, `$focus.wrap()`) are
  omitted — hosted content owns its own focusables.
- `ValueExpr` / `TriggerLeading` expressions resolve up Alpine's scope chain to
  the host page's `x-data` (e.g. `classLabel('surface-dark')`), so Select stays
  generic and the host owns the display logic.

### B. `components/palette` — new generic color picker grid

A presentational + interactive color grid. No trigger of its own (the Select
shell provides that). Rendered inside a host `x-data`; internal x-data only
holds `{ hovered: '' }`.

```go
type Config struct {
    ID           string
    AlpineModel  string   // optional: var assigned on pick (e.g. "myColor")
    OnSelectExpr string   // optional: Alpine expr run on pick, receives $value
    Hues         []string // default: DefaultHues (Tailwind v4 hue families)
    Shades       []string // default: DefaultShades (50..950)
    ShowNeutral  bool     // white/black quick swatches (default true)
    ShowHex      bool     // native <input type=color> + hex text input
    ShowReset    bool     // Reset action (default true)
    Class        string   // extra wrapper classes
}
```

Markup (mirrors today's `tailwindPalette`, generalized):
- Top row (when `ShowNeutral`): white + black swatch buttons, a live
  `x-text="hovered || …"` label, and (when `ShowReset`) a Reset button.
- Grid: `Hues × Shades` buttons; background `var(--color-<hue>-<shade>)` so the
  grid reflects the live theme (including overrides). Hover/focus sets `hovered`.
- Hex section (when `ShowHex`): `<input type="color">` + a `#rrggbb` text input,
  both feeding the same pick path.

**Pick contract** — a single internal `pick(value)` JS helper:
1. if `AlpineModel` set → assign it `value`;
2. if `OnSelectExpr` set → evaluate it with `$value === value` in scope;
3. always `$dispatch('select-close')` so a hosting Select shell closes.

`value` is one of: a Tailwind class `"blue-700"`, `"white"`, `"black"`, a hex
`"#aabbcc"`, or `""` (Reset). The component does not interpret the value beyond
previewing the grid; semantics belong to the host.

Exported defaults: `palette.DefaultHues`, `palette.DefaultShades` (moved out of
theme.templ's `tailwindHues`/`tailwindShades`, which are then deleted).

Escaping: follow CLAUDE.md rules — build the x-data JS via a single-quoted
string builder (no `json.Marshal`), reference via `x-data="paletteX"` /
`Alpine.data()` if it grows complex; inline object form is acceptable if it has
no quoted string literals that templ would mangle.

### C. Theme page integration

`colorRow(t colorToken)` is rebuilt as a Select shell hosting a Palette:

```templ
templ colorRow(t colorToken) {
    @selectfield.Select(selectfield.Config{
        ID:             "color-" + t.Key,
        Shell:          true,
        TriggerLeading: colorSwatch(t.Key),          // swatch templ component
        ValueExpr:      "classLabel('" + t.Key + "')",
    }) {
        @palette.Palette(palette.Config{
            ID:           "palette-" + t.Key,
            OnSelectExpr: "pickColor('" + t.Key + "', $value)",
            ShowHex:      true,
        })
    }
}
```

- The theme page keeps `pickColor`, `clearColor`, `classLabel`, `currentClass`,
  `resolved`, `overrides`, `refreshResolved`. `pickColor(token, value)` is
  updated so an empty `value` delegates to `clearColor(token)` (Reset path),
  then refreshes — giving the Palette a single pick entry point for both
  selection and reset.
- `pickColor` already does `requestAnimationFrame(() => refreshResolved())`, so
  the live preview swatch and value text update after a pick. The shell closes
  via the `select-close` event.
- The page-level `openPalette` state and the `togglePalette` helper become
  unnecessary (the Select shell owns open/close) and are removed.
- `colorSwatch(token)` is a small templ component rendering the leading swatch
  (`:style` = `'background-color:' + (resolved['<token>'] || '#888')`).

## Data flow

```
user clicks a grid swatch
  → Palette.pick("blue-700")
      → OnSelectExpr: pickColor("surface-dark", "blue-700")   [host scope]
          → setColor + rAF→refreshResolved  → resolved/overrides update
      → $dispatch('select-close')
  → Select shell: isOpen=false
  → trigger re-renders: ValueExpr classLabel(...) + leading swatch :style
```

No new global state; Alpine scope-chain resolution links shell ↔ palette ↔ host.

## Edge cases

- **Reset**: Palette Reset button picks `""` → `pickColor(token,"")` clears the
  override; trigger value falls back to the theme default label; swatch falls
  back to `resolved[token]` (theme value) or `#888`.
- **Hex vs class**: native color input only round-trips hex; if the current
  value is a class the input shows its default — acceptable, hex is an alternate
  entry path, not a mirror.
- **Disabled/dynamic**: not needed for color rows; shell mode ignores
  `AlpineBindDisabled` unless wired (out of scope).
- **templ escaping**: `ValueExpr` and `OnSelectExpr` contain single quotes only;
  verify rendered HTML has no `&quot;`/`&#39;` inside `x-data`/`x-text`.

## Components workflow obligations (CLAUDE.md)

- Add a demo page for Palette: `internal/pages/demo/components/palette.templ`,
  register route in `internal/server/server.go:handleComponent()`, add sidebar
  entry in `layout.templ:getSidebarItems()` and the registry. (Select already
  has a demo; add a "shell mode" example there.)
- `templ generate` after editing `.templ`; never edit `*_templ.go`.

## Testing

- **Select shell (new)** `tests/e2e/select_test.go` or a new test: render a
  shell-mode Select with arbitrary children; assert trigger shows `ValueExpr`,
  panel opens on click, child content visible, `select-close` event closes it.
- **Palette (new)** `tests/e2e/palette_test.go`: grid renders `len(Hues)*
  len(Shades)` swatches; clicking a swatch sets `AlpineModel` / fires
  `OnSelectExpr` with the right `$value`; Reset fires `""`; hex input path sets a
  hex value; hover updates the label.
- **Theme page (existing)** `tests/e2e/theme_page_test.go`: update any test that
  drove the old `colorRow`/`tailwindPalette` markup (`openPalette`,
  `togglePalette`, `data-token` selectors) to the new shell+palette structure.
  Confirm picking a color updates `--color-<token>` and the trigger swatch/label;
  confirm Reset clears it.
- Full theme + select + palette E2E green; full suite unaffected.

## Files touched

- `components/select/types.go` — new fields + shell branch in classes if needed.
- `components/select/select.templ` — shell-mode trigger + children panel +
  `shellData()`; `templ generate`.
- `components/palette/types.go` + `components/palette/palette.templ` — new.
- `internal/pages/demo/components/theme.templ` — rewrite `colorRow`, add
  `colorSwatch`, delete `tailwindPalette` + `tailwindHues`/`tailwindShades` +
  `togglePalette`/`openPalette` usage.
- `internal/pages/demo/components/palette.templ` + registry + `server.go` +
  `layout.templ` — demo page + route + sidebar.
- `tests/e2e/*` — new palette/shell tests; update theme tests.

## Out of scope (YAGNI)

- Migrating the `contrastBase` inline select (separate composite widget).
- Disabled/readonly states for shell mode.
- Multi-select / multiple colors per row.

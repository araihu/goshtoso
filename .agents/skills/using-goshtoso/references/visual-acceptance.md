# Visual acceptance for Goshtoso applications

Run this checklist after build and interaction tests. It is intentionally short
enough to use on every application surface and strict enough to catch a page
that merely compiles.

## Required matrix

Inspect every changed archetype at:

| Viewport | Theme | Color mode |
|---|---|---|
| 390 px wide | Goshtoso | light |
| 390 px wide | Goshtoso | dark |
| 390 px wide | Minimal | light |
| 390 px wide | Minimal | dark |
| 1440 px wide | Goshtoso | light |
| 1440 px wide | Goshtoso | dark |
| 1440 px wide | Minimal | light |
| 1440 px wide | Minimal | dark |

Also inspect loading, empty, error, and success. Add permission-denied, stale,
partial, destructive confirmation, and filtered-empty when the domain supports
them.

## Hierarchy and composition

- The rendered surface still matches the task, register, information priority,
  density, and visual direction recorded in `design-intelligence.md`.
- The page has one obvious primary task and one `h1`.
- Primary, secondary, and destructive actions do not compete visually.
- Dense work surfaces use stable rows, columns, and alignment instead of equal
  decorative cards.
- Secondary rails and metadata remain subordinate to the main task.
- Text line length stays readable and long labels do not break controls.
- Skeletons match the shape of final content and empty states teach the next
  useful action.

## Responsive behavior

- No horizontal page scroll at 390 px. A data table may own horizontal overflow.
- Drawers begin below a sticky top bar and do not trap content behind overlays.
- Toolbars stack or collapse without changing control order unexpectedly.
- Detail rails follow main content on small screens.
- Sticky actions do not cover focused inputs, errors, or the last result row.
- Inspect internal overflow containers, not only document width. A clipped
  alert inside a sidebar can pass a document-level horizontal-scroll check.

## Themes and color

- Goshtoso and Minimal preserve the same hierarchy even when radius, font, and
  contrast personality change.
- Light and dark modes expose readable text, outlines, hover, focus, selected,
  disabled, loading, error, warning, success, and info states.
- Status never depends on color alone.
- Primary color marks actions and selection, not decoration.
- Typography belongs to the product. Do not reflexively choose Inter, Geist,
  Roboto, or another fashionable default across unrelated domains; use the
  system stack or make a deliberate, tested type choice.
- Do not replace that monoculture with a reflexive Georgia/Times editorial
  stack. Type should express this product's domain, not merely look different
  from the fashionable sans-serif default.
- Avoid the generated “ghost card” formula: a border plus a broad diffuse
  shadow on every surface. Use borders for structure and reserve elevation for
  a real layering relationship such as a dialog, popover, or dragged item.
- Reject decorative gradient text, glass effects, side stripes, category-coded
  palettes, and excessive pill/rounded containers unless the surface brief
  records a product-specific reason for them.

## Keyboard and focus

- Use Tab and Shift+Tab through the complete task.
- Activate every control with the expected keyboard command.
- Escape closes overlays and returns focus to the trigger.
- HTMX navigation and validation move focus deliberately.
- Focus is always visible and never hidden under sticky UI.

## Automated checks

- Build and unit/render tests pass in a fresh consumer module.
- Browser tests exercise direct navigation, HTMX navigation, Back, and refresh.
- Capture JavaScript console errors and fail on unexpected messages.
- Run axe or an equivalent accessibility scanner and resolve every P1 issue.
- Assert that every utility required by the recipe exists in the delivered CSS.
- Save screenshots for the required viewport, theme, color-mode, and state
  matrix so regressions can be compared rather than remembered.

## Completion statement

Record the tested routes, viewport matrix, states, console result, accessibility
result, and known exceptions. Do not describe a surface as visually verified if
it was only inspected as templ source or rendered HTML.

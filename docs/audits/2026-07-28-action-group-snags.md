# Action Group snag journal

## Scope

Generic responsive `ActionGroup` for Goshtoso consumers. Primary action stays
visible; ordered secondary actions collapse into one flat overflow Dropdown.

## Snags and decisions

- Source inspection found Dropdown intentionally flat: `Section` owns
  `[]Item`, with no submenu model, nested `role="menu"`, or Right/Left submenu
  keyboard behavior. Final design preserves that contract. A stacked
  `Action.Items` renders its own ordinary Dropdown at normal width, then
  flattens into a labeled section of ActionGroup's shared overflow Dropdown.
- Dropdown `Section` has no heading field. Flattened groups therefore preserve
  context with a disabled label item at the start of their existing section.
  This avoids a Dropdown API change and keeps menu structure flat.
- Browser verification exposed a pre-existing keyboard gap: ArrowDown opened a
  flat Dropdown but left focus on its trigger, so menu Up/Down handlers could
  not run. The composition-required fix adds trigger/menu refs, focuses the
  first enabled item on Space/Enter/ArrowDown, and returns focus on Escape.
- Responsive transform needs actual container width because action labels and
  consumer placements vary. Container queries on an intrinsic-width action
  cluster risk self-referential sizing, so measurement is isolated in the
  embedded `action-group.js` ResizeObserver helper. Without JavaScript, all
  secondary actions remain visible and wrap; overflow stays hidden.
- Exact compact Button API is `button.SizeSmall` through
  `button.WithSize(button.SizeSmall)`. ActionGroup uses it for native button
  actions; button-appearance links use the matching `link.SizeSmall`.
- Axe 4.12.1 reports zero violations when scoped to
  `[data-goshtoso-action-group]`. The full documentation page reports three
  existing `scrollable-region-focusable` findings on shared demo code-preview
  containers, not on ActionGroup markup; this slice records but does not widen
  into the shared documentation shell.

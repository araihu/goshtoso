# Component API Naming

This guide defines the naming grammar for public Goshtoso component config
fields. Use it when adding a component, changing a component API, or normalizing
similar props across components.

This maintainer guide complements the public interfaces documented in the
[Goshtoso Component Model](COMPONENT_MODEL.md). Read that guide first for the
common component interface, concrete return values, constructor styles, and
stable `Kind` identity.

Breaking changes are acceptable during the normalization pass. Prefer the
canonical name over preserving an old component-specific spelling.

## Core Rule

Name a field by the concept it controls and, when needed, by the rendered target
it applies to. A shared concept should use the same word everywhere; a target
specific hook should name the DOM or component surface it affects.

## Targets

Use `Root` for the outermost rendered element.

Use target prefixes for nested surfaces:

- `InputAttrs`, `InputClass` for native input elements.
- `TriggerAttrs`, `TriggerClass` for buttons or controls that open something.
- `PanelAttrs`, `PanelClass` for dropdown, drawer, modal, tooltip, or popover
  surfaces.
- `ListAttrs`, `ListClass` and `ItemAttrs`, `ItemClass` for repeated lists.
- `NavAttrs`, `NavClass` only when the rendered element is semantically a
  `<nav>`.

Avoid a bare `Class` or `Attrs` in new public APIs unless a component renders a
single obvious element and no likely future nested target exists. For most
components, prefer `RootClass` and `RootAttrs`.

## Text Fields

Use these names consistently:

- `Label`: visible form-control or choice label.
- `Title`: heading/title text.
- `Description`: descriptive body copy.
- `HelperText`: form-control helper, validation, or error text.
- `Placeholder`: native placeholder or empty-selection placeholder.
- `AriaLabel`: accessible label when no visible label exists.

Avoid generic `Text` when a more specific role exists. Keep `Text` only for a
component whose primary job is rendering one short text value and no clearer
semantic name applies.

## Actions And Triggers

Action fields should name the role of the action:

- `TriggerLabel`: label for the control that opens or reveals the component.
- `ActionLabel`: label for a single generic action.
- `PrimaryLabel` and `SecondaryLabel`: labels for paired actions.
- `AddLabel`, `RemoveLabel`, `ClearLabel`, `DismissLabel`: labels for named
  actions when the action itself is the public concept.

For action behavior, group related transport details in a config struct when the
component has more than one action or more than one transport option. Prefer
`HTMX *HTMXConfig` and `Alpine *AlpineConfig` over scattered `HX*` or
`Alpine*` fields.

## Form Controls

Form controls should converge on these field names where applicable:

- `ID`
- `Name`
- `Value`
- `Label`
- `Placeholder`
- `HelperText`
- `State`
- `Disabled`
- `Readonly`
- `Required`
- `Autocomplete`
- `InputAttrs`
- `RootClass`
- `RootAttrs`

Use `Readonly` with a lowercase `o`, matching the HTML attribute spelling. Use
`Disabled` only for controls that cannot be interacted with.

## Options And Items

For selectable options, use:

- `Value`: submitted or selected value.
- `Label`: visible option text.
- `Description`: supporting option text.
- `Disabled`: unavailable option.

Use `Items` for navigation/menu/list entries, `Options` for user-selectable
choices, and `Rows`/`Columns` for tabular data.

## Transport And Reactivity

Prefer grouped integration structs:

- `HTMX *HTMXConfig`
- `Alpine *AlpineConfig`

Inside those structs, name fields after the attribute without inconsistent
capitalization:

- `Get`, `Post`, `Target`, `Swap`, `Trigger`, `PushURL`
- `Model`, `BindDisabled`, `Data`

Avoid new public fields named `HXGet`, `HxGet`, `HTMXEndpoint`, or
`AlpineModel` on the root config when the same behavior can live under a grouped
config.

## Migration Checklist

When renaming public fields:

1. Update `components/<name>/types.go`.
2. Update the `.templ` source and demo pages.
3. Regenerate templ output with `templ generate`.
4. Run `go run ./scripts/skillgen` if a component `types.go` or entry point
   changed.
5. Update API reference tables and examples.
6. Add focused tests for render contracts or renamed behavior when risk is not
   purely mechanical.

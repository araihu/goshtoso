# Structured Input Consolidation Design

## Summary

Replace the separate `keyvalue` and `triplet` public component APIs with one
`structuredinput` component. The new component renders repeatable rows of
schema-defined inputs for structured form data: the former key/value editor is a
two-column structured input, and the former triplet editor is a three-column
structured input with a select column.

This is an intentional breaking change. The old `components/keyvalue` and
`components/triplet` packages should be removed rather than preserved as
compatibility wrappers.

## Goals

- Provide one reusable component for repeatable structured rows.
- Support the current key/value and key/value/effect use cases without duplicate
  rendering logic.
- Make submission shape explicit and predictable for arbitrary columns.
- Update the form field integration so consumers see one built-in field type.
- Replace the two demo pages with one docs page following the component docs
  pattern.

## Non-Goals

- Do not fold this into `schemafield`; schema-derived whole-form rendering is a
  different responsibility from repeatable row editing.
- Do not add server-side validation or HTMX behavior in this pass.
- Do not support every possible field type immediately. Text and select fields
  cover the existing behavior.
- Do not keep deprecated public wrappers for `keyvalue` or `triplet`.

## Public API

Create `components/structuredinput` with these core types:

```go
type ColumnType string

const (
    ColumnText   ColumnType = "text"
    ColumnSelect ColumnType = "select"
)

type Option struct {
    Value string
    Label string
}

type Column struct {
    Key         string
    Label       string
    Type        ColumnType
    Placeholder string
    Options     []Option
    Default     string
}

type Entry map[string]string

type Config struct {
    ID       string
    Name     string
    Columns  []Column
    Entries  []Entry
    AddLabel string
    Disabled bool
    Class    string
}
```

The public entry point is:

```go
templ StructuredInput(cfg Config)
```

`Column.Key` is required for every column. Empty keys are ignored. Duplicate keys
keep the first column and ignore later duplicates so generated input names stay
unique and predictable. `Column.Type` defaults to `ColumnText` when omitted.

## Rendering Behavior

The component renders:

- A root `<div>` with optional `id`, user classes, and Alpine `x-data`.
- One row per entry via `x-for`.
- One control per configured column.
- Text inputs for `ColumnText`.
- Native selects for `ColumnSelect`.
- Remove buttons per row when not disabled.
- An add-row button when not disabled.

New rows use each column's `Default` value. Select columns without an explicit
default use the first option's value. Text columns default to an empty string.

Disabled mode disables every visible control and hides add/remove actions, as
the old components do.

## Alpine Data

The existing components put JSON directly in the `x-data` attribute. The new
component should follow the repo's templ/Alpine escaping rule and avoid
`json.Marshal` for attribute-bound Alpine data. Build a small JavaScript object
literal with unquoted structural keys and single-quoted string values escaped for
JavaScript. This component only needs data, not functions, so no global
`Alpine.data()` registration is required.

Nil `Entries` must initialize as an empty array, not `null`.

## Submission Shape

Use an explicit structured submission shape:

```text
name[index][columnKey]=value
```

For example:

```text
labels[0][key]=app
labels[0][value]=web
taints[0][key]=node-role.kubernetes.io/control-plane
taints[0][value]=true
taints[0][effect]=NoSchedule
```

This replaces the old special-case hidden values:

- `keyvalue`: `name[key]=value`
- `triplet`: `name[key]=value:effect`

The new shape is more consistent for arbitrary structured rows and avoids
encoding multiple column values into one string.

## Docs And Navigation

Replace the two demo pages with one `Structured Input` page at
`/components/structured-input`.

The page should follow the Goshtoso component docs pattern:

- `demo.ComponentDemo` for the primary key/value metadata example.
- `demo.DemoSection` for a Kubernetes taints example with a select column.
- `demo.DemoSection` for an empty starter example.
- `demo.APIReference` outside `#structured-input-fragment`.
- Stable preview container IDs such as `structured-input-key-value`,
  `structured-input-taints`, and `structured-input-empty`.

Remove `Key Value` and `Triplet` from the component registry, sidebar, ordered
component nav, sidebar E2E expectations, and usage docs. Add `Structured Input`
in the form component group.

## Form Integration

Update `components/form` so `FieldGroupConfig` exposes:

```go
StructuredInput *structuredinput.Config
```

Remove the old `KeyValue` and `Triplet` fields and imports. Rendering precedence
should keep the new field near the other multi-value form components, after
`TagsList` and before `FileInput`.

## Testing

Add or update focused tests to cover:

- Rendering seeded key/value rows through `structuredinput`.
- Adding and removing rows.
- Rendering a select column and using the default select value for new rows.
- Empty starter behavior.
- Hidden input names and values for the structured submission shape.
- Form field integration renders `StructuredInput`.
- Sidebar includes `Structured Input` and no longer includes `Key Value` or
  `Triplet`.

Run `templ generate` after editing `.templ` files and `go run ./scripts/skillgen`
after changing component public types. Rebuild the site binary and run focused
E2E tests for the new component plus sidebar coverage.

## Migration Impact

Consumer code importing `components/keyvalue` or `components/triplet` must migrate
to `components/structuredinput`. Consumers relying on the old form submission
shape must update request parsing to handle indexed structured rows.

This break is acceptable because the requested change is to replace both public
APIs immediately.

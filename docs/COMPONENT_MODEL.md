# Working with Goshtoso Components

Every public component constructor returns a concrete renderable value. These
values implement `components.Component`, which embeds `templ.Component` and
adds a stable component identity:

```go
type Component interface {
    templ.Component
    Kind() Kind
}
```

You can render a Goshtoso component anywhere templ accepts a
`templ.Component`. Keep its concrete return value when application code needs
the component-specific type, or store it through `components.Component` when a
collection contains different components.

## Constructor styles

There is no shared configuration type. Each component package exposes the
fields or options that its constructor supports.

Components with structured data use package-specific config structs:

```go
badge.Badge(badge.Config{
    Label:      "Requires attention",
    Tone:       badge.ToneDanger,
    Appearance: badge.AppearanceSoft,
    Size:       badge.SizeSM,
})
```

Button, Link, Kbd, and Tooltip use functional options instead of config
structs. Link, Kbd, and Tooltip begin with their required values; Button accepts
options directly:

```go
button.Button(
    button.WithTone(button.TonePrimary),
    button.WithSize(button.SizeSmall),
    button.WithType("submit"),
)
```

Use the API reference on each component page for its exact constructor,
configuration fields, option functions, and rendered defaults.

## Concrete return values

Constructors return exported values such as `button.Instance`,
`modal.AlertDialogInstance`, and `table.Instance`. The values still satisfy
both common renderable interfaces:

```go
saveButton := button.Button(button.WithType("submit"))

var component components.Component = saveButton
var renderable templ.Component = saveButton
```

Concrete return types are useful when storing constructors as function values,
writing adapters, or preserving component-specific type information.

## Stable identity with Kind

Use `Kind()` when generic code needs to identify values from different
component packages:

```go
for _, component := range pageComponents {
    switch component.Kind() {
    case components.KindAlertDialog:
        // Alert-dialog orchestration.
    case components.KindTable:
        // Table orchestration.
    }
}
```

Kind values are stable kebab-case identifiers for registries, diagnostics,
tests, and switch statements. They are not CSS classes or HTML element names.
`components.AllKinds()` returns a copy of the complete public Kind list.

## Rendered defaults

A zero-valued field may render a documented default rather than its literal Go
zero value. Each component page records the effective values consumers see in
rendered output.

For installation and asset wiring, see the
[Consumer Integration Guide](USAGE.md).

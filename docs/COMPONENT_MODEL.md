# Goshtoso Component Model

Goshtoso separates application-wide styling, component identity, and
component configuration so each concept has one clear meaning. This guide
explains the vocabulary used by the Go API, component documentation, and
consumer agent references.

## Theme

A **theme** is the token set selected for a page or application. It controls
colors, typography, radii, and dark-mode behavior. `goshtoso` and `minimal` are
themes.

A theme is independent of component identity. Changing a page from the
Goshtoso theme to the Minimal theme does not turn a `Button` into another
primitive, and it does not change the button's `Kind`.

```html
<html data-theme="minimal">
```

## Primitive

A **primitive** is one meaningful public renderable concept with a semantic and
interaction contract. `Button`, `Modal`, `AlertDialog`, `Rating`, and
`RatingDisplay` are primitives.

Each public primitive has a constructor in its component package. A package may
provide more than one primitive when the concepts have distinct contracts.

## Kind

**Kind** is the stable identity returned by a concrete
`components.Component` value's `Kind()` method. It answers “what primitive is
this?” It does not describe styling, state, layout, or the current theme.

Kind values are useful when consumers keep different Goshtoso primitives in a
`[]components.Component` and need component-specific orchestration:

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

Kind string values are stable kebab-case identifiers for registries,
diagnostics, documentation tooling, tests, and switch statements. They are not
CSS classes or HTML element names.

## Configuration dimension

A **configuration dimension** is one independent choice within a primitive,
such as `Tone`, `Appearance`, `Size`, `Orientation`, `Position`, or `Mode`.
This is documentation vocabulary; Goshtoso does not expose a public
`Dimension` or `Axis` type.

For example, a badge remains the same primitive while its semantic tone,
visual treatment, and size change independently:

```go
badge.Badge(badge.Config{
    Label:      "Requires attention",
    Tone:       badge.ToneDanger,
    Appearance: badge.AppearanceSoft,
    Size:       badge.SizeSM,
})
```

The canonical names describe distinct choices:

- `Tone` is semantic intent or a palette role.
- `Appearance` is a visual treatment that preserves semantics and behavior.
- `Size` is physical scale.
- `Orientation` is horizontal or vertical direction.
- `Position` or `Side` is placement relative to another surface or the
  viewport.
- `Layout` is arrangement of the same content model.
- `Mode` is a behavior or presentation strategy when a more specific name is
  unavailable.
- `State` is validation or lifecycle state supplied by the consumer.

Each component package owns the typed values that make sense for its primitive.
For example, badge and alert tones may use the same words without being
assignable Go types.

## There is no universal Variant

`Variant` is too imprecise to be Goshtoso's canonical field or method. It can
conflate semantic tone, visual appearance, layout, behavior, and even distinct
component identities.

Goshtoso therefore has no universal `Variant` field and no common `Variant()`
method. Look for the specific configuration dimension that describes the
choice, or for another primitive when the semantic or interaction contract
changes.

## One primitive or two

Two presentations remain one primitive when their semantic role,
accessibility contract, required content, interaction and lifecycle behavior,
consumer-visible DOM responsibilities, and server integration contract stay
materially the same.

Style or configuration differences remain dimensions of one primitive:

- a drawer's left or right placement is `Position`;
- horizontal or vertical tabs use `Orientation`;
- a button's solid or outline treatment is `Appearance`;
- a table's plain or striped treatment is `Appearance`;
- numbered or simple pagination uses `Mode`.

Distinct semantics, DOM responsibilities, accessibility, or interaction
contracts become separate primitives:

- `Modal` and `AlertDialog`;
- `Toast` and `MessageToast`;
- `Carousel` and `CardCarousel`;
- `Banner` and `CookieBanner`;
- `Rating` and `RatingDisplay`.

This rule tells consumers whether to look for another configuration field or
another constructor.

## Effective defaults and atomic primitives

Documentation describes **effective defaults**: the behavior consumers see
after rendering. A Go zero value may trigger component defaulting rather than
appear literally in the output. For example, an empty size can resolve to the
primitive's normal size, and a zero row count can render a documented non-zero
number of rows.

Data-heavy components use configuration structs because their inputs are
naturally structured. Atomic primitives with a small required core and mostly
independent optional choices use functional options. `Button`, `Link`, `Kbd`,
and `Tooltip` follow that functional-options pattern, with their zero values
resolved through private defaults.

For public field naming conventions used by config-based components, see
[Component API Naming](COMPONENT_API_NAMING.md).

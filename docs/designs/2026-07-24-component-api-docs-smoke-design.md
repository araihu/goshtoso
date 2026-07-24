# Component API, Documentation, and Smoke-Test Design

**Status:** Design approved in conversation on 2026-07-24; written
specification pending review  
**Scope:** Goshtoso component library, documentation site, consumer agent
reference, and component smoke tests

## Context

Goshtoso is still alpha, so this pass may break public APIs when the existing
surface is ambiguous, non-idiomatic, unused, or impossible to document
truthfully. Compatibility aliases are not a goal. A smaller, deliberate public
surface is preferable to preserving accidental exports.

The audit found four related problems:

1. `Variant` currently means several unrelated things: semantic color, visual
   treatment, behavior, layout, and sometimes an entirely different component
   structure.
2. Rendered components have no shared runtime identity. Generic Go code can
   hold `templ.Component` values, but cannot reliably enumerate or switch on
   Goshtoso primitives.
3. Documentation tables are hand-written, flat, and incomplete. Main config
   fields, nested public config types, effective defaults, and public secondary
   primitives can drift independently.
4. The existing E2E suite proves many focused behaviors, but it does not
   mechanically prove that every registered component page loads directly and
   through HTMX navigation with working assets, examples, and API references.

At the audited `origin/main` commit (`1110624`), all existing unit, site, and
E2E tests pass. That green baseline does not contradict the findings: several
incorrect claims and undocumented fields are outside the assertions made by
those tests.

## Goals

- Give every intentional public renderable primitive a stable `Kind`.
- Use concrete configuration dimension names instead of a universal
  `Variant`.
- Define when two presentations belong to one primitive and when they are
  separate primitives.
- Curate the exported API before promising to support and document it.
- Document every supported public entry point and configuration type,
  including effective rendered defaults.
- Put the component-model rationale in consumer-facing source, site, and agent
  documentation.
- Make documentation completeness and all-page smoke behavior enforceable.

## Non-goals

- Preserve source compatibility with alpha APIs that are being corrected.
- Create a universal runtime schema for arbitrary component configuration.
- Make every styling choice a shared root-package enum.
- Generate prose or examples from Go source.
- Replace focused accessibility and interaction tests with broad smoke tests.
- Convert every component to functional options for uniformity.

## Consumer Vocabulary

The public documentation will use the following terms consistently.

### Theme

A theme is the application-wide visual system: tokens, typography, colors,
radii, and dark-mode behavior. `goshtoso` and `minimal` are themes. A theme does
not identify a component and is not a component configuration dimension.

### Primitive

A primitive is an intentional public renderable unit with one semantic and
interaction contract. `Button`, `Modal`, `AlertDialog`, `Toast`, and
`MessageToast` are examples.

### Kind

`Kind` is the stable runtime identity of a primitive. It answers "what
primitive is this?" It does not describe styling, state, or layout.

Examples:

```text
button
avatar
avatar-stack
modal
alert-dialog
toast
message-toast
toast-container
```

### Configuration dimension

A configuration dimension is one independent choice within a primitive.
"Dimension" is preferred in consumer documentation; "axis" is design-system
jargon and will not be part of the public Goshtoso API.

For example:

```go
badge.Config{
    Label:      "Requires attention",
    Tone:       badge.ToneDanger,
    Appearance: badge.AppearanceSoft,
    Size:       badge.SizeSmall,
}
```

`Tone`, `Appearance`, and `Size` can change independently. They are dimensions
of one badge, not themes and not different component identities.

Canonical dimension names are:

- `Tone`: semantic intent or palette role, such as neutral, primary, success,
  warning, or danger.
- `Appearance`: visual treatment that preserves semantics and behavior, such
  as solid, soft, outline, ghost, plain, or striped.
- `Size`: physical scale.
- `Orientation`: horizontal or vertical direction.
- `Position` or `Side`: placement relative to the viewport, trigger, or
  conversation.
- `Layout`: arrangement of the same content model.
- `Mode`: a behavior or presentation strategy when a more specific name is
  unavailable.
- `State`: validation or lifecycle state supplied by the consumer.

These are naming categories, not shared root enums. Each package owns the typed
values that make sense for its primitive. A button tone and an alert tone may
share words without being assignable Go types.

### Why there is no universal Variant

`Variant` is acceptable as an informal umbrella in design-system discussion,
but it is too imprecise as Goshtoso's canonical public field or method. A
universal `Variant()` would force unrelated values such as danger, striped,
simple pagination, and card carousel into one vocabulary.

Goshtoso therefore has no common `Variant()` method. Existing `Variant` fields
will be renamed according to their actual dimension or removed when they
represent a different primitive.

## One Primitive or Two

Two presentations remain one primitive only when all of these stay materially
the same:

- semantic role and accessibility contract;
- required data/content model;
- interaction and lifecycle behavior;
- DOM responsibilities exposed to the consumer;
- server integration contract.

If only an independent choice changes, it is a configuration dimension:

- drawer left/right is `Position`;
- tabs horizontal/vertical is `Orientation`;
- button solid/outline is `Appearance`;
- table plain/striped is `Appearance`;
- pagination numbered/simple is `Mode`.

If semantics or the interaction/content contract changes, it is a separate
primitive:

- `Modal` and `AlertDialog`;
- `Toast` and `MessageToast`;
- `Carousel` and `CardCarousel`.

The plain carousel may infer an overlay and CTA from optional slide content;
that remains the same slide and navigation contract. A card carousel owns an
article/card wrapper and a separate body, so it is a different primitive.

This rule is part of the public component-model documentation so consumers can
predict whether to look for another field or another constructor.

## Runtime Component Identity

The root `components` package will define:

```go
package components

type Kind string

type Component interface {
    templ.Component
    Kind() Kind
}

func AllKinds() []Kind
```

`Kind` constants use stable kebab-case string values. `AllKinds` returns a copy
in stable order. String values are intended for registries, diagnostics,
documentation tooling, tests, and switch statements; they are not CSS classes
or HTML element names.

Every intentional public renderable constructor returns a value implementing
`components.Component`. Each package uses a concrete instance type so callers
are not forced through one opaque wrapper and Go type assertions remain useful.
The generated templ implementation becomes private behind that value.

Conceptually:

```go
type Instance struct {
    // private rendering inputs
}

func Button(cfg Config) Instance

func (Instance) Kind() components.Kind
func (Instance) Render(context.Context, io.Writer) error
```

Packages with multiple primitives use distinct concrete types, such as
`avatar.Instance` and `avatar.StackInstance`. Internal fragments do not receive
a `Kind`.

### Public renderable curation

An exported renderable is public only when a consumer is expected to invoke it
directly and its contract can be documented and tested independently.

Examples that remain public include primary component constructors and
deliberate secondary primitives such as `AvatarStack`, `CheckboxGroup`,
`ToastContainer`, and form sections.

Implementation fragments such as combobox OOB internals and table head/body/row
render fragments become private unless a documented server-handler workflow
requires direct consumer use. Convenience renderers that duplicate another
component, such as a card-local star rating, are removed or replaced by the
canonical primitive.

The implementation plan will contain the complete entry-point inventory and
its keep/private/remove decision before edits begin.

## Configuration API Policy

### Config structs

Config structs remain the default for data-heavy or naturally structured
components, including tables, forms, sidebars, comboboxes, carousels, and
schema-driven forms. Their nested public types are part of the supported API
and must be documented.

### Functional options

Functional options are used selectively when:

- the primitive has a small, obvious required core;
- most consumers rely on defaults;
- the current config has many independent optional fields;
- a caller commonly changes only one or two properties.

Options are not introduced merely to make APIs look uniform. Initial
candidates are small atomic primitives such as buttons and spinners; the
implementation inventory decides each case against the criteria above.
Options use typed `With...` functions and are applied to an unexported default
configuration. Data-heavy components do not receive dozens of options that
mirror every struct field.

### Exported helper curation

CSS-class assemblers, resolved-ID helpers, defaulting helpers, and render-only
predicates become private unless consumers have a documented non-rendering use.
Intentional behavior helpers such as validation, URL construction, state
initialization, initials derivation, or schema transformation may remain
public.

No-op and contradictory APIs are removed. Known examples include:

- accordion `SingleOpen`, which duplicates `AllowMultiple == false`;
- table `WithCheckbox`, which does not drive checkbox rendering;
- card `Price`, `Rating`, `HasPrice`, and `HasRating`, which are not rendered;
- combobox option presentation fields that are not consumed by the renderer.

If a field is retained, its visible or behavioral effect must be demonstrated
and tested.

## Dimension Migration

The implementation uses this classification:

| Existing component | New model |
| --- | --- |
| Accordion | `Appearance`; remove `SingleOpen` |
| Alert | `Tone` |
| Avatar | `Tone` |
| Badge | `Tone` plus `Appearance` for solid/soft |
| Banner | `Tone` |
| Button | `Tone`; introduce `Appearance` only when a distinct treatment exists |
| Card | `Emphasis`; keep `Layout` |
| Carousel | remove `Variant`; content determines overlay, `CardCarousel` is separate |
| Checkbox | `Tone` |
| FileInput | `Appearance` |
| Modal | split `Modal` and `AlertDialog`; alert dialog owns `Tone` |
| Pagination | `Mode` |
| Radio | `Tone` |
| Spinner | `Tone` |
| Table | `Appearance`; checkbox selection remains explicit behavior |
| Toast | `Tone`; split `MessageToast` |
| Toggle | `Tone` |

The implementation plan must inspect rendered behavior before fixing final
constant names. A more precise word than the table requires an explicit reason
in the plan; it may not fall back to an ambiguous catch-all `Variant`.

## Documentation Architecture

### Consumer component-model guide

The rationale above is published in all consumer-facing channels:

1. `docs/COMPONENT_MODEL.md` in the library source.
2. A navigable site page at `/docs/component-model`.
3. A concise section and link in Getting Started.
4. The generated `using-goshtoso` consumer skill/reference.
5. A link from `docs/COMPONENT_API_NAMING.md`, which remains the maintainer
   naming guide.

The guide includes the theme/kind/dimension example, the one-primitive test,
and an idiomatic Go loop/switch example:

```go
for _, component := range pageComponents {
    switch component.Kind() {
    case components.KindAlertDialog:
        // Component-specific orchestration.
    case components.KindTable:
        // Component-specific orchestration.
    }
}
```

### Structured API references

The site replaces the one-flat-table assumption with reusable API sections.
Each component page declares:

- public constructor and `Kind`;
- primary config or functional options;
- every supported nested public config/data type;
- every supported secondary primitive;
- fields/options with exact Go type;
- whether a value is required;
- effective rendered default;
- behavioral description and constraints.

`Default` always means effective consumer-visible behavior, not the Go zero
value. For example, a zero `Rows` value that renders three textarea rows is
documented as `3`.

API metadata lives in named package-level values or functions that tests can
inspect. It is not buried only inside generated templ fragments.

Reflection verifies struct-field coverage and type spelling. Prose, examples,
and effective defaults remain authored because source reflection cannot explain
behavior.

### Content corrections

The pass corrects every claim against source and smoke behavior, including
known drift:

- missing `avatar.Config.SrcExpr`;
- missing `textarea.Config.InputAttrs`;
- missing `toggle.Config.Value` and `toggle.Config.InputAttrs`;
- wrong or raw-zero defaults for textarea, select, tags list, sidebar,
  tooltip, and form;
- nested config types that are public but absent from API references;
- stale registry/SEO claims about file-input validation, spinner labels,
  structured-input semantics, textarea counters, tag duplicate handling,
  range helper text, toast positions, and the table checkbox variant.

Descriptions are owned by the same registry used for navigation and smoke
tests so route titles, search text, metadata, and test enumeration do not drift
through independent lists.

## Contract and Smoke Testing

### Compile-time identity contracts

Each public constructor has an assertion that its return type implements
`components.Component`. Tests verify that every constant from `AllKinds` is
unique, non-empty, documented exactly once, and mapped to a component page.

### Documentation contract tests

For every registered component page, tests verify:

- every exported field of each registered public struct appears exactly once;
- documented field names and Go types match reflection;
- descriptions and effective-default cells are non-empty where applicable;
- every public primitive and nested config type is included;
- no removed/private symbol remains in docs or snippets;
- every component route is represented in navigation and search metadata.

Reflection is a guardrail, not the source of truth for defaults. Focused
render tests prove non-zero effective defaults and conditional behavior.

### Browser smoke tests

A registry-driven direct-load test visits every component route and asserts:

- successful HTTP response;
- expected title and `h1`;
- non-empty component description;
- at least one marked preview and matching code example;
- structured API reference;
- no `pageerror`, error-level console event, failed request, or failed local
  asset.

A second registry-driven test navigates through every component route using the
actual HTMX/sidebar path and asserts:

- URL and heading update;
- API reference and preview are swapped in;
- active navigation is correct;
- Alpine/HTMX remain initialized;
- no browser errors or failed assets accumulate.

Stable `data-demo-section`, `data-demo-preview`, and API-section hooks support
these assertions without coupling tests to presentation classes.

Broad smoke tests establish catalog integrity. Existing focused tests remain
responsible for keyboard behavior, ARIA state, form submission, HTMX swaps,
Alpine state, sorting, pagination, and other component-specific contracts.
When documentation makes a claim not covered by an existing test, this pass
adds a focused render or E2E assertion.

### Themes

The catalog smoke runs under the default theme. A smaller representative matrix
runs the shared component model under dark mode and the Minimal theme to catch
token and visibility failures without multiplying all 42-page navigation
loops. Components changed in ways that affect theme classes receive focused
light, dark, and Minimal checks.

## Delivery Sequence

1. Inventory and classify every exported renderable, helper, config type, and
   current `Variant`.
2. Add the root identity contract and concrete component instance pattern.
3. Curate and rename public APIs, then update all internal consumers.
4. Publish the consumer component-model guide and extend structured API
   metadata.
5. Correct and enrich all component pages and generated consumer-agent
   references.
6. Add identity, documentation, registry, and browser smoke contracts.
7. Regenerate templ output, CSS when needed, and the consumer skill reference.
8. Run focused tests, then complete root/site unit tests, lint/fix/build gates,
   and the full E2E suite.

## Completion Evidence

The work is complete only when current repository evidence proves all of the
following:

- every intentional public renderable implements `components.Component`;
- `AllKinds`, the public-renderable inventory, documentation mapping,
  navigation, and component routes agree;
- no ambiguous component `Variant` remains from the migration matrix;
- every supported public config field and nested type is documented;
- every documented default and behavioral claim has source/render evidence;
- all component pages pass direct-load and HTMX-navigation smoke tests;
- focused interaction tests and the full existing E2E suite pass;
- `templ generate`, skill generation, formatting, lint, build, root tests, site
  tests, and `git diff --check` pass.

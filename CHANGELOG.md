# Changelog

All notable changes to Goshtoso are documented in this file.

## [Unreleased]

### Breaking component API changes

This is an intentionally **Breaking** alpha release line. The exact migration
base is `v0.0.11`
(`10b4dcbf3da3c1dd534d8d2baa949d043b9d0f1f`). No next version or release date
has been assigned.

Read the complete [component API migration guide](docs/MIGRATING_COMPONENT_API.md)
before upgrading.

#### Source-breaking

- Added the root `components.Component` / `Kind` identity contract and a stable
  74-entry `AllKinds()` registry. Sixty-seven same-name public constructors now
  return exported concrete instance types instead of `templ.Component`.
  Seven new or renamed split constructors complete the 74-constructor
  inventory.
- Replaced ambiguous component `Variant` and `Style` APIs with package-owned
  dimensions such as `Tone`, `Appearance`, and `Mode`. There is no universal
  `Variant` API and there are no compatibility aliases.
- Split primitives whose semantics or interaction contracts differ:
  `Banner` / `CookieBanner`, `Carousel` / `CardCarousel`, `Modal` /
  `AlertDialog`, `Rating` / `RatingDisplay`, and `Toast` / `MessageToast`,
  including their OOB toast forms.
- Replaced the exported Button, Link, Kbd, and Tooltip config constructors with
  required core arguments plus typed functional options.
- Removed or privatized accidental render fragments, mutable defaults,
  CSS-class assemblers, resolved-ID/default helpers, and render-only
  predicates. Notable removals include
  `accordion.AccordionItemData`, Card's inert price/rating API and
  `StarRating`, Combobox's six public fragments and five inert option
  presentation fields, Table's duplicate action/status helpers and initial
  head/body/pagination fragments, and `toast.Container`.

#### Behavior and effective defaults

- Carousel overlays are inferred from slide content; card carousel behavior is
  a separate primitive.
- Alert dialogs now own `role="alertdialog"`, semantic Tone, and their
  single-action contract. Read-only rating output now has its own non-form
  primitive and accessible-label behavior.
- Toast events distinguish semantic and message kinds; message actions render
  only when configured. Custom Tooltip triggers preserve consumer state and
  wire accessibility to the actual focus target. Search navigation targets are
  revalidated at the client-side sink.
- Constructor defaults are explicit for the four functional-option APIs.
  Several default Appearance/Mode constants now use the empty string while
  preserving their effective rendered treatment; consumers should compare
  typed constants rather than raw strings.

### Documentation and verification

- Published the consumer component model across source docs, the demo site, and
  the generated `using-goshtoso` reference.
- Replaced flat hand-written API tables with source-checked structured
  references covering 42 component pages, 74 Kinds, every supported public
  constructor/configuration type, and effective rendered defaults.
- Added registry-driven direct-load and HTMX-navigation smoke coverage for the
  complete component catalog plus a representative light/dark Goshtoso/Minimal
  theme matrix.

[Unreleased]: https://github.com/araihu/goshtoso/compare/v0.0.11...main

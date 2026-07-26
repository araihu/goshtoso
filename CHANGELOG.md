# Changelog

All notable changes to Goshtoso are documented in this file.

## [v0.0.13] - 2026-07-26

### Resilient dependency loading

- Refactored `head.Dependencies` and `head.DependenciesMinimal` to accept typed
  functional options while preserving zero-argument and zero-value use.
- Made exact-version unpkg URLs the default third-party sources, with ordered
  automatic fallback to the matching embedded assets when a browser cannot
  download a primary resource.
- Added generated SHA-384 Subresource Integrity for CDN and local runtime bytes,
  CSP nonce propagation, a public readiness promise, and fallback/ready/error
  browser events.
- Added Alpine Mask to the full runtime set so public components that emit
  `x-mask` work with the default head helper.
- Added `WithLocalRuntime` for offline PWAs, desktop/mobile WebViews, air-gapped
  deployments, and explicit no-network policies. Per-dependency URL, integrity,
  omission, fallback, stylesheet, loader, and combobox controls remain
  available for application-owned infrastructure.

### Documentation and verification

- Updated the consumer guide, README, component demo, installable agent skill,
  and generated component reference with the dependency modes and migration
  guidance.
- Added a durable head dependency audit and a real-browser fixture that forces
  every primary request to fail, then verifies Alpine Collapse, Focus, Mask,
  HTMX, and combobox behavior through the embedded fallbacks.

### Upgrade note

The default now makes version-pinned CDN requests before using local fallback.
Applications that must make no external runtime request should pass
`head.WithLocalRuntime()`. Custom runtime bytes should supply matching CDN and
local URLs plus `head.WithDependencyIntegrity`; an empty integrity string
explicitly disables SRI for that dependency.

## [v0.0.12] - 2026-07-25

### Breaking component API changes

This is an intentionally **Breaking** alpha release. The exact migration base
is `v0.0.11` (`10b4dcbf3da3c1dd534d8d2baa949d043b9d0f1f`).

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

[v0.0.12]: https://github.com/araihu/goshtoso/compare/v0.0.11...v0.0.12

# Changelog

All notable changes to Goshtoso are documented in this file.

## [v0.1.11] - 2026-08-09

### Arbitrary consumer icon packs

- Extended `iconpack` beyond the Arai Hu Assets catalog: consumers can bring
  any SVG icon pack, or a single icon, through a hashed JSON or YAML source
  manifest and a source root or verified archive.
- Added deterministic standalone SVG normalization, typed consumer-local Go
  bindings, sprite output, provenance, licenses, and exact source hashes while
  preserving the existing `components/icon` rendering and accessibility
  contract.
- Added a checked-in Bootstrap Icons example and browser proof in the
  Goshtoso site, plus the `/docs/iconpack` usage guide for both source modes.

### Upgrade note

- Existing embedded Heroicons and the Arai Hu Assets release mode remain
  compatible. Consumers may adopt arbitrary packs without modifying the
  embedded Goshtoso icon package.

## [v0.1.9] - 2026-08-08

### Consumer-local icon packs

- Added consumer-local `iconpack` generation from a verified Arai Hu Assets
  release root or archive, preserving the catalog's canonical names and sprite
  symbols while generating a parallel Go package.
- Added generated `Icon`, `Lookup`, `Name`, `Glyph`, `Config`, and `SpriteURL`
  usage through the core `components/icon` API, with licenses, provenance, and
  manifest outputs kept beside the generated package.
- Added the Goshtoso `/docs/iconpack` guide, Icon workbench examples, and
  focused external-consumer/browser proof for generated icons and their sprite.

### Upgrade note

- Existing bundled Heroicons and the core `components/icon` API remain
  unchanged. Consumers should use the generator's exact release hashes and
  serve the generated sprite from the configured same-origin URL.

## [v0.1.8] - 2026-08-07

### Scroll region boundary cues

- Added `components/scrollregion` for bounded, independently scrollable content
  with semantic top and bottom boundary indicators.
- Added HTMX-aware runtime initialization and cleanup for scroll, resize, and
  dynamic-content changes without intercepting pointer input.
- Published the generated JavaScript, Tailwind CSS, and consumer-agent API
  references required by downstream renderers.

### Upgrade note

- Existing components are unchanged. Consumers can render
  `scrollregion.ScrollRegion(scrollregion.Config{...})` wherever a bounded
  vertical viewport needs non-interactive overflow cues.
- Stable component identity and the public catalog page follow in the site pin
  update after this root package is reachable as `v0.1.8`.

## [v0.1.7] - 2026-08-02

### Reduced-motion pressed cards

- Corrected `card.InteractionPressed` to transition the Tailwind v4 individual
  `translate` property, so its pressed movement animates as designed.
- Neutralized hover and active translation when `prefers-reduced-motion` is
  enabled, keeping pressed cards spatially still for reduced-motion users.

### Upgrade note

- No API changes are required. Consumers using `InteractionPressed` should
  update their Goshtoso stylesheet and dependency pin together.

## [v0.1.6] - 2026-08-02

### Expressive project cards and directional drawers

- Added `card.Config.Media` and `MediaClass` so consumers can supply arbitrary
  templ media while retaining Card structure, semantics, and content styling.
- Added the opt-in `card.InteractionPressed` treatment for linked or clickable
  cards, including reduced-motion-safe transform and transition behavior.
- Added `drawer.SideTop` and `drawer.SideBottom` with height presets from small
  through full-screen, while preserving existing left/right defaults.
- Added component demos, focused rendering tests, generated CSS, and updated
  consumer-agent references for all new configuration choices.

### Upgrade note

- Existing Card and Drawer configurations remain source compatible. Consumers
  opt into custom media, pressed interaction, or vertical drawer directions.

## [v0.1.5] - 2026-08-02

### Inline code

- Added `components/inlinecode`, a semantic, theme-aware primitive for short
  code fragments inside prose and documentation.
- Added consumer hooks through `WithRootClass` and `WithRootAttrs`, stable
  component identity, generated skill reference, usage guidance, and tests.
- Added `codeblock.DensityCompact` for short install and command snippets that
  need tighter header and code-body spacing without consumer CSS overrides.
- Preserved standalone site deployability with a version-aware catalog bridge;
  the public demo page follows after the site module pins this release.

## [v0.1.4] - 2026-08-02

### Muamba runtime acquisition

- Replaced the private vendoring downloader and compatibility JSON views with
  Muamba v0.0.2 SHA-384 locks plus a Goshtoso-owned metadata overlay.
- Added `assets.MuambaResources`, `assets.MuambaHash`, and
  `assets.RuntimeHash` for typed acquisition inspection and cache busting while
  preserving the original nine-field `RuntimeAsset` layout.
- Upgraded standalone Tailwind CSS to v4.3.3 with target-specific Muamba locks
  and retained its MIT license alongside runtime licenses and provenance.
- Kept runtime URLs, versions, ordering, enablement, SRI, attributions, and
  CDN-first/local-fallback behavior compatible.

## [v0.1.3] - 2026-08-02

### Canonical embedded runtime manifest

- Made the legacy JSON runtime inventory the ordered source of truth for
  embedded JavaScript pins, CDN URLs, SRI, local paths, loader defaults,
  attribution data, package provenance, and retained license notices.
- Added `head.WithRuntimeManifest` for typed ownership of dependency URLs,
  integrity, enablement, minimal-mode membership, and safe execution order.
  Existing per-dependency options remain supported and apply after the custom
  manifest snapshot.
- Added generated, caller-owned `assets.DefaultRuntimeMetadata()` for runtime
  identity and licensing. `assets.RuntimeAsset` remains the supported loading
  and override contract, with its original nine-field layout preserved for
  source compatibility, including positional literals.
- Added deterministic transactional generation/downloads, duplicate-module
  rejection, exact package-version provenance checks, and remote CDN/license
  hash verification. Configuration freedom does not imply arbitrary-version
  compatibility; the manifest pins remain the tested combination.
- Made the demo site select and order local runtime assets from the Goshtoso
  module linked into its binary, with explicit site-only enablement and
  v0.1.0-compatible fallbacks for optional roles.

### Upgrade note

- No migration is required for existing `head.Dependencies` or
  `head.DependenciesMinimal` consumers. Default full, minimal, CDN-first, and
  local-only rendering remain compatible. The manifest pins are the tested
  runtime combination; overriding versions configures loading but does not
  guarantee compatibility with arbitrary combinations.

### Release verification

- Updated Playwright Go to v0.6100.0 and its current
  `github.com/mxschmitt/playwright-go` module path after the retired v1.57.0
  driver archive made clean release-runner installation fail.

## [v0.1.2] - 2026-07-30

### Automated Arai Hû fallback assets

- Added a transactional updater for the bundled Arai Hû theme, mark, wordmark,
  favicon, and release manifest.
- Added authenticated release-dispatch enrollment so new Assets releases can
  propose verified fallback updates without personal access tokens.
- Preserved local bundled files as the server-rendered and no-JavaScript
  fallback while the optional presentation channel manages live branding.
- Documented archive-extraction and crash-durability hardening as follow-up
  technical debt.

## [v0.1.0] - 2026-07-29

### Accessible sprite icons and typed catalog

- Added `components/icon`, a brand-neutral accessible SVG sprite component with
  safe same-origin external references, inline-document mode, fixed sizes, and
  explicit labelled versus decorative rendering.
- Bundled the immutable 67-symbol Heroicons v2.2.0 UI sprite, MIT notice,
  typed `heroicons.Icon...` constants, enumerable glyph metadata, and the
  default same-origin `heroicons.SpriteURL` of `/assets/icons/heroicons.svg`.
- Added `iconcatalog`, a schema-v1 catalog generator for project-local typed
  bindings. It validates catalog structure, selected SVG sprite assets,
  monochrome/tintable color behavior, duplicate symbols, and Go identifier
  collisions before generating formatted source.
- Added the responsive icon showcase, configuration modal, copyable Go example,
  asset serving checks, generated-file drift checks, and real-browser coverage
  for every bundled symbol.

### Release-candidate compatibility evidence

- Prepared against Arai Hû Assets P1 source `246cb28`, integrated release
  candidate `80d43a3`; catalog schema is `1`.
- Recorded immutable upstream Arai Hû Assets `dist/catalog.json`: 302 records,
  schema `1`, SHA-256
  `d83be964fa411e87c61b49f0a0b6a2a1465f33ad43bea7cd93b2e434b59266af`.
- Recorded exact local 67-record Heroicons generator subset: namespace `ui`,
  product `heroicons`, and only objects from that immutable upstream catalog;
  SHA-256
  `0a420ad65e2fe7db3e2cc5dbb6c87167fcd6e85f64a3ebc409e2a58c9bd111ef`.
- Recorded immutable UI sprite SHA-256
  `75e282de7a19efba9cf0285b44af0641c1527361f921b7d7f8020efc1f1f0fb7`.
- Generated bundled bindings from that local subset with
  `go run ./cmd/iconcatalog -catalog internal/iconcatalog/testdata/heroicons-catalog.json -namespace ui -product heroicons -sprite-url /assets/icons/heroicons.svg -package heroicons -const-prefix Icon -out components/icon/heroicons/names_gen.go`.
- Release evidence covers approved Goshtoso functional head `ab05821`; tagging
  and downstream dependency pins remain deliberately deferred.

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

[v0.1.9]: https://github.com/araihu/goshtoso/compare/v0.1.8...v0.1.9
[v0.1.3]: https://github.com/araihu/goshtoso/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/araihu/goshtoso/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/araihu/goshtoso/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/araihu/goshtoso/compare/v0.0.13...v0.1.0
[v0.0.13]: https://github.com/araihu/goshtoso/compare/v0.0.12...v0.0.13
[v0.0.12]: https://github.com/araihu/goshtoso/compare/v0.0.11...v0.0.12

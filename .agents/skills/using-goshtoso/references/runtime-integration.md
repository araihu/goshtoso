# Runtime Integration

Read this reference only when changing Goshtoso's default runtime delivery,
offline inventory, cache identity, or Content Security Policy integration.

## Runtime choices

- Keep `head.Dependencies()` for resilient CDN-first loading with exact embedded
  fallback bytes.
- Use `head.Dependencies(head.WithLocalRuntime())` for offline, air-gapped, or
  explicit no-CDN policy. Do not select it merely to simplify a probe.
- Override primary/local URL pairs together and supply
  `WithDependencyIntegrity` when bytes change. Primary and fallback bytes share
  one integrity value and must match.
- Use `WithRuntimeManifest(assets.DefaultRuntimeManifest())` only when the app
  must enable, omit, add, or reorder the full typed dependency set.
- Use `WithoutLocalFallback()` only when failure is preferable to local retry.
- Await `window.goshtosoDependencies.ready` before app JavaScript requiring the
  complete dynamically loaded runtime.

## Manifest and inventory

Use `assets.DefaultRuntimeManifest()` instead of copying generated versioned
paths. Its caller-owned dependency slice preserves execution order, roles, CDN
primary, Handler-served local URL, SRI, enabled state, minimal-set membership,
defer, and readiness semantics. Cache stylesheet and loader URLs separately.
Execute either the loader or direct local scripts, never both.

Custom roles must remain unique and safe. Preserve Alpine plugin and first-party
ordering before Alpine. Load HTMX, then `hx-alpine-compat`, then Alpine core;
keep SSE/WS after HTMX. The default manifest already supplies this order. For custom local-only loading,
set each `PrimaryURL` to its `LocalURL`, keep the loader local, and use
`WithoutLocalFallback()`. `WithLocalRuntime` applies only to the default
manifest. Invalid manifests fail before emitting HTML.

Goshtoso guarantees its pinned runtime combination, not arbitrary overrides.
Bind same-version caches only when `assets.GoshtosoVersion().Status` is
`assets.VersionExact`; development, replacement, and unavailable builds do not
identify exact Goshtoso bytes.

## Alpine and fragment lifecycle

The bundled `hx-alpine-compat` extension coordinates Alpine with HTMX settling
and preserves reactive scope during `innerMorph` / `outerMorph` swaps. Register
app providers before Alpine starts. Do not add a second Alpine runtime,
`@alpinejs/morph`, or unconditional `Alpine.initTree()` after each swap.
Replacement swaps still destroy the replaced state. Keep stable IDs for morphs
and test edited fields, focus, reactive IDs, and repeated navigation.

See [the release migration reference](migration-v0.3.0.md) for request/event
changes and the App Shells dependency boundary.

## CSP

Test the rendered app under its exact policy. Standard Alpine requires dynamic
function evaluation and component state mutates inline styles. Local files alone
do not make `script-src 'self'` sufficient. Allow required Alpine behavior or
replace the default runtime with a verified CSP-compatible stack.
`templ.WithNonce` propagates to the loader and child scripts for nonce and
`strict-dynamic` policies. Do not weaken unrelated directives.

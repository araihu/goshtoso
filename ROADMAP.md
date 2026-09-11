# Roadmap and Stability

Goshtoso is alpha-stage software. The components are already usable in real Go
applications, but the public API is still allowed to change while the project
moves toward a stable release line.

## Alpha Stability Policy

- Tags use `v0.x.y` while the API is still being refined.
- Only the latest `v0.x.y` tag is supported during alpha. See
  [SECURITY.md](SECURITY.md) for the security support policy.
- Breaking component API changes may ship in `v0.x.y`, but they should be
  intentional, documented in the release notes, and reflected in demos, tests,
  and generated references.
- Generated files are part of the published module. Consumers should not need to
  run `templ generate` for Goshtoso's own components.
- Runtime JavaScript and CSS should stay locally bundled and versioned through
  the asset pipeline, not CDN-dependent at page load time.

## Path to v1.0

The `v0.3.x` line establishes HTMX 4 and Alpine 3.17 as the runtime baseline.
The first stable `v1.x` line should provide consistent public APIs and predictable
upgrade guidance.

Before `v1.0.0`, the project should have:

- Stable naming conventions across public component config fields.
- Current component demos, API tables, and E2E coverage for the supported
  component catalog.
- A release process that keeps `README.md`, `VERSIONS.md`, release notes, and
  the `site/` module pin in sync.
- Clear consumer docs for bundled assets, custom Tailwind builds, HTMX, Alpine,
  theming, and common pitfalls.
- CI coverage for generated drift, linting, root/site tests, E2E behavior, and
  documentation links.

## Near-Term Focus

- Keep public docs synchronized with the generated component catalog.
- Tighten component API consistency as new components and variants land.
- Improve examples that show complete server-rendered workflows, not only
  isolated component previews.
- Reduce contributor friction around generated assets, release hygiene, and
  documentation drift.

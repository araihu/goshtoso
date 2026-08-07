# Roadmap and Stability

Goshtoso is alpha-stage software. The components are already usable in real Go
applications, but the public API is still allowed to change while the project
moves toward a stable release line.

## Pre-1.0 stability policy

- Tags use semver pre-1.0 versions while the API is still being refined.
- Only the latest supported release line receives routine fixes. See
  [SECURITY.md](SECURITY.md) for the security support policy.
- Breaking component API changes must be intentional, documented in the release
  notes, and reflected in demos, tests, and generated references.
- Generated files are part of the published module. Consumers should not need to
  run `templ generate` for Goshtoso's own components.
- Runtime JavaScript and CSS should stay locally bundled and versioned through
  the asset pipeline, not CDN-dependent at page load time.

## Release-line criteria

Each new release line should keep the core component APIs coherent enough for
early adopters to upgrade with normal release-note guidance. Before a release,
the project should have:

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

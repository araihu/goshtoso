---
name: component-authoring
description: Use when creating, editing, reviewing, or hardening Goshtoso component packages under components/<name>/, especially public Config structs, types.go helpers, .templ entry points, HTMX/Alpine attributes, generated IDs, validation, URL builders, JavaScript expressions, or component tests.
---

# Goshtoso Component Authoring

Use this before changing component source under `components/`. For demo/docs
pages, also use `component-docs`. For consumer examples, also use
`using-goshtoso`. For interactive markup details, use `htmx`, `alpinejs`, and
`templ` as needed.

## Interface Rules

- Keep simple presentational components zero-value usable.
- Add `Validate() error` when a component has required IDs/endpoints,
  mutually exclusive modes, server/client modes, dependency fields, or invariants
  that cannot be safely inferred in the template.
- Prefer typed constants for variants, sizes, modes, positions, and states.
- Keep `Config` structs plain and documented; use value receivers for pure
  helpers unless mutation is intentional.
- Keep `Class` additive. It may extend root classes, but must not replace
  required layout, accessibility, focus, or dark-mode classes.
- Use `templ.Attributes` for caller escape hatches when arbitrary HTML, HTMX, or
  Alpine attributes are expected.

## Safety Rules

- Never build URLs with string concatenation. Use `net/url`, preserve existing
  query params, and encode generated params.
- Never interpolate caller-provided values into JavaScript without a helper for
  JavaScript string literals or identifiers.
- Avoid package-level mutable render state. Render output must be safe under
  concurrent HTTP requests.
- Do not place complex Alpine objects with quoted strings in HTML attributes.
  Register complex behavior through `<script>` + `Alpine.data()` and `templ.Raw`.
- Do not use `json.Marshal` for data that lands inside an HTML attribute.
- Keep htmx first: server-rendered fragments and `hx-*` before client-side state
  when the interaction can round-trip cleanly.

## Testing Rules

- Add helper tests for URL building, validation, escaping, generated IDs, and
  generated JavaScript expressions.
- Add render tests when template output semantics change.
- Test zero-value behavior and at least one configured behavior for new
  components.
- For Alpine/HTMX swaps or browser-only behavior, add or update E2E coverage in
  `site/tests/e2e`.

## Required Commands

After `.templ` edits:

```bash
templ generate
```

After component API or entry-point edits:

```bash
go run ./scripts/skillgen
```

Before finishing:

```bash
go test ./components/...
```

Run narrower package tests first while developing, then the full component
package baseline.

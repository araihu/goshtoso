# Component Interface Hardening Design

## Goal

Harden the highest-risk Goshtoso component interfaces without starting a broad
API migration, then capture the rules in a reusable component-authoring skill so
new and modified components follow the safer shape by default.

## Scope

This first pass covers the components and interfaces with known safety,
concurrency, or semantic sharp edges:

- `components/codeblock`: remove package-level render state and improve repeated
  Copy button accessibility.
- `components/button`: make `LoadingText` render as user-visible text instead of
  being appended as a CSS class.
- `components/table`: replace manual query-string construction with `net/url`
  helpers and remove or safely encode inline JavaScript string interpolation for
  row links.
- `components/modal` and `components/drawer`: audit ID-derived Alpine state and
  event expressions, then add focused helper coverage for safe generated names.
- `.agents/skills/component-authoring` and `.claude/skills/component-authoring`:
  add concise rules for future component interface design.

The pass intentionally does not add validation to every component package. That
belongs in later, smaller migrations after this first safety layer lands.

## Architecture

The code hardening stays local to each component. It should preserve existing
call sites unless a current field is demonstrably broken. Public fields remain
plain Go config structs because that is the library's current idiom.

Shared helper extraction should be minimal. If one component needs a URL builder
or JavaScript string helper, place it in that component package first. Promote to
a shared internal helper only if two or more component packages need identical
behavior during this pass.

## Component Changes

### CodeBlock

`Config.GetID()` currently increments a package-level counter during render. This
is not safe for concurrent HTTP rendering and makes default output depend on
render order. The hardened behavior should derive a stable ID without mutable
package state:

- If `Config.ID` is set, use it unchanged.
- If `Config.ID` is empty, derive an ID from stable config data such as language,
  label, and code content.
- Keep the ID deterministic within a render and safe across concurrent renders.
- Add a test that concurrent calls do not race or share mutable state.

The Copy button should also receive an accessible name that distinguishes
repeated code blocks, using label/language or the resolved code element ID.

### Button

`LoadingText` should mean text. The template should render a hidden indicator
span containing `cfg.LoadingText` while preserving the existing child content
span. If a visual spinner class is needed later, it should be a separately named
field, not overloaded into `LoadingText`.

Add a render test that proves `LoadingText: "Saving..."` appears as text and is
not emitted as a class name.

### Table

Manual URL construction should be replaced with `net/url` helpers for
`PaginationBaseURL`, `SortURL`, `PageURL`, and `NextPageURL`.

Rules:

- Preserve existing query parameters in `HTMXEndpoint`.
- Encode generated parameters with `url.Values`.
- Preserve `ExtraQueryParams` compatibility for existing callers, but parse and
  merge it rather than append raw text where practical.
- Keep the sort-reset behavior: cycling to `SortNone` omits sort params.

Row link behavior should avoid inline JavaScript with raw `row.Link` in a string
literal. Prefer declarative HTMX/plain-link attributes. If inline JavaScript is
still needed for middle-click or full navigation, encode the URL with a helper
that produces a valid JavaScript single-quoted string.

Add tests for:

- Existing endpoint query params are preserved.
- Sort keys and page values are encoded.
- `SortNone` omits sort params.
- Row links containing `'`, `&`, or query strings render safely.

### Modal and Drawer

These components build Alpine state names and event expressions from component
IDs. The first pass should keep their public API intact while adding helper
coverage for the generated names:

- Generated state identifiers must be valid JavaScript identifiers.
- String values embedded in Alpine expressions must be JavaScript-escaped.
- Empty IDs should fall back to a deterministic safe identifier or avoid ID-based
  state generation.

Add focused tests around helper functions or rendered attributes. Do not
redesign modal/drawer behavior in this pass.

## Component Authoring Skill

Create a local skill named `component-authoring` in both skill trees used by this
repo:

- `.agents/skills/component-authoring/SKILL.md`
- `.claude/skills/component-authoring/SKILL.md`

The skill should trigger when creating, editing, reviewing, or hardening
`components/<name>/types.go`, component `.templ` entry points, or component
public config APIs.

The skill body should stay concise and include these rules:

- Keep zero values useful for simple presentational components.
- Add `Validate() error` for complex components with required IDs, endpoints,
  mutually exclusive modes, server/client modes, or dependent fields.
- Never build URLs by string concatenation; use `net/url`.
- Never interpolate untrusted or caller-provided values into JavaScript without a
  JavaScript string/identifier helper.
- Avoid package-level mutable render state.
- Prefer typed constants over free-form strings for modes, variants, and states.
- Use `templ.Attributes` as an escape hatch when consumers need arbitrary HTML,
  HTMX, or Alpine attributes.
- Keep `Class` additive and do not let it replace required accessibility or
  layout classes.
- Add unit/render tests for helper logic, validation, escaping, and generated
  attributes.
- After API changes, run `templ generate`, `go test ./components/...`, and
  `go run ./scripts/skillgen`.

The skill should reference `component-docs` for demo-page changes,
`using-goshtoso` for consumer usage, and `htmx`/`alpinejs`/`templ` for
interactivity and escaping details.

## Testing

Use targeted component tests first, then the package baseline:

- New render/helper tests in the touched component packages.
- `go test ./components/...`
- `templ generate` after `.templ` edits.
- `go run ./scripts/skillgen` after component API changes.

Full site E2E is not required for this first pass unless a template change
affects browser-only behavior that unit/render tests cannot cover.

## Compatibility

The first pass should be backward compatible except for fixing broken output.
Existing public config fields should remain unless a field is clearly misleading
and can be corrected without breaking callers. New helper methods and tests are
preferred over sweeping public API changes.

## Out of Scope

- Adding `Validate()` to every component package.
- Renaming all public config fields for uniformity.
- Reworking the table component's full interaction model.
- Changing demo page structure or docs-page patterns.
- Running the full E2E suite unless implementation uncovers browser-only risk.

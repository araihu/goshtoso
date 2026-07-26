# Goshtoso Application Patterns

This directory is a standalone consumer module and an external-style benchmark
for Goshtoso. It demonstrates a small operations application using only public
library packages. It does not import the repository's `site/` module, load a
CDN, or require an application-specific Tailwind build.

## Routes and patterns

| Route | Pattern | Purpose |
| --- | --- | --- |
| `GET /` | App Shell | Persistent navigation, status, appearance controls, and pattern entry points |
| `GET /operations` | Operations List | Operational table plus loading, empty, error, and success states |
| `GET /operations/{id}` | Detail Workspace | Focused evidence, guardrails, ownership, and run metadata |
| `GET /workflows/deploy` | Multi-step Workflow | Target, release, and review steps |
| `POST /workflows/deploy` | Multi-step Workflow | Server-side validation and workflow advancement |

The operations state matrix is available with `?state=loading`, `empty`,
`error`, or `success`. HTMX requests to the same route receive only the
`#operations-state` fragment. Appearance is explicit and stateless:
`?theme=goshtoso|minimal&mode=light|dark`.

## File map

| File | Responsibility |
| --- | --- |
| `main.go` | Domain fixtures, method-qualified `http.ServeMux`, handlers, view models, and component configs |
| `views.templ` | The four server-rendered patterns and state matrix |
| `views_templ.go` | Generated templ output; never edit manually |
| `app.css` | Product-specific CSS embedded into the binary, with a 500-line regression budget |
| `main_test.go` | Render, HTTP, asset, method, workflow, theme, and import-boundary tests |
| `SNAGS.md` | Source-dives and consumer friction encountered during the benchmark |
| `go.mod`, `go.sum` | Isolated consumer module with a local replace to the repository root |

## Public Goshtoso components

- `head.Dependencies` and `assets.Handler` provide local CSS, Alpine.js, HTMX,
  and component scripts.
- `appshell`, `navbar`, and `sidebar` provide the responsive frame, one skip
  link, desktop navigation, and the mobile drawer.
- `pageheader`, `card`, `badge`, `link`, and `button` provide hierarchy,
  status, native navigation, arbitrary Card body content, and multi-action
  form submits through `button.WithAttrs`.
- `table` provides the success-state operations list and full-row navigation.
- `toolbar` groups the state controls; `skeleton`, `emptystate`, and `alert`
  represent loading, empty, and error outcomes.
- `breadcrumbs` structures the detail workspace; `select` and `textinput`
  provide the workflow fields; `steps` represents workflow progress.

The custom `app.css` is deliberately application-owned: its 398 lines define
Atlas-specific editorial layouts and branding while the public composition
components own the shell, page hierarchy, navigation, links, and fields. The
tests cap it at 500 lines; the pre-composition baseline was 728. It is embedded
with the binary and loaded after `/assets/styles.css`, so this fixture needs
neither a CDN nor an application-specific Tailwind build. Theme selection is applied with
`data-theme="goshtoso|minimal"` and dark mode with the `.dark` class; no browser
storage is required.

## Commands

Run all commands from this directory:

```bash
templ generate
GOWORK=off go test ./...
GOWORK=off go run .
```

Then open <http://localhost:3000>. The tests also prove that:

- every pattern route renders;
- the operations state matrix renders as both a full document and an HTMX
  fragment;
- Goshtoso and application CSS are served locally;
- application-owned CSS stays within the 500-line composition budget;
- Minimal/Goshtoso and light/dark appearance markers are emitted;
- the shell exposes exactly one skip link, one `h1`, and one desktop main
  scroll region;
- empty and error states include useful next actions;
- unsupported HTTP methods return `405 Method Not Allowed`;
- the module has no `github.com/araihu/goshtoso/site` import.

# Goshtoso application patterns

Use this reference after the library is integrated and the first component
renders correctly. It turns component choices into four common product
surfaces. Start with the closest pattern, keep domain decisions in the app, and
use `visual-acceptance.md` before calling the result complete.

For dashboards, settings, onboarding, content/marketing, or an ambiguous brief,
start with `design-intelligence.md`. It routes those archetypes into the task
patterns below without inventing category widgets or an aesthetic preset.

## Choose a pattern

| User task | Start here | Primary Goshtoso packages |
|---|---|---|
| Navigate an authenticated product | App Shell | `appshell`, `navbar`, `sidebar`, `search` |
| Find and act on many resources | Operations List | `pageheader`, `toolbar`, `panel`, `table`, `emptystate`, `skeleton` |
| Inspect and change one resource | Detail Workspace | `pageheader`, `breadcrumbs`, `panel`, `badge`, `tabs`, `button` |
| Complete a long or risky task | Multi-step Workflow | `pageheader`, `steps`, `panel`, `form`, `alert`, `button` |
| Publish a static public presence | Static Brand Site | tokens plus product-owned templ/CSS |

For every pattern, model `loading`, `empty`, `error`, and `success` before
polishing the happy path. Add permission-denied, stale, partial, and destructive
confirmation states when the domain can produce them.

## Static Brand Site

### Problem

Publish an organization, product, portfolio, or publication without inventing
an application dashboard or coupling basic content to a server runtime.

### Start here

Copy `examples/brand-site` with `goshtoso -init-brand-site=./my-site`. Its Go
binary writes `public/index.html`, Goshtoso's compiled stylesheet, and a
product-owned brand stylesheet. The output is deployable on any static host.

### Contract

- The product owns identity, art direction, imagery, copy, typography, and
  content rhythm in its templ and CSS.
- Goshtoso owns its semantic token vocabulary and any added controls, forms,
  feedback, or navigation primitives.
- Use ordinary document landmarks, native links, and one clear heading
  hierarchy. Do not add an App Shell, sidebar, dashboard metrics, or generic
  hero/features/testimonials/CTA sequence without a real product reason.
- Keep custom CSS deliberately scoped to the product; it is not a component
  workaround and must not leak into Goshtoso's base themes.

### Completion checks

- Static build produces HTML and every referenced stylesheet.
- At 390 px and 1440 px, navigation, content measure, and all primary links
  remain usable.
- Light/dark behavior is intentional when the brand theme supports both;
  otherwise set the declared color scheme and verify contrast.
- Honor `prefers-reduced-motion` for brand animation.

## App Shell

### Problem

Provide stable orientation while content changes. The shell owns global
navigation, responsive positioning, the main scroll container, page title
updates, and overlays. Feature pages own their content and local tools.

### Anatomy

1. A skip link is the first focusable element.
2. A compact top bar contains product identity, global search, theme controls,
   and account actions.
3. A sidebar contains stable sections and the current-page state.
4. Exactly one main region scrolls. `appshell.AppShell` defaults it to
   `id="main-content"` and `tabindex="-1"` so navigation can restore focus;
   override those values only when the application has an equivalent target.
5. Breadcrumbs and the page title live immediately above feature content.
6. Drawers, search results, and toasts render outside the scroll region so
   clipping does not hide them.

Start with `appshell.AppShell` for the frame. AppShell renders the `<header>` landmark,
so supply the top region's contents rather than nesting another `<header>`.
Supply `navbar.Navbar` or another
header component, `sidebar.Sidebar` for desktop navigation, and the route's
content either as templ children or `Config.Content`. Templ children are the
shortest path for page-local markup; `Config.Content` remains useful when the
content is already a component value. Its defaults include the skip link and
one scrollable `main`; app code still owns the mobile navigation trigger and
route state. When `sidebar.Sidebar` is nested inside AppShell, set
`DisableSkipLink: true` so the shell remains the single owner of page-level
skip navigation.

### HTMX contract

- Swap only the feature region for routine navigation.
- Push the canonical URL and update the document title with the same response.
- Return breadcrumbs or counters out of band when they belong to the shell.
- After a navigation swap, focus `#main-content` and move its scroll position to
  the top.
- Do not cache pages whose Alpine state cannot be reconstructed. Prefer a full
  body boost or `hx-history="false"` for those routes.

### Responsive contract

- At 390 px, keep the top bar visible, place the sidebar in a drawer below it,
  and preserve a full-width content column. Use viewport-owned geometry such as
  `fixed top-16 bottom-0`, adjusted to the actual top-bar height; `absolute
  top-full` inside a header child can open the panel below the viewport.
- At 1440 px, keep the sidebar visible and fixed-width while the content region
  receives the remaining width.
- The shell owns responsive positioning. The sidebar component owns its
  internal borders, background, and navigation layout.

### Accessibility

- Label the primary navigation and every icon-only action.
- Keep the active route structural with `aria-current="page"`.
- Open mobile navigation with a real button and return focus to its trigger.
- Give the shell one primary mobile navigation trigger. A Navbar right action
  may create a second Navbar menu; do not place it beside a Sidebar overlay
  hamburger unless their names and destinations are intentionally distinct.
- Escape closes every mobile navigation surface, restores a truthful trigger
  name, and leaves `aria-expanded="false"`.
- Global search must work with keyboard navigation and Escape.

### What stays application-specific

Product name, navigation information architecture, authorization, search
domain, account actions, telemetry, and persistence policy.

### Completion checks

- Direct navigation and HTMX navigation render the same page.
- Browser back/forward restores URL, title, focus, and a usable state.
- Only one vertical content region scrolls at desktop width.
- The mobile drawer starts below the top bar and never obscures focused content.
- Opening the drawer at 390 px yields a panel that intersects the viewport and
  has positive width and height; `display: block` and `aria-expanded=true` alone
  do not prove the navigation is reachable.

## Operations List

### Problem

Help a user scan, filter, compare, and act on many resources without losing
context. This is a work surface, not a dashboard gallery.

### Anatomy

1. A Page Header names the resource, explains scope, shows a count, and exposes
   one primary action.
2. A Toolbar contains search, filters, saved views, or bulk actions. Keep it
   close to the results it controls.
3. A compact live region reports the result count and active filters.
4. Table owns sortable columns, rows, pagination or infinite loading.
5. A state region replaces the result surface during loading, empty, or error
   states without moving the page tools.

Use `pageheader.PageHeader` for identity and the primary action,
`toolbar.Toolbar` for search/filter/action regions, `skeleton.Skeleton` while
rows load, and `emptystate.EmptyState` when there is no data. These composition
components provide strong layout and accessibility defaults while keeping
queries, permissions, and domain language in the app.

For rich status cells and predictable row navigation, use the public cell and
link-mode types directly:

```go
table.Row{
	Link:     "/operations/op-104",
	LinkMode: table.LinkFull,
	Cells: map[string]table.Cell{
		"status": {Component: badge.Badge(badge.Config{Label: "Healthy"})},
	},
}
```

`Cell.Component` accepts any `templ.Component`. Prefer `LinkFull` when the
destination depends on a fresh document or complex Alpine state; use
`LinkBoost` only when HTMX navigation preserves the destination contract.

When a row needs both primary navigation and trailing controls, set `Link` and
`Actions` together. Goshtoso keeps the `<tr>` non-interactive, renders the link
inside the first data cell, and leaves the action buttons as independent
keyboard targets. Do not add row-level click handlers as a second navigation
path.

### State matrix

| State | Required response |
|---|---|
| `loading` | Skeletons match the final columns and preserve surface height. |
| `empty` | Explain what belongs here and offer a relevant first action. |
| `filtered-empty` | Preserve filters and offer a clear-filters action. |
| `error` | Keep the query visible, explain the failure, and provide retry. |
| `success` | Report result count and keep actions near the rows they affect. |

### Responsive contract

- At 390 px, stack the Page Header action, collapse secondary filters, and keep
  the table horizontally scrollable rather than squeezing data into illegible
  columns.
- At 1440 px, keep frequently used filters visible and use width for readable
  columns, not decorative metrics.

### Accessibility

- Every filter needs a persistent label.
- Result counts use `aria-live="polite"` without announcing every keystroke.
- Loading regions use `aria-busy="true"` and retain a useful accessible name.
- Row actions must be reachable without relying on a click anywhere behavior.

### What stays application-specific

Column priority, filter vocabulary, status mapping, authorization, destructive
actions, sorting defaults, and whether pagination or infinite loading best fits
the task.

### Completion checks

- Search and filters survive refresh and browser navigation when their state is
  meaningful to the user.
- A retry does not clear the user's query.
- The empty state distinguishes no data from no matching data.
- Keyboard users can reach the first result without traversing hidden controls.

## Detail Workspace

### Problem

Let a user understand one resource, see its current status, move between related
views, and perform safe actions without losing identity or context.

### Anatomy

1. Breadcrumbs return to the parent collection.
2. An identity strip contains the title, stable identifier, status, and primary
   action.
3. Tabs or a compact local navigation divide substantial, peer-level views.
4. The main column contains the active task or data surface.
5. A narrow rail contains ownership, timestamps, links, and secondary facts.
6. Destructive actions stay visually separated from routine actions.

Use `panel.Panel` for the main decision surface and secondary rail when they
need a neutral frame. Supply their headings explicitly; Panel deliberately does
not choose a heading level or landmark role.

### State matrix

- `loading`: preserve the identity strip if it is already known and skeleton the
  active view.
- `partial`: show available facts and name the unavailable section.
- `stale`: show when data was observed and offer refresh.
- `error`: scope the error to the failed region when the rest remains usable.
- `success`: keep status and last-updated information visible after mutations.

### Responsive contract

- At 390 px, stack identity metadata, make actions full-width only when needed,
  and move the detail rail below the main content.
- At 1440 px, use a dominant main column and a clearly secondary rail. Do not
  make both columns equal when their tasks are not equal.

### Accessibility

- Tabs follow the ARIA tab keyboard model or use ordinary links when navigation
  changes the route.
- Status is text, not color alone.
- Mutation feedback moves focus only when the user's next action requires it.
- Collection search/filter swaps preserve focus and caret in the initiating
  control; only a response with an explicit focus target may move it.
- Destructive confirmation names the resource and the irreversible effect.

### What stays application-specific

Identity fields, status lifecycle, tab information architecture, action
permissions, refresh model, and destructive policy.

### Completion checks

- A copied detail URL opens the same active view.
- Status and action availability never contradict each other.
- Stale or partial evidence has an explicit mutation policy; irreversible
  actions are disabled when the available evidence cannot support them.
- The rail follows the main content on small screens.
- Partial failure does not replace the entire workspace unnecessarily.

## Multi-step Workflow

### Problem

Guide a user through a long, consequential, or dependency-ordered task while
preserving progress and making the final submission reviewable.

### Anatomy

1. A short Page Header explains the outcome, not the implementation.
2. Steps show progress and announce the current step.
3. Each step groups fields by the user's mental model, with persistent labels
   and honest helper text.
4. Server validation returns the same step with field errors and a summary.
5. A stable action footer contains Back and the single forward action.
6. The final step reviews the exact submission and provides Change links back
   to each group.

`panel.Panel` is appropriate around the step body or final review when the form
needs a stable full-width surface. Keep the actual form and action semantics in
`form.Form`, `FieldGroup`, `FormErrors`, links, and buttons.

### State matrix

| State | Required response |
|---|---|
| `editing` | Preserve entered values and show the current step. |
| `invalid` | Focus the error summary, associate errors with fields, keep values. |
| `submitting` | Prevent double submit and announce progress. |
| `error` | Preserve the review and offer a safe retry. |
| `success` | Name the created result and the next useful action. |

### Responsive contract

- At 390 px, keep fields in one column and ensure the action footer does not
  cover focused controls or validation messages.
- At 1440 px, limit form line length and use multiple columns only for fields
  that users naturally compare.

### Accessibility

- The step indicator is additional context, not the only heading.
- Error summaries link to fields and errors are not represented by color alone.
- Every action has at least a 44 px target on touch layouts.
- Back never discards valid input without warning.

### What stays application-specific

Step order, validation rules, persistence, authorization, defaults, side
effects, retry semantics, and success destination.

### Completion checks

- Refresh and Back behavior are explicitly chosen and tested.
- Successful native POSTs use Post/Redirect/Get when the receipt or completed
  workflow is a durable route; Back must not land on a resubmission error page.
- Compare a Back-restored task document with current server revision and status.
  Use `Cache-Control: no-store` on sensitive task pages or refresh persisted
  history entries on `pageshow`; stale-write rejection alone does not prevent a
  misleading Available, pending, or Completed view.
- Invalid submission preserves every valid value.
- Review shows the exact payload in language the user understands.
- Double submission cannot create duplicate work.
- Draft restoration keeps each composite control's visible label, submitted
  value, status text, and dependent UI in agreement. When restoring client-side,
  set the public input value and dispatch its documented `input` or `change`
  event; otherwise rerender the selected config from the server.

## CSS boundary

The embedded Goshtoso stylesheet guarantees component markup and the application
layout utilities used by the official recipes. It is not a general Tailwind
compiler for arbitrary classes supplied through `RootClass` or attributes.

The audited guaranteed additions are `max-w-7xl`, `xl:grid-cols-4`,
`lg:col-span-2`, `min-h-64`, `sm:text-4xl`, `first:pt-0`, `last:pb-0`,
`min-w-[220px]`, and `sm:col-span-2`. If an application needs other Tailwind
utilities, give it its own Tailwind build and load that stylesheet after
`/assets/styles.css`. A valid class name can otherwise fail silently because no
matching selector exists in the embedded CSS.

Built-in themes are selected with `data-theme="goshtoso"` or
`data-theme="minimal"` on `<html>`. Dark mode uses the `dark` class on the same
element. Apply these markers before first paint when the app persists a choice.

Application-owned CSS may consume Goshtoso's stable semantic custom properties
instead of copying theme colors into a second palette. Start with
`--color-surface`, `--color-surface-alt`, `--color-on-surface`,
`--color-on-surface-strong`, `--color-on-surface-muted`, `--color-outline`, and
`--color-primary`; use `--color-info`, `--color-success`, `--color-warning`, and
`--color-danger` for status fills, borders, and icons. Small status text on a
surface uses the contrast-safe derived pairs such as
`text-danger-text dark:text-danger-text-dark` and
`text-success-text dark:text-success-text-dark`; do not assume a single fill
color also contrasts as text in both modes. Dark tokens follow their semantic group:
`--color-surface-dark-alt`, `--color-on-surface-dark-strong`,
`--color-outline-dark-strong`, and `--color-primary-dark` are representative
names. Base status tokens are shared across modes; their derived `*-text-dark`
variants follow the dark surface foreground. Alias these into product-named
variables when that makes app CSS clearer. See `docs/THEMING.md` for the
complete contract.

For a filled semantic action, use `button.WithTone` rather than composing raw
status utilities. Button's derived `*-action` pairs guarantee a matching
foreground in both modes and keep hover contrast without lowering the whole
control's opacity.

## Field-proven compositions

These deeper compositions were recovered from independent consumer builds. Use
them as extensions of the four base patterns, not as screens to copy literally.

### Decision Queue

Combine an Operations List with a Detail Workspace when a reviewer must keep a
queue, amount or risk summary, selected record, policy context, and audit trail
visible while choosing an outcome. Keep the selected row and decision state
structural, separate destructive actions, use a real form for mutations, and
stack the selected record after the queue at 390 px. An error may replace only
the detail region when the queue can remain useful.

If detail navigation swaps only the workspace, synchronize the collection after
settle: URL, detail key, focus target, selected-row styling, and
`aria-current`/`aria-selected` must identify the same record. A stale highlight
is a state-integrity defect even when the correct detail is visible. Prefer a
server-rendered collection update; otherwise update all representations in one
small navigation handler and test them together.

Treat the decision lifecycle as a state machine. Test a second terminal action,
two stale tabs, the exact conflict-resolution control, and a repeated request
with the same idempotency identity. Stale evidence must refresh to new evidence
or block the decision; reloading the same stale fixture is not recovery.

Hold each real mutation in flight and assert pending copy plus a disabled or
otherwise deduplicated submitter. `button.WithLoadingText` follows an ancestor
HTMX form; put `hx-disabled-elt="find button[type='submit']"` on that form. A
fixture labeled “Loading” is not evidence that the mutation path announces or
deduplicates work.

### Interruption-safe Workflow

Extend the Multi-step Workflow for dock, field, or tablet use. Default every
line to the common case, ask the user to edit exceptions only, persist a bounded
draft, review the exact changes before submission, and make retries idempotent.
A sticky footer may keep Back/Review visible, but it must leave room for focused
fields and the last error at 390 px. Exercise every step at 390 px: long action
labels stack or wrap inside the viewport, and scrolling the final field into
view leaves it completely above the footer.

### Content-first Review

Combine a compact queue with a reading surface and a narrow decision rail when
the user's primary work is judgment over long-form content. Typography and
rhythm should belong to the publication or domain, while Goshtoso supplies
status, form, feedback, and responsive primitives. Keep author, deadline,
channel, current status, and the handoff note adjacent to the decision. Avoid
turning prose into equal dashboard cards.

## Verified standalone recipe

[`examples/application-patterns`](https://github.com/araihu/goshtoso/tree/main/examples/application-patterns)
is the repository's standalone consumer module. Its file map includes handlers,
domain fixtures, templ views, app CSS, tests, and module boundaries; it imports
public Goshtoso packages only. In the repository, its `replace` directive tests
the local candidate. To copy it into another workspace, follow its README:
remove the local replace, pin a released Goshtoso version, run `go mod tidy`,
then regenerate and test. The interactive examples under `site/` are demo-site
applications and may depend on `site/internal` packages, so do not describe
them as copyable standalone apps.

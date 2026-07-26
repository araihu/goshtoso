# Goshtoso application patterns

Use this reference after the library is integrated and the first component
renders correctly. It turns component choices into four common product
surfaces. Start with the closest pattern, keep domain decisions in the app, and
use `visual-acceptance.md` before calling the result complete.

## Choose a pattern

| User task | Start here | Primary Goshtoso packages |
|---|---|---|
| Navigate an authenticated product | App Shell | `appshell`, `navbar`, `sidebar`, `search` |
| Find and act on many resources | Operations List | `pageheader`, `toolbar`, `table`, `emptystate`, `skeleton` |
| Inspect and change one resource | Detail Workspace | `pageheader`, `breadcrumbs`, `badge`, `tabs`, `button` |
| Complete a long or risky task | Multi-step Workflow | `pageheader`, `steps`, `form`, `alert`, `button` |

For every pattern, model `loading`, `empty`, `error`, and `success` before
polishing the happy path. Add permission-denied, stale, partial, and destructive
confirmation states when the domain can produce them.

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
4. Exactly one main region scrolls. Give it `id="main-content"` and
   `tabindex="-1"` so navigation can restore focus.
5. Breadcrumbs and the page title live immediately above feature content.
6. Drawers, search results, and toasts render outside the scroll region so
   clipping does not hide them.

Start with `appshell.AppShell` for the frame. Supply `navbar.Navbar` or another
header component, `sidebar.Sidebar` for desktop navigation, and the route's
content as `Config.Content`. Its defaults include the skip link and one
scrollable `main`; app code still owns the mobile navigation trigger and route
state.

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
  and preserve a full-width content column.
- At 1440 px, keep the sidebar visible and fixed-width while the content region
  receives the remaining width.
- The shell owns responsive positioning. The sidebar component owns its
  internal borders, background, and navigation layout.

### Accessibility

- Label the primary navigation and every icon-only action.
- Keep the active route structural with `aria-current="page"`.
- Open mobile navigation with a real button and return focus to its trigger.
- Global search must work with keyboard navigation and Escape.

### What stays application-specific

Product name, navigation information architecture, authorization, search
domain, account actions, telemetry, and persistence policy.

### Completion checks

- Direct navigation and HTMX navigation render the same page.
- Browser back/forward restores URL, title, focus, and a usable state.
- Only one vertical content region scrolls at desktop width.
- The mobile drawer starts below the top bar and never obscures focused content.

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
		"status": {Component: badge.Badge(badge.Config{Text: "Healthy"})},
	},
}
```

`Cell.Component` accepts any `templ.Component`. Prefer `LinkFull` when the
destination depends on a fresh document or complex Alpine state; use
`LinkBoost` only when HTMX navigation preserves the destination contract.

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
- Destructive confirmation names the resource and the irreversible effect.

### What stays application-specific

Identity fields, status lifecycle, tab information architecture, action
permissions, refresh model, and destructive policy.

### Completion checks

- A copied detail URL opens the same active view.
- Status and action availability never contradict each other.
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
- Invalid submission preserves every valid value.
- Review shows the exact payload in language the user understands.
- Double submission cannot create duplicate work.

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

# Benchmark snags

This ledger records consumer source-dives and integration friction encountered
while building the standalone application-patterns benchmark.

## 2026-07-25

### Constructor styles required a model/source check

The public packages do not share one constructor style: `button.Button` uses
functional options while `badge`, `alert`, `card`, `steps`, and `table` use
package-specific config structs. The benchmark consulted
`docs/COMPONENT_MODEL.md`, `.agents/skills/using-goshtoso`, and the public
`types.go` files to keep calls idiomatic. This is documented behavior, but it
slows blind composition until the consumer learns which style each package
uses.

### Table status cells required a type source-dive

The generated consumer reference identified `table.Table`, but composing a
status badge inside a row required checking `components/table/types.go` to
confirm that `table.Cell.Component` accepts `templ.Component`, and checking
`table.LinkMode` to choose safe full navigation for detail workspaces. The API
worked without an adapter once those types were found.

### Workflow submit buttons cannot carry native name/value attributes

`button.Button` exposes type, ID, class, HTMX, Alpine, and loading options, but
not arbitrary native attributes. A multi-action form therefore cannot put
`name="action"` and `value="back|next"` on Goshtoso buttons. This benchmark uses
a normal link for Back and derives Continue versus Deploy from the posted step.

### Theme mechanics required source confirmation

The consumer guide covers embedded CSS, but the exact selectors for the two
benchmark appearances were confirmed in `all-themes.css`: themes use
`data-theme="goshtoso|minimal"` and dark mode uses `.dark`. The application
keeps these choices in query/form values instead of writing browser storage,
and its embedded CSS consumes the same public custom properties.

### templ injects its own package import

The first generation pass failed because `views.templ` explicitly imported
`github.com/a-h/templ` while also using `templ.URL` and `templ.KV`; the generated
file already injects that import, producing a redeclaration. Removing the
explicit import and regenerating is the correct consumer pattern.

## Control-plane review and remediation

The benchmark implementation was reviewed after integration. The following
issues were fixed before accepting the slice:

- the shell now starts with a keyboard-visible skip link and exposes one main
  desktop scroll region;
- empty and error states now offer concrete next actions instead of only
  describing them;
- decorative hero effects and broad shadows were removed so the operational
  hierarchy carries the interface;
- appearance parameters are preserved when state actions navigate;
- the consumer reference now documents `table.Cell.Component`,
  `table.LinkMode`, exact theme selectors, and templ's injected import. Those
  additions remove four source lookups from the next blind build.

The button attribute gap was closed in the consolidated branch with
`button.WithAttrs(templ.Attributes)`. The workflow now exercises it with real
Back and Continue submit actions, and the handler tests Back without validating
the step being left. This converts the original snag into a public API contract.

The consolidated benchmark also replaced its hand-built state toolbar, spinner,
and empty alert with `toolbar.Toolbar`, `skeleton.Skeleton`, and
`emptystate.EmptyState`, and exercises `card.Config.Body`. These APIs were
promoted from repeated consumer composition after the original benchmark, so
they are recorded as remediation rather than hidden in the initial source-dive
measurement.

### Final mobile matrix exposed min-content overflow

The 390 px checks found that the success table and loading skeleton could
expand their grid/flex ancestors and create document-level horizontal
scrolling. The table already owned an `overflow-x-auto` viewport; the missing
contract was on its ancestors. `.app-stack` now uses `min-width: 0` and an
explicit `minmax(0, 1fr)` track, direct children opt out of intrinsic minimums,
and the loading card stacks at the mobile breakpoint. Wide data remains locally
scrollable without moving the application shell. The asset test protects the
containment rules.

### The benchmark outlived its own composition baseline

After AppShell, PageHeader, Link, Toolbar, EmptyState, and Skeleton were
promoted, this benchmark still hand-built the shell, responsive navigation,
page headings, action links, and workflow fields. That left 728 lines of
application CSS and understated the value of the new public kit. Migrating the
fixture to `appshell`, `navbar`, `sidebar`, `pageheader`, `link`, `select`, and
`textinput` reduced the application-owned stylesheet to 398 lines (45.3%)
without introducing a public Stack/Grid CSS vocabulary.

The migration exposed one real composition snag: `AppShell` and `Sidebar`
both emitted a skip link. `sidebar.Config.DisableSkipLink` now lets the
containing shell own page-level skip navigation while preserving Sidebar's
standalone default. AppShell also accepts templ children as its content slot,
so a basic page no longer has to manufacture `Config.Content`.

The first browser pass then served a stale pre-migration stylesheet because the
standalone fixture cached `/app.css` for one hour. The benchmark now revalidates
that app-owned asset (`Cache-Control: no-cache`). It also gives AppShell an
application-specific `100dvh` root so content scrolls in `main` instead of
growing the document; both behaviors are covered by the live 1440 px checks.

The same browser pass exposed an accessibility mismatch in `Link`: choosing
`AppearanceButton` emitted `role="button"` even though the component remained
an anchor with `href` and native link keyboard behavior. Appearance is now
visual-only by default; `WithRole` remains available when a consumer has an
explicit semantic contract. Component and benchmark tests prevent the implicit
button role from returning.

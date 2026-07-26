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

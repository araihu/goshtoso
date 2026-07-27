# Component documentation shell adoption

## Outcome

The Goshtoso demo component pages now consume the public
`github.com/araihu/goshtoso-app-shells` module instead of owning a private copy
of the documentation frame. The adoption pins
`v0.0.0-20260727225437-d51e9ebd333b` and preserves the demo as the parity target.

The module separates two reusable contracts:

- `componentdocshell` owns the configurable outer documentation frame: brand
  header, search-first fixed navigation and mobile drawer, appearance controls,
  optional table of contents, document containment, and HTMX content swaps.
- `componentpage` owns repeated semantic component-reference sections and
  preview/code examples. A future catalog shell remains a separate pattern.

The consumer continues to own social metadata, structured data, favicon and
manifest links, runtime policy, brand assets, navigation/search entries,
application content, storage consent, and analytics decisions.

## Parity and regression evidence

- The full Goshtoso root tests, vet, and build pass.
- The full nested site tests, vet, and `go build ./cmd/server` pass, including
  the browser E2E suite.
- Component-page E2E covers direct and HTMX navigation, 1440 px, persistent
  736 px navigation, the 390 px drawer, Goshtoso and Minimal themes, dark/light
  icon state, hash/TOC alignment, document overflow, and console errors.
- Live 1280 px approval smoke confirmed the 288 px content offset, fixed
  search-first navigation, right-hand TOC, theme switching, icon swapping,
  search focus, zero document overflow, and zero console errors.

## Snags recorded

- The demo's legacy dark-mode store and element IDs needed explicit shell
  bindings rather than shell-owned assumptions.
- Native hash navigation split scroll state across the document and the
  internal content scroller; the shell now normalizes both and reserves trailing
  anchor range.
- A fixed document body created false 390 px overflow even when wide tables had
  correct local containment; root document containment is sufficient.
- Prose immediately following an inline `<code>` element cannot begin with the
  reserved word `for` in a templ source because it is parsed as control syntax.
- Running a newer templ generator touched unrelated generated files; those
  mechanical diffs were discarded so the adoption remains focused.

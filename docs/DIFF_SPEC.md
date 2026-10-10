# Generic text diff component

Status: implementation contract for issue #323.

Issue: [araihu/goshtoso#323](https://github.com/araihu/goshtoso/issues/323).

## Purpose

Add a generic, server-rendered component that presents differences between any
two texts. Supported content includes source code, prose, logs, configuration,
manifests, and plain text. Configuration is one example, not the component's
identity or a special rendering mode.

Consumers may supply precomputed comparison rows or calculate them with
`RowsFromText(before, after)`. Goshtoso uses `github.com/sergi/go-diff` v1.4.0 for
optional line comparison and owns presentation, responsive layout, theme
integration, and accessibility. Consumers retain parsing, normalization, and
application rules. Rendering does not interpret text or require a lexer,
parser, or client runtime.

## Public API

Use package `github.com/araihu/goshtoso/components/diff`, constructor `diff.Diff`,
and stable component identity `components.KindDiff` with value `"diff"`.
`Diff` returns a concrete `Instance` implementing `components.Component`.

Public types:

```go
type Operation string

const (
    OperationUnchanged Operation = ""
    OperationInsert    Operation = "insert"
    OperationRemove    Operation = "remove"
    OperationReplace   Operation = "replace"
)

type Line struct {
    Text   string
    Number int
}

type Row struct {
    Operation Operation
    Before    *Line
    After     *Line
}

type Config struct {
    ID          string
    BeforeLabel string
    AfterLabel  string
    Rows        []Row
    MaxHeight   string
    AriaLabel   string
    EmptyText   string
    RootClass   string
    RootAttrs   templ.Attributes
}

func RowsFromText(before, after string) []Row
func Diff(Config) Instance
func (Instance) WithExpressions(expressions.Diff) Instance
```

`Line.Text` contains a complete source line, excluding its newline separator.
When supplying rows directly, the consumer chooses how to split source text
and represent a trailing empty line. The renderer preserves all supplied text,
including leading and trailing spaces, tabs, Unicode, and HTML-looking
characters. It does not trim, tokenize, normalize, or calculate differences
within a line.

`Line.Number` is optional and refers to the original source. Positive values
are displayed verbatim; zero and negative values omit the number. Each side
has independent numbering. The renderer never invents or renumbers lines;
`RowsFromText` assigns source numbers before rendering.

A non-nil `Line` with empty text represents a real blank line. A nil side
represents alignment padding and has no source text, line number, or change
marker. This distinction is required for unequal replacement blocks.

## Row contract

Rows render in supplied order; each row reads before, then after.

| Operation | Before | After | Presentation |
| --- | --- | --- | --- |
| Unchanged | Source line | Matching source line | Neutral on both sides |
| Insert | nil | Source line | Addition on the after side |
| Remove | Source line | nil | Removal on the before side |
| Replace | Source line or nil | Source line or nil | Removal before, addition after; missing sides remain padding |

A replacement block is a sequence of replacement rows, pairing lines in
source order and using nil for the shorter side's remaining positions. For
example, replacing two lines with three uses two paired rows and a third row
with `Before: nil`.

When supplying rows directly, consumers are responsible for choosing operations
and pairing lines. The renderer performs no equality check. Unknown operations
render populated sides neutrally. A side supplied contrary to the operation contract still
renders its text, but only the operation's applicable side receives a change
marker. A row with both sides nil is omitted.

Nil or empty `Rows`, or rows with no populated sides, show the empty state.
Identical texts use unchanged rows and retain their text and optional numbers.
An entirely added or removed text uses insertion or removal rows, respectively.

## Calculating rows from text

`RowsFromText` uses sergi/go-diff's line-to-rune encoding, compares those tokens,
and rehydrates complete lines. It never uses character-level refinement or
returns library-specific types. Sources receive independent one-based numbers.
Adjacent additions and removals between unchanged blocks become replacement
rows, paired in source order with nil padding for the shorter side. A block
containing only additions or only removals uses insert or remove rows.

Empty strings contain no lines. LF and CRLF terminate lines; a final terminator
does not create an extra blank line. A real empty line, such as the second line
of `"a\n\n"`, remains populated. The helper removes only the line separator
from displayed text, preserving spaces, tabs, Unicode, and bare carriage
returns. Comparisons include the separators: changes to newline style or a
final newline mark the affected lines changed, even when their displayed text
is identical. The viewer does not display a separate newline warning.

The helper uses the library's default one-second comparison timeout. When that
search deadline is exceeded, the library returns coarser delete/insert changes;
all input lines still appear. Tokenization and row conversion are outside that
search deadline. Consumers needing other algorithms or policies can still
supply `Config.Rows` directly. Rendering never calculates rows implicitly.

## Layout and scrolling

At viewport widths of at least 640px, present paired before/after columns with
visible labels. Rows remain vertically aligned across both sides, including
padding in unequal blocks.

Below 640px, stack each pair in before-then-after order. Show the corresponding
consumer label alongside every populated side. Hide padding without hiding
real blank lines. An unchanged pair may display both labeled copies; this
preserves the same source association and reading order at every width.

Keep each source line intact in one source element with preserved whitespace.
Long lines use horizontal scrolling within the component. They must not overlap
the opposite column or cause the surrounding page to overflow. Column widths
must remain consistent across a comparison, even when one line is much longer
than the rest. Labels remain readable when source text is wide.

Provide a bounded scroll region, with `MaxHeight` defaulting to `32rem`.
Consumers may supply another CSS length such as `400px`. Vertical overflow
stays within this region. The region must be keyboard focusable and have a
visible focus indicator, allowing native keyboard scrolling without JavaScript.

## Semantics and accessibility

- Use `BeforeLabel` and `AfterLabel` to associate every populated side with its
  source. Labels are plain escaped text and must support arbitrary languages.
- Name the scroll region using `AriaLabel`. When omitted, derive its name from
  the effective before and after labels.
- Identify removal with a visible `−` marker and addition with a visible `+`
  marker. Accessible marker names describe the operation; color is supplemental.
- Expose source line numbers when supplied. Reading order associates the source
  label, operation, number, and complete text without replacing source content
  with an `aria-label`.
- Keep decorative padding out of the accessibility tree. Mobile labels must
  not produce duplicate announcements of the same source name.
- Render source text through normal escaping. HTML-looking text must never
  create elements, execute scripts, or acquire application behavior.

Built-in text belongs to `expressions.Diff`: `BeforeLabel`, `AfterLabel`,
`AddedLabel`, `RemovedLabel`, `UnchangedLabel`, and `EmptyText`. English defaults
are `Before`, `After`, `Added`, `Removed`, `Unchanged`, and `No lines to compare.`
Explicit config text takes precedence over expressions. Follow the existing
expression inheritance and immutable `WithExpressions` conventions, and
regenerate the expression-file schema.

## Themes and integration

Use Goshtoso semantic surface, text, outline, success, and danger tokens,
including their dark-mode counterparts and readable action colors for markers.
Use the theme's border radius and monospace font for source content. The visual
layout follows a pull-request split diff: a compact document-style header,
continuous 20px source lines at 12px, source numbers before operation markers,
and stronger addition/removal tint in the number gutters than the text area.
Missing sides use neutral diagonal padding; rows have no horizontal dividers.
Mobile source labels retain the same document icon and neutral surface. Preserve
readability in Goshtoso and Minimal themes, in both light and dark modes.

The first render and responsive behavior require no JavaScript. Consumers can
render the instance in a templ page or replace it with a normal HTMX fragment.
`ID`, `RootClass`, and `RootAttrs` provide ordinary integration hooks without
introducing a component runtime or automatic HTTP endpoints.

## Documentation and examples

Register `/components/diff` as **Diff** in the Display catalog and demo registry.
Document the precomputed-row contract, nil-versus-blank distinction, breakpoint,
scrolling behavior, defaults, and ordinary fragment replacement. Public Go doc
comments remain the API reference; regenerate component skill references.

Provide one preview and one matching code example per variant:

1. Plain text calculated from two strings with `RowsFromText`.
2. A source-code replacement block with unequal line counts and source numbers.
3. Long lines, tabs, indentation, Unicode, and literal HTML-looking text.
4. Empty inputs and identical inputs as separately identified examples.
5. A long comparison with a custom height bound.

Use generic labels such as `Original` and `Revised` across the primary examples.
A configuration example may be added within this generic documentation.

## Verification and acceptance

Calculation tests must cover empty and identical text, complete insertions and
removals, unequal replacements, independent numbering after edits, repeated
lines, LF/CRLF/final separators, whitespace, Unicode, and HTML-looking text.
Source reconstruction must retain every line in order on both sides.

Rendering tests must verify operations, independent source numbers, padding
versus blank lines, empty and identical inputs, unknown-operation fallback,
verbatim escaped text, expression precedence, and the common component identity.

Browser tests must verify:

- Aligned columns on desktop and readable stacked pairs at 390px.
- Both sides of unequal replacements, without visible mobile padding.
- Long source lines remain intact, with scrolling contained in the component
  and no overlap into the other source or surrounding page.
- Custom and default height bounds, plus keyboard scrolling and focus styling.
- Literal HTML-looking text stays text, including with JavaScript disabled.
- Light and dark modes in Goshtoso and Minimal, with token-derived colors and
  visible markers that distinguish operations without relying on color.
- Ordinary HTMX fragment replacement preserves layout and source associations,
  with no initialization step or console errors.

Before opening the implementation PR, regenerate templ, CSS, the expression
schema, and component references; run relevant root/site lint and unit checks,
current-source site integration, the published-consumer contract, and browser
regressions required by repository guidance.

## Out of scope

Custom diff algorithms, word or character highlighting, syntax
highlighting, parsers, normalization, editing, merge controls, submission
workflows, authorization, and language-specific validation. These can be
considered separately without restricting this component's generic text API.

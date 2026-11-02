# Spec 001: Generic layout containers and item movement

Status: implemented; public API and interaction reviews accepted. Preparing v0.5.0; publication pending.
Date: 2026-11-01. Owner: Goshtoso. Release stage: v0.1.0+; tagged baseline v0.4.7.
Current milestone: complete upstream release verification before downstream integration.
The component, generated catalog entry and demo are implemented. No publication or deployment is claimed.
Implementation and review are tracked in [PR #363](https://github.com/araihu/goshtoso/pull/363).

## 1. Purpose and verified gap

A consumer needs to arrange arbitrary `templ.Component` content in containers and
move direct children within or between compatible containers. Presentation, domain
membership, persistence and authorization stay with the consumer.

Baseline source: `6fed2b98f6b13e65e0f419a2d028b94d91afc1b9`.
The first consumer is Balemoh, which currently pins Goshtoso
`v0.4.8-0.20261101060924-c5f68d0bb167`.

Existing [component identity](../../COMPONENT_MODEL.md), Panel slots, Button,
Select, Form, Dialog, Accordion and ScrollRegion remain reusable. Table's sortable
columns sort tabular values; FileInput's drop surface selects files. Neither owns
movement of arbitrary existing renderable children. No public natural-height
Stack/Grid/Masonry container or cross-container movement contract was found in the
selected component inventory, configs, runtime sources and usage examples.

This upstream component must land, pass its consumer contract and be released before
Balemoh ships the new editor. It must not be copied from demo internals or rebuilt
as a Balemoh-only sortable/masonry runtime.

## 2. Responsibility and exclusions

One public package, `components/layout`, owns container/item wrappers,
responsive direct-child geometry, optional movement handles, keyboard/pointer
interaction, announcements and cleanup. It composes arbitrary child components.
The consumer owns all data semantics, sorting rules, validation, API calls, draft
state, Save/Cancel, group labels and permission checks.

Excluded: application-specific Cards, SQL/API clients, source-kind inspection,
alphabetical sorting, saved ranks, arbitrary-depth tree editors, absolute x/y
positioning, resizing a widget's internal content, browser storage and remote code.
A library `Kind` identifies type; it never substitutes for an item instance ID.

## 3. Public API

Add `layout.Container(Config) Instance`, implementing `components.Component` with
stable Kind `layout`. The declarations below are normative public Go API. They follow
[component API naming](../../COMPONENT_API_NAMING.md); the tagged v0.4.7 baseline does not include this package.

| Type | Fields |
|---|---|
| `Config` | `ID`, accessible `Label`, `Mode`, maximum `Columns`, `MinItemWidthRem`, `Gap`, `Items []Item`, optional `Reorder *ReorderConfig`, `RootClass`, `RootAttrs` |
| `Item` | Opaque stable `ID`, accessible `Label`, arbitrary `Content templ.Component`, `Span`, `RootClass`, `RootAttrs` |
| `ReorderConfig` | Compatible movement `Group`, `Disabled` |
| `Mode` | Stack, Grid, Masonry; Stack is zero/default |
| `Span` | Auto/default or FullWidth |
| `Gap` | Small/Medium/Large, 12/16/24 CSS px respectively |

Columns default to 1 and accept 1 through 4. MinItemWidthRem defaults to 12 and
accepts finite values from 8 through 40. IDs are nonempty, unique among direct
children, at most 256 UTF-8 bytes; container ID is both logical ID and the rendered root HTML `id`, with no
whitespace. Document uniqueness is a runtime activation precondition. Labels are nonempty and rendered as escaped text.
Item IDs must be unique across containers sharing the same nonempty movement
Group; empty Group requires uniqueness only within its own container.
Per-Config invalid input returns a render error before writing partial markup.
Cross-container/document collisions are detected by the runtime before movement
activation, not claimed as independently enforceable during server rendering.
Consumers may supply stable DOM IDs through Item.RootAttrs for existing fragment
targets; otherwise logical identity uses the item marker without inventing a DOM ID.
Duplicate owned DOM IDs within this Config, including generated instruction IDs,
and attempts to replace library-owned data markers are render errors. Cross-Config DOM collisions block runtime movement. Normal consumer semantic/ARIA attrs and non-owned classes remain supported.

No universal component-config interface is introduced. Content may be a Card,
Panel, chart, plain note, table or another layout component. No automatic visual
card wrapper, clipping or forced fixed height is added.

### 3.1 Exact exported declarations and defaults

```go
// Package layout imports context, io, templ, components and expressions.
type Mode string
const (
    ModeStack Mode = ""
    ModeGrid Mode = "grid"
    ModeMasonry Mode = "masonry"
)
type Span string
const (
    SpanAuto Span = ""
    SpanFullWidth Span = "full"
)
type Gap string
const (
    GapMedium Gap = ""
    GapSmall Gap = "small"
    GapLarge Gap = "large"
)
type Config struct {
    ID string
    Label string
    Mode Mode
    Columns int
    MinItemWidthRem float64
    Gap Gap
    Items []Item
    Reorder *ReorderConfig
    RootClass string
    RootAttrs templ.Attributes
}
type Item struct {
    ID string
    Label string
    Content templ.Component
    Span Span
    RootClass string
    RootAttrs templ.Attributes
}
type ReorderConfig struct {
    Group string
    Disabled bool
}
type ValidationError struct {
    Path string
    Code string
}
func (ValidationError) Error() string
// Instance has unexported immutable configuration/expression fields.
type Instance struct { /* unexported fields */ }
func Container(Config) Instance
func (Instance) WithContentDrag(bool) Instance
func (Instance) Kind() components.Kind
func (Instance) Render(context.Context, io.Writer) error
func (Instance) WithExpressions(expressions.Layout) Instance
```

Add `components.KindLayout = "layout"` and its generated inventory entry.
Zero Columns means 1; zero MinItemWidthRem means 12. Other out-of-range/non-finite
values are invalid. Zero Gap means 16 px. Nil Items means an empty collection.
Nil Content is permitted; enabled handle chrome still contributes
to natural item height. Nonempty labels containing non-whitespace text
are required for container and items, at most 256 Unicode scalar values.
Validation visits Config fields in declaration order, then items in slice order.
`ValidationError.Error()` is `layout: <Path>: <Code>`. Paths use Go field notation,
for example `Items[2].RootAttrs["id"]`. Stable codes are `invalid_id`,
`invalid_label`, `duplicate_id`, `invalid_mode`, `invalid_span`, `invalid_gap`,
`invalid_columns`, `invalid_min_width`, `duplicate_attribute`, `reserved_attribute`,
`unsupported_attribute` and `invalid_attribute_value`. Errors write zero bytes.

RootAttrs is `templ.Attributes`. Canonicalize keys with ASCII lowercase; reject
case-insensitive duplicate keys before any merge. Both roots accept `class` as a
string, `title`, `lang`, `dir`, `aria-*` except reserved names, consumer `data-*`
except owned markers, `hx-*`, `x-on:*` and only `x-bind:aria-current`, `x-bind:aria-expanded`,
`x-bind:aria-busy`. Other reactive ARIA bindings are not allowed. Item roots additionally
accept `id` as a valid non-whitespace string. Other attributes are rejected.
Config.RootAttrs cannot set `id`; Config.ID is the sole container ID authority.
Item.RootAttrs.id is the sole optional item DOM-ID authority. A Config rejects an
item DOM ID equal to its own ID, its generated instruction ID when rendered, or
another explicit item DOM ID in that Config.

Reserved on both roots: `role`, `tabindex`, `aria-label`, `aria-labelledby`,
`aria-describedby`, `hidden`, `draggable`, all `data-goshtoso-layout*`, all
`data-layout-*`, `x-layout`, `x-data`, `x-init`, and `style`. Inline styles belong
inside Content, not on structural roots. Reactive aliases for owned attributes are
also reserved, including `x-bind:aria-label`, `x-bind:aria-labelledby`,
`x-bind:aria-describedby`, and their colon shorthands. Apply the same canonical
case-insensitive duplicate/reserved validation to aliases. No reactive alias may
write id, role, tabindex, hidden, draggable, owned markers or geometry styles. Class output is library base classes,
then RootClass, then RootAttrs.class, space-separated in that order. Consumer CSS
must not override structural display/grid properties; such overrides are unsupported,
not something render validation claims to detect. Content styles remain unrestricted.
Only library geometry custom properties are merged/preserved during morph; class,
ARIA and other consumer attributes still reconcile normally.

### 3.2 Expressions and semantic text

Extend the root `expressions` package with this exact typed group, corresponding
Set/merge/schema/file-decoder support and generated inventory:

```go
type Layout struct {
    HandleLabel func(itemLabel string) string
    Instructions string
    MoveAnnouncement func(itemLabel, containerLabel, position string) string
    CancelAnnouncement func(itemLabel string) string
    MovementUnavailable string
}
```

WithExpressions follows existing per-field config/component/context/renderer/English
precedence; empty/nil inherits. Labels supplied in Config/Item remain app-owned text.
English defaults are `Move <item>`, complete keyboard instructions, `<item> moved to
<container>, position <position>.`, `Move cancelled for <item>.`, and
`Movement controls are unavailable. Use the application's explicit move controls.`

Handle/cancel functions resolve with each actual item label during server rendering.
MoveAnnouncement resolves once with three opaque string slot tokens. Its returned
complete sentence is serialized as a literal/slot token sequence, never an executable
JS expression. It must retain each supplied token at least once. Invalid/missing
slot tokens cause `invalid_attribute_value` at `Expressions.MoveAnnouncement` before
markup. At runtime substitute item label, destination label and ASCII decimal
1-based position as text nodes. Consumer text containing token-like characters is
inserted only after parsing and is never reparsed. This preserves translation word
order without an item-by-destination matrix or calling Go functions in JavaScript.
Numeric pluralization and locale negotiation are not new guarantees of this component.

## 4. Public markup and ownership

Library-owned markers are:

```html
<div id="operations" role="list" aria-label="Operations" tabindex="-1"
     data-goshtoso-layout="container" data-layout-id="operations">
  <div role="listitem" aria-label="Example item"
       data-goshtoso-layout="item" data-layout-item-id="opaque-item-id">
    <!-- Consumer's arbitrary component. -->
  </div>
</div>
```

The container iterates only matching direct children. Nested container contents
never become children of an outer algorithm. Item roots contain a neutral flow-root content boundary and, when enabled, an
in-flow handle row. Measurement covers the complete natural item border box and
its nonnegative block margins, not a prior grid-row allocation. The live region
is truly out of flow and does not contribute to track sizing.

Existing consumer markers such as `data-dashboard-widget` and widget fragment IDs
are passed through public RootAttrs. Runtime behavior selects public data markers,
not Card CSS classes or private child markup. Attribute values are escaped; IDs
are compared as opaque strings without constructing unescaped CSS selectors.

Grid/Masonry columns respond to their container's available inline width, up to
Columns, with the configured minimum and gap. Stack is one column. FullWidth spans
all columns of the parent. Sparse Grid placement preserves DOM reading order;
`grid-auto-flow: dense`, CSS columns and shortest-column DOM shuffling are excluded.
Masonry uses small row tracks and measured spans while retaining DOM order.

No JavaScript or no ResizeObserver yields ordinary responsive Grid for Masonry.
Content remains readable. Per-row stretching and equal-card heights are outside this component slice.
Container geometry does not become an accessibility grid;
use ordinary labelled collection/item semantics and preserve content's own semantics.

### 4.1 Accessible enhancement and registry

The normative container is a labelled `div[role=list][tabindex=-1]`; item roots are
`div[role=listitem]` labelled from Item.Label. Content retains its nested semantics.
When Reorder is non-nil and not Disabled, a native `button[type=button]` handle is rendered inside each item,
with a complete expression-derived accessible name. It references the owning
container's instruction element by aria-describedby. Instructions have a generated
ID exactly `Config.ID + "--layout-instructions"`, owned by that container and
rendered only when its enabled Reorder configuration renders handles. This ID
participates in render-time and document-wide DOM-ID collision checks. One container-local `role=status` live region has
`aria-live=polite` and `aria-atomic=true`; it receives complete outcome text.
The container root is the programmatic fallback focus target, not an extra Tab stop.

For this enabled configuration, server markup renders handles `hidden disabled` and enhancement-only instructions
hidden. Valid runtime initialization reveals/enables them; a blocked/removed runtime
hides/disables them again. No-JS users see ordinary content and a noscript
MovementUnavailable explanation only when Reorder was requested. They do not receive
an enabled dead drag button. Consumer explicit form controls remain their own alternative.

One registry per document tracks connected container elements by ID and movement
Group. Empty Group means same-container-only movement, never cross-container access;
nonempty Group permits transfers among enabled, visible, connected compatible roots.
Disabled/nil Reorder never activates handles. Registry validation separately enforces
document-unique owned DOM IDs and group-wide unique item IDs. Owned DOM IDs include
container root IDs, explicit item root IDs and every library-generated instruction
ID. Each owned ID must resolve to exactly its owning element among all connected
document elements with that ID, including unrelated consumer markup. Compare opaque
ID strings directly, without unescaped selectors. A generated/explicit/container
collision blocks every layout root owning that ID; no owner wins. Duplicate
container logical IDs use `duplicate_container_id`; other DOM-ID collisions use
`duplicate_dom_id`. Observe connected ID changes, additions/removals and replacements
so clearing a collision restores a single valid owner and correct IDREFs.

On a collision, every affected root has `data-layout-reorder-state=blocked`; hide
and disable its handles, cancel any pickup and emit one bubbling `layout:error`
per transition with `{containerId, code, itemId?}`. Codes are
`duplicate_container_id`, `duplicate_item_id` and `duplicate_dom_id`. Diagnostics do
not identify an arbitrary winning element. No move/before-move event is allowed while
blocked. Removal/reconciliation revalidates the registry; clearing a collision
reactivates valid roots once. Cleanup removes registry ownership and live resources.

## 5. Movement protocol

Reorder is opt-in. Nil or Disabled means no movement handles and no active move
listeners; reading and geometry still work. Movement Group is an explicit compatibility
scope. Distinct outer group and inner widget scopes prevent mixing levels.

Each movement handle displays a six-dot grip with a visually hidden localized
movement label and a matching tooltip. Its target is at least 44 by 44 CSS pixels.
Pointer pickup starts from the explicit handle by default; links, text selection,
fields and controls inside content remain usable. With `WithContentDrag(true)`, mouse or
pen movement of at least 6px also picks up non-interactive content. A click or smaller
movement has no move event or cancellation announcement. This opt-in replaces text
selection with dragging in non-interactive content; use handle-only mode when text
selection must remain available. Links, controls and editable
regions are excluded. Consumers that render an inactive editing preview may mark its
content `inert`; the surrounding item remains draggable and the named handle remains
available to keyboard users. Touch retains native content scrolling and uses the
handle or consumer's named move controls; pointer interaction must not require hover.

During an active pointer gesture, a passive ghost follows the pointer at the pickup
offset. A placeholder occupies the prospective insertion slot and displaces the
surrounding items in Stack, Grid and Masonry. Hover does not reorder real items or
emit movement events. Short displacement animations respect reduced motion. The
original item retains its identity and state until release. The ghost has no IDs,
runtime bindings or active custom elements, and stays outside keyboard and assistive
technology navigation. Cancellation, veto, blur, resize and source removal remove
all temporary visuals and restore the original presentation.
Escape cancels and returns focus. Keyboard handle movement uses Alt+Up/Down in logical
DOM order; Start/End actions can be supplied through consumer Button/Dropdown controls.
Instructions and outcomes are announced through an accessible live region.

Before committing a move, emit a bubbling, cancelable `layout:before-move` event:

```json
{
  "itemId": "opaque-id",
  "fromContainerId": "source",
  "toContainerId": "destination",
  "beforeItemId": null
}
```

JSON null means append. Otherwise beforeItemId is the opaque ID of a different
direct member of the destination. The consumer may cancel
for domain validation. Without cancellation, move the existing item root in the DOM
without recreating its content, recompute geometry and emit one `layout:move` with
the same normalized detail. This event means a local accepted move, never a server
save. Consumers update their local draft and later persist through their own API.

No-op drops emit no accepted-move event. Cancelled/out-of-scope/invalid-anchor moves
leave DOM order and focus unchanged. Cross-container transfer uses matching Group,
revalidates direct-child identity at drop and updates both containers once. Nested
handlers stop at the owning container. Root removal, configuration replacement or
membership change during pickup cancels before applying a stale move.

A consumer may use named explicit move/order controls instead of dragging. Applying
its new canonical child order through a server fragment or consumer-owned draft
reconciliation must not require simulating pointer events. The component observes
or reconciles direct-child changes through the same geometry/lifecycle path.

A successful move announces item label, destination label and logical position.
Focus remains on the same logical item's handle, with a container fallback if that
item disappears. A cancelled move restores its pre-pickup placement and focus.

### 5.1 Keyboard and programmatic requests

On a focused handle, Alt+Up/Down moves one logical sibling earlier/later;
Alt+Home/End moves to first/last within its current container. Boundaries are no-ops.
Alt+PageUp/PageDown transfers to the previous/next compatible container in connected
DOM preorder, appending there. This reaches nonadjacent containers by repetition and
empty destinations by append. Invisible, blocked and disabled containers are skipped.
Distinct movement Groups never share this traversal. Prevent shortcut defaults only
while handling a valid handle command; Escape cancels a pickup.

Consumers' named destination/start/end controls dispatch a bubbling
`layout:request-move` on the source container with the same normalized detail as
before-move. The runtime accepts it only when item, source, target, Group and anchor
validate. It runs exactly the same cancelable before-move, root move, accepted event,
announcement and focus logic as a pointer or shortcut. It makes no HTTP request.
Start maps to the first other member; End/empty destination maps to JSON null append.
No-op requests emit zero accepted events. Unsupported request detail emits
`layout:error` code `invalid_move_request` and changes no order. These request events
are the public non-pointer command protocol, not a private DOM helper.

All accepted input paths retain the same item handle as focus target even across
containers; missing item falls back to the source container root, or destination
root if source was removed. Focus is resolved by registry element identity and opaque
item equality, never by an ambiguous duplicate label.

### 5.2 Native disabled controls

There is no public temporary-disable command in this slice. Consumers own workflow
state and may place a configured container inside a native disabled fieldset.
Define a container's native-disabled predicate independently of its items: for each
ancestor `fieldset` with its `disabled` attribute present, find its first direct
`legend` child. That fieldset disables the container unless that legend exists and
contains the container root. The container is natively disabled if any ancestor
fieldset disables it. This matches native fieldset inheritance, including nested
fieldsets and the first-legend exception, and works for empty containers.

An available source/target requires this predicate to be false, valid enabled
Reorder configuration and an unblocked connected registry owner. Pickup and handle
commands also require the source handle's native `:disabled` state to be false.
Check actual current ancestors synchronously on pointer pickup, keyboard and
programmatic requests, before emitting before-move, and again immediately before
DOM commit after any synchronous before-move listeners. Revalidate source and
destination identity, compatibility and anchor at that final boundary too. Never
cache an enabled decision across consumer callbacks or reparenting. Programmatic
request-move does not bypass these guards. Disabled destinations are incompatible,
including empty destinations. An attempted unavailable move emits `layout:error`
code `reorder_unavailable`, changes no order and emits no accepted event.

The handle row remains in flow and visible when native fieldset disabling applies;
only native enabled state changes. Consumer code owns focus movement before disabling
its workflow surface. The runtime cancels an active pickup when its source becomes
natively disabled, observing ancestor disabled/legend/ancestry changes through the
existing lifecycle observation path, and does not restore
focus to a disabled control. Same-root morph reads effective native state, so no
cached flag can silently re-enable it. A replacement root derives state from its
actual ancestors/configuration; cleanup leaves no transient enable/disable registry.
Nil/Disabled Reorder and registry collisions still prevent activation regardless of
native fieldset state. Any library-owned transition hiding a focused handle moves
focus to its owning container root before hiding it. Re-activation never steals
focus from another current control.

## 6. HTMX, Alpine and geometry lifecycle

Initialize from the public markup after dependencies are ready, on normal page load
and on HTMX-processed subtrees. Initialization is idempotent. Use an Alpine directive
with cleanup for local state; vanilla measurement code is limited to browser geometry
that declarative state does not supply.

ResizeObserver and direct-child/content observation schedule at most one animation
frame. Nested measurement settles parent widths before child heights and measures
inner containers before assigning outer height spans. Font loading, fragment content,
width changes, hidden-to-visible groups and removed children are handled.

Cleanup disconnects every observer, removes handlers, cancels pending frames and
invalidates deferred font/async callbacks. Re-entry creates one fresh owner. Do not
retain detached DOM nodes or depend on consumer navigation to release resources.

The component owns only its row-span/column geometry properties and scoped classes.
HTMX morph reconciliation must preserve that geometry until a new measurement,
without suppressing consumer style/class/ARIA/content updates. Do not veto an entire
style attribute as a generic workaround. New configuration rereads current direct
children and options, rather than retaining stale arrays captured on first mount.

Content-only updates preserve consumer controls and focus. Structural changes while
moving cancel safely or are deferred by an explicit consumer boundary; they cannot
commit an obsolete anchor. Consumer draft semantics are outside this package.

### 6.1 Geometry algorithm and settlement

Let W be container content-box inline width in CSS px, g the configured gap, m
MinItemWidthRem times root computed font size, and M effective Columns. Grid/Masonry
use `c = max(1, min(M, floor((W + g) / (m + g))))`. Stack uses c=1. Equal-width
tracks use `repeat(c,minmax(0,1fr))`. FullWidth uses grid-column 1/-1. At widths
smaller than m the one column shrinks to available width without adding overflow.

The item root uses auto natural block size and align-self:start in Masonry. Its
Content boundary is flow-root so child margins cannot collapse out of measurement.
Enabled movement adds an in-flow handle row with a 44 CSS px minimum button block
size before Content, plus 4px spacing when nonempty Content follows. H is the entire
natural item root border-box height from getBoundingClientRect, plus computed
nonnegative block-start/end margins. Thus handle, content, consumer root padding and
borders all contribute. Negative root margins and consumer forced root block-size,
height, stretch or structural grid overrides are unsupported; sizing belongs inside
Content if needed. Instructions/live region are visually hidden and absolutely
positioned out of the grid tracks. Prior allocation is not measured: never stretch
Masonry roots/Content to span height, including when changing modes. For Masonry use row
track r=1 CSS px, row-gap=0, column-gap=g and `span=max(1,ceil((H+g)/r))`. Item roots
align to start and do not stretch natural Content to the row allocation. The
allocation error `span*r-(H+g)` lies in [0,1) CSS px. Sparse placement may leave
larger visual holes for reading order; the tolerance is on allocated span, not a
promise that all visible vertical gaps are equal. Stack/Grid use ordinary auto rows.

A scheduled frame first computes parent widths/columns in outer-to-inner order,
then measures/assigns spans in inner-to-outer order, skipping hidden roots. Only
changed library-owned values are written. Observe natural measurement boundaries
and container inline sizes; library-owned style/marker writes do not become new
external mutation invalidations. Content, font, visibility and external width changes
invalidate the affected ancestor chain. At most one frame is queued at a time.

For quiescent inputs, settlement means two consecutive passes have identical column
counts, measured natural heights within 0.5 CSS px and identical spans. This must
occur within four animation frames after the final external invalidation. Expose
`data-layout-geometry-state=pending|settled|fallback`; failure to settle uses readable
Grid, marks fallback and emits `layout:error` code `geometry_unstable` once per
transition. A later external invalidation may retry Masonry. Continuous content
animation is not claimed to have a quiescent settlement deadline.

## 7. Implementation acceptance

1. Render mixed Card/Panel/chart/plain content without extra visual card chrome.
2. All modes work at 390/1440 px and within a narrow group in a wide viewport.
3. Nested direct-child containers do not cross-measure or cross-move levels.
4. Alphabetical/custom domain ordering can be applied by the consumer; the library
   never sorts by labels or silently rewrites that canonical order.
5. Pointer and keyboard movement emit equivalent before/accepted event detail once.
6. Cancellation, invalid anchors, no-op drops and incompatible scopes produce no move.
7. Existing links/forms remain operable without initiating a drag. Text selection
   remains available in default handle-only mode; content-drag opt-in replaces it.
8. Body content updates, font changes, visibility and resize preserve usable geometry.
9. No-script/no-observer Masonry uses readable Grid; movement instructions do not
   promise unavailable dragging. Consumers retain explicit server/form alternatives.
10. Root/child replacement and HTMX navigation release observers/listeners and do not
    duplicate effects on remount. Focus, scroll and ARIA updates remain correct.
11. Existing consumers compile without opting in. New public Kind and concrete return
    type follow the component inventory/expression contracts.
12. Generated instruction IDs colliding with another container or explicit item DOM
    ID block every affected root, hide/disable handles and emit one transition error.
    Clearing/replacing the collision restores one owner and correct aria-describedby.
13. Empty destinations under disabled fieldsets reject pointer, keyboard and request
    moves with zero accepted events; destinations inside the first legend remain
    eligible unless another disabled ancestor applies. Ancestor changes, reparenting
    and synchronous before-move callbacks cannot commit using stale availability.
14. Pointer movement keeps a ghost under the pickup point and opens a prospective
    insertion gap, including populated destinations and full-width source items.
    Hover emits no move event. Escape, blur, resize, veto and source removal leave
    no ghost, placeholder or temporary source styles. Reduced motion skips reflow
    animations. Ghost creation does not mount consumer components or duplicate IDs.
15. Goshtoso/Minimal in light/dark pass the meaningful interaction states, visible focus,
    contrast, reduced-motion and console/accessibility checks.

Implementation must include a public leaf demo with mixed/nested examples, browser
coverage of these journeys, generated templ/CSS/runtime outputs, component catalog
registration, `cmd/skillgen` output and the published-consumer/current-source checks.
Follow AGENTS.md. These gates apply to each release; CI supplies execution evidence.

## 8. Integration and release boundary

Goshtoso owns these generic behaviors first. Balemoh owns the API and grouping/order
policy that consume them. No upstream component embeds Balemoh IDs or imports it.
Implementation and tests have passed independent review. A verified upstream
release must be published before Balemoh updates its dependency. An unreleased
branch or this specification does not provide a published module version. The library's existing supported components remain backward compatible.

The first downstream slice excludes tabs, absolute placement, resizing and nested
group editing. Broader editor orchestration is separate application work. Existing
published-consumer compatibility and reachable accessibility failures remain blockers
at the library's v0.1.0+ maturity; hypothetical future scale is not a new promise.

Geometry proof includes nil/short Content, reorder enabled/disabled with a 44px
handle, positive root padding/border/margins and every Gap value. Verify the full
natural-box span allocation error remains below 1px and no neighboring item overlaps.
Reactive-attribute proof rejects owned x-bind/shorthand names, including mixed-case
aliases, while allowed aria-current/busy/expanded leave collection/item names and
instruction associations unchanged after Alpine initialization.

Native disabled proof wraps a populated Masonry container in a fieldset, moves focus
to a consumer status target, disables it during pickup and attempts all move entry
paths. No accepted event occurs; the visible handle row remains measured. Same-root
morph, replacement and remount derive disabled state from current ancestors/config;
reenabling produces one valid interaction owner. No new disabled-state API is exported.

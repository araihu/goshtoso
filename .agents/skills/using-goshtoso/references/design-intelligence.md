# Design intelligence for Goshtoso applications

Use this reference before choosing components when asked to build, redesign, or
extend an application surface. Its purpose is a strong first pass with few
questions, not a generated brand identity.

## Authority order

Resolve design decisions in this order:

1. Existing product identity, content, and semantic tokens.
2. The user's real task, operating context, data, and consequential states.
3. Goshtoso application patterns and exported component contracts.
4. A deliberate visual direction and anti-reflex critique.
5. Browser evidence from the required acceptance matrix.

Never let a product category choose a palette, typeface, layout, or style. A
monitoring tool is not automatically dark, a restaurant is not automatically
warm orange, and an editorial product is not automatically serif.

## First-pass interaction budget

Proceed without an aesthetic clarification when the existing app, content, or
Goshtoso defaults provide a reversible choice. Record the assumption in the
handoff. Ask only when a missing answer changes permissions, destructive
behavior, required data, workflow truth, or a supplied brand contract.

If no identity is supplied, say so in the brief. Use the selected Goshtoso
theme and its system/title stacks; do not add a remote font or invent a brand.

## Write a surface brief first

Before coding, write one compact block:

```text
Primary user and task:
Usage scene and constraints:
Product register: product | brand
Archetype:
Information priority:
Navigation model: document links | HTMX fragments | local state
Consequential states:
Existing identity or explicit no-identity assumption:
Density: compact | standard | relaxed
Motion: none | restrained | expressive with reason
Visual direction:
Chosen Goshtoso pattern and primitives:
```

Keep the visual direction concrete: name hierarchy, density, geometry, type
roles, color role, and motion. “Modern SaaS,” “clean,” “glass,” or a category
name is not a direction.

### Worked route: operational dashboard

```text
Primary user and task: reliability lead deciding which service needs intervention
Usage scene and constraints: desktop control room; scan in under 30 seconds; keyboard actions
Product register: product
Archetype: dashboard + operations list
Information priority: active incidents > degraded services > stable trend context
Navigation model: HTMX fragments with native detail links
Consequential states: loading, no incidents, stale data, partial telemetry, permission denied, retry success
Existing identity or explicit no-identity assumption: no identity; Goshtoso system/title stacks
Density: compact
Motion: restrained; only explain row/status changes
Visual direction: one dense incident queue, restrained dividers, tabular identifiers, semantic state color, no elevation
Chosen Goshtoso pattern and primitives: App Shell + Operations List; PageHeader, Toolbar, Alert, Table, Badge, Panel, EmptyState, Skeleton
No-match/application CSS: compact trend strip; application-owned because no public chart primitive matches
```

This brief makes the incident queue dominant. It does not create one Card per
metric. The trend strip stays subordinate, labels its time range, has a textual
equivalent, and cannot replace the actionable rows.

### Worked route: public evidence page

```text
Primary user and task: prospective partner verifying outcomes and implementation fit
Product register: brand
Archetype: marketing or public content
Information priority: claim > evidence > method > bounded contact action
Navigation model: document links
Consequential states: slow media, missing case study, form validation, submission success
Existing identity or explicit no-identity assumption: supplied wordmark and editorial type roles
Density: relaxed
Motion: none without a supplied brand reason
Visual direction: readable editorial measure, square evidence figures, quiet dividers, one accent action
Chosen Goshtoso pattern and primitives: Navbar, Panel for bounded evidence regions, Button, Link, Alert, Form
No-match/application CSS: publication typography and evidence figure layout
```

Goshtoso supplies durable controls and feedback here; the brand owns typography,
content rhythm, and art direction. A generic hero/features/testimonials/CTA
sequence is not implied.

## Route by task archetype

| Archetype | Composition route | Avoid |
|---|---|---|
| Operations or queue | Operations List; add Detail Workspace when selection needs context | Metric-card gallery before the work queue |
| Dashboard | App Shell + Operations List; make the actionable queue or decision dominant, then add grouped facts or an application-owned trend only when it changes that decision | Equal cards for every number |
| Resource detail | Detail Workspace with a dominant main task and subordinate rail | Two equal columns when importance differs |
| Long, risky, or interruptible task | Multi-step Workflow; add bounded draft and idempotent retry | Wizard chrome around a task that fits one form |
| Settings | App Shell + Page Header + neutral `panel.Panel` groups + Form controls | Rebuilding fields and section panels with raw classes |
| Onboarding | Multi-step Workflow when order matters; otherwise one focused setup page | Promotional feature tour disguised as setup |
| Content review | Detail Workspace or queue + reading surface + decision rail | Turning prose into dashboard cards |
| Marketing or public content | Brand register; content hierarchy and evidence first, components second | Applying product-dashboard density or generic hero/features/CTA reflexes |

The four application patterns are composition contracts, not page templates.
Combine them when the task requires it and keep domain rules in the app.

## Map hierarchy to primitives

- Use `appshell.AppShell` for one header landmark, one scrollable main region,
  skip navigation, and responsive shell ownership.
- Use `pageheader.PageHeader` for page identity and the primary page action.
- Use `toolbar.Toolbar` for controls that change a nearby collection.
- Use `panel.Panel` for a neutral full-width operational, settings, or detail
  surface. An ordinary heading is enough for document structure. When the panel
  must be a named landmark, set both `role="region"` and a matching
  `aria-labelledby`; it intentionally owns no heading rank, article role,
  maximum width, or shadow.
- Use `card.Card` only for genuinely card-like content with its own title and
  article semantics. Do not use Card as a universal bordered rectangle.
- Use `table.Table` for comparable records and preserve native links and row
  actions instead of adding row click handlers.
- Use `form.FieldGroup` for labels, required state, linked errors, and hints.
  `FieldGroup.ID` remains the historical wrapper/HTMX target; give a built-in
  component its own ID or let FieldGroup derive a collision-free one. Populate
  `FormErrorItem.TargetID` with `field.FocusTargetID()` after validation binding
  instead of guessing composite suffixes. `Required` supplies native validation
  where the control supports it and accessible required state for composites;
  the server must still validate Select, Combobox, TagsList, and StructuredInput
  values. For custom children, wire labels, errors, and required state manually.
- Use `skeleton.Skeleton`, `emptystate.EmptyState`, `alert.Alert`, and success
  feedback as first-class states, not late decorations.

When no Goshtoso primitive matches, keep the composition application-owned and
say “no public primitive matched.” Do not silently pretend Card, Tabs, or a
button has different semantics.

## Choose a deliberate direction

Make five bounded decisions:

1. **Hierarchy:** one primary task and one dominant surface; supporting facts
   remain subordinate.
2. **Density:** compact for scanning/repeated action, standard for mixed work,
   relaxed for focused reading or low-frequency settings.
3. **Geometry:** use the theme radius; add borders for grouping and elevation
   only for a real layering relationship.
4. **Typography:** use existing product roles. Keep body measure readable and
   use at most three type roles. Monospace is for identifiers/code, not a
   shortcut for “technical.”
5. **Color and motion:** semantic color communicates action or state. Motion
   explains change or hierarchy and respects reduced motion.

## Anti-reflex critique

Before browser review, remove any unsupported instance of:

- Inter, Geist, Roboto, Georgia, or Times chosen only as a fashionable default;
- gradient text, glass effects, broad diffuse shadows, or cream/parchment as a
  generic “designed by AI” treatment;
- decorative side stripes, excessive rounding, pill-shaped containers, or
  icons placed only to fill space;
- repeated bordered “ghost cards” when rows, sections, dividers, or one Panel
  would express the hierarchy more honestly;
- huge headings, tiny low-contrast body text, or metadata that competes with
  the primary task;
- category-coded palettes or placeholder copy that replaces domain language.

These are rejection checks, not a second aesthetic preset. Keep a choice when
the product context genuinely justifies it and record the reason.

## Required state and recovery pass

Model default, loading, empty, error, and success before polishing. Add
filtered-empty, stale, partial, permission-denied, validation, interrupted,
destructive confirmation, draft recovery, and duplicate-submit prevention when
the domain can produce them.

For every consequential mutation, write the allowed transition table before
writing the handler: starting states, terminal states, authorized reversal,
stale/partial evidence policy, idempotency identity, and the exact effect of
retry or conflict recovery. A warning badge is not a safety policy. Terminal
decisions cannot be overwritten by another terminal decision unless reversal
is a named, authorized workflow; an offered recovery control must not reapply a
notification, cancellation, receipt, or other side effect.

Validation must retain valid values, focus a summary, link summary items to real
controls, associate field errors and hints, and survive the same HTMX swap used
in production. A simulated 422 or 503 must visibly swap or recover; a correct
fragment that HTMX ignores is still a broken interface.

Keep route and display context through every recovery: selected record, filters,
theme, color mode, and valid form values. Render not-found and unknown-record
responses inside the application shell with a route back to known state. For
native consequential POSTs, use Post/Redirect/Get when refresh and Back should
return to a stable document; idempotency remains required even with the
redirect.

## Evidence before handoff

Run `visual-acceptance.md` and report:

- routes and primary journeys tested;
- viewport/theme/mode/state matrix inspected;
- keyboard and focus result;
- console and accessibility result;
- authored application CSS and why it remains application-owned;
- assumptions, explicit no-match decisions, and remaining exceptions.

Rendered browser evidence outranks source inspection. A compile-successful page
with weak hierarchy, inaccessible recovery, or an untested theme is not done.

# Design Intelligence Distillation Audit

Status: in progress

This audit preserves the evidence, decisions, implementation map, and blind-probe
results for distilling UI UX Pro Max and Impeccable into Goshtoso's public
agent-facing surface. It is intentionally versioned early so work survives
context compaction and can be reviewed independently of the implementation.

## Frozen inputs

- Goshtoso: `origin/main` at `5d2e74e4c693ffb17a7443b8b77ed195f815cd05`
- UI UX Pro Max: `nextlevelbuilder/ui-ux-pro-max-skill` at
  `1307d97a72e6c1cda572cb65471ae5ce82995218`
- Impeccable: the ignored repository-local installation under
  `/Users/guilhermecastro/repos/araihu/goshtoso/.agents/skills/impeccable/`.
  Because it is not part of the Goshtoso commit, the consulted files are frozen
  by SHA-256 instead: `SKILL.md` `881f679f4c96b4cb38b6b9a4bfb8d78275fd2b54860c60c313ee094cb7b1a76a`,
  `brand.md` `01296e660b59d6ca91fceb5f5a05af648319707d873c0d3186d5c5ae0dc7f086`,
  `product.md` `9ea8cc99ec208f4c1addc66369980ec6ee2a0ae7732aa8f14672aeed3418b6aa`,
  and `distill.md` `14e45d26cc64a8c6794de0cb2c2e23fcad4850cbd0ef075d848184f101fb6c52`
- Prior Goshtoso agent-quality audits under `docs/audits/`, including the
  37/40 confirmation result and round-two control record

## Objective

Improve the probability that a blind agent produces a coherent, attractive,
accessible Goshtoso application on its first attempt, using only the public
consumer skill and documentation, without needing source dives or repeated
user correction.

The target is not a larger catalog of aesthetic presets. The target is a small,
opinionated decision system that helps an agent choose a product register,
compose the right application pattern, use supported primitives, reject common
AI reflexes, and verify the result before handoff.

## Control record

```yaml
control_record:
  parent_thread_id: 019f9b57-a227-7f63-9797-7e14c9426bf5
  capabilities:
    durable_sessions: {create: available, list: available, read: available, send: available, wait: available, interrupt: unavailable, archive: available}
    subagents: {create: available, send: available, wait: available, interrupt: available}
    native_parallel: {fan_out: available for bounded tool calls, aggregate: parent, max_concurrency: 4}
  dag_ownership_gate:
    - {node: external-mechanism-audit, depends_on: frozen-inputs, exclusive_paths: [], parent_owned_paths: [all repository paths], acceptance: evidence-backed transfer/reject matrix, placement: subagent, merge_order: 1}
    - {node: goshtoso-gap-audit, depends_on: frozen-inputs, exclusive_paths: [], parent_owned_paths: [all repository paths], acceptance: evidence-backed primitive/guidance gap map, placement: subagent, merge_order: 1}
    - {node: blind-probe-contract, depends_on: frozen-inputs, exclusive_paths: [], parent_owned_paths: [all repository paths], acceptance: repeatable first-pass probe rubric, placement: subagent, merge_order: 1}
    - {node: synthesis-and-implementation, depends_on: [external-mechanism-audit, goshtoso-gap-audit, blind-probe-contract], exclusive_paths: [repository], parent_owned_paths: [repository], acceptance: generated artifacts and focused/full gates, placement: local_sequential, merge_order: 2}
  placement_decisions:
    - {node: external-mechanism-audit, primitive: subagent, rationale: bounded read-only source analysis with no durable mutation, capability_requirements: [isolated reasoning, terminal result], lifecycle_owner: parent, fallback: local_sequential, promotion_trigger: any repository mutation}
    - {node: goshtoso-gap-audit, primitive: subagent, rationale: bounded read-only inventory independent of external corpus mechanics, capability_requirements: [isolated reasoning, terminal result], lifecycle_owner: parent, fallback: local_sequential, promotion_trigger: any repository mutation}
    - {node: blind-probe-contract, primitive: subagent, rationale: bounded read-only experimental-design task with frozen prior audits, capability_requirements: [isolated reasoning, terminal result], lifecycle_owner: parent, fallback: local_sequential, promotion_trigger: any repository mutation}
    - {node: synthesis-and-implementation, primitive: local_sequential, rationale: shared docs, skill, generated references, components, and site registries require one writer, capability_requirements: [worktree isolation], lifecycle_owner: parent, fallback: none, promotion_trigger: none}
  dispatch_inputs:
    - external-mechanism-audit prompt with immutable source SHAs and read-only paths
    - goshtoso-gap-audit prompt with immutable source SHA and read-only paths
    - blind-probe-contract prompt with immutable source SHA and prior audit paths
  monitor_registry: []
  completion_protocols:
    persistent_session: callback envelope plus cursor recovery
    subagent: parent-collected terminal result with status, evidence, concerns, and recommended action
    native_parallel: deterministic aggregate of frozen input identities
  acknowledgement_owner: parent
  integration_order: [external-mechanism-audit, goshtoso-gap-audit, blind-probe-contract, synthesis-and-implementation]
  combined_gates: [templ generate, just css, go run ./cmd/skillgen, root and site tests, lints, build, relevant E2E, blind probes]
  cleanup_owner: parent
  lifecycle_ledger: /Users/guilhermecastro/.codex/state/orchestrating-control-planes/lifecycle.json
```

## Initial hypothesis

UI UX Pro Max contributes breadth and retrieval: structured datasets for product
types, interface archetypes, typography, color, motion, charts, UX rules, and
stack-specific advice. Impeccable contributes judgment: product/brand registers,
context preservation, anti-reflex constraints, visual acceptance, and refusal of
generic AI scaffolding. Goshtoso already contributes implementation truth: typed
templ components, semantic theme tokens, server-rendered patterns, HTMX/Alpine
contracts, and a verified acceptance matrix.

The likely missing layer is a compact decision path joining those strengths:

1. infer the interface archetype and user task;
2. pick one composition pattern and information hierarchy;
3. choose a deliberate visual direction without category-default aesthetics;
4. map the composition to real Goshtoso primitives;
5. cover responsive, interaction, data, error, empty, and loading states;
6. run a deterministic quality pass before handoff.

## Evidence ledger

### Upstream corpus and first recommendations

The UI UX Pro Max checkout validates successfully and contains 4,232 CSV rows:
192 product records, 192 color records, 84 style records, 74 typography pairs,
98 web UX rules, 161 category reasoning rules, 34 landing patterns, 29 native
app-interface rules, 25 chart types, 104 icon records, 16 motion records, a
1,923-font lookup, and 22 stack-specific files. Its MIT license permits reuse
with the license notice when substantial source is copied.

Three unmodified design-system queries demonstrate why Goshtoso should reuse
the retrieval shape but not accept its aesthetic result as authority:

| Query | Upstream result | Conflict or miss |
|---|---|---|
| self-hosted monitoring/operations | dark-only OLED, Fira Code headings, glow | Converts a work context into a category costume, drops required light parity, and uses mono as a technical shorthand. |
| API documentation portal | FAQ landing, vibrant block style, dark status palette | Misclassifies the primary task and combines unrelated pattern/style/color matches. |
| coastal Italian restaurant booking | warm amber background, Playfair Display SC, vibrant blocks | Reproduces the hospitality/category reflex that Impeccable explicitly rejects. |

The useful mechanism is structured multi-domain retrieval followed by one
compact decision output. The unsafe mechanism is allowing category similarity
to decide aesthetic identity. Goshtoso's distilled layer must therefore make
task archetype and implementation contracts deterministic while treating
visual direction as a contextual decision with explicit anti-reflex checks.

No upstream dataset, search engine, or generated recommendation is currently
selected for vendoring. The expected first implementation is smaller: a
consumer-facing surface brief, Goshtoso-specific pattern/component mappings,
and executable acceptance checks. If later work copies substantial upstream
source rather than independently expressing the mechanism, its MIT notice must
ship with that copy.

### External mechanism audit

The independent read-only audit reached the same conclusion and added four
material facts:

1. Loading the canonical corpus wholesale would cost roughly 385k tokens. A
   targeted query is fast and compact, so progressive retrieval is worth
   preserving, but only over Goshtoso-native records.
2. Sixty-six of 161 `ui-reasoning.csv` rows contain duplicate JSON keys. The
   upstream validator accepts them while Python parsing silently discards the
   earlier value. Schema validation without semantic golden tests is
   insufficient.
3. A gibberish query still returns Glassmorphism, Inter, SaaS colors, and a
   hero/features/CTA layout because the generator falls back to generic
   defaults. A Goshtoso query must preserve an explicit no-match result rather
   than invent confidence.
4. Goshtoso's generated component reference is already about 95 KB of the
   roughly 120 KB installed skill. A small exact-query index can reduce context,
   but it should be generated from the real Go API and tested with golden
   semantic queries instead of importing the upstream style/font corpus.

| Upstream mechanism | Distilled decision |
|---|---|
| Local top-k retrieval | Adopt only for Goshtoso component/pattern records, with at most three results and explicit source/confidence. |
| Product/category style, palette, and font maps | Reject. They encode the category reflex Impeccable is meant to prevent. |
| Context intake | Adopt: task, users, usage scene, density, existing identity, consequential states, and navigation model. |
| Master plus route overrides | Defer; if added, persist consumer-owned decisions only, never generated identity. |
| Variance, motion, and density dials | Retain density and motion as explicit brief axes; reject automatic variance/style selection. |
| Browser delivery loop | Adopt and specialize to the existing Goshtoso theme/mode/state matrix. |
| Generic browser audit | Adapt only as supporting evidence; approximate contrast or sampled focus is not a WCAG verdict. |
| Multi-platform installer and duplicated corpus | Reject; retain `.agents` plus `npx skills` as the neutral distribution path. |

The source audit recommended a generated compact component index plus a
stdlib-only query helper. This remains a candidate, not yet an implementation
commitment; blind-probe value and added maintenance cost will decide it after
the Goshtoso gap audit.

### Goshtoso gap audit

The independent public-surface audit found no P0 issue and four P1 gaps. Two
are API gaps rather than documentation gaps:

1. **Form recovery is not implementable from the current public contract.**
   `FieldGroup` applies `cfg.ID` to its wrapper while the built-in control can
   receive the same ID, producing duplicate IDs. Labels can therefore target a
   wrapper instead of the real control. Built-in text controls do not derive
   `aria-invalid` or `aria-describedby`, and `FormErrorItem` cannot link a
   summary entry back to a field.
2. **There is no neutral, full-width content surface.** `Card` deliberately
   owns article/title semantics, an `h3`, and content widths. Operational
   panels, settings sections, and detail rails repeatedly reconstruct a flat
   bordered surface with raw classes. The appropriate missing primitive is a
   small `Panel`/`Surface`, not a generic dashboard widget.
3. **The agent route starts too late.** It exposes four implementation patterns
   but does not first distinguish product register, primary task, application
   archetype, deliberate visual direction, and consequential states.
4. **The example contract overstates portability.** Several site examples
   import `site/internal` packages and cannot be extracted by consumers. Only
   `examples/application-patterns` currently passes as a standalone external
   module.

The next tier contains promising but less-proven additions: semantic
description/metric displays, native route-local navigation, a more complete
layout-token contract, and stricter generated-reference documentation checks.
Timeline, icons, and charts remain probe-before-API candidates. Generic
`Stack`/`Grid`, category widgets, chart components, and style/font catalogs are
explicitly deferred.

### Impeccable distillation

Impeccable supplies the judgment layer the upstream recommender lacks:

- preserve an existing product's identity before proposing a new direction;
- distinguish product UI from brand/marketing register;
- use the task, hierarchy, content, and operating context as visual inputs;
- reject category costumes, generic Inter/Geist/Roboto defaults, technical
  monospace reflexes, card soup, ghost cards, excessive rounding, decorative
  side stripes, gradient text, and cream/parchment AI defaults;
- keep one primary goal visible and use progressive disclosure for supporting
  information;
- verify hierarchy, measure, type, responsive behavior, states, keyboard use,
  and theme parity in the rendered product.

The combined rule precedence is therefore:

1. existing consumer identity and semantic tokens;
2. real user task, content, environment, and state model;
3. Goshtoso application patterns and exported component contracts;
4. a deliberate visual direction plus anti-reflex critique;
5. browser evidence across the Goshtoso acceptance matrix.

The external corpus is neither a design authority nor a dependency. Its
useful contribution is the shape of a compact intake, exact retrieval, source
accounting, no-match behavior, and a build-see-review loop.

## Implementation decision

This slice will implement the smallest set with direct P1 evidence:

1. repair form identity, description, invalid-state, and linked-summary
   semantics with rendered-output tests written before the implementation;
2. add a neutral `Panel` primitive only if its minimal contract can replace
   repeated raw surfaces without owning heading rank, width, or card semantics;
3. add a compact design-intelligence reference and route the public skill
   through task/register/archetype/state/visual-direction preflight;
4. make portable-example claims exact and point consumers to the verified
   standalone application-patterns module;
5. update the docs site so the same contract is discoverable without installing
   the skill;
6. regenerate the component reference and verify the public install artifact;
7. run a fresh public-only blind-probe suite before calling the distillation
   successful.

The generated component-query index is deferred in this slice. The main blind
failure mode is currently decision quality and incomplete contracts, not lookup
latency. A compact index becomes justified only if fresh probes exceed ten
public lookups, need more than two API-recovery cycles, or source-dive because
the generated reference is too large to use.

## Blind-probe contract

The new protocol uses three fresh, memory-naive builders in empty standalone
consumer modules. They receive an immutable public-only bundle and may inspect
exported APIs through `go doc`, but not Goshtoso implementation, `site/`,
examples, tests, audits, Git history, other consumers, memory, the internet, or
prior probe artifacts. Post-dispatch product/API/design hints are forbidden;
contamination invalidates rather than qualifies a result.

The three frozen briefs are:

- **Northstar maintenance-change review**, the stable comparison anchor;
- **Library Holds Desk**, a novel operational list/detail domain;
- **Watershed Sample Handoff**, a novel interruption-safe mobile workflow.

Each builder records public lookups, compile/runtime recoveries, authored CSS,
snags, and all verification output. A reviewer who has not read the builder's
report freezes the visual critique before an independent evidence checker runs
DOM, accessibility, console, overflow, contrast, keyboard, state, and theme
checks.

Scoring is 40 observable points: four each for visibility, real-world match,
control/freedom, consistency, error prevention, recognition, efficiency,
minimalist hierarchy, recovery, and help/documentation. Hard gates include
stable generation, tests/vet/build/HTTP checks, local assets, native route and
mutation semantics, unique IDs, linked validation, keyboard completion, zero
critical/serious accessibility findings, no accidental overflow, the full
390/1440 by Goshtoso/Minimal by light/dark matrix, and no unresolved P0-P2.

The round passes only when the Northstar result is at least its historical
37/40 vector, every novel probe is at least 35/40, the median is at least
37/40, no heuristic is below three, and all hard gates pass. A strong pass
requires all three at 37 or higher and a median of at least 38.

## Implementation checkpoint

The P1 implementation now includes:

- a new public `panel.Panel` primitive with outlined, subtle, and plain
  appearances; compact, standard, and relaxed density; arbitrary header,
  actions, body, and footer slots; and target-specific class/attribute hooks;
- no implicit article/section role, heading level, maximum width, shadow, or
  title string in Panel's contract;
- a catalog page with three variants and public Kind/inventory integration;
- form wrapper/control identity separation with backward compatibility for the
  validation package's established distinct wrapper/control IDs;
- `aria-invalid` and merged `aria-describedby` propagation through built-in
  FieldGroup controls, plus standalone TextInput/Textarea helper association;
- focusable, auto-focused `FormErrors` summaries and `TargetID` fragment links
  back to invalid controls;
- target hooks added only where FieldGroup needed to reach real public controls:
  Combobox `TriggerAttrs`, Checkbox/TagsList `InputAttrs`, and StructuredInput
  `RootAttrs`;
- a public `design-intelligence.md` reference with authority order, a compact
  low-interaction surface brief, seven-archetype routing, deliberate direction,
  anti-reflex critique, primitive mapping, state/recovery contract, and evidence
  handoff;
- a broader skill trigger so build/design/redesign requests activate the
  consumer skill, not only dependency-integration requests;
- docs-site copies of the surface-brief route and exact links to design
  intelligence, patterns, and visual acceptance;
- corrected example claims: `site/` applications are runnable demo-site apps;
  `examples/application-patterns` is the verified standalone consumer recipe;
- migration of the standalone recipe's detail main surface, detail rail, and
  workflow body from semantic Card misuse to Panel.

The standalone recipe remains below its 500-line application-CSS budget at 418
lines. Twenty new lines name application-owned panel heading roles; the earlier
398-line value remains the historical pre-Panel checkpoint. A decorative
three-pixel side stripe found by the combined critique was replaced by a
structural one-pixel divider.

### Tests written against the missing contracts

The form tests first failed to compile because `TargetID` did not exist, and the
Panel tests first failed because the package and API did not exist. They now
render the real components and assert:

- globally unique wrapper/control IDs for every built-in FieldGroup type;
- labels target the real input, textarea, combobox trigger, toggle, checkbox,
  tags input, or file input, while StructuredInput receives group labeling;
- errors/hints reach controls through `aria-invalid` and `aria-describedby`;
- summaries are programmatically focusable and link to field targets;
- Panel renders arbitrary regions and templ children without owning document
  semantics, widths, or headings;
- Panel appearances and densities remain bounded.

The root component packages, site non-E2E packages, and standalone external
module pass. The first full root run exposed only the expected inventory-count
guard (`49` packages, `48` docs pages, `80` renderables versus the old
`48/47/79` constants); the guard and README have been advanced with the new
catalog.

### Snag: module mode during in-repo site validation

Running `site/` with `GOWORK=off` resolves the released version pinned in
`site/go.mod`, so it cannot see a newly added in-repo package and was also behind
several existing composition packages. In-repo site gates use the worktree
workspace; `GOWORK=off` portability is exercised in the true standalone example
module, whose local replace points at the candidate root.

### Snag: native validation hid the server response

The first updated empty-submit E2E reported zero server-rendered errors even
though a direct POST returned linked error and hint nodes. The new FieldGroup
contract correctly propagates `Required` to the real control, so Chromium's
native constraint validation stopped the empty submit before HTMX issued a
request. The server-response test now sets `form.noValidate` only for that
scenario; normal rendering retains native required semantics. All form
validation E2Es then passed, including field swaps, dependency updates, value
preservation, error clearing, and the full invalid-submit accessibility checks.

## Independent review checkpoint

Two read-only reviewers inspected the frozen candidate. Panel's public API was
clean, but the first form contract still had three P1 gaps: `FieldGroup.ID`
would have moved an established wrapper/HTMX target to a control, public prose
promised built-in Select without exporting it, and composite required state plus
error-summary focus targets were not reliable. They also found P2 truth gaps in
FileInput description merging, standalone-recipe externalization, dashboard
routing, and Panel landmark wording.

The candidate now:

- preserves `FieldGroup.ID` as the historical wrapper target while deriving
  collision-free built-in IDs; equal wrapper/component IDs are normalized
  without duplicate DOM IDs;
- exposes `FieldGroupConfig.FocusTargetID()` so error summaries do not guess
  Combobox, Select, TagsList, or StructuredInput suffixes;
- adds Select as a first-class FieldGroup built-in with trigger attribute hooks,
  linked helper/error state, accessible required state, and a correctly targeted
  label;
- gives Combobox required state to its real trigger and gives all FieldGroup
  built-ins accessible required state, plus native validation where supported;
- merges FileInput component helper IDs with FieldGroup errors and hints instead
  of emitting competing `aria-describedby` attributes;
- replaces the old form demo's false custom-TagsList example with built-in
  Select and TagsList examples and a `FocusTargetID` recipe;
- adds completed dashboard and public-evidence briefs, an exact named-region
  Panel recipe, a GitHub source link, and copy-out commands that replace the
  repository-local development pin with a reviewed release.

The first full E2E run otherwise completed and exposed only two inventory guards
still fixed at 47 component pages. Both now assert the 48-page catalog, and the
fragment-navigation/sidebar tests pass with Panel included.

### Candidate gate result

After the review remediations, generation remained byte-stable across
`templ generate`, `just css`, and `go run ./cmd/skillgen`. Root tests, site
non-E2E tests, the standalone module with `GOWORK=off`, root/site golangci-lint,
`go vet`, the demo-server build, and `git diff --check` all passed. The final
full browser suite passed in 309.353 seconds. Its preceding run found only
outdated Select E2E selectors that still targeted the hidden submission input;
the tests now assert the public label-to-trigger and helper-ID contract.

### Snag: mobile documentation header and nested table overflow

Manual inspection at 390 px found the documentation header's action group
extending seven pixels past the viewport. The code examples were correctly
contained scroll regions; the actual leak came from keeping the full Goshtoso
wordmark beside the mobile menu, theme selector, and mode control. The wordmark
now collapses below `sm` while the icon and accessible home label remain.

The application-patterns operations preview also wrapped Table in a redundant
48-rem scroll container even though Table already owns its horizontal access.
Removing that wrapper leaves the component's scroll region as the single source
of truth. A 390-pixel E2E assertion now verifies that the shared document body
fits the viewport while the operations table itself remains horizontally
scrollable.

A post-fix full-suite rerun, while three blind builds and browser evidence
captures were also active, timed out only in the unchanged Wizard sidebar
`clickUntil` condition after five attempts. The exact
`TestWizard_SidebarNavNoErrors` test passed immediately in isolation. The frozen
candidate's earlier full suite was green; this is retained as concurrency/flaky
test evidence rather than attributed to the documentation-layout change.

## Blind probe round 1: builder evidence versus independent review

Three agents received only the public skill, public docs, generated component
reference, and exported `go doc` surface from frozen commit
`a99415343fad5c26673bde34875ab1ad79d5b745`. They could compile through a local
replace but could not inspect component source, the demo site, examples, tests,
audits, history, other consumers, memory, or the internet. None asked the user a
question. All three generated, tested, vetted, built, served local assets,
captured the required eight-view matrix, and reported zero critical browser
failures.

Fresh reviewers then froze their visual and interaction critique before reading
the builder report or supplied evidence. This separation exposed false
confidence in each builder-owned harness:

| Probe | Domain and archetype | Builder claim | Independent score | Result |
|---|---|---|---:|---|
| Northstar | reliability maintenance Decision Queue | all gates pass | 29/40 `[3,2,3,4,2,4,2,3,3,3]` | FAIL |
| Riverbend | library holds Decision Queue/detail | all gates pass | 29/40 `[3,4,2,3,2,3,3,4,2,3]` | FAIL |
| Watershed | interruption-safe custody workflow | all gates pass | 24/40 `[3,2,2,3,2,2,3,3,2,2]` | FAIL |

Northstar allowed terminal decisions to overwrite one another, moved focus away
from debounced search after every HTMX swap, allowed approval against stale
evidence that could not actually refresh, and rendered unknown records as bare
404s. Riverbend's offered conflict recovery reapplied a pickup notification,
cancelled records remained in the active queue without undo, its animated skip
link was still offscreen at the instant of focus, and soft semantic badges had
insufficient text contrast. Watershed missed horizontal overflow on its second
step, restored hidden Select values without synchronizing the visible composite
state, contradicted its own temperature boundary, returned success directly
from POST so Back reached `ERR_CACHE_MISS`, and used a sticky action row that
covered content. Its Select inspection also found incomplete live combobox ARIA
and keyboard behavior.

### Round-1 feedback decisions

The first round is not accepted as a pass. It changes the candidate in two ways:

1. Goshtoso-owned component defects are fixed at their source: immediate visible
   AppShell skip focus, contrast-safe soft Badge labels, visible Minimal form
   control boundaries, Select external-value synchronization, and Select
   listbox/keyboard semantics. The Form footer is independently audited for a
   safe 390-pixel sticky/action baseline.
2. Public guidance now requires a mutation transition table, terminal-state and
   stale-evidence policy, two-tab conflict plus side-effect-count tests,
   selective post-settle focus that preserves search/filter caret, native
   constraint attributes, in-shell unknown-route recovery, display-context
   preservation, Post/Redirect/Get for durable success documents, composite
   draft synchronization, and route-by-route 390-pixel evidence instead of an
   initial-page-only matrix.

The probe artifacts remain outside the repository at
`/tmp/gs-blind-northstar`, `/tmp/gs-blind-library`, and
`/tmp/gs-blind-watershed`; the durable scorecard and decisions live here.

### Round-1 remediation implementation

The component fixes were developed and reviewed as isolated worktree slices,
then cherry-picked onto this control-plane branch:

- Select now has one correctly owned listbox, trigger `aria-controls`, live
  `aria-selected`, deterministic wrapping ArrowUp/ArrowDown behavior, trigger
  focus recovery, and visible-state synchronization when a consumer updates the
  hidden public input and dispatches bubbling `input` or `change`.
- AppShell removes motion from the offscreen-to-focused skip link so its first
  focused frame is visible. Soft semantic Badges retain tone in border/tint/dot
  but use the theme's strong surface foreground for normal-size label contrast.
  A `control-outline` semantic token separates functional form boundaries from
  Minimal's intentionally transparent structural outline; Textarea consumes it.
- Form's existing Footer keeps its API but now stacks full-width actions below
  `sm`, wraps long labels, preserves 44-pixel targets, returns to a desktop row,
  and gives sticky mode an opaque semantic surface, `z-20`, safe-area padding,
  and a normal-flow footprint. Arbitrary multi-action or POST-Back workflow bars
  remain an explicit application-owned no-match.

The combined public guidance adds CSP requirements, selective
`htmx:afterSettle` focus, state-transition/idempotency tests, native constraint
parity, unknown-route recovery, display-context preservation,
Post/Redirect/Get, composite draft synchronization, and route-by-route mobile
evidence.

One isolation snag is retained: the Select slice's first compound shell command
created its worktree but did not change the caller's working directory, briefly
cherry-picking the frozen candidate into the primary checkout. The slice owner
detected it immediately, verified the only other item was the user's pre-existing
untracked Impeccable critique, restored primary to `origin/main` with that file
preserved, and performed all actual edits and the final commit in the dedicated
worktree. The control plane independently reverified primary afterward.

## Blind probe round 2: novel domains

The remediated candidate was frozen at `1bf02f5` and given to three new builders
without access to the round-1 applications. Each builder received only the
consumer skill, public Goshtoso documentation, and its assigned domain brief.
Independent reviewers again froze browser and HTTP evidence before reading the
builder report.

| Probe | Domain and archetype | Independent score | Result |
|---|---|---:|---|
| Editorial | multi-state publishing review | 19/40 `[2,2,2,2,0,3,3,2,1,2]` | FAIL |
| Specimen | three-step field collection | 34/40 `[4,4,3,3,3,4,3,3,4,3]` | CHANGES REQUESTED |
| Dispatch | safety-critical railway restriction decisions | 24/40 `[2,2,3,2,4,3,2,3,1,2]` | FAIL |

Editorial contained three P0 authorization/state defects: stale or
Changes-Requested records could still publish, and an error/no-article state
could accept a comment. It also erased drafts on validation and conflict,
silently swallowed network failure, reported malformed mobile submissions as
immediate success, left focus and queue state stale after swaps, hid urgency
below the first viewport, and failed settled helper-text and hover contrast.
Terminal replay and one-dispatch idempotency did pass.

Specimen had no P0 defect and passed its 24-case responsive matrix, draft/Select
agreement, offline and 503 checks, Post/Redirect/Get history, exact replay,
unknown-route shell, local assets, semantics, and clean detached Go gates. Its
mobile sticky footer nevertheless covered `owner_ref` and `collected_at` hit
targets. Settled Goshtoso styles exposed marker contrast of 2.50 in dark mode,
4.33 in light mode, 4.44 small-link contrast, and a 3.23 alert heading. Select
Escape focus was inconsistent, validation order was mixed, cancel/404 recovery
preserved values but returned to step one, and the app emitted a favicon 404 and
unbounded receipt/debug information.

Dispatch's raw browser packet is preserved at
`/tmp/dispatch-r2-blind-browser.hnrE4Q/raw-probe-manifest.json`; full-page
screenshots are explicitly excluded because internal-scroll capture produced
blank images. Viewport screenshots plus DOM geometry are authoritative. The
probe found no horizontal document overflow across its theme and route matrix,
kept exact replay at one side effect, rejected forged stale, partial, and
terminal-alternate actions, preserved the same receipt across completion,
Back, Forward, and refresh, and kept invalid routes inside the application
shell. It also found material workflow gaps: delayed HTMX filtering never set
`aria-busy` or disabled duplicate controls, abort and raw 503 failures exposed
no visible recovery, the sidebar count remained stale after a fragment swap,
missing or malformed versions returned bare HTTP 400 responses, advertised
secondary routes were generic 404 recovery screens, and a mobile row selection
left the detail at y=1382 and its action at y=2008 while scroll remained at zero.
The live shell also showed `14:08 local` while refreshed evidence and decisions
were recorded around `09:11`/`09:12`. Source verification added a P1 permission
failure: the permission-denied fixture still exposed and accepted an
irreversible action because the server never enforced authority. Refresh
idempotency was also decorative, while decision replay correctly retained one
receipt and one actual side effect. Real Tab traversal remains unverified in
this one browser harness because both supported keyboard APIs left focus
unchanged; it is not counted as a pass. Dispatch's exact frozen score is 24/40
with no P0, seven P1, and three P2 findings.

### Round-2 remediation in progress

The public skill now includes an executable state/action invariant ledger.
Every consequential row must prove the server-side permission decision, retained
values and context, focus or destination, final URL or receipt identity, and
exact side-effect count. It requires forged hidden actions, terminal starting
states, real transport failure, loading deduplication, 422 validation, PRG, and
an independent observer; a builder-owned happy-path gallery is insufficient.

The Goshtoso-owned settled-style defects are fixed at their source: helper text
uses semantic muted foreground tokens instead of opacity, button-like hover
states use contrast rather than lowering opacity, non-shell Select no longer
traps focus and keyboard-open advances relative to the selected option, and
Form's stacked mobile footer remains in normal flow while sticky behavior starts
at `sm`. Semantic status text now has contrast-safe `*-text`/`*-text-dark`
tokens, and filled status actions have explicit foreground/background pairs
instead of undefined `danger-dark`-style utilities. Goshtoso's primary link
color also clears AA on `surface-alt`. A focused E2E contrast matrix covers
Button including Danger, Link, Alert status titles, required markers, Textarea,
TextInput, and FileInput in Goshtoso/Minimal light/dark states.

Final acceptance remains blocked until a fresh immutable candidate passes the
full quality gates plus a new blind confirmation probe. The confirmation must
start from the new invariant ledger and prove server authorization, actual
loading/transport failures, retained recovery, mobile selected-detail focus or
scroll, status-action contrast, and refresh idempotency rather than presentation
fixtures.

### Final-confirmation snag and upstream correction

The fresh blind confirmation started from immutable candidate `b274a5c` and,
before source inspection, measured the empty TextInput boundary at 1.64:1 in
Goshtoso light mode. The public guidance successfully made the consumer detect
and temporarily compensate for the defect, but a basic application should not
need application CSS to perceive a form control.

The upstream correction makes TextInput and its search variant consume the
existing semantic `control-outline` pair, gives Goshtoso light/dark and Minimal
dark explicit control boundaries above 3:1, and extends the rendered browser
matrix to assert the TextInput border itself at the non-text contrast threshold.
This is retained as a library snag rather than hidden by the consumer override.
The independent confirmation score remains tied to `b274a5c`; the corrected
candidate must rerun its focused contrast gate and the full repository gates
before merge.

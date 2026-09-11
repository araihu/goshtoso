# Pair A / reviewer 1: Alpine inputs and data components

Reviewed 2026-09-10 at Goshtoso commit `707126f3`, with htmx 4.0.0, Alpine/core plugins 3.17.2, and `hx-alpine-compat`. This is a report for later implementation; no component changes are included. The strongest immediate cleanup is Select's ineffective watcher bookkeeping. The strongest behavioral improvement is restoring client Combobox persistence when a new root arrives through fragment navigation. Whole-component morphing has useful potential, but requires a deliberate distinction between retained UI state and refreshed server data.

**Evidence labels:** confirmed = exercised with current vendored runtime; static = directly established by authored/upstream source; candidate = replacement requires a behavior prototype. P1 means a prerequisite before broad adoption, P2 a useful follow-up, P3 a small cleanup. Priorities are planning priorities, not assertions of production incidents. All source links below are pinned to the reviewed commit.

## Findings

### A1-01 — Remove ineffective Select watcher teardown bookkeeping

**P3 · confirmed · longstanding Alpine behavior.** [select.js:24–49](https://github.com/araihu/goshtoso/blob/707126f3/assets/js/src/components/select.js#L24) stores the return of two `$watch` calls and later invokes entries only if they are functions. This appears intended to prevent parent-model observers surviving component removal. Alpine's `$watch` magic registers its own cleanup and returns nothing; the probe returned `['undefined', 'undefined']`. Remove `modelUnwatches`, its push wrappers, and this `destroy()` implementation, while retaining the watchers themselves until A1-02 is proven. The benefit is simpler ownership with no false impression of manual cleanup.

Risk is accidentally applying this rule to other resources: explicit `AbortController`, timers, observers and listeners still need teardown. Preserve `TestSelectDependentBinding` and `TestSelectExternalDraftRestoration`; add repeated mount/remove coverage that changes the parent model after removal and checks for stale updates/errors. Tagged upstream: [$watch implementation](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/magics/%24watch.js). This is not a new 3.17.2 feature.

### A1-02 — Replace the Select parent-model bridge with `x-modelable`

**P2 · candidate · longstanding Alpine capability.** [select.js:30](https://github.com/araihu/goshtoso/blob/707126f3/assets/js/src/components/select.js#L30) implements two-way binding with paired watches and generated assignment expressions through `Alpine.evaluate`. [Select markup:31](https://github.com/araihu/goshtoso/blob/707126f3/components/select/select.templ#L31) separately watches input/change on its hidden submission control. The original reason is sound: a custom composite control must synchronize its internal selection, enclosing Alpine model, and external draft restoration.

Prototype a scalar `selectedValue` exposed by `x-modelable="selectedValue"` plus `x-model={cfg.Alpine.Model}` on the component root, deriving `selectedOption` from that scalar. This could remove the expression-concatenation bridge and duplicated selected-array state in this single-select component. Keep external hidden-input input/change handling unless the replacement demonstrably preserves that public contract. Do not replace it with `_x_model` internals.

Risks: current initialization only applies a truthy parent value, unknown values clear selection, selection restores trigger focus, and shell mode has a different contract. Native entanglement's initial outer value may change the current empty-parent/server-default precedence. Tests must include empty, undefined and invalid parent values, preselected defaults, dotted parent expressions, two dependent selects, external draft events, keyboard focus, readonly submission, fragment removal and morphing. Preserve existing Select E2E/ARIA tests. Upstream [v3.17.2 x-modelable](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/directives/x-modelable.js) owns entanglement cleanup; [v3.14.9 already had it](https://github.com/alpinejs/alpine/blob/v3.14.9/packages/alpinejs/src/directives/x-modelable.js). Adoption is consolidation, not an upgrade necessity.

### A1-03 — Initialize persisted client Combobox selection per mounted root

**P2 · confirmed gap in a minimal fragment probe; replacement candidate.** [combobox-client.js:199](https://github.com/araihu/goshtoso/blob/707126f3/assets/js/src/components/combobox-client.js#L199) restores session state, but its only initialization entry points are the document-ready scan and [pageshow:248](https://github.com/araihu/goshtoso/blob/707126f3/assets/js/src/components/combobox-client.js#L248). A newly inserted client root misses restoration. The probe seeded `goshtoso:combobox:probe:selected=["a"]`, inserted a root with htmx, and observed no hidden selected input; a synthetic pageshow then restored value `a`.

Move root restoration into a client Combobox Alpine provider's `init()` (or another tightly scoped mount hook), available in the runtime before roots initialize. This replaces full-document first-mount discovery with instance lifecycle and covers fragment navigation. `hx-alpine-compat` coordinates htmx's mutation lifecycle, but does not run application persistence on its own. Preserve a scoped pageshow path where browser-restored DOM still needs reconciliation. Avoid reapplying storage on every body/options swap: server/URL ownership must win when `DisablePersistence` is set.

Keep sessionStorage, its try/catch, and the persistence opt-out. Morph preservation only lasts while a node remains alive; it does not replace session persistence or full navigation behavior. Preserve `TestCombobox_BFCache_RestoresSelection`, `TestComboboxDisablePersistence`, multi-toggle/clear and no-HTTP client tests. Add direct-load versus fragment-entry equivalence, two roots, unavailable storage, malformed storage, and repeated fragment visits without duplicated listeners. Upstream: [Alpine x-data init/cleanup](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/directives/x-data.js), [compat extension](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-alpine-compat.js). The init lifecycle predates this migration.

### A1-04 — Prototype one server Combobox morph instead of body plus label swaps

**P2 · candidate · htmx 4 native morph opportunity.** [combobox.templ:173](https://github.com/araihu/goshtoso/blob/707126f3/components/combobox/combobox.templ#L173) deliberately puts only ephemeral open/focus state in an unswapped Alpine shell. [oob.templ:3](https://github.com/araihu/goshtoso/blob/707126f3/components/combobox/oob.templ#L3) documents the workaround, while [handler.go:92](https://github.com/araihu/goshtoso/blob/707126f3/components/combobox/handler.go#L92) emits a body replacement plus a separate trigger-label OOB update.

Prototype targeting the stable server-mode root with `outerMorph`, rendering the full Combobox once. `hx-alpine-compat` can retain the shell's UI state while the server replaces options, hidden selection values, labels and classes. This could eliminate the bespoke split response and ensure trigger styling is reconciled with selection, rather than requiring a separate label-only response. Keep a small options-only response for active search if it remains cheaper and preserves focus better; full-root rendering is not inherently preferable for every action.

Risks: focusIndex must remain valid after option filtering; stable option identity is needed; the focused search input's live value must agree with server search state; single-select close behavior and cascading invalidation must remain explicit. Client-mode hidden input mutations have different ownership and are excluded from this initial prototype. Preserve `TestCombobox_Toggle_PreservesIsOpen`, lazy search, cascading providers, keyboard/chevron, provider-error retargeting and full/fragment navigation tests. Add open dropdown morphs, focused search selection/caret, concurrent search/toggle, selected-option removal and class/ARIA checks. Upstream: [tagged compat morph hook](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-alpine-compat.js), [htmx 4 changes](https://four.htmx.org/docs/whats-new-in-htmx-4).

### A1-05 — Gate broad morph adoption on server-data reconciliation

**P1 adoption prerequisite · confirmed synthetic probe · not a demonstrated shipped defect.** [Select factory:9](https://github.com/araihu/goshtoso/blob/707126f3/assets/js/src/components/select.js#L9), [StructuredInput:15](https://github.com/araihu/goshtoso/blob/707126f3/assets/js/src/components/structured-input.js#L15), and [Search caches:38](https://github.com/araihu/goshtoso/blob/707126f3/assets/js/src/components/search.js#L38) seed or cache data for the component's lifetime. These are valid under replacement initialization. Morphing retains this data alongside desired UI state.

A minimal Select-root probe changed server configuration label `Alpha` to `Beta` with `outerMorph`. Actual output was `{same:true,label:'Alpha',datasetLabel:'Beta'}`, without page errors. This demonstrates why changing swap strings globally is unsafe. No existing production morph route for this root was established by this review.

For each proposed morph, define which values belong to the server (options, source URLs/configuration), which belong to the user (drafts/selections when permitted), and which are ephemeral UI state. Use explicit reconciliation into the live provider for changed server configuration, narrow the morph boundary, or keep intentional replacement. Do not use unconditional `Alpine.initTree()` to force freshness. Changing `x-data` merely to trigger initialization also risks resetting state the morph was intended to preserve.

Add live-option relabel/removal, changed search source, updated StructuredInput defaults/columns, edited drafts, and server reset tests before broadening morph defaults. The existing generic `TestHTMX4AlpineSwapLifecycle` proves preservation/cleanup mechanics, not these ownership contracts. Upstream: [compat transfers existing data stacks](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/ext/hx-alpine-compat.js), [Alpine clone-aware x-data](https://github.com/alpinejs/alpine/blob/v3.17.2/packages/alpinejs/src/directives/x-data.js).

### A1-06 — Render field-validation metadata as static `hx-vals`

**P3 · static · existing htmx capability used to simplify a migration bridge.** [form.templ:267–282](https://github.com/araihu/goshtoso/blob/707126f3/components/form/form.templ#L267) adds a `name` to the field wrapper and evaluates `this.getAttribute('name')` to identify the validation field. This replaced reliance on the old htmx trigger header, but `cfg.Meta.FieldName` is already known at render time.

Generate static JSON containing `X-Goshtoso-Validation` and `X-Goshtoso-Field`, using Go JSON serialization and ordinary templ attribute escaping. Remove wrapper `name` if no consumer contract depends on it. Keep request parsing in `validation.Handle`; the custom field identity is still needed because the request originates at the wrapper. This removes runtime evaluation and an indirect DOM lookup without changing transport semantics.

Risks: missing Meta, unusual field names, and consumers selecting the wrapper by name. Preserve `TestFormValidation_NativeChangeCarriesFieldIdentity`, dependency validation, field errors and all escaping tests; add metadata absence and quote/newline names. This is not a reason to remove inert data encoding elsewhere. Upstream: [htmx 4 hx-vals JSON form](https://four.htmx.org/attributes/hx-vals).

### A1-07 — Give search and validation explicit request synchronization policies

**P2 · candidate · longstanding `hx-sync`, verified against htmx 4.** [Combobox search:82](https://github.com/araihu/goshtoso/blob/707126f3/components/combobox/combobox.templ#L82) delays input but does not state a latest-result policy; [field validation:267](https://github.com/araihu/goshtoso/blob/707126f3/components/form/form.templ#L267) and form submission use separate source elements. Delay reduces request frequency; it does not establish which overlapping operation wins.

Prototype `hx-sync="this:replace"` for search and a deliberate form/field synchronization policy such as field `hx-sync="closest form:abort"` when submission should supersede validation. This native mechanism is preferable to introducing sequence counters or global cancellation handlers later. There is no existing custom cancellation workaround here to delete, and no race was reproduced in this review; this is a resilience opportunity.

Scope synchronization carefully: a shared abort strategy can discard a second independent field validation, while broad replace can cancel a write. Add controlled delayed responses, search/toggle interaction, dependent-field changes, and validation immediately followed by submit. Assert final values/errors and request cancellation, not just event names. Preserve existing FormValidation and Combobox search suites. Upstream [htmx 4 hx-sync examples](https://four.htmx.org/attributes/hx-sync) explicitly describe search replacement and form/validation races.

### A1-08 — Separate dependent validation response routing from FieldGroup state

**P3 · candidate · htmx 4 partial-response opportunity.** [validation/response.go:22](https://github.com/araihu/goshtoso/blob/707126f3/components/form/validation/response.go#L22) temporarily sets each dependent's shared `FieldGroup.OOB` flag and resets it to false after rendering, including an error path. The reason is to make normal component markup behave as a secondary response update. It adds mutable response-only state to reusable field configuration.

Prototype rendering an ordinary dependent FieldGroup inside `<hx-partial hx-target="#field-wrapper" hx-swap="outerHTML">…</hx-partial>`, escaping the actual wrapper selector correctly. This native response wrapper can remove the temporary flag mutation without changing the reusable component's first-paint HTML. The narrower non-framework alternative is rendering a copied FieldGroup config with OOB set; use that if partial wrappers do not simplify the surrounding response API. Neither approach requires changing ordinary public OOB support elsewhere.

Risks: primary/dependent overlap, absent targets, selector escaping, accidental nested OOB processing and differences in error propagation. Preserve dependency slug updates, primary field value preservation, render cancellation/error tests and unique-wrapper-ID accessibility contracts. Add reuse of the same config before/after successful and failed response rendering, and a multi-dependent response. No partial refactor was implemented or benchmarked. Upstream [htmx 4.0.0 partial-task processing](https://github.com/bigskysoftware/htmx/blob/v4.0.0/dist/htmx.js#L1140) reads the wrapper target/swap and clones its content into separate tasks.

## Retain or defer

- Keep native checkbox/radio/toggle form semantics, readonly/disabled conventions, ARIA relationships and CSS states. These are not Alpine workarounds. Radio's transparent full-segment input addresses browser scrolling on focus; a framework upgrade does not invalidate that reason.
- Keep FileInput's DataTransfer-to-files bridge and bubbling change event. Alpine has no native file-model replacement for that interaction. Keep TextInput's mask plugin and its standalone Alpine root.
- Keep Select's public external draft restoration events, keyboard focus behavior and hidden submission control unless A1-02 proves equivalence. Do not replace its visibility-aware focus logic wholesale with `$focus`: Pair A2 checked the plugin's different visibility semantics.
- Keep schemaform's Go schema walking, default resolution, type selection and native fieldset controls. Keep schematree's native details/summary and caller-owned lazy content. Neither needs framework code for its existing behavior.
- Keep StructuredInput/TagsList hidden serialization and disabled display contracts. Index keys deserve testing before future reorder/morph work, but are not evidence that Alpine 3.17.2 requires a rewrite. Stable identities should be designed before adding reordering, not inferred from editable values.
- Keep Search's ranking, escaped highlighting, safe navigation validation, cached results and fetch abort on destroy. Native htmx HTML requests would change its JSON-source API. The compatibility extension does not make application caches or pending network work obsolete.
- Keep Palette's lazy rendering decision pending a benchmark. Its `x-html` swatch builder avoids shipping the whole grid in initial markup, and uses escaped names plus delegated events. Replacing it with many `x-for`/per-swatch handlers may add runtime work. A renderer consolidation should prove equal lazy output, reduced-motion styles, hover/focus, keyboard activation and escaping first.
- Keep safe Go string/JSON encoding, nil-array normalization, and runtime registration guards. The refreshed stack does not remove those responsibilities.

## Coverage and limits

The sweep inventoried all authored `.templ`, `.go` and JS in these directories, read the interactive markup and implementation paths, and inspected relevant unit/E2E contracts. Generated `_templ.go` files were excluded as generated artifacts; no per-pixel CSS audit or exhaustive review of every test assertion was performed.

| Directory | Reviewed behavior and disposition |
|---|---|
| `components/checkbox` | Native control, descriptions, group rendering, attribute/ARIA merge; retain |
| `components/combobox` | Static/lazy/cascading selection, provider responses, OOB, client persistence and shell state; A1-03/04/07 |
| `components/fileinput` | Dropzone and upload filename state, file event bridge; retain |
| `components/form` and `validation` | Field wrappers, built-in control configs, submit/enter behavior, sections, validation identity/dependents/OOB; A1-06/07/08 |
| `components/palette` | Eager/lazy swatches, model event, hex normalization; retain pending measured consolidation |
| `components/radio` | Standard/segmented/group variants, HTMX and Alpine hooks, keyboard semantics; retain |
| `components/range` | Native slider, value badge and numeric seed sanitization; retain |
| `components/rating` | Native radio groups, numeric state and static display; retain |
| `components/schemaform` | Go walking/defaults, generated native fields and TagsList delegation; retain |
| `components/schematree` | Native disclosures, slots and lazy ownership; retain |
| `components/search` | DOM/JSON sources, cache/filter/ranking, focus and navigation safety; retain, A1-05 gate |
| `components/select` | Model bridge, options, hidden input, shell, focus, external restore, ARIA; A1-01/02/05 |
| `components/structuredinput` | Row append/remove, column defaults, names and hidden serialization; A1-05 gate |
| `components/tagslist` | Add/remove, escaped initial data, indexed submission; retain |
| `components/textarea` | Native attributes, validation relationships and action variant; retain |
| `components/textinput` | Native inputs, search/password/mask variants and escaping; retain |
| `components/toggle` | Native checked/off submission and styling; retain |
| `assets/js/src/components` | Complete input-related `select.js`, `combobox-client.js`, `structured-input.js`, `palette.js`, `search.js`, `data.js`; other shared behaviors belong to partner reports |

Unit suites for all 18 covered packages (including `form/validation`) passed with `-count=1`. Three narrow browser observations ran against the exact vendored scripts: undefined watcher returns, stale Select data under a proposed morph, and client Combobox fragment restoration. [Reproduction scripts and outputs](01-alpine-inputs-probes.md) document their construction. The first morph probe invocation used an incorrect positional `htmx.swap` signature, failed, and was corrected to the documented object argument before collecting the reported result. No full E2E suite, broad network race experiment, implementation prototype or proposed-refactor benchmark was run in this report-only task.

Pair A2 corroborated A1-01 and challenged broad focus/cleanup replacement: only the ineffective watcher bookkeeping should disappear. We agreed that scoped `x-id` can help implicit identities, but explicit/public IDs must survive and existing server namespacing may be the smaller fix. A1-05 was shared with Pair B1 as a gate for its morph recommendations. Pair B1 agreed to retain explicit field identity while simplifying its static encoding and left dependent OOB flag mutation to A1-08. These are cross-review checks, not independent reproductions of the browser probes.

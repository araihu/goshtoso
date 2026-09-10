# Component simplification review after the htmx 4 migration

Review baseline: Goshtoso [`707126f3`](https://github.com/araihu/goshtoso/commit/707126f3), app-shells [`7b3d1f0`](https://github.com/araihu/goshtoso-app-shells/commit/7b3d1f0c7aaccd29138cceafb7d28a516ce1293c). Runtime: htmx 4.0.0, Alpine and collapse/focus/mask 3.17.2, official hx-alpine-compat 4.0.0.

This is a report-only review. Recommendations are proposals for later work, not implemented changes or a blanket authorization to remove safeguards. The preceding migration passed its full browser suite; that establishes its current tested behavior, not proof that every workaround is necessary or every proposed replacement is safe.

## Four reviewer reports

| Pair | Reviewer | Scope | Report |
|---|---|---|---|
| A: Alpine | A1 | Inputs, selection, data binding, validation and composite controls | [Alpine inputs](01-alpine-inputs.md) |
| A: Alpine | A2 | Overlays, navigation, actions, layout and display components | [Alpine overlays](02-alpine-overlays.md) |
| B: htmx | B1 | Component request/response contracts, swaps, concurrency and server fragments | [htmx components](03-htmx-components.md) |
| B: htmx | B2 | Runtime loading, initialization/cleanup, demo streams and app-shells | [htmx lifecycle](04-htmx-lifecycle.md) |

## Recommended next work

The reviewers use local planning priorities; the sequence below reconciles those labels and deduplicates overlapping findings. Confirmed fixture behavior still needs a rendered-component regression before changing production code.

| Batch | Work proposed for later | Evidence and gate |
|---|---|---|
| 1: focused corrections | Single lazy-tab request owner (A2-01/B1-01); navbar keyboard ARIA (A2-02); per-root client Combobox persistence (A1-03) | Reproduced with locked runtimes in narrow fixtures. Add exact request-count, keyboard ARIA/focus and fragment-entry regressions. |
| 2: bounded pruning | Select/site watcher bookkeeping (A1-01/B2-04); old manifest branches (B2-01); static field identity values (A1-06); native table sentinel (B1-02) | Bookkeeping, manifest and static metadata are source-backed simplifications. Sentinel capability was browser-probed, but scroll roots, failures and fallback loading still need parity tests. |
| 3: lifecycle and identity | ActionGroup cleanup (A2-03), toast timer ownership (A2-04), accordion implicit IDs (A2-05); single TOC owner (B2-03); redundant OOB enhancement scans (B2-02) | TOC duplication was fixture-probed; other items are source findings. Test disposal, reduced motion, focus, OOB initialization and repeated instances before consolidation. |
| 4: request contracts | Shared-result synchronization (B1-03/A1-07), native filter payloads (B1-04), concrete form status/disable hooks (B1-05), dependent field partials (A1-08) | Candidates with explicit payload, concurrency, error and mutation semantics to prove. |
| 5: controlled morph adoption | Server Combobox whole-root experiment (A1-04/B1-06); Select modelable bridge (A1-02) | Server-data reconciliation (A1-05/A2-06) is a prerequisite. Preserving Alpine state can preserve obsolete options and configuration too. |

A2-01 and B1-01 describe one duplicated-request issue. A1-04 and B1-06 propose one server Combobox experiment. A1-05 and A2-06 describe the shared morph-adoption constraint. Native morphing, partial responses and the compatibility extension are new htmx 4 opportunities; `$watch` cleanup, `x-modelable`, `x-id` and provider lifecycle are established Alpine capabilities worth using more consistently.

Docs-drawer focus-trap consolidation (B2-05) and declarative streaming listeners (B2-04) are separate later experiments with focus and first-message timing gates.

Retained safeguards include transition-aware focus ownership, explicit browser-resource disposal, safe encoding, JSON-search API behavior, intentional server reset/OOB boundaries and the SSE cancellation-order workaround. Each report explains what evidence would justify changing them.

## Validation during the sweep

A1 ran all 18 scoped unit packages successfully and recorded three narrow browser observations in [its evidence appendix](01-alpine-inputs-probes.md). A2 and B1 independently reproduced the tab duplicate request; A2 also reproduced the navbar ARIA mismatch. The coordinator repeated the A2 fixture with an initial `x-cloak` guard and got the same results. B1 additionally probed native table intersection root/margin handling and row replacement. B2 instrumented a narrow shared-shell fixture and observed two live TOC observers; its report documents the limits of that result. These probes validate specific mechanisms, not whole-component replacement designs. The coordinator verified report source anchors against the baseline files and checked named test references. The full migration suites were green before this documentation-only review; they were not rerun merely to add reports.

## Component coverage

The inventory contains **55 component directories**. Pair A owns exhaustive component coverage; pair B provides a second, cross-cutting review of htmx-bearing components and their integration. Static components are included so “no justified change” is an explicit result rather than an omission.

| Component | Alpine reviewer | Authored templ files | Authored Go files | Go test files |
|---|---|---:|---:|---:|
| `accordion` | A2 | 1 | 2 | 3 |
| `actiongroup` | A2 | 1 | 3 | 1 |
| `alert` | A2 | 1 | 2 | 1 |
| `appshell` | A2 | 1 | 2 | 1 |
| `avatar` | A2 | 1 | 2 | 2 |
| `badge` | A2 | 1 | 2 | 2 |
| `banner` | A2 | 1 | 2 | 4 |
| `breadcrumbs` | A2 | 1 | 2 | 2 |
| `button` | A2 | 1 | 2 | 2 |
| `card` | A2 | 1 | 2 | 1 |
| `carousel` | A2 | 1 | 2 | 4 |
| `chatbubble` | A2 | 1 | 2 | 1 |
| `checkbox` | A1 | 1 | 2 | 2 |
| `codeblock` | A2 | 1 | 3 | 2 |
| `combobox` | A1 | 3 | 5 | 5 |
| `drawer` | A2 | 1 | 2 | 2 |
| `dropdown` | A2 | 1 | 2 | 2 |
| `emptystate` | A2 | 1 | 2 | 1 |
| `fileinput` | A1 | 1 | 2 | 1 |
| `form` | A1 | 2 | 6 | 4 |
| `head` | A2 | 1 | 2 | 4 |
| `icon` | A2 | 1 | 3 | 2 |
| `inlinecode` | A2 | 1 | 2 | 1 |
| `kbd` | A2 | 1 | 2 | 2 |
| `link` | A2 | 1 | 2 | 2 |
| `modal` | A2 | 2 | 3 | 4 |
| `navbar` | A2 | 1 | 2 | 3 |
| `pageheader` | A2 | 1 | 2 | 1 |
| `pagination` | A2 | 1 | 2 | 3 |
| `palette` | A1 | 1 | 2 | 2 |
| `panel` | A2 | 1 | 2 | 1 |
| `popover` | A2 | 1 | 2 | 1 |
| `radio` | A1 | 1 | 2 | 3 |
| `range` | A1 | 1 | 2 | 2 |
| `rating` | A1 | 1 | 3 | 3 |
| `schemaform` | A1 | 1 | 2 | 1 |
| `schematree` | A1 | 1 | 2 | 1 |
| `scrollregion` | A2 | 1 | 1 | 1 |
| `search` | A1 | 1 | 2 | 2 |
| `select` | A1 | 1 | 2 | 4 |
| `sidebar` | A2 | 1 | 2 | 1 |
| `skeleton` | A2 | 1 | 2 | 1 |
| `spinner` | A2 | 1 | 2 | 1 |
| `splitbutton` | A2 | 1 | 3 | 1 |
| `steps` | A2 | 1 | 2 | 1 |
| `structuredinput` | A1 | 1 | 2 | 2 |
| `table` | A2 | 1 | 2 | 7 |
| `tabs` | A2 | 1 | 2 | 2 |
| `tagslist` | A1 | 1 | 2 | 2 |
| `textarea` | A1 | 1 | 2 | 2 |
| `textinput` | A1 | 1 | 2 | 3 |
| `toast` | A2 | 1 | 2 | 4 |
| `toggle` | A1 | 1 | 2 | 1 |
| `toolbar` | A2 | 1 | 2 | 1 |
| `tooltip` | A2 | 1 | 2 | 4 |

The supporting runtime inventory contains 20 authored library JavaScript files and 14 authored site JavaScript files. Generated Go/minified outputs are implementation artifacts, not separate components. B2 additionally covers the component-docs, console and landing shells in app-shells. Individual reports record evidence, tests actually run, retained workarounds, confidence and remaining validation needs.

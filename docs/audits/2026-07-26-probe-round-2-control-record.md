# Goshtoso consumer probe round 2: control record

Date: 2026-07-26

Status: builds and dual assessment complete; remediation pending

This file is the durable control-plane and evidence record for a second blind
consumer round. It intentionally tests whether the current public Goshtoso
surface generalizes beyond the operational Control Room benchmark.

## Frozen experiment boundary

- parent thread: `019f9b57-a227-7f63-9797-7e14c9426bf5`
- branch under evaluation: `codex/agent-quality-improvements`
- immutable base SHA: `34915148826ab3bd508c1856d493ab89a4ff8ec5`
- read-only library snapshot: `/tmp/goshtoso-probes/round2/library`
- probe root: `/tmp/goshtoso-probes/round2`
- public files allowed to guide a probe:
  - `.agents/skills/using-goshtoso/SKILL.md`
  - `.agents/skills/using-goshtoso/references/application-patterns.md`
  - `.agents/skills/using-goshtoso/references/visual-acceptance.md`
  - `.agents/skills/using-goshtoso/references/components-reference.md`
  - `docs/COMPONENT_MODEL.md`
  - `docs/USAGE.md`
  - exported API through `go doc`
- forbidden evidence: `components/**` implementation files, `site/**`,
  `examples/**`, audit reports, Manja, Gostoso, Projeto Manga, Sourceboard,
  Control Room, and `tks-console`.
- context contamination rule: do not search memory. If the harness supplies
  prior Goshtoso facts, declare them, do not open the named artifacts, and keep
  them out of design/API decisions.

## Capability map

| Capability | Runtime capability |
|---|---|
| create bounded worker | `collaboration.spawn_agent` |
| list workers | `collaboration.list_agents` |
| read worker output | callback messages plus `collaboration.wait_agent` |
| send/steer | `collaboration.send_message` / `collaboration.followup_task` |
| archive worker | unavailable for local subagents |
| bounded wait | `collaboration.wait_agent` |

## DAG and ownership gate

| Node | Depends on | Exclusive mutable path | Parent-owned paths | Independent acceptance | Integration order |
|---|---|---|---|---|---|
| approvals probe | frozen SHA | `/tmp/goshtoso-probes/round2/approvals/**` | repository worktree and this report | standalone generate/test/vet/build/HTTP checks | 1, evidence only |
| receiving probe | frozen SHA | `/tmp/goshtoso-probes/round2/receiving/**` | repository worktree and this report | standalone generate/test/vet/build/HTTP checks | 1, evidence only |
| editorial probe | frozen SHA | `/tmp/goshtoso-probes/round2/editorial/**` | repository worktree and this report | standalone generate/test/vet/build/HTTP checks | 1, evidence only |
| synthesis | all three reports | repository report only | library implementation | cross-probe comparison plus independent visual/detector evidence | 2 |
| remediation | synthesis | dedicated Goshtoso worktree | probe artifacts | focused tests plus repository gates | 3 |
| confirmation | remediation | new external probe path | repository worktree | repeat probe against new immutable SHA | 4 |

All dispatch gates are yes: the base and allowed docs are frozen, mutable paths
are disjoint, each app is a standalone module, no child commits or edits the
library, and the parent owns synthesis and all product changes.

## Shared completion envelope

Every child must send this envelope to the parent before its final response:

```yaml
control_plane_completion:
  child_thread_id: <registered child id>
  status: DONE | DONE_WITH_CONCERNS | BLOCKED | NEEDS_INPUT
  branch: null
  base_sha: 34915148826ab3bd508c1856d493ab89a4ff8ec5
  head_sha: null
  pushed_sha: null
  owned_paths: [<exclusive probe path>]
  tests_evidence: [<command plus result or report reference>]
  dependencies_handoffs: [<findings for parent synthesis>]
  conflicts_concerns: [<issue>]
  report_path: <absolute PROBE_REPORT.md path>
  recommended_parent_action: <one concrete action>
```

## Dispatch prompt: approvals probe

```yaml
parent_thread_id: 019f9b57-a227-7f63-9797-7e14c9426bf5
child_thread_id: <pending runtime result>
goal: Build a polished standalone vendor-invoice approval desk using only the frozen public Goshtoso consumer surface, then report every snag and lookup.
base_ref: codex/agent-quality-improvements
base_sha: 34915148826ab3bd508c1856d493ab89a4ff8ec5
owned_paths: [/tmp/goshtoso-probes/round2/approvals/**]
forbidden_paths: [/tmp/gs-agent-quality-improvements/**, /tmp/goshtoso-probes/round2/library/components/**, /tmp/goshtoso-probes/round2/library/site/**, /tmp/goshtoso-probes/round2/library/examples/**]
frozen_inputs: [/tmp/goshtoso-probes/round2/library at 34915148826ab3bd508c1856d493ab89a4ff8ec5, allowed public files listed in this control record]
acceptance: [standalone Go module with replace to frozen snapshot, own templ source, local Goshtoso assets, queue/detail/decision journey, loading/empty/error/success, responsive structure for 390 and 1440, no CDN, templ generate, go test, go vet, go build, HTTP smoke, PROBE_REPORT.md with timing/files opened/source dives/compile recoveries/components/LOC/unmet needs]
git_policy: Do not create branches or commits. Never edit the library snapshot or parent repository.
report_path: /tmp/goshtoso-probes/round2/approvals/PROBE_REPORT.md
```

Scene: a finance reviewer works through vendor invoices on a 14-inch laptop in
a bright office and must approve, reject, or request information without losing
the amount, vendor, policy flags, or audit context. Keep the product restrained,
task-first, and specific. Do not copy an existing Goshtoso app.

## Dispatch prompt: receiving probe

```yaml
parent_thread_id: 019f9b57-a227-7f63-9797-7e14c9426bf5
child_thread_id: <pending runtime result>
goal: Build a polished standalone warehouse receiving application using only the frozen public Goshtoso consumer surface, then report every snag and lookup.
base_ref: codex/agent-quality-improvements
base_sha: 34915148826ab3bd508c1856d493ab89a4ff8ec5
owned_paths: [/tmp/goshtoso-probes/round2/receiving/**]
forbidden_paths: [/tmp/gs-agent-quality-improvements/**, /tmp/goshtoso-probes/round2/library/components/**, /tmp/goshtoso-probes/round2/library/site/**, /tmp/goshtoso-probes/round2/library/examples/**]
frozen_inputs: [/tmp/goshtoso-probes/round2/library at 34915148826ab3bd508c1856d493ab89a4ff8ec5, allowed public files listed in this control record]
acceptance: [standalone Go module with replace to frozen snapshot, own templ source, local Goshtoso assets, purchase-order lookup/line discrepancies/review-submit journey, loading/empty/error/success, mobile-first 390 plus 1440 structure, no CDN, templ generate, go test, go vet, go build, HTTP smoke, PROBE_REPORT.md with timing/files opened/source dives/compile recoveries/components/LOC/unmet needs]
git_policy: Do not create branches or commits. Never edit the library snapshot or parent repository.
report_path: /tmp/goshtoso-probes/round2/receiving/PROBE_REPORT.md
```

Scene: a warehouse receiver uses a tablet one-handed at a bright loading dock,
checks a purchase order, records discrepancies, and submits a receiving review.
Touch targets, interruption recovery, and dense line items matter more than
decoration.

## Dispatch prompt: editorial probe

```yaml
parent_thread_id: 019f9b57-a227-7f63-9797-7e14c9426bf5
child_thread_id: <pending runtime result>
goal: Build a polished standalone editorial planning application using only the frozen public Goshtoso consumer surface, then report every snag and lookup.
base_ref: codex/agent-quality-improvements
base_sha: 34915148826ab3bd508c1856d493ab89a4ff8ec5
owned_paths: [/tmp/goshtoso-probes/round2/editorial/**]
forbidden_paths: [/tmp/gs-agent-quality-improvements/**, /tmp/goshtoso-probes/round2/library/components/**, /tmp/goshtoso-probes/round2/library/site/**, /tmp/goshtoso-probes/round2/library/examples/**]
frozen_inputs: [/tmp/goshtoso-probes/round2/library at 34915148826ab3bd508c1856d493ab89a4ff8ec5, allowed public files listed in this control record]
acceptance: [standalone Go module with replace to frozen snapshot, own templ source, local Goshtoso assets, editorial queue/story workspace/status-edit journey, loading/empty/error/success, responsive structure for 390 and 1440, no CDN, templ generate, go test, go vet, go build, HTTP smoke, PROBE_REPORT.md with timing/files opened/source dives/compile recoveries/components/LOC/unmet needs]
git_policy: Do not create branches or commits. Never edit the library snapshot or parent repository.
report_path: /tmp/goshtoso-probes/round2/editorial/PROBE_REPORT.md
```

Scene: a small nonprofit editor plans stories on a laptop, moves one draft from
pitch through review, and needs author, deadline, channel, and notes visible at
the decision point. This probe must not default to an operations dashboard or a
grid of interchangeable cards.

## Required child protocol

For every probe:

1. Start from an empty owned directory and record start/end timestamps.
2. Read only the allowed public files; list each opened file and `go doc` query.
3. Do not search memory. Declare injected context contamination if present.
4. Log every compiler/runtime error and the information used to recover.
5. Use `apply_patch` for files and preserve all paths outside the owned root.
6. Do not use Browser; the parent performs comparable visual inspection later.
7. Before the final response, use `collaboration.send_message` to send the full
   completion envelope to `/root`. Do this for DONE, DONE_WITH_CONCERNS,
   BLOCKED, and NEEDS_INPUT. If sending is unavailable, begin the final response
   with `CALLBACK_UNAVAILABLE` followed by the same envelope.

## Monitor registry

| Probe | Child id | ID acknowledged | Last event | Expected heartbeat | Callback received | Parent acknowledgement |
|---|---|---|---|---|---|---|
| approvals | `/root/probe_approvals` | yes | done with concerns | 10 minutes | yes | evidence accepted |
| receiving | `/root/probe_receiving` | yes | done with concerns | 10 minutes | yes | evidence accepted |
| editorial | `/root/probe_editorial` | yes | done with concerns | 10 minutes | yes | evidence accepted |
| design review A | `/root/probe_design_review` | yes | blocked: Browser backend unavailable | 12 minutes | yes | clean fallback accepted |
| parent visual fallback | parent | n/a | complete before detector | n/a | n/a | evidence frozen |

### Control-plane snag: late persistence of Assessment A dispatch

The complete Assessment A prompt, ownership, frozen inputs, acceptance matrix
and callback envelope were constructed and sent at runtime, but this durable
record was updated immediately after dispatch instead of before it. The worker
still received every required field and acknowledged its runtime ID before
work. This is a protocol-ordering miss, not an ownership or evidence collision;
it will be counted in the control-plane self-evaluation.

Assessment A owns only
`/tmp/goshtoso-probes/round2/design-review/**`, reads the three completed apps
except their `PROBE_REPORT.md` files, and is forbidden from the library,
repository, audit, memory and deterministic detector. Acceptance requires fresh
browser tabs, 390/1440 inspection, Goshtoso light and Minimal dark,
representative states and primary journeys, separate Nielsen 0-4 scores,
cognitive-load/persona review, prioritized issues, server cleanup and the
shared completion envelope. Its report path is
`/tmp/goshtoso-probes/round2/design-review/ASSESSMENT_A.md`.

## Combined gates after remediation

- generation: `templ generate`, `just css`, `go run ./cmd/skillgen` with no drift;
- static: `git diff --check`, root/site lint, affected standalone lint/vet;
- tests: root, site non-E2E, affected examples, full E2E when shared UI changes;
- consumer: fresh external confirmation module against the new immutable SHA;
- visual: 390 and 1440, Goshtoso light and Minimal dark, no document overflow,
  one `h1`, one skip link, keyboard path, state matrix, and no console errors;
- durable synthesis appended to this file before declaring convergence.

## Probe outcomes

### Editorial planning (`/root/probe_editorial`)

- status: `DONE_WITH_CONCERNS`, callback received and acknowledged;
- duration: 10m04s;
- authored surface: 332 templ, 297 Go including tests, 265 CSS lines;
- public lookups: six allowed files and 13 logged `go doc` queries;
- source dives: zero; injected memory context declared but not searched or used;
- parent recheck: `templ generate` no drift, test/vet/build pass;
- journey: queue states plus inline HTMX status edit, validation and success;
- feedback:
  - `textarea.Config` has no typed `Required`, unlike TextInput;
  - `pageheader.Config` has no title-class hook, forcing two scoped
    `!important` declarations for an editorial heading;
  - AppShell already owns the `<header>` landmark, but this slot boundary was
    easy to duplicate in consumer markup;
  - first setup failed because `go mod tidy` ran before templ generation and
    removed templ; the documented command order should be explicit.
- remaining evidence: parent visual, keyboard, console and accessibility pass.

Full external report at the time of synthesis:
`/tmp/goshtoso-probes/round2/editorial/PROBE_REPORT.md`.

### Vendor invoice approvals (`/root/probe_approvals`)

- status: `DONE_WITH_CONCERNS`, callback received and acknowledged;
- duration: 14m46s;
- authored surface: 384 templ, 667 Go including tests, 343 CSS lines;
- public lookups: six allowed files/ranges and seven logged `go doc` queries;
- source dives: zero; injected memory context declared but not searched or used;
- parent recheck: `templ generate` no drift, test/vet/build pass;
- journey: queue/detail plus approve, reject and request-information
  confirmations, validation, PRG success and filtered-empty;
- feedback:
  - public docs disagree on 47 versus 48 packages and 13 versus 15 themes;
  - the Go 1.26.5 floor surfaced only during module resolution;
  - a button-looking no-JS navigation action needs a semantic GET form because
    Button has no href; Link may cover the visual need but this was not obvious;
  - the documented CSS boundary was clear, but a polished work surface still
    needed 343 app-owned lines.
- remaining evidence: parent visual, keyboard, console and accessibility pass.

Full external report at the time of synthesis:
`/tmp/goshtoso-probes/round2/approvals/PROBE_REPORT.md`.

### Warehouse receiving (`/root/probe_receiving`)

- status: `DONE_WITH_CONCERNS`, callback received and acknowledged;
- duration: 13m52s;
- authored surface: 404 templ, 747 Go including tests, 262 CSS lines;
- public lookups: six allowed files/ranges and 12 logged `go doc` queries;
- source dives: zero; injected memory context declared but not searched or used;
- parent recheck: `templ generate` no drift, test/vet/build pass;
- journey: purchase-order lookup, dense line discrepancies, review, submit,
  edit/discard, cookie interruption recovery and idempotent retry;
- feedback:
  - Go 1.26.5 was again discovered only during module resolution;
  - adjacent templ text after a component call became unrendered child content
    until wrapped in an explicit element;
  - HTMX did not swap intentional 422/503 fragments, so the app mapped expected
    HTMX failures to transport 200 plus `X-Receiving-Status`;
  - 262 lines of CSS repeated semantic application colors because the public
    guidance names utility classes but not the stable theme custom properties;
  - dense editable rows are an application composition problem, not a reason
    to force the display Table component into a form.
- remaining evidence: parent visual, keyboard, console and accessibility pass.

Full external report at the time of synthesis:
`/tmp/goshtoso-probes/round2/receiving/PROBE_REPORT.md`.

## Visual Assessment A

The delegated reviewer correctly returned `BLOCKED` rather than fabricating
browser evidence because its runtime had no Browser backend. It had already
smoke-tested all three apps at HTTP 200, then stopped their servers and
confirmed the ports closed. The parent performed a sequential fallback before
running any deterministic detector. This degrades reviewer independence because
the parent knew the builder reports, but preserves detector independence and
provides real comparable browser evidence.

The complete fallback report is
`/tmp/goshtoso-probes/round2/design-review/PARENT_ASSESSMENT_A.md`.

| App | Nielsen total | Direct evidence summary |
|---|---:|---|
| Ledgerline approvals | 34/40 | polished master-detail and complete request-information flow; one P2 internal horizontal overflow clips the dark error alert |
| Dockline receiving | 37/40 | strongest task flow, safe defaults, review/submit and effective mobile sticky actions; P3 preview/progress polish only |
| Larkspur editorial | 34/40 | distinct editorial voice and successful inline status workflow; one P2 nested mobile queue scroller |

Across 1440x900 and 390x844, Goshtoso light and Minimal dark, the three apps had
zero document-level horizontal overflow, one `h1`, one skip link, usable primary
journeys, and no console errors or warnings. No P0 or P1 visual/usability issue
was found. Assessment A did not use or claim axe/detector output.

## Deterministic Assessment B

Assessment B ran only after Assessment A was frozen. The full report is
`/tmp/goshtoso-probes/round2/design-review/ASSESSMENT_B.md`.

| App | Detector findings |
|---|---|
| Ledgerline approvals | `overused-font`: Inter |
| Dockline receiving | `overused-font`: Inter |
| Larkspur editorial | `overused-font`: Inter; `side-tab` on a rendered editorial blockquote |

The recurring result is significant: three unrelated probes independently
defaulted to Inter, so the public visual-acceptance guidance should actively
discourage reflexive fashionable-font convergence. The editorial side accent is
accepted as P3 context-appropriate blockquote styling, not a repeated card
status tell.

Browser overlay injection could not be performed because the in-app Browser
surface exposes read-only Playwright evaluation and no script/HTML mutation
capability; its only tab capability was `pageAssets`. A rendered-URL detector
attempt also reported missing Puppeteer. The existing real-browser screenshots,
DOM inspection and console checks remain the visibility evidence; no overlay is
claimed. No live server or overlay artifact was created. Application servers
were stopped and ports 8121-8123 were confirmed closed.

## Remediation decision

The round found no P0/P1. The next implementation slice will address the
recurring or contract-level findings that can improve a new blind consumer:

1. publish the minimum Go version and a setup order that generates templ before
   dependency tidying;
2. remove public inventory/theme count drift and lock it with executable tests;
3. document stable semantic CSS custom properties for application-owned styling;
4. document HTMX non-2xx swap behavior, semantic button-like GET navigation,
   AppShell's header landmark, and templ's adjacent-child-content trap;
5. strengthen visual acceptance against reflexive Inter/Geist/Roboto-style font
   choices;
6. evaluate focused typed hooks for `textarea.Required` and PageHeader title
   styling with tests and generated consumer reference updates.

The two visual P2s belong to the frozen probe applications, not the library's
base components, so they remain evidence rather than speculative component
changes.

## Independent remediation review dispatch

The completed approvals worker is reused as a read-only reviewer while the
parent owns all implementation and integration.

```yaml
child_thread_id: /root/probe_approvals
goal: Review the complete uncommitted round-2 remediation diff for correctness, regression risk, public API coherence, documentation accuracy, and generated/test coverage.
worktree: /tmp/gs-agent-quality-improvements
allowed_reads: [git diff, changed source/templates/generated files, public docs/skill references, focused tests]
forbidden_actions: [editing files, generating files, committing, staging, pushing, reading memory]
acceptance: [validate textarea.Required on both primitives, PageHeader title hooks and routing, AppShell landmark wording, dynamic inventory tests, HTMX/navigation/templ/CSS/font guidance, field-proven site section, demo/E2E coverage; report only actionable P0-P3 findings with file anchors]
callback: Send one completion message to /root with status, commands run, findings, and recommendation before the final response.
```

The parent continues generation, tests, lint and browser verification in
parallel. The review is evidence only and owns no mutable path.

## Remediation implementation

The implementation distilled the recurring consumer evidence into public
contracts instead of copying the three pre-remediation probe apps into the
site. Their stable decision structures are useful; their app-owned P2s and
domain-specific code are not suitable as canonical examples yet.

- `textarea.Config.Required` now reaches the native `required` attribute in
  both Textarea entry points, with unit, demo and E2E coverage.
- `pageheader.Config` now exposes `TitleClass` and `TitleAttrs` on the `h1`,
  with a consumer-oriented demo, unit coverage and an E2E element-routing
  assertion.
- AppShell's generated API reference now says that the component owns the
  persistent `header` landmark. Skillgen now collapses complete multiline
  field comments and escapes HTML delimiters for Markdown tables; regression
  tests freeze the full Header, Content and MainAttrs descriptions.
- README, the usage guide and generated consumer references publish the Go
  floor and source-derived inventory: 48 public component packages, 47 docs
  pages, 79 primitives and 15 themes.
- The consumer skill records generation-before-tidy setup, semantic GET
  navigation, expected HTMX error swaps and templ adjacent-text behavior.
- The application-pattern and visual-acceptance references publish stable
  semantic CSS variables, internal-overflow inspection and deliberate product
  typography.
- The docs site adds three compact field-proven extensions: Decision Queue,
  Interruption-safe Workflow and Content-first Review.

## Remediation verification before confirmation

Generation was rerun after all edits: `templ generate` reported zero updates,
`go run ./cmd/skillgen` reproduced both consumer reference mirrors, and
`just css` rebuilt the theme and Tailwind assets. `git diff --check` is clean.

Focused component/docs tests, focused site E2E, root tests, non-E2E site tests,
the site build, and both golangci-lint modules passed. One parent lint command
initially exited because two golangci-lint processes were started concurrently;
the serial rerun reported zero issues. This is recorded as orchestration rework,
not a code finding.

The parent inspected the docs and component demos in the real in-app browser:

- `/docs/application-patterns` renders one `h1` and all three field-proven
  patterns at 1440x900 and 390x844;
- the main content region has zero horizontal overflow in both viewports (the
  intentionally scrollable recipe table is internally wider on mobile);
- `PageHeader` puts `tracking-tight` and `data-heading-voice="editorial"` on
  the custom `h1`, while the page remains responsive;
- the required demo emits `textarea#demo-required[required]`;
- the browser console reported zero errors or warnings.

The local site server was stopped and port 8124 was confirmed closed. The
temporary responsive viewport override and browser tabs were finalized.

## Independent remediation review result

`/root/probe_approvals` returned `DONE_WITH_CONCERNS` after a literal full-diff
review, credential-pattern scan, focused component/docs tests, and site demo
tests. CodeRabbit was attempted but its service rate limit prevented a review.
The reviewer found no runtime or component defect, but reported three public
documentation problems:

1. setup prose said to create the consumer template before the “last two”
   commands, which could be read as after `templ generate`;
2. semantic-token prose incorrectly generalized every dark token to a `-dark`
   suffix even though name placement varies and status tokens are shared;
3. skillgen truncated multiline field docs and left raw HTML tags in Markdown
   tables, already cutting AppShell Content and MainAttrs descriptions.

All three were accepted. The setup instruction now says “before `templ
generate`”; the token guide names representative exact dark tokens and the
shared status rule; skillgen preserves full comments and escapes table HTML.
The generator regression test and AppShell generated-reference contract now
cover the formerly truncated fields. The reviewer made no mutations.

The same reviewer then performed a narrow recheck and returned `CLEAN`: both
generated channels contain complete Markdown-safe Header, Content and MainAttrs
descriptions, focused `internal/skillgen` and `docs` tests pass, and no
actionable anchor remains in the three accepted findings.

## Confirmation probe dispatch

The remediation implementation is frozen at
`0badb7e2dd34702ad36f243ac3ac4f77b0da13df`. A fresh worker receives no parent
conversation and may read only the public consumer surface in the detached
snapshot at `/tmp/goshtoso-probes/round2/confirmation-library`.

```yaml
child_thread_id: /root/probe_confirmation
goal: Build a new standalone maintenance-change review app that proves the recurring round-2 contracts can be discovered and used from public Goshtoso guidance alone.
workspace: /tmp/goshtoso-probes/round2/confirmation
library_snapshot: /tmp/goshtoso-probes/round2/confirmation-library@0badb7e2dd34702ad36f243ac3ac4f77b0da13df
allowed_library_reads: [README.md, docs/USAGE.md, docs/THEMING.md, .agents/skills/using-goshtoso/SKILL.md, references linked directly by that skill]
forbidden_reads: [components source, site source, tests, audit files, prior probe apps or reports, memory]
required_contracts: [discover Go floor before setup, create templ before generate then tidy, local embedded assets, AppShell header content without nested header, PageHeader title hooks, native required Textarea, semantic GET Link appearance, deliberate HTMX non-2xx swap policy, explicit adjacent text wrapper, Goshtoso semantic CSS variables, deliberate non-Inter product typography, one field-proven composition]
required_states: [default, empty, validation error, server error, success, interruption-safe retry or draft recovery]
verification: [templ generate without drift, go test, go vet, go build, HTTP smoke]
deliverable: CONFIRMATION_REPORT.md with exact lookups, source-dive count, authored line counts, snags, commands and PASS or FAIL
forbidden_actions: [editing Goshtoso, branch or commit creation, publishing, memory writes]
callback: Send one completion envelope to /root; parent owns browser, detector, integration and final audit.
```

Confirmation acceptance requires zero forbidden source dives, passing build
gates, no recurrence of the remediated setup/API/guidance defects, no P0/P1,
and browser usability at 1440x900 and 390x844 in Goshtoso light and Minimal
dark. The detector remains sequenced after the parent's visual assessment.

## Final branch gates before confirmation integration

After the reviewed documentation fixes, the parent reran generation and the
complete non-browser quality surface:

- `templ generate`: zero updates;
- `go run ./cmd/skillgen`: both reference mirrors reproduced;
- `just css`: theme and Tailwind assets rebuilt;
- `go test ./... -count=1`: pass;
- site non-E2E tests: pass;
- root and site `golangci-lint run`: `0 issues` in both modules;
- `go fix ./...` in both modules: clean;
- demo server build: pass;
- full `go test ./site/tests/e2e/... -count=1 -timeout 15m`: pass in
  315.465 seconds;
- `git diff --check`: clean.

## Confirmation result and iteration

The fresh worker returned an initial `PASS` after about 18 minutes. It authored
626 non-generated lines (97 CSS), read eight allowed public files, made zero
forbidden source dives, and passed stable generation, test, vet, build, page,
asset, 422, 503, created-success and idempotent-replay smokes. The report is
`/tmp/goshtoso-probes/round2/confirmation/CONFIRMATION_REPORT.md`.

All required recovered contracts were discovered and exercised. The host's Go
1.26.1 was below the documented floor, so the worker selected 1.26.5 before
setup. The resulting Northstar app used Content-first Review, AppShell without
a nested header, PageHeader title hooks, native required Textarea, semantic GET
Link appearance, an explicit `htmx:beforeSwap` policy, wrapped adjacent text,
semantic theme tokens, Georgia/Times product typography, and bounded
interruption recovery.

The worker disclosed one metadata-only isolation snag: a recovery `find`
enumerated unrelated `/tmp` filenames after the assigned directory did not yet
exist. No unrelated content was opened; the source-dive count remains zero.

The parent did not accept the first result blindly. Direct artifact inspection
found a literal residual ```)``` line at the end of the app-owned CSS, and the
report found a genuine public-doc inconsistency: an application-pattern example
used obsolete `badge.Config.Text` while the generated API correctly documents
`Label`. The CSS cleanup follow-up was sent immediately and this control record
was updated afterward, a low-severity late-recording event retained for the
orchestration self-evaluation.

The public badge example and an executable no-regression assertion were fixed
in `fb1755e52f15cdb6700004d10aba6437f4913903`. The latest immutable snapshot is
`/tmp/goshtoso-probes/round2/confirmation-library-v2`.

```yaml
child_thread_id: /root/probe_confirmation
followup_goal: Remove the trailing CSS artifact, point the unchanged consumer app at confirmation-library-v2, rerun generation/test/vet/build and complete HTTP smokes, and update the report.
mutable_path: /tmp/goshtoso-probes/round2/confirmation
library_snapshot: /tmp/goshtoso-probes/round2/confirmation-library-v2@fb1755e52f15cdb6700004d10aba6437f4913903
forbidden_actions: [editing Goshtoso, reading new non-public surfaces, branch or commit creation, memory writes]
acceptance: [clean CSS parse surface, stable generation, all gates and smokes pass, badge example uses Label, report identifies latest snapshot]
callback: Send the revised result to /root before visual assessment dispatch.
```

The follow-up returned `PASS` against v2. The app CSS is now 96 lines and
contains no residual trailer; a regression assertion covers it. Two stable
generation runs, test, vet, build, the full HTTP/state matrix and idempotent
replay pass. The report records 628 authored lines, ten public lookups, zero
forbidden source dives, and the badge snag as resolved in v2.

## Confirmation visual assessment dispatch

Visual judgment is delegated to a fresh read-only reviewer with no builder
conversation. The builder report is frozen before this dispatch and no detector
has run on the confirmation app.

```yaml
child_thread_id: /root/confirmation_visual
goal: Independently judge the Northstar confirmation app's usability and visual quality in a real browser before deterministic detector output exists.
app: /tmp/goshtoso-probes/round2/confirmation
run: PORT=18473 GOTOOLCHAIN=go1.26.5 go run .
target: http://127.0.0.1:18473/review
allowed_reads: [rendered UI, browser DOM, console, app run command only]
forbidden_reads: [builder report, app source, Goshtoso source, prior visual reports, detector output, memory]
matrix: [1440x900 Goshtoso light, 1440x900 Minimal dark, 390x844 Goshtoso light, 390x844 Minimal dark]
journeys: [default review, empty state, validation 422 swap, interrupted 503 swap and safe retry, success]
checks: [hierarchy, composition, product voice, one h1, skip link, document and internal overflow, sticky/focus safety, keyboard landmarks, console errors, state clarity, Nielsen 1-5 scores]
deliverable: /tmp/goshtoso-probes/round2/confirmation/VISUAL_ASSESSMENT.md with screenshots or exact DOM evidence, P0-P3 findings, total score and PASS/FAIL
forbidden_actions: [editing any file, running detector, publishing, memory writes]
callback: Send one completion envelope to /root and confirm port 18473 is closed.
```

The fresh reviewer returned `BLOCKED` rather than fabricating evidence. Its
runtime exposed no in-app Browser: explicit `iab` selection reported unavailable
and browser discovery returned an empty list. All four visual cells and all UI
journeys remain explicitly untested by this delegated assessment. The report is
`/tmp/goshtoso-probes/round2/confirmation/VISUAL_ASSESSMENT.md`; the server was
stopped and port 18473 was confirmed closed.

## Confirmation deterministic assessment B

Assessment B ran only after the blocked visual report was frozen.

- Static Impeccable detection over `page.templ` and `app.css` returned `[]`:
  no registered anti-pattern finding, including no overused-font warning.
- Rendered-URL detection could not run because the detector environment lacks
  Puppeteer. It returned no rendered findings, and none are claimed.
- The detector server was stopped and port 18473 was confirmed closed.

## Current convergence status

The public-contract and functional confirmation is proven against v2: zero
source dives, all required recovered contracts exercised, all build/state gates
passing, a focused 96-line application stylesheet, and a clean static detector.
The confirmation itself found one new public-doc snag and one app-owned artifact;
both were fixed, regression-covered, and revalidated.

The full round is not yet visually proven because Browser was finalized after
the remediation-site inspection and child runtimes have no Browser backend.
The next control-plane turn must run the v2 confirmation app in the root in-app
Browser at the four required viewport/theme cells, freeze that assessment, and
only then decide whether the confirmation app is canonical enough to publish.
Until that pass, the docs site keeps the distilled field-proven patterns rather
than copying the entire app.

## Control-plane checkpoint

Evaluation `353694ab-2683-4cbe-9739-33c2f55fceaf` recorded this orchestration as
`partial`, medium severity, with failure code `browser-backend-unavailable`.
The measured slice used six children, zero callback misses, zero path
collisions, zero integration failures and zero user interventions. Rework count
is two: the serial lint rerun plus the confirmation CSS/doc iteration. One
hidden dependency is the child Browser backend unavailable after the parent had
already finalized its Browser session. One confirmation follow-up dispatch was
also written to this record after dispatch and remains disclosed above.

The five-evaluation summary recommends a control-plane skill review because an
unrelated earlier high-severity incident exists in the rolling window; there is
no repeated current failure code, so this task does not auto-edit the shared
orchestration skill.

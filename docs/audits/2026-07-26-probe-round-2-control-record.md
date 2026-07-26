# Goshtoso consumer probe round 2: control record

Date: 2026-07-26

Status: builds complete; independent assessment pending

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
| design review A | `/root/probe_design_review` | yes | acknowledged and running | 12 minutes | no | pending |

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

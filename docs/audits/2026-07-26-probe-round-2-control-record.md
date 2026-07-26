# Goshtoso consumer probe round 2: control record

Date: 2026-07-26

Status: dispatch prepared

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
| approvals | pending | no | dispatch prepared | 10 minutes | no | pending |
| receiving | pending | no | dispatch prepared | 10 minutes | no | pending |
| editorial | pending | no | dispatch prepared | 10 minutes | no | pending |

## Combined gates after remediation

- generation: `templ generate`, `just css`, `go run ./cmd/skillgen` with no drift;
- static: `git diff --check`, root/site lint, affected standalone lint/vet;
- tests: root, site non-E2E, affected examples, full E2E when shared UI changes;
- consumer: fresh external confirmation module against the new immutable SHA;
- visual: 390 and 1440, Goshtoso light and Minimal dark, no document overflow,
  one `h1`, one skip link, keyboard path, state matrix, and no console errors;
- durable synthesis appended to this file before declaring convergence.

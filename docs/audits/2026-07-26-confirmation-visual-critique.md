# Northstar confirmation visual critique

Date: 2026-07-26

Target: `/tmp/goshtoso-probes/round2/confirmation/page.templ`, served at
`http://127.0.0.1:18473/review`

Library snapshot: `/tmp/goshtoso-probes/round2/confirmation-library-v2` at
`fb1755e52f15cdb6700004d10aba6437f4913903`

## Assessment protocol

Assessment A was performed in the root in-app Browser and frozen in
`/tmp/goshtoso-probes/round2/confirmation/ROOT_VISUAL_ASSESSMENT_A.md` before
the deterministic detector was rerun. Independence is degraded because the
parent had seen an earlier static-detector result in the prior turn, but no
detector output was consulted during the current visual pass.

Assessment B then ran against the primary markup artifact:

```text
node /Users/guilhermecastro/.agents/skills/impeccable/scripts/detect.mjs --json page.templ
[]
```

No rendered-URL detector result is claimed: that detector environment lacks
Puppeteer. Browser inspection, exact DOM measurements, computed styles and live
journeys remain the rendered evidence.

## Verdict

**PASS WITH P2 REMEDIATION, 36/40.** There are no P0 or P1 findings. Northstar
is a coherent Content-first Review surface with unusually good interruption
recovery. Three P2s must be corrected before the whole app is canonical
reference material.

| Nielsen heuristic | Score |
|---|---:|
| Visibility of system status | 4 |
| Match with the real world | 4 |
| User control and freedom | 3 |
| Consistency and standards | 4 |
| Error prevention | 4 |
| Recognition rather than recall | 4 |
| Flexibility and efficiency | 3 |
| Aesthetic and minimalist design | 3 |
| Error recognition and recovery | 4 |
| Help and documentation | 3 |
| **Total** | **36/40** |

## Matrix evidence

| Cell | Result |
|---|---|
| 1440x900, Goshtoso light | Default, empty, live 422, live 503, safe retry and success inspected. One h1; zero document/main/internal overflow; zero console errors or warnings. |
| 1440x900, Minimal dark | Default inspected. One h1 and zero horizontal overflow. Three-column hierarchy remained clear. |
| 390x844, Goshtoso light | Default and empty inspected. Zero document/main overflow; state deck overflowed internally by 191 px and queue by 76 px. |
| 390x844, Minimal dark | Default and decision rail inspected. Same two internal scrollers; form controls remained large and stacked correctly. |

The live HTMX journeys preserved the review note across 422 and 503 fragments.
The 503 fragment made safe retry explicit. Retrying reached `Window approved`
and cleared the textarea. The DOM exposed one h1, a skip link to
`#main-content`, main and navigation landmarks, two complementary rails and no
nested headers. Browser keypress automation did not move focus from `body`, so
this pass does not claim complete keyboard traversal.

## Priority findings

### P2: stacked horizontal scrollers on mobile

At 390 px, the six-state deck and two-item queue render as two consecutive
horizontal scrollers. Native scrollbar tracks are especially prominent in
Minimal dark. Wrap the state deck and stack queue items vertically below
44rem; these small sets do not need carousel behavior.

### P2: Minimal-dark small copy contrast

The 14 px PageHeader description resolves to
`color(srgb .541474 .541577 .541587)` over `oklch(.269 0 0)`, a computed
contrast of **4.383:1**. The source token is Minimal
`--color-on-surface-dark: var(--color-neutral-400)`. Raise the token to clear
4.5:1 against the dark surfaces and add a regression.

### P2: empty-state no-op header action

`View cleared queue` remains visible while already on `/review?state=empty`,
beside a separate `Return to active review` action. Omit the page-level action
for the empty state.

### P3: ghost-card treatment

The primary reading surface combines a 1 px border and a broad
`0 1.25rem 3rem` shadow. Keep the structural border and remove the shadow.

### P3: type guidance can converge on a second monoculture

Georgia/Times is attractive here but not inherently tied to maintenance work.
Public guidance should say explicitly that avoiding Inter, Geist and Roboto
does not mean defaulting to Georgia or Times; typography should express the
actual product domain.

## What is already strong

1. Queue, long-form context and the decision rail separate the work without
   degenerating into a dashboard card grid.
2. Validation and interruption states preserve input, name what happened and
   expose an honest next action.
3. Goshtoso and Minimal retain hierarchy across modes; semantic labels preserve
   meaning independent of color.

## Remediation contract

- Correct all three P2 findings and the ghost-card P3.
- Distill the type and surface lessons into consumer guidance.
- Cut a new immutable Goshtoso snapshot, repoint the unchanged confirmation
  app, rerun generation/test/vet/build and HTTP journeys.
- Repeat the affected 1440/390 Goshtoso-light and Minimal-dark browser cells.

## Snags

- The worktree copy of the Impeccable skill lacked
  `scripts/context.mjs`, `scripts/critique-storage.mjs` and
  `reference/heuristics-scoring.md`. The root used the repository's primary
  checkout copy of the scripts and the scoring/persona material embedded in
  `reference/critique.md` while keeping all feature edits in the dedicated
  worktree.
- Rendered-URL detector execution remains unavailable because its Node
  environment lacks Puppeteer. This did not prevent direct Browser evidence.

Questions were skipped because the findings are concrete and this control-plane
goal already authorizes the remediation and revalidation cycle.

## Remediation implementation checkpoint

The external Northstar app corrected all app-owned findings with focused red to
green regressions:

- the state deck wraps and the two-item queue stacks below 44rem;
- the empty state omits the no-op page-level action; and
- the primary reading surface retains its border without the broad shadow.

`templ generate`, `go test ./...`, the three focused tests, `go vet ./...` and
`go build ./...` pass in the consumer. Evidence is preserved in
`/tmp/goshtoso-probes/round2/confirmation/ROOT_VISUAL_REMEDIATION.md`.

The Goshtoso-side contrast remediation changes Minimal
`--color-on-surface-dark` from `neutral-400` to `neutral-300`. A browser E2E
regression now requires the semantic small-copy token to clear 4.5:1 against
both Minimal dark surfaces. Consumer guidance now rejects reflexive
Georgia/Times convergence and the border-plus-broad-shadow ghost-card formula.

An independent source review exposed a second problem while checking the token:
the docs theme editor manually duplicates each built-in theme and was already
stale. Arctic exported three obsolete white foreground tokens, while Minimal
exported an obsolete outline and the old dark text token. The exported values
are aligned with `all-themes.css`, and a new all-theme contract test prevents
the manual blocks from drifting again.

Pre-snapshot gates are green: stable `templ generate`, `just css`, root and site
unit suites, both `golangci-lint` runs, `go fix` in both modules, site build,
focused contrast E2E and the full E2E suite (`319.452s`). The final verdict
remains open only until the immutable v3 consumer snapshot and affected Browser
matrix are reverified.

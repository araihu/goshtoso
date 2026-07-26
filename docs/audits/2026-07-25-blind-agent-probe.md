# Blind agent consumer probe: Control Room

Date: 2026-07-25

Branch: `codex/blind-agent-probe`

Base: detached worktree at `fa80bbc0eca6f061aa25754c6c2884f257614143`,
verified equal to the local `codex/agent-quality-improvements` ref before the
probe branch was created.

Result: **PASS, 36/40** (target: at least 30/40)

Approximate wall time: **26 minutes** from source loading through final browser
acceptance.

## Executive result

A fresh external Go module named **Control Room** was created at
`/tmp/gs-control-room-probe.ar9nCq`. It used:

```go
replace github.com/araihu/goshtoso => /Users/guilhermecastro/.codex/worktrees/ea2f/goshtoso
```

The temporary app was written from the public consumer skill, its bundled
references, the four permitted public documents, compiler/linter diagnostics,
and the app's own rendered HTML. It was not copied from an existing Goshtoso
application.

The app composes:

- `AppShell`
- `PageHeader`
- `Toolbar`
- `Table`
- `Skeleton`
- `EmptyState`
- `Badge`
- `Button` with `WithAttrs`
- `Card` with `Body`

It exposes method-qualified routes, mounts embedded local assets, models
loading/empty/error/success, and includes a three-step simulated maintenance
workflow whose Back and Continue actions both submit by POST.

The final app passed template generation, tests, vet, lint, and build. Browser
inspection covered the full 390/1440, Goshtoso/Minimal, light/dark matrix and
all four operational states. The console contained zero warnings or errors.

The app and screenshots remain temporary and were not copied into the
repository.

## Heuristic score

| Heuristic | Score | Evidence |
|---|---:|---|
| System visibility | 4/4 | Environment badge, fleet summary, state selector, live result announcement, accessible loading state, and explicit workflow progress all expose current state. |
| Match with the real world | 4/4 | Production services, regions, owners, SLO language, canary strategy, rollback threshold, and maintenance review match an operator's vocabulary. |
| User control and freedom | 3/4 | Retry, Back, Continue, browser Back/Forward, theme/mode controls, and safe simulated confirmation work. Deducted because the mobile sidebar is intentionally hidden and the browser automation surface could not complete a real Tab traversal. |
| Consistency and standards | 4/4 | Goshtoso components supply consistent controls and status treatments; native forms, landmarks, labels, `aria-current`, and method-qualified routing follow platform conventions. |
| Error prevention | 3/4 | Step-one Back is disabled, safety controls default on, review names the exact simulated effect, and POST redirects prevent resubmission. There is no full validation/error-summary cycle because the requested probe did not require persistence. |
| Recognition rather than recall | 4/4 | Stable navigation, visible filters, persistent field labels, step names, status text, and review values keep choices in view. |
| Flexibility and efficiency | 3/4 | Direct state/theme URLs, sticky desktop toolbar, compact table, responsive workflow, and browser history are efficient. Deducted for the large app-owned CSS layer and reduced mobile global navigation. |
| Aesthetic and minimalist design | 4/4 | One dominant work surface, restrained summary, subordinate metadata, no dashboard-card sprawl, and strong responsive hierarchy in both themes. |
| Error recovery | 4/4 | Error state preserves workspace controls and offers Retry; Back preserves the workflow route; browser Back/Forward restores steps; no destructive real effect occurs. |
| Help and documentation | 3/4 | Page descriptions, helper copy, empty/error guidance, and public application-pattern references were sufficient. Deducted for integration sharp edges listed below. |
| **Total** | **36/40** | **PASS** |

## Runtime and browser evidence

### Routes

The temporary `http.ServeMux` registered:

```text
GET  /assets/
GET  /app.css
GET  /{$}
GET  /operations
GET  /workflow
POST /workflow/back
POST /workflow/continue
```

Tests confirmed that wrong-method requests to `/operations`,
`/workflow/back`, and `/workflow/continue` return `405 Method Not Allowed`.

### Assets

The final rendered page referenced only:

```text
/assets/styles.css
/assets/js/runtime/alpinejs-collapse/3.14.9/alpine-collapse.min.js
/assets/js/runtime/alpinejs-focus/3.14.9/alpine-focus.min.js
/assets/js/runtime/alpinejs/3.14.9/alpine.min.js
/assets/js/runtime/htmx.org/2.0.8/htmx.min.js
/assets/js/combobox.js
/app.css
```

`GET /assets/styles.css` and `GET /app.css` returned 200 in the app tests.
There were no remote asset URLs in rendered HTML.

### Responsive/theme matrix

Every final matrix entry had exactly one `h1`, one `main`, the expected
`data-theme`/dark marker, and no page-level horizontal overflow:

| Width | Theme | Mode | Page overflow | Table overflow |
|---:|---|---|---|---|
| 390 | Goshtoso | light | none | owned by table frame |
| 390 | Goshtoso | dark | none | owned by table frame |
| 390 | Minimal | light | none | owned by table frame |
| 390 | Minimal | dark | none | owned by table frame |
| 1440 | Goshtoso | light | none | not needed |
| 1440 | Goshtoso | dark | none | not needed |
| 1440 | Minimal | light | none | not needed |
| 1440 | Minimal | dark | none | not needed |

Temporary screenshots:

```text
/tmp/gs-control-room-probe.ar9nCq/screenshot-390-goshtoso-light-success.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-390-goshtoso-dark-success.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-390-minimal-light-success.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-390-minimal-dark-success.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-1440-goshtoso-light-success.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-1440-goshtoso-dark-success.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-1440-minimal-light-success.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-1440-minimal-dark-success.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-390-goshtoso-light-workflow-step2.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-1440-goshtoso-light-workflow-step2.png
```

The in-app browser's full-page screenshot path produced a blank mobile capture
despite visible DOM and valid element bounds. Viewport screenshots rendered
correctly and were used for the visual matrix; the blank attempt was superseded
and is not cited as acceptance evidence. The additional
`/tmp/gs-control-room-probe.ar9nCq/screenshot-390-goshtoso-light-success-viewport.png`
file is a valid viewport capture retained as a temporary diagnostic.

### State matrix

| State | Browser evidence |
|---|---|
| Loading | Skeleton surface present, `aria-busy="true"`, live text: “Refreshing four production services.” |
| Empty | EmptyState explains what belongs here and offers “New maintenance run.” |
| Error | Alert semantics present, retained search/state controls, Retry action, and “Signal delayed” summary badge. |
| Success | Four table rows, text status plus Badge indicator, count announcement, and explicit Inspect actions. |

State screenshots:

```text
/tmp/gs-control-room-probe.ar9nCq/screenshot-1440-goshtoso-light-loading.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-1440-goshtoso-light-empty.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-1440-goshtoso-light-error.png
/tmp/gs-control-room-probe.ar9nCq/screenshot-1440-goshtoso-light-success.png
```

### Workflow and history

Final real-browser interaction:

```text
Continue:
  POST /workflow/continue
  -> 303 /workflow?mode=light&step=2&theme=goshtoso

Back:
  POST /workflow/back
  -> 303 /workflow?mode=light&step=1&theme=goshtoso

Browser Back:
  step 2 -> step 1

Browser Forward:
  step 1 -> step 2
```

No persistence or external side effect occurs.

### Accessibility and keyboard

A bounded rendered-DOM accessibility check was run on Operations and Workflow.
Both had:

- one `h1` and one `main`;
- heading sequence `h1`, `h2`, `h3` with no skipped levels;
- no duplicate IDs;
- no unlabeled form controls;
- no nameless buttons or links;
- no images missing `alt`;
- no nested interactive elements;
- a skip link as the first focusable DOM control;
- a matching `#main-content` target with `tabindex="-1"`;
- structural current-page markers.

The first browser inspection exposed a nested Button inside a linked Table row.
The external app removed the row link and retained the explicit Inspect action;
the final rendered result contained zero nested interactive elements.

The in-app browser automation surface focused a known button but did not advance
focus when Tab was sent and did not activate the focused control when Enter was
sent. Static focus order, native control semantics, the skip target, focus CSS,
and unique accessible names were verified, but a complete real Tab/Shift+Tab
traversal is **not claimed**. This is recorded as a browser-verification
limitation and is reflected in the heuristic deductions.

The final browser console contained **0 warnings and 0 errors**.

Temporary machine-readable evidence:

```text
/tmp/gs-control-room-probe.ar9nCq/accessibility-audit.json
/tmp/gs-control-room-probe.ar9nCq/browser-final-results.json
```

## Quality gates

Final results:

```text
templ generate
  PASS, updates=1 on the last source change

go test ./... -count=1
  PASS: ok example.com/control-room

go vet ./...
  PASS

golangci-lint run
  PASS: 0 issues

go build -o control-room .
  PASS
```

The first lint run found three unchecked `response.Body.Close()` calls in the
temporary app tests. They were fixed from lint diagnostics and the full gate
was rerun.

## APIs that required extra work

1. **`table.Row.Link` plus `table.Row.Actions`**

   The generated reference documents both fields but does not warn that using a
   whole-row link together with a Button action can create nested interactive
   markup. The problem was visible only in the app's rendered DOM. The safe
   solution was to leave the row unlinked and keep the explicit row action.

2. **`AppShell` skip target focusability**

   The application-pattern reference requires `main` to have
   `tabindex="-1"`, while `AppShell` did not emit it from the initial config.
   `MainAttrs: templ.Attributes{"tabindex": "-1"}` was needed.

3. **Templ conditional ARIA attributes**

   Using `templ.KV` as an `aria-current` value did not emit the expected
   rendered attribute. A small app helper returning `page` or `false` produced
   the correct HTML. This is templ friction rather than a Goshtoso component
   failure, but it is easy for an agent to miss without a rendered-DOM check.

4. **`Button.WithAttrs`**

   The API was flexible enough for `formaction`, `name`, `value`,
   `aria-label`, test hooks, and a boolean `disabled` attribute. The generated
   component reference was sufficient.

5. **`Card.Body` and composition slots**

   `Card.Body`, `PageHeader.Actions`, Toolbar slots, Table Cell components, and
   EmptyState actions were straightforward and accepted normal templ
   components without adapters.

## Utility/CSS findings

No required Goshtoso component selector or documented guaranteed recipe utility
was missing. `min-w-[220px]` worked for the mobile table contract.

The public guidance correctly says that embedded Goshtoso CSS is not a general
consumer Tailwind compiler. To make a branded, polished app without guessing at
unemitted utilities, the probe supplied an app-owned stylesheet:

```text
app.css: 570 lines
```

That is a meaningful adoption cost. The resulting UI was strong, but the probe
did not demonstrate that a similarly branded operational surface can be built
with Goshtoso's embedded stylesheet alone.

## Source protocol and zero-source-dive verdict

### Direct forbidden repository source dive: **PASS — zero**

No file under these paths was opened or searched:

```text
components/
site/
examples/
docs/audits/
```

The target report path was checked only for existence before creation. No
existing audit content was read. No Manja, Manga, tks-console, totvs-work,
GitHub, local pkg.go.dev source, `git log`, `git show`, or historical diff was
consulted.

No `rg`, `find`, `cat`, or `sed` command targeted Goshtoso implementation
source outside the permitted files.

### Perfect blind-context isolation: **QUALIFIED**

Before app construction, the harness injected Mem0 snippets that mentioned
earlier probes and example names. Mandatory agent policy also required two Mem0
searches and one registry lookup, which surfaced similar historical summaries.
Those snippets were not treated as product/API evidence, no linked source was
opened, and no artifact was reused. Nonetheless, a perfectly memory-naive run
was impossible in this harness. This contamination is disclosed rather than
hidden.

The direct zero-source-dive result remains PASS; the stronger claim “the agent
had never seen any prior Goshtoso probe outcome” cannot be made.

## Exact consulted files

Permitted Goshtoso sources:

```text
AGENTS.md
README.md
docs/USAGE.md
docs/COMPONENT_MODEL.md
docs/COMPONENT_API_NAMING.md
.agents/skills/using-goshtoso/SKILL.md
.agents/skills/using-goshtoso/references/application-patterns.md
.agents/skills/using-goshtoso/references/components-reference.md
.agents/skills/using-goshtoso/references/visual-acceptance.md
```

Harness/tool instructions, not used as Goshtoso API evidence:

```text
/Users/guilhermecastro/.codex/plugins/cache/openai-bundled/browser/26.721.31836/skills/control-in-app-browser/SKILL.md
/Users/guilhermecastro/.codex/memories/MEMORY.md
Mem0 decision/task_learning search results injected by the harness
Browser screenshot and viewport capability documentation
```

The Go toolchain also resolved the local module graph through the declared
`replace`; no library source file was opened manually.

## Exact command log

Repository/source consultation:

```bash
rg -n -C 2 'Goshtoso consumer agent skill|agent-quality|blind|using-goshtoso' /Users/guilhermecastro/.codex/memories/MEMORY.md

wc -l .agents/skills/using-goshtoso/SKILL.md && cat .agents/skills/using-goshtoso/SKILL.md

wc -l /Users/guilhermecastro/.codex/plugins/cache/openai-bundled/browser/26.721.31836/skills/control-in-app-browser/SKILL.md && cat /Users/guilhermecastro/.codex/plugins/cache/openai-bundled/browser/26.721.31836/skills/control-in-app-browser/SKILL.md

pwd && git branch --show-current && git status --short --branch && wc -l AGENTS.md && cat AGENTS.md

wc -l README.md docs/USAGE.md docs/COMPONENT_MODEL.md docs/COMPONENT_API_NAMING.md && cat README.md docs/USAGE.md docs/COMPONENT_MODEL.md docs/COMPONENT_API_NAMING.md

rg --files .agents/skills/using-goshtoso/references | sort && for probe_ref in .agents/skills/using-goshtoso/references/*; do wc -l "$probe_ref"; cat "$probe_ref"; done

rg --files .agents/skills/using-goshtoso/references | sort

wc -l .agents/skills/using-goshtoso/references/application-patterns.md .agents/skills/using-goshtoso/references/components-reference.md .agents/skills/using-goshtoso/references/visual-acceptance.md

cat .agents/skills/using-goshtoso/references/application-patterns.md && cat .agents/skills/using-goshtoso/references/visual-acceptance.md

sed -n '1,500p' .agents/skills/using-goshtoso/references/components-reference.md
sed -n '501,1000p' .agents/skills/using-goshtoso/references/components-reference.md
sed -n '1001,1500p' .agents/skills/using-goshtoso/references/components-reference.md
sed -n '1501,1899p' .agents/skills/using-goshtoso/references/components-reference.md

cat README.md
cat docs/USAGE.md
cat docs/COMPONENT_MODEL.md && cat docs/COMPONENT_API_NAMING.md

git rev-parse HEAD codex/agent-quality-improvements && if git show-ref --verify --quiet refs/heads/codex/blind-agent-probe; then echo branch-exists; else echo branch-available; fi && mktemp -d /tmp/gs-control-room-probe.XXXXXX

git switch -c codex/blind-agent-probe
```

Temporary app generation and validation:

```bash
go version && templ version && golangci-lint version && go mod tidy && templ generate && gofmt -w main.go main_test.go

go test ./... -count=1
go vet ./...
golangci-lint run
go build -o control-room .

gofmt -w main_test.go && templ generate && go test ./... -count=1 && go vet ./... && golangci-lint run && go build -o control-room .

gofmt -w main.go && templ generate && go test ./... -count=1 && go vet ./... && golangci-lint run && go build -o control-room .

templ generate && go test ./... -count=1 && go vet ./... && golangci-lint run && go build -o control-room .
```

The last combined gate command was executed three times after the skip target,
heading hierarchy, and `aria-current` fixes. `./control-room` was started five
times, with the preceding process stopped before each rebuilt binary was
started.

Evidence inventory:

```bash
date '+%Y-%m-%d %H:%M:%S %Z' && find /tmp/gs-control-room-probe.ar9nCq -maxdepth 1 -type f -print | sort

wc -l main.go page.templ app.css main_test.go && git status --short --branch

if test -e docs/audits/2026-07-25-blind-agent-probe.md; then echo target-exists; else echo target-available; fi && git status --short --branch
```

The `wc ... && git status ...` command intentionally recorded above was run in
the temporary non-git directory: `wc` succeeded, then `git status` returned
fatal “not a git repository.” No recovery action or source lookup followed.

Browser actions (not shell commands):

```text
Open local Operations route.
Read rendered DOM.
Capture 390 and 1440 viewport screenshots.
Navigate the eight viewport/theme/mode combinations.
Navigate loading, empty, error, and success.
Inspect page and table overflow metrics.
Navigate Workflow steps.
Click Continue and Back with expected navigation.
Exercise browser Back and Forward.
Run bounded rendered-DOM accessibility checks.
Read console warnings/errors.
Read final stylesheet/script URLs.
```

Report-only repository completion commands:

```bash
git diff --check
git diff -- docs/audits/2026-07-25-blind-agent-probe.md
git status --short
git add docs/audits/2026-07-25-blind-agent-probe.md
git commit -m "docs: record blind agent consumer probe"
git status --short --branch
git rev-parse HEAD
```

## Snags

1. Harness memory contamination prevented a perfectly memory-naive probe.
2. The initial combined public-source read exceeded tool output limits; the
   selected files were reread individually or in explicit line chunks until
   complete.
3. The first lint run found three unchecked response-body closes in temporary
   tests.
4. Table row linking plus a Button action created nested interactive markup in
   rendered HTML.
5. AppShell required explicit `MainAttrs` for the documented focus target.
6. Conditional `aria-current` needed rendered-HTML verification.
7. Full-page mobile screenshot output was blank while viewport screenshots were
   valid.
8. The browser automation surface did not perform a complete Tab/Enter
   traversal, so that part of keyboard acceptance remains qualified.
9. One evidence command ran `git status` from the temporary non-git directory
   after `wc`; it failed harmlessly and is preserved above.

## Final assessment

The public Goshtoso adoption surface is sufficient for an agent to build a
coherent external operational application without source diving. The strongest
parts are the generated API reference, task-oriented application patterns,
composition slots, local asset contract, and explicit visual matrix.

The principal opportunities are to document unsafe whole-row-link/action
composition, make the AppShell focus target match the recipe by default, and
reduce the amount of app-owned CSS needed for a polished shell. Subject to the
disclosed harness-memory and keyboard-automation qualifications, the probe
clears the requested quality threshold.

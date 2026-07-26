# Composition components snag journal

Date: 2026-07-25
Base: `origin/main` at `bd8edd1c3d9baa188654b93e7f049dd94d414c69`
Slice: `AppShell`, `PageHeader`, `Toolbar`, `EmptyState`, `Skeleton`, and
`card.Config.Body`

## Snags and source-dives

### The audit was not present in the base branch

The requested audit file did not exist in `origin/main`, the isolated slice's
base. The evidence and approved candidate list were read from the coordinator
worktree at
`/private/tmp/gs-agent-quality-improvements/docs/audits/2026-07-25-agent-quality-audit.md`.
The critique in `/private/tmp/gs-agent-quality-audit/.impeccable/critique/` was
also consulted to confirm the Sourceboard Card-body failure and the repeated
composition contracts. No files in either worktree were changed.

### Public component registration spans several inventories

Adding one public component required source-diving across the root `Kind`
inventory, public-renderable surface tests, the site catalog, the demo registry,
sidebar/search derivation, direct/fragment smoke tests, and both generated skill
references. The site test suite also keeps a separate expected sidebar-section
count. The catalog is the source for navigation and search, but these root
runtime, public-function, and consumer-test inventories must still be updated in
lockstep.

### Full-height AppShell defaults need a constrained demo frame

`AppShell` deliberately defaults to `min-h-screen` and a single scrollable main
region. A component-gallery preview cannot safely override that Tailwind utility
through `RootClass`, because utility precedence is stylesheet-owned rather than
attribute-order-owned. The demo therefore clips the real full-height component
inside a fixed preview frame instead of adding a demo-only public mode.

### Missing-route E2E failures are slow by default

The initial RED run covered all five new routes before registration. Each
missing preview waited for Playwright's 30-second locator timeout, so the useful
404 proof took roughly 150 seconds. After that proof, the focused tests were run
only in the positive state. Future route-first TDD should probe one missing route
or assert the HTTP status before waiting on preview selectors.

### Skill generation keeps only the first field-comment line

`cmd/skillgen` uses the first line of each public field comment. All new field
documentation was kept to one line so the generated component reference would
not publish truncated sentences. This is a generator constraint rather than a
component API requirement.

### Count prose remains outside this slice

The new inventory is 47 component pages and 79 renderable primitives. Existing
count prose in `README.md` and `docs/USAGE.md` still said 42 and 74. Those files
were explicitly excluded from the isolated slice and were updated by the
coordinator after integration. Migration and changelog counts remain unchanged
because they document the earlier release surface.

### CSS generation downloaded its pinned binary

`just css` fetched Tailwind CSS v4.3.0 before rebuilding `assets/styles.css`.
The generated stylesheet includes the new component and demo utilities; no
hand-edited CSS was required.

### The documentation smoke test assumed components never render an h1

The direct-load and fragment-navigation smoke helpers selected `main h1`, which
became ambiguous as soon as `PageHeader` rendered its intended page heading.
The component documentation template already marks its canonical title with
`data-toc-heading`, so the smoke contract now targets that marker and leaves
preview semantics unconstrained.

### Visual inspection caught an escaping AppShell skip link

The first desktop capture showed the keyboard skip link above the constrained
demo frame even though its translated position should have been clipped. The
absolute link had no positioned shell ancestor, so its containing block could
escape the root overflow boundary. A regression assertion and `relative` root
default now keep the link hidden until focus while preserving the real skip
target.

### Integration exposed a private templ helper collision

The isolated component demo and the independently built application recipe both
used `appShellCreateAction` in the same Go package. Each slice compiled alone,
but the integrated package did not. The coordinator prefixed every recipe-local
AppShell helper with `applicationPattern`, regenerated templ output, and reran
the merged-package tests.

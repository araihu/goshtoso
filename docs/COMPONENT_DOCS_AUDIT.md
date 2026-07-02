# Component Docs Audit

Date: 2026-07-02

This audit tracks component demo pages that lag behind the current Goshtoso
documentation pattern: one preview and one code block per variant, followed by
a standalone `demo.APIReference` outside the fragment wrapper. The goal is to
make the site useful for first-time evaluation and returning API lookup without
turning component pages into marketing copy.

## Starting Findings

1. `codeblock`: API props were embedded in `ComponentDemoProps` instead of a
   standalone `demo.APIReference`.
2. `steps`: API props were embedded in `ComponentDemoProps` instead of a
   standalone `demo.APIReference`.
3. `palette`: the API table omitted `LazyWhen`, and the examples did not explain
   the lazy-rendering path used inside dropdown shells. The page also had too few
   practical variants for a broad color-picker component.
4. `drawer`: examples covered right and left placement, but not persistent
   dismissal behavior or HTMX body targeting/open triggers.
5. `tooltip`: the API table had a stale duplicate `Trigger` row where
   `TriggerMode` should have been documented.
6. `toast`: the API table omitted server-rendered action toasts via
   `ActionLabel` and `ActionHTMX`.
7. `select`: the page demonstrated Alpine and large lists, but shell-mode fields
   (`Shell`, `TriggerLeading`, `ValueExpr`, `TriggerLabel`) needed clearer
   standalone documentation.
8. `table`: the API reference compressed infinite-scroll behavior; it needed to
   surface `PaginationConfig.ContainerHeight` and the default page-scroll model.

## Completed Work

- Move CodeBlock and Steps API docs into standalone `demo.APIReference` tables.
- Add the missing Palette, Tooltip, and Toast API rows.
- Add Palette examples for restricted brand hues and compact required-field
  pickers.
- Add Drawer examples for persistent drawers and HTMX-targeted drawer bodies.
- Add a Select shell-mode section and API rows.
- Expand the Table API table with nested pagination, infinite-scroll, and filter
  HTMX fields.
- Regenerate templ output for every edited `.templ` page.

## Verification

- `templ generate`
- `go test ./components/codeblock ./components/steps ./components/palette ./components/drawer ./components/tooltip ./components/toast ./components/select ./components/table -count=1`
- `cd site && go test ./internal/pages/demo/components -count=1`
- Structural docs scan: no component demo pages missing a standalone
  `demo.APIReference`, and no remaining broad weak-docs signals from the audit
  heuristic.
- `git diff --check`

## Future Polish

- Consider splitting Table's single API reference into grouped sections for
  columns, rows, pagination, filters, and HTMX behavior if the page gets a
  dedicated reference-layout component.

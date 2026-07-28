# V11 brand-assets snag journal

Date: 2026-07-27
Worktree: `/tmp/gs-goshtoso-v11-assets`

## External SVG theme propagation required a source dive

The approved v11 marks are standalone SVG documents. Their `surface`, `ink`,
and `signal` values use internal custom properties and switch through
`prefers-color-scheme`; custom properties from an embedding `<img>` do not
provide a dependable `.dark` contract. Inspecting the approved SVG styles and
the site's `darkmode.js` behavior showed that the site toggled only `.dark`.

`css/main.css` now maps the semantic values to Goshtoso theme tokens and sets
`color-scheme: light|dark` on `:root` / `:root.dark`. This lets the approved
external SVGs follow the existing dark-mode toggle without modifying their
geometry. The focused landing E2E test and manual browser inspection cover both
states.

# Modernization task list

Canonical scope and acceptance criteria: [MODERNIZATION.md](MODERNIZATION.md). Status is updated after implementation, focused verification and reciprocal review; retained decisions require evidence.

| Task | Driver | Task | Status | Partner review / evidence |
|---|---|---|---|---|
| A01 | A1 | Select binding | Pending | — |
| A02 | A1 | Client Combobox mount persistence | Pending | — |
| A03 | A1 | Server Combobox morph | Pending | — |
| A04 | A1 | Input morph ownership | Pending | — |
| A05 | A1 | Static field identity | Pending | — |
| A06 | A1 | Dependent validation partials | Pending | — |
| A07 | A2 | Form and Combobox synchronization | Pending | — |
| A08 | A2 | Native table sentinel | Pending | — |
| A09 | A2 | Native table filter payloads | Pending | — |
| A10 | A2 | Shared table request queue | Pending | — |
| A11 | A2 | Form status and disable policy | Pending | — |
| A12 | A2 | Request boundary regression review | Pending | — |
| B01 | B1 | Single lazy Tabs request | Pending | — |
| B02 | B1 | Navbar menu ownership | Pending | — |
| B03 | B1 | ActionGroup resource ownership | Pending | — |
| B04 | B1 | Toast timer lifecycle | Pending | — |
| B05 | B1 | Accordion identity | Pending | — |
| B06 | B1 | ScrollRegion lifecycle | Pending | — |
| B07 | B2 | Runtime manifest pruning | Pending | — |
| B08 | B2 | One active shell controller | Pending | — |
| B09 | B2 | Site Alpine lifecycle | Pending | — |
| B10 | B2 | Docs drawer native focus trap | Pending | — |
| B11 | B2 | Overlay morph and focus boundaries | Pending | — |
| B12 | B2 | Streaming and shell integration guard | Pending | — |

## Combined integration gates

- [ ] All 24 tasks resolved with implementation or tested retention rationale.
- [ ] Pair A reciprocal code review closed.
- [ ] Pair B reciprocal code review closed.
- [ ] Generated templ/JS/CSS/skill and integrity checks clean.
- [ ] Root/site units and lint pass.
- [ ] Current-source and public-pinned site contracts pass.
- [ ] Full Goshtoso browser suite passes.
- [ ] Full app-shells unit/browser suite passes.
- [ ] Final commits and public dependency pins persisted on feature branches.

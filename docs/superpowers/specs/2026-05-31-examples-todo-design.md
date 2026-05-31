# Examples — Todo List (design)

Date: 2026-05-31
Status: approved (design), pending implementation plan

## Goal

Add an `/examples/*` route family to Goshtoso: real, runnable example apps that
showcase the component library in actual use (not isolated component docs). First
app: a classic todo list. The family is built to grow — more examples later.

## Constraints & principles

- Server is **stateless**: per-user todo state lives in a cookie, never in server
  memory (no per-user map to leak/grow).
- Demonstrate the library's real pattern: **HTMX server round-trips returning HTML
  fragments**, templ for rendering, Alpine for local interactivity, Tailwind themes.
- Reuse existing infrastructure (registry, `renderDemo`, `Layout`/`Fragment`,
  theme + dark-mode + fragment-swap nav) rather than a parallel shell.
- Deterministic E2E (project value): drag is not the only reorder path.

## 1. Architecture & layout

**Routes** (mirror `/components/*`):

- `/examples` → index page: cards linking to each example app.
- `/examples/todo` → the todo app.

Both resolve through the **existing demo registry + `renderDemo`**, giving free
fragment-swap navigation, theme selector, and dark mode. A new `/examples/`
HTTP handler trims the prefix and looks up the registry key (`examples`,
`examples/todo`), exactly like `handleComponent` does for `/components/`.

**Sidebar:** add a new **"Examples"** `sidebar.Section` (collapsible) in
`getSidebarSections`, with item "Todo List" → `/examples/todo`. The index page is
reachable from a top-level "Examples" entry / the section header.

**Packages** (separation of concerns, each independently testable):

- `internal/examples/todo/` — pure domain. Types, cookie encode/decode, and all
  state mutations. No `net/http`. Fully unit-tested.
- `internal/pages/demo/examples/` — templ content. Exported `IndexContent()` and
  `TodoContent()` plus the row/list partials. The registry (in package
  `components`) imports this package and references the exported content fns.
- `internal/server/` — thin todo HTTP handlers: read cookie → call domain mutation
  → write cookie → render fragment.

## 2. State model (cookie-backed)

Cookie `gt_todo`, value = `base64url(JSON(State))`. No server memory.

```go
type Todo struct {
    ID       int    // assigned from State.Seq (deterministic, no rand)
    Title    string
    Done     bool
    Priority string // "low" | "med" | "high"
    Due      string // "2006-01-02" or "" (optional)
    Order    int    // explicit sort index, drives reorder
}

type State struct {
    Todos  []Todo
    Filter string // "all" | "active" | "done"
    Seq    int    // next ID to assign
}
```

**Guards:**

- Cap **50 todos** (reject/ignore beyond cap — keeps cookie under the ~4KB limit).
- Title trimmed/truncated to **200 chars**.
- Malformed, oversized, or absent cookie → treated as empty `State` (no panic, no
  500). Bad base64/JSON resets silently.

**Cookie attributes:** `Path=/` (must cover both `/examples/*` and the
`/api/examples/todo/*` endpoints — a narrower path would not be sent to the API),
`HttpOnly`, `SameSite=Lax`, `MaxAge` ≈ 30 days.

**Security:** cookie holds only the user's own todos; titles are rendered through
templ (auto-escaped) → no HTML/JS injection. No signing required (tampering only
affects the tamperer's own view).

**Domain functions** (pure, `internal/examples/todo`):

- `Decode([]byte) (State, error)` / `Encode(State) string`
- `(s *State) Add(title, priority, due string)` — assigns ID from `Seq`, appends
  with next `Order`, respects 50-cap + 200-char truncation.
- `(s *State) Toggle(id int)`
- `(s *State) Delete(id int)`
- `(s *State) Edit(id int, title, priority, due string)`
- `(s *State) Reorder(ids []int)` — reassigns `Order` from the given id sequence;
  ignores unknown ids, keeps any omitted todos at the end in prior order.
- `(s *State) ClearCompleted()`
- `(s *State) SetFilter(f string)`
- `(s State) Visible() []Todo` — applies `Filter`, sorted by `Order`.
- count helpers (total, active, done).

## 3. Endpoints & HTMX flow

All under `/api/examples/todo/`, all **POST**, each: read cookie → mutate → set
updated cookie → render an HTML fragment.

| Endpoint          | Action                                   | Response |
|-------------------|------------------------------------------|----------|
| `/add`            | create from form (title, priority, due)  | new `<li>` appended + OOB count badge + OOB toast |
| `/toggle?id=`     | flip `Done`                              | swapped `<li>` (outerHTML) + OOB count |
| `/delete?id=`     | remove todo                              | empty body (htmx removes row) + OOB count + OOB toast |
| `/edit?id=`       | update title/priority/due                | swapped `<li>` (outerHTML) |
| `/filter?f=`      | set filter                               | full `<ul>` list re-render |
| `/clear-completed`| drop all `Done`                          | full `<ul>` + OOB count + OOB toast |
| `/reorder`        | apply id order (`ids=3,1,2`)             | 204 No Content + OOB count (UI already moved optimistically) |

**OOB swaps:** live count badge updated via `hx-swap-oob`; action feedback toast
reuses the existing toast OOB pattern (`handleToastOOB`). Mirrors the established
table/toast precedents.

**Components showcased:** Card (app shell), Text Input + Button (add form), Select
(priority), Checkbox (toggle done), Badge (priority + live count), Tabs (filter
All/Active/Done), Toast (action feedback), Tooltip (icon buttons), Modal (optional
edit dialog), plus a styled empty-state.

## 4. Reorder mechanism

**Decision: native HTML5 drag-and-drop + up/down buttons.**

- A per-row drag handle uses native `draggable` wired via Alpine
  (`dragstart`/`dragover`/`drop`) to reorder the list in the DOM, then fires
  `htmx.ajax('POST', '/api/examples/todo/reorder', { values: { ids } })` to persist.
- Each row also has **↑ / ↓ buttons** (plain HTMX POSTs) — accessible, and the
  **deterministic** path exercised by E2E.
- No new vendor dependency (no SortableJS).

## 5. Testing

**Unit** — `internal/examples/todo/*_test.go`, no browser:

- encode/decode round-trip; malformed/oversized/absent cookie → empty State.
- 50-todo cap; 200-char title truncation.
- Add/Toggle/Delete/Edit/Reorder/ClearCompleted/SetFilter behavior.
- `Visible()` applies filter + Order sort; `Reorder` handles unknown/omitted ids.
- `Seq` monotonic; IDs never reused.

**E2E** — `tests/e2e/todo_example_test.go` (random free port, shared browser):

- add → row appears, live count increments.
- toggle checkbox → done styling, count updates (read Alpine live value via
  `Evaluate`, not the static attribute).
- delete → row removed.
- filter tabs All / Active / Done.
- clear-completed.
- reorder via **↑/↓ buttons** (deterministic). Native DnD persistence covered by a
  direct `/reorder` endpoint test rather than flaky browser drag simulation.
- post-swap clicks use the **`clickUntil`** helper (HTMX rebind race — see
  CLAUDE.md and the `e2e-suite-flaky-full-run` memory).
- sidebar: "Examples" section present; Todo nav swaps `#main-content`.

Full suite stays green; no skipped tests.

## Out of scope

- Persistence beyond the cookie (no DB).
- Multi-user / accounts / sharing.
- Additional example apps (this spec is todo-only; the route family is built to
  accept more later).

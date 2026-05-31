---
target: examples/todo
total_score: 30
p0_count: 0
p1_count: 2
timestamp: 2026-05-31T05-31-08Z
slug: internal-pages-demo-examples-todo-templ
---
## Critique — `/examples/todo` (product register)

### Design Health Score

| # | Heuristic | Score | Key Issue |
|---|-----------|-------|-----------|
| 1 | Visibility of System Status | 4 | Live "N active" badge + toasts + active filter state. Strong. |
| 2 | Match System / Real World | 4 | All/Active/Done, Clear completed, priorities — familiar vocabulary. |
| 3 | User Control and Freedom | 3 | Delete is immediate with no undo (only an info toast). |
| 4 | Consistency and Standards | 2 | Native `<input type=date>` and glyph buttons break the component vocabulary used by TextInput/Select/Button. |
| 5 | Error Prevention | 2 | Delete and Clear-completed have no confirm/undo; a full-cookie add silently no-ops with no feedback. |
| 6 | Recognition Rather Than Recall | 4 | Everything visible; labels clear. |
| 7 | Flexibility and Efficiency | 3 | Drag OR ↑/↓ is good; no keyboard shortcuts, no bulk ops. |
| 8 | Aesthetic and Minimalist | 3 | Clean and themed, but text-heavy intro + dense 7-element row + single-card wrap read a little generic. |
| 9 | Error Recovery | 2 | No error states; mutations can't fail visibly; no undo. |
| 10 | Help and Documentation | 3 | New intro explains the app well; no inline hint that rows are draggable. |
| **Total** | | **30/40** | **Solid; clear, fixable gaps in consistency + error prevention** |

### Anti-Patterns Verdict

**LLM assessment:** Does not scream "AI made this" — it's restrained, themed (inherits Minimal), and uses real components. But it fails the *product* slop test in two spots: the unstyled native date picker and the text-glyph buttons (↑ ↓ ✕ ⠿) are exactly the "subtly-off component" a user fluent in Linear/Things would pause at. The single bordered panel wrapping the whole app + a two-paragraph preamble is the generic-tool silhouette.

**Deterministic scan:** `detect.mjs` crashed (bundled rules missing) — ran a manual ban scan instead. Found: **2 em-dash violations in my copy** (`todo.templ:107`, `:138`) — the "No em dashes" ban. No gradient text, no glassmorphism, no identical card grids in the todo markup. `border-l-2` and one `backdrop-blur` belong to the sidebar/toast layout chrome, not this feature.

**Visual overlays:** Not available — no interactive browser/injection in this session. Reviewed captured rendered HTML (seeded + empty) + source.

### Overall Impression
A correct, legible task tool that works. The biggest opportunity is component consistency: the add row mixes first-class Goshtoso fields with a raw browser date input and ASCII glyph buttons, which undercuts the whole point of an example showcasing the library. Fix that and the screen jumps from "fine" to "trustworthy."

### What's Working
- **System status is genuinely good:** the OOB live count + per-action toasts + active-filter highlight give continuous feedback — the strongest part.
- **Two reorder affordances:** native drag for mouse users, ↑/↓ for everyone else (and for deterministic testing). Thoughtful.
- **Empty state teaches** ("add your first task above") instead of a dead "nothing here".

### Priority Issues

- **[P1] Em dashes in copy.** `todo.templ:107` and `:138`. Direct ban violation. **Fix:** replace with periods/colons ("Nothing here yet. Add your first task above."; split the intro sentence). **Command:** `clarify`.
- **[P1] Native date input breaks component vocabulary.** The `<input type=date>` is raw-styled while title/priority use TextInput/Select. In an example *showcasing the component library*, this is the worst place to drop a browser-default control. **Fix:** wrap it in the same field chrome (label + bordered control matching TextInput), or render a Goshtoso-styled date field. **Command:** `extract` / `craft`.
- **[P2] Destructive actions have no guard or undo.** Delete and Clear-completed mutate immediately. Delete fires an info toast but no undo; Clear-completed can wipe many rows silently. **Fix:** add an Undo action to the delete/clear toast (re-POST prior state), or a lightweight confirm on Clear-completed. **Command:** `harden`.
- **[P2] Row is dense with sub-44px touch targets.** Per row: drag handle, checkbox, title, due, badge, ↑, ↓, ✕ — eight elements; the glyph buttons are `px-1`. Native HTML5 drag also doesn't work on touch, so mobile users get only the tiny ↑/↓. **Fix:** group secondary actions (overflow/hover-reveal), enlarge tap targets to ≥40px, ensure ↑/↓ are the primary mobile reorder. **Command:** `adapt` / `layout`.
- **[P3] "Clear completed" always present + silent full-cookie no-op.** The button shows even with zero completed; adding past the size budget silently does nothing. **Fix:** disable/hide Clear-completed when done-count is 0; surface a toast when an add is rejected. **Command:** `harden`.

### Persona Red Flags

**Alex (power user):** No keyboard path to focus the add field or toggle a row; moving an item far up means many ↑ clicks (drag helps only on desktop). No bulk complete/delete.

**Jordan (first-timer):** The ⠿ drag handle is visually cryptic (aria-label exists, but nothing signals "drag me"); priorities have no legend explaining what low/med/high change.

**Sam (touch / mobile):** Native DnD doesn't fire on touch, so drag silently does nothing on a phone; the ↑/↓/✕ glyphs at `px-1` are below the 44px tap minimum.

### Minor Observations
- Intro "Demonstrates:" line is a long middot run-on; tighten or bullet it.
- Whole app sits in one bordered panel (`bg-surface-alt`) — acceptable for a focused tool, but consider letting the form/list breathe without the wrapper.
- Glyphs (↑ ↓ ✕) instead of a consistent icon set read slightly unpolished next to the library's components.

### Questions to Consider
- Should an example app model *best-practice* destructive-action UX (undo), since people will copy it?
- Is the raw date input acceptable in the one place whose job is to showcase components?
- What would the row look like if it carried three actions instead of six?

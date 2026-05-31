---
target: examples/todo
total_score: 38
p0_count: 0
p1_count: 0
timestamp: 2026-05-31T06-03-12Z
slug: internal-pages-demo-examples-todo-templ
---
## Re-critique — `/examples/todo` (after fixes)

Follow-up to the 30/40 baseline. All chosen issues addressed and test-covered (full E2E suite green, 179s; lint clean).

### Design Health Score

| # | Heuristic | Score | Change | Note |
|---|-----------|-------|--------|------|
| 1 | Visibility of System Status | 4 | = | live count + toasts + list-full warning |
| 2 | Match System / Real World | 4 | = | |
| 3 | User Control & Freedom | 4 | +1 | undo-on-delete restores item to original position |
| 4 | Consistency & Standards | 4 | +2 | due date now native `<input type=date>` rendered through TextInput, matching the other fields |
| 5 | Error Prevention | 4 | +2 | Clear-completed disabled when 0 done; list-full rejection now surfaces a toast |
| 6 | Recognition > Recall | 4 | = | |
| 7 | Flexibility & Efficiency | 3 | = | still no keyboard shortcuts / bulk ops (P3) |
| 8 | Aesthetic & Minimalist | 4 | +1 | intro trimmed to a scannable bullet list; denser, consistent 40px-target rows |
| 9 | Error Recovery | 4 | +2 | undo affordance + rejected-add feedback |
| 10 | Help & Documentation | 3 | = | clear intro; no inline "drag to reorder" hint |
| **Total** | | **38/40** | **+8** | **Excellent** |

### What changed
- **Copy:** removed both em dashes (ban); intro is now a 5-item bullet list instead of a run-on.
- **Consistency:** added `TypeDate`/`TypeDateTimeLocal` to TextInput; the due field renders as a native date picker with the same label+border chrome as Task and Priority. No custom date component (native input is the correct, accessible affordance).
- **Touch/mobile:** action buttons are 40px tap targets (denser on `sm+`); the drag handle is hidden on touch (native DnD does not fire there) so mobile relies on the always-visible ↑/↓; long titles truncate instead of overflowing.
- **Destructive actions:** delete now shows an inline Undo bar that restores the task (id/order/state preserved) via a `/restore` endpoint; Clear-completed is disabled when nothing is done; a full list surfaces a "List full" toast instead of silently dropping the add.

### Remaining (P3, not addressed by design)
- No keyboard shortcuts / bulk complete (Flexibility).
- ASCII glyph buttons (↑ ↓ ✕ ⠿) rather than a consistent icon set.
- Undo bar's 8s auto-dismiss depends on Alpine initializing the OOB-swapped node; functionally the Undo button works regardless, and any later action clears the bar.

### Process note
Found and fixed two issues the baseline run had masked: a missing Tailwind rebuild (new row utilities produced no CSS) and a first-run cookie-notice banner that occluded the row controls and was silently failing the reorder E2E. The earlier "full suite green" claim was inaccurate; the suite is now genuinely green and independently re-run.

# Adversarial acceptance for Goshtoso applications

Use this reference before implementing a consequential mutation or a workflow
that can become stale, invalid, partial, interrupted, or permission-limited. A
list of states is not a test plan. Convert product rules into an invariant
ledger, then derive browser and HTTP checks from every row.

## Write the invariant ledger

Copy this table into the consumer repository. Fill one row for every visible
action in every state where the action can be attempted, including forged or
stale requests that the UI does not normally offer.

| Route or state | Action or request | Allowed? | Expected response | Input and context preserved | Focus or destination | Effect count |
|---|---|---:|---|---|---|---:|
| review / ready | Publish | yes | durable success receipt | article, filter, theme | receipt for this article | 1 |
| review / stale | Publish | no | in-shell conflict recovery | draft and selected article | conflict heading | 0 |
| review / changes requested | Publish | no | in-shell policy error | draft and selected article | error summary | 0 |
| comment / transport failure | Add comment | retryable | visible retry fragment | complete comment draft | retry control | 0 |
| receipt / already published | repeated Publish | idempotent | same receipt | article identity | same receipt | still 1 |

Use concrete domain states and effects. “Error handled” and “success shown” are
not testable. Record the exact object, redirect or swap target, retained values,
focus target, idempotency key, and externally observable side-effect count.

## Derive tests from the ledger

The test list is the ledger, not a hand-picked happy path. For each row:

1. Arrange the server state independently of the UI. Include stale tabs,
   terminal records, missing records, expired permissions, and partial data.
2. Attempt the real action or equivalent forged request.
3. Assert the HTTP status and HTMX swap behavior. A correct 422 or 503 fragment
   that is not visible is a failure.
4. Assert server truth after the response. A hidden or disabled button does not
   replace authorization and transition guards.
5. Assert retained valid values, selected record, filters, theme, color mode,
   and composite visible/hidden values. For a collection/detail surface, the
   URL, detail identity, visible selected style, and `aria-current` or
   `aria-selected` must all identify the same record after the swap settles.
6. Assert `document.activeElement`, the final URL or fragment identity, and the
   exact effect count.

Every consequential action must have at least one allowed row, one denied or
stale row when the domain permits it, and one repeated request. Terminal states
must be tested as starting states, not only rendered as badges.

## Recovery and transport gates

- Submit loading disables or otherwise deduplicates the initiating action while
  preserving an announced progress state.
- Native and server validation agree. A 422 retains valid values, links the
  summary to real controls, and focuses the summary after the swap settles.
- A real transport failure is visible inside the application shell. Silent
  network failure is a hard failure even if a simulated server 503 works.
- Conflict recovery cannot bypass a terminal-state rule or repeat a side
  effect. Exercise two stale tabs and the offered recovery control.
- Durable native POST success uses Post/Redirect/Get. Refresh and Back resolve
  to stable documents; a success query or receipt identifies the object that
  actually changed. Compare the Back-restored revision and status with a fresh
  server read; stale-write guards prevent duplicate effects but do not make a
  cached Available or pending form truthful.
- Unknown IDs and routes retain the shell and provide a route back to known
  state.

## Browser evidence contract

Drive the real rendered app, then save a compact evidence manifest containing:

```text
ledger rows tested / total:
denied transitions and server truth:
422 retained values and focus:
transport failure and retry:
two-tab conflict and effect count:
repeated terminal request and effect count:
success URL or fragment identity:
390 px first-viewport priority and sticky clearance:
1440 px hierarchy:
Goshtoso and Minimal, light and dark contrast:
keyboard path and final focus:
console and accessibility findings:
```

At 390 px, inspect every step and recovery route. Assert the recorded primary
information appears in the first viewport, internal rails do not create a
second horizontal scroller, and sticky actions leave the last focused control
fully visible. At every theme/mode combination, measure normal text, helper
text, control boundaries, and default/hover/focus action contrast.

## Completion gate

Do not declare acceptance when any consequential ledger row is untested, a
denied transition changes server truth, a retry can duplicate an effect, a
transport error is silent, or a P0/P1 accessibility or workflow defect remains.
Builder self-attestation is not independent evidence: the observer must exercise
the frozen application and preserve its scorecard and artifacts.

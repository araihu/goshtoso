# Alpine.js gotchas (Goshtoso: Alpine + templ + htmx)

Baseline target: Alpine 3.17.2. For htmx 4 swap lifecycle read [integration](htmx4.md); check `muamba.yaml` before assuming that integration is already installed.

## Generated expressions

Follow AGENTS.md's preference for simple data attributes and shared `Alpine.data()` providers for complex behavior. Inspect both the browser console and live `element.getAttribute('x-data')` value when debugging. HTML entities in serialized HTML are normal and decoded by the browser; double encoding, malformed JavaScript, or unavailable providers are distinct problems. Do not use `templ.Raw()` to bypass escaping on user data or interpolate untrusted strings into scripts.

Normalize nil Go slices to empty arrays before Alpine calls methods such as `.includes()`: JSON `null` is not `[]`.

## Registration and lifecycle

Prefer providers in the first-party bundle before Alpine starts. If registration must arrive later, do not rely exclusively on an `alpine:init` event that already fired:

```js
function register() {
  Alpine.data('logFeed', () => ({ entries: [] }))
}
if (window.Alpine) register()
else document.addEventListener('alpine:init', register, { once: true })
```

This registration still needs to run before the new component initializes. The compatibility extension cannot fix a provider that has not been registered. Alpine starts once; repeated navigation must not duplicate global listeners or timers. Release resources in `destroy()`.

Alpine observes DOM changes; htmx 4 compatibility coordinates those observations with swaps. When Alpine or application JS inserts new htmx markup itself, process the relevant subtree with `htmx.process()` if necessary. Avoid unconditional reinitialization of already live Alpine trees.

## UI and browser verification

- Initially hidden components need `x-cloak` and `[x-cloak]{display:none!important}`.
- `x-for` / `x-if` belong on a `<template>` with one root; use stable keys for reordered lists.
- Custom events bubble to ancestors, not siblings. Use `.window` where cross-component listening is intended.
- Keep update-only OOB attributes out of initial fragments; test main/OOB ordering after the htmx upgrade.
- For Goshtoso E2E assertions, use the existing helpers and live DOM evaluation for bound attributes. Await Alpine readiness and settled updates; verify actual `input` events when diagnosing `x-model` tests.
- Test both state-preserving morphs and intentional replacement/teardown. Compatibility does not make every swap preserve state.

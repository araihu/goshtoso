---
name: alpinejs
description: Use when writing, debugging, or reviewing Alpine.js markup — any x-* directive (x-data, x-init, x-show, x-bind/`:`, x-on/`@`, x-text, x-html, x-model, x-modelable, x-for, x-transition, x-effect, x-ignore, x-ref, x-cloak, x-teleport, x-if, x-id), magics ($el, $refs, $store, $watch, $dispatch, $nextTick, $root, $data, $id), globals (Alpine.data, Alpine.store, Alpine.bind), the init()/destroy() lifecycle, alpine:init, Alpine plugins (persist, intersect, collapse, focus, mask, morph), extending via Alpine.directive/Alpine.magic, the CSP build, or Alpine interacting with templ attribute escaping / htmx swaps.
---

# Alpine.js

## Overview

Alpine is a lightweight declarative framework: you sprinkle reactive behavior onto plain HTML via `x-*` attributes — no build step, no virtual DOM. Mental model: "Tailwind for JavaScript." Expressions in attributes are evaluated against the nearest `x-data` scope, which cascades to descendants.

Reference crawled from alpinejs.dev (Alpine v3.x).

## When to use

- Writing/editing any element with `x-data`, `x-show`, `x-on`/`@`, `x-bind`/`:`, `x-model`, `x-for`, `x-if`, `x-transition`, etc.
- Registering reusable behavior with `Alpine.data()`, global state with `Alpine.store()`, or attribute bundles with `Alpine.bind()`.
- Using magics (`$refs`, `$dispatch`, `$watch`, `$nextTick`, `$store`, `$id`) or the `init()` / `destroy()` lifecycle.
- Adding an Alpine plugin (persist, intersect, collapse, focus, mask, morph) or extending Alpine (`Alpine.directive`, `Alpine.magic`).
- Debugging why a component silently does nothing (Alpine swallows expression parse failures with no console error — usually attribute escaping; see gotchas).

**This repo (Goshtoso) uses Alpine v3 + templ + htmx.** Alpine is bundled locally at `assets/js/vendor/`. Read `reference/gotchas.md` whenever Alpine markup is generated through templ or combined with htmx — templ escaping and fragment-nav registration silently break Alpine.

## The core: data + behavior on one element

```html
<div x-data="{ open: false }">
    <button @click="open = ! open">Toggle</button>
    <div x-show="open" @click.outside="open = false" x-transition>Contents…</div>
</div>
```

`x-data` holds reactive state (and can hold methods, getters, and an `init()`). Everything inside reads that scope. Nested `x-data` inherits the parent scope and may shadow same-named keys.

## Directive quick reference

| Directive | Does | Note |
|-----------|------|------|
| `x-data` | Declare component + reactive scope | Supports methods, `get` getters, `init()` |
| `x-init` | Run expression on element init | Works without `x-data`; async OK |
| `x-show` | Toggle CSS `display` | `.important`; pair `x-cloak` if start hidden |
| `x-bind` / `:` | Set attribute from expression | `:class="{ 'hidden': !open }"`, `:style="{…}"`, bind whole object |
| `x-on` / `@` | Run code on a DOM event | `@click`, `@keyup.enter`; rich modifiers (below) |
| `x-text` | Set `textContent` | Plain text only |
| `x-html` | Set `innerHTML` | XSS risk — trusted content only |
| `x-model` | Two-way bind a form input | `.lazy .number .boolean .debounce .throttle .fill` |
| `x-modelable` | Expose an internal prop as an `x-model` target | For reusable components |
| `x-for` | Loop list/object/range | On `<template>`, single root, use `:key` |
| `x-if` | Add/remove DOM by condition | On `<template>`, single root, **no** `x-transition` |
| `x-transition` | Animate show/hide | `.duration.Nms .delay .opacity .scale .origin.*` |
| `x-effect` | Re-run expression when its deps change | Implicit watcher |
| `x-ignore` | Skip Alpine init for a subtree | — |
| `x-ref` | Tag element for `$refs.name` | Static only in v3 |
| `x-cloak` | Hide until Alpine inits | Needs `[x-cloak]{display:none!important}` CSS |
| `x-teleport` | Move `<template>` content elsewhere | `x-teleport="body"`, `.prepend/.append` |
| `x-id` | Scoped namespace for `$id()` | `x-id="['foo']"` → matching `$id('foo')` |

Most directives (except `x-data`, `x-init`, `x-ignore`) require a parent `x-data`.

## `x-on` / `x-model` modifiers (the ones you reach for)

- **Events:** `.prevent` `.stop` `.self` `.outside` `.window` `.document` `.once` `.passive` `.capture` `.camel` `.dot`
- **Debounce/throttle:** `.debounce` (250ms default, `.debounce.500ms`) · `.throttle`
- **Keys:** `.enter` `.escape` `.tab` `.space` `.up/.down/.left/.right` + chords `.shift` `.ctrl` `.cmd` `.alt` `.meta` (e.g. `@keyup.shift.enter`)
- **`x-model`:** `.lazy` `.number` `.boolean` `.debounce` `.throttle` `.fill` (seed from the input's `value`)

## Magics & globals

| Magic | Use |
|-------|-----|
| `$el` | Current DOM element |
| `$refs` | Elements tagged with `x-ref` |
| `$store` | Access an `Alpine.store()` |
| `$watch('prop', (v,old)=>…)` | React to a property change |
| `$dispatch('evt', detail)` | Fire a custom event (bubbles; use `.window` to catch across components) |
| `$nextTick(cb)` | Run after Alpine flushes DOM updates (awaitable) |
| `$root` | Nearest `x-data` ancestor element |
| `$data` | Current scope as an object (pass to external fns) |
| `$id('name')` | Collision-free scoped IDs |

```js
document.addEventListener('alpine:init', () => {
    Alpine.data('dropdown', () => ({
        open: false,
        init() { /* before first render */ },
        toggle() { this.open = ! this.open },
        destroy() { /* on teardown — clear timers/listeners */ },
    }))
    Alpine.store('darkMode', { on: false, toggle() { this.on = !this.on } })
})
```

`Alpine.bind('Name', () => ({...}))` defines a reusable attribute bundle applied via `x-bind="Name"`.

## Detailed references (load when needed)

- **`reference/directives.md`** — every directive with full syntax, all modifiers, and gotchas.
- **`reference/magics-globals.md`** — every magic, the three globals, and the full lifecycle (`alpine:init`, `alpine:initialized`, `init()`, `destroy()`, `x-init`, manual `Alpine.start()`).
- **`reference/plugins.md`** — persist, intersect, collapse, focus, mask, morph (install + usage), plus extending (`Alpine.directive`/`Alpine.magic`), reactivity primitives, async expressions, and the CSP build.
- **`reference/gotchas.md`** — templ attribute escaping, null-array crashes, fragment-nav `Alpine.data` registration, htmx-inserted nodes needing `htmx.process()`, `x-cloak` FOUC. **Read this before generating Alpine through templ or mixing with htmx.**

## Common mistakes

- **Silent dead component, no console error.** Alpine swallows expression parse failures. #1 cause in this repo: templ escapes `"`/`'`/`&` inside `x-data` (`&quot;`). Inspect rendered HTML in devtools. See `reference/gotchas.md`.
- **`json.Marshal` into an HTML attribute.** Produces double-quoted JSON; templ escapes the quotes; Alpine sees broken syntax. Build single-quoted JS, or register via `Alpine.data()` in a `<script>`.
- **`null` arrays.** `json.Marshal([]T(nil))` → `null`; Alpine `.includes()` throws. Guard to `[]`.
- **`x-for` / `x-if` not on `<template>`,** or with more than one root element — won't work. Add `:key` to `x-for`.
- **Forgetting `x-cloak`** on initially-hidden elements → flash of content before Alpine boots.
- **`Alpine.data()` registered only on `alpine:init`** — undefined when the node arrives via an htmx fragment swap after Alpine already started. Register immediately too, and call `htmx.process()` on Alpine-inserted nodes. See `reference/gotchas.md`.
- **`$dispatch` to a sibling** — events bubble to ancestors, not siblings; listen with `@evt.window`.

# Alpine.js Directives — full reference

Every directive with syntax, modifiers, and gotchas. Crawled from alpinejs.dev (v3.x).
Most directives require a parent `x-data`; exceptions noted.

## x-data
Defines an HTML chunk as an Alpine component and supplies its reactive data.
```html
<div x-data="{ open: false }">
    <button @click="open = ! open">Toggle</button>
    <div x-show="open">Content...</div>
</div>
```
- Supports methods (`toggle() { this.open = !this.open }`), getters (`get isOpen() {...}`), and an `init()` method (auto-runs on init).
- Scope cascades to children; reusable via `Alpine.data('name', () => ({...}))`.
- **Gotcha:** Properties in nested `x-data` shadow same-named parent properties (inner overrides outer).

## x-init
Hooks into an element's initialization phase. (No `x-data` required.)
```html
<div x-init="console.log('Initializing!')"></div>
```
- Supports async (`await`); pair with `$nextTick()` for post-render work.
- **Gotcha:** When both exist, the `x-data` object's `init()` runs first, then the `x-init` directive.

## x-show
Shows/hides an element via CSS `display`.
```html
<div x-show="open">Content</div>
```
- Modifier: `.important` — applies `display: none !important`.
- **Gotcha:** If initial state is `false`, add `x-cloak` to prevent flicker before Alpine loads.

## x-bind
Sets HTML attributes dynamically. **Shorthand `:`**.
```html
<input type="text" :placeholder="placeholderText">
<div :class="{ 'hidden': !show }"></div>      <!-- class object syntax, merges -->
<div :style="{ color: 'red', display: 'flex' }"></div>
<button x-bind="trigger"></button>            <!-- bind a whole object of attrs -->
```
- **Gotcha:** Requires a parent `x-data`.

## x-on
Runs code on a DOM event. **Shorthand `@`**.
```html
<button @click="alert('Hi')">Say Hi</button>
```
- Modifiers: `.prevent` `.stop` `.outside` `.window` `.document` `.once` `.self` `.camel` `.dot` `.passive` `.passive.false` `.capture`; `.debounce` (default 250ms, e.g. `.debounce.500ms`); `.throttle`.
- Keyboard: `.enter` `.space` `.escape` `.tab` `.up` `.down` `.left` `.right` `.page-down` …; chords `.shift` `.ctrl` `.cmd` `.alt` `.meta` (e.g. `@keyup.shift.enter`).
- **Gotcha:** Requires a parent `x-data`.

## x-text
Sets `textContent`.
```html
<strong x-text="username"></strong>
```
- **Gotcha:** Plain text only; HTML tags render as literal text (use `x-html` otherwise).

## x-html
Sets `innerHTML`.
```html
<span x-html="username"></span>
```
- **Gotcha:** XSS risk — trusted content only, never user input.

## x-model
Two-way binds a form input to a data property.
```html
<input type="text" x-model="message">
```
- Works with text/textarea/checkbox/radio/select/range.
- Modifiers: `.lazy` (sync on change), `.change`, `.blur`, `.enter`, `.number` (cast to number), `.boolean` (cast to boolean), `.debounce` (250ms default), `.throttle`, `.fill` (seed property from the input's `value` attribute).
- Programmatic: `el._x_model.get()` / `el._x_model.set(value)`.
- **Gotcha:** Requires a parent `x-data`. (In E2E: Playwright `Fill()` does not fire a native `input` event, so `x-model` won't update — dispatch `input` manually.)

## x-modelable
Exposes a component's internal property as an `x-model` target.
```html
<div x-data="{ count: 0 }" x-modelable="count" x-model="number">
    <button @click="count++">Increment</button>
</div>
```
- Use case: reusable components exposing internal state like a native input.

## x-for
Renders elements by iterating lists, objects, or ranges. **Must be on a `<template>` with one root element.**
```html
<template x-for="item in items" :key="item.id">
    <li x-text="item.name"></li>
</template>
```
- With index: `x-for="(item, index) in items"`. Object: `x-for="(value, key) in object"`. Range: `x-for="i in 10"`.
- **Gotcha:** Without a unique `:key`, reorder/remove causes odd side-effects. Requires parent `x-data`.

## x-transition
Animates show/hide.
```html
<div x-show="open" x-transition>Content</div>
```
- Modifiers: `.duration.{ms}` (default 150ms enter / 75ms leave), `.delay.{ms}`, `.opacity`, `.scale` / `.scale.{n}`, `.origin.{top|bottom|left|right|combos}`. Target enter/leave separately: `x-transition:enter.duration.500ms`.
- CSS class helpers: `:enter`, `:enter-start`, `:enter-end`, `:leave`, `:leave-start`, `:leave-end`.
- **Gotcha:** Does NOT work with `x-if` (only `x-show`). Requires parent `x-data`.

## x-effect
Re-runs an expression whenever its tracked dependencies change (implicit watcher).
```html
<div x-data="{ label: 'Hello' }" x-effect="console.log(label)"></div>
```
- Runs once on init, then on any referenced-property change; deps auto-detected.

## x-ignore
Stops Alpine from processing a DOM subtree.
```html
<div x-ignore><span x-text="label"></span></div>   <!-- x-text is skipped -->
```

## x-ref + $refs
Tags a DOM element for direct access.
```html
<button @click="$refs.text.remove()">Remove</button>
<span x-ref="text">Hello 👋</span>
```
- **Gotcha:** v3 supports static refs only (no dynamic `:x-ref`). Requires parent `x-data`.

## x-cloak
Hides elements until Alpine initializes (prevents FOUC). **Requires the CSS rule.**
```html
<style>[x-cloak] { display: none !important; }</style>
<span x-cloak x-show="false">Hidden until Alpine loads</span>
```
- Alpine removes the attribute once initialized. No-CSS alternative: wrap in `<template x-if="true">`.

## x-teleport
Moves `<template>` content to another DOM location (modals, z-index escapes).
```html
<template x-teleport="body">
  <div x-show="open">Modal contents...</div>
</template>
```
- Selector = any CSS query. Teleported content keeps parent scope, `$refs`, `$root`.
- Modifiers: `.prepend`, `.append`.
- **Gotcha:** Must be on a `<template>`. Alpine re-dispatches forwarded events to avoid double-bubbling.

## x-if
Adds/removes DOM by condition (vs `x-show` which only CSS-hides). **Must be on a `<template>` with one root element.**
```html
<template x-if="open">
    <div>Contents...</div>
</template>
```
- **Gotcha:** Does NOT support `x-transition`. Requires parent `x-data`.

## x-id + $id
Scoped namespace for generating unique IDs across repeated instances.
```html
<div x-id="['text-input']">
    <label :for="$id('text-input')">Username</label>
    <input type="text" :id="$id('text-input')">
</div>
```
- Same name within a scope yields the same ID (for label/input pairing); different instances get unique suffixes.

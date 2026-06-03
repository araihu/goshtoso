# Alpine.js Magics, Globals & Lifecycle

Crawled from alpinejs.dev (v3.x).

## Magics

### $el
Current DOM node.
```html
<button @click="$el.innerHTML = 'Hello'">Replace me</button>
```

### $refs
Elements tagged with `x-ref`.
```html
<button @click="$refs.text.remove()">Remove</button>
<span x-ref="text">Hello 👋</span>
```
- **Gotcha:** v3 supports static refs only; dynamic `:x-ref="item.name"` yields the literal string.

### $store
Access a global store registered with `Alpine.store()`.
```html
<button x-data @click="$store.darkMode.toggle()">Toggle</button>
<div x-data :class="$store.darkMode.on && 'bg-black'">Content</div>
```
- Stores can be primitives, not just objects.

### $watch
Run a callback when a watched property changes.
```html
<div x-data="{ open: false }" x-init="$watch('open', (value, old) => console.log(value))">
    <button @click="open = ! open">Toggle</button>
</div>
```
- Can watch nested props: `$watch('foo.bar', …)`.
- **Gotcha:** Mutating a property of the watched object inside the callback can cause an infinite loop.

### $dispatch
Shortcut for `el.dispatchEvent(new CustomEvent(...))`.
```html
<button @click="$dispatch('notify', { id: 1 })">Notify</button>
```
- **Gotcha:** Events bubble to ancestors, NOT siblings. To catch a sibling's event, listen at window: `@notify.window="..."`. Read detail via `$event.detail`.

### $nextTick
Run AFTER Alpine has applied its reactive DOM updates. Returns a Promise (awaitable).
```js
title = 'Hello World!'
$nextTick(() => { console.log($el.innerText) })   // reads updated DOM
```

### $root
Root element of the component (closest `x-data` ancestor).
```html
<div x-data data-message="Hi">
    <button @click="alert($root.dataset.message)">Say Hi</button>
</div>
```

### $data
Current scope as a plain object — pass an entire scope to an external function.
```html
<div x-data="{ greeting: 'Hello' }">
    <button @click="sayHello($data)">Say Hello</button>
</div>
```
- Usually unnecessary — access data directly in expressions.

### $id
Collision-free scoped IDs (pair with the `x-id` directive — see directives.md).
```html
<input :id="$id('text-input')">   <!-- id="text-input-1", "-2", … per instance -->
```

## Globals

### Alpine.data — reusable `x-data` contexts
```js
document.addEventListener('alpine:init', () => {
    Alpine.data('dropdown', () => ({
        open: false,
        toggle() { this.open = ! this.open },
    }))
})
```
Bundled/module form:
```js
import Alpine from 'alpinejs'
import dropdown from './dropdown.js'
Alpine.data('dropdown', dropdown)
Alpine.start()
```
Use in markup: `<div x-data="dropdown">`.

### Alpine.store — global reactive state
```js
document.addEventListener('alpine:init', () => {
    Alpine.store('darkMode', {
        on: false,
        toggle() { this.on = ! this.on },
    })
})
```
- `init()` inside a store runs immediately after registration, before any render:
```js
Alpine.store('darkMode', {
    init() { this.on = window.matchMedia('(prefers-color-scheme: dark)').matches },
    on: false,
})
```
- Read/write from JS: `Alpine.store('darkMode').on = true`.

### Alpine.bind — reusable attribute bundles
```js
Alpine.bind('SomeButton', () => ({
    type: 'button',
    '@click'() { this.doSomething() },
    ':disabled'() { return this.shouldDisable },
}))
```
```html
<button x-bind="SomeButton"></button>
```

## Lifecycle

Order: `alpine:init` → each component's `init()` / `x-init` → `alpine:initialized`.

- **`alpine:init`** (on `document`) — fires before Alpine boots; register `Alpine.data`/`Alpine.store`/`Alpine.bind` here.
- **`alpine:initialized`** — fires after Alpine finishes initializing the page.
- **`init()`** — method on an `x-data` / `Alpine.data` object; runs before that element initializes.
- **`destroy()`** — method on the data object; runs on teardown (element removed via `x-if`/`x-for`, etc.). Clear timers/listeners here to avoid leaks.
  ```js
  Alpine.data('timer', () => ({
      timer: null,
      init() { this.timer = setInterval(() => {/*…*/}, 1000) },
      destroy() { clearInterval(this.timer) },
  }))
  ```
- **`x-init`** — attribute form to run an expression on element init.
- **`Alpine.start()`** — call once after import when NOT using the auto-starting CDN build.

> **In this repo:** Alpine is bundled and already running by the time fragments arrive via htmx/fragment-nav. Registering `Alpine.data()` ONLY inside an `alpine:init` listener leaves it undefined for later-loaded fragments — register immediately as well. See `gotchas.md`.

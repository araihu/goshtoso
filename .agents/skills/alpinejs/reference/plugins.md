# Alpine.js Plugins & Advanced

Crawled from alpinejs.dev (v3.x). For CDN use, **plugin script tags must load BEFORE the core Alpine script.** Versions below are illustrative — treat as latest 3.x.

## Plugins

### persist — state survives page loads (localStorage)
```html
<script defer src="https://cdn.jsdelivr.net/npm/@alpinejs/persist@3.x.x/dist/cdn.min.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
```
```bash
npm install @alpinejs/persist
```
```html
<div x-data="{ count: $persist(0) }">
    <button @click="count++">Increment</button>
    <span x-text="count"></span>
</div>
```
- Custom key: `$persist(0).as('my-key')`. Other storage: `.using(sessionStorage)`.

### intersect — run on viewport enter/leave (IntersectionObserver)
```html
<div x-intersect="shown = true">Triggers when visible</div>
<div x-intersect.once="shown = true">Fires only once</div>
<div x-intersect:enter="shown = true">On entering</div>
<div x-intersect:leave="shown = true">On leaving</div>
<div x-intersect.half="shown = true">50% visible</div>
<div x-intersect.full="shown = true">99% visible</div>
<div x-intersect.threshold.50="shown = true">custom threshold %</div>
<div x-intersect.margin.200px="loaded = true">200px buffer</div>
```
- Common use: lazy-load / infinite scroll sentinel.

### collapse — animate height on show/hide
```html
<div x-data="{ expanded: false }">
    <button @click="expanded = ! expanded">Toggle</button>
    <p x-show="expanded" x-collapse>Content</p>
</div>
```
- `x-collapse.min.50px` keeps a minimum visible height when collapsed.

### focus — focus trapping (`x-trap`) + `$focus` magic
```html
<div x-data="{ open: false }">
    <button @click="open = true">Open</button>
    <span x-show="open" x-trap="open">
        <input type="text">
        <button @click="open = false">Close</button>
    </span>
</div>

<div @keydown.right="$focus.next()" @keydown.left="$focus.previous()">
    <button>First</button><button>Second</button>
</div>
```

### mask — auto-format input as the user types
```html
<input x-mask="99/99/9999" placeholder="MM/DD/YYYY">
<input x-mask:dynamic="$input.startsWith('34') ? '9999 999999 99999' : '9999 9999 9999 9999'">
<input x-mask:dynamic="$money($input)">
```
- Wildcards: `*` any char, `a` letters, `9` numbers.

### morph — morph an element into new HTML, preserving DOM + Alpine state
```js
Alpine.morph(el, `
    <div x-data="{ message: 'hi' }">
        <input type="text" x-model="message">
        <span x-text="message"></span>
    </div>
`)
```
- Keeps focus, scroll, and component state — useful for server-rendered HTML diffing.

## Reactivity primitives
```js
let data = Alpine.reactive({ count: 1 })   // tracking proxy
Alpine.effect(() => console.log(data.count)) // runs now + on each dep change
data.count = 2 // logs "2"
```
These are what Alpine is built on; `effect` auto-tracks reactive props read inside it.

## Extending Alpine

Directive — `name` → `x-name`:
```js
Alpine.directive('uppercase', (el, { value, modifiers, expression }, { Alpine, effect, cleanup }) => {
    el.textContent = el.textContent.toUpperCase()
})
```
- `value` = colon part (`x-foo:bar` → `'bar'`); `modifiers` = dot array (`x-foo.baz` → `['baz']`); `expression` = the attribute value. `effect` = auto-cleaning reactive runner; `cleanup` = register teardown.

Magic — `name` → `$name`:
```js
Alpine.magic('now', () => (new Date).toLocaleTimeString())          // $now
Alpine.magic('clipboard', () => subject => navigator.clipboard.writeText(subject)) // $clipboard(x)
```

## Async expressions
Alpine awaits async functions anywhere it accepts sync ones:
```html
<span x-text="await getLabel()"></span>
<span x-text="getLabel"></span>   <!-- reference w/o parens: auto-detected & awaited -->
```

## CSP build
The CSP-friendly build forbids `eval`-style inline evaluation: **no arrow functions, destructuring, template literals, or global access in attributes.** Move all logic into `Alpine.data()` components (getters/methods); keep attributes to plain property references.
```html
<div x-data="userManager" x-show="hasActiveAdmins"></div>
<script nonce="...">
  Alpine.data('userManager', () => ({
    users: [],
    get hasActiveAdmins() { return this.users.some(u => u.active && u.role === 'admin') },
  }))
</script>
```

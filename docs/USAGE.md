# Goshtoso Consumer Integration Guide

This guide explains how to use Goshtoso components in Go web applications built
with templ, Tailwind CSS, HTMX, and Alpine.js.

Before choosing component constructors and options, read the
[Goshtoso Component Model](COMPONENT_MODEL.md). It defines themes, primitives,
stable `Kind` identity, configuration dimensions, and the rule for deciding
whether a difference is a dimension of one primitive or a separate primitive.

## Installation

### 1. Add the dependency

```bash
go get github.com/araihu/goshtoso@latest
```

Goshtoso's own components ship **pre-generated** — you never run `templ generate`
against the library. But your *own* `.templ` pages do need the templ toolchain:

```bash
go get github.com/a-h/templ                       # runtime (your generated code imports it)
go install github.com/a-h/templ/cmd/templ@latest  # the CLI, if not already installed
templ generate                                    # YOUR .templ → _templ.go
```

### 2. Serve the bundled assets (recommended)

The fastest, deterministic path: serve Goshtoso's embedded assets and let
`head.Dependencies()` link them. No CDN, no Tailwind build, no extraction.

```go
// main.go
import "github.com/araihu/goshtoso/assets"

http.Handle("/assets/", assets.Handler()) // serves styles.css + js/ + fonts/ + images/
                                          // NOTE: self-strips /assets/ — don't wrap in StripPrefix
```

```go
// page.templ — emits the matching /assets/* <link>/<script> tags
import "github.com/araihu/goshtoso/components/head"

templ Layout() {
    <html>
        <head>
            @head.Dependencies()        // CSS + Alpine + collapse/focus + HTMX + combobox nav
            // or @head.DependenciesMinimal() — CSS + Alpine core + HTMX + combobox nav (no collapse/focus plugins)
        </head>
        ...
    </html>
}
```

The served `styles.css` already carries every component style + the theme system
(13 themes). **Stock CDN Tailwind will not work** — the theme tokens
(`bg-primary`, `text-on-surface`, …) live only in this compiled CSS.

Skip the rest of this section unless you maintain your own Tailwind build.

### 2b. Extract Goshtoso CSS (only for a custom Tailwind build)

Goshtoso ships a CLI that extracts the pre-built CSS from embedded assets. Register it as a Go tool for version-pinned reproducibility:

```bash
# Add to go.mod (alongside your other tools)
# tool github.com/araihu/goshtoso/cmd/goshtoso
go mod tidy

# Extract CSS
go tool goshtoso -out=css/goshtoso-base.css
```

Or use `go run` for one-off extraction:

```bash
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -out=css/goshtoso-base.css
```

Then import it in your Tailwind entry point:

```css
/* your-project/css/main.css */
@import "tailwindcss";
@import "./goshtoso-base.css";
```

The extracted CSS includes all Goshtoso component styles, the theme system (13 themes), and base utilities. Add it to `.gitignore` since it's a build artifact.

### 3. Required JavaScript

If you use `@head.Dependencies()` (section 2), the JS is already wired — skip
this. Only hand-roll the tags if you are not using the `head` package. In that
case mirror what `head.Dependencies()` emits, in this order (plugins **before**
Alpine core):

```html
<!-- Versions are pinned in assets/js/runtime/versions.json
     (see assets.AlpineVersion(), assets.HTMXVersion()). -->
<link rel="stylesheet" href="/assets/styles.css"/>
<script defer src="/assets/js/runtime/alpinejs-collapse/3.14.9/alpine-collapse.min.js"></script>
<script defer src="/assets/js/runtime/alpinejs-focus/3.14.9/alpine-focus.min.js"></script>
<script defer src="/assets/js/runtime/alpinejs/3.14.9/alpine.min.js"></script>
<script src="/assets/js/runtime/htmx.org/2.0.8/htmx.min.js"></script>
<script defer src="/assets/js/combobox.js"></script>
```

These are the vendored files `assets.Handler()` serves — the version is in the
path, so there is no floating CDN tag to drift. **These versioned paths change
when you upgrade a dep; prefer `@head.Dependencies()` so you never hardcode
them.** Don't forget `combobox.js` (the combobox component's keyboard nav is dead
without it) — it is first-party, so it stays unversioned at `/assets/js/`.

## Using your own Tailwind build

`goshtoso -version` prints the Tailwind version Goshtoso's CSS was built with
(also in [`VERSIONS.md`](../VERSIONS.md)). Match your own Tailwind to it.

### Path A — two stylesheets (recommended, low coupling)

Serve Goshtoso's prebuilt CSS and run your own Tailwind into a *separate* file.
No recompiling Goshtoso.

```html
<link rel="stylesheet" href="/assets/styles.css"/>  <!-- Goshtoso, via assets.Handler() -->
<link rel="stylesheet" href="/css/app.css"/>          <!-- your own Tailwind output -->
```

```css
/* your app.css — your own tokens/classes only */
@import "tailwindcss";
@theme { --color-brand: oklch(0.7 0.15 250); }
```

### Path B — unified build (one tree-shaken stylesheet)

Compile Goshtoso's theme source together with your own. Requires your Tailwind
to match `goshtoso -version`.

```bash
# 1. extract the theme SOURCE next to your CSS
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -theme -out=css/goshtoso-theme.css

# 2. discover the components dir Tailwind must scan
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -source-path
# -> /…/go/pkg/mod/github.com/araihu/goshtoso@vX.Y.Z/components
```

```css
/* your main.css */
@import "tailwindcss";
@import "./goshtoso-theme.css";                 /* tokens + selectors + themes */
@source "/…/goshtoso@vX.Y.Z/components";        /* emit Goshtoso's classes (path from -source-path) */
@theme { --color-brand: oklch(0.7 0.15 250); }  /* your own tokens too */
```

Goshtoso's fonts/images are still served by `assets.Handler()` at `/assets/`,
so mount it regardless of which path you choose.

## Component Catalog

All components are imported from `github.com/araihu/goshtoso/components/<name>`.
Run the demo server (`go run ./site/cmd/server`) or visit
[goshtoso.araihu.com](https://goshtoso.araihu.com/) for interactive examples,
configuration previews, and API tables.

| Component | Import | Description |
|-----------|--------|-------------|
| `accordion` | `components/accordion` | Collapsible sections with default, plain, and split appearances |
| `alert` | `components/alert` | Dismissable alert banners with info, success, warning, and danger tones |
| `avatar` | `components/avatar` | User avatar with image, initials fallback, status indicator |
| `badge` | `components/badge` | Inline status badges with independent tone, appearance, and size dimensions |
| `banner` | `components/banner` | Full-width notifications and consent dialogs as separate `Banner` and `CookieBanner` primitives |
| `breadcrumbs` | `components/breadcrumbs` | Navigation breadcrumb trail with custom separators |
| `button` | `components/button` | Buttons with tone and size options plus HTMX and Alpine.js integration |
| `card` | `components/card` | Content cards with image, rating, price, and multiple layouts |
| `carousel` | `components/carousel` | Image carousel with autoplay, navigation, and HTMX lazy loading |
| `chatbubble` | `components/chatbubble` | Chat/message bubbles with sender alignment and avatar support |
| `checkbox` | `components/checkbox` | Checkboxes with semantic tones, group layout, and indeterminate state |
| `codeblock` | `components/codeblock` | Code display block with copy button and max-height scrolling |
| `combobox` | `components/combobox` | Searchable dropdown with single/multi-select, HTMX server search |
| `drawer` | `components/drawer` | Slide-over drawers for navigation and contextual panels |
| `dropdown` | `components/dropdown` | Context menus, action menus with icons, shortcuts, sections |
| `fileinput` | `components/fileinput` | File input controls with labels, helper text, and validation states |
| `form` | `components/form` | Form orchestrator: Section, FlipSection, CollapsibleSection, FieldGroup |
| `kbd` | `components/kbd` | Semantic keyboard shortcut and user input hints |
| `link` | `components/link` | Styled link primitives with external-link and navigation affordances |
| `modal` | `components/modal` | General and confirmation dialogs as separate `Modal` and `AlertDialog` primitives; `Tone` belongs to `AlertDialog` |
| `navbar` | `components/navbar` | Top navigation bar with links, user profile dropdown, action items |
| `pagination` | `components/pagination` | Page navigation with HTMX, ellipsis, prev/next buttons |
| `palette` | `components/palette` | Color palette and swatch utilities for theme demos and pickers |
| `radio` | `components/radio` | Radio inputs and groups with validation and semantic tones |
| `range` | `components/range` | Range sliders with labels, helper text, and icon slots |
| `rating` | `components/rating` | Rating controls and display states |
| `schemaform` | `components/schemaform` | Schema Form: generate form controls from JSON Schema, defaults, current values, and allow-list rules |
| `search` | `components/search` | Search input and command-palette style result lists |
| `select` | `components/select` | HTML select dropdown with validation states, readonly mode |
| `sidebar` | `components/sidebar` | Collapsible sidebar with sections, nested items, badges |
| `spinner` | `components/spinner` | Loading spinner with independent size and tone dimensions |
| `steps` | `components/steps` | Stepper/progress navigation for multi-step flows |
| `structuredinput` | `components/structuredinput` | Repeatable structured row editor (for labels, taints, rules) |
| `table` | `components/table` | Data table with sorting, pagination, infinite scroll, filters, row links |
| `tabs` | `components/tabs` | Tab navigation with badges, HTMX lazy content loading |
| `tagslist` | `components/tagslist` | Dynamic tag list editor (add/remove string tags) |
| `textarea` | `components/textarea` | Multi-line text input with validation states |
| `textinput` | `components/textinput` | Text input with types (text, email, password, number), validation |
| `toast` | `components/toast` | Notifications as separate `Toast` and `MessageToast` primitives; sender and avatar content belongs to `MessageToast` |
| `toggle` | `components/toggle` | Toggle switch with semantic tones |
| `tooltip` | `components/tooltip` | Hover tooltips with position options, rich content support |

## Basic Component Pattern

Import the component package you need and follow its constructor contract.
Atomic primitives such as Button use functional options; provide child content
when the component accepts children:

```go
import "github.com/araihu/goshtoso/components/button"

templ Actions() {
    @button.Button(
        button.WithTone(button.TonePrimary),
        button.WithType("submit"),
    ) {
        Save changes
    }
}
```

For component-specific fields, prefer the generated Go documentation and the
demo site's API tables:

- [Go package reference](https://pkg.go.dev/github.com/araihu/goshtoso)
- [Live component docs](https://goshtoso.araihu.com/components/button)

Public config fields follow
[`docs/COMPONENT_API_NAMING.md`](COMPONENT_API_NAMING.md). Shared extension
points generally use target-specific names such as `RootClass`, `InputAttrs`,
`HTMX`, and `Alpine`.

## Theming

### Available Themes

Goshtoso ships 13 built-in themes. The default theme is `goshtoso`; the Minimal
theme is useful for checking no-radius edge cases.

### Switching Themes

Set the theme via data attribute on `<html>`:

```html
<html data-theme="modern">
```

Or with JavaScript:

```javascript
document.documentElement.setAttribute('data-theme', 'modern');
```

### Dark Mode

Add/remove the `dark` class on `<html>`:

```javascript
document.documentElement.classList.toggle('dark');
```

## Best Practices

### 1. Prefer templ components for rich content

Create separate templ components for accordion content to keep code clean:

```go
templ SettingsContent() {
    <div class="space-y-4">
        <div>
            <label class="block text-sm font-medium">Name</label>
            <input type="text" class="mt-1 block w-full" />
        </div>
        <div>
            <label class="block text-sm font-medium">Email</label>
            <input type="email" class="mt-1 block w-full" />
        </div>
    </div>
}

// Use it
@accordion.Accordion(accordion.AccordionConfig{
    Items: []accordion.AccordionItem{
        {Title: "Settings", Content: SettingsContent()},
    },
})
```

### 2. Pass icons and custom slots as templ components

Many components accept icons, actions, details, or custom bodies as
`templ.Component`. Keep those as normal templ functions when possible. Use
`templ.Raw` only for trusted, static HTML or scripts you fully control.

### 3. HTMX Integration

Components work seamlessly with HTMX for dynamic updates:

```go
// Initial render
@accordion.Accordion(accordion.AccordionConfig{
    ID: "cluster-accordion",
    Items: []accordion.AccordionItem{
        {
            ID:      "node-pools",
            Title:   "Node Pools",
            Content: NodePoolsTable(cluster.NodePools),
        },
    },
})

// Update fragment via HTMX
func HandleNodePoolsUpdate(w http.ResponseWriter, r *http.Request) {
    clusterID := r.URL.Query().Get("cluster_id")
    nodePools := fetchNodePools(clusterID)
    
    // Render just the content that changed
    accordion.AccordionItem{
        ID:      "node-pools",
        Title:   "Node Pools",
        Content: NodePoolsTable(nodePools),
    }.Render(r.Context(), w)
}
```

### 4. Testing

For application tests, render your own templ pages and assert on the generated
HTML, then cover important browser behavior with Playwright or your preferred
E2E tool. The Goshtoso repository's `components/*/*_test.go` and
`site/tests/e2e/*_test.go` files are useful examples.

## Known Pitfalls

### HTMX History Cache vs Alpine.js State

When using HTMX SPA navigation (`hx-get` + `hx-target="#main-content-area"` + `hx-push-url`), HTMX caches the raw `document.body.innerHTML` for back-button history restore. The problem: Alpine-generated DOM nodes (from `x-for`, `x-text`, etc.) are saved in the cache, but Alpine scope objects are lost. On back-button restore, the page shows stale Alpine-generated elements with no reactivity — combobox dropdowns with blank items, broken toggles, etc.

**Recommended approaches (pick one per use case):**

1. **`LinkMode: LinkBoost`** on table rows - swaps the full `<body>` via `hx-select="body"` + `hx-target="body"`. Back-button re-fetches from server, so Alpine re-initializes cleanly. No stale cache.

2. **`LinkMode: LinkFull`** on table rows - plain `window.location.href` navigation. Simplest, safest. Use when the target page has complex Alpine state.

3. **`hx-history="false"`** on a container - tells HTMX not to cache this page. Back-button will fetch from server. Useful when you can't control the navigation source.

4. **Alpine re-init on history restore** - listen for `htmx:historyRestore` and call `Alpine.initTree(document.body)`. Works in theory but is fragile: HTMX strips `<script>` tags from cached HTML, so Alpine data registrations may be missing.

```go
// Example: table rows with boost mode (recommended for lists → detail navigation)
row := table.Row{
    ID:       "cluster-1",
    Link:     "/clusters/abc-123",
    LinkMode: table.LinkBoost,
    Cells:    cells,
}
```

### IntersectionObserver in Nested Scroll Containers

HTMX's `intersect` and `revealed` triggers use `IntersectionObserver` with the **viewport** as root. If the table is inside a container with `overflow-y-auto` (e.g., a scrollable main content area), the sentinel element may already be in the viewport even though it's scrolled out of view within its parent. The observer fires immediately or never fires on scroll.

Goshtoso's table infinite scroll sentinel includes a built-in scroll-listener fallback that attaches to the nearest `.overflow-y-auto` ancestor. This handles the nested-scroll case automatically.

If you're building custom infinite scroll outside the table component, use this pattern:

```html
<tr id="sentinel"
    hx-get="/next-page"
    hx-trigger="intersect once"
    hx-swap="outerHTML">
</tr>
<script>
// Fallback for nested scroll containers
(function() {
    var sentinel = document.getElementById('sentinel');
    if (!sentinel) return;
    var container = sentinel.closest('.overflow-y-auto');
    if (!container) return;
    function check() {
        var rect = sentinel.getBoundingClientRect();
        var cRect = container.getBoundingClientRect();
        if (rect.top < cRect.bottom + 200) {
            container.removeEventListener('scroll', check);
            htmx.trigger(sentinel, 'intersect');
        }
    }
    container.addEventListener('scroll', check);
    check(); // check immediately in case already visible
})();
</script>
```

## Troubleshooting

### Component not styled correctly?

1. Confirm `/assets/styles.css` is being served by `assets.Handler()`.
2. Confirm `@head.Dependencies()` or an equivalent `<link>` tag is present.
3. If you run a custom Tailwind build, match the version in `VERSIONS.md` and
   include Goshtoso's theme source and component `@source` path.
4. Ensure the `data-theme` attribute is set on `<html>`.

### Alpine.js not working?

1. Prefer `@head.Dependencies()` so Alpine core and plugins are loaded in the
   supported order.
2. Check browser console for Alpine errors.
3. Avoid embedding marshaled JSON directly in Alpine attributes; templ escapes
   quotes and Alpine can fail silently. Register complex behavior with
   `Alpine.data()` instead.
4. For collapse animations, ensure the collapse plugin is loaded before Alpine
   core.

### Dark mode not working?

1. Add the `dark` class to `<html>`.
2. Verify Goshtoso's CSS is loaded before app-specific overrides.
3. Check that theme state is applied before first paint if your app persists
   theme preferences.

## Examples and References

See the `/components` directory for component implementations and the demo site
for complete examples of each component.

Run the demo server:

```bash
cd /path/to/goshtoso
go run ./site/cmd/server -port 8090
```

Then visit:
- http://localhost:8090/components/accordion
- http://localhost:8090/components/table
- http://localhost:8090/examples/todo

The public documentation site is available at
[https://goshtoso.araihu.com/](https://goshtoso.araihu.com/).

## Contributing

For contribution workflow, generated-file rules, and local quality gates, see
[`CONTRIBUTING.md`](../CONTRIBUTING.md). For release expectations, see
[`docs/RELEASE_CHECKLIST.md`](RELEASE_CHECKLIST.md).

## License

MIT. See [`LICENSE`](../LICENSE).

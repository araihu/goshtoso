# Goshtoso Components - Usage Guide

This guide explains how to use Goshtoso (Go + Alpine.js + Tailwind CSS + HTMX + Templ) components from the tks-console and other projects.

## Installation

### 1. Add Dependency

```bash
go get github.com/araihu/goshtoso@latest
```

### 2. Extract Goshtoso CSS

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

Include these CDN links in your HTML head:

```html
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
<script src="https://unpkg.com/htmx.org@2.0.8/dist/htmx.min.js"></script>
```

For components using Alpine.js collapse plugin (like Accordion):
```html
<script defer src="https://cdn.jsdelivr.net/npm/@alpinejs/collapse@3.x.x/dist/cdn.min.js"></script>
```

## Component Catalog

All components are imported from `github.com/araihu/goshtoso/components/<name>`. Run the demo server (`go run cmd/server/main.go`) to see interactive examples.

| Component | Import | Description |
|-----------|--------|-------------|
| `accordion` | `components/accordion` | Collapsible sections with multiple variants (default, no-background, bordered) |
| `alert` | `components/alert` | Dismissable alert banners with info/success/warning/danger variants |
| `avatar` | `components/avatar` | User avatar with image, initials fallback, status indicator |
| `badge` | `components/badge` | Inline status badges with solid/soft variants and sizes |
| `banner` | `components/banner` | Full-width notification banners with CTAs, cookie consent variant |
| `breadcrumbs` | `components/breadcrumbs` | Navigation breadcrumb trail with custom separators |
| `button` | `components/button` | Buttons with 8 variants, 4 sizes, HTMX and Alpine.js integration |
| `card` | `components/card` | Content cards with image, rating, price, and multiple layouts |
| `carousel` | `components/carousel` | Image carousel with autoplay, navigation, and HTMX lazy loading |
| `checkbox` | `components/checkbox` | Checkboxes with 6 color variants, group layout, indeterminate state |
| `codeblock` | `components/codeblock` | Code display block with copy button and max-height scrolling |
| `combobox` | `components/combobox` | Searchable dropdown with single/multi-select, HTMX server search |
| `dropdown` | `components/dropdown` | Context menus, action menus with icons, shortcuts, sections |
| `form` | `components/form` | Form orchestrator: Section, FlipSection, CollapsibleSection, FieldGroup |
| `keyvalue` | `components/keyvalue` | Dynamic key-value pair editor (for labels, env vars) |
| `modal` | `components/modal` | Dialogs with info/danger/warning variants, custom actions |
| `navbar` | `components/navbar` | Top navigation bar with links, user profile dropdown, action items |
| `pagination` | `components/pagination` | Page navigation with HTMX, ellipsis, prev/next buttons |
| `select` | `components/select` | HTML select dropdown with validation states, readonly mode |
| `sidebar` | `components/sidebar` | Collapsible sidebar with sections, nested items, badges |
| `spinner` | `components/spinner` | Loading spinner with size and color variants |
| `table` | `components/table` | Data table with sorting, pagination, infinite scroll, filters, row links |
| `tabs` | `components/tabs` | Tab navigation with badges, HTMX lazy content loading |
| `tagslist` | `components/tagslist` | Dynamic tag list editor (add/remove string tags) |
| `textarea` | `components/textarea` | Multi-line text input with validation states |
| `textinput` | `components/textinput` | Text input with types (text, email, password, number), validation |
| `toast` | `components/toast` | Toast notifications with auto-dismiss, position, sender avatar |
| `toggle` | `components/toggle` | Toggle switch with 6 color variants |
| `tooltip` | `components/tooltip` | Hover tooltips with position options, rich content support |
| `triplet` | `components/triplet` | Key-value-effect editor (for Kubernetes taints) |

## Detailed Examples

### Accordion

A collapsible accordion component with multiple variants.

**Import:**
```go
import "github.com/araihu/goshtoso/components/accordion"
```

**Basic Usage:**
```go
@accordion.Accordion(accordion.AccordionConfig{
    Items: []accordion.AccordionItem{
        {
            ID:      "section1",
            Title:   "Section 1",
            Content: templ.Raw("<p>This is the content for section 1.</p>"),
        },
        {
            ID:      "section2", 
            Title:   "Section 2",
            Content: templ.Raw("<p>This is the content for section 2.</p>"),
        },
    },
})
```

**Allow Multiple Open:**
```go
@accordion.Accordion(accordion.AccordionConfig{
    AllowMultiple: true,
    Items: []accordion.AccordionItem{
        {ID: "item1", Title: "Item 1", Content: content1},
        {ID: "item2", Title: "Item 2", Content: content2},
    },
})
```

**No Background Variant:**
```go
@accordion.Accordion(accordion.AccordionConfig{
    Variant: accordion.NoBackground,
    Items:   items,
})
```

**With Icons:**
```go
@accordion.Accordion(accordion.AccordionConfig{
    Items: []accordion.AccordionItem{
        {
            ID:      "settings",
            Title:   "Settings",
            Icon:    icons.SettingsIcon(), // Your icon component
            Content: SettingsContent(),
        },
    },
})
```

**Initially Expanded:**
```go
@accordion.Accordion(accordion.AccordionConfig{
    Items: []accordion.AccordionItem{
        {
            ID:                "details",
            Title:             "Details",
            Content:           DetailsContent(),
            InitiallyExpanded: true,
        },
    },
})
```

**Disabled Item:**
```go
@accordion.Accordion(accordion.AccordionConfig{
    Items: []accordion.AccordionItem{
        {
            ID:       "restricted",
            Title:    "Restricted Section",
            Content:  Content(),
            Disabled: true,
        },
    },
})
```

### Button

A versatile button component with multiple variants.

**Import:**
```go
import "github.com/araihu/goshtoso/components/button"
```

**Basic Usage:**
```go
@button.Button(button.Config{
    Variant: button.Primary,
    Type:    "button",
}) {
    Click Me
}
```

**With HTMX:**
```go
@button.Button(button.Config{
    Variant: button.Primary,
    HTMX: &button.HTMXConfig{
        Post:   "/api/action",
        Target: "#result",
        Swap:   "innerHTML",
    },
}) {
    Submit
}
```

**With Alpine.js:**
```go
@button.Button(button.Config{
    Variant: button.Primary,
    Alpine: &button.AlpineConfig{
        OnClick: "modalIsOpen = true",
    },
}) {
    Open Modal
}
```

**All Variants:**
- `button.Primary` - Black background, white text
- `button.Secondary` - Dark gray, white text
- `button.Alternate` - Light gray, dark text
- `button.Inverse` - Inverted colors
- `button.Info` - Sky blue, white text
- `button.Danger` - Red, white text
- `button.Warning` - Yellow/amber, dark text
- `button.Success` - Green, dark text

**Sizes:**
```go
@button.Button(button.Config{
    Variant: button.Primary,
    Size:    button.SizeSmall,  // xs text
})

@button.Button(button.Config{
    Variant: button.Primary,
    Size:    button.SizeMedium, // sm text (default)
})

@button.Button(button.Config{
    Variant: button.Primary,
    Size:    button.SizeLarge,  // base text
})

@button.Button(button.Config{
    Variant: button.Primary,
    Size:    button.SizeXLarge, // lg text
})
```

## Configuration Types

### AccordionItem

```go
type AccordionItem struct {
    ID                string          // Unique identifier
    Title             string          // Header text
    Content           templ.Component // Body content
    Icon              templ.Component // Optional icon
    Disabled          bool            // Disable interaction
    InitiallyExpanded bool            // Start expanded
}
```

### AccordionConfig

```go
type AccordionConfig struct {
    Items         []AccordionItem  // Accordion sections
    AllowMultiple bool             // Multiple open at once
    Variant       Variant          // Visual style
    ID            string           // Container ID
    Class         string           // Additional CSS classes
}
```

### Button Config

```go
type Config struct {
    Variant     Variant       // Button style
    Size        Size          // Button size
    Type        string        // HTML type attribute
    Disabled    bool          // Disabled state
    ID          string        // Element ID
    Class       string        // Additional classes
    HTMX        *HTMXConfig   // HTMX attributes
    Alpine      *AlpineConfig // Alpine.js directives
    LoadingText string        // Loading state text
}
```

## Theming

### Available Themes

Goshtoso supports all 13 PenguinUI themes:

1. **Minimal** (default) - Black/white, no border radius
2. **Arctic** - Cool blue tones
3. **Modern** - Clean professional look
4. **High-Contrast** - Maximum accessibility
5. **And more...**

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

### 1. Content Components

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

### 2. Icons

Pass icons as templ.Components:

```go
func InfoIcon() templ.Component {
    return templ.Raw(`<svg xmlns="http://www.w3.org/2000/svg" class="size-5" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
    </svg>`)
}
```

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

Always test your component implementations:

```go
// In your project tests
func TestAccordion_Integration(t *testing.T) {
    // Use Goshtoso's visual testing utilities
    // See Goshtoso tests for examples
}
```

## Known Pitfalls

### HTMX History Cache vs Alpine.js State

When using HTMX SPA navigation (`hx-get` + `hx-target="#main-content-area"` + `hx-push-url`), HTMX caches the raw `document.body.innerHTML` for back-button history restore. The problem: Alpine-generated DOM nodes (from `x-for`, `x-text`, etc.) are saved in the cache, but Alpine scope objects are lost. On back-button restore, the page shows stale Alpine-generated elements with no reactivity — combobox dropdowns with blank items, broken toggles, etc.

**Recommended approaches (pick one per use case):**

1. **`LinkMode: LinkBoost`** on table rows — swaps the full `<body>` via `hx-select="body"` + `hx-target="body"`. Back-button re-fetches from server, so Alpine re-initializes cleanly. No stale cache.

2. **`LinkMode: LinkFull`** on table rows — plain `window.location.href` navigation. Simplest, safest. Use when the target page has complex Alpine state.

3. **`hx-history="false"`** on a container — tells HTMX not to cache this page. Back-button will fetch from server. Useful when you can't control the navigation source.

4. **Alpine re-init on history restore** — listen for `htmx:historyRestore` and call `Alpine.initTree(document.body)`. Works in theory but is fragile: HTMX strips `<script>` tags from cached HTML, so Alpine data registrations may be missing.

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

1. Check that Tailwind CSS is processing the component files
2. Verify `all-themes.css` is imported
3. Ensure the `data-theme` attribute is set on `<html>`

### Alpine.js not working?

1. Verify Alpine.js is loaded before components render
2. Check browser console for Alpine.js errors
3. For collapse animations, ensure `@alpinejs/collapse` is loaded

### Dark mode not working?

1. Add `dark` class to `<html>` element
2. Verify CSS custom properties are defined in `all-themes.css`
3. Check that `dark:` prefixes are in component classes

## Examples

See the `/components` directory in the Goshtoso repository for complete examples of each component with visual parity tests.

Run the demo server:

```bash
cd /path/to/goshtoso
go run cmd/server/main.go -port 8090
```

Then visit:
- http://localhost:8090/components/button
- http://localhost:8090/components/accordion

## Contributing

To add new components:

1. Create component in `components/<name>/`
2. Copy original HTML to `fixtures/`
3. Create demo page in `internal/pages/demo/components/`
4. Write E2E tests with visual parity checks
5. Document in this guide

## License

MIT License (same as PenguinUI)

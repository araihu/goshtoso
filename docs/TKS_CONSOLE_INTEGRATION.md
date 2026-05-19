# tks-console Integration Guide

This guide demonstrates how to integrate Goshtoso components into the tks-console project.

## Quick Start

### 1. Add Goshtoso Dependency

```bash
cd /path/to/tks-console
go get github.com/araihu/goshtoso@latest
```

Or manually add to `go.mod`:

```go
require (
    // ... existing dependencies ...
    github.com/araihu/goshtoso v0.1.0
)
```

Then run:

```bash
go mod tidy
```

### 2. Extract Goshtoso CSS

tks-console uses Goshtoso's CLI tool to extract the pre-built CSS:

```bash
# Already registered in go.mod as a tool dependency
go tool goshtoso -out=css/goshtoso-base.css

# Or via Makefile
make sync-goshtoso-css
```

This replaces the old manual `cp` approach. The extracted CSS is imported in `css/main.css` and is gitignored as a build artifact. `make generate` runs this automatically.

### 3. Test Component

Create a test page to verify the component works:

```go
// internal/ports/http/ssr/pages/test_gottha.templ
package pages

import (
    "github.com/araihu/goshtoso/components/accordion"
    "github.com/cloud104/tks-console/internal/ports/http/ssr/components/page"
)

templ TestGoshtosoAccordion() {
    @page.Base(page.BaseData{
        Title: "Goshtoso Accordion Test",
    }) {
        <div class="max-w-2xl mx-auto p-8">
            <h1 class="text-2xl font-bold mb-6">Goshtoso Accordion Test</h1>
            
            @accordion.Accordion(accordion.AccordionConfig{
                Items: []accordion.AccordionItem{
                    {
                        ID:      "test1",
                        Title:   "Test Section 1",
                        Content: TestContent1(),
                    },
                    {
                        ID:      "test2",
                        Title:   "Test Section 2",
                        Content: TestContent2(),
                    },
                },
            })
        </div>
    }
}

templ TestContent1() {
    <p>This is test content for section 1.</p>
}

templ TestContent2() {
    <p>This is test content for section 2.</p>
}
```

Add route in your handler:

```go
// In your router setup
r.Get("/dev/test-gottha-accordion", func(w http.ResponseWriter, r *http.Request) {
    pages.TestGoshtosoAccordion().Render(r.Context(), w)
})
```

## Real-World Usage Examples

### Example 1: Cluster Details Accordion

Replace the existing cluster details sections with an accordion:

```go
// Before: Multiple separate sections
// After: Organized accordion

templ ClusterDetailsAccordion(cluster domain.Cluster) {
    @accordion.Accordion(accordion.AccordionConfig{
        AllowMultiple: true,
        Items: []accordion.AccordionItem{
            {
                ID:      "overview",
                Title:   "Overview",
                Content: ClusterOverview(cluster),
            },
            {
                ID:      "node-pools",
                Title:   fmt.Sprintf("Node Pools (%d)", len(cluster.NodePools)),
                Content: NodePoolsSection(cluster.NodePools),
            },
            {
                ID:      "addons",
                Title:   fmt.Sprintf("Addons (%d)", len(cluster.Addons)),
                Content: AddonsSection(cluster.Addons),
            },
            {
                ID:      "networking",
                Title:   "Networking",
                Content: NetworkingSection(cluster.NetworkConfig),
            },
            {
                ID:      "security",
                Title:   "Security",
                Content: SecuritySection(cluster.SecurityConfig),
            },
        },
    })
}
```

### Example 2: Create Cluster Form - Step Sections

Use accordion to organize a multi-step form:

```go
templ CreateClusterForm(data dto.CreateClusterData) {
    <form id="create-cluster-form" hx-post="/api/clusters" hx-target="#form-result">
        @accordion.Accordion(accordion.AccordionConfig{
            ID:            "create-cluster-accordion",
            AllowMultiple: true,
            Items: []accordion.AccordionItem{
                {
                    ID:                "basic-info",
                    Title:             "1. Basic Information",
                    Content:           BasicInfoSection(data),
                    InitiallyExpanded: true,
                },
                {
                    ID:      "node-pools",
                    Title:   "2. Node Pools",
                    Content: NodePoolsSection(data.NodePools),
                },
                {
                    ID:      "networking",
                    Title:   "3. Networking",
                    Content: NetworkingSection(data.Networking),
                },
                {
                    ID:      "addons",
                    Title:   "4. Addons",
                    Content: AddonsSection(data.Addons),
                },
            },
        })
        
        <div class="mt-6 flex justify-end gap-3">
            <button type="button" class="btn-secondary" onclick="history.back()">
                Cancel
            </button>
            <button type="submit" class="btn-primary">
                Create Cluster
            </button>
        </div>
    </form>
}
```

### Example 3: Provider Details - Collapsible Sections

Replace the provider tabs with an accordion for better mobile experience:

```go
templ ProviderDetailsAccordion(provider domain.Provider) {
    @accordion.Accordion(accordion.AccordionConfig{
        Variant: accordion.NoBackground,
        Items: []accordion.AccordionItem{
            {
                ID:      "instance-types",
                Title:   "Instance Types",
                Icon:    icons.ServerIcon(),
                Content: InstanceTypesTable(provider.InstanceTypes),
            },
            {
                ID:      "regions",
                Title:   "Regions",
                Icon:    icons.GlobeIcon(),
                Content: RegionsTable(provider.Regions),
            },
            {
                ID:      "pricing",
                Title:   "Pricing",
                Icon:    icons.CurrencyIcon(),
                Content: PricingTable(provider.Pricing),
            },
        },
    })
}
```

### Example 4: Settings Page - Organized Sections

Use accordion to organize settings into collapsible sections:

```go
templ SettingsPage(settings domain.UserSettings) {
    @accordion.Accordion(accordion.AccordionConfig{
        ID: "settings-accordion",
        Items: []accordion.AccordionItem{
            {
                ID:      "profile",
                Title:   "Profile Settings",
                Content: ProfileSettings(settings.Profile),
            },
            {
                ID:      "notifications",
                Title:   "Notifications",
                Content: NotificationSettings(settings.Notifications),
            },
            {
                ID:      "security",
                Title:   "Security",
                Content: SecuritySettings(settings.Security),
            },
            {
                ID:      "api-keys",
                Title:   "API Keys",
                Content: APIKeysSection(settings.APIKeys),
            },
            {
                ID:      "billing",
                Title:   "Billing",
                Content: BillingSection(settings.Billing),
            },
        },
    })
}
```

## Migration from Existing tks-console Components

### Current Implementation Analysis

Looking at existing tks-console components in `internal/ports/http/ssr/components/`:

**Current Components:**
- `badge/` - Custom implementation ✓ (compatible)
- `modal/` - Custom implementation ✓ (compatible)
- `toast/` - Custom implementation ✓ (compatible)
- `table/` - Custom implementation ✓ (compatible)
- `sidebar/` - Custom implementation ✓ (compatible)
- `form/` - Complex forms (can use alongside Goshtoso)

**New Components from Goshtoso:**
- `accordion/` - ✓ New addition
- `button/` - Can supplement existing
- Future: `alert/`, `card/`, `spinner/`, etc.

### Migration Strategy

1. **Phase 1: Add New Components**
   - Accordion for cluster details, settings pages
   - Keep existing components unchanged

2. **Phase 2: Gradual Replacement**
   - Replace custom implementations with Goshtoso versions
   - Test visual parity before replacing

3. **Phase 3: Standardize**
   - Use Goshtoso components as primary library
   - Keep custom components only for tks-specific features

## Testing Integration

### Visual Regression Tests

Add tests to verify Goshtoso components render correctly in tks-console:

```go
// tests/e2e/gottha_integration_test.go
package e2e

import (
    "testing"
    
    "github.com/playwright-community/playwright-go"
    "github.com/stretchr/testify/require"
)

func TestGoshtosoAccordion_InTksConsole(t *testing.T) {
    // Setup
    pool, err := dockertest.NewPool("")
    require.NoError(t, err)
    
    env, err := bootstrapE2ERuntime(t, pool)
    require.NoError(t, err)
    
    // Launch browser
    pw, err := playwright.Run()
    require.NoError(t, err)
    defer pw.Stop()
    
    browser, err := pw.Chromium.Launch()
    require.NoError(t, err)
    defer browser.Close()
    
    page, err := browser.NewPage()
    require.NoError(t, err)
    
    // Login
    loginWithOIDC(t, page, env.ConsoleURL)
    
    // Navigate to page using Goshtoso accordion
    _, err = page.Goto(env.ConsoleURL+"/dev/test-gottha-accordion")
    require.NoError(t, err)
    
    // Test accordion renders
    accordion := page.Locator(".divide-y")
    visible, err := accordion.IsVisible()
    require.NoError(t, err)
    require.True(t, visible, "accordion should be visible")
    
    // Test interaction
    firstButton := accordion.Locator("button").First()
    err = firstButton.Click()
    require.NoError(t, err)
    
    // Verify expanded
    ariaExpanded, err := firstButton.GetAttribute("aria-expanded")
    require.NoError(t, err)
    require.Equal(t, "true", ariaExpanded)
}
```

## Best Practices for tks-console

### 1. Consistent Import Pattern

Create an internal wrapper for commonly used components:

```go
// internal/components/gottha.go
package components

import (
    gotthabutton "github.com/araihu/goshtoso/components/button"
    gotthaaccordion "github.com/araihu/goshtoso/components/accordion"
)

// Re-export with tks-console defaults
type ButtonConfig = gotthabutton.Config
type AccordionConfig = gotthaaccordion.AccordionConfig
type AccordionItem = gotthaaccordion.AccordionItem

func Button(cfg ButtonConfig) templ.Component {
    return gotthabutton.Button(cfg)
}

func Accordion(cfg AccordionConfig) templ.Component {
    return gotthaaccordion.Accordion(cfg)
}
```

### 2. Theme Compatibility

tks-console uses these CSS variables (verify they match Goshtoso):

```css
/* Check these variables are defined */
--color-primary
--color-surface
--color-surface-alt
--color-outline
--radius-radius

/* Dark mode variants */
--color-primary-dark
--color-surface-dark
--color-surface-dark-alt
--color-outline-dark
```

### 3. Icon System

tks-console uses SVG icons. Pass them to Goshtoso components:

```go
import "github.com/cloud104/tks-console/internal/ports/http/ssr/components/icons"

@accordion.Accordion(accordion.AccordionConfig{
    Items: []accordion.AccordionItem{
        {
            Title: "Settings",
            Icon:  icons.SpriteIcon("settings", "size-5"),
            Content: SettingsContent(),
        },
    },
})
```

### 4. HTMX Integration

Goshtoso components work with tks-console's HTMX patterns:

```go
// Existing tks-console HTMX pattern
hx-post="/api/clusters/validate"
hx-target="#validation-result"
hx-swap="outerHTML"

// Works seamlessly with Goshtoso components
```

## Troubleshooting

### Build Errors

If you get import errors:

```bash
# Clean and rebuild
go clean -cache
go mod tidy
go build ./...
```

### Styling Issues

If components don't look right:

1. Check Tailwind CSS version (v4+ required)
2. Verify theme variables are loaded
3. Compare computed styles with browser DevTools

### Alpine.js Conflicts

tks-console already uses Alpine.js. Goshtoso components use standard Alpine.js patterns and should work seamlessly. If you see conflicts:

1. Check Alpine.js version compatibility
2. Ensure `x-cloak` CSS is defined
3. Verify no duplicate Alpine.js instances

## Next Steps

1. **Test Integration**
   - Run the test page: `/dev/test-gottha-accordion`
   - Verify visual parity with existing tks-console components

2. **Gradual Adoption**
   - Start with new features using Goshtoso components
   - Replace existing components over time

3. **Contribute Back**
   - Report any issues to Goshtoso repository
   - Suggest new components based on tks-console needs

## Support

- Goshtoso Repository: https://github.com/araihu/goshtoso
- Issues: Create issue in Goshtoso repo for component bugs
- Discussions: Use tks-console team channels for integration questions

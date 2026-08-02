# E2E Tests

This directory contains end-to-end tests using Playwright for browser automation.

## Running Tests

### Run all E2E tests
```bash
make test-e2e
```

### Run only impacted component and example identities

All Go unit tests remain unconditional. For Playwright, compare committed
changes with a base revision and run the directly affected identities plus
their reverse Go-package consumers:

```bash
make test-e2e-focused E2E_BASE=origin/main
```

The selector writes `.e2e-impact.json`. Unknown paths, shared runtime/theme
changes, generated-only diffs, deletions, renames, and unsafe history select
`e2e,full` rather than risking a false-negative focused run.

### Run specific test
```bash
go test -tags=e2e,full ./tests/e2e/... -v -run TestAccordion_StaticContent
```

### Run in short mode (skip E2E)
```bash
go test -tags=e2e,full ./tests/e2e/... -short
```

## Test Coverage

### Components Tests (`components_test.go`)

#### Accordion Tests
- `TestAccordion_StaticContent` - Tests accordion expand/collapse with static content
- `TestAccordion_ServerLoadedContent` - Tests HTMX lazy loading functionality
- `TestAccordion_AllVariants` - Tests all accordion variants (Default, NoBackground, ServerLoaded)
- `TestAccordion_Visual_Parity` - Visual regression tests with screenshots

#### Button Tests
- `TestButton_HTMXInteractions` - Tests HTMX POST/GET requests, loading states, confirm dialogs
- `TestButton_Variants_Render_Correctly` - Verifies all 8 button variants render

#### Integration Tests
- `TestComponent_DarkMode` - Tests dark mode toggle and persistence
- `TestAPIEndpoints` - Direct API endpoint testing
- `TestPerformance` - Performance benchmarks
- `TestIntegration` - Full user workflows
- `TestErrorHandling` - Error scenarios

### Visual Tests (`visual_helpers.go`)

Screenshot comparison utilities for visual regression testing.

## Test Structure

Each test follows this pattern:

1. **Setup** - Start server, initialize Playwright
2. **Action** - Navigate, interact with components
3. **Assertion** - Verify expected behavior
4. **Cleanup** - Stop server, close browser

## Writing New Tests

Example test structure:

```go
func TestYourFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping E2E test in short mode")
    }

    // Setup
    cleanupServer := setupServer(t)
    defer cleanupServer()

    _, browser, cleanupPW := setupPlaywright(t)
    defer cleanupPW()

    page, err := browser.NewPage()
    require.NoError(t, err)

    // Navigate
    _, err = page.Goto(baseURL+"/your-page", playwright.PageGotoOptions{
        WaitUntil: playwright.WaitUntilStateNetworkidle,
    })
    require.NoError(t, err)

    // Test
    t.Run("Your_Subtest", func(t *testing.T) {
        // Interact
        element := page.Locator(".your-selector")
        err := element.Click()
        require.NoError(t, err)

        // Assert
        visible, err := element.IsVisible()
        require.NoError(t, err)
        assert.True(t, visible)
    })
}
```

## Utilities

### Screenshot Comparison
```go
config := ScreenshotConfig{
    OriginalURL:    baseURL + "/original",
    GoshtosoURL:      baseURL + "/gottha",
    ComponentName:  "your-component",
    Threshold:      0.95, // 95% match required
}
result := CompareScreenshots(t, config)
```

### Class Verification
```go
htmlResult := ExtractAndCompareHTML(t, page, 
    "original-selector",
    "gottha-selector")
PrintComparisonReport(t, htmlResult, nil)
```

## Continuous Integration

To run in CI/CD:

```bash
# Install Playwright browsers
make install-playwright

# Run all E2E tests
go test -tags=e2e,full ./tests/e2e/... -v
```

## Debugging

### View Screenshots
Failed tests save screenshots to `test-results/screenshots/`

### View Test Output
Run with verbose flag:
```bash
go test -tags=e2e,full ./tests/e2e/... -v 2>&1 | tee test-output.log
```

### Run Single Test
```bash
go test -tags=e2e,full ./tests/e2e/... -v -run TestAccordion_StaticContent/Accordion_Expands_And_Collapses
```

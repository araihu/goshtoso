# E2E Tests

This directory contains end-to-end tests using Playwright for browser automation.

## Running Tests

### Run all E2E tests
```bash
just test-e2e
```

### Run only impacted component and example identities

All Go unit tests remain unconditional. For Playwright, compare committed
changes with a base revision and run the directly affected identities plus
their reverse Go-package consumers:

```bash
just test-e2e-focused origin/main
```

The selector writes `.e2e-impact.json`. Unknown paths, shared runtime/theme
changes, generated-only diffs, deletions, renames, and unsafe history select
`e2e,full` rather than risking a false-negative focused run.

In CI, the selected suite runs once as normal. If every failed top-level test
failed on a Playwright timeout, the launcher retries only those exact tests
once. The classifier recognizes timeout error formats emitted before and after
the Go Playwright binding migration. Assertion failures and non-timeout
failures are never retried, and a second timeout remains a failed check. Local
runs stay strict and do not retry.

### Run specific test
```bash
go test -tags=e2e,full ./tests/e2e/... -v -run TestAccordion_StaticContent
```

### Check native dialog layout in WebKit

This opt-in smoke test checks that compact, desktop and fullscreen dialogs expose
their content in Playwright WebKit. It does not reproduce the content-height
collapse observed in native Safari 26.6.2: Playwright WebKit 26.5 passes with
both `flex-1` and `flex-auto`. The default CI browser is Chromium, and the shared
test setup also requires it for this command.

```bash
go run github.com/mxschmitt/playwright-go/cmd/playwright@v0.6201.1 install chromium webkit
go test -tags=e2e,webkit ./tests/e2e -run TestDialogContentWebKit -count=1
```

Also check native Safari in Responsive Design Mode at 390 × 844. Open the modal
demo's Quick note dialog, enter text, dismiss it, and reopen it. The field and
Done button must stay visible and interactive. Repeat with the fullscreen demo.
Open with Tab and Enter, then press Escape and check that focus returns to the
previously focused button. A Safari mouse click need not focus a button first.
For consumers with compact search dialogs, open search, enter a query, check
matching results, then dismiss search and open navigation. A dimmed page with
an invisible dialog is a failure even if the input appears in accessibility tools.

### Prove a generated consumer-local icon pack

This focused browser test creates a verified release fixture, generates a
temporary consumer package, serves its sprite over HTTP, and renders the
generated helper through Goshtoso's core `components/icon` path:

```bash
go test -tags=e2e,iconpack ./tests/e2e -v -run TestIconpackGeneratedConsumerBrowserProof
```

The arbitrary-source proof uses a Goshtoso-owned `.iconpack.yaml`, an
explicit `.iconpack.lock.yaml` first-trust boundary, and the same browser
assertions through `components/icon`:

```bash
go test -tags=e2e,iconpack ./tests/e2e -v -run TestIconpackConfigGeneratedConsumerBrowserProof
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
- `TestAccordion_Visual_Parity` - Verifies the rendered accordion uses the expected classes

#### Button Tests
- `TestButton_HTMXInteractions` - Tests HTMX POST/GET requests, loading states, confirm dialogs
- `TestButton_Variants_Render_Correctly` - Verifies all 8 button variants render

#### Integration Tests
- `TestComponent_DarkMode` - Tests dark mode toggle and persistence
- `TestAPIEndpoints` - Direct API endpoint testing
- `TestPerformance` - Performance benchmarks
- `TestIntegration` - Full user workflows
- `TestErrorHandling` - Error scenarios

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

### Tailwind Class Verification

```go
VerifyTailwindClasses(t, page.Locator(".your-selector"), []string{
    "flex",
    "items-center",
})
```

## Continuous Integration

To run in CI/CD:

```bash
# Install Playwright browsers
just install-playwright

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

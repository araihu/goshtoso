//go:build e2e && (full || button)

package e2e

import (
	"strings"
	"sync"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// waitForText waits until the element matched by selector (a CSS selector that
// may include a :has-text() filter) becomes visible.
func waitForText(t *testing.T, page playwright.Page, selector string) {
	t.Helper()
	require.NoError(t, page.Locator(selector).First().WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
}

// TestButtonCoverageDemo exercises the /components/button demo: variant/size/
// disabled containers, the disabled boolean attribute, and the three HTMX-wired
// buttons (POST, GET, DELETE-with-confirm) that swap server fragments into their
// result targets. Console errors are captured inline per harness convention.
func TestButtonCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	// Capture any console errors surfaced during the run.
	var (
		mu            sync.Mutex
		consoleErrors []string
	)
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			mu.Lock()
			consoleErrors = append(consoleErrors, msg.Text())
			mu.Unlock()
		}
	})

	// Auto-accept the hx-confirm dialog so the DELETE request fires.
	page.On("dialog", func(d playwright.Dialog) {
		_ = d.Accept()
	})

	_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	t.Run("PageAndContainersRender", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		require.Contains(t, title, "Buttons")

		for _, id := range []string{
			"#button-variants",
			"#button-sizes",
			"#button-disabled",
			"#button-form-actions",
			"#button-htmx",
		} {
			visible, err := page.Locator(id).IsVisible()
			require.NoError(t, err)
			require.True(t, visible, "container %s should be visible", id)
		}
	})

	t.Run("NativeFormActionsRender", func(t *testing.T) {
		back := page.Locator(`#button-form-actions button[name="action"][value="back"]`)
		advance := page.Locator(`#button-form-actions button[name="action"][value="continue"]`)
		for _, action := range []playwright.Locator{back, advance} {
			visible, err := action.IsVisible()
			require.NoError(t, err)
			require.True(t, visible)
		}
	})

	t.Run("AllEightVariantsRender", func(t *testing.T) {
		count, err := page.Locator("#button-fragment button").Count()
		require.NoError(t, err)
		require.Equal(t, 8, count, "expected 8 variant buttons")
	})

	t.Run("SizeButtonsRender", func(t *testing.T) {
		cases := []struct {
			size  string
			label string
		}{
			{"sm", "Small"},
			{"md", "Medium"},
			{"lg", "Large"},
			{"xl", "Extra Large"},
		}
		for _, tc := range cases {
			require.NoError(t, page.Locator("label[for='button-size-"+tc.size+"']").Click())
			require.NoError(t, page.Locator("[data-testid='button-size-selected']").Filter(playwright.LocatorFilterOptions{
				HasText: tc.size,
			}).WaitFor())
			require.NoError(t, page.Locator("[data-testid='button-size-preview-"+tc.size+"']").WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateVisible,
			}))
			loc := page.Locator("[data-testid='button-size-preview-"+tc.size+"'] button", playwright.PageLocatorOptions{HasText: tc.label})
			visible, err := loc.First().IsVisible()
			require.NoError(t, err)
			require.True(t, visible, "size button %q should render", tc.label)
		}
	})

	t.Run("DisabledButtonsHaveDisabledAttribute", func(t *testing.T) {
		first := page.Locator("#button-disabled button").First()
		disabled, err := first.IsDisabled()
		require.NoError(t, err)
		require.True(t, disabled, "disabled-section button should be disabled")
	})

	t.Run("HTMXPostSwapsResult", func(t *testing.T) {
		btn := page.Locator("#button-htmx button", playwright.PageLocatorOptions{HasText: "Send POST"})
		require.NoError(t, btn.Click())
		waitForText(t, page, "#htmx-result-post:has-text('Hello from HTMX')")
	})

	t.Run("HTMXGetSwapsResult", func(t *testing.T) {
		btn := page.Locator("#button-htmx button", playwright.PageLocatorOptions{HasText: "Send GET"})
		require.NoError(t, btn.Click())
		waitForText(t, page, "#htmx-result-get:has-text('Hello from HTMX')")
	})

	t.Run("HTMXDeleteWithConfirmSwapsResult", func(t *testing.T) {
		btn := page.Locator("#button-htmx button", playwright.PageLocatorOptions{HasText: "Delete"})
		require.NoError(t, btn.Click())
		waitForText(t, page, "#htmx-result-delete:has-text('Hello from HTMX')")
	})

	t.Run("NoConsoleErrors", func(t *testing.T) {
		mu.Lock()
		defer mu.Unlock()
		require.Empty(t, consoleErrors, "unexpected console errors: %s", strings.Join(consoleErrors, "; "))
	})
}

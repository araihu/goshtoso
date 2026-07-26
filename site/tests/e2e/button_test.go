package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestButton_GoshtosoComponent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("PageLoads", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		require.Contains(t, title, "Buttons", "page title should contain 'Buttons'")

		t.Logf("✓ Page loaded with title: %s", title)
	})

	t.Run("AllVariantsRender", func(t *testing.T) {
		// The button fragment has 8 buttons in a grid
		buttons := page.Locator("#button-fragment button")
		count, err := buttons.Count()
		require.NoError(t, err)
		require.GreaterOrEqual(t, count, 8, "expected at least 8 buttons")

		t.Logf("✓ Found %d Goshtoso buttons", count)
	})

	t.Run("PrimaryButtonExists", func(t *testing.T) {
		primary := page.Locator("#button-fragment button:has-text('Primary')")
		visible, err := primary.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "primary button should be visible")

		classAttr, err := primary.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, classAttr, "bg-primary", "should have bg-primary class")

		t.Logf("✓ Primary button exists with correct styling")
	})

	t.Run("DangerButtonExists", func(t *testing.T) {
		danger := page.Locator("#button-fragment button:has-text('Danger')")
		visible, err := danger.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "danger button should be visible")

		classAttr, err := danger.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, classAttr, "bg-danger", "should have bg-danger class")

		t.Logf("✓ Danger button exists with correct styling")
	})

	t.Run("AncestorFormOwnsLoadingAndDisablesSubmitter", func(t *testing.T) {
		form := page.Locator("#button-form-loading")
		submit := form.Locator("button[type=submit]")
		loading := submit.Locator(".goshtoso-loading-label")

		require.NoError(t, submit.Click())
		require.NoError(t, loading.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
		disabled, err := submit.IsDisabled()
		require.NoError(t, err)
		require.True(t, disabled, "form-owned HTMX request must disable its submitter")

		result := page.Locator("#button-form-loading-result")
		require.NoError(t, result.WaitFor())
		require.Eventually(t, func() bool {
			text, textErr := result.TextContent()
			return textErr == nil && strings.Contains(text, "Hello from HTMX")
		}, 3*time.Second, 20*time.Millisecond)
	})
}

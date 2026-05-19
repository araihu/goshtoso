package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestButton_OriginalPenguinUI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/original/buttons/default-button.html", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("AllVariantsPresent", func(t *testing.T) {
		buttons := page.Locator("button")
		count, err := buttons.Count()
		require.NoError(t, err)
		require.GreaterOrEqual(t, count, 8, "expected at least 8 buttons")

		for i := 0; i < count; i++ {
			button := buttons.Nth(i)
			visible, err := button.IsVisible()
			require.NoError(t, err)
			require.True(t, visible, "button %d should be visible", i)
		}

		t.Logf("✓ Found %d buttons", count)
	})

	t.Run("PrimaryButtonStyling", func(t *testing.T) {
		button := page.Locator("button").First()

		text, err := button.TextContent()
		require.NoError(t, err)
		require.Contains(t, strings.ToLower(text), "primary", "first button should be Primary")

		classAttr, err := button.GetAttribute("class")
		require.NoError(t, err)

		require.Contains(t, classAttr, "bg-primary", "should have bg-primary class")
		require.Contains(t, classAttr, "text-on-primary", "should have text-on-primary class")
		require.Contains(t, classAttr, "rounded-radius", "should have rounded-radius class")

		t.Logf("✓ Primary button has correct styling")
	})

	t.Run("ButtonIsClickable", func(t *testing.T) {
		button := page.Locator("button").First()

		disabled, err := button.IsDisabled()
		require.NoError(t, err)
		require.False(t, disabled, "button should be enabled")

		err = button.Click()
		require.NoError(t, err)

		t.Logf("✓ Button is clickable")
	})
}

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
}

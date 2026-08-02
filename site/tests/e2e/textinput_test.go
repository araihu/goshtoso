//go:build e2e

package e2e

import (
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestTextInput_GoshtosoComponent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/text-input", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	takeScreenshot(t, page, "text-input-demo")

	t.Run("PageLoads", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		require.Contains(t, title, "Text Input", "page title should contain 'Text Input'")
		t.Logf("Page loaded with title: %s", title)
	})

	t.Run("DefaultInputRenders", func(t *testing.T) {
		input := page.Locator("#textInputDefault")
		visible, err := input.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "default text input should be visible")

		// Check attributes
		inputType, err := input.GetAttribute("type")
		require.NoError(t, err)
		require.Equal(t, "text", inputType)

		placeholder, err := input.GetAttribute("placeholder")
		require.NoError(t, err)
		require.Equal(t, "Enter your name", placeholder)

		name, err := input.GetAttribute("name")
		require.NoError(t, err)
		require.Equal(t, "name", name)

		// Check styling classes
		classAttr, err := input.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, classAttr, "rounded-radius")
		require.Contains(t, classAttr, "border-control-outline")

		t.Logf("Default input renders correctly")
	})

	t.Run("ErrorStateRenders", func(t *testing.T) {
		input := page.Locator("#inputError")
		visible, err := input.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "error input should be visible")

		// Check border class for error state
		classAttr, err := input.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, classAttr, "border-danger")

		// Check error label has error icon (SVG)
		errorLabel := page.Locator("label[for='inputError']")
		classAttr, err = errorLabel.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, classAttr, "text-danger")

		// Check error icon SVG exists within label
		errorIcon := errorLabel.Locator("svg")
		count, err := errorIcon.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count, "error label should have an SVG icon")

		// Check helper text
		helperText := page.Locator("small.text-danger-text")
		count, err = helperText.Count()
		require.NoError(t, err)
		require.GreaterOrEqual(t, count, 1, "should have error helper text")

		t.Logf("Error state renders correctly")
	})

	t.Run("SuccessStateRenders", func(t *testing.T) {
		input := page.Locator("#inputSuccess")
		visible, err := input.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "success input should be visible")

		// Check border class for success state
		classAttr, err := input.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, classAttr, "border-success")

		// Check value
		value, err := input.GetAttribute("value")
		require.NoError(t, err)
		require.Equal(t, "John", value)

		// Check success label
		successLabel := page.Locator("label[for='inputSuccess']")
		classAttr, err = successLabel.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, classAttr, "text-success")

		t.Logf("Success state renders correctly")
	})

	t.Run("PasswordInputRenders", func(t *testing.T) {
		input := page.Locator("#passwordInput")
		visible, err := input.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "password input should be visible")

		// Check toggle button exists
		toggleBtn := page.Locator("button[aria-label='Show password']")
		count, err := toggleBtn.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count, "should have password toggle button")

		t.Logf("Password input renders correctly")
	})

	t.Run("PasswordToggleWorks", func(t *testing.T) {
		time.Sleep(150 * time.Millisecond) // Wait for Alpine.js

		input := page.Locator("#passwordInput")

		// Initially should be password type
		inputType, err := input.GetAttribute("type")
		require.NoError(t, err)
		require.Equal(t, "password", inputType)

		// Click toggle
		toggleBtn := page.Locator("button[aria-label='Show password']")
		err = toggleBtn.Click()
		require.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		// Should now be text type
		inputType, err = input.GetAttribute("type")
		require.NoError(t, err)
		require.Equal(t, "text", inputType)

		// Click toggle again
		err = toggleBtn.Click()
		require.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		// Should be back to password
		inputType, err = input.GetAttribute("type")
		require.NoError(t, err)
		require.Equal(t, "password", inputType)

		t.Logf("Password toggle works correctly")
	})

	t.Run("SearchInputRenders", func(t *testing.T) {
		input := page.Locator("input[type='search']")
		count, err := input.Count()
		require.NoError(t, err)
		require.GreaterOrEqual(t, count, 1, "should have search input")

		// Check search icon SVG is present near the input
		searchSection := page.Locator("div:has(> input[type='search'])")
		svgIcon := searchSection.Locator("svg")
		svgCount, err := svgIcon.Count()
		require.NoError(t, err)
		require.GreaterOrEqual(t, svgCount, 1, "search input should have search icon")

		t.Logf("Search input renders correctly")
	})

	t.Run("DisabledInputRenders", func(t *testing.T) {
		input := page.Locator("#disabledInput")
		visible, err := input.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "disabled input should be visible")

		disabled, err := input.IsDisabled()
		require.NoError(t, err)
		require.True(t, disabled, "disabled input should be disabled")

		t.Logf("Disabled input renders correctly")
	})

	t.Run("MaskedInputRenders", func(t *testing.T) {
		input := page.Locator("#phoneInput")
		visible, err := input.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "phone input should be visible")

		// Check x-mask attribute
		mask, err := input.GetAttribute("x-mask")
		require.NoError(t, err)
		require.Equal(t, "(999) 999-9999", mask)

		t.Logf("Masked input renders correctly")
	})

	t.Run("InputLabelsPresent", func(t *testing.T) {
		// Check that labels exist for inputs with IDs
		labels := page.Locator("label[for]")
		count, err := labels.Count()
		require.NoError(t, err)
		require.GreaterOrEqual(t, count, 4, "should have at least 4 labels")

		t.Logf("Found %d input labels", count)
	})
}

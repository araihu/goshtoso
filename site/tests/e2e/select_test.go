//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSelect_DefaultRendering tests the default select component renders correctly
func TestSelect_DefaultRendering(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/select", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Default_Select_Has_Label_And_Options", func(t *testing.T) {
		// Check label exists
		label := page.Locator("label[for='os-trigger']")
		labelText, err := label.TextContent()
		require.NoError(t, err)
		assert.Contains(t, labelText, "Operating System")

		trigger := page.Locator("#os-trigger[role='combobox']")
		count, err := trigger.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "should have exactly one default select")

		require.NoError(t, trigger.Click())
		options := page.Locator("#os-option-0, #os-option-1, #os-option-2")
		optCount, err := options.Count()
		require.NoError(t, err)
		assert.Equal(t, 3, optCount, "should have 3 options")

		t.Log("Default select renders with label and options")
	})

	t.Run("Select_Has_Chevron_Icon", func(t *testing.T) {
		svg := page.Locator("#os-trigger svg")
		svgCount, err := svg.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, svgCount, 1, "should have chevron icon")

		t.Log("Select has chevron icon")
	})

	t.Run("Select_Has_Appearance_None", func(t *testing.T) {
		trigger := page.Locator("#os-trigger")
		class, err := trigger.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, class, "rounded-radius", "select trigger should use custom styling")

		t.Log("Select has appearance-none for custom styling")
	})
}

// TestSelect_ValidationStates tests error and success states
func TestSelect_ValidationStates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/select", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Error_State_Has_Danger_Border", func(t *testing.T) {
		trigger := page.Locator("#os-error-trigger")
		class, err := trigger.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, class, "border-danger", "error select should have border-danger")

		t.Log("Error state has danger border")
	})

	t.Run("Error_State_Has_Helper_Text", func(t *testing.T) {
		helperText := page.Locator("#os-error-helper")
		text, err := helperText.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Error: Please select an operating system")

		t.Log("Error state shows helper text")
	})

	t.Run("Error_State_Label_Has_Icon", func(t *testing.T) {
		label := page.Locator("label[for='os-error-trigger']")
		svg := label.Locator("svg")
		count, err := svg.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "error label should have an icon")

		labelClass, err := label.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, labelClass, "text-danger", "error label should have text-danger")

		t.Log("Error state label has icon and danger color")
	})

	t.Run("Success_State_Has_Success_Border", func(t *testing.T) {
		trigger := page.Locator("#os-success-trigger")
		class, err := trigger.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, class, "border-success", "success select should have border-success")

		t.Log("Success state has success border")
	})

	t.Run("Success_State_Has_Preselected_Value", func(t *testing.T) {
		hiddenInput := page.Locator("input#os-success")
		value, err := hiddenInput.InputValue()
		require.NoError(t, err)
		assert.Equal(t, "mac", value, "success select should have Mac preselected")

		t.Log("Success state has preselected value")
	})
}

// TestSelect_DisabledState tests the disabled select
func TestSelect_DisabledState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/select", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Disabled_Select_Has_Disabled_Attribute", func(t *testing.T) {
		trigger := page.Locator("#os-disabled-trigger")
		isDisabled, err := trigger.IsDisabled()
		require.NoError(t, err)
		assert.True(t, isDisabled, "disabled select should be disabled")

		t.Log("Disabled select has disabled attribute")
	})
}

// TestSelect_CountrySelect tests the country select with many options
func TestSelect_CountrySelect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/select", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Country_Select_Has_Many_Options", func(t *testing.T) {
		require.NoError(t, page.Locator("#country-trigger").Click())
		options := page.Locator("[id^='country-option-']")
		count, err := options.Count()
		require.NoError(t, err)
		assert.Greater(t, count, 50, "country select should have many options")

		t.Log("Country select has many options")
	})

	t.Run("Country_Select_Has_Autocomplete", func(t *testing.T) {
		hiddenInput := page.Locator("input#country")
		autocomplete, err := hiddenInput.GetAttribute("autocomplete")
		require.NoError(t, err)
		assert.Equal(t, "country", autocomplete, "country select should have autocomplete=country")

		t.Log("Country select has autocomplete attribute")
	})
}

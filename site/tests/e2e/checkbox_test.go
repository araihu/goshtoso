//go:build e2e && (full || checkbox)

package e2e

import (
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckbox(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/checkbox", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Page_Loads", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		assert.Contains(t, title, "Checkbox")
		t.Log("Checkbox page loads correctly")
	})

	t.Run("Default_Checkbox_Renders", func(t *testing.T) {
		// Checked checkbox should be visible
		checkedInput := page.Locator("#defaultChecked")
		visible, err := checkedInput.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "default checked checkbox should be visible")

		// Verify it's checked
		checked, err := checkedInput.IsChecked()
		require.NoError(t, err)
		assert.True(t, checked, "default checkbox should be checked")

		// Unchecked checkbox
		uncheckedInput := page.Locator("#defaultUnchecked")
		unchecked, err := uncheckedInput.IsChecked()
		require.NoError(t, err)
		assert.False(t, unchecked, "unchecked checkbox should not be checked")

		t.Log("Default checkboxes render correctly")
	})

	t.Run("Checkbox_Toggle", func(t *testing.T) {
		input := page.Locator("#defaultUnchecked")

		// Click to check
		err := input.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		checked, err := input.IsChecked()
		require.NoError(t, err)
		assert.True(t, checked, "checkbox should be checked after click")

		// Click to uncheck
		err = input.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		checked, err = input.IsChecked()
		require.NoError(t, err)
		assert.False(t, checked, "checkbox should be unchecked after second click")

		t.Log("Checkbox toggling works correctly")
	})

	t.Run("Color_Variants_Render", func(t *testing.T) {
		variants := []string{
			"variantPrimary",
			"variantSecondary",
			"variantInfo",
			"variantSuccess",
			"variantWarning",
			"variantDanger",
		}

		for _, id := range variants {
			input := page.Locator("#" + id)
			visible, err := input.IsVisible()
			require.NoError(t, err)
			assert.True(t, visible, "%s checkbox should be visible", id)

			checked, err := input.IsChecked()
			require.NoError(t, err)
			assert.True(t, checked, "%s checkbox should be checked", id)
		}

		t.Log("All color variants render correctly")
	})

	t.Run("Custom_Icons_Render", func(t *testing.T) {
		icons := []string{"iconXmark", "iconMinus", "iconPlus"}

		for _, id := range icons {
			input := page.Locator("#" + id)
			visible, err := input.IsVisible()
			require.NoError(t, err)
			assert.True(t, visible, "%s checkbox should be visible", id)

			// Verify SVG icon is present
			svg := page.Locator("label[for='" + id + "'] svg")
			svgVisible, err := svg.IsVisible()
			require.NoError(t, err)
			assert.True(t, svgVisible, "%s should have an SVG icon", id)
		}

		t.Log("Custom icon checkboxes render correctly")
	})

	t.Run("Description_Checkbox_Renders", func(t *testing.T) {
		input := page.Locator("#descriptionCheckbox")
		visible, err := input.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "description checkbox should be visible")

		// Verify description text exists
		desc := page.Locator("#checkboxDescription")
		descVisible, err := desc.IsVisible()
		require.NoError(t, err)
		assert.True(t, descVisible, "description text should be visible")

		descText, err := desc.TextContent()
		require.NoError(t, err)
		assert.Contains(t, descText, "good news", "description should contain expected text")

		// Verify aria-describedby
		ariaDesc, err := input.GetAttribute("aria-describedby")
		require.NoError(t, err)
		assert.Equal(t, "checkboxDescription", ariaDesc, "should have aria-describedby attribute")

		t.Log("Description checkbox renders correctly")
	})

	t.Run("Container_Checkbox_Renders", func(t *testing.T) {
		input := page.Locator("#containerCheckbox")
		visible, err := input.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "container checkbox should be visible")

		// Verify container label has correct classes
		label := page.Locator("label[for='containerCheckbox']")
		classes, err := label.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classes, "border", "container should have border class")
		assert.Contains(t, classes, "justify-between", "container should use justify-between layout")

		t.Log("Container checkbox renders correctly")
	})

	t.Run("Checkbox_Group_Renders", func(t *testing.T) {
		// Verify group title
		groupTitle := page.Locator("h3:has-text('Notifications')")
		visible, err := groupTitle.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "group title should be visible")

		// Verify group items
		emailCheck := page.Locator("#groupEmail")
		emailChecked, err := emailCheck.IsChecked()
		require.NoError(t, err)
		assert.True(t, emailChecked, "email checkbox should be checked")

		pushCheck := page.Locator("#groupPush")
		pushChecked, err := pushCheck.IsChecked()
		require.NoError(t, err)
		assert.False(t, pushChecked, "push checkbox should not be checked")

		smsCheck := page.Locator("#groupSMS")
		smsChecked, err := smsCheck.IsChecked()
		require.NoError(t, err)
		assert.True(t, smsChecked, "sms checkbox should be checked")

		t.Log("Checkbox group renders correctly")
	})

	t.Run("Disabled_Checkbox", func(t *testing.T) {
		disabledInput := page.Locator("#disabledUnchecked")
		disabled, err := disabledInput.IsDisabled()
		require.NoError(t, err)
		assert.True(t, disabled, "disabled checkbox should be disabled")

		t.Log("Disabled checkbox renders correctly")
	})

	// Take screenshot for visual verification
	takeScreenshot(t, page, "checkbox-component")
}

//go:build e2e && (full || textarea)

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTextarea_GoshtosoComponent tests the Goshtoso textarea component
func TestTextarea_GoshtosoComponent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/textarea", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Page_Loads_Successfully", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		assert.Contains(t, title, "Textarea")
		t.Log("✓ Textarea page loads successfully")
	})

	t.Run("Default_Textarea_Renders", func(t *testing.T) {
		ta := page.Locator("textarea#demo-default")
		visible, err := ta.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "default textarea should be visible")

		classAttr, err := ta.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classAttr, "w-full")
		assert.Contains(t, classAttr, "rounded-radius")
		assert.Contains(t, classAttr, "border-control-outline")

		placeholder, err := ta.GetAttribute("placeholder")
		require.NoError(t, err)
		assert.Equal(t, "We'd love to hear from you...", placeholder)

		t.Log("✓ Default textarea renders with correct classes and placeholder")
	})

	t.Run("Minimal_Light_Textarea_Has_A_Visible_Boundary", func(t *testing.T) {
		contrast, err := page.Locator("textarea#demo-default").Evaluate(`element => {
			document.documentElement.dataset.theme = 'minimal'
			document.documentElement.classList.remove('dark')
			const canvas = document.createElement('canvas')
			canvas.width = 1
			canvas.height = 1
			const ctx = canvas.getContext('2d', { willReadFrequently: true })
			const pixel = color => {
				ctx.clearRect(0, 0, 1, 1)
				ctx.fillStyle = color
				ctx.fillRect(0, 0, 1, 1)
				return [...ctx.getImageData(0, 0, 1, 1).data]
			}
			const luminance = rgb => {
				const linear = rgb.map(value => {
					const channel = value / 255
					return channel <= 0.03928 ? channel / 12.92 : Math.pow((channel + 0.055) / 1.055, 2.4)
				})
				return 0.2126 * linear[0] + 0.7152 * linear[1] + 0.0722 * linear[2]
			}
			const style = getComputedStyle(element)
			const surfacePixel = pixel(style.backgroundColor)
			const borderPixel = pixel(style.borderTopColor)
			const alpha = borderPixel[3] / 255
			const renderedBorder = borderPixel.slice(0, 3).map((channel, index) =>
				channel * alpha + surfacePixel[index] * (1 - alpha)
			)
			const border = luminance(renderedBorder)
			const surface = luminance(surfacePixel.slice(0, 3))
			return (Math.max(border, surface) + 0.05) / (Math.min(border, surface) + 0.05)
		}`, nil)
		require.NoError(t, err)
		var ratio float64
		switch value := contrast.(type) {
		case float64:
			ratio = value
		case int:
			ratio = float64(value)
		default:
			require.Failf(t, "unexpected contrast ratio", "%T", contrast)
		}
		require.GreaterOrEqual(t, ratio, 3.0, "Minimal light control boundary must meet WCAG non-text contrast")
	})

	t.Run("Default_Textarea_Has_Label", func(t *testing.T) {
		label := page.Locator("label[for='demo-default']")
		text, err := label.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "Comment", text)
		t.Log("✓ Default textarea has correct label")
	})

	t.Run("Error_State_Renders", func(t *testing.T) {
		ta := page.Locator("textarea#demo-error")
		visible, err := ta.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "error textarea should be visible")

		classAttr, err := ta.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classAttr, "border-danger")

		// Check error label has icon
		label := page.Locator("label[for='demo-error']")
		labelClass, err := label.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, labelClass, "text-danger")

		// Check error icon exists
		icon := label.Locator("svg")
		iconCount, err := icon.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, iconCount, "error label should have an icon")

		// Check helper text
		helperText := page.Locator("textarea#demo-error ~ small")
		text, err := helperText.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Error:")

		t.Log("✓ Error state renders with danger border, icon, and helper text")
	})

	t.Run("Success_State_Renders", func(t *testing.T) {
		ta := page.Locator("textarea#demo-success")
		visible, err := ta.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "success textarea should be visible")

		classAttr, err := ta.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classAttr, "border-success")

		// Check success label
		label := page.Locator("label[for='demo-success']")
		labelClass, err := label.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, labelClass, "text-success")

		// Check success icon
		icon := label.Locator("svg")
		iconCount, err := icon.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, iconCount, "success label should have an icon")

		t.Log("✓ Success state renders with success border and icon")
	})

	t.Run("Disabled_Textarea_Renders", func(t *testing.T) {
		ta := page.Locator("textarea#demo-disabled")
		visible, err := ta.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "disabled textarea should be visible")

		disabled, err := ta.IsDisabled()
		require.NoError(t, err)
		assert.True(t, disabled, "textarea should be disabled")

		t.Log("✓ Disabled textarea renders correctly")
	})

	t.Run("Required_Textarea_Renders_Native_Constraint", func(t *testing.T) {
		ta := page.Locator("textarea#demo-required")
		required, err := ta.GetAttribute("required")
		require.NoError(t, err)
		assert.Equal(t, "", required)
	})

	t.Run("Actions_Textarea_Renders", func(t *testing.T) {
		ta := page.Locator("textarea#demo-actions")
		visible, err := ta.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "actions textarea should be visible")

		rows, err := ta.GetAttribute("rows")
		require.NoError(t, err)
		assert.Equal(t, "6", rows)

		// Check send button exists
		sendBtn := page.Locator("button[aria-label='send']")
		btnVisible, err := sendBtn.IsVisible()
		require.NoError(t, err)
		assert.True(t, btnVisible, "send button should be visible")

		// Check action buttons exist
		actionBtns := page.Locator("button[aria-label='Emojis'], button[aria-label='Attach a file'], button[aria-label='Send voice']")
		count, err := actionBtns.Count()
		require.NoError(t, err)
		assert.Equal(t, 3, count, "should have 3 action buttons")

		t.Log("✓ Textarea with actions renders with buttons")
	})

	t.Run("Textarea_Accepts_Input", func(t *testing.T) {
		ta := page.Locator("textarea#demo-default")
		err := ta.Fill("Hello, this is a test message!")
		require.NoError(t, err)

		value, err := ta.InputValue()
		require.NoError(t, err)
		assert.Equal(t, "Hello, this is a test message!", value)

		t.Log("✓ Textarea accepts user input")
	})
}

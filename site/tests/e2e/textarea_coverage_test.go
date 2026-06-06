package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTextareaCoverageDemo drives the /components/textarea demo page to exercise
// browser-side behavior for the textarea component: variant rendering, the
// "with actions" footer buttons, live typing, and console cleanliness.
func TestTextareaCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, cleanup := setupPlaywright(t)
	defer cleanup()

	page := newPage(t, browser)

	// Collect any console errors raised while loading/interacting.
	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			consoleErrors = append(consoleErrors, msg.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/textarea", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	mainText, err := page.Locator("main").TextContent()
	require.NoError(t, err)
	assert.Contains(t, mainText, "Textarea")

	t.Run("Actions_Footer_Buttons_Are_Interactive", func(t *testing.T) {
		actionsTA := page.Locator("textarea#demo-actions")
		require.NoError(t, actionsTA.ScrollIntoViewIfNeeded())
		visible, err := actionsTA.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "actions textarea should be visible")

		for _, label := range []string{"Emojis", "Attach a file", "Send voice"} {
			btn := page.Locator("button[aria-label='" + label + "']")
			btnVisible, err := btn.IsVisible()
			require.NoError(t, err)
			assert.True(t, btnVisible, "%s button should be visible", label)
			// Buttons are type=button no-ops; clicking must not navigate or error.
			require.NoError(t, btn.Click())
		}

		sendBtn := page.Locator("button[aria-label='send']")
		sendVisible, err := sendBtn.IsVisible()
		require.NoError(t, err)
		assert.True(t, sendVisible, "send button should be visible")
		require.NoError(t, sendBtn.Click())
	})

	t.Run("Actions_Textarea_Accepts_Input", func(t *testing.T) {
		actionsTA := page.Locator("textarea#demo-actions")
		require.NoError(t, actionsTA.Fill("voice memo draft"))
		val, err := actionsTA.InputValue()
		require.NoError(t, err)
		assert.Equal(t, "voice memo draft", val)
	})

	t.Run("State_Variants_Use_Distinct_Borders", func(t *testing.T) {
		errClass, err := page.Locator("textarea#demo-error").GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, errClass, "border-danger")

		okClass, err := page.Locator("textarea#demo-success").GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, okClass, "border-success")
	})

	t.Run("No_Console_Errors", func(t *testing.T) {
		assert.Empty(t, consoleErrors, "console errors: %s", strings.Join(consoleErrors, "; "))
	})
}

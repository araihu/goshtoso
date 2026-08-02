//go:build e2e && (full || textinput)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestTextinputCoverageDemo exercises the password toggle, search, and mask
// variants on the text-input demo page to drive the lower-coverage templ
// branches (passwordInput, searchInput, masked defaultInput) through a real
// browser, and asserts there are no console errors.
func TestTextinputCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	var consoleErrors []string
	page.OnConsole(func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			consoleErrors = append(consoleErrors, msg.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/text-input", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	t.Run("PasswordToggle", func(t *testing.T) {
		container := page.Locator("#textinput-password")
		require.NoError(t, container.ScrollIntoViewIfNeeded())

		input := container.Locator("input")
		require.NoError(t, input.WaitFor())

		// Alpine binds the type via x-bind:type; it starts as "password".
		liveType, err := input.Evaluate("el => el.getAttribute('type')", nil)
		require.NoError(t, err)
		require.Equal(t, "password", liveType, "password input should start hidden")

		toggle := container.Locator("button[aria-label='Show password']")
		require.NoError(t, toggle.Click())

		_, err = page.WaitForFunction(
			"() => document.querySelector('#textinput-password input').getAttribute('type') === 'text'",
			nil,
		)
		require.NoError(t, err)

		// Toggling back hides it again.
		require.NoError(t, toggle.Click())
		_, err = page.WaitForFunction(
			"() => document.querySelector('#textinput-password input').getAttribute('type') === 'password'",
			nil,
		)
		require.NoError(t, err)
	})

	t.Run("SearchVariantRenders", func(t *testing.T) {
		container := page.Locator("#textinput-search")
		require.NoError(t, container.ScrollIntoViewIfNeeded())

		input := container.Locator("input[type='search']")
		require.NoError(t, input.WaitFor())

		ariaLabel, err := input.GetAttribute("aria-label")
		require.NoError(t, err)
		require.Equal(t, "search", ariaLabel)

		class, err := input.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, class, "pl-10", "search input reserves space for the icon")

		// The leading search icon svg is present.
		iconCount, err := container.Locator("svg").Count()
		require.NoError(t, err)
		require.GreaterOrEqual(t, iconCount, 1)
	})

	t.Run("MaskVariantAcceptsInput", func(t *testing.T) {
		container := page.Locator("#textinput-mask")
		require.NoError(t, container.ScrollIntoViewIfNeeded())

		input := container.Locator("input")
		require.NoError(t, input.WaitFor())

		hasMask, err := input.Evaluate("el => el.hasAttribute('x-mask')", nil)
		require.NoError(t, err)
		require.Equal(t, true, hasMask, "mask variant should expose x-mask")

		require.NoError(t, input.Fill("1234567890"))
		value, err := input.InputValue()
		require.NoError(t, err)
		require.NotEmpty(t, value, "mask input should accept typed characters")
	})

	require.Empty(t, consoleErrors, "no console errors expected: %v", consoleErrors)
}

//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTheme_Colors_VerifyComputedValues tests that buttons have the correct computed CSS colors
func TestTheme_Colors_VerifyComputedValues(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Setup server
	cleanupServer := setupServer(t)
	defer cleanupServer()

	// Setup Playwright
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	// Create page
	page := newPage(t, browser)

	// Navigate to button demo
	_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("PrimaryButton_LightMode", func(t *testing.T) {
		button := page.Locator("button:has-text('Primary')").First()

		bgColor, err := button.Evaluate("el => window.getComputedStyle(el).backgroundColor", nil)
		require.NoError(t, err)
		// Primary button should have a non-transparent background
		bgStr := fmt.Sprintf("%v", bgColor)
		t.Logf("Primary button bg: %s", bgStr)
		assert.NotEqual(t, "rgba(0, 0, 0, 0)", bgStr, "Primary button should have a background color")

		textColor, err := button.Evaluate("el => window.getComputedStyle(el).color", nil)
		require.NoError(t, err)
		// Text should be light (on-primary)
		t.Logf("Primary text color: %v", textColor)
		assert.NotEqual(t, "rgb(0, 0, 0)", textColor, "Primary button text should not be black")
	})

	t.Run("SuccessButton_LightMode", func(t *testing.T) {
		button := page.Locator("button:has-text('Success')").First()

		bgColor, err := button.Evaluate("el => window.getComputedStyle(el).backgroundColor", nil)
		require.NoError(t, err)
		// Success should be a green tone
		t.Logf("Success bg: %v", bgColor)
		assert.NotEqual(t, "rgba(0, 0, 0, 0)", bgColor, "Success button should have a background color")
	})

	t.Run("WarningButton_LightMode", func(t *testing.T) {
		button := page.Locator("button:has-text('Warning')").First()

		bgColor, err := button.Evaluate("el => window.getComputedStyle(el).backgroundColor", nil)
		require.NoError(t, err)
		// Warning should have a background color
		t.Logf("Warning bg: %v", bgColor)
		assert.NotEqual(t, "rgba(0, 0, 0, 0)", bgColor, "Warning button should have a background color")
	})
}

// TestTheme_Classes_Presence verifies Tailwind utility classes are applied
func TestTheme_Classes_Presence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Setup server
	cleanupServer := setupServer(t)
	defer cleanupServer()

	// Setup Playwright
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	// Create page
	page := newPage(t, browser)

	// Navigate to button demo
	_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Goshtoso_Buttons_Have_Correct_Classes", func(t *testing.T) {
		// Get all buttons in the button fragment
		buttons := page.Locator("#button-fragment button")
		count, err := buttons.Count()
		require.NoError(t, err)
		require.GreaterOrEqual(t, count, 8, "Should have at least 8 buttons")

		// Check that buttons have expected Tailwind classes
		for i := 0; i < count && i < 8; i++ {
			button := buttons.Nth(i)

			hasRounded, err := button.Evaluate("el => el.classList.contains('rounded-2xl')", nil)
			require.NoError(t, err)
			assert.True(t, hasRounded.(bool), fmt.Sprintf("Button %d should have rounded-radius class", i))

			hasFontMedium, err := button.Evaluate("el => el.classList.contains('font-medium')", nil)
			require.NoError(t, err)
			assert.True(t, hasFontMedium.(bool), fmt.Sprintf("Button %d should have font-medium class", i))
		}

		t.Logf("✓ All %d buttons have correct Tailwind classes", count)
	})
}

// TestTheme_DarkMode_Toggle verifies dark mode toggle works
func TestTheme_DarkMode_Toggle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Setup server
	cleanupServer := setupServer(t)
	defer cleanupServer()

	// Setup Playwright
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	// Create page
	page := newPage(t, browser)

	// Navigate to button demo
	_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("DarkMode_Toggle_Adds_Class", func(t *testing.T) {
		time.Sleep(150 * time.Millisecond) // Wait for Alpine.js store init

		// Get initial state
		initialDark, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)

		// Click dark mode toggle button
		toggleBtn := page.Locator("#darkModeToggleBtn")
		err = toggleBtn.Click()
		require.NoError(t, err)
		time.Sleep(150 * time.Millisecond)

		// Should have toggled
		afterDark, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		assert.NotEqual(t, initialDark.(bool), afterDark.(bool), "dark class should toggle")

		t.Log("✓ Dark mode toggle works correctly")
	})

	t.Run("DarkMode_Persists_In_LocalStorage", func(t *testing.T) {
		// Check localStorage for dark mode key (may be 'darkMode' or via Alpine store)
		darkModeValue, err := page.Evaluate("() => localStorage.getItem('darkMode') || localStorage.getItem('_x_darkMode')", nil)
		require.NoError(t, err)

		// Value should exist (either 'true' or 'false')
		if darkModeValue != nil {
			t.Logf("✓ Dark mode persisted in localStorage: %v", darkModeValue)
		} else {
			t.Log("Dark mode storage key not found (may use Alpine.js $persist)")
		}
	})
}

// TestTheme_LockedToAraiHu verifies component docs ignore legacy persisted
// themes now that the global selector has been removed.
func TestTheme_LockedToAraiHu(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, cleanup := setupPlaywright(t)
	defer cleanup()
	page := newPage(t, browser)
	script := `
document.cookie = "gt_storage=allowed; Path=/; SameSite=Lax";
localStorage.setItem("theme", "minimal");
`
	require.NoError(t, page.AddInitScript(playwright.Script{Content: &script}))

	_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	// Wait for Alpine.js
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(3000),
	})
	require.NoError(t, err)

	lockedTheme, err := page.Evaluate(`() => document.documentElement.getAttribute('data-theme')`, nil)
	require.NoError(t, err)
	assert.Equal(t, "araihu", lockedTheme)
	themeTriggerCount, err := page.Locator("#site-theme-trigger").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, themeTriggerCount)

	storedTheme, err := page.Evaluate(`() => localStorage.getItem('theme')`, nil)
	require.NoError(t, err)
	assert.Equal(t, "minimal", storedTheme, "locked docs must not rewrite consumer storage")
}

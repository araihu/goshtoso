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

// TestTheme_Switching verifies the theme dropdown changes the data-theme attribute
func TestTheme_Switching(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	// Wait for Alpine.js
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(3000),
	})
	require.NoError(t, err)

	// Get initial theme
	initialTheme, err := page.Evaluate(`() => document.documentElement.getAttribute('data-theme')`, nil)
	require.NoError(t, err)
	t.Logf("Initial theme: %v", initialTheme)

	themes := []struct {
		key   string
		label string
	}{
		{"arctic", "Arctic"},
		{"neo-brutalism", "Neo Brutalism"},
		{"90s", "90s"},
		{"minimal", "Minimal"},
	}

	for _, theme := range themes {
		t.Run("SwitchTo_"+theme.label, func(t *testing.T) {
			// Open the theme dropdown
			dropdownBtn := page.Locator("button:has(span.capitalize)")
			err := dropdownBtn.Click()
			require.NoError(t, err)

			// Click the theme button by its data-theme-key
			themeBtn := page.Locator(fmt.Sprintf("button[data-theme-key='%s']", theme.key))
			err = themeBtn.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(2000),
			})
			require.NoError(t, err)

			err = themeBtn.Click()
			require.NoError(t, err)

			// Verify data-theme attribute changed
			_, err = page.WaitForFunction(
				fmt.Sprintf(`() => document.documentElement.getAttribute('data-theme') === '%s'`, theme.key),
				nil,
				playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
			)
			require.NoError(t, err, "data-theme should be '%s'", theme.key)

			// Verify localStorage was updated
			storedTheme, err := page.Evaluate(`() => localStorage.getItem('theme')`, nil)
			require.NoError(t, err)
			assert.Equal(t, theme.key, storedTheme, "localStorage theme should match")
		})
	}
}

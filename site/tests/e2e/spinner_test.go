package e2e

import (
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSpinner_PageLoads tests that the spinner demo page loads correctly
func TestSpinner_PageLoads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/spinner", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Page_Title_Contains_Spinner", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		assert.Contains(t, title, "Spinner", "page title should contain Spinner")
		t.Log("✓ Page title contains Spinner")
	})

	t.Run("Spinner_SVGs_Are_Rendered", func(t *testing.T) {
		// Wait for Alpine.js
		time.Sleep(150 * time.Millisecond)

		// Check that spinner SVGs exist on the page
		spinners := page.Locator("svg.motion-safe\\:animate-spin")
		count, err := spinners.Count()
		require.NoError(t, err)
		assert.Greater(t, count, 0, "should have at least one spinner SVG on the page")
		t.Logf("✓ Found %d spinner SVGs on the page", count)
	})

	t.Run("Spinner_Has_Correct_Attributes", func(t *testing.T) {
		time.Sleep(150 * time.Millisecond)

		// Check first spinner has correct attributes
		firstSpinner := page.Locator("svg.motion-safe\\:animate-spin").First()
		ariaHidden, err := firstSpinner.GetAttribute("aria-hidden")
		require.NoError(t, err)
		assert.Equal(t, "true", ariaHidden, "spinner should have aria-hidden=true")

		viewBox, err := firstSpinner.GetAttribute("viewBox")
		require.NoError(t, err)
		assert.Equal(t, "0 0 24 24", viewBox, "spinner should have correct viewBox")

		t.Log("✓ Spinner has correct SVG attributes")
	})

	t.Run("Spinner_Color_Variants_Exist", func(t *testing.T) {
		time.Sleep(150 * time.Millisecond)

		// Check for color variant classes
		variants := []string{
			"fill-primary",
			"fill-secondary",
			"fill-info",
			"fill-success",
			"fill-warning",
			"fill-danger",
		}

		for _, variant := range variants {
			locator := page.Locator("svg." + variant)
			count, err := locator.Count()
			require.NoError(t, err)
			assert.Greater(t, count, 0, "should have spinner with class %s", variant)
		}
		t.Log("✓ All color variant spinners are present")
	})

	t.Run("Spinner_Size_Variants_Exist", func(t *testing.T) {
		time.Sleep(150 * time.Millisecond)

		// Check for size variant classes
		sizes := []string{"size-4", "size-5", "size-8", "size-12"}
		for _, size := range sizes {
			locator := page.Locator("svg." + size + ".motion-safe\\:animate-spin")
			count, err := locator.Count()
			require.NoError(t, err)
			assert.Greater(t, count, 0, "should have spinner with size class %s", size)
		}
		t.Log("✓ All size variant spinners are present")
	})
}

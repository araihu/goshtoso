//go:build e2e && (full || alert)

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertComponentDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/alert", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	t.Run("semantic variants render titles", func(t *testing.T) {
		defaults := page.Locator("#alert-default [role='alert']")
		count, err := defaults.Count()
		require.NoError(t, err)
		assert.Equal(t, 4, count)

		for _, title := range []string{
			"Update Available",
			"Successfully Subscribed",
			"Credit Card Expires Soon",
			"Invalid Email Address",
		} {
			require.NoError(t, page.Locator("#alert-default [role='alert']").Filter(playwright.LocatorFilterOptions{
				HasText: title,
			}).WaitFor())
		}
	})

	t.Run("dismissible close hides one alert", func(t *testing.T) {
		first := page.Locator("#alert-dismissible [role='alert']").First()
		require.NoError(t, first.Locator("button[aria-label='dismiss alert']").Click())
		require.NoError(t, first.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateHidden,
			Timeout: playwright.Float(3000),
		}))

		remaining, err := page.Locator("#alert-dismissible [role='alert']:visible").Count()
		require.NoError(t, err)
		assert.Equal(t, 3, remaining)
	})

	t.Run("link list and action variants expose their content", func(t *testing.T) {
		linkCount, err := page.Locator("#alert-link a").Filter(playwright.LocatorFilterOptions{HasText: "Details"}).Count()
		require.NoError(t, err)
		assert.Equal(t, 4, linkCount)

		for _, item := range []string{
			"has minimum 8 characters",
			"includes both upper and lower cases",
			"contains at least one number",
		} {
			count, err := page.Locator("#alert-list li").Filter(playwright.LocatorFilterOptions{HasText: item}).Count()
			require.NoError(t, err)
			assert.Equal(t, 2, count)
		}

		updateCount, err := page.Locator("#alert-action button").Filter(playwright.LocatorFilterOptions{HasText: "Update Now"}).Count()
		require.NoError(t, err)
		assert.Equal(t, 2, updateCount)

		dismissCount, err := page.Locator("#alert-action button").Filter(playwright.LocatorFilterOptions{HasText: "Dismiss"}).Count()
		require.NoError(t, err)
		assert.Equal(t, 4, dismissCount)
	})
}

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBadgeCoverageDemo complements TestBadgeComponentDemo with checks that the
// badge demo loads cleanly (no JS exceptions / console errors), that the
// notification and animating-dot helpers render their expected hooks, and that
// the badges survive a dark-mode toggle. filterIgnorable is defined in
// alert_coverage_test.go.
func TestBadgeCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)

	var pageErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/badge", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	t.Run("page heading renders", func(t *testing.T) {
		require.NoError(t, page.Locator("main").Filter(playwright.LocatorFilterOptions{
			HasText: "Badge",
		}).First().WaitFor())
	})

	t.Run("notification badge caps at 99 and exposes button labels", func(t *testing.T) {
		require.NoError(t, page.Locator("#badge-notification button[aria-label='notifications']").WaitFor())

		// NotificationBadge(count) caps display at "99" for count > 99.
		require.NoError(t, page.Locator("#badge-notification").GetByText("99", playwright.LocatorGetByTextOptions{
			Exact: playwright.Bool(true),
		}).WaitFor())

		// NotificationDot() renders a size-3 dot with no text.
		dots, err := page.Locator("#badge-notification span.size-3").Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, dots, 1)
	})

	t.Run("animating dots render aria-label and ping hook", func(t *testing.T) {
		dots := page.Locator("#badge-animating span[aria-label='notification']")
		count, err := dots.Count()
		require.NoError(t, err)
		assert.Equal(t, 6, count)

		// Each AnimatingDot emits an inner animate-ping span.
		ping, err := page.Locator("#badge-animating span.animate-ping").Count()
		require.NoError(t, err)
		assert.Equal(t, 6, ping)
	})

	t.Run("indicator dots are aria-hidden decorations", func(t *testing.T) {
		count, err := page.Locator("#badge-indicators span[aria-hidden='true']").Count()
		require.NoError(t, err)
		assert.Equal(t, 6, count)
	})

	t.Run("badges survive a dark-mode toggle", func(t *testing.T) {
		before, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		beforeBool, _ := before.(bool)

		_, err = page.Evaluate(`() => { document.documentElement.classList.toggle('dark'); }`, nil)
		require.NoError(t, err)

		after, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		afterBool, _ := after.(bool)
		assert.NotEqual(t, beforeBool, afterBool, "dark class should toggle")

		// Solid badges stay rendered after the theme switch.
		require.NoError(t, page.Locator("#badge-solid").GetByText("Primary", playwright.LocatorGetByTextOptions{
			Exact: playwright.Bool(true),
		}).First().WaitFor())
	})

	require.Empty(t, pageErrors, "uncaught JS exceptions on badge demo")
	require.Empty(t, filterIgnorable(consoleErrors), "console errors on badge demo")
}

package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAlertCoverageDemo complements TestAlertComponentDemo with checks that the
// alert demo loads cleanly (no JS exceptions / console errors), renders a
// variant icon per default alert, keeps content visible across a dark-mode
// toggle, and exposes live Alpine state on dismissible alerts.
func TestAlertCoverageDemo(t *testing.T) {
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

	_, err := page.Goto(baseURL+"/components/alert", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	t.Run("page heading renders", func(t *testing.T) {
		require.NoError(t, page.Locator("main").Filter(playwright.LocatorFilterOptions{
			HasText: "Alert",
		}).First().WaitFor())
	})

	t.Run("each default alert renders an icon badge svg", func(t *testing.T) {
		// defaultAlert/iconBadge emit one aria-hidden svg per alert. The default
		// section shows the four semantic variants.
		icons := page.Locator("#alert-default [role='alert'] svg[aria-hidden='true']")
		count, err := icons.Count()
		require.NoError(t, err)
		assert.Equal(t, 4, count)
	})

	t.Run("dismissible alert exposes live Alpine state", func(t *testing.T) {
		first := page.Locator("#alert-dismissible [role='alert']").First()
		require.NoError(t, first.WaitFor())
		// Read the live attribute via the DOM, not the static HTML.
		visible, err := first.Evaluate("el => el.getAttribute('x-show')", nil)
		require.NoError(t, err)
		assert.Equal(t, "alertIsVisible", visible)
	})

	t.Run("alerts survive a dark-mode toggle", func(t *testing.T) {
		before, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		beforeBool, _ := before.(bool)

		_, err = page.Evaluate(`() => { document.documentElement.classList.toggle('dark'); }`, nil)
		require.NoError(t, err)

		after, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		afterBool, _ := after.(bool)
		assert.NotEqual(t, beforeBool, afterBool, "dark class should toggle")

		// Default alerts stay rendered after the theme switch.
		require.NoError(t, page.Locator("#alert-default [role='alert']").First().WaitFor())
	})

	require.Empty(t, pageErrors, "uncaught JS exceptions on alert demo")
	require.Empty(t, filterIgnorable(consoleErrors), "console errors on alert demo")
}

// filterIgnorable drops console errors that are not caused by the component
// under test (e.g. unrelated network 404s for optional assets).
func filterIgnorable(errs []string) []string {
	var kept []string
	for _, e := range errs {
		if strings.Contains(e, "404") || strings.Contains(e, "favicon") {
			continue
		}
		kept = append(kept, e)
	}
	return kept
}

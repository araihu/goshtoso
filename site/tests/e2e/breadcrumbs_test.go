//go:build e2e && (full || breadcrumbs)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBreadcrumbsComponentDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/breadcrumbs", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	for _, id := range []string{"#breadcrumbs-chevron", "#breadcrumbs-slash", "#breadcrumbs-icon"} {
		nav := page.Locator(id + " nav[aria-label='breadcrumb']")
		require.NoError(t, nav.WaitFor())

		require.NoError(t, nav.Locator("a").Filter(playwright.LocatorFilterOptions{HasText: "Components"}).WaitFor())
		current := nav.Locator("[aria-current='page']")
		assert.Equal(t, "Breadcrumb", mustText(t, current))
	}

	slashCount, err := page.Locator("#breadcrumbs-slash span[aria-hidden='true']").Filter(playwright.LocatorFilterOptions{HasText: "/"}).Count()
	require.NoError(t, err)
	assert.Equal(t, 2, slashCount)

	chevronCount, err := page.Locator("#breadcrumbs-chevron svg[aria-hidden='true']").Count()
	require.NoError(t, err)
	assert.Equal(t, 2, chevronCount)

	homeTooltip, err := page.Locator("#breadcrumbs-icon a").First().Locator("span").First().GetAttribute("title")
	require.NoError(t, err)
	assert.Equal(t, "Home", homeTooltip)
}

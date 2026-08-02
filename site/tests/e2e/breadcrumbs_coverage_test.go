//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBreadcrumbsCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	var pageErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/breadcrumbs", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	require.NoError(t, page.Locator("main").Filter(playwright.LocatorFilterOptions{
		HasText: "Breadcrumbs",
	}).First().WaitFor())
	for _, id := range []string{"#breadcrumbs-chevron", "#breadcrumbs-slash", "#breadcrumbs-icon"} {
		require.NoError(t, page.Locator(id+" nav[aria-label='breadcrumb']").WaitFor())
		require.NoError(t, page.Locator(id+" [aria-current='page']").WaitFor())
	}

	chevronCount, err := page.Locator("#breadcrumbs-chevron svg[aria-hidden='true'][stroke='currentColor']").Count()
	require.NoError(t, err)
	assert.Equal(t, 2, chevronCount)

	slashCount, err := page.Locator("#breadcrumbs-slash span[aria-hidden='true']").Filter(playwright.LocatorFilterOptions{
		HasText: "/",
	}).Count()
	require.NoError(t, err)
	assert.Equal(t, 2, slashCount)

	iconTitle, err := page.Locator("#breadcrumbs-icon a").First().Locator("span").First().GetAttribute("title")
	require.NoError(t, err)
	assert.Equal(t, "Home", iconTitle)

	entry, ok := catalog.Lookup("components/breadcrumbs")
	require.True(t, ok)
	requireComponentGoAPILink(t, page, entry)

	before, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
	require.NoError(t, err)
	_, err = page.Evaluate(`() => document.documentElement.classList.toggle('dark')`, nil)
	require.NoError(t, err)
	after, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
	require.NoError(t, err)
	assert.NotEqual(t, before, after, "dark class should toggle while breadcrumbs remain mounted")
	require.NoError(t, page.Locator("#breadcrumbs-fragment").WaitFor())

	require.Empty(t, pageErrors, "uncaught JS exceptions on breadcrumbs demo: %v", pageErrors)
	require.Empty(t, filterIgnorable(consoleErrors), "console errors on breadcrumbs demo: %s", strings.Join(consoleErrors, "; "))
}

func TestBreadcrumbsCoverageFragmentNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/getting-started", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	link := page.Locator("a[href='/components/breadcrumbs']").First()
	require.NoError(t, link.ScrollIntoViewIfNeeded())
	require.NoError(t, link.Click())
	require.NoError(t, page.Locator("#breadcrumbs-fragment").WaitFor())
	require.NoError(t, page.Locator("#breadcrumbs-slash nav[aria-label='breadcrumb']").WaitFor())
	require.NoError(t, page.Locator("main").Filter(playwright.LocatorFilterOptions{
		HasText: "Slash Separator",
	}).First().WaitFor())

	require.Empty(t, filterIgnorable(consoleErrors), "console errors after fragment navigation: %s", strings.Join(consoleErrors, "; "))
}

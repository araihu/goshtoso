//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginationCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	var jsErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})
	page.On("pageerror", func(err error) {
		jsErrors = append(jsErrors, err.Error())
	})

	_, err := page.Goto(baseURL+"/components/pagination", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("main").GetByText("Pagination").First().WaitFor())
	require.NoError(t, page.Locator("#pagination-simple").ScrollIntoViewIfNeeded())

	next := page.Locator("#pagination-simple a[aria-label='next page']")
	nextHref, err := next.GetAttribute("href")
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(nextHref, "?page=4"), "simple next link should target page 4")
	require.NoError(t, next.Click())
	require.NoError(t, page.WaitForURL("**/components/pagination?page=4"))

	_, err = page.Goto(baseURL+"/components/pagination", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	beginEllipses, err := page.Locator("#pagination-ellipsis-begin span[aria-label='more pages']").Count()
	require.NoError(t, err)
	assert.Equal(t, 1, beginEllipses)
	require.NoError(t, page.Locator("#pagination-ellipsis-begin a[aria-current='page'][aria-label='page 2']").WaitFor())

	endEllipses, err := page.Locator("#pagination-ellipsis-end span[aria-label='more pages']").Count()
	require.NoError(t, err)
	assert.Equal(t, 1, endEllipses)
	require.NoError(t, page.Locator("#pagination-ellipsis-end a[aria-current='page'][aria-label='page 29']").WaitFor())

	htmxPage := page.Locator("#pagination-small a[aria-label='page 4']")
	hxGet, err := htmxPage.GetAttribute("hx-get")
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(hxGet, "?page=4"), "HTMX page link should target page 4")
	hxSwap, err := htmxPage.GetAttribute("hx-swap")
	require.NoError(t, err)
	assert.Equal(t, "innerHTML", hxSwap)

	require.Empty(t, jsErrors, "no JS console/page errors on pagination demo: %v", jsErrors)
}

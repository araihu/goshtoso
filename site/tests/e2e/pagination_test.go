package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPagination_FirstPageDisablesPreviousAndKeepsNextAvailable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/pagination", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	previousDisabled := page.Locator("#pagination-first [aria-disabled='true']").Filter(playwright.LocatorFilterOptions{HasText: "Previous"})
	require.NoError(t, previousDisabled.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))

	nextHref, err := page.Locator("#pagination-first a[aria-label='next page']").GetAttribute("href")
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(nextHref, "?page=2"), "next link should target page 2 from the first page")
}

func TestPagination_EllipsisMiddleShowsCurrentAndCollapsedRanges(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/pagination", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	current := page.Locator("#pagination-ellipsis-mid a[aria-current='page']")
	assert.Equal(t, "15", mustText(t, current))

	ellipsisCount, err := page.Locator("#pagination-ellipsis-mid span[aria-label='more pages']").Count()
	require.NoError(t, err)
	assert.Equal(t, 2, ellipsisCount, "middle pages should collapse both leading and trailing ranges")

	for _, label := range []string{"page 1", "page 14", "page 15", "page 16", "page 30"} {
		count, err := page.Locator("#pagination-ellipsis-mid a[aria-label='" + label + "']").Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "expected pagination link %s", label)
	}
}

func TestPagination_DocumentedHTMXVariantExposesHTMXAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/pagination", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	link := page.Locator("#pagination-small a[aria-label='page 4']")
	hxGet, err := link.GetAttribute("hx-get")
	require.NoError(t, err)
	assert.NotEmpty(t, hxGet, "small-page demo docs describe HTMX.Target + HTMX.Swap, so the rendered preview should expose hx-get")

	hxTarget, err := link.GetAttribute("hx-target")
	require.NoError(t, err)
	assert.Equal(t, "#items-tbody", hxTarget)
}

func TestPagination_DeepLinkKeepsTOCRailAttachedToContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 2048, Height: 1280},
	})

	_, err := page.Goto(baseURL+"/components/pagination#api-reference", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#api-reference").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, page.Locator(`#toc-list [data-toc-link="api-reference"]`).WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	gap, err := page.Evaluate(`() => {
		const content = document.getElementById('main-content');
		const rail = document.getElementById('toc-rail');
		if (!content || !rail) return Number.POSITIVE_INFINITY;
		const railStyle = getComputedStyle(rail);
		if (railStyle.display === 'none') return Number.POSITIVE_INFINITY;
		return rail.getBoundingClientRect().left - content.getBoundingClientRect().right;
	}`)
	require.NoError(t, err)
	assert.LessOrEqual(t, jsFloat(gap), 1.0, "desktop TOC rail should sit flush with the content frame")

	apiTop, err := page.Locator("#api-reference").Evaluate("el => el.getBoundingClientRect().top", nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, jsFloat(apiTop), 64.0, "deep-linked heading should clear the sticky header")
	assert.Less(t, jsFloat(apiTop), 1280.0, "deep-linked heading should remain visible without adding blank tail space")
}

func jsFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}

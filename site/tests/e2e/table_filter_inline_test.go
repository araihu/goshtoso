//go:build e2e && (full || table)

package e2e

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTableFilter_InlineVariant pins the `FilterVariantInline` contract:
// same reactive filter behavior, no bordered block, no collapsible toggle,
// no `x-show/x-collapse` wrapper. Designed for modal bodies and toolbar
// strips where the host container already owns chrome.
func TestTableFilter_InlineVariant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	page.SetDefaultTimeout(5000)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "Alpine.js should be available")

	_, err = page.WaitForFunction(`() => {
		var bar = document.getElementById('inline-filtered-table-filters');
		var el = bar && bar.closest('[data-table-filters]');
		if (!el) return false;
		try { return !!Alpine.$data(el); } catch(e) { return false; }
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "inline filter Alpine component should initialize")

	t.Run("NoBorderedWrapper", func(t *testing.T) {
		// The bar-variant wrapper carries the chrome classes; inline must not.
		bar := page.Locator("#inline-filtered-table-filters")
		className, err := bar.GetAttribute("class")
		require.NoError(t, err)
		assert.NotContains(t, className, "border", "inline variant must not render a bordered block")
		assert.NotContains(t, className, "rounded-radius", "inline variant must not render rounded-radius chrome")
		assert.Contains(t, className, "flex", "inline variant must render filters as a flex row")
	})

	t.Run("NoCollapsibleToggle", func(t *testing.T) {
		toggle := page.Locator("#inline-filtered-table-filters button:has-text('Filters')")
		count, err := toggle.Count()
		require.NoError(t, err)
		assert.Equal(t, 0, count, "inline variant must not render the collapsible header button")
	})

	t.Run("HeadersStartNonSortable", func(t *testing.T) {
		headers := page.Locator("#inline-filtered-table-thead th[hx-get*='order_by']")
		count, err := headers.Count()
		require.NoError(t, err)
		assert.Zero(t, count, "inline filter appearance must not gain an unrelated sorting contract")
	})

	t.Run("SearchStillSwapsTbody", func(t *testing.T) {
		// Inline variant should still fire HTMX on input, same as bar.
		input := page.Locator("#inline-filtered-table-filters input[type='search']")
		paginator := page.Locator("#inline-filtered-table-pagination")
		text, err := paginator.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Page 1 of 4")

		require.NoError(t, input.Fill("alice"))
		_, err = input.Evaluate(`(el) => el.dispatchEvent(new Event('input', {bubbles: true}))`, nil)
		require.NoError(t, err)

		_, err = page.WaitForFunction(
			`() => {
				var rows = document.querySelectorAll('#inline-filtered-table tbody tr');
				return rows.length === 1;
			}`, nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err, "inline filter input should still swap the tbody on input")

		text, err = page.Locator("#inline-filtered-table tbody").TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Alice Brown")
		hidden, err := paginator.IsHidden()
		require.NoError(t, err)
		require.True(t, hidden, "one-page filtered results must hide stale pagination controls")

		// Clear back to four pages and prove the stable host restores controls.
		require.NoError(t, input.Fill(""))
		_, err = input.Evaluate(`(el) => el.dispatchEvent(new Event('input', {bubbles: true}))`, nil)
		require.NoError(t, err)
		_, err = page.WaitForFunction(
			`() => document.querySelectorAll('#inline-filtered-table tbody tr').length === 3`,
			nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)
		_, err = page.WaitForFunction(
			`() => {
				const pagination = document.querySelector('#inline-filtered-table-pagination');
				return pagination && !pagination.hidden && pagination.textContent.includes('Page 1 of 4');
			}`,
			nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)
		activeText, err := paginator.Locator("a[aria-current='page']").TextContent()
		require.NoError(t, err)
		assert.Equal(t, "1", strings.TrimSpace(activeText))
		hxGet, err := paginator.Locator("a[aria-current='page']").GetAttribute("hx-get")
		require.NoError(t, err)
		assert.Contains(t, hxGet, "table_id=inline-filtered-table")
	})

	t.Run("PaginationKeepsTableAndPaginatorState", func(t *testing.T) {
		currentPage := page.Locator("#inline-filtered-table-pagination a[aria-current='page']")
		hxGet, err := currentPage.GetAttribute("hx-get")
		require.NoError(t, err)
		assert.Contains(t, hxGet, "table_id=inline-filtered-table")

		require.NoError(t, currentPage.Click())
		assert.Contains(t, page.URL(), "/components/table", "active page click must remain an HTMX swap")

		page2 := page.Locator("#inline-filtered-table-pagination a[aria-label='page 2']")
		require.NoError(t, page2.Click())
		_, err = page.WaitForFunction(
			`() => document.querySelector('#inline-filtered-table-pagination')?.textContent.includes('Page 2 of 4')`, nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err, "page 2 should update inline table paginator")

		activeText, err := page.Locator("#inline-filtered-table-pagination a[aria-current='page']").TextContent()
		require.NoError(t, err)
		assert.Equal(t, "2", strings.TrimSpace(activeText))

		next := page.Locator("#inline-filtered-table-pagination a[aria-label='next page']")
		require.NoError(t, next.Click())
		_, err = page.WaitForFunction(
			`() => document.querySelector('#inline-filtered-table-pagination')?.textContent.includes('Page 3 of 4')`, nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err, "Next should advance inline table paginator")

		headers := page.Locator("#inline-filtered-table-thead th[hx-get*='order_by']")
		count, err := headers.Count()
		require.NoError(t, err)
		assert.Zero(t, count, "pagination responses must preserve non-sortable inline headers")
	})
}

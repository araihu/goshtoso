//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fillSearchInput fills the search input and dispatches an input event so Alpine's x-model picks up the change.
// Playwright's Fill() doesn't trigger Alpine's x-model binding.
func fillSearchInput(t *testing.T, page playwright.Page, value string) {
	t.Helper()
	searchInput := page.Locator("#filtered-table-filters input[type='search']")
	err := searchInput.Fill(value)
	require.NoError(t, err)
	// Dispatch input event to trigger Alpine x-model
	_, err = searchInput.Evaluate(`(el) => el.dispatchEvent(new Event('input', {bubbles: true}))`, nil)
	require.NoError(t, err)
}

// ensureFiltersExpanded ensures the filter bar is expanded by checking Alpine state.
func ensureFiltersExpanded(t *testing.T, page playwright.Page) {
	t.Helper()
	_, _ = page.Evaluate(`() => {
		var bar = document.getElementById('filtered-table-filters');
		var el = bar && bar.closest('[data-table-filters]');
		if (el) { Alpine.$data(el).filtersExpanded = true; }
	}`, nil)
	// Wait for the search input to be visible
	err := page.Locator("#filtered-table-filters input[type='search']").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(2000),
	})
	require.NoError(t, err, "filters should be expanded")
}

func TestTableFilter(t *testing.T) {
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

	// Wait for Alpine.js to initialize (loaded locally, no CDN dependency)
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(3000),
	})
	require.NoError(t, err, "Alpine.js should be available (bundled locally)")

	// Wait for Alpine to process the filter component
	_, err = page.WaitForFunction(`() => {
		var bar = document.getElementById('filtered-table-filters');
		var el = bar && bar.closest('[data-table-filters]');
		if (!el) return false;
		try { return !!Alpine.$data(el); } catch(e) { return false; }
	}`, nil, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(3000),
	})
	require.NoError(t, err, "Alpine filter component should initialize")

	// --- Filter bar structure ---

	t.Run("FilterBar_Renders", func(t *testing.T) {
		filterBar := page.Locator("#filtered-table-filters")
		visible, err := filterBar.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "filter bar should be visible")
	})

	t.Run("FilterBar_HasSearchInput", func(t *testing.T) {
		searchInput := page.Locator("#filtered-table-filters input[type='search']")
		count, err := searchInput.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "should have 1 search input")

		placeholder, err := searchInput.GetAttribute("placeholder")
		require.NoError(t, err)
		assert.Contains(t, placeholder, "Search by name", "search should have placeholder")
	})

	t.Run("FilterBar_HasMembershipSelect", func(t *testing.T) {
		sel := page.Locator("#filtered-table-filters select")
		count, err := sel.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "should have 1 select")

		options := sel.Locator("option")
		optCount, err := options.Count()
		require.NoError(t, err)
		assert.Equal(t, 3, optCount, "select should have 3 options (All, Gold, Silver)")
	})

	t.Run("FilterBar_IsCollapsible", func(t *testing.T) {
		toggleBtn := page.Locator("#filtered-table-filters button:has-text('Filters')")
		count, err := toggleBtn.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "should have collapsible Filters toggle")

		searchInput := page.Locator("#filtered-table-filters input[type='search']")

		// Click to collapse
		err = toggleBtn.Click()
		require.NoError(t, err)

		// Wait for search input to become hidden (x-collapse animation)
		err = searchInput.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateHidden,
			Timeout: playwright.Float(2000),
		})
		require.NoError(t, err, "search should be hidden after collapse")

		// Click to expand again
		err = toggleBtn.Click()
		require.NoError(t, err)

		// Wait for search input to become visible
		err = searchInput.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(2000),
		})
		require.NoError(t, err, "search should be visible after expand")
	})

	// --- Search filter ---

	t.Run("Search_FiltersRows", func(t *testing.T) {
		ensureFiltersExpanded(t, page)

		fillSearchInput(t, page, "alice")

		// Wait for HTMX to swap and show filtered results
		_, err := page.WaitForFunction(
			`() => {
				var rows = document.querySelectorAll('#filtered-table tbody tr');
				return rows.length === 1;
			}`, nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err, "should filter to 1 row")

		tbodyText, err := page.Locator("#filtered-table tbody").TextContent()
		require.NoError(t, err)
		assert.Contains(t, tbodyText, "Alice Brown", "should find Alice")
	})

	t.Run("Search_ClearedShowsAll", func(t *testing.T) {
		ensureFiltersExpanded(t, page)

		fillSearchInput(t, page, "")

		// Wait for rows to return to default page size
		_, err := page.WaitForFunction(
			`() => {
				var rows = document.querySelectorAll('#filtered-table tbody tr');
				return rows.length === 3;
			}`, nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err, "clearing search should show default page (3 rows)")
	})

	// --- Select filter ---

	t.Run("SelectGold_FiltersToGoldOnly", func(t *testing.T) {
		ensureFiltersExpanded(t, page)

		sel := page.Locator("#filtered-table-filters select")
		_, err := sel.SelectOption(playwright.SelectOptionValues{Values: &[]string{"Gold"}})
		require.NoError(t, err)

		// Wait for HTMX swap
		_, err = page.WaitForFunction(
			`() => {
				var text = document.querySelector('#filtered-table tbody').textContent;
				return text.includes('Gold') && !text.includes('Silver');
			}`, nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err, "Gold filter should show only Gold rows")
	})

	t.Run("SelectAll_ClearsFilter", func(t *testing.T) {
		ensureFiltersExpanded(t, page)

		sel := page.Locator("#filtered-table-filters select")
		_, err := sel.SelectOption(playwright.SelectOptionValues{Values: &[]string{""}})
		require.NoError(t, err)

		_, err = page.WaitForFunction(
			`() => {
				var rows = document.querySelectorAll('#filtered-table tbody tr');
				return rows.length === 3;
			}`, nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err, "clearing filter should show default page")
	})

	// --- Combined search + select ---

	t.Run("CombinedFilter_SearchAndSelect", func(t *testing.T) {
		ensureFiltersExpanded(t, page)

		// Set Gold filter
		sel := page.Locator("#filtered-table-filters select")
		_, err := sel.SelectOption(playwright.SelectOptionValues{Values: &[]string{"Gold"}})
		require.NoError(t, err)

		_, err = page.WaitForFunction(
			`() => !document.querySelector('#filtered-table tbody').textContent.includes('Silver')`,
			nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)

		// Then search within Gold
		fillSearchInput(t, page, "bob")

		_, err = page.WaitForFunction(
			`() => document.querySelectorAll('#filtered-table tbody tr').length === 1`,
			nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err, "should find only Bob in Gold members")

		tbodyText, err := page.Locator("#filtered-table tbody").TextContent()
		require.NoError(t, err)
		assert.Contains(t, tbodyText, "Bob Johnson")

		// Clear both
		fillSearchInput(t, page, "")
		_, err = sel.SelectOption(playwright.SelectOptionValues{Values: &[]string{""}})
		require.NoError(t, err)

		_, err = page.WaitForFunction(
			`() => document.querySelectorAll('#filtered-table tbody tr').length === 3`,
			nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err, "should restore all rows")
	})

	// --- Filter preserves across sort ---

	t.Run("FilterPreservedAcrossSort", func(t *testing.T) {
		ensureFiltersExpanded(t, page)

		// Set Gold filter
		sel := page.Locator("#filtered-table-filters select")
		_, err := sel.SelectOption(playwright.SelectOptionValues{Values: &[]string{"Gold"}})
		require.NoError(t, err)

		_, err = page.WaitForFunction(
			`() => !document.querySelector('#filtered-table tbody').textContent.includes('Silver')`,
			nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)

		// Click a sort header - filter should be preserved
		sortHeader := page.Locator("#filtered-table thead th[hx-get*='order_by']").First()
		count, err := sortHeader.Count()
		require.NoError(t, err)
		if count > 0 {
			err = sortHeader.Click()
			require.NoError(t, err)

			// Wait for sort swap to complete, then verify Gold filter still applied
			_, err = page.WaitForFunction(
				`() => {
					var text = document.querySelector('#filtered-table tbody').textContent;
					return !text.includes('Silver');
				}`, nil,
				playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
			)
			require.NoError(t, err, "Gold filter should persist after sort")
		}

		// Clean up
		_, err = sel.SelectOption(playwright.SelectOptionValues{Values: &[]string{""}})
		require.NoError(t, err)
	})
}

//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitForPaginationPage(t *testing.T, page playwright.Page, pageNumber int) {
	t.Helper()
	_, err := page.WaitForFunction(`pageNumber => document.querySelector('#paginated-table-pagination')?.textContent.includes('Page ' + pageNumber + ' of 4')`, pageNumber)
	require.NoError(t, err)
}

// TestTablePaginationNav tests that pagination controls and table rows update
// correctly when navigating between pages via HTMX.
func TestTablePaginationNav(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	page.SetDefaultTimeout(3000)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	// --- Initial state (Page 1 of 4) ---

	t.Run("InitialState_Page1", func(t *testing.T) {
		// Page info text
		pageInfo, err := page.Locator("#paginated-table-pagination").TextContent()
		require.NoError(t, err)
		assert.Contains(t, pageInfo, "Page 1 of 4")

		// Active page button should be "1" with primary background
		activePage := page.Locator("#paginated-table-pagination a[aria-current='page']")
		count, err := activePage.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count, "should have exactly 1 active page")

		text, err := activePage.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "1", "active page should be 1")

		classAttr, err := activePage.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classAttr, "bg-primary", "active page should have bg-primary")
		assert.Contains(t, classAttr, "font-bold", "active page should be bold")
	})

	t.Run("InitialState_PrevDisabled", func(t *testing.T) {
		// Prev should be a <span> with aria-disabled, not an <a>
		prevDisabled := page.Locator("#paginated-table-pagination span[aria-disabled='true']:has-text('Previous')")
		count, err := prevDisabled.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Prev should be disabled on page 1")

		classAttr, err := prevDisabled.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classAttr, "cursor-not-allowed", "disabled Prev should have cursor-not-allowed")
		assert.Contains(t, classAttr, "text-on-surface/40", "disabled Prev should have muted text")
	})

	t.Run("InitialState_NextEnabled", func(t *testing.T) {
		nextLink := page.Locator("#paginated-table-pagination a[aria-label='next page']")
		count, err := nextLink.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Next should be an enabled link on page 1")
	})

	t.Run("InitialState_Rows", func(t *testing.T) {
		rows := page.Locator("#paginated-table tbody tr")
		count, err := rows.Count()
		require.NoError(t, err)
		assert.Equal(t, 3, count, "page 1 should have 3 rows")

		// First row should contain first record
		firstRow, err := rows.First().TextContent()
		require.NoError(t, err)
		assert.Contains(t, firstRow, "2335", "first row should have ID 2335")
		assert.Contains(t, firstRow, "Alice Brown")
	})

	// --- Navigate to Page 2 ---

	t.Run("NavigateToPage2", func(t *testing.T) {
		// Click page 2 link
		page2Link := page.Locator("#paginated-table-pagination a[aria-label='page 2']")
		count, err := page2Link.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count, "page 2 link should exist")

		err = page2Link.Click()
		require.NoError(t, err)
		waitForPaginationPage(t, page, 2)
	})

	t.Run("Page2_RowsChanged", func(t *testing.T) {
		rows := page.Locator("#paginated-table tbody tr")
		count, err := rows.Count()
		require.NoError(t, err)
		assert.Equal(t, 3, count, "page 2 should have 3 rows")

		// Page 2 should NOT have Alice (she's on page 1)
		tbodyText, err := page.Locator("#paginated-table tbody").TextContent()
		require.NoError(t, err)
		assert.NotContains(t, tbodyText, "Alice Brown", "page 2 should not contain page 1 records")
		assert.Contains(t, tbodyText, "2345", "page 2 should have record 4")
	})

	t.Run("Page2_PaginatorUpdated", func(t *testing.T) {
		// Page info should now show "Page 2 of 4"
		pageInfo, err := page.Locator("#paginated-table-pagination").TextContent()
		require.NoError(t, err)
		assert.Contains(t, pageInfo, "Page 2 of 4", "page info should update to page 2")

		// Active page should be "2"
		activePage := page.Locator("#paginated-table-pagination a[aria-current='page']")
		text, err := activePage.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "2", "active page should be 2")

		classAttr, err := activePage.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classAttr, "bg-primary", "page 2 button should have primary bg")
	})

	t.Run("Page2_PrevEnabled", func(t *testing.T) {
		// Prev should now be an <a> link, not disabled <span>
		prevLink := page.Locator("#paginated-table-pagination a[aria-label='previous page']")
		count, err := prevLink.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Prev should be enabled on page 2")

		// Disabled prev span should be gone
		prevDisabled := page.Locator("#paginated-table-pagination span[aria-disabled='true']:has-text('Previous')")
		disabledCount, err := prevDisabled.Count()
		require.NoError(t, err)
		assert.Equal(t, 0, disabledCount, "Prev should not be disabled on page 2")
	})

	t.Run("Page2_InactivePageStyles", func(t *testing.T) {
		// Page 1 button should now be inactive (no bg-primary, no font-bold)
		page1Link := page.Locator("#paginated-table-pagination a[aria-label='page 1']")
		count, err := page1Link.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count, "page 1 link should exist")

		classAttr, err := page1Link.GetAttribute("class")
		require.NoError(t, err)
		assert.NotContains(t, classAttr, "bg-primary", "page 1 should not have primary bg when inactive")
		assert.NotContains(t, classAttr, "font-bold", "page 1 should not be bold when inactive")
		assert.Contains(t, classAttr, "hover:text-primary", "inactive page should have hover effect")
	})

	// --- Navigate to Last Page ---

	t.Run("NavigateToPage4", func(t *testing.T) {
		page4Link := page.Locator("#paginated-table-pagination a[aria-label='page 4']")
		err := page4Link.Click()
		require.NoError(t, err)
		waitForPaginationPage(t, page, 4)
	})

	t.Run("Page4_RowsAreLastRecords", func(t *testing.T) {
		tbodyText, err := page.Locator("#paginated-table tbody").TextContent()
		require.NoError(t, err)
		assert.Contains(t, tbodyText, "Emma Harris", "last page should contain Emma Harris")
		assert.Contains(t, tbodyText, "2355")
	})

	t.Run("Page4_NextDisabled", func(t *testing.T) {
		// Next should be disabled on last page
		nextDisabled := page.Locator("#paginated-table-pagination span[aria-disabled='true']:has-text('Next')")
		count, err := nextDisabled.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Next should be disabled on last page")

		classAttr, err := nextDisabled.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classAttr, "cursor-not-allowed")
	})

	t.Run("Page4_PaginatorShowsPage4", func(t *testing.T) {
		pageInfo, err := page.Locator("#paginated-table-pagination").TextContent()
		require.NoError(t, err)
		assert.Contains(t, pageInfo, "Page 4 of 4")

		activePage := page.Locator("#paginated-table-pagination a[aria-current='page']")
		text, err := activePage.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "4", "active page should be 4")
	})

	// --- Navigate back to Page 1 via Prev ---

	t.Run("NavigateBackToPage1", func(t *testing.T) {
		// Click Previous repeatedly to go back to page 1
		prevLink := page.Locator("#paginated-table-pagination a[aria-label='previous page']")

		// Page 4 -> 3
		err := prevLink.Click()
		require.NoError(t, err)
		waitForPaginationPage(t, page, 3)

		// Page 3 -> 2
		prevLink = page.Locator("#paginated-table-pagination a[aria-label='previous page']")
		err = prevLink.Click()
		require.NoError(t, err)
		waitForPaginationPage(t, page, 2)

		// Page 2 -> 1
		prevLink = page.Locator("#paginated-table-pagination a[aria-label='previous page']")
		err = prevLink.Click()
		require.NoError(t, err)
		waitForPaginationPage(t, page, 1)
	})

	t.Run("BackOnPage1_StateRestored", func(t *testing.T) {
		// Should be back to page 1
		pageInfo, err := page.Locator("#paginated-table-pagination").TextContent()
		require.NoError(t, err)
		assert.Contains(t, pageInfo, "Page 1 of 4")

		// Alice should be back
		tbodyText, err := page.Locator("#paginated-table tbody").TextContent()
		require.NoError(t, err)
		assert.Contains(t, tbodyText, "Alice Brown")

		// Prev should be disabled again
		prevDisabled := page.Locator("#paginated-table-pagination span[aria-disabled='true']:has-text('Previous')")
		count, err := prevDisabled.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Prev should be disabled again on page 1")
	})

	// --- Navigate via Next button ---

	t.Run("NavigateViaNext", func(t *testing.T) {
		nextLink := page.Locator("#paginated-table-pagination a[aria-label='next page']")
		err := nextLink.Click()
		require.NoError(t, err)
		waitForPaginationPage(t, page, 2)

		pageInfo, err := page.Locator("#paginated-table-pagination").TextContent()
		require.NoError(t, err)
		assert.Contains(t, pageInfo, "Page 2 of 4", "Next should advance to page 2")
	})
}

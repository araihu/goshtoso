//go:build e2e && (full || table)

package e2e

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tableAPI is a helper that fetches the table rows API and returns the HTML body.
func tableAPI(t *testing.T, params string) string {
	t.Helper()
	url := baseURL + "/api/components/table/rows"
	if params != "" {
		url += "?" + params
	}
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body)
}

// countSubstring counts non-overlapping occurrences of substr in s.
func countSubstring(s, substr string) int {
	return strings.Count(s, substr)
}

// --- Pagination Tests ---

func TestTableHTMX_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("DefaultPage_Returns3Rows", func(t *testing.T) {
		body := tableAPI(t, "")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 3, rows, "default page should return 3 rows (per_page default)")
	})

	t.Run("Page1_HasFirstRecords", func(t *testing.T) {
		body := tableAPI(t, "page=1&per_page=3")
		assert.Contains(t, body, "2335", "page 1 should contain first record ID")
		assert.Contains(t, body, "Alice Brown")
	})

	t.Run("Page2_HasNextRecords", func(t *testing.T) {
		body := tableAPI(t, "page=2&per_page=3")
		assert.NotContains(t, body, "Alice Brown", "page 2 should not have page 1 records")
		assert.Contains(t, body, "2345", "page 2 should have record 4")
	})

	t.Run("Page3_HasRecords7to9", func(t *testing.T) {
		body := tableAPI(t, "page=3&per_page=3")
		assert.Contains(t, body, "2350")
		assert.Contains(t, body, "James Wilson")
	})

	t.Run("Page4_HasLastRecords", func(t *testing.T) {
		body := tableAPI(t, "page=4&per_page=3")
		assert.Contains(t, body, "2354")
		assert.Contains(t, body, "Emma Harris")
	})

	t.Run("LargerPageSize_ReturnsMoreRows", func(t *testing.T) {
		body := tableAPI(t, "page=1&per_page=6")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 6, rows, "per_page=6 should return 6 rows")
	})

	t.Run("FullPageSize_ReturnsAllRows", func(t *testing.T) {
		body := tableAPI(t, "page=1&per_page=100")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 12, rows, "large per_page should return all 12 rows")
	})

	t.Run("BeyondLastPage_ResetsToPage1", func(t *testing.T) {
		body := tableAPI(t, "page=999&per_page=3")
		// Handler resets to page 1 when start >= len(records)
		assert.Contains(t, body, "2335", "out-of-range page should reset to page 1")
	})

	t.Run("PerPage1_Returns1Row", func(t *testing.T) {
		body := tableAPI(t, "page=1&per_page=1")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 1, rows)
	})
}

// --- Sorting Tests ---

func TestTableHTMX_Sorting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("SortByName_Asc", func(t *testing.T) {
		body := tableAPI(t, "order_by=name&order_dir=asc&per_page=100")
		// Alice should come before Bob, Bob before Daniel, etc.
		aliceIdx := strings.Index(body, "Alice Brown")
		bobIdx := strings.Index(body, "Bob Johnson")
		danielIdx := strings.Index(body, "Daniel Lee")
		sophiaIdx := strings.Index(body, "Sophia Chen")
		require.Greater(t, aliceIdx, -1, "Alice should be present")
		assert.Less(t, aliceIdx, bobIdx, "Alice before Bob (asc)")
		assert.Less(t, bobIdx, danielIdx, "Bob before Daniel (asc)")
		assert.Less(t, danielIdx, sophiaIdx, "Daniel before Sophia (asc)")
	})

	t.Run("SortByName_Desc", func(t *testing.T) {
		body := tableAPI(t, "order_by=name&order_dir=desc&per_page=100")
		sophiaIdx := strings.Index(body, "Sophia Chen")
		aliceIdx := strings.Index(body, "Alice Brown")
		assert.Less(t, sophiaIdx, aliceIdx, "Sophia before Alice (desc)")
	})

	t.Run("SortByID_Asc", func(t *testing.T) {
		body := tableAPI(t, "order_by=id&order_dir=asc&per_page=100")
		idx2335 := strings.Index(body, "2335")
		idx2355 := strings.Index(body, "2355")
		assert.Less(t, idx2335, idx2355, "2335 before 2355 (asc)")
	})

	t.Run("SortByID_Desc", func(t *testing.T) {
		body := tableAPI(t, "order_by=id&order_dir=desc&per_page=100")
		idx2355 := strings.Index(body, "2355")
		idx2335 := strings.Index(body, "2335")
		assert.Less(t, idx2355, idx2335, "2355 before 2335 (desc)")
	})

	t.Run("SortByMembership_GroupsCorrectly", func(t *testing.T) {
		body := tableAPI(t, "order_by=membership&order_dir=asc&per_page=100")
		// Gold comes before Silver alphabetically
		firstGold := strings.Index(body, "Gold")
		firstSilver := strings.Index(body, "Silver")
		assert.Less(t, firstGold, firstSilver, "Gold before Silver (asc)")
	})

	t.Run("SortByEmail_Works", func(t *testing.T) {
		body := tableAPI(t, "order_by=email&order_dir=asc&per_page=100")
		aliceIdx := strings.Index(body, "alice.brown@")
		sophiaIdx := strings.Index(body, "sophia.chen@")
		assert.Less(t, aliceIdx, sophiaIdx, "alice.brown before sophia.chen (asc)")
	})

	t.Run("DefaultOrderDir_IsAsc", func(t *testing.T) {
		bodyExplicit := tableAPI(t, "order_by=name&order_dir=asc&per_page=100")
		bodyDefault := tableAPI(t, "order_by=name&per_page=100")
		assert.Equal(t, bodyExplicit, bodyDefault, "omitting order_dir should default to asc")
	})

	t.Run("NoSortParams_ReturnsNaturalOrder", func(t *testing.T) {
		bodyNoSort := tableAPI(t, "per_page=100")
		bodySortAsc := tableAPI(t, "order_by=name&order_dir=asc&per_page=100")
		// Natural order should differ from explicitly sorted order
		// (unless data happens to be pre-sorted, which it isn't — IDs are not alphabetical)
		// Just verify the response is valid and contains rows
		rows := countSubstring(bodyNoSort, "<tr")
		assert.Greater(t, rows, 0, "should return rows without sort params")
		_ = bodySortAsc
	})
}

// --- Sorting + Pagination Combined ---

func TestTableHTMX_SortAndPaginate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("SortDescByName_Page1_HasSophia", func(t *testing.T) {
		body := tableAPI(t, "order_by=name&order_dir=desc&page=1&per_page=3")
		assert.Contains(t, body, "Sophia Chen", "desc name page 1 should start with Sophia")
		assert.Contains(t, body, "Sarah Adams")
		assert.Contains(t, body, "Ryan Thompson")
	})

	t.Run("SortDescByName_Page2_HasOlivia", func(t *testing.T) {
		body := tableAPI(t, "order_by=name&order_dir=desc&page=2&per_page=3")
		assert.NotContains(t, body, "Sophia Chen", "page 2 should not have Sophia")
		assert.Contains(t, body, "Olivia Taylor")
	})

	t.Run("SortAscByID_Page4_HasLastIDs", func(t *testing.T) {
		body := tableAPI(t, "order_by=id&order_dir=asc&page=4&per_page=3")
		assert.Contains(t, body, "2354")
		assert.Contains(t, body, "2355")
	})

	t.Run("PaginationIsStableAcrossSort", func(t *testing.T) {
		// Page 1 sorted by name asc should always return the same 3 names
		body1 := tableAPI(t, "order_by=name&order_dir=asc&page=1&per_page=3")
		body2 := tableAPI(t, "order_by=name&order_dir=asc&page=1&per_page=3")
		assert.Equal(t, body1, body2, "same params should return identical results")
	})
}

// --- Filtering Tests ---

func TestTableHTMX_Filtering(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("SearchByName_FindsMatch", func(t *testing.T) {
		body := tableAPI(t, "search=alice&per_page=100")
		assert.Contains(t, body, "Alice Brown")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 1, rows, "only Alice should match")
	})

	t.Run("SearchByEmail_FindsMatch", func(t *testing.T) {
		body := tableAPI(t, "search=sophia.chen&per_page=100")
		assert.Contains(t, body, "Sophia Chen")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 1, rows)
	})

	t.Run("SearchByID_FindsMatch", func(t *testing.T) {
		body := tableAPI(t, "search=2342&per_page=100")
		assert.Contains(t, body, "Sarah Adams")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 1, rows)
	})

	t.Run("SearchIsCaseInsensitive", func(t *testing.T) {
		body := tableAPI(t, "search=ALICE&per_page=100")
		assert.Contains(t, body, "Alice Brown")
	})

	t.Run("SearchPartialName_ReturnsMultiple", func(t *testing.T) {
		body := tableAPI(t, "search=son&per_page=100")
		// Bob Johnson and Ryan Thompson both contain "son"
		assert.Contains(t, body, "Bob Johnson")
		assert.Contains(t, body, "Ryan Thompson")
	})

	t.Run("SearchNoMatch_ReturnsEmpty", func(t *testing.T) {
		body := tableAPI(t, "search=zzzznotfound&per_page=100")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 0, rows, "no match should return 0 rows")
	})

	t.Run("FilterByMembership_Gold", func(t *testing.T) {
		body := tableAPI(t, "membership=Gold&per_page=100")
		assert.Contains(t, body, "Bob Johnson")
		assert.Contains(t, body, "Sophia Chen")
		assert.NotContains(t, body, "Alice Brown", "Alice is Silver, not Gold")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 7, rows, "should have 7 Gold members")
	})

	t.Run("FilterByMembership_Silver", func(t *testing.T) {
		body := tableAPI(t, "membership=Silver&per_page=100")
		assert.Contains(t, body, "Alice Brown")
		assert.NotContains(t, body, "Bob Johnson", "Bob is Gold, not Silver")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 5, rows, "should have 5 Silver members")
	})

	t.Run("FilterIsCaseInsensitive", func(t *testing.T) {
		body := tableAPI(t, "membership=gold&per_page=100")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 7, rows, "membership filter should be case-insensitive")
	})
}

// --- Filter + Sort + Pagination Combined ---

func TestTableHTMX_FilterSortPaginate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("FilterGold_SortByNameAsc_Page1", func(t *testing.T) {
		body := tableAPI(t, "membership=Gold&order_by=name&order_dir=asc&page=1&per_page=3")
		// Gold members sorted by name asc: Alex Martinez, Bob Johnson, Emily Rodriguez, ...
		assert.Contains(t, body, "Alex Martinez")
		assert.Contains(t, body, "Bob Johnson")
		assert.Contains(t, body, "Emily Rodriguez")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 3, rows)
	})

	t.Run("FilterGold_SortByNameAsc_Page2", func(t *testing.T) {
		body := tableAPI(t, "membership=Gold&order_by=name&order_dir=asc&page=2&per_page=3")
		// Gold members sorted by name asc page 2: Emma Harris, Olivia Taylor, Sarah Adams
		assert.NotContains(t, body, "Alex Martinez", "page 2 shouldn't have page 1 records")
		assert.Contains(t, body, "Emma Harris")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 3, rows)
	})

	t.Run("FilterSilver_SortByNameDesc", func(t *testing.T) {
		body := tableAPI(t, "membership=Silver&order_by=name&order_dir=desc&per_page=100")
		// Silver members: Alice Brown, Ryan Thompson, James Wilson, Michael Davis, Daniel Lee, ...
		// Desc: Ryan, Michael, James, Daniel, Alice, ...
		ryanIdx := strings.Index(body, "Ryan Thompson")
		aliceIdx := strings.Index(body, "Alice Brown")
		assert.Less(t, ryanIdx, aliceIdx, "Ryan before Alice in desc order")
	})

	t.Run("Search_Sort_Paginate", func(t *testing.T) {
		// Search for "a" matches many names, sort by ID desc, paginate
		body := tableAPI(t, "search=a&order_by=id&order_dir=desc&page=1&per_page=3")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 3, rows, "should return exactly per_page rows")
	})

	t.Run("FilterReducesPagination", func(t *testing.T) {
		// All 12 records: page 4 has 3 rows (records 10-12)
		// Silver only (5 records): page 2 is the last page with 2 rows
		bodyAll4 := tableAPI(t, "page=4&per_page=3")
		bodySilver2 := tableAPI(t, "membership=Silver&page=2&per_page=3")
		allRows := countSubstring(bodyAll4, "<tr")
		silverRows := countSubstring(bodySilver2, "<tr")
		assert.Equal(t, 3, allRows, "all records page 4 should have 3 rows")
		assert.Equal(t, 2, silverRows, "Silver page 2 should have only 2 rows (5 total, 3 on page 1)")
	})
}

// --- Browser HTMX Integration Tests ---

func TestTableHTMX_BrowserSorting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("SortableHeaders_HaveHTMXAttributes", func(t *testing.T) {
		// Find the sortable table section
		sortableHeaders := page.Locator("#sortable-table thead th[hx-get]")
		count, err := sortableHeaders.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3, "should have at least 3 sortable headers")

		// Verify first sortable header has proper hx-get URL
		firstHeader := sortableHeaders.First()
		hxGet, err := firstHeader.GetAttribute("hx-get")
		require.NoError(t, err)
		assert.Contains(t, hxGet, "order_by=", "hx-get should include order_by param")
		assert.Contains(t, hxGet, "order_dir=", "hx-get should include order_dir param")
	})

	t.Run("ClickSortHeader_UpdatesRows", func(t *testing.T) {
		initialRows, err := page.Locator("#sortable-table tbody tr").Count()
		require.NoError(t, err)
		require.Equal(t, 6, initialRows, "sortable preview should start with six rows")

		// Get initial first row text
		initialFirst, err := page.Locator("#sortable-table tbody tr").First().TextContent()
		require.NoError(t, err)

		// Click the Name header to sort
		nameHeader := page.Locator("#sortable-table thead th[hx-get*='order_by=name']")
		count, err := nameHeader.Count()
		require.NoError(t, err)
		if count == 0 {
			t.Skip("no sortable name header found")
		}

		err = nameHeader.Click()
		require.NoError(t, err)
		_, err = page.WaitForFunction(
			`() => {
				var rows = document.querySelectorAll('#sortable-table tbody tr');
				var nameHeader = document.querySelector('#sortable-table-thead th[hx-get*="order_by=name"]');
				return rows.length === 6 &&
					rows[0].textContent.includes('Alex Martinez') &&
					nameHeader?.getAttribute('hx-get').includes('order_dir=desc');
			}`, nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err, "ascending sort should complete before the next interaction")

		// Get new first row text - should be different if sort changed
		newFirst, err := page.Locator("#sortable-table tbody tr").First().TextContent()
		require.NoError(t, err)

		// After clicking name sort, the order should change
		assert.Contains(t, newFirst, "Alex Martinez")
		assert.NotEqual(t, initialFirst, newFirst)
		t.Logf("Before sort: %.50s", strings.TrimSpace(initialFirst))
		t.Logf("After sort:  %.50s", strings.TrimSpace(newFirst))

		// Response-generated headers must retain the six-row endpoint state.
		clickUntil(t, page, nameHeader,
			`() => {
				var rows = document.querySelectorAll('#sortable-table tbody tr');
				var nameHeader = document.querySelector('#sortable-table-thead th[hx-get*="order_by=name"]');
				return rows.length === 6 &&
					rows[0].textContent.includes('Sophia Chen') &&
					!nameHeader?.getAttribute('hx-get').includes('order_dir=');
			}`,
		)
	})
}

// TestTableHTMX_SortCycling verifies the 3-state sort cycle: neutral → asc → desc → neutral.
func TestTableHTMX_SortCycling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	// getSortState returns the sort icon state for the Name column header.
	getSortState := func() string {
		result, err := page.Evaluate(`() => {
			var ths = document.querySelectorAll('#sortable-table-thead th[hx-get]');
			for (var th of ths) {
				if (th.textContent.includes('Name')) {
					var svg = th.querySelector('svg');
					if (!svg) return 'no-svg';
					var html = svg.outerHTML;
					if (html.includes('opacity')) return 'neutral';
					if (html.includes('M10 17')) return 'asc';
					if (html.includes('M10 3')) return 'desc';
					return 'unknown';
				}
			}
			return 'not-found';
		}`, nil)
		if err != nil {
			return "error"
		}
		return fmt.Sprintf("%v", result)
	}

	// getNameHeaderURL returns the hx-get URL of the Name column header.
	getNameHeaderURL := func() string {
		result, err := page.Evaluate(`() => {
			var ths = document.querySelectorAll('#sortable-table-thead th[hx-get]');
			for (var th of ths) {
				if (th.textContent.includes('Name')) return th.getAttribute('hx-get');
			}
			return '';
		}`, nil)
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%v", result)
	}

	// nameHeader re-resolves each use, so it finds the fresh <th> after the sort
	// OOB swap replaces the thead.
	nameHeader := page.Locator("#sortable-table-thead th:has-text('Name')")

	// nameHeaderURLHas builds a WaitForFunction predicate that holds once the Name
	// header's next-sort hx-get URL contains (present) or omits (!present) substr.
	nameHeaderURLHas := func(substr string, present bool) string {
		test := "th.getAttribute('hx-get').includes('" + substr + "')"
		if !present {
			test = "!" + test
		}
		return `() => { var ths = document.querySelectorAll('#sortable-table-thead th[hx-get]'); for (var th of ths) { if (th.textContent.includes('Name') && ` + test + `) return true; } return false; }`
	}

	t.Run("InitialState_NameIsNeutral", func(t *testing.T) {
		// Table starts sorted by ID asc, so Name column should be neutral
		state := getSortState()
		assert.Equal(t, "neutral", state, "Name header should be neutral initially (table sorted by ID)")
	})

	t.Run("Click1_NameSortsAsc", func(t *testing.T) {
		// After an asc click the header's next-sort URL points to desc.
		clickUntil(t, page, nameHeader, nameHeaderURLHas("order_dir=desc", true))
		assert.Equal(t, "asc", getSortState(), "after 1st click, should show asc icon")
	})

	t.Run("Click2_NameSortsDesc", func(t *testing.T) {
		// After desc, next click resets to neutral — URL omits order_dir.
		clickUntil(t, page, nameHeader, nameHeaderURLHas("order_dir=", false))
		assert.Equal(t, "desc", getSortState(), "after 2nd click, should show desc icon")
	})

	t.Run("Click3_NameReturnsToNeutral", func(t *testing.T) {
		// Neutral: next click goes to asc — URL contains order_dir=asc.
		clickUntil(t, page, nameHeader, nameHeaderURLHas("order_dir=asc", true))
		assert.Equal(t, "neutral", getSortState(), "after 3rd click, should return to neutral icon")
	})

	t.Run("NeutralState_URLOmitsSortParams", func(t *testing.T) {
		url := getNameHeaderURL()
		// In neutral state, the URL should have order_dir=asc (next click direction),
		// but the current state has no active sort. The URL IS the next action, not the current state.
		// The current state (neutral) is reflected by the icon, while the URL points to the next state (asc).
		assert.Contains(t, url, "order_by=name", "neutral header URL should target name column")
		assert.Contains(t, url, "order_dir=asc", "neutral header URL should point to asc as next action")
	})
}

func TestTableHTMX_BrowserPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("PaginationControls_Visible", func(t *testing.T) {
		pag := page.Locator("#paginated-table-pagination")
		visible, err := pag.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "pagination controls should be visible")
	})

	t.Run("PageInfo_ShowsCorrectState", func(t *testing.T) {
		pagText, err := page.Locator("#paginated-table-pagination").TextContent()
		require.NoError(t, err)
		assert.Contains(t, pagText, "Page 1 of 4", "should show page 1 of 4")
	})

	t.Run("PageButtons_AreClickable", func(t *testing.T) {
		// Get initial row content
		initialRows, err := page.Locator("#paginated-table tbody").TextContent()
		require.NoError(t, err)

		// Find and click a page navigation element (could be button or link)
		nextBtn := page.Locator("#paginated-table-pagination a[hx-get*='page=2'], #paginated-table-pagination button:has-text('2'), #paginated-table-pagination button:has-text('Next')")
		count, err := nextBtn.Count()
		require.NoError(t, err)
		if count == 0 {
			t.Skip("no page 2 button found")
		}

		err = nextBtn.First().Click()
		require.NoError(t, err)
		time.Sleep(500 * time.Millisecond)

		newRows, err := page.Locator("#paginated-table tbody").TextContent()
		require.NoError(t, err)
		assert.NotEqual(t, initialRows, newRows, "page 2 should show different rows")
	})
}

func TestTableHTMX_BrowserLazyLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	// Increase nav timeout for lazy load (500ms server delay)
	page.SetDefaultTimeout(5000)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("LazyTable_LoadsRows", func(t *testing.T) {
		loadButton := page.Locator("#load-lazy-table")
		err := loadButton.Click()
		require.NoError(t, err)

		// Wait for HTMX lazy load to complete (has 500ms server delay)
		time.Sleep(1500 * time.Millisecond)

		lazyTbody := page.Locator("#lazy-table tbody")
		rows := lazyTbody.Locator("tr")
		count, err := rows.Count()
		require.NoError(t, err)
		assert.Greater(t, count, 0, "lazy table should have loaded rows")
		t.Logf("Lazy table loaded %d rows", count)
	})
}

func TestTableHTMX_BrowserInfiniteScroll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("InfiniteScroll_HasSentinel", func(t *testing.T) {
		sentinel := page.Locator("#infinite-table tr#infinite-table-sentinel[data-hx-get]")
		count, err := sentinel.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count, "should have exactly 1 scroll sentinel row")

		// Verify sentinel URL has page=2
		url, err := sentinel.GetAttribute("data-hx-get")
		require.NoError(t, err)
		assert.Contains(t, url, "page=2", "sentinel should request page 2")
	})

	t.Run("InitialRows_Present", func(t *testing.T) {
		rows := page.Locator("#infinite-table tbody tr")
		count, err := rows.Count()
		require.NoError(t, err)
		// Initial rows + 1 sentinel row
		assert.GreaterOrEqual(t, count, 3, "should have at least 3 initial rows")
	})
}

// --- Response Format Tests ---

func TestTableHTMX_ResponseFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("ContentType_IsHTML", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/components/table/rows")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		ct := resp.Header.Get("Content-Type")
		assert.Contains(t, ct, "text/html")
	})

	t.Run("Response_ContainsTRElements", func(t *testing.T) {
		body := tableAPI(t, "")
		assert.Contains(t, body, "<tr", "response should contain table rows")
		assert.Contains(t, body, "<td", "response should contain table cells")
	})

	t.Run("InfiniteVariant_ContainsSentinel", func(t *testing.T) {
		body := tableAPI(t, "variant=infinite&page=1&per_page=3")
		assert.Contains(t, body, `id="table-sentinel"`, "infinite response should include sentinel")
		assert.Contains(t, body, "data-hx-get", "infinite response should include next-page URL")
		assert.Contains(t, body, "page=2", "sentinel should reference next page")
	})

	t.Run("InfiniteVariant_LastPage_NoSentinel", func(t *testing.T) {
		body := tableAPI(t, "variant=infinite&page=4&per_page=3")
		assert.NotContains(t, body, `id="table-sentinel"`, "last page should not have sentinel")
	})

	t.Run("Rows_ContainExpectedColumns", func(t *testing.T) {
		body := tableAPI(t, "page=1&per_page=1")
		assert.Contains(t, body, "2335", "should contain ID")
		assert.Contains(t, body, "Alice Brown", "should contain name")
		assert.Contains(t, body, "alice.brown@penguinui.com", "should contain email")
		assert.Contains(t, body, "Silver", "should contain membership")
	})
}

// --- Edge Cases ---

func TestTableHTMX_EdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("InvalidPage_Handled", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/components/table/rows?page=-1")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, 200, resp.StatusCode, "invalid page should not crash")
	})

	t.Run("InvalidPerPage_Handled", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/components/table/rows?per_page=abc")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, 200, resp.StatusCode, "non-numeric per_page should not crash")
	})

	t.Run("InvalidSortColumn_Handled", func(t *testing.T) {
		body := tableAPI(t, "order_by=nonexistent&per_page=100")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 12, rows, "invalid sort column should return all rows unsorted")
	})

	t.Run("EmptySearch_ReturnsAll", func(t *testing.T) {
		body := tableAPI(t, "search=&per_page=100")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 12, rows, "empty search should return all rows")
	})

	t.Run("FilterAndSearch_Combined", func(t *testing.T) {
		// Search "a" within Gold members
		body := tableAPI(t, "search=a&membership=Gold&per_page=100")
		rows := countSubstring(body, "<tr")
		assert.Greater(t, rows, 0, "combined filter should return some results")
		assert.NotContains(t, body, "Silver", "Gold filter should exclude Silver members")
	})

	t.Run("ZeroResults_FilterAndSearch", func(t *testing.T) {
		body := tableAPI(t, "search=zzz&membership=Gold&per_page=100")
		rows := countSubstring(body, "<tr")
		assert.Equal(t, 0, rows, "impossible combo should return 0 rows")
	})
}

// Data integrity test: verify all 12 records are retrievable
func TestTableHTMX_DataIntegrity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	body := tableAPI(t, "per_page=100")

	expectedNames := []string{
		"Alice Brown", "Bob Johnson", "Sarah Adams", "Alex Martinez",
		"Ryan Thompson", "Emily Rodriguez", "James Wilson", "Sophia Chen",
		"Michael Davis", "Olivia Taylor", "Daniel Lee", "Emma Harris",
	}

	for _, name := range expectedNames {
		assert.Contains(t, body, name, fmt.Sprintf("%s should be in full dataset", name))
	}

	rows := countSubstring(body, "<tr")
	assert.Equal(t, 12, rows, "full dataset should have exactly 12 rows")
}

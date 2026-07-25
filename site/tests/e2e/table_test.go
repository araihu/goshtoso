package e2e

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTable_DefaultTable tests the default table variant
func TestTable_DefaultTable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Page_Loads_Successfully", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		assert.Contains(t, title, "Table")
		t.Log("✓ Table page loads successfully")
	})

	t.Run("Default_Table_Renders", func(t *testing.T) {
		tables := page.Locator("table")
		count, err := tables.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "should have at least one table")

		firstTable := tables.First()
		classAttr, err := firstTable.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classAttr, "w-full")
		assert.Contains(t, classAttr, "text-left")
		assert.Contains(t, classAttr, "text-sm")

		t.Log("✓ Default table renders with correct classes")
	})

	t.Run("Table_Has_Headers", func(t *testing.T) {
		headers := page.Locator("table").First().Locator("thead th")
		count, err := headers.Count()
		require.NoError(t, err)
		assert.Equal(t, 4, count, "default table should have 4 headers")

		firstHeader, err := headers.Nth(0).TextContent()
		require.NoError(t, err)
		assert.Equal(t, "CustomerID", firstHeader)

		t.Log("✓ Table headers render correctly")
	})

	t.Run("Table_Has_Rows", func(t *testing.T) {
		rows := page.Locator("table").First().Locator("tbody tr")
		count, err := rows.Count()
		require.NoError(t, err)
		assert.Equal(t, 3, count, "default table should have 3 rows")

		firstCell, err := rows.Nth(0).Locator("td").First().TextContent()
		require.NoError(t, err)
		assert.Equal(t, "2335", firstCell)

		t.Log("✓ Table rows render correctly")
	})

	t.Run("Table_Container_Has_Border", func(t *testing.T) {
		container := page.Locator("table").First().Locator("xpath=..")
		classAttr, err := container.GetAttribute("class")
		require.NoError(t, err)

		outerContainer := container.Locator("xpath=..")
		outerClass, err := outerContainer.GetAttribute("class")
		require.NoError(t, err)

		combined := classAttr + " " + outerClass
		assert.Contains(t, combined, "border", "container should have border")
		assert.Contains(t, combined, "rounded-radius", "container should have rounded corners")

		t.Log("✓ Table container has correct styling")
	})
}

// TestTableStripedAppearance tests the striped table appearance.
func TestTableStripedAppearance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("StripedRowsHaveAppearanceClasses", func(t *testing.T) {
		// Scope to the per-variant preview wrapper (#table-striped) so the
		// codeblock's syntax-highlight spans never leak into the table query.
		stripedSection := page.Locator("#table-striped")

		rows := stripedSection.Locator("table tbody tr")
		count, err := rows.Count()
		require.NoError(t, err)
		assert.Equal(t, 6, count, "striped table should have 6 rows")

		firstRow := rows.First()
		classAttr, err := firstRow.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classAttr, "odd:bg-surface-alt", "striped rows should have the striped appearance class")

		t.Log("✓ Striped table rows have correct classes")
	})
}

// TestTableCheckboxSelection tests the checkbox selection behavior.
func TestTableCheckboxSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Checkbox_Table_Has_Checkboxes", func(t *testing.T) {
		checkboxSection := page.Locator("#table-checkbox")

		checkboxes := checkboxSection.Locator("input[type='checkbox']")
		count, err := checkboxes.Count()
		require.NoError(t, err)
		assert.Equal(t, 4, count, "checkbox table should have 4 checkboxes (1 header + 3 rows)")

		t.Log("✓ Checkbox table has correct number of checkboxes")
	})

	t.Run("Check_All_Selects_All_Rows", func(t *testing.T) {
		checkboxSection := page.Locator("#table-checkbox")

		headerCheckbox := checkboxSection.Locator("thead input[type='checkbox']")
		err := headerCheckbox.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		rowCheckboxes := checkboxSection.Locator("tbody input[type='checkbox']")
		rowCount, err := rowCheckboxes.Count()
		require.NoError(t, err)

		for i := range rowCount {
			checked, err := rowCheckboxes.Nth(i).IsChecked()
			require.NoError(t, err)
			assert.True(t, checked, "row checkbox %d should be checked", i)
		}

		t.Log("✓ Check all selects all row checkboxes")
	})
}

// TestTable_WithAction tests the action button table variant
func TestTable_WithAction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Action_Table_Has_Edit_Buttons", func(t *testing.T) {
		actionSection := page.Locator("#table-action")

		editButtons := actionSection.Locator("button:has-text('Edit')")
		count, err := editButtons.Count()
		require.NoError(t, err)
		assert.Equal(t, 3, count, "action table should have 3 Edit buttons")

		firstButton := editButtons.First()
		classAttr, err := firstButton.GetAttribute("class")
		require.NoError(t, err)
		for _, className := range []string{
			"!bg-transparent",
			"!border-transparent",
			"!text-primary",
			"dark:!text-primary-dark",
			"!p-0.5",
			"!rounded-radius",
		} {
			assert.Contains(t, classAttr, className, "Edit button should preserve text-action styling")
		}
		assert.Equal(t, "button", mustAttribute(t, firstButton, "type"))
		assert.Equal(t, "Edit", strings.TrimSpace(mustText(t, firstButton)))

		appearanceValue, err := firstButton.Evaluate(`el => {
			const style = getComputedStyle(el);
			const probe = document.createElement("span");
			probe.style.color = "var(--color-primary)";
			probe.style.borderRadius = "var(--radius-radius)";
			el.parentElement.appendChild(probe);
			const expectedColor = getComputedStyle(probe).color;
			const expectedRadius = getComputedStyle(probe).borderRadius;
			probe.remove();
			return {
				backgroundColor: style.backgroundColor,
				borderColor: style.borderTopColor,
				color: style.color,
				expectedColor,
				paddingTop: style.paddingTop,
				paddingRight: style.paddingRight,
				borderRadius: style.borderRadius,
				expectedRadius,
			};
		}`, nil)
		require.NoError(t, err)
		appearance, ok := appearanceValue.(map[string]any)
		require.True(t, ok, "computed appearance result should be an object")
		assert.Equal(t, "rgba(0, 0, 0, 0)", appearance["backgroundColor"])
		assert.Equal(t, "rgba(0, 0, 0, 0)", appearance["borderColor"])
		assert.Equal(t, appearance["expectedColor"], appearance["color"])
		assert.Equal(t, "2px", appearance["paddingTop"])
		assert.Equal(t, "2px", appearance["paddingRight"])
		assert.Equal(t, appearance["expectedRadius"], appearance["borderRadius"])

		_, err = page.Evaluate("() => document.documentElement.classList.add('dark')", nil)
		require.NoError(t, err)
		_, err = page.WaitForFunction(
			"() => document.documentElement.classList.contains('dark')",
			nil,
		)
		require.NoError(t, err)
		_, err = page.WaitForFunction(`() => {
			const el = document.querySelector("#table-action button");
			if (!el) return false;
			const probe = document.createElement("span");
			probe.style.color = "var(--color-primary-dark)";
			el.parentElement.appendChild(probe);
			const ready = getComputedStyle(el).color === getComputedStyle(probe).color;
			probe.remove();
			return ready;
		}`, nil)
		require.NoError(t, err)
		darkAppearanceValue, err := firstButton.Evaluate(`el => {
			const probe = document.createElement("span");
			probe.style.color = "var(--color-primary-dark)";
			el.parentElement.appendChild(probe);
			const expectedColor = getComputedStyle(probe).color;
			probe.remove();
			return {
				color: getComputedStyle(el).color,
				expectedColor,
			};
		}`, nil)
		require.NoError(t, err)
		darkAppearance, ok := darkAppearanceValue.(map[string]any)
		require.True(t, ok, "dark computed appearance result should be an object")
		assert.Equal(t, darkAppearance["expectedColor"], darkAppearance["color"])

		t.Log("✓ Action table has correctly styled Edit buttons")
	})
}

// TestTable_UsersTable tests the users table with avatars and status badges
func TestTable_UsersTable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Users_Table_Has_Avatars", func(t *testing.T) {
		usersSection := page.Locator("#table-users")

		avatars := usersSection.Locator("td img[src*='avatar-']")
		count, err := avatars.Count()
		require.NoError(t, err)
		assert.Equal(t, 5, count, "users table should have 5 avatar images")

		t.Log("✓ Users table has avatar images")
	})

	t.Run("Users_Table_Has_Status_Badges", func(t *testing.T) {
		usersSection := page.Locator("#table-users")

		activeBadges := usersSection.Locator("span:has-text('Active')")
		activeCount, err := activeBadges.Count()
		require.NoError(t, err)
		assert.Equal(t, 4, activeCount, "should have 4 Active badges")

		canceledBadges := usersSection.Locator("span:has-text('Canceled')")
		canceledCount, err := canceledBadges.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, canceledCount, "should have 1 Canceled badge")

		t.Log("✓ Users table has correct status badges")
	})
}

// TestTable_SortableTable tests the sortable table with HTMX
func TestTable_SortableTable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Sortable_Headers_Have_HTMX_Attributes", func(t *testing.T) {
		sortableTable := page.Locator("#sortable-table")
		sortableHeaders := sortableTable.Locator("thead th[hx-get]")
		count, err := sortableHeaders.Count()
		require.NoError(t, err)
		assert.Equal(t, 3, count, "should have 3 sortable headers (id, name, membership)")

		// Verify hx-get contains order params
		hxGet, err := sortableHeaders.First().GetAttribute("hx-get")
		require.NoError(t, err)
		assert.Contains(t, hxGet, "order_by=")
		assert.Contains(t, hxGet, "order_dir=")

		t.Log("✓ Sortable headers have HTMX attributes")
	})

	t.Run("Sort_Icons_Render", func(t *testing.T) {
		sortableTable := page.Locator("#sortable-table")
		// Should have sort icons (SVGs) in sortable headers
		sortIcons := sortableTable.Locator("thead th svg")
		count, err := sortIcons.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3, "should have sort icons in sortable headers")

		t.Log("✓ Sort icons render in sortable headers")
	})

	t.Run("Clicking_Sort_Header_Fetches_Data", func(t *testing.T) {
		sortableTable := page.Locator("#sortable-table")
		nameHeader := sortableTable.Locator("thead th:has-text('Name')")

		err := nameHeader.Click()
		require.NoError(t, err)

		// Wait for HTMX to swap content
		time.Sleep(50 * time.Millisecond)

		// Verify the tbody still has rows (HTMX replaced content)
		rows := sortableTable.Locator("tbody tr")
		count, err := rows.Count()
		require.NoError(t, err)
		assert.Greater(t, count, 0, "table should still have rows after sorting")

		t.Log("✓ Clicking sort header fetches and displays sorted data")
	})
}

// TestTable_LazyLoad tests lazy-loaded table via HTMX
func TestTable_LazyLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Lazy_Table_Has_HTMX_Trigger", func(t *testing.T) {
		lazyTbody := page.Locator("#lazy-table-tbody")

		trigger, err := lazyTbody.GetAttribute("hx-trigger")
		require.NoError(t, err)
		assert.Equal(t, "click from:#load-lazy-table", trigger)

		initialRows, err := lazyTbody.Locator("tr").Count()
		require.NoError(t, err)
		assert.Equal(t, 1, initialRows, "lazy table should show the loading placeholder before click")

		loadButton := page.Locator("#load-lazy-table")
		err = loadButton.Click()
		require.NoError(t, err)

		time.Sleep(1500 * time.Millisecond)

		rows := lazyTbody.Locator("tr")
		count, err := rows.Count()
		require.NoError(t, err)
		assert.Greater(t, count, 0, "lazy table should have rows after loading")

		t.Log("✓ Lazy-loaded table fetches and displays data")
	})
}

// TestTable_Pagination tests paginated table
func TestTable_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Pagination_Controls_Render", func(t *testing.T) {
		paginationNav := page.Locator("#paginated-table-pagination")
		visible, err := paginationNav.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "pagination controls should be visible")

		// Should show page info
		pageInfo, err := paginationNav.TextContent()
		require.NoError(t, err)
		assert.Contains(t, pageInfo, "Page 1 of 4")

		t.Log("✓ Pagination controls render correctly")
	})

	t.Run("Pagination_Has_Page_Buttons", func(t *testing.T) {
		paginationNav := page.Locator("#paginated-table-pagination")

		buttons := paginationNav.Locator("button, a")
		count, err := buttons.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "should have at least 1 pagination control")

		t.Logf("✓ Pagination has %d controls", count)
	})

	t.Run("Prev_Button_Disabled_On_First_Page", func(t *testing.T) {
		paginationNav := page.Locator("#paginated-table-pagination")
		prevBtn := paginationNav.Locator("button:has-text('Prev'), button:has-text('Previous'), button:has-text('←')")
		count, err := prevBtn.Count()
		require.NoError(t, err)

		if count > 0 {
			disabled, _ := prevBtn.First().GetAttribute("disabled")
			assert.NotNil(t, disabled, "Prev button should be disabled on first page")
			t.Log("✓ Prev button is disabled on first page")
		} else {
			t.Log("No prev button found — pagination may use different controls")
		}
	})

	t.Run("Clicking_Next_Page_Fetches_Data", func(t *testing.T) {
		paginationNav := page.Locator("#paginated-table-pagination")
		nextBtn := paginationNav.Locator("button:has-text('Next'), button:has-text('2'), button:has-text('→')")
		count, err := nextBtn.Count()
		require.NoError(t, err)

		if count == 0 {
			t.Skip("No next/page-2 button found")
		}

		err = nextBtn.First().Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		paginatedTable := page.Locator("#paginated-table")
		rows := paginatedTable.Locator("tbody tr")
		count, err = rows.Count()
		require.NoError(t, err)
		assert.Greater(t, count, 0, "should have rows after clicking page 2")

		t.Log("✓ Clicking page button fetches new data")
	})
}

// TestTable_InfiniteScroll tests infinite scroll table
func TestTable_InfiniteScroll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Infinite_Scroll_Has_Sentinel", func(t *testing.T) {
		infiniteTable := page.Locator("#infinite-table")

		// Should have a script-driven sentinel row for the next page.
		sentinel := infiniteTable.Locator("tr#infinite-table-sentinel[data-hx-get]")
		count, err := sentinel.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count, "should have one scroll sentinel row")

		// Sentinel should point at the next page. The URL is stored in data-hx-get
		// because the component drives htmx.ajax from IntersectionObserver.
		hxGet, err := sentinel.GetAttribute("data-hx-get")
		require.NoError(t, err)
		assert.Contains(t, hxGet, "page=2", "sentinel should request page 2")

		t.Log("✓ Infinite scroll table has sentinel row")
	})

	t.Run("Initial_Rows_Render", func(t *testing.T) {
		infiniteTable := page.Locator("#infinite-table")

		// Should have initial rows + sentinel
		rows := infiniteTable.Locator("tbody tr")
		count, err := rows.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3, "should have at least 3 initial rows")

		t.Log("✓ Infinite scroll table has initial rows")
	})
}

// TestTable_API tests the table rows API endpoint directly
func TestTable_API(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	t.Run("Table_Rows_API_Returns_HTML", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/components/table/rows")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		bodyStr := string(body)

		assert.Contains(t, bodyStr, "Alice Brown")
		assert.Contains(t, bodyStr, "<tr")

		t.Log("✓ Table rows API returns HTML with data")
	})

	t.Run("Table_Rows_API_Sorting", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/components/table/rows?order_by=name&order_dir=desc&per_page=20")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		bodyStr := string(body)

		// With desc sort by name, Sophia should come before Alice
		sophiaIdx := strings.Index(bodyStr, "Sophia Chen")
		aliceIdx := strings.Index(bodyStr, "Alice Brown")
		assert.Greater(t, aliceIdx, sophiaIdx, "Sophia should come before Alice in desc name sort")

		t.Log("✓ Table rows API sorts correctly")
	})

	t.Run("Table_Rows_API_Pagination", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/components/table/rows?page=2&per_page=3")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		bodyStr := string(body)

		// Page 2 with 3 per page should have records 4-6 (Alex, Ryan, Emily)
		assert.Contains(t, bodyStr, "Alex Martinez")
		assert.NotContains(t, bodyStr, "Alice Brown", "page 2 should not contain page 1 data")

		t.Log("✓ Table rows API paginates correctly")
	})

	t.Run("Table_Rows_API_Lazy_Load", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/components/table/rows?variant=lazy")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		bodyStr := string(body)

		assert.Contains(t, bodyStr, "<tr")
		assert.Contains(t, bodyStr, "Alice Brown")

		t.Log("✓ Table rows API lazy load works")
	})

	t.Run("Table_Rows_API_Infinite_Scroll", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/components/table/rows?variant=infinite&page=2")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		bodyStr := string(body)

		// Should have rows and possibly a sentinel for next page
		assert.Contains(t, bodyStr, "<tr")

		t.Log("✓ Table rows API infinite scroll works")
	})
}

// TestTable_AllVariantsRender tests that all table variants render without errors
func TestTable_AllVariantsRender(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	t.Run("Table_Page_Returns_200", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/components/table")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		t.Log("✓ Table page returns 200")
	})

	t.Run("Table_Rows_API_Returns_200", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/components/table/rows")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		t.Log("✓ Table rows API returns 200")
	})
}

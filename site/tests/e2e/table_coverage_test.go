//go:build e2e && (full || table)

package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTableCoverageDemo complements the existing table E2E suite by driving the
// demo variants that map to low-coverage component branches: the check-all
// Alpine toggle, the select-filter HTMX swap, and the infinite-scroll sentinel
// (IntersectionObserver + append). It also asserts the page loads with no JS
// exceptions or console errors.
func TestTableCoverageDemo(t *testing.T) {
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

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	t.Run("page heading renders", func(t *testing.T) {
		require.NoError(t, page.Locator("main").Filter(playwright.LocatorFilterOptions{
			HasText: "Table",
		}).First().WaitFor())
	})

	t.Run("checkbox variant renders header check-all and row checkboxes", func(t *testing.T) {
		container := page.Locator("#table-checkbox")
		require.NoError(t, container.WaitFor())

		header := container.Locator("thead input[type='checkbox']")
		require.NoError(t, header.WaitFor())
		// Header checkbox is bound to Alpine checkAll via x-model.
		model, err := header.Evaluate("el => el.getAttribute('x-model')", nil)
		require.NoError(t, err)
		assert.Equal(t, "checkAll", model)

		// Row checkboxes mirror checkAll via x-bind:checked.
		rowBoxes := container.Locator("tbody input[type='checkbox']")
		count, err := rowBoxes.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "checkbox variant should have row checkboxes")

		// Toggle the header check-all and confirm a row checkbox follows.
		require.NoError(t, header.Check())
		_, err = page.WaitForFunction(
			"() => { const c = document.querySelector('#table-checkbox tbody input[type=\"checkbox\"]'); return c && c.checked === true; }",
			nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
		)
		require.NoError(t, err, "row checkbox should reflect checkAll after toggle")
	})

	t.Run("select filter swaps rows via HTMX", func(t *testing.T) {
		// The filtered demo table carries a membership select filter. Changing it
		// fires applyFilters() -> htmx.ajax targeting the tbody, exercising
		// the bundled filter runtime's configRequest hook end to end. The membership column
		// is the 4th cell (id, name, email, membership).
		membership := page.Locator("#filtered-table-filters select")
		require.NoError(t, membership.WaitFor())

		// Baseline: the unfiltered first page mixes memberships, so not every
		// membership cell is "Gold" before the swap.
		allGold := `() => {
			const cells = Array.from(document.querySelectorAll('#filtered-table-tbody tr td:nth-child(4)'));
			return cells.length > 0 && cells.every(c => c.textContent.trim() === 'Gold');
		}`
		before, err := page.Evaluate(allGold, nil)
		require.NoError(t, err)
		require.Equal(t, false, before, "precondition: unfiltered page should not be all-Gold")

		_, err = membership.SelectOption(playwright.SelectOptionValues{
			Values: &[]string{"Gold"},
		})
		require.NoError(t, err)

		// After the HTMX swap settles, every remaining row must be a Gold member —
		// proving the tbody was actually replaced with filtered server rows.
		_, err = page.WaitForFunction(allGold, nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
		require.NoError(t, err, "filtered tbody should contain only Gold members after swap")
	})

	t.Run("infinite scroll sentinel appends rows", func(t *testing.T) {
		container := page.Locator("#table-infinite")
		require.NoError(t, container.WaitFor())

		sentinel := page.Locator("#infinite-table-sentinel")
		require.NoError(t, sentinel.WaitFor())
		// Sentinel carries the next-page URL for the IntersectionObserver script.
		hxGet, err := sentinel.Evaluate("el => el.getAttribute('data-hx-get')", nil)
		require.NoError(t, err)
		assert.Contains(t, hxGet, "variant=infinite")

		initial, err := page.Locator("#infinite-table-tbody tr").Count()
		require.NoError(t, err)

		// Reveal the sentinel inside its capped 300px scroller to trip the
		// IntersectionObserver and append the next page.
		require.NoError(t, sentinel.ScrollIntoViewIfNeeded())
		_, err = page.WaitForFunction(
			fmt.Sprintf("() => document.querySelectorAll('#infinite-table-tbody tr').length > %d", initial),
			nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(4000)},
		)
		require.NoError(t, err, "infinite scroll should append more rows past the sentinel")
	})

	t.Run("sortable header carries an HTMX sort URL", func(t *testing.T) {
		// Assert the sortable header wiring without triggering a swap: sorting the
		// non-paginated demo table makes the demo handler emit an OOB pagination
		// update against a missing target (a site-handler quirk, not a component
		// bug), which would dirty the strict console-error gate below. The sort
		// swap itself is exercised by table_htmx_test.go and unit-covered.
		header := page.Locator("#sortable-table thead th[hx-get]").First()
		require.NoError(t, header.WaitFor())
		hxGet, err := header.Evaluate("el => el.getAttribute('hx-get')", nil)
		require.NoError(t, err)
		assert.Contains(t, hxGet, "order_by=")
	})

	t.Run("no JS exceptions or console errors", func(t *testing.T) {
		assert.Empty(t, pageErrors, "expected no page errors")
		assert.Empty(t, consoleErrors, "expected no console errors")
	})
}

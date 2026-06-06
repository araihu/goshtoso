package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// gotoExpense navigates to the expense page and waits for Alpine.js to be ready.
// It pre-sets the storage-consent cookie so the first-run notice banner does not
// overlap the row controls, and uses seed=0 to start from an empty list so tests
// can assert exact counts.
func gotoExpense(t *testing.T, page playwright.Page) {
	t.Helper()
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: new("try{document.cookie='gt_storage=allowed; Path=/; SameSite=Lax'}catch(e){}"),
	}))
	_, err := page.Goto(baseURL + "/examples/expense?seed=0")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
}

// rowCountExpr is a JS expression returning the number of real expense rows
// (the dashed empty-state li has no data-expense-id, so it is excluded).
const rowCountExpr = "document.querySelectorAll('#expense-list li[data-expense-id]').length"

// addExpense fills the add form (leaving the default category) and submits,
// waiting until the visible row count reaches want. Use only when the result
// stays within a single page (want <= PerPage).
func addExpense(t *testing.T, page playwright.Page, desc, amount string, want int) {
	t.Helper()
	require.NoError(t, page.Locator("input[name='desc']").Fill(desc))
	require.NoError(t, page.Locator("input[name='amount']").Fill(amount))
	require.NoError(t, page.Locator("#expense-fragment form button[type='submit']").Click())
	_, err := page.WaitForFunction(
		fmt.Sprintf("() => %s === %d", rowCountExpr, want),
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

// addExpenseFirst submits the add form and waits until desc is the first visible
// row. Newest expenses sort first, so this works regardless of how many pages
// the list spans (unlike a row-count assertion, which caps at PerPage).
func addExpenseFirst(t *testing.T, page playwright.Page, desc, amount string) {
	t.Helper()
	require.NoError(t, page.Locator("input[name='desc']").Fill(desc))
	require.NoError(t, page.Locator("input[name='amount']").Fill(amount))
	require.NoError(t, page.Locator("#expense-fragment form button[type='submit']").Click())
	_, err := page.WaitForFunction(
		fmt.Sprintf("() => { const r = document.querySelector('#expense-list li[data-expense-id] span'); return r && r.textContent.trim() === %q; }", desc),
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

// TestExpenseExample_AddAndTotal adds two expenses and verifies the rows appear
// and the running total badge reflects their sum.
func TestExpenseExample_AddAndTotal(t *testing.T) {
	page := newIsolatedPage(t)
	gotoExpense(t, page)

	addExpense(t, page, "Lunch", "12.50", 1)
	addExpense(t, page, "Bus fare", "2.75", 2)

	// Total is computed in integer cents server-side: 1250 + 275 = 1525 → $15.25.
	_, err := page.WaitForFunction(
		"() => document.querySelector('#expense-total').textContent.includes('$15.25')",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

// TestExpenseExample_RejectsBadAmount verifies a non-numeric amount is refused
// (no row added) and a warning toast is shown.
func TestExpenseExample_RejectsBadAmount(t *testing.T) {
	page := newIsolatedPage(t)
	gotoExpense(t, page)

	require.NoError(t, page.Locator("input[name='desc']").Fill("Bad one"))
	require.NoError(t, page.Locator("input[name='amount']").Fill("not-a-number"))
	require.NoError(t, page.Locator("#expense-fragment form button[type='submit']").Click())

	// A "Not added" toast appears and the list stays empty.
	_, err := page.WaitForFunction(
		"() => Array.from(document.querySelectorAll('#toast-container')).some(c => c.textContent.includes('Not added'))",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
	count, err := page.Evaluate("() => " + rowCountExpr)
	require.NoError(t, err)
	require.EqualValues(t, 0, count)

	// Rejected submit must KEEP the typed values so the user can correct them.
	desc, err := page.Locator("input[name='desc']").InputValue()
	require.NoError(t, err)
	require.Equal(t, "Bad one", desc, "rejected submit should retain the description")
}

// TestExpenseExample_ValidSubmitClearsInputs verifies a successful add clears the
// description and amount inputs (ready for the next entry).
func TestExpenseExample_ValidSubmitClearsInputs(t *testing.T) {
	page := newIsolatedPage(t)
	gotoExpense(t, page)

	addExpense(t, page, "Snack", "3.00", 1)
	_, err := page.WaitForFunction(
		"() => document.querySelector(\"input[name='desc']\").value === '' && document.querySelector(\"input[name='amount']\").value === ''",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

// TestExpenseExample_SearchFilter verifies the search box filters the list.
func TestExpenseExample_SearchFilter(t *testing.T) {
	page := newIsolatedPage(t)
	gotoExpense(t, page)

	addExpense(t, page, "Coffee", "4.00", 1)
	addExpense(t, page, "Train", "9.00", 2)

	// Search for "cof" → only the Coffee row remains.
	require.NoError(t, page.Locator("input[name='search']").Fill("cof"))
	_, err := page.WaitForFunction(
		fmt.Sprintf("() => %s === 1 && document.querySelector('#expense-list li[data-expense-id] span').textContent.trim() === 'Coffee'", rowCountExpr),
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

// TestExpenseExample_CategoryFilter adds an item with the default category
// (Food) and verifies the category select filters it in and out. This also
// confirms the add-form combobox submits its default selection (Food) with the
// form.
func TestExpenseExample_CategoryFilter(t *testing.T) {
	page := newIsolatedPage(t)
	gotoExpense(t, page)

	addExpense(t, page, "Groceries", "30.00", 1)

	// Filter to Transport (no rows) → empty state. The category filter is a
	// segmented radio; clicking its label fires a change that posts the filter.
	clickUntil(t, page,
		page.Locator("label[for='expense-cat-Transport']"),
		fmt.Sprintf("() => %s === 0", rowCountExpr))

	// Filter back to Food → the Groceries row returns, proving the add-form
	// combobox stored Food with the new expense.
	clickUntil(t, page,
		page.Locator("label[for='expense-cat-Food']"),
		fmt.Sprintf("() => %s === 1", rowCountExpr))
}

// TestExpenseExample_DeleteUndo deletes a row and restores it via the toast Undo.
func TestExpenseExample_DeleteUndo(t *testing.T) {
	page := newIsolatedPage(t)
	gotoExpense(t, page)

	addExpense(t, page, "Refundable", "5.00", 1)

	// Delete → wait for the toast carrying an Undo action.
	clickUntil(t, page,
		page.Locator("#expense-list li[data-expense-id] button[aria-label='Delete expense']").First(),
		"() => Array.from(document.querySelectorAll('#toast-container button')).some(b => b.textContent.trim() === 'Undo')")

	// Undo → the row returns.
	clickUntil(t, page,
		page.Locator("#toast-container button", playwright.PageLocatorOptions{HasText: "Undo"}),
		fmt.Sprintf("() => %s === 1", rowCountExpr))
}

// TestExpenseExample_Pagination adds enough rows to span two pages, then jumps
// to page two and verifies the remaining rows show.
func TestExpenseExample_Pagination(t *testing.T) {
	page := newIsolatedPage(t)
	gotoExpense(t, page)

	// PerPage is 8; nine rows → two pages (8 + 1). Newest sorts first, so assert
	// on the first-row description rather than a row count (which caps at PerPage).
	for i := 1; i <= 9; i++ {
		addExpenseFirst(t, page, fmt.Sprintf("item %d", i), "1.00")
	}

	// Page 1 shows a full page of 8.
	count, err := page.Evaluate("() => " + rowCountExpr)
	require.NoError(t, err)
	require.EqualValues(t, 8, count)

	// Jump to page 2 → exactly one row remains.
	clickUntil(t, page,
		page.Locator("#expense-pagination a[aria-label='page 2']"),
		fmt.Sprintf("() => %s === 1", rowCountExpr))
}

// TestExpenseExample_ClearAllModal opens the confirm-to-clear modal and confirms,
// emptying the list.
func TestExpenseExample_ClearAllModal(t *testing.T) {
	page := newIsolatedPage(t)
	gotoExpense(t, page)

	addExpense(t, page, "Disposable", "1.00", 1)

	// Open the modal (trigger button), then confirm via the dialog's CTA.
	require.NoError(t, page.Locator("#expense-fragment button:has-text('Clear all')").First().Click())
	clickUntil(t, page,
		page.Locator("[role='dialog'] button:has-text('Clear all')"),
		fmt.Sprintf("() => %s === 0", rowCountExpr))
}

// TestExpenseExample_FragmentNavNoErrors lands elsewhere, navigates to the
// expense app via the sidebar (htmx fragment swap), and verifies there are no
// console/page errors and that adding a row works through that path.
func TestExpenseExample_FragmentNavNoErrors(t *testing.T) {
	page := newIsolatedPage(t)
	dismissCookieBanner(t, page)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("a[href='/examples/expense']").First().Click())
	_, err = page.WaitForFunction("() => !!document.querySelector('#expense-fragment')", nil)
	require.NoError(t, err)

	// Add a row through the fragment-loaded page to confirm HTMX + Alpine work.
	// Sidebar nav does not pass seed=0, so the list opens with sample rows; assert
	// the new row lands first (newest sorts first) rather than an exact count.
	require.NoError(t, page.Locator("input[name='desc']").Fill("via sidebar"))
	require.NoError(t, page.Locator("input[name='amount']").Fill("3.00"))
	// clickUntil rather than a single click: right after the fragment swap, htmx
	// may not have rebound the freshly-inserted add form yet, so the first click
	// can fall through. clickUntil retries until the new row lands first.
	clickUntil(t, page,
		page.Locator("#expense-fragment form button[type='submit']"),
		"() => { const r = document.querySelector('#expense-list li[data-expense-id] span'); return r && r.textContent.trim() === 'via sidebar'; }")

	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav expense page: %v", jsErrors)
}

// TestExpenseExample_SidebarPresent verifies the sidebar Examples section has the
// Expense Tracker link.
func TestExpenseExample_SidebarPresent(t *testing.T) {
	page := newIsolatedPage(t)
	_, err := page.Goto(baseURL + "/examples/expense")
	require.NoError(t, err)
	require.NoError(t, page.Locator("text=Examples").First().WaitFor())
	require.NoError(t, page.Locator("a[href='/examples/expense']").First().WaitFor())
}

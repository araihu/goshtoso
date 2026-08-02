//go:build e2e && (full || example_todo)

package e2e

import (
	"fmt"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

// addTodo fills the add form and submits, waiting for the row to appear.
func addTodo(t *testing.T, page playwright.Page, title string) {
	t.Helper()
	require.NoError(t, page.Locator("input[name='title']").Fill(title))
	require.NoError(t, page.Locator("#todo-fragment form button[type='submit']").Click())
	_, err := page.WaitForFunction(
		fmt.Sprintf("() => Array.from(document.querySelectorAll('#todo-list li span')).some(s => s.textContent.trim() === %q)", title),
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

// gotoTodo navigates to the todo page and waits for Alpine.js to be ready.
func gotoTodo(t *testing.T, page playwright.Page) {
	t.Helper()
	// Suppress the first-run cookie/localStorage notice: a dismissable fixed
	// banner pinned bottom-right (layout.templ). On a fresh browser context it
	// would overlap the right-hand row controls (move/delete) and intercept
	// their clicks; a returning user has already dismissed it.
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: new("try{document.cookie='gt_storage=allowed; Path=/; SameSite=Lax'}catch(e){}"),
	}))
	// seed=0 opts out of the first-visit sample data so tests start with an
	// empty list and assert exact counts.
	_, err := page.Goto(baseURL + "/examples/todo?seed=0")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
}

// TestTodoExample_AddAndCount verifies adding a todo and the active count badge.
func TestTodoExample_AddAndCount(t *testing.T) {
	page := newIsolatedPage(t)
	gotoTodo(t, page)

	addTodo(t, page, "write tests")

	count, err := page.Locator("#todo-list > li").Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)

	badge, err := page.Locator("#todo-count").TextContent()
	require.NoError(t, err)
	require.Contains(t, badge, "1 active")
}

// TestTodoExample_ToggleDelete verifies toggling done state and deleting a todo.
func TestTodoExample_ToggleDelete(t *testing.T) {
	page := newIsolatedPage(t)
	gotoTodo(t, page)

	addTodo(t, page, "toggle me")

	// Toggle via clickUntil — HTMX outerHTML swap rebind race means the checkbox
	// is replaced; we wait for the title span (second span in the li) to have
	// line-through. Use querySelectorAll + Array.from to keep the result boolean.
	clickUntil(t, page, page.Locator("#todo-list > li").First().Locator("input[type='checkbox']"),
		"() => Array.from(document.querySelectorAll('#todo-list li span')).some(s => s.className.includes('line-through'))")

	// Delete the (re-queried after swap) row.
	clickUntil(t, page,
		page.Locator("#todo-list > li button[aria-label='Delete']").First(),
		"() => !document.querySelector('#todo-list > li [data-todo-id]')")
}

// TestTodoExample_Filters verifies the All/Active/Done segmented radio filter:
// the list filters AND the selected radio is checked (the active highlight).
func TestTodoExample_Filters(t *testing.T) {
	page := newIsolatedPage(t)
	gotoTodo(t, page)

	addTodo(t, page, "active one")
	addTodo(t, page, "done one")

	// Mark the second todo done. Use Array.from+filter to ensure boolean return.
	clickUntil(t, page,
		page.Locator("#todo-list > li").Nth(1).Locator("input[type='checkbox']"),
		"() => Array.from(document.querySelectorAll('#todo-list li span')).filter(s => s.className.includes('line-through')).length === 1")

	// Active filter → only the not-done todo visible, and its radio is checked
	// (the highlight survives the list swap because the radios live outside it).
	clickUntil(t, page,
		page.Locator("label[for='todo-filter-active']"),
		"() => document.querySelectorAll('#todo-list > li').length === 1 && document.querySelector('#todo-filter-active').checked")

	// Done filter → only the done todo visible, Done radio checked.
	clickUntil(t, page,
		page.Locator("label[for='todo-filter-done']"),
		"() => document.querySelectorAll('#todo-list > li').length === 1 && document.querySelector('#todo-list li span.line-through') && document.querySelector('#todo-filter-done').checked")

	// All filter → both todos visible, All radio checked.
	clickUntil(t, page,
		page.Locator("label[for='todo-filter-all']"),
		"() => document.querySelectorAll('#todo-list > li').length === 2 && document.querySelector('#todo-filter-all').checked")
}

// TestTodoExample_FragmentNavNoErrors lands elsewhere, navigates to the todo app
// via the sidebar link (htmx fragment swap), and verifies there are no console /
// page errors and that a filter radio and a reorder arrow both work through that
// path. This is the regression guard for the SPA-style navigation a real user
// takes, which a direct-load test cannot catch.
func TestTodoExample_FragmentNavNoErrors(t *testing.T) {
	page := newIsolatedPage(t)
	dismissCookieBanner(t, page)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	// Land on another page first, then navigate via the sidebar (fragment swap).
	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("a[href='/examples/todo']").First().Click())
	_, err = page.WaitForFunction("() => !!document.querySelector('#todo-list > li')", nil)
	require.NoError(t, err)

	// Filter via radio works through the fragment-loaded page.
	clickUntil(t, page,
		page.Locator("label[for='todo-filter-active']"),
		"() => document.querySelector('#todo-filter-active').checked && !!document.querySelector('#todo-list > li')")

	// Reorder arrow works through the fragment-loaded page.
	before, err := page.Evaluate("() => document.querySelector('#todo-list > li').getAttribute('data-todo-id')")
	require.NoError(t, err)
	clickUntil(t, page,
		page.Locator("#todo-list > li").Nth(1).Locator("button[aria-label='Move up']"),
		fmt.Sprintf("() => document.querySelector('#todo-list > li').getAttribute('data-todo-id') !== %q", before))

	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav todo page: %v", jsErrors)
}

// TestTodoExample_ReorderButtons verifies the ↑/↓ reorder buttons.
func TestTodoExample_ReorderButtons(t *testing.T) {
	page := newIsolatedPage(t)
	gotoTodo(t, page)

	addTodo(t, page, "first")
	addTodo(t, page, "second")

	// Move "second" (row index 1) up — it should become the first row.
	clickUntil(t, page,
		page.Locator("#todo-list > li").Nth(1).Locator("button[aria-label='Move up']"),
		"() => { const spans = Array.from(document.querySelectorAll('#todo-list > li:first-child span')); return spans.some(s => s.textContent.trim() === 'second'); }")
}

// TestTodoExample_ClearCompleted verifies the "Clear completed" button.
func TestTodoExample_ClearCompleted(t *testing.T) {
	page := newIsolatedPage(t)
	gotoTodo(t, page)

	addTodo(t, page, "keep")
	addTodo(t, page, "remove")

	// Mark the second todo done. Use Array.from+filter to ensure boolean return.
	clickUntil(t, page,
		page.Locator("#todo-list > li").Nth(1).Locator("input[type='checkbox']"),
		"() => Array.from(document.querySelectorAll('#todo-list li span')).filter(s => s.className.includes('line-through')).length === 1")

	// Clear completed → only "keep" remains.
	clickUntil(t, page,
		page.Locator("#todo-fragment button:has-text('Clear completed')"),
		"() => document.querySelectorAll('#todo-list > li').length === 1")
}

// TestTodoExample_ClearCompletedDisabled verifies that #todo-clear is disabled
// when no task is done and becomes enabled after marking a task done.
func TestTodoExample_ClearCompletedDisabled(t *testing.T) {
	page := newIsolatedPage(t)
	gotoTodo(t, page)

	addTodo(t, page, "only task")

	// The Clear-completed Button lives inside the #todo-clear OOB wrapper.
	// Must be disabled — no completed tasks yet.
	isDisabled, err := page.Evaluate("() => document.querySelector('#todo-clear button').disabled", nil)
	require.NoError(t, err)
	require.Equal(t, true, isDisabled, "Clear completed should be disabled when no done tasks")

	// Mark the task done. Wait for the ClearButton OOB swap to land and the
	// button to become enabled.
	clickUntil(t, page,
		page.Locator("#todo-list > li").First().Locator("input[type='checkbox']"),
		"() => document.querySelector('#todo-clear button') && !document.querySelector('#todo-clear button').disabled")

	// Button must now be enabled.
	isDisabled2, err := page.Evaluate("() => document.querySelector('#todo-clear button').disabled", nil)
	require.NoError(t, err)
	require.Equal(t, false, isDisabled2, "Clear completed should be enabled when there is a done task")
}

// TestTodoExample_UndoDelete adds a todo, deletes it, then clicks Undo and
// verifies the task reappears in #todo-list.
func TestTodoExample_UndoDelete(t *testing.T) {
	page := newIsolatedPage(t)
	gotoTodo(t, page)

	addTodo(t, page, "restore me")

	// Delete the todo. Wait for the toast (carrying an Undo action) to appear.
	clickUntil(t, page,
		page.Locator("#todo-list > li button[aria-label='Delete']").First(),
		"() => Array.from(document.querySelectorAll('#toast-container button')).some(b => b.textContent.trim() === 'Undo')")

	// Click the toast's Undo action. Wait for the todo to reappear in the list.
	clickUntil(t, page,
		page.Locator("#toast-container button", playwright.PageLocatorOptions{HasText: "Undo"}),
		fmt.Sprintf("() => Array.from(document.querySelectorAll('#todo-list li span')).some(s => s.textContent.trim() === %q)", "restore me"))
}

// TestTodoExample_SidebarPresent verifies the sidebar has an Examples section
// and the Todo List navigation link.
func TestTodoExample_SidebarPresent(t *testing.T) {
	page := newIsolatedPage(t)
	_, err := page.Goto(baseURL + "/examples/todo")
	require.NoError(t, err)
	require.NoError(t, page.Locator("text=Examples").First().WaitFor())
	require.NoError(t, page.Locator("a[href='/examples/todo']").First().WaitFor())
}

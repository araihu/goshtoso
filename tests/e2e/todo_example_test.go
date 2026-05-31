package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// newIsolatedPage creates a new browser context (fresh cookies) and page for
// tests that rely on cookie-backed state (like the todo app). Each call returns
// both the context and page; the caller must close the context in t.Cleanup.
func newIsolatedPage(t *testing.T) playwright.Page {
	t.Helper()
	ctx, err := sharedBrowser.NewContext()
	require.NoError(t, err)
	t.Cleanup(func() { _ = ctx.Close() })

	page, err := ctx.NewPage()
	require.NoError(t, err)
	page.SetDefaultTimeout(3000)
	page.SetDefaultNavigationTimeout(5000)
	t.Cleanup(func() { _ = page.Close() })
	return page
}

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
		Content: playwright.String("try{localStorage.setItem('cookieConsent','accepted')}catch(e){}"),
	}))
	_, err := page.Goto(baseURL + "/examples/todo")
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

// TestTodoExample_Filters verifies the All/Active/Done filter segmented buttons.
func TestTodoExample_Filters(t *testing.T) {
	page := newIsolatedPage(t)
	gotoTodo(t, page)

	addTodo(t, page, "active one")
	addTodo(t, page, "done one")

	// Mark the second todo done. Use Array.from+some to ensure boolean return.
	clickUntil(t, page,
		page.Locator("#todo-list > li").Nth(1).Locator("input[type='checkbox']"),
		"() => Array.from(document.querySelectorAll('#todo-list li span')).filter(s => s.className.includes('line-through')).length === 1")

	// Active filter → only the not-done todo visible.
	clickUntil(t, page,
		page.Locator("#todo-fragment button:has-text('Active')"),
		"() => document.querySelectorAll('#todo-list > li').length === 1")

	// Done filter → only the done todo visible.
	clickUntil(t, page,
		page.Locator("#todo-fragment button:has-text('Done')"),
		"() => Array.from(document.querySelectorAll('#todo-list li span')).some(s => s.className.includes('line-through')) && document.querySelectorAll('#todo-list > li').length === 1")

	// All filter → both todos visible.
	clickUntil(t, page,
		page.Locator("#todo-fragment button:has-text('All')"),
		"() => document.querySelectorAll('#todo-list > li').length === 2")
}

// TestTodoExample_ReorderButtons verifies the ↑/↓ reorder buttons.
func TestTodoExample_ReorderButtons(t *testing.T) {
	page := newIsolatedPage(t)
	gotoTodo(t, page)

	addTodo(t, page, "first")
	addTodo(t, page, "second")

	// Move "second" (row index 1) up — should now be the first row.
	// The first span is the ⠿ grab handle; the title span is the second one.
	// Use Array.from + indexing to get the title span of the first li.
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

	// Button must be disabled — no completed tasks yet. Use JS .disabled property
	// (the IDL attribute) which is a boolean, reliable across Playwright versions.
	isDisabled, err := page.Evaluate("() => document.querySelector('#todo-clear').disabled", nil)
	require.NoError(t, err)
	require.Equal(t, true, isDisabled, "#todo-clear should be disabled when no done tasks")

	// Mark the task done. Wait for the ClearButton OOB swap to land and the
	// button to become enabled.
	clickUntil(t, page,
		page.Locator("#todo-list > li").First().Locator("input[type='checkbox']"),
		"() => document.querySelector('#todo-clear') && !document.querySelector('#todo-clear').disabled")

	// Button must now be enabled.
	isDisabled2, err := page.Evaluate("() => document.querySelector('#todo-clear').disabled", nil)
	require.NoError(t, err)
	require.Equal(t, false, isDisabled2, "#todo-clear should be enabled when there is a done task")
}

// TestTodoExample_UndoDelete adds a todo, deletes it, then clicks Undo and
// verifies the task reappears in #todo-list.
func TestTodoExample_UndoDelete(t *testing.T) {
	page := newIsolatedPage(t)
	gotoTodo(t, page)

	addTodo(t, page, "restore me")

	// Delete the todo. Wait for the undo bar to appear.
	clickUntil(t, page,
		page.Locator("#todo-list > li button[aria-label='Delete']").First(),
		"() => document.querySelector('#todo-undo button') !== null")

	// Click Undo. Wait for the todo to reappear in the list.
	clickUntil(t, page,
		page.Locator("#todo-undo button"),
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

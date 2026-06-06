package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const searchActiveMarker = "shadow-[inset_2px_0_0"

// openComponentSearch opens the demo search dialog and waits for the input to
// be focusable. It returns the input locator.
func openComponentSearch(t *testing.T, page playwright.Page) playwright.Locator {
	t.Helper()
	trigger := page.Locator("#component-search button[aria-haspopup='dialog']")
	require.NoError(t, trigger.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, trigger.Click())

	input := page.Locator("#component-search-input")
	require.NoError(t, input.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	return input
}

// resultIsActive reports whether the result with the given id currently carries
// the active-row inset-shadow class applied by the isActive() Alpine binding.
func resultIsActive(t *testing.T, page playwright.Page, id string) bool {
	t.Helper()
	class, err := page.Locator(id).GetAttribute("class")
	require.NoError(t, err)
	return strings.Contains(class, searchActiveMarker)
}

// waitResultActive waits until the result with the given id gains the active-row
// class, polling the live DOM so Alpine's class binding has time to settle.
func waitResultActive(t *testing.T, page playwright.Page, id string) {
	t.Helper()
	_, err := page.WaitForFunction(
		"sel => { const el = document.querySelector(sel); return !!el && el.className.includes('"+searchActiveMarker+"'); }",
		id,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err)
}

// TestSearchCoverageKeyboardNavigation drives the dialog with the keyboard so
// the Alpine move()/isActive()/setActive() and Escape-close paths execute.
func TestSearchCoverageKeyboardNavigation(t *testing.T) {
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

	_, err := page.Goto(baseURL+"/components/search", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	require.NoError(t, page.Locator("#search-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	input := openComponentSearch(t, page)

	// "search" matches the Sidebar and Text Input descriptions only, giving a
	// deterministic two-result list in DOM order.
	require.NoError(t, input.Fill("search"))
	require.NoError(t, page.Locator("#search-demo-sidebar:visible").WaitFor())
	require.NoError(t, page.Locator("#search-demo-text-input:visible").WaitFor())

	visible, err := page.Locator("#component-search-dialog [data-search-result]:visible").Count()
	require.NoError(t, err)
	require.Equal(t, 2, visible, "query should yield exactly two visible results")

	// First result is active on a fresh query (activeIndex 0).
	waitResultActive(t, page, "#search-demo-sidebar")
	assert.False(t, resultIsActive(t, page, "#search-demo-text-input"))

	// ArrowDown advances the active row to the second result.
	require.NoError(t, input.Press("ArrowDown"))
	waitResultActive(t, page, "#search-demo-text-input")
	assert.False(t, resultIsActive(t, page, "#search-demo-sidebar"))

	// ArrowUp returns the active row to the first result.
	require.NoError(t, input.Press("ArrowUp"))
	waitResultActive(t, page, "#search-demo-sidebar")

	// Hovering the second result moves the active row via setActive().
	require.NoError(t, page.Locator("#search-demo-text-input").Hover())
	waitResultActive(t, page, "#search-demo-text-input")

	// Escape closes the dialog and clears the query.
	require.NoError(t, input.Press("Escape"))
	require.NoError(t, page.Locator("#component-search-dialog").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden,
	}))

	assert.Empty(t, pageErrors, "no uncaught page errors")
	assert.Empty(t, consoleErrors, "no console errors")
}

// TestSearchCoverageEnterNavigates exercises choose()/selectResult() by pressing
// Enter on a single match and asserting client-side navigation to the result
// href.
func TestSearchCoverageEnterNavigates(t *testing.T) {
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

	_, err := page.Goto(baseURL+"/components/search", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	input := openComponentSearch(t, page)

	// "kbd" matches only the KBD result by title.
	require.NoError(t, input.Fill("kbd"))
	require.NoError(t, page.Locator("#search-demo-kbd:visible").WaitFor())
	visible, err := page.Locator("#component-search-dialog [data-search-result]:visible").Count()
	require.NoError(t, err)
	require.Equal(t, 1, visible)

	require.NoError(t, input.Press("Enter"))
	require.NoError(t, page.WaitForURL("**/components/kbd"))

	assert.Empty(t, pageErrors, "no uncaught page errors")
	assert.Empty(t, consoleErrors, "no console errors")
}

// TestSearchCoverageClickOutsideCloses verifies the backdrop click.self handler
// closes the dialog without navigating.
func TestSearchCoverageClickOutsideCloses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/search", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	_ = openComponentSearch(t, page)

	dialog := page.Locator("#component-search-dialog")
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	// Dispatch a click whose target is the backdrop element itself so the
	// x-on:click.self handler fires (a real click(el) sets target===currentTarget),
	// without depending on backdrop geometry inside the transformed layout.
	_, err = dialog.Evaluate("el => el.click()", nil)
	require.NoError(t, err)

	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden,
	}))
	// Still on the search page; backdrop click must not navigate.
	assert.Contains(t, page.URL(), "/components/search")
	assert.Empty(t, consoleErrors, "no console errors")
}

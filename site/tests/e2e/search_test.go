//go:build e2e && (full || search)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_PageLoadsAndFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/search", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#search-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	trigger := page.Locator("#component-search button[aria-haspopup='dialog']")
	require.NoError(t, trigger.Click())

	input := page.Locator("#component-search-input")
	require.NoError(t, input.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	visible, err := page.Locator("#component-search-dialog [data-search-result]:visible").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, visible, "search should open with an empty result list")

	require.NoError(t, input.Fill("kbd"))
	assert.NoError(t, page.Locator("#search-demo-kbd:visible").WaitFor())
	assert.NoError(t, page.Locator("#search-demo-kbd .search-highlight").WaitFor())

	visible, err = page.Locator("#component-search-dialog [data-search-result]:visible").Count()
	require.NoError(t, err)
	assert.Equal(t, 1, visible)

	require.NoError(t, input.Fill("zzzz"))
	assert.NoError(t, page.Locator("#component-search-dialog p", playwright.PageLocatorOptions{HasText: "No results found."}).WaitFor())
}

func TestSearch_ItemsURLFetchesAndFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/search", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	trigger := page.Locator("#remote-search button[aria-haspopup='dialog']")
	require.NoError(t, trigger.Click())

	input := page.Locator("#remote-search-input")
	require.NoError(t, input.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	require.NoError(t, input.Fill("fetched"))
	require.NoError(t, page.Locator("#search-remote-table:visible").WaitFor())

	visible, err := page.Locator("#remote-search-dialog [data-search-result]:visible").Count()
	require.NoError(t, err)
	assert.Equal(t, 1, visible)

	require.NoError(t, input.Fill("teams"))
	result := page.Locator("#search-remote-list-teams:visible")
	require.NoError(t, result.WaitFor())
	assert.NoError(t, result.Locator("text=GET").WaitFor())
	assert.NoError(t, result.Locator("text=/teams").WaitFor())
	method, err := result.GetAttribute("data-search-method")
	require.NoError(t, err)
	assert.Equal(t, "GET", method)
	path, err := result.GetAttribute("data-search-path")
	require.NoError(t, err)
	assert.Equal(t, "/teams", path)

	// "rc" is not a substring of "Remote combobox". Fuzzy mode must rank it
	// from the same normalized client-side records used by ItemsURL.
	require.NoError(t, input.Fill("rc"))
	require.NoError(t, page.Locator("#search-remote-combobox:visible").WaitFor())
	visible, err = page.Locator("#remote-search-dialog [data-search-result]:visible").Count()
	require.NoError(t, err)
	assert.Equal(t, 2, visible)
	firstID, err := page.Locator("#remote-search-dialog [data-search-result]:visible").First().GetAttribute("id")
	require.NoError(t, err)
	assert.Equal(t, "search-remote-combobox", firstID)
}

func TestSearch_FuzzyModeFiltersDOMItems(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/search", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)

	trigger := page.Locator("#fuzzy-search button[aria-haspopup='dialog']")
	require.NoError(t, trigger.Click())
	input := page.Locator("#fuzzy-search-input")
	require.NoError(t, input.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))

	// "srm" is a compact subsequence, not a literal substring.
	require.NoError(t, input.Fill("srm"))
	require.NoError(t, page.Locator("#search-fuzzy-result:visible").WaitFor())

	// Explicit score tiers keep even a sparse title match ahead of a compact
	// text match; the secondary fuzzy score must never cross priority bands.
	ranked, err := page.Evaluate(`() => {
		const root = document.createElement("div");
		root.dataset.searchMatchMode = "fuzzy";
		const modal = window.goshtosoSearchModal(root);
		const values = [
			{ id: "text", title: "none", text: "a-b-c" },
			{ id: "title", title: "a" + "x".repeat(500) + "b" + "x".repeat(500) + "c", text: "none" },
		];
		return modal.rankedMatches(values, (value) => modal.resultScore("abc", value.title, value.text)).map((value) => value.id);
	}`, nil)
	require.NoError(t, err)
	assert.Equal(t, []any{"title", "text"}, ranked)

	contextRanked, err := page.Evaluate(`() => {
		const root = document.createElement("div");
		root.dataset.searchMatchMode = "substring";
		const modal = window.goshtosoSearchModal(root);
		const values = [
			{ id: "fallback", title: "Line chart", text: "chart component", priority: 0 },
			{ id: "active", title: "Line chart", text: "chart component", priority: 1 },
		];
		return modal.rankedMatches(values, (value) => modal.resultScore("line", value.title, value.text, value.priority)).map((value) => value.id);
	}`, nil)
	require.NoError(t, err)
	assert.Equal(t, []any{"active", "fallback"}, contextRanked, "active family should win equal-quality matches")
}

func TestSidebarSearch_UsesKbdAndNavigates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	trigger := page.Locator("#docs-search button[aria-haspopup='dialog']")
	require.NoError(t, trigger.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	kbdText, err := trigger.Locator("kbd").TextContent()
	require.NoError(t, err)
	assert.Contains(t, kbdText, "K")
	assert.NotContains(t, kbdText, "Esc")

	require.NoError(t, page.Keyboard().Press("Meta+K"))
	input := page.Locator("#docs-search-input")
	require.NoError(t, input.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	parentTag, err := page.Locator("#docs-search-dialog").Evaluate("el => el.parentElement.tagName", nil)
	require.NoError(t, err)
	assert.Equal(t, "BODY", parentTag, "docs search dialog should be teleported out of the transformed sidebar")
	dialogKbdText, err := page.Locator("#docs-search-dialog kbd").TextContent()
	require.NoError(t, err)
	assert.Contains(t, dialogKbdText, "Esc")

	visible, err := page.Locator("#docs-search-dialog [data-search-result]:visible").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, visible, "docs sidebar search should open empty")

	require.NoError(t, input.Fill("button"))
	assert.NoError(t, page.Locator("#search-button:visible").WaitFor())
	visible, err = page.Locator("#docs-search-dialog [data-search-result]:visible").Count()
	require.NoError(t, err)
	assert.Equal(t, 4, visible, "docs sidebar search should cap matches like PenguinUI")

	firstTitle, err := page.Locator("#docs-search-dialog [data-search-result]:visible strong").First().TextContent()
	require.NoError(t, err)
	assert.Equal(t, "Button", firstTitle)
	assert.NoError(t, page.Locator("#docs-search-dialog [data-search-result]:visible .search-highlight").First().WaitFor())
	firstClass, err := page.Locator("#docs-search-dialog [data-search-result]:visible").First().GetAttribute("class")
	require.NoError(t, err)
	assert.NotContains(t, firstClass, "bg-primary", "search results should not use full primary row fill")
	assert.NotContains(t, firstClass, "color-primary", "active row accent should not reuse the match-highlight color")

	buttonDescription, err := page.Locator("#search-button p").TextContent()
	require.NoError(t, err)
	assert.Contains(t, buttonDescription, "variants")
	assert.NotContains(t, buttonDescription, "component documentation and examples")

	require.NoError(t, input.Fill("search"))
	require.NoError(t, input.Press("Enter"))

	require.NoError(t, page.WaitForURL("**/components/search"))
	require.NoError(t, page.Locator("#search-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
}

func TestSidebarSearch_DoesNotOpenDuringSidebarNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/search", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#docs-search button[aria-haspopup='dialog']").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	visible, err := page.Locator("#docs-search-dialog:visible").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, visible)

	require.NoError(t, page.Locator("a[data-sidebar-item='Badge']").Click())
	require.NoError(t, page.WaitForURL("**/components/badge"))
	require.NoError(t, page.Locator("#main-content h1", playwright.PageLocatorOptions{HasText: "Badge"}).WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	visible, err = page.Locator("#docs-search-dialog:visible").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, visible, "sidebar navigation should not open the docs search modal")
}

func TestSidebarSearch_GlobalShortcutOnlyOpensDocsSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/search", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Keyboard().Press("Meta+K"))
	require.NoError(t, page.Locator("#docs-search-input").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	componentVisible, err := page.Locator("#component-search-dialog:visible").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, componentVisible, "demo search should not open from the sidebar global shortcut")

	customVisible, err := page.Locator("#custom-search-dialog:visible").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, customVisible, "custom demo search should not open from the sidebar global shortcut")
}

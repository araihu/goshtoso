package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSidebarCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/sidebar", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	require.NoError(t, page.Locator("#sidebar-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	mainText, err := page.Locator("main").TextContent()
	require.NoError(t, err)
	require.Contains(t, mainText, "Sidebar")

	t.Run("simple and section variants expose active states and badges", func(t *testing.T) {
		simple := page.Locator("#sidebar-simple")
		require.NoError(t, simple.ScrollIntoViewIfNeeded())
		require.NoError(t, simple.Locator("input[type='search'][placeholder='Search...']").WaitFor())
		require.NoError(t, simple.Locator("a.text-primary").Filter(playwright.LocatorFilterOptions{HasText: "Profile"}).WaitFor())
		require.NoError(t, simple.Locator("a").Filter(playwright.LocatorFilterOptions{HasText: "Inbox"}).GetByText("3", playwright.LocatorGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())

		sections := page.Locator("#sidebar-sections")
		require.NoError(t, sections.ScrollIntoViewIfNeeded())
		require.NoError(t, sections.Locator("[data-sidebar-section='Components']").GetByText("Table", playwright.LocatorGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())
		require.NoError(t, sections.Locator("a.pointer-events-none").Filter(playwright.LocatorFilterOptions{HasText: "Overview"}).WaitFor())
	})

	t.Run("nested section children and badges render", func(t *testing.T) {
		subItems := page.Locator("#sidebar-sub-items")
		require.NoError(t, subItems.ScrollIntoViewIfNeeded())

		require.False(t, sidebarLinkVisibleContaining(t, page, "#sidebar-sub-items", "Create User"))
		require.False(t, sidebarLinkVisibleContaining(t, page, "#sidebar-sub-items", "List Products"))

		users := subItems.Locator("a[aria-controls='ep-users-children']")
		products := subItems.Locator("a[aria-controls='ep-products-children']")
		require.NoError(t, users.Click())
		waitForSidebarLinkVisibleContaining(t, page, "#sidebar-sub-items", "Create User")

		require.NoError(t, products.Click())
		waitForSidebarLinkVisibleContaining(t, page, "#sidebar-sub-items", "Create User")
		waitForSidebarLinkVisibleContaining(t, page, "#sidebar-sub-items", "List Products")

		for _, badge := range []string{"POST", "GET", "PUT", "DEL"} {
			require.NoError(t, subItems.Locator("sup").Filter(playwright.LocatorFilterOptions{HasText: badge}).First().WaitFor())
		}
	})

	t.Run("collapsible demo updates live Alpine visibility", func(t *testing.T) {
		collapsible := page.Locator("#sidebar-collapsible")
		require.NoError(t, collapsible.ScrollIntoViewIfNeeded())

		require.True(t, sidebarVisible(t, page, "#sidebar-collapsible", "Dashboard"))
		require.False(t, sidebarVisible(t, page, "#sidebar-collapsible", "Articles"))

		contentButton := collapsible.Locator("button").Filter(playwright.LocatorFilterOptions{HasText: "Content"})
		require.NoError(t, contentButton.Click())
		_, err := page.WaitForFunction(
			`() => Array.from(document.querySelectorAll("#sidebar-collapsible a")).some((el) => el.textContent.trim() === "Articles" && el.offsetParent !== null)`,
			nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)
	})

	t.Run("overlay opens and closes with Alpine state", func(t *testing.T) {
		overlay := page.Locator("#sidebar-overlay")
		require.NoError(t, overlay.ScrollIntoViewIfNeeded())

		panel := overlay.GetByText("Overlay", playwright.LocatorGetByTextOptions{Exact: playwright.Bool(true)})
		visible, err := panel.IsVisible()
		require.NoError(t, err)
		require.False(t, visible)

		require.NoError(t, overlay.Locator("button[aria-label='Open sidebar']").Click())
		require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		}))
		require.NoError(t, overlay.Locator("input[type='search'][placeholder='Search...']").WaitFor())

		require.NoError(t, overlay.Locator("button[aria-label='Close sidebar']").Click())
		require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateHidden,
			Timeout: playwright.Float(3000),
		}))
	})

	require.Empty(t, jsErrors, "no JS console/page errors on sidebar demo: %v", jsErrors)
}

func sidebarVisible(t *testing.T, page playwright.Page, scope string, text string) bool {
	t.Helper()

	result, err := page.Evaluate(`([scope, text]) => Array.from(document.querySelectorAll(scope + " a")).some((el) => el.textContent.trim() === text && el.offsetParent !== null)`, []string{scope, text})
	require.NoError(t, err)

	return result.(bool)
}

func sidebarLinkVisibleContaining(t *testing.T, page playwright.Page, scope string, text string) bool {
	t.Helper()

	result, err := page.Evaluate(`([scope, text]) => Array.from(document.querySelectorAll(scope + " a")).some((el) => el.textContent.includes(text) && el.offsetParent !== null)`, []string{scope, text})
	require.NoError(t, err)

	return result.(bool)
}

func waitForSidebarLinkVisibleContaining(t *testing.T, page playwright.Page, scope string, text string) {
	t.Helper()

	_, err := page.WaitForFunction(
		`([scope, text]) => Array.from(document.querySelectorAll(scope + " a")).some((el) => el.textContent.includes(text) && el.offsetParent !== null)`,
		[]string{scope, text},
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err)
}

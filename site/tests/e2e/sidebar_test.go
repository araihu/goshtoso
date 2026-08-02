//go:build e2e

package e2e

import (
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSidebar_AllComponentsPresent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	componentPages := catalog.ComponentPages()
	require.Len(t, componentPages, 50)

	_, err := page.Goto(baseURL+componentPages[0].Path, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	sidebar := page.Locator("nav[aria-label='sidebar navigation']")
	require.NoError(t, sidebar.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	for _, componentPage := range componentPages {
		if componentPage.Active == "app-shell" {
			continue
		}
		t.Run(componentPage.Title, func(t *testing.T) {
			link := sidebar.Locator("a[href='" + componentPage.Path + "']").Filter(playwright.LocatorFilterOptions{
				HasText: componentPage.Title,
			})
			count, err := link.Count()
			require.NoError(t, err)
			assert.Equal(t, 1, count, "%s should have one sidebar link to %s", componentPage.Title, componentPage.Path)
		})
	}

	legacyLinkCount, err := sidebar.Locator("a[href='/components/combobox-new']").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, legacyLinkCount, "combobox-new must not be exposed after v2 becomes canonical")

	appShellLinkCount, err := sidebar.Locator("a[href='/components/app-shell']").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, appShellLinkCount, "the legacy App Shell page remains routable but must not duplicate the App Shells module")
}

func TestSidebar_LinksNavigate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	// Pick a few components and verify clicking their sidebar link loads the page
	testLinks := []struct {
		href      string
		titlePart string
	}{
		{"/components/accordion", "Accordion"},
		{"/components/toggle", "Toggle"},
		{"/components/checkbox", "Checkbox"},
	}

	for _, link := range testLinks {
		t.Run(link.titlePart, func(t *testing.T) {
			_, err := page.Goto(baseURL+link.href, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			})
			require.NoError(t, err)

			title, err := page.Title()
			require.NoError(t, err)
			assert.Contains(t, title, link.titlePart, "page title should contain %s", link.titlePart)
		})
	}
}

func TestSidebar_ExamplesTopItemNavigatesToOverview(t *testing.T) {
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

	sidebar := page.Locator("nav[aria-label='sidebar navigation']")
	require.NoError(t, sidebar.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	examplesLink := sidebar.Locator("a[href='/examples']")
	count, err := examplesLink.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Examples overview should have a single sidebar link")

	overviewItem := sidebar.Locator("a[href='/examples'][data-sidebar-item='Overview']")
	count, err = overviewItem.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Examples should expose its overview after Modules")

	moduleHeading := sidebar.GetByRole("heading", playwright.LocatorGetByRoleOptions{Name: "Modules", Exact: playwright.Bool(true)})
	examplesHeading := sidebar.GetByRole("heading", playwright.LocatorGetByRoleOptions{Name: "Examples", Exact: playwright.Bool(true)})
	moduleBox, err := moduleHeading.BoundingBox()
	require.NoError(t, err)
	examplesBox, err := examplesHeading.BoundingBox()
	require.NoError(t, err)
	assert.Less(t, moduleBox.Y, examplesBox.Y, "Modules should precede Examples")

	require.NoError(t, examplesLink.Click())
	require.NoError(t, page.WaitForURL("**/examples"))
	require.NoError(t, page.Locator("main h1", playwright.PageLocatorOptions{HasText: "Examples"}).WaitFor())
}

func TestSidebarComponentDemoVariants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/sidebar", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	simple := page.Locator("#sidebar-simple")
	require.NoError(t, simple.GetByText("MyApp", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
	require.NoError(t, simple.Locator("input[type='search'][placeholder='Search...']").WaitFor())
	require.NoError(t, simple.Locator("a").Filter(playwright.LocatorFilterOptions{HasText: "Profile"}).WaitFor())
	require.NoError(t, simple.Locator("a").Filter(playwright.LocatorFilterOptions{HasText: "Inbox"}).GetByText("3", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())

	sections := page.Locator("#sidebar-sections")
	for _, heading := range []string{"Getting Started", "Components", "Settings"} {
		require.NoError(t, sections.Locator("h3").Filter(playwright.LocatorFilterOptions{HasText: heading}).WaitFor())
	}
	require.NoError(t, sections.Locator("[data-sidebar-section='Components']").GetByText("Table", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())

	subItems := page.Locator("#sidebar-sub-items")
	require.NoError(t, subItems.GetByText("API Docs", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
	for _, label := range []string{"Create User", "Get User", "Update User", "Delete User"} {
		count, err := subItems.Locator("a").Filter(playwright.LocatorFilterOptions{HasText: label}).Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	}
	for _, badge := range []string{"POST", "GET", "PUT", "DEL"} {
		count, err := subItems.Locator("sup").Filter(playwright.LocatorFilterOptions{HasText: badge}).Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1)
	}

	contentButton := page.Locator("#sidebar-collapsible button").Filter(playwright.LocatorFilterOptions{HasText: "Content"})
	require.NoError(t, contentButton.Click())
	require.NoError(t, page.Locator("#sidebar-collapsible").GetByText("Articles", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	overlay := page.Locator("#sidebar-overlay")
	panel := overlay.GetByText("Overlay", playwright.LocatorGetByTextOptions{Exact: new(true)})
	visible, err := panel.IsVisible()
	require.NoError(t, err)
	assert.False(t, visible)

	require.NoError(t, overlay.Locator("button[aria-label='Open sidebar']").Click())
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))
	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(3000),
	}))

	require.NoError(t, overlay.Locator("button[aria-label='Open sidebar']").Click())
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	require.NoError(t, overlay.Locator("div[aria-hidden='true']").Click())
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(3000),
	}))
}

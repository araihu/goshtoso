package e2e

import (
	"testing"

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

	_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	sidebar := page.Locator("nav[aria-label='sidebar navigation']")
	require.NoError(t, sidebar.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	// Every component directory should have a sidebar link
	expectedComponents := []struct {
		href  string
		label string
	}{
		{"/components/accordion", "Accordion"},
		{"/components/alert", "Alert"},
		{"/components/avatar", "Avatar"},
		{"/components/badge", "Badge"},
		{"/components/banner", "Banner"},
		{"/components/button", "Buttons"},
		{"/components/card", "Card"},
		{"/components/carousel", "Carousel"},
		{"/components/checkbox", "Checkbox"},
		{"/components/combobox", "Combobox"},
		{"/components/dropdown", "Dropdown"},
		{"/components/form", "Form"},
		{"/components/kbd", "KBD"},
		{"/components/structured-input", "Structured Input"},
		{"/components/link", "Link"},
		{"/components/modal", "Modal"},
		{"/components/navbar", "Navbar"},
		{"/components/pagination", "Pagination"},
		{"/components/palette", "Palette"},
		{"/components/range", "Range"},
		{"/components/rating", "Rating"},
		{"/components/select", "Select"},
		{"/components/sidebar", "Sidebar"},
		{"/components/spinner", "Spinner"},
		{"/components/steps", "Steps"},
		{"/components/table", "Table"},
		{"/components/tabs", "Tabs"},
		{"/components/tags-list", "Tags List"},
		{"/components/text-input", "Text Input"},
		{"/components/textarea", "Textarea"},
		{"/components/toast", "Toast"},
		{"/components/toggle", "Toggle"},
		{"/components/tooltip", "Tooltip"},
	}

	for _, comp := range expectedComponents {
		t.Run(comp.label, func(t *testing.T) {
			link := sidebar.Locator("a[href='" + comp.href + "']")
			count, err := link.Count()
			require.NoError(t, err)
			assert.Equal(t, 1, count, "%s should have a sidebar link to %s", comp.label, comp.href)
		})
	}

	legacyLinkCount, err := sidebar.Locator("a[href='/components/combobox-new']").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, legacyLinkCount, "combobox-new must not be exposed after v2 becomes canonical")
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
	assert.Equal(t, 1, count, "Examples overview should have a single top-level sidebar link")

	oldOverviewItem := sidebar.Locator("a[href='/examples'][data-sidebar-item='Overview']")
	count, err = oldOverviewItem.Count()
	require.NoError(t, err)
	assert.Zero(t, count, "old Examples > Overview nav item should be gone")

	require.NoError(t, examplesLink.Click())
	require.NoError(t, page.WaitForURL("**/examples"))
	require.NoError(t, page.Locator("main h1", playwright.PageLocatorOptions{HasText: "Examples"}).WaitFor())
}

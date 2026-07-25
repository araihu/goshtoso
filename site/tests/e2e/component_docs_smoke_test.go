package e2e

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestAllComponentDocsDirectLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	for _, entry := range catalog.ComponentPages() {
		entry := entry
		t.Run(entry.Active, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			failures := watchPageFailures(page)
			response, err := page.Goto(baseURL+entry.Path, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			})
			require.NoError(t, err)
			require.NotNil(t, response)
			require.Equal(t, http.StatusOK, response.Status())

			heading, err := page.Locator("main h1").TextContent()
			require.NoError(t, err)
			require.Equal(t, entry.Title, strings.TrimSpace(heading))

			descriptionCount, err := page.Locator("[data-component-description]").Count()
			require.NoError(t, err)
			require.Equal(t, 1, descriptionCount)

			previewCount, err := page.Locator("[data-demo-preview]").Count()
			require.NoError(t, err)
			require.GreaterOrEqual(t, previewCount, 1)

			codeCount, err := page.Locator("[data-demo-code]").Count()
			require.NoError(t, err)
			require.GreaterOrEqual(t, codeCount, 1)

			apiCount, err := page.Locator("[data-api-reference]").Count()
			require.NoError(t, err)
			require.Equal(t, 1, apiCount)
			waitForPageSettled(t, page)
			failures.RequireEmpty(t)
		})
	}
}

func TestAllComponentDocsFragmentNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	entries := catalog.ComponentPages()
	require.Len(t, entries, 42)

	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	response, err := page.Goto(baseURL+entries[0].Path, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, http.StatusOK, response.Status())

	for index, entry := range entries {
		if index > 0 {
			link := page.Locator(fmt.Sprintf(
				`#sidebar-nav-content a[href=%q]`,
				entry.Path,
			))
			require.NoError(t, link.ScrollIntoViewIfNeeded())
			clickUntil(t, page, link, componentDocsDestinationReady(entry))
		}

		requireComponentDocsDestination(t, page, failures, entry)
	}
}

func componentDocsDestinationReady(entry catalog.Entry) string {
	return fmt.Sprintf(
		`() => window.location.pathname === %q &&
			document.querySelector("main h1")?.textContent.trim() === %q`,
		entry.Path,
		entry.Title,
	)
}

func requireComponentDocsDestination(
	t *testing.T,
	page playwright.Page,
	failures *pageFailures,
	entry catalog.Entry,
) {
	t.Helper()

	activeSelector := fmt.Sprintf(
		`#sidebar-nav-content a[href=%q][aria-current="page"]`,
		entry.Path,
	)
	_, err := page.WaitForFunction(
		fmt.Sprintf(
			`() => window.location.pathname === %q &&
				document.querySelector("main h1")?.textContent.trim() === %q &&
				document.querySelector(%q) !== null`,
			entry.Path,
			entry.Title,
			activeSelector,
		),
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err)

	path, err := page.Evaluate("() => window.location.pathname")
	require.NoError(t, err)
	require.Equal(t, entry.Path, path)

	heading, err := page.Locator("main h1").TextContent()
	require.NoError(t, err)
	require.Equal(t, entry.Title, strings.TrimSpace(heading))

	apiCount, err := page.Locator("[data-api-reference]").Count()
	require.NoError(t, err)
	require.Equal(t, 1, apiCount)

	previewCount, err := page.Locator("[data-demo-preview]").Count()
	require.NoError(t, err)
	require.GreaterOrEqual(t, previewCount, 1)

	activeLink := page.Locator(fmt.Sprintf(
		`#sidebar-nav-content a[href=%q]`,
		entry.Path,
	))
	current, err := activeLink.GetAttribute("aria-current")
	require.NoError(t, err)
	require.Equal(t, "page", current)

	runtimesReady, err := page.Evaluate(
		"() => typeof window.Alpine !== 'undefined' && typeof window.htmx !== 'undefined'",
	)
	require.NoError(t, err)
	require.Equal(t, true, runtimesReady)

	waitForPageSettled(t, page)
	failures.RequireEmpty(t)
}

func TestComponentDocsThemeMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	pages := []string{
		"/components/button",
		"/components/form",
		"/components/modal",
		"/components/table",
		"/components/toast",
	}
	states := []struct {
		name  string
		theme string
		dark  bool
	}{
		{name: "goshtoso-light", theme: "goshtoso", dark: false},
		{name: "goshtoso-dark", theme: "goshtoso", dark: true},
		{name: "minimal-light", theme: "minimal", dark: false},
		{name: "minimal-dark", theme: "minimal", dark: true},
	}

	for _, path := range pages {
		for _, state := range states {
			t.Run(strings.TrimPrefix(path, "/components/")+"/"+state.name, func(t *testing.T) {
				page := newPage(t, sharedBrowser)
				failures := watchPageFailures(page)
				script := fmt.Sprintf(`
    document.cookie = "gt_storage=allowed; Path=/; SameSite=Lax";
    localStorage.setItem("theme", %q);
    localStorage.setItem("darkMode", %q);
`, state.theme, strconv.FormatBool(state.dark))
				require.NoError(t, page.AddInitScript(playwright.Script{
					Content: &script,
				}))

				response, err := page.Goto(baseURL+path, playwright.PageGotoOptions{
					WaitUntil: playwright.WaitUntilStateDomcontentloaded,
				})
				require.NoError(t, err)
				require.NotNil(t, response)
				require.Equal(t, http.StatusOK, response.Status())

				html := page.Locator("html")
				theme, err := html.GetAttribute("data-theme")
				require.NoError(t, err)
				require.Equal(t, state.theme, theme)

				className, err := html.GetAttribute("class")
				require.NoError(t, err)
				require.Equal(t, state.dark, hasClass(className, "dark"))

				preview := page.Locator("[data-demo-preview]").First()
				visible, err := preview.IsVisible()
				require.NoError(t, err)
				require.True(t, visible)

				box, err := preview.BoundingBox()
				require.NoError(t, err)
				require.NotNil(t, box)
				require.Greater(t, box.Width, float64(0))
				require.Greater(t, box.Height, float64(0))

				value, err := page.Evaluate(`() => getComputedStyle(document.documentElement)
    .getPropertyValue("--color-surface").trim()`)
				require.NoError(t, err)
				surfaceToken, ok := value.(string)
				require.True(t, ok)
				require.NotEmpty(t, surfaceToken)

				waitForPageSettled(t, page)
				failures.RequireEmpty(t)
			})
		}
	}
}

func hasClass(className, target string) bool {
	for _, name := range strings.Fields(className) {
		if name == target {
			return true
		}
	}
	return false
}

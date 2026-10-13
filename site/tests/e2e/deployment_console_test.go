//go:build e2e && (full || example_deployments)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestDeploymentConsoleFlow(t *testing.T) {
	page := newPage(t, sharedBrowser)
	dismissCookieBanner(t, page)
	page.OnPageError(func(err error) { t.Errorf("browser error: %v", err) })
	_, err := page.Goto(baseURL + "/examples")
	require.NoError(t, err)
	_, err = page.Evaluate("window.deploymentNavigationSentinel = true")
	require.NoError(t, err)
	require.NoError(t, page.Locator("#componentdocshell-sidebar-content a[href='/examples/deployments']").Click())
	require.NoError(t, page.Locator("#deployment-console").WaitFor())
	require.NoError(t, page.Locator("#componentdocshell-sidebar-content a[href='/examples/deployments'][aria-current='page']").WaitFor())
	require.NoError(t, page.Locator("#deployment-navigation").WaitFor())
	assertDeploymentFragmentNavigation(t, page)
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "New deployment", Exact: playwright.Bool(true)}).First().Click())
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Create deployment", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.GetByText("Enter a service name between 2 and 60 characters.", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())
	require.NoError(t, page.Locator("#deployment-service").Fill("billing-api"))
	require.NoError(t, page.Locator("#deployment-version").Fill("v9.1.0"))
	require.NoError(t, page.Locator("#deployment-environment-trigger").Click())
	require.NoError(t, page.GetByRole("option", playwright.PageGetByRoleOptions{Name: "Production", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Create deployment", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: "billing-api", Exact: playwright.Bool(true)}).WaitFor())
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Approve deployment", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.GetByText("Deployment approved", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())
	assertDeploymentFragmentNavigation(t, page)
	_, err = page.Reload()
	require.NoError(t, err)
	require.NoError(t, page.GetByText("Deployment approved", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())
	require.NoError(t, page.Locator("#componentdocshell-sidebar-content").WaitFor())
	_, err = page.Evaluate("window.deploymentNavigationSentinel = true")
	require.NoError(t, err)
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Back to deployments", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.Locator("#deployment-search").Fill("billing"))
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Search", Exact: playwright.Bool(true)}).Click())
	_, err = page.WaitForFunction("() => document.querySelectorAll('#deployment-table tbody tr').length === 1", nil)
	require.NoError(t, err)
	count, err := page.Locator("#deployment-table tbody tr").Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.NoError(t, page.Locator("#deployment-search").Fill("no-such-service"))
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Search", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.GetByText("No matching deployments", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Clear search", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.SetViewportSize(390, 844))
	require.NoError(t, page.Locator("#deployment-navigation a").Filter(playwright.LocatorFilterOptions{HasText: "Overview"}).Click())
	require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: "Operations overview", Exact: playwright.Bool(true)}).WaitFor())
	fits, err := page.Evaluate("document.body.scrollWidth <= document.body.clientWidth")
	require.NoError(t, err)
	require.Equal(t, true, fits)
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Back to examples", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.Locator("#examples-overview").WaitFor())
	assertDeploymentFragmentNavigation(t, page)
}

func TestDeploymentConsoleAppearanceAndHistory(t *testing.T) {
	for _, mode := range []struct {
		name, theme string
		dark        bool
	}{
		{"GoshtosoLight", "goshtoso", false},
		{"GoshtosoDark", "goshtoso", true},
		{"MinimalLight", "minimal", false},
		{"MinimalDark", "minimal", true},
	} {
		t.Run(mode.name, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			dismissCookieBanner(t, page)
			require.NoError(t, page.SetViewportSize(1175, 1109))
			page.OnPageError(func(err error) { t.Errorf("browser error: %v", err) })
			_, err := page.Goto(baseURL + "/examples")
			require.NoError(t, err)
			require.NoError(t, waitForAlpine(page))
			_, err = page.Evaluate(`([theme, dark]) => {
				document.documentElement.dataset.theme = theme;
				document.documentElement.classList.toggle('dark', dark);
				window.deploymentNavigationSentinel = true;
			}`, []any{mode.theme, mode.dark})
			require.NoError(t, err)
			require.NoError(t, page.Locator("#componentdocshell-sidebar-content a[href='/examples/deployments']").Click())
			require.NoError(t, page.Locator("#deployment-console").WaitFor())
			assertDeploymentFragmentNavigation(t, page)
			takeScreenshot(t, page, "deployment-fragment-"+mode.name)
			require.NoError(t, page.Locator("#deployment-navigation a").Filter(playwright.LocatorFilterOptions{HasText: "New deployment"}).Click())
			require.NoError(t, page.Locator("#deployment-service").WaitFor())
			_, err = page.GoBack()
			require.NoError(t, err)
			require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: "Operations overview", Exact: playwright.Bool(true)}).WaitFor())
			_, err = page.GoForward()
			require.NoError(t, err)
			require.NoError(t, page.Locator("#deployment-service").WaitFor())
			assertDeploymentFragmentNavigation(t, page)
			require.NoError(t, page.SetViewportSize(390, 844))
			menu := page.Locator("button[aria-controls='componentdocshell-sidebar']")
			expanded, err := menu.Evaluate("el => el.getAttribute('aria-expanded')", nil)
			require.NoError(t, err)
			if expanded == "true" {
				require.NoError(t, menu.Click())
			}
			_, err = page.WaitForFunction("() => document.querySelector('#componentdocshell-sidebar').getBoundingClientRect().right <= 1", nil)
			require.NoError(t, err)
			fits, err := page.Locator("#deployment-example").Evaluate("el => el.scrollWidth <= el.clientWidth", nil)
			require.NoError(t, err)
			require.Equal(t, true, fits, "the embedded console must fit the mobile content region")
			takeScreenshot(t, page, "deployment-fragment-mobile-"+mode.name)
		})
	}
}

func assertDeploymentFragmentNavigation(t *testing.T, page playwright.Page) {
	t.Helper()
	preserved, err := page.Evaluate("window.deploymentNavigationSentinel === true")
	require.NoError(t, err)
	require.Equal(t, true, preserved, "navigation must preserve the current document")
}

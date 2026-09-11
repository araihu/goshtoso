//go:build e2e && full

package e2e

import (
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeploymentConsoleFlow(t *testing.T) {
	page := newPage(t, sharedBrowser)
	page.OnPageError(func(err error) { t.Errorf("browser error: %v", err) })
	_, err := page.Goto(baseURL + "/examples")
	require.NoError(t, err)
	require.NoError(t, page.Locator("#componentdocshell-sidebar-content a[href='/examples/deployments']").Click())
	require.NoError(t, page.Locator("#deployment-console").WaitFor())
	require.NoError(t, page.Locator(".console-shell__header").WaitFor())
	require.NoError(t, page.Locator("#consoleshell-sidebar").WaitFor())
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
	_, err = page.Reload()
	require.NoError(t, err)
	require.NoError(t, page.GetByText("Deployment approved", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Back to deployments", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.Locator("#deployment-search").Fill("billing"))
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Search", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.Locator("#deployment-table").WaitFor())
	count, err := page.Locator("#deployment-table tbody tr").Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.NoError(t, page.Locator("#deployment-search").Fill("no-such-service"))
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Search", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.GetByText("No matching deployments", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Clear search", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.SetViewportSize(390, 844))
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Open navigation", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.Locator("#consoleshell-sidebar a").Filter(playwright.LocatorFilterOptions{HasText: "Overview"}).Click())
	require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: "Operations overview", Exact: playwright.Bool(true)}).WaitFor())
	fits, err := page.Evaluate("document.body.scrollWidth <= document.body.clientWidth")
	require.NoError(t, err)
	require.Equal(t, true, fits)
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Back to examples", Exact: playwright.Bool(true)}).Click())
	require.NoError(t, page.Locator("#examples-overview").WaitFor())
}

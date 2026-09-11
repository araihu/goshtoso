//go:build e2e && full

package e2e

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExamplesOverviewNavigation(t *testing.T) {
	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	defer failures.RequireEmpty(t)
	_, err := page.Goto(baseURL + "/docs/application-patterns")
	require.NoError(t, err)
	require.NoError(t, page.WaitForURL("**/examples"))
	require.NoError(t, page.Locator("#examples-overview").WaitFor())
	count, err := page.Locator("a[href='/docs/application-patterns']").Count()
	require.NoError(t, err)
	require.Zero(t, count)
	require.NoError(t, page.Locator("#goshtoso-site-secondary-navigation a[href='/examples'][aria-current='location']").WaitFor())
	for _, example := range []struct{ path, marker string }{
		{"expense", "#expense-fragment"}, {"profile", "#profile-fragment"}, {"wizard", "#wizard-app"},
	} {
		require.NoError(t, page.Locator("#componentdocshell-sidebar-content a[href='/examples/"+example.path+"']").Click())
		require.NoError(t, page.Locator(example.marker).WaitFor())
		require.NoError(t, page.Locator("#goshtoso-site-secondary-navigation a[href='/examples']").Click())
		require.NoError(t, page.Locator("#examples-overview").WaitFor())
	}
	require.NoError(t, page.SetViewportSize(390, 844))
	fits, err := page.Evaluate("document.body.scrollWidth <= document.body.clientWidth")
	require.NoError(t, err)
	require.Equal(t, true, fits)
}

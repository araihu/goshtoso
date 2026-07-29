package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestActionGroupFragmentNavigationUsesBundledProvider(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: new("document.cookie='gt_storage=allowed; Path=/; SameSite=Lax'"),
	}))
	failures := watchPageFailures(page)

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("a[href='/components/action-group']").First().Click())
	require.NoError(t, page.Locator("#action-group-responsive").WaitFor())
	require.NoError(t, page.Locator("#action-group-publish").Click())
	_, err = page.WaitForFunction(`() =>
		document.querySelector("#action-group-responsive [x-text='lastAction']")?.textContent === "publish"
	`, nil)
	require.NoError(t, err, "bundled actionGroupDemo provider must initialise after HTMX fragment navigation")

	waitForPageSettled(t, page)
	failures.RequireEmpty(t)
}

//go:build e2e && (full || iconpack)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIconpackDocumentationPageBrowserProof(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	t.Cleanup(func() {
		waitForPageSettled(t, page)
		failures.RequireEmpty(t)
	})

	_, err := page.Goto(baseURL+"/docs/iconpack", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#iconpack-fragment").WaitFor())
	require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{
		Name:  "Consumer-local icon packs",
		Exact: playwright.Bool(true),
	}).WaitFor())

	body, err := page.Locator("body").TextContent()
	require.NoError(t, err)
	for _, expected := range []string{
		"Generate from an Assets release",
		"Serve the sprite and render through core",
		"IconBrandDeveloperIconsTRPC",
		"appicons.Lookup",
		"manifest.json",
	} {
		assert.Contains(t, body, expected)
	}

	link := page.GetByRole("link", playwright.PageGetByRoleOptions{
		Name:  "Open the Icon component workbench",
		Exact: playwright.Bool(true),
	})
	require.NoError(t, link.WaitFor())
	assert.Equal(t, "/components/icon", mustAttribute(t, link, "href"))

	sidebarLinks := page.Locator("a[href='/docs/iconpack']")
	count, err := sidebarLinks.Count()
	require.NoError(t, err)
	require.GreaterOrEqual(t, count, 1)
	visible := false
	for index := 0; index < count; index++ {
		candidate := sidebarLinks.Nth(index)
		isVisible, err := candidate.IsVisible()
		require.NoError(t, err)
		if isVisible {
			visible = true
			assert.Equal(t, "/docs/iconpack", mustAttribute(t, candidate, "href"))
		}
	}
	assert.True(t, visible, "Icon Packs navigation link must be visible in the active docs layout")
}

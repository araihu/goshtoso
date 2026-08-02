//go:build e2e && (full || head)

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependenciesDemoPage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/dependencies", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	title, err := page.Locator("main h1").TextContent()
	require.NoError(t, err)
	assert.Contains(t, title, "Dependencies")

	linkCount, err := page.Locator("#dependencies-full code").Filter(playwright.LocatorFilterOptions{
		HasText: "head.Dependencies()",
	}).Count()
	require.NoError(t, err)
	assert.Equal(t, 1, linkCount)

	require.NoError(t, page.Locator("#dependencies-full").GetByText("Fallback: /assets/js/runtime/*").WaitFor())
	require.NoError(t, page.Locator("#dependencies-minimal").GetByText("DependenciesMinimal").WaitFor())
	require.NoError(t, page.Locator("#dependencies-asset-contract").GetByText("assets.Handler()").WaitFor())
	require.NoError(t, page.Locator("#dependencies-options").GetByText("Strong defaults, explicit escape hatches").WaitFor())
	require.NoError(t, page.Locator("#dependencies-manifest").GetByText("One typed, ordered baseline").WaitFor())
	require.NoError(t, page.GetByText("not arbitrary versions").WaitFor())
}

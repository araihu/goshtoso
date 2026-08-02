//go:build e2e && full

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestModulesSidebarAndAppShellsShowcase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/modules/app-shells", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	require.NoError(t, page.Locator(`a[href="/modules/app-shells"][aria-current="page"]`).WaitFor())
	require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: "This documentation site uses Component Docs Shell"}).WaitFor())
	require.Equal(t, 1, mustCount(t, page.Locator(`a[href="/modules/charts"]`)))
	require.Equal(t, 0, mustCount(t, page.Locator(`a[href="/components/app-shell"]`)))

	moduleHeading := page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: "Modules", Exact: playwright.Bool(true)})
	examplesHeading := page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: "Examples", Exact: playwright.Bool(true)})
	moduleBox, err := moduleHeading.BoundingBox()
	require.NoError(t, err)
	examplesBox, err := examplesHeading.BoundingBox()
	require.NoError(t, err)
	require.Less(t, moduleBox.Y, examplesBox.Y, "Modules must precede Examples in the sidebar")
}

func TestChartsModuleLazyLoadsStaticInteractiveAndThreeD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	_, err := page.Goto(baseURL+"/modules/charts", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	require.NoError(t, page.Locator(`a[href="/modules/charts"][aria-current="page"]`).WaitFor())

	for _, kind := range []string{"static", "interactive", "line-3d"} {
		section := page.Locator(`[data-chart-module-showcase="` + kind + `"]`)
		require.NoError(t, section.ScrollIntoViewIfNeeded())
		frameElement := page.Locator(`iframe[data-charts-showcase-frame="` + kind + `"]`)
		require.NoError(t, frameElement.WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(10_000)}))
		frame := page.FrameLocator(`iframe[data-charts-showcase-frame="` + kind + `"]`)
		require.NoError(t, frame.Locator(`[data-charts-showcase-kind="`+kind+`"]`).WaitFor())
		if kind == "static" {
			require.NoError(t, frame.Locator("figure svg").WaitFor())
		} else {
			require.NoError(t, frame.Locator("canvas").First().WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(15_000)}))
		}
	}

	failures.RequireEmpty(t)
}

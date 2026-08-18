//go:build e2e && full

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
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
	require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: "Goshtoso App Shells"}).WaitFor())
	require.NoError(t, page.Locator(`[data-sidebar-section="Frames"]`).WaitFor())
	require.NoError(t, page.Locator(`[data-sidebar-section="Shells"]`).WaitFor())
	require.NoError(t, page.Locator(`a[href="/modules/app-shells/frames/component-page"]`).WaitFor())
	require.NoError(t, page.Locator(`a[href="/modules/app-shells/shells/component-docs-shell"]`).WaitFor())
	require.Equal(t, 1, mustCount(t, page.Locator(`a[href="/modules/charts"]`)))
	require.Equal(t, 0, mustCount(t, page.Locator(`a[href="/components/app-shell"]`)))

	framesBox, err := page.Locator(`[data-sidebar-section="Frames"]`).BoundingBox()
	require.NoError(t, err)
	shellsBox, err := page.Locator(`[data-sidebar-section="Shells"]`).BoundingBox()
	require.NoError(t, err)
	require.Less(t, framesBox.Y, shellsBox.Y, "Frames must precede Shells in the sidebar")
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

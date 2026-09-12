//go:build e2e && full

package e2e

import (
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCoreDocsContainsAgentsAndIcons(t *testing.T) {
	_, browser, cleanup := setupPlaywright(t)
	defer cleanup()
	page := newPage(t, browser)
	_, err := page.Goto(baseURL + "/docs/agents")
	require.NoError(t, err)
	for _, path := range []string{"/components/icon", "/docs/icon-catalog", "/docs/iconpack", "/docs/agents"} {
		require.NoError(t, page.Locator(`nav[aria-label='sidebar navigation'] a[href='`+path+`']`).Click())
		require.NoError(t, page.WaitForURL("**"+path))
		core := page.Locator(`#goshtoso-site-secondary-navigation a[data-site-secondary-family='core']`)
		value, err := core.GetAttribute("aria-current")
		require.NoError(t, err)
		require.Equal(t, "location", value)
		count, err := page.Locator(`#goshtoso-site-secondary-navigation a[href='/docs/agents'], #goshtoso-site-secondary-navigation a[href='/components/icon']`).Count()
		require.NoError(t, err)
		require.Zero(t, count)
		active := page.Locator(`nav[aria-label='sidebar navigation'] a[href='` + path + `']`)
		require.NoError(t, active.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
		current, err := active.GetAttribute("aria-current")
		require.NoError(t, err)
		require.Equal(t, "page", current)
	}
}

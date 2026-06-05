package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDrawerDemoPage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/drawer", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	heading := page.Locator("main h1")
	text, err := heading.TextContent()
	require.NoError(t, err)
	assert.Contains(t, text, "Drawer")

	dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{
		Name: "Project details",
	})
	visible, err := dialog.IsVisible()
	require.NoError(t, err)
	assert.False(t, visible)

	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Open details",
	}).Click())
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, dialog.GetByText("Deployment target").WaitFor())

	require.NoError(t, dialog.GetByRole("button", playwright.LocatorGetByRoleOptions{
		Name: "Close",
	}).Click())
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden,
	}))
}

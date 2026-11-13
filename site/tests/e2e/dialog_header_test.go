//go:build e2e && (full || modal)

package e2e

import (
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDialogCustomHeaderLifecycle(t *testing.T) {
	page := newIsolatedPage(t)
	_, err := page.Goto(baseURL + "/components/modal")
	require.NoError(t, err)
	trigger := page.Locator("[data-custom-header-open]")
	require.NoError(t, trigger.Click())
	dialog := page.Locator("#assetDetails")
	require.NoError(t, dialog.WaitFor())
	count, err := dialog.GetByText("Selected asset").Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
	require.NoError(t, page.SetViewportSize(390, 844))
	require.NoError(t, trigger.Click())
	require.NoError(t, dialog.WaitFor())
	require.NoError(t, dialog.GetByText("Done").Click())
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
}

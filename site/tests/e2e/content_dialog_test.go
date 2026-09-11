//go:build e2e && (full || modal)

package e2e

import (
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContentDialogAboveDrawer(t *testing.T) {
	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	_, err := page.Goto(baseURL + "/components/modal")
	require.NoError(t, err)
	_, err = page.WaitForFunction("()=>window.Alpine && window.htmx", nil)
	require.NoError(t, err)
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Open content editor", Exact: playwright.Bool(true)}).Click())
	draft := page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Draft", Exact: playwright.Bool(true)})
	require.NoError(t, draft.Fill("Changed draft"))
	trigger := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Choose item", Exact: playwright.Bool(true)})
	for _, width := range []int{390, 1440} {
		require.NoError(t, page.SetViewportSize(width, 994))
		require.NoError(t, trigger.Click())
		dialog := page.Locator("#contentPicker")
		require.NoError(t, dialog.WaitFor())
		fits, err := dialog.Evaluate(`e=>{const r=e.getBoundingClientRect();return e.matches(':modal') && r.left>=0 && r.right<=innerWidth && r.top>=0 && r.bottom<=innerHeight && (innerWidth>=640 ? r.height < innerHeight * 0.8 : (r.width===innerWidth && r.height===innerHeight));}`, nil)
		require.NoError(t, err)
		require.Equal(t, true, fits)
		require.NoError(t, page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Search items"}).Fill("query"))
		require.NoError(t, page.Keyboard().Press("Escape"))
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
		visible, err := draft.IsVisible()
		require.NoError(t, err)
		require.True(t, visible)
		value, err := draft.InputValue()
		require.NoError(t, err)
		require.Equal(t, "Changed draft", value)
		focused, err := trigger.Evaluate("e=>document.activeElement===e", nil)
		require.NoError(t, err)
		require.Equal(t, true, focused)
		require.NoError(t, trigger.Click())
		require.NoError(t, page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Search items"}).Fill("Release"))
		require.NoError(t, page.GetByRole("radio", playwright.PageGetByRoleOptions{Name: "Release notes", Exact: playwright.Bool(false)}).Check())
		require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Select item", Exact: playwright.Bool(true)}).Click())
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
		selected, err := page.Locator("#modal-content-dialog").Evaluate("e => Alpine.$data(e).selectedItem", nil)
		require.NoError(t, err)
		require.Equal(t, "Release notes", selected)
	}
	failures.RequireEmpty(t)
}

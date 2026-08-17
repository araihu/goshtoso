//go:build e2e && (full || splitbutton)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitButton_Primitives(t *testing.T) {
	page := newPage(t, sharedBrowser)
	defer func() { _ = page.Close() }()

	_, err := page.Goto(baseURL+"/components/splitbutton", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	root := page.Locator("#splitbutton-demo")
	primary := root.Locator("a").First()
	trigger := page.Locator("#splitbutton-demo-menu button[aria-haspopup='menu']")
	panel := page.Locator("#splitbutton-demo-menu [data-popover-panel]")

	require.NoError(t, primary.WaitFor())
	require.NoError(t, trigger.WaitFor())
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}))

	primaryHref, err := primary.GetAttribute("href")
	require.NoError(t, err)
	assert.Equal(t, "/components/splitbutton", primaryHref)

	visible, err := panel.IsVisible()
	require.NoError(t, err)
	assert.False(t, visible, "split button menu should be hidden before activation")

	require.NoError(t, trigger.Click())
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	caption := panel.GetByText("Review the draft", playwright.LocatorGetByTextOptions{Exact: playwright.Bool(true)})
	require.NoError(t, caption.WaitFor())

	newTab := panel.Locator("a[target='_blank'][rel='noopener noreferrer']")
	assert.Equal(t, 1, mustCount(t, newTab))

	menuItems := panel.Locator("[role='menuitem']")
	assert.Equal(t, 3, mustCount(t, menuItems))

	expanded, err := trigger.GetAttribute("aria-expanded")
	require.NoError(t, err)
	assert.Equal(t, "true", expanded)

	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden,
	}))
	require.NoError(t, trigger.Focus())
	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	_, err = page.WaitForFunction(
		`() => {
			const first = document.querySelector("#splitbutton-demo-menu [role='menuitem']");
			return first && document.activeElement === first;
		}`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
	)
	require.NoError(t, err)
	activeTag, err := page.Locator("#splitbutton-demo-menu [role='menuitem']").First().Evaluate("el => el === document.activeElement", nil)
	require.NoError(t, err)
	assert.True(t, activeTag.(bool), "keyboard activation should focus the first menu item")
}

//go:build e2e && (full || popover)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopover_Primitives(t *testing.T) {
	page := newPage(t, sharedBrowser)
	defer func() { _ = page.Close() }()

	_, err := page.Goto(baseURL+"/components/popover", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	trigger := page.Locator("#filters-popover button")
	panel := page.Locator("#filters-popover [data-popover-panel]")

	require.NoError(t, trigger.WaitFor())
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}))

	visible, err := panel.IsVisible()
	require.NoError(t, err)
	assert.False(t, visible, "popover panel should be hidden before activation")

	role, err := panel.GetAttribute("role")
	require.NoError(t, err)
	assert.Equal(t, "dialog", role)

	label, err := panel.GetAttribute("aria-label")
	require.NoError(t, err)
	assert.Equal(t, "Filters", label)

	require.NoError(t, trigger.Click())
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	expanded, err := trigger.GetAttribute("aria-expanded")
	require.NoError(t, err)
	assert.Equal(t, "true", expanded)

	content, err := panel.TextContent()
	require.NoError(t, err)
	assert.Contains(t, content, "Filter results")

	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden,
	}))
}

func TestPopover_NestedEscapeClosesInnermostFirst(t *testing.T) {
	page := newPage(t, sharedBrowser)
	defer func() { _ = page.Close() }()
	_, err := page.Goto(baseURL + "/components/popover")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	outer := page.Locator("#nested-inbox-trigger")
	inner := page.Locator("#nested-inbox-actions [data-popover-trigger] button")
	require.NoError(t, outer.Click())
	require.NoError(t, inner.Click())
	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, page.Locator("#nested-inbox-actions-panel").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
	visible, err := page.Locator("#nested-inbox-panel").IsVisible()
	require.NoError(t, err)
	assert.True(t, visible, "inner Escape must leave outer dialog open")
	_, err = page.WaitForFunction(`() => document.activeElement === document.querySelector('#nested-inbox-actions [data-popover-trigger] button')`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, page.Locator("#nested-inbox-panel").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
	_, err = page.WaitForFunction(`() => document.activeElement.id === 'nested-inbox-trigger'`, nil)
	require.NoError(t, err)
}

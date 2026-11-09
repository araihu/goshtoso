//go:build e2e && (full || popover)

package e2e

import (
	"fmt"
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

func TestPopover_SidebarTriggerOwnership(t *testing.T) {
	for _, theme := range []string{"goshtoso", "minimal"} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/dark=%t", theme, dark), func(t *testing.T) {
				page := newPage(t, sharedBrowser)
				defer func() { _ = page.Close() }()
				_, err := page.Goto(baseURL + "/components/popover")
				require.NoError(t, err)
				require.NoError(t, waitForAlpine(page))
				_, err = page.Evaluate(`({theme, dark}) => {
                    document.documentElement.dataset.theme = theme;
                    document.documentElement.classList.toggle('dark', dark);
                }`, map[string]any{"theme": theme, "dark": dark})
				require.NoError(t, err)
				navigation := page.Locator("#sidebar-popover-example > div > button")
				account := page.Locator("#sidebar-account-trigger")
				accountPanel := page.Locator("#sidebar-account-panel")
				navigationPanel := page.Locator("#account-navigation-panel")
				assertOwnership := func() {
					t.Helper()
					_, err := page.WaitForFunction(`() => {
                        const nav = document.querySelector('#sidebar-popover-example > div > button');
                        const account = document.querySelector('#sidebar-account-trigger');
                        const state = Alpine.$data(document.querySelector('#sidebar-account'));
                        const inherited = {$refs: {trigger: nav, panel: document.querySelector('#account-navigation-panel')}};
                        return state.triggerElement.call(inherited) === account &&
                            state.panelElement.call(inherited) === document.querySelector('#sidebar-account-panel') &&
                            nav.getAttribute('aria-controls') === 'account-navigation-panel' &&
                            !nav.hasAttribute('aria-haspopup') &&
                            account.getAttribute('aria-controls') === 'sidebar-account-panel' &&
                            account.getAttribute('aria-haspopup') === 'dialog';
                    }`, nil)
					require.NoError(t, err)
				}
				assertOwnership()
				require.NoError(t, navigation.Click())
				require.NoError(t, navigationPanel.WaitFor())
				for pass := range 2 {
					assertOwnership()
					require.NoError(t, account.Focus())
					require.NoError(t, page.Keyboard().Press("Enter"))
					require.NoError(t, accountPanel.WaitFor())
					require.NoError(t, page.Keyboard().Press("Tab"))
					_, err = page.WaitForFunction(`() => document.activeElement.id === 'sidebar-account-close'`, nil)
					require.NoError(t, err)
					require.NoError(t, page.Keyboard().Press("Enter"))
					require.NoError(t, accountPanel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
					_, err = page.WaitForFunction(`() => document.activeElement.id === 'sidebar-account-trigger'`, nil)
					require.NoError(t, err)
					assertOwnership()
					if pass == 0 {
						// A fragment remount must not bind the new popover to its ancestor's ref.
						_, err = page.Evaluate(`() => {
                            const root = document.querySelector('#sidebar-account');
                            root.replaceWith(root.cloneNode(true));
                        }`)
						require.NoError(t, err)
					}
				}
				require.NoError(t, page.Keyboard().Press("Escape"))
				require.NoError(t, navigationPanel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
				assertOwnership()
			})
		}
	}
}

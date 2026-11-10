//go:build e2e && (full || tooltip)

package e2e

import (
	"fmt"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestTooltip_PortalDismissesBeforeActionOpensDrawer(t *testing.T) {
	for _, width := range []int{390, 1117} {
		for _, theme := range []string{"goshtoso", "minimal"} {
			for _, dark := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%s/dark=%t", width, theme, dark), func(t *testing.T) {
					page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
						Viewport: &playwright.Size{Width: width, Height: 900},
					})
					_, err := page.Goto(baseURL + "/components/tooltip")
					require.NoError(t, err)
					require.NoError(t, waitForAlpine(page))
					_, err = page.Evaluate(`config => {
  document.documentElement.dataset.theme=config.theme;
  document.documentElement.classList.toggle('dark',config.dark);
  window.drawerTooltipOpen=null;
  window.addEventListener('drawer:open', event => {
    if(event.detail?.id==='tooltip-details')
      drawerTooltipOpen=document.querySelector('#details-tip').matches(':popover-open');
  });
}`, map[string]any{"theme": theme, "dark": dark})
					require.NoError(t, err)
					for _, action := range []string{"click", "Enter", "Space"} {
						t.Run(action, func(t *testing.T) {
							trigger := page.Locator("#tooltip-action-trigger")
							require.NoError(t, trigger.Hover())
							require.NoError(t, trigger.Focus())
							_, err := page.WaitForFunction(`()=>document.querySelector('#details-tip').matches(':popover-open')`, nil)
							require.NoError(t, err)
							if action == "click" {
								require.NoError(t, trigger.Click())
							} else {
								require.NoError(t, trigger.Press(action))
							}
							require.NoError(t, page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "Details", Exact: playwright.Bool(true)}).WaitFor())
							open, err := page.Evaluate(`()=>drawerTooltipOpen`)
							require.NoError(t, err)
							require.Equal(t, false, open, "dismiss before the trigger's propagation-stopping action runs")
							_, err = page.WaitForFunction(`()=>!document.querySelector('#details-tip').matches(':popover-open')`, nil)
							require.NoError(t, err)
							require.NoError(t, page.Keyboard().Press("Escape"))
							require.NoError(t, page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "Details", Exact: playwright.Bool(true)}).WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
						})
					}
					t.Run("focus-transfer", func(t *testing.T) {
						trigger := page.Locator("#tooltip-action-trigger")
						require.NoError(t, trigger.Focus())
						require.NoError(t, page.Locator("#details-tip").Hover())
						_, err := page.Evaluate(`()=>window.dispatchEvent(new CustomEvent('drawer:open',{detail:{id:'tooltip-details'}}))`)
						require.NoError(t, err)
						dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "Details", Exact: playwright.Bool(true)})
						require.NoError(t, dialog.WaitFor())
						require.NoError(t, dialog.GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Close", Exact: playwright.Bool(true)}).Focus())
						open, err := page.Locator("#details-tip").Evaluate(`tip=>tip.matches(':popover-open')`, nil)
						require.NoError(t, err)
						require.Equal(t, false, open, "moving focus into a drawer dismisses a hovered portal tooltip")
						require.NoError(t, page.Keyboard().Press("Escape"))
						require.NoError(t, page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "Details", Exact: playwright.Bool(true)}).WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
					})
				})
			}
		}
	}
}

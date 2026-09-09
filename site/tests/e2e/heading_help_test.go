//go:build e2e && (full || pageheader || table)

package e2e

import (
	"fmt"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHeadingHelpClickDismissal(t *testing.T) {
	for _, fixture := range []struct{ route, label, id string }{
		{"/components/page-header", "About staging", "staging-help"},
		{"/components/table", "About services", "service-help"},
	} {
		t.Run(fixture.id, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			failures := watchPageFailures(page)
			_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})
			require.NoError(t, err)
			_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined' && typeof htmx !== 'undefined'", nil)
			require.NoError(t, err)
			_, err = page.Evaluate("() => window.__headingHelpNavigation = true", nil)
			require.NoError(t, err)
			link := page.Locator(fmt.Sprintf(`#componentdocshell-sidebar-content a[href=%q]`, fixture.route))
			clickUntil(t, page, link, fmt.Sprintf("!!document.getElementById(%q)", fixture.id))
			survived, err := page.Evaluate("() => window.__headingHelpNavigation === true", nil)
			require.NoError(t, err)
			require.Equal(t, true, survived, "navigation must preserve the document through HTMX")
			require.NoError(t, page.SetViewportSize(390, 900))
			button := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: fixture.label, Exact: playwright.Bool(true)})
			require.NoError(t, button.Focus())
			open, err := page.Locator("#"+fixture.id).Evaluate("el => el.matches(':popover-open')", nil)
			require.NoError(t, err)
			require.Equal(t, false, open, "focus alone must not open help")
			require.NoError(t, button.Press("Enter"))
			require.Eventually(t, func() bool {
				value, err := page.Locator("#"+fixture.id).Evaluate("el => el.matches(':popover-open') && getComputedStyle(el).opacity === '1'", nil)
				return err == nil && value == true
			}, 5*time.Second, 50*time.Millisecond)
			bounds, err := page.Locator("#"+fixture.id).Evaluate(`el => {
                const r = el.getBoundingClientRect();
                return el.matches(':popover-open') && getComputedStyle(el).translate === 'none' && r.left >= 0 && r.top >= 0 && r.right <= innerWidth && r.bottom <= innerHeight;
            }`, nil)
			require.NoError(t, err)
			debug, _ := page.Locator("#"+fixture.id).Evaluate(`el => ({open: el.matches(':popover-open'), rect: el.getBoundingClientRect().toJSON(), attrs: el.outerHTML.slice(0, 700)})`, nil)
			require.Equal(t, true, bounds, "help must escape table clipping and stay inside the viewport: %v", debug)
			require.NoError(t, button.Press("Escape"))
			open, err = page.Locator("#"+fixture.id).Evaluate("el => el.matches(':popover-open')", nil)
			require.NoError(t, err)
			require.Equal(t, false, open)
			require.NoError(t, button.Click())
			require.Eventually(t, func() bool {
				open, err := page.Locator("#"+fixture.id).Evaluate("el => el.matches(':popover-open')", nil)
				return err == nil && open == true
			}, time.Second, 20*time.Millisecond)
			require.NoError(t, page.Locator("body").Click(playwright.LocatorClickOptions{Position: &playwright.Position{X: 380, Y: 5}}))
			require.Eventually(t, func() bool {
				v, _ := page.Locator("#"+fixture.id).Evaluate("el => el.matches(':popover-open')", nil)
				return v == false
			}, time.Second, 20*time.Millisecond)
			require.NoError(t, button.Click())
			require.Eventually(t, func() bool {
				open, err := page.Locator("#"+fixture.id).Evaluate("el => el.matches(':popover-open')", nil)
				return err == nil && open == true
			}, time.Second, 20*time.Millisecond)
			require.NoError(t, button.Click())
			require.Eventually(t, func() bool {
				v, _ := page.Locator("#"+fixture.id).Evaluate("el => el.matches(':popover-open')", nil)
				return v == false
			}, time.Second, 20*time.Millisecond)
			failures.RequireEmpty(t)
		})
	}
}

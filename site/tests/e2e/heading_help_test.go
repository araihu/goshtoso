//go:build e2e && (full || pageheader || table)

package e2e

import (
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHeadingHelpSupportsKeyboardFocus(t *testing.T) {
	for _, fixture := range []struct{ route, label, id string }{
		{"/components/page-header", "About staging", "staging-help"},
		{"/components/table", "About services", "service-help"},
	} {
		t.Run(fixture.id, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			require.NoError(t, page.SetViewportSize(390, 900))
			_, err := page.Goto(baseURL+fixture.route, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})
			require.NoError(t, err)
			button := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: fixture.label, Exact: playwright.Bool(true)})
			require.NoError(t, button.Focus())
			require.Eventually(t, func() bool {
				value, err := page.Locator("#"+fixture.id).Evaluate("el => getComputedStyle(el).opacity", nil)
				return err == nil && value == "1"
			}, 5*time.Second, 50*time.Millisecond)
			bounds, err := page.Locator("#"+fixture.id).Evaluate(`el => {
                const r = el.getBoundingClientRect();
                return el.matches(':popover-open') && r.left >= 0 && r.top >= 0 && r.right <= innerWidth && r.bottom <= innerHeight;
            }`, nil)
			require.NoError(t, err)
			debug, _ := page.Locator("#"+fixture.id).Evaluate(`el => ({open: el.matches(':popover-open'), rect: el.getBoundingClientRect().toJSON(), attrs: el.outerHTML.slice(0, 700)})`, nil)
			require.Equal(t, true, bounds, "help must escape table clipping and stay inside the viewport: %v", debug)
			require.NoError(t, button.Press("Escape"))
			open, err := page.Locator("#"+fixture.id).Evaluate("el => el.matches(':popover-open')", nil)
			require.NoError(t, err)
			require.Equal(t, false, open)
		})
	}
}

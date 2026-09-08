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
			_, err := page.Goto(baseURL+fixture.route, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})
			require.NoError(t, err)
			button := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: fixture.label, Exact: playwright.Bool(true)})
			require.NoError(t, button.Focus())
			require.Eventually(t, func() bool {
				value, err := page.Locator("#"+fixture.id).Evaluate("el => getComputedStyle(el).opacity", nil)
				return err == nil && value == "1"
			}, 5*time.Second, 50*time.Millisecond)
		})
	}
}

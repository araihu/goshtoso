//go:build e2e && full

package e2e

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSelectTrailingBadgeFollowsValueAndKeyboard(t *testing.T) {
	page := newIsolatedPage(t)
	_, err := page.Goto(baseURL + "/components/select")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	for _, width := range []int{390, 1440} {
		require.NoError(t, page.SetViewportSize(width, 900))
		for _, choice := range []struct{ index, kind string }{{"0", "Built-in"}, {"1", "Custom"}} {
			require.NoError(t, page.Locator("#rich-theme-trigger").Click())
			require.NoError(t, page.Locator("#rich-theme-option-"+choice.index).Click())
			_, err = page.WaitForFunction(`kind=>[...document.querySelectorAll('#rich-theme-trigger .demo-theme-kind')].some(e=>e.getClientRects().length&&e.textContent.trim()===kind)`, choice.kind)
			require.NoError(t, err)
			valid, err := page.Evaluate(`()=>{
    const t=document.querySelector('#rich-theme-trigger'),l=t.querySelector('[x-text]'),b=[...t.querySelectorAll('.demo-theme-kind')].find(e=>e.getClientRects().length),r=t.getBoundingClientRect(),a=l.getBoundingClientRect(),c=b.getBoundingClientRect();
    return c.left>=a.right&&c.right<=r.right&&r.right<=innerWidth+1&&!l.textContent.includes('Built-in')&&!l.textContent.includes('Custom');
   }`, nil)
			require.NoError(t, err)
			require.Equal(t, true, valid)
		}
		require.NoError(t, page.Locator("#rich-theme-trigger").Click())
		require.NoError(t, page.Locator("#rich-theme-option-1").Focus())
		require.NoError(t, page.Locator("#rich-theme-option-1").Press("ArrowUp"))
		require.NoError(t, page.Locator("#rich-theme-option-0").Press("Enter"))
		_, err = page.WaitForFunction(`()=>document.querySelector('#rich-theme').value==='code'`, nil)
		require.NoError(t, err)
	}
}

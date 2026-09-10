//go:build e2e && full

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSidebarNavigationPreservesScrollThroughoutSwap(t *testing.T) {
	page := newPage(t, sharedBrowser)
	require.NoError(t, page.SetViewportSize(1162, 1240))
	_, err := page.Goto(baseURL + "/components/kbd")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	for _, route := range []string{"schema-form", "kbd", "schema-form"} {
		_, err = page.Evaluate(`() => {
   const sidebar = document.querySelector('.sidebar-scroll');
   sidebar.scrollTop = 800;
   window.sidebarProbe = {node: sidebar, start: sidebar.scrollTop, samples: [], done: false};
   const sample = () => {
    const current = document.querySelector('.sidebar-scroll');
    sidebarProbe.samples.push({same: current === sidebar, top: current.scrollTop});
   };
   document.addEventListener('htmx:after:swap', function after(event) {
    if (event.detail.ctx.target.id !== 'main-content') return;
    document.removeEventListener('htmx:after:swap', after, true);
    sample();
    requestAnimationFrame(() => { sample(); requestAnimationFrame(() => { sample(); sidebarProbe.done = true; }); });
   }, true);
  }`)
		require.NoError(t, err)
		require.NoError(t, page.Locator(`.sidebar-scroll a[href="/components/`+route+`"]`).Click())
		_, err = page.WaitForFunction(`() => window.sidebarProbe?.done`, nil)
		require.NoError(t, err)
		result, err := page.Evaluate(`() => ({scrolled: sidebarProbe.start > 0, stable: sidebarProbe.samples.every(s => s.same && Math.abs(s.top - sidebarProbe.start) <= 1), active: document.querySelector('.sidebar-scroll a[aria-current="page"]')?.getAttribute('href')})`)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"scrolled": true, "stable": true, "active": "/components/" + route}, result)
	}
}

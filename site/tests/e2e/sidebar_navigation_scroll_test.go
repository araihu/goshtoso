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

func TestSidebarRevealsActiveItemOnlyWhenOutsideViewport(t *testing.T) {
	page := newPage(t, sharedBrowser)
	require.NoError(t, page.SetViewportSize(1162, 900))
	_, err := page.Goto(baseURL + "/components/schema-form")
	require.NoError(t, err)
	visible := `() => {
  const sidebar = document.querySelector('.sidebar-scroll');
  const active = sidebar.querySelector('a[aria-current="page"]');
  if (!active) return false;
  const box = sidebar.getBoundingClientRect(), item = active.getBoundingClientRect();
  return item.top >= box.top - 1 && item.bottom <= box.bottom + 1;
 }`
	_, err = page.WaitForFunction(visible, nil)
	require.NoError(t, err, "direct load reveals the active item")
	_, err = page.Evaluate(`async () => {
  document.querySelector('.sidebar-scroll').scrollTop = 0;
  await htmx.ajax('GET', '/components/schema-form', {target: '#main-content', swap: 'innerHTML'});
 }`)
	require.NoError(t, err)
	_, err = page.WaitForFunction(visible, nil)
	require.NoError(t, err, "navigation reveals an offscreen active item")
	result, err := page.Evaluate(`async () => {
  const sidebar = document.querySelector('.sidebar-scroll');
  const before = sidebar.scrollTop;
  await htmx.ajax('GET', '/components/schema-form', {target: '#main-content', swap: 'innerHTML'});
  return sidebar.scrollTop === before;
 }`)
	require.NoError(t, err)
	require.Equal(t, true, result, "a visible active item must not move the sidebar")
}

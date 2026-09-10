//go:build e2e && full

package e2e

import (
	"bytes"
	"context"
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/scrollregion"
	"github.com/araihu/goshtoso/components/toast"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestModernTabsSingleRequestAndRetry(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	require.NoError(t, page.AddInitScript(playwright.Script{Content: new(`window.tabRequests=0; window.tabMode='error'; const fetchOriginal=window.fetch; window.fetch=function(input, init){ if(String(input).includes('/api/components/tab-content/details')) { window.tabRequests++; if(window.tabMode==='error') return Promise.resolve(new Response('Retry after selection',{status:503})); if(window.tabMode==='network') return Promise.reject(new TypeError('offline')); } return fetchOriginal(input,init); };`)}))
	_, err := page.Goto(baseURL + "/components/tabs")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	value, err := page.Evaluate(`() => window.tabRequests`, nil)
	require.NoError(t, err)
	require.EqualValues(t, 0, value)
	details := page.Locator("#tabs-htmx").GetByRole("tab", playwright.LocatorGetByRoleOptions{Name: "Details"})
	overview := page.Locator("#tabs-htmx [role=tab]").First()
	for _, mode := range []string{"error", "network", "success"} {
		_, err = page.Evaluate(`mode => window.tabMode=mode`, mode)
		require.NoError(t, err)
		if mode == "error" {
			_, err = page.Evaluate(`() => Alpine.$data(document.querySelector('#tabpanelhtmxdetails')).selectedTab='details'`, nil)
			require.NoError(t, err)
		} else {
			require.NoError(t, details.Click())
		}
		_, err = page.WaitForFunction(`() => !document.querySelector('#tabpanelhtmxdetails').dataset.loading`, nil)
		require.NoError(t, err)
		require.NoError(t, overview.Click())
	}
	value, err = page.Evaluate(`() => window.tabRequests`, nil)
	require.NoError(t, err)
	require.EqualValues(t, 3, value)
	require.NoError(t, details.Click())
	require.NoError(t, overview.Click())
	require.NoError(t, details.Click())
	value, err = page.Evaluate(`() => window.tabRequests`, nil)
	require.NoError(t, err)
	require.EqualValues(t, 3, value)
	content, err := page.Locator("#tabpanelhtmxdetails").TextContent()
	require.NoError(t, err)
	require.Contains(t, content, "Details (Lazy Loaded)")
}

func TestModernNavbarKeyboardOwnsARIAAndFocus(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	_, err := page.Goto(baseURL + "/components/navbar")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	trigger := page.Locator("button[aria-label='user menu']").First()
	for _, key := range []string{"Enter", "Space", "ArrowDown"} {
		require.NoError(t, trigger.Focus())
		require.NoError(t, trigger.Press(key))
		_, err = page.WaitForFunction(`() => { const t=document.querySelector('button[aria-label="user menu"]'); return t.getAttribute('aria-expanded')==='true' && t.parentElement.querySelector('[role=menu]').contains(document.activeElement); }`, nil)
		require.NoError(t, err)
		require.NoError(t, page.Keyboard().Press("Escape"))
		_, err = page.WaitForFunction(`() => { const t=document.querySelector('button[aria-label="user menu"]'); return t.getAttribute('aria-expanded')==='false' && document.activeElement===t; }`, nil)
		require.NoError(t, err)
	}
}

func TestModernEnhancerRemovalDisposesObservers(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	for _, spec := range []struct{ route, selector string }{{"/components/action-group", "[data-goshtoso-action-group]"}, {"/components/tabs", "[data-goshtoso-scroll-region]"}} {
		t.Run(spec.route, func(t *testing.T) {
			page := newPage(t, browser)
			require.NoError(t, page.AddInitScript(playwright.Script{Content: new(`window.resizeOwners=[]; const Original=ResizeObserver; window.ResizeObserver=class extends Original { constructor(cb){super(cb); this.targets=new Set();window.resizeOwners.push(this);} observe(el){this.targets.add(el);super.observe(el);} unobserve(el){this.targets.delete(el);super.unobserve(el);} disconnect(){this.targets.clear();super.disconnect();} };`)}))
			_, err := page.Goto(baseURL + spec.route)
			require.NoError(t, err)
			require.NoError(t, waitForAlpine(page))
			if spec.selector == "[data-goshtoso-scroll-region]" {
				var markup bytes.Buffer
				require.NoError(t, scrollregion.ScrollRegion(scrollregion.Config{Content: templ.Raw(`<p>Measured content</p>`)}).Render(context.Background(), &markup))
				_, err = page.Evaluate(`html => document.body.insertAdjacentHTML('beforeend',html)`, markup.String())
				require.NoError(t, err)
			}
			require.NoError(t, page.Locator(spec.selector).First().WaitFor())
			_, err = page.Evaluate(`selector => {const root=document.querySelector(selector);window.removedRoot=root;window.ownedObservers=resizeOwners.filter(o=>Array.from(o.targets).some(el=>el===root||root.contains(el)));root.remove();}`, spec.selector)
			require.NoError(t, err)
			_, err = page.WaitForFunction(`() => ownedObservers.length>0 && ownedObservers.every(o=>o.targets.size===0)`, nil)
			require.NoError(t, err)
			_, err = page.Evaluate(`() => { const host=document.createElement('div'); host.setAttribute('x-data','{shown:true}'); host.innerHTML='<template x-if="shown">'+removedRoot.outerHTML+'</template>'; window.regionHost=host; document.body.append(host); }`, nil)
			require.NoError(t, err)
			_, err = page.WaitForFunction(`selector => {const root=regionHost.querySelector(selector);return root && resizeOwners.some(o=>Array.from(o.targets).some(el=>el===root||root.contains(el)));}`, spec.selector)
			require.NoError(t, err)
			_, err = page.Evaluate(`selector => {const root=regionHost.querySelector(selector);window.ownedObservers=resizeOwners.filter(o=>Array.from(o.targets).some(el=>el===root||root.contains(el)));Alpine.$data(regionHost).shown=false;}`, spec.selector)
			require.NoError(t, err)
			_, err = page.WaitForFunction(`() => ownedObservers.length>0 && ownedObservers.every(o=>o.targets.size===0)`, nil)
			require.NoError(t, err)
		})
	}
}

func TestModernToastBurstAndIdempotentDismiss(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	for _, reduced := range []*playwright.ReducedMotion{playwright.ReducedMotionReduce, playwright.ReducedMotionNoPreference} {
		t.Run(string(*reduced), func(t *testing.T) {
			page := newPage(t, browser)
			require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: reduced}))
			_, err := page.Goto(baseURL + "/components/toast")
			require.NoError(t, err)
			require.NoError(t, waitForAlpine(page))
			_, err = page.Evaluate(`() => {for(let i=0;i<25;i++) window.dispatchEvent(new CustomEvent('notify',{detail:{tone:'info',title:'Burst '+i}}));}`, nil)
			require.NoError(t, err)
			_, err = page.WaitForFunction(`() => document.querySelectorAll('#toast-container [role=alert]').length===20`, nil)
			require.NoError(t, err)
			_, err = page.Evaluate(`() => {const el=document.querySelector('#toast-container [role=alert]');window.closedToast=el;Alpine.$data(el).dismiss();Alpine.$data(el).dismiss();}`, nil)
			require.NoError(t, err)
			_, err = page.WaitForFunction(`() => !closedToast.isConnected && document.querySelectorAll('#toast-container [role=alert]').length===19`, nil)
			require.NoError(t, err)
		})
	}
}

func TestModernToastTimersAreContainerScopedAndDisposable(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	require.NoError(t, page.AddInitScript(playwright.Script{Content: new(`window.toastTimers=new Set();const realSetTimeout=window.setTimeout;const realClearTimeout=window.clearTimeout;window.setTimeout=function(callback,delay,...args){const id=realSetTimeout(callback,delay,...args);if(delay===12345)toastTimers.add(id);return id;};window.clearTimeout=function(id){toastTimers.delete(id);return realClearTimeout(id);};`)}))
	_, err := page.Goto(baseURL + "/components/toast")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	var markup bytes.Buffer
	for _, id := range []string{"first-timers", "second-timers"} {
		require.NoError(t, toast.ToastContainer(toast.ContainerConfig{ID: id, DisplayDuration: 12345}).Render(context.Background(), &markup))
	}
	_, err = page.Evaluate(`html=>document.body.insertAdjacentHTML('beforeend',html)`, markup.String())
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => {const el=document.getElementById('second-timers');return typeof Alpine.$data(el).addNotification==='function';}`, nil)
	require.NoError(t, err)
	_, err = page.Evaluate(`() => window.dispatchEvent(new CustomEvent('notify',{detail:{tone:'info',title:'Timer'}}))`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => toastTimers.size===2`, nil)
	require.NoError(t, err)
	_, err = page.Evaluate(`() => window.dispatchEvent(new CustomEvent('pause-auto-dismiss',{detail:{container:document.getElementById('first-timers')}}))`, nil)
	require.NoError(t, err)
	value, err := page.Evaluate(`() => toastTimers.size`, nil)
	require.NoError(t, err)
	require.EqualValues(t, 1, value)
	_, err = page.Evaluate(`() => document.getElementById('second-timers').remove()`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => toastTimers.size===0`, nil)
	require.NoError(t, err)
}

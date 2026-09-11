//go:build e2e && full

package e2e

import (
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestModernSiteHasOneTOCOwner(t *testing.T) {
	page := newPage(t, sharedBrowser)
	require.NoError(t, page.AddInitScript(playwright.Script{Content: new(`window.tocObservers=[]; const Original=IntersectionObserver; window.IntersectionObserver=class extends Original { constructor(cb, options){super(cb,options);this.targets=new Set();window.tocObservers.push(this);} observe(el){this.targets.add(el);super.observe(el);} disconnect(){this.targets.clear();super.disconnect();} };`)}))
	_, err := page.Goto(baseURL + "/components/pagination")
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => document.querySelectorAll('[data-toc-link]').length > 1`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => tocObservers.filter(o=>Array.from(o.targets).some(el=>el.matches('[data-toc-heading]'))).length===1`, nil)
	require.NoError(t, err, "only the current shell must observe TOC headings")
	_, err = page.Evaluate(`async () => { await htmx.ajax('GET','/components/tabs',{target:'#main-content',swap:'innerHTML'}); }`)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => tocObservers.filter(o=>Array.from(o.targets).some(el=>el.matches('[data-toc-heading]'))).length===1`, nil)
	require.NoError(t, err, "navigation must release the previous TOC observer")
}

func TestModernTickerDeclarativeCancellation(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/examples/ticker")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	result, err := page.Evaluate(`async () => {const root=document.querySelector('#ticker-fragment'); const state=Alpine.$data(root); const connector=root.querySelector('[hx-sse\\:connect]');state.paused=true; await Alpine.nextTick();const paused=new CustomEvent('htmx:sse:before:message',{bubbles:true,cancelable:true});connector.dispatchEvent(paused);const connected=state.connected;state.paused=false;const resumed=new CustomEvent('htmx:sse:before:message',{bubbles:true,cancelable:true});connector.dispatchEvent(resumed);connector.dispatchEvent(new CustomEvent('htmx:sse:error',{bubbles:true}));return {paused:paused.defaultPrevented,resumed:resumed.defaultPrevented,connected,errorDisconnected:!state.connected};}`)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"paused": true, "resumed": false, "connected": true, "errorDisconnected": true}, result)
}

// Configuration is server-owned on mount. Replacing the root resets local state
// deliberately; retaining a root must not silently be made the default swap.
func TestModernOverlayConfigurationReplacement(t *testing.T) {
	for _, spec := range []struct{ route, selector, dataset, state string }{
		{"tabs", "[data-tabs-config]", "tabsConfig", "selectedTab"},
		{"carousel", "[data-carousel-slides]", "carouselSlides", "slides"},
	} {
		t.Run(spec.route, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			_, err := page.Goto(baseURL + "/components/" + spec.route)
			require.NoError(t, err)
			require.NoError(t, waitForAlpine(page))
			result, err := page.Evaluate(`async spec => {
    const root=document.querySelector(spec.selector);const next=root.cloneNode(true);
    next.id='replacement-config';
    if(spec.dataset==='tabsConfig'){const config=goshtosoParseData(root.dataset.tabsConfig, {});config.default=config.ids.at(-1);next.dataset.tabsConfig=btoa(JSON.stringify(config));window.expectedConfig=config.default;}
    else {const slides=goshtosoParseData(root.dataset.carouselSlides, []).slice(0,1);next.dataset.carouselSlides=btoa(JSON.stringify(slides));window.expectedConfig=slides;}
    await htmx.swap({sourceElement:root,target:root,text:next.outerHTML,swap:'outerHTML'});await Alpine.nextTick();
    const replaced=document.querySelector('#replacement-config');return {newRoot:replaced!==root,actual:JSON.stringify(Alpine.$data(replaced)[spec.state]),expected:JSON.stringify(window.expectedConfig)};
   }`, map[string]any{"selector": spec.selector, "dataset": spec.dataset, "state": spec.state})
			require.NoError(t, err)
			got := result.(map[string]any)
			require.Equal(t, true, got["newRoot"])
			require.Equal(t, got["expected"], got["actual"])
		})
	}
}

func TestModernSecondaryPartialOwnsInitialization(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/components/button")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	_, err = page.Evaluate(`async () => {
  window.partialLifecycle={init:0,destroy:0};Alpine.data('partialLifecycle',()=>({init(){window.partialLifecycle.init++},destroy(){window.partialLifecycle.destroy++}}));
  document.body.insertAdjacentHTML('beforeend','<aside id="sidebar-nav-content"></aside>');
  const main=document.querySelector('#main-content');
  window.partialHTML='<hx-partial hx-target="#sidebar-nav-content" hx-swap="innerHTML"><div x-data="partialLifecycle"><button hx-get="/components/button">Nested request</button></div></hx-partial>';
  await htmx.swap({sourceElement:main,target:main,text:window.partialHTML,swap:'innerHTML'});await Alpine.nextTick();
 }`)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => partialLifecycle.init===1 && partialLifecycle.destroy===0`, nil)
	require.NoError(t, err)
	_, err = page.Evaluate(`async () => {const main=document.querySelector('#main-content');await htmx.swap({sourceElement:main,target:main,text:window.partialHTML,swap:'innerHTML'});await Alpine.nextTick();}`)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => partialLifecycle.init===2 && partialLifecycle.destroy===1`, nil)
	require.NoError(t, err)
}

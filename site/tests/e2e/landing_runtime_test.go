//go:build e2e && full

package e2e

import (
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestLandingPlaygroundWaitsForComponentRuntime(t *testing.T) {
	page := newPage(t, sharedBrowser)
	held := make(chan playwright.Route, 1)
	require.NoError(t, page.Route("**/assets/js/goshtoso.min.js", func(route playwright.Route) { held <- route }))
	require.NoError(t, page.AddInitScript(playwright.Script{Content: playwright.String(`
 window.runtimeRequests=[]; window.runtimeErrors=[];
 const originalError=console.error.bind(console);
 console.error=(...args)=>{runtimeErrors.push(args.map(String).join(' '));originalError(...args);};
 window.addEventListener('error',e=>runtimeErrors.push(e.message));
 document.addEventListener('htmx:before:request',e=>{
   if(e.detail.ctx.sourceElement.closest('[data-table-requests]')) runtimeRequests.push(typeof window.goshtosoRequests);
 });
 `)}))
	_, err := page.Goto(baseURL+"/playground/theme", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateCommit})
	require.NoError(t, err)
	var bundle playwright.Route
	select {
	case bundle = <-held:
	case <-time.After(5 * time.Second):
		t.Fatal("playground must load the first-party component runtime")
	}
	// The parser finishes while this defer script is held, but htmx's initial
	// processing must wait for the declaration's complete dependency chain.
	_, err = page.WaitForFunction(`()=>document.readyState==='interactive' && window.htmx`, nil)
	require.NoError(t, err)
	requests, err := page.Evaluate(`()=>runtimeRequests`)
	require.NoError(t, err)
	require.Empty(t, requests)
	missing, err := page.Evaluate(`()=>typeof window.goshtosoRequests==='undefined'`)
	require.NoError(t, err)
	require.Equal(t, true, missing)
	require.NoError(t, bundle.Continue())
	_, err = page.WaitForFunction(`()=>document.querySelector('#home-table-tbody')?.textContent.includes('Alice Brown')`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#home-table th").First().Click())
	_, err = page.WaitForFunction(`()=>runtimeRequests.length>=2`, nil)
	require.NoError(t, err)
	requests, err = page.Evaluate(`()=>runtimeRequests`)
	require.NoError(t, err)
	for _, kind := range requests.([]interface{}) {
		require.Equal(t, "object", kind, "every table request must have its coordination runtime")
	}
	errors, err := page.Evaluate(`()=>runtimeErrors`)
	require.NoError(t, err)
	require.Empty(t, errors)
}

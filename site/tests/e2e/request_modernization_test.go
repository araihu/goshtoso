//go:build e2e && full

package e2e

import (
	"fmt"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/combobox"
	"github.com/araihu/goshtoso/components/form"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

// Requests use the rendered components and vendored runtimes. A controllable
// fetch transport lets responses arrive in the opposite order without sleeps.
func modernizationPage(t *testing.T, body string) playwright.Page {
	t.Helper()
	page := newPage(t, sharedBrowser)
	require.NoError(t, page.AddInitScript(playwright.Script{Content: playwright.String(`
 window.requests=[];
 const originalFetch=window.fetch.bind(window);
 window.fetch=(url, options)=>{
   if(!String(url).includes('/modern-api')) return originalFetch(url, options);
   return new Promise((resolve,reject)=>{
     const request={url:String(url), body:options.body ? String(options.body) : '', aborted:false,
       resolve:(html,status=200,headers={})=>resolve(new Response(html,{status,headers:{'Content-Type':'text/html',...headers}}))};
     options.signal.addEventListener('abort',()=>{request.aborted=true;reject(new DOMException('aborted','AbortError'));},{once:true});
     window.requests.push(request);
   });
 };
 `)}))
	document := renderInteractiveDocument(t, head.DependenciesMinimal(head.WithLocalRuntime()), templ.Raw(body))
	require.NoError(t, page.Route("**/modern-fixture", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{Status: playwright.Int(200), ContentType: playwright.String("text/html"), Body: document}))
	}))
	_, err := page.Goto(baseURL + "/modern-fixture")
	require.NoError(t, err)
	_, err = page.WaitForFunction(`()=>window.htmx && window.Alpine && document.querySelector('[data-htmx-powered]')`, nil)
	require.NoError(t, err)
	return page
}

func modernizationWait(t *testing.T, page playwright.Page, condition string) {
	t.Helper()
	_, err := page.WaitForFunction(condition, nil)
	if err != nil {
		debug, _ := page.Evaluate(`()=>({requests:requests.map(r=>({url:r.url,aborted:r.aborted}))})`)
		t.Logf("failed state: %#v", debug)
	}
	require.NoError(t, err)
}

func TestModernTablePayloadAndLatestRequest(t *testing.T) {
	cfg := table.Config{ID: "one", HTMX: &table.HTMXConfig{Endpoint: "/modern-api/rows?search=stale&active=true&page=7"},
		ExtraQueryParams: "&extra=a%26b&per_page=9", Columns: []table.Column{{Key: "name", Label: "Name", Sortable: true}},
		Rows:    []table.Row{{ID: "old", Cells: map[string]table.Cell{"name": {Text: "Original"}}}},
		Filters: &table.FilterConfig{Filters: []table.Filter{{Key: "search", Type: table.FilterSearch}, {Key: "active", Type: table.FilterToggle, DefaultValue: "false"}, {Key: "odd'[]&", Type: table.FilterSearch, DefaultValue: "0"}}},
	}
	other := cfg
	other.ID = "two"
	body := `<form><input name="unrelated" value="private">` + renderComponentFragment(t, table.Table(cfg)) + renderComponentFragment(t, table.Table(other)) + `</form>`
	page := modernizationPage(t, body)
	require.NoError(t, page.Locator("#one-filters input[type=search]").First().Fill("first"))
	modernizationWait(t, page, `()=>requests.length===1`)
	_, err := page.Evaluate(`()=>{const u=new URL(requests[0].url,location.href);window.initialPayload=Object.fromEntries(u.searchParams);}`)
	require.NoError(t, err)
	payload, err := page.Evaluate(`()=>initialPayload`)
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"_filter": "1", "extra": "a&b", "per_page": "9", "search": "first", "table_id": "one", "odd'[]&": "0"}, payload)
	// Sorting shares the first table's stable queue, but cannot cancel table two.
	require.NoError(t, page.Locator("#two-filters input[type=search]").First().Fill("independent"))
	modernizationWait(t, page, `()=>requests.length===2`)
	require.NoError(t, page.Locator("#one th").Click())
	modernizationWait(t, page, `()=>requests.length===3 && requests[0].aborted`)
	aborted, err := page.Evaluate(`()=>requests[1].aborted`)
	require.NoError(t, err)
	require.Equal(t, false, aborted)
	// Third intent arrives after the first abort's finally: native replace alone
	// loses the second request's handle in htmx 4.0.0.
	require.NoError(t, page.Locator("#one-filters input[type=search]").First().Fill("newest"))
	modernizationWait(t, page, `()=>requests.length===4 && requests[2].aborted`)
	_, err = page.Evaluate(`()=>{requests[3].resolve('<tr><td>Newest sort</td></tr>');requests[2].resolve('<tr><td>Stale second</td></tr>');requests[0].resolve('<tr><td>Stale</td></tr>');requests[1].resolve('<tr><td>Independent</td></tr>');}`)
	require.NoError(t, err)
	modernizationWait(t, page, `()=>document.querySelector('#one-tbody').textContent.includes('Newest sort') && document.querySelector('#two-tbody').textContent.includes('Independent')`)
	require.NoError(t, page.Locator("#one-filters input[type=search]").First().Fill(""))
	modernizationWait(t, page, `()=>requests.length===5`)
	absent, err := page.Evaluate(`()=>{const q=new URL(requests[4].url,location.href).searchParams;return !q.has('search')&&!q.has('active')&&!q.has('page')&&!q.has('unrelated');}`)
	require.NoError(t, err)
	require.Equal(t, true, absent)
	_, err = page.Evaluate(`()=>document.querySelector('#one-filters input').dispatchEvent(new CustomEvent('htmx:before:cleanup',{bubbles:true}))`)
	require.NoError(t, err)
	aborted, err = page.Evaluate(`()=>requests[4].aborted`)
	require.NoError(t, err)
	require.Equal(t, false, aborted, "descendant cleanup must not abort a root request")
	_, err = page.Evaluate(`()=>{const root=document.querySelector('#one').closest('[data-table-requests]');htmx.ajax('GET','/modern-api/remove',{source:document.body,target:root,swap:'outerHTML'});}`)
	require.NoError(t, err)
	modernizationWait(t, page, `()=>requests.length===6`)
	_, err = page.Evaluate(`()=>requests[5].resolve('<div>Removed</div>')`)
	require.NoError(t, err)
	modernizationWait(t, page, `()=>requests[4].aborted && !document.querySelector('#one')`)
}

func TestModernTableNativeSentinelRetryAndChain(t *testing.T) {
	for _, contained := range []bool{false, true} {
		t.Run(fmt.Sprint(contained), func(t *testing.T) {
			cfg := table.Config{ID: "scroll", HTMX: &table.HTMXConfig{Endpoint: "/modern-api/rows"}, Columns: []table.Column{{Key: "name", Label: "Name"}},
				Rows:       []table.Row{{ID: "first", Cells: map[string]table.Cell{"name": {Text: "First"}}}},
				Pagination: &table.PaginationConfig{Mode: table.PaginationInfiniteScroll, CurrentPage: 1, PerPage: 1, HasMore: true}}
			body := renderComponentFragment(t, table.Table(cfg))
			if !contained {
				body = `<div style="overflow-y:auto;height:10px">Unrelated scroller</div>` + body
			}
			if contained {
				body = `<div id="scroller" class="overflow-y-auto" style="height:100px;overflow-y:auto"><div style="height:900px"></div>` + body + `</div>`
			}
			page := modernizationPage(t, body)
			if contained {
				count, err := page.Evaluate(`()=>requests.length`)
				require.NoError(t, err)
				require.EqualValues(t, 0, count)
				_, err = page.Evaluate(`()=>document.querySelector('#scroller').scrollTop=1000`)
				require.NoError(t, err)
			}
			modernizationWait(t, page, `()=>requests.length===1`)
			status := 500
			if contained {
				status = 403
			}
			_, err := page.Evaluate(`status=>requests[0].resolve('<p>Unexpected failure</p>',status)`, status)
			require.NoError(t, err)
			modernizationWait(t, page, `()=>!document.querySelector('#scroll-sentinel button').disabled`)
			require.NoError(t, page.Locator("#scroll-sentinel button").Click())
			modernizationWait(t, page, `()=>requests.length===2`)
			next := cfg
			next.Pagination = &table.PaginationConfig{Mode: table.PaginationInfiniteScroll, CurrentPage: 2, PerPage: 1, HasMore: true}
			next.Rows = []table.Row{{ID: "second", Cells: map[string]table.Cell{"name": {Text: "Second"}}}}
			_, err = page.Evaluate(`html=>requests[1].resolve(html)`, renderComponentFragment(t, table.TableRows(next)))
			require.NoError(t, err)
			modernizationWait(t, page, `()=>requests.length===3`)
			_, err = page.Evaluate(`()=>requests[2].resolve('<tr id="third"><td>Third</td></tr>')`)
			require.NoError(t, err)
			modernizationWait(t, page, `()=>document.querySelector('#third') && !document.querySelector('#scroll-sentinel')`)
			count, err := page.Locator("#scroll tbody tr").Count()
			require.NoError(t, err)
			require.Equal(t, 3, count)
			urls, err := page.Evaluate(`()=>requests.map(r=>new URL(r.url,location.href).searchParams.get('page'))`)
			require.NoError(t, err)
			require.Equal(t, []interface{}{"2", "2", "3"}, urls)
		})
	}
}

func TestModernFormSubmissionOwnsValidationAndStatus(t *testing.T) {
	field := form.FieldGroup(form.FieldGroupConfig{ID: "name", Input: &textinput.Config{Name: "name"}, Validation: &form.ValidationConfig{Endpoint: "/modern-api/validate"}})
	root := renderComponentFragment(t, form.Form(form.Config{ID: "edit", HTMX: &form.HTMXConfig{Post: "/modern-api/submit", Target: "#result", Sync: "this:drop", Disable: "findAll button[type=submit]:not(:disabled)"}, RootAttrs: templ.Attributes{"hx-status:5xx": "swap:none", "hx-indicator": "#busy"}}))
	body := root[:len(root)-len("</form>")] + renderComponentFragment(t, field) + `<button type="submit">Save</button><button id="locked" type="submit" disabled>Locked</button></form><span id="busy"></span><div id="result">Original</div>`
	page := modernizationPage(t, body)
	require.NoError(t, page.Locator("#name input").Fill("First"))
	require.NoError(t, page.Locator("#name input").Press("Tab"))
	modernizationWait(t, page, `()=>requests.length===1`)
	require.NoError(t, page.Locator("#name input").Fill("Latest"))
	require.NoError(t, page.Locator("#name input").Press("Tab"))
	modernizationWait(t, page, `()=>requests.length===2 && requests[0].aborted`)
	require.NoError(t, page.Locator("#edit button").First().Click())
	modernizationWait(t, page, `()=>requests.length===3 && requests[1].aborted`)
	// Validation attempted during a submit must not run even with an external indicator.
	_, err := page.Evaluate(`()=>document.querySelector('#name input').dispatchEvent(new Event('change',{bubbles:true}))`)
	require.NoError(t, err)
	// Programmatic duplicate submit also drops; HTML disabling alone is insufficient.
	_, err = page.Evaluate(`()=>document.querySelector('#edit').dispatchEvent(new Event('submit',{bubbles:true,cancelable:true}))`)
	require.NoError(t, err)
	_, err = page.Evaluate(`()=>requests[2].resolve('<p>Do not replace this form</p>',500)`)
	require.NoError(t, err)
	modernizationWait(t, page, `()=>!document.querySelector('#edit button').disabled`)
	count, err := page.Evaluate(`()=>requests.length`)
	require.NoError(t, err)
	require.EqualValues(t, 3, count)
	text, err := page.Locator("#result").InnerText()
	require.NoError(t, err)
	require.Equal(t, "Original", text)
	locked, err := page.Locator("#locked").IsDisabled()
	require.NoError(t, err)
	require.True(t, locked)
	require.NoError(t, page.Locator("#edit button").First().Click())
	modernizationWait(t, page, `()=>requests.length===4`)
	_, err = page.Evaluate(`()=>requests[3].resolve('<p>Validation feedback</p>',422)`)
	require.NoError(t, err)
	modernizationWait(t, page, `()=>document.querySelector('#result').textContent==='Validation feedback'`)
}

func TestModernComboboxMutationCancelsStaleSearch(t *testing.T) {
	cfg := combobox.Config{ID: "choices", Name: "choice", Mode: combobox.ModeMultiple, Source: combobox.Source{LazyEndpoint: "/modern-api/options"}, OptionsEndpoint: "/modern-api/options", ToggleEndpoint: "/modern-api/toggle", ClearEndpoint: "/modern-api/clear", EnableSearch: true, EnableClearAll: true, DisablePersistence: true, DependsOn: []string{"provider"}}
	state := combobox.State{Options: []combobox.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}}
	page := modernizationPage(t, `<input name="provider" value="cloud"><input name="unrelated" value="private">`+renderComponentFragment(t, combobox.Combobox(cfg, state)))
	require.NoError(t, page.Locator("#choices-trigger").Click())
	require.NoError(t, page.Locator("[data-combobox-search]").Fill("a"))
	modernizationWait(t, page, `()=>requests.length===1`)
	require.NoError(t, page.Locator("[data-combobox-search]").Fill("al"))
	modernizationWait(t, page, `()=>requests.length===2 && requests[0].aborted`)
	_, err := page.Evaluate(`()=>{const input=document.querySelector('[data-combobox-search]');input.setSelectionRange(1,1);document.querySelector('#choices-options [data-combobox-option]').dispatchEvent(new MouseEvent('click',{bubbles:true}));}`)
	require.NoError(t, err)
	modernizationWait(t, page, `()=>requests.length===3 && requests[1].aborted`)
	readOnly, err := page.Locator("[data-combobox-search]").Evaluate(`el=>el.readOnly`, nil)
	require.NoError(t, err)
	require.Equal(t, true, readOnly)
	// Native drop admission guards stale selection snapshots even synthetic clicks.
	_, err = page.Evaluate(`()=>document.querySelectorAll('#choices-options [data-combobox-option]')[1].dispatchEvent(new MouseEvent('click',{bubbles:true}))`)
	require.NoError(t, err)
	count, err := page.Evaluate(`()=>requests.length`)
	require.NoError(t, err)
	require.EqualValues(t, 3, count)
	// A real component response morphs the root while the user keeps query focus.
	state.Selected = []string{"a"}
	state.Search = "al"
	_, err = page.Evaluate(`html=>requests[2].resolve(html)`, renderComponentFragment(t, combobox.Combobox(cfg, state)))
	require.NoError(t, err)
	modernizationWait(t, page, `()=>document.querySelector('#choices input[type=hidden]')?.value==='a' && !document.querySelector('[data-combobox-search]').readOnly`)
	focus, err := page.Evaluate(`()=>{const input=document.querySelector('[data-combobox-search]');return {focus:document.activeElement===input,value:input.value,caret:input.selectionStart};}`)
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"focus": true, "value": "al", "caret": 1}, focus)
	_, err = page.Evaluate(`()=>document.querySelectorAll('#choices-options [data-combobox-option]')[1].dispatchEvent(new MouseEvent('click',{bubbles:true}))`)
	require.NoError(t, err)
	modernizationWait(t, page, `()=>requests.length===4`)
	payload, err := page.Evaluate(`()=>Object.fromEntries(new URLSearchParams(requests[3].body))`)
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"choice": "a", "q": "al", "provider": "cloud", "value": "b"}, payload)
	state.Selected = []string{"a", "b"}
	_, err = page.Evaluate(`html=>requests[3].resolve(html)`, renderComponentFragment(t, combobox.Combobox(cfg, state)))
	require.NoError(t, err)
	modernizationWait(t, page, `()=>document.querySelectorAll('#choices input[type=hidden]').length===2 && !document.querySelector('#choices-trigger').disabled`)
	require.NoError(t, page.Locator("#choices [data-combobox-clear-all]").Click())
	modernizationWait(t, page, `()=>requests.length===5`)
	selected, err := page.Evaluate(`()=>new URLSearchParams(requests[4].body).getAll('choice')`)
	require.NoError(t, err)
	require.Equal(t, []interface{}{"a", "b"}, selected)
	state.Selected = nil
	_, err = page.Evaluate(`html=>requests[4].resolve(html)`, renderComponentFragment(t, combobox.Combobox(cfg, state)))
	require.NoError(t, err)
	modernizationWait(t, page, `()=>document.querySelectorAll('#choices input[type=hidden]').length===0 && !document.querySelector('#choices-trigger').disabled`)
	require.NoError(t, page.Locator("#choices-options [data-combobox-option]").First().Click())
	modernizationWait(t, page, `()=>requests.length===6`)
	// Provider 502 intentionally retargets the options and keeps the root alive.
	_, err = page.Evaluate(`()=>requests[5].resolve('<ul id="choices-options"><li>Retry provider</li></ul>',502,{'HX-Retarget':'#choices-options'})`)
	require.NoError(t, err)
	modernizationWait(t, page, `()=>document.querySelector('#choices-options').textContent.includes('Retry provider') && !document.querySelector('#choices-trigger').disabled`)
}

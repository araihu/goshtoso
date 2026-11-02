//go:build e2e && full

package e2e

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/layout"
	"github.com/araihu/goshtoso/expressions"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func layoutFixture(t *testing.T, contentDrag ...bool) *httptest.Server {
	t.Helper()
	source := func(changed bool) layout.Instance {
		c := layout.Config{ID: "alpha", Label: "Alpha", Mode: layout.ModeMasonry, Columns: 3, Reorder: &layout.ReorderConfig{Group: "widgets"}, Items: []layout.Item{
			{ID: "a", Label: `A <unsafe>`, RootClass: "p-3 border m-2", Content: templ.Raw(`<p>Plain content</p><span id="inside-markers" data-layout-content data-layout-drop-target>Consumer content</span><label>Draft<input id="layout-draft" value="initial"></label><a href="#help" id="content-link">Help</a><span id="interactive-checkbox" role="checkbox" style="display:inline-block;width:80px;height:20px">Checkbox</span><span id="focusable-content" tabindex="0" style="display:inline-block;width:80px;height:20px">Focusable</span><audio id="media-audio" controls style="width:100px;height:20px"></audio><video id="media-video" controls style="width:100px;height:20px"></video>`)},
			{ID: "b", Label: "Beta", Content: templ.Raw(`<div style="height:111.5px">Tall content</div>`)},
			{ID: "c", Label: "Gamma"},
			{ID: "d", Label: "Delta", Span: layout.SpanFullWidth, Content: templ.Raw(`<p>Full width</p>`)},
		}}
		if changed {
			c.Label = "Alpha updated"
			c.Gap = layout.GapLarge
			c.RootClass = "border"
			c.RootAttrs = templ.Attributes{"title": "Updated"}
		}
		return layout.Container(c).WithContentDrag(len(contentDrag) > 0 && contentDrag[0]).WithExpressions(expressions.Layout{MoveAnnouncement: func(item, container, position string) string { return position + " in " + container + " for " + item }})
	}
	empty := func(id string) layout.Instance {
		return layout.Container(layout.Config{ID: id, Label: id, Mode: layout.ModeGrid, Columns: 2, Reorder: &layout.ReorderConfig{Group: "widgets"}, RootClass: "min-h-24"}).WithContentDrag(len(contentDrag) > 0 && contentDrag[0])
	}
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.HandleFunc("GET /fragment", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		require.NoError(t, source(true).Render(r.Context(), w))
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, `<!doctype html><html data-theme="goshtoso"><head>`)
		require.NoError(t, head.Dependencies(head.WithLocalRuntime()).Render(r.Context(), w))
		_, _ = io.WriteString(w, `</head><body class="bg-surface text-on-surface dark:bg-surface-dark dark:text-on-surface-dark"><main style="max-width:1000px;padding:16px"><h1>Layout fixture</h1><span id="outside-markers" data-layout-content data-layout-drop-target>Unrelated content</span><div id="status" tabindex="-1">Status</div><button id="refresh" hx-get="/fragment" hx-target="#alpha" hx-swap="outerMorph">Refresh</button><div id="bounds" style="width:100%">`)
		require.NoError(t, source(false).Render(r.Context(), w))
		_, _ = io.WriteString(w, `</div><h2>Inbox</h2>`)
		require.NoError(t, empty("inbox").Render(r.Context(), w))
		_, _ = io.WriteString(w, `<fieldset disabled id="disabled"><legend>Disabled destinations</legend>`)
		require.NoError(t, empty("unavailable").Render(r.Context(), w))
		_, _ = io.WriteString(w, `</fieldset><fieldset disabled id="legend-fieldset"><legend>`)
		require.NoError(t, empty("legend-allowed").Render(r.Context(), w))
		_, _ = io.WriteString(w, `</legend></fieldset>`)
		inner := layout.Container(layout.Config{ID: "inner", Label: "Inner", Mode: layout.ModeMasonry, Columns: 3, Reorder: &layout.ReorderConfig{Group: "inner"}, Items: []layout.Item{{ID: "inner-a", Label: "Inner A", Content: templ.Raw(`<p>Nested</p>`)}, {ID: "inner-b", Label: "Inner B"}}})
		require.NoError(t, layout.Container(layout.Config{ID: "outer", Label: "Outer", Mode: layout.ModeGrid, Columns: 2, Reorder: &layout.ReorderConfig{Group: "groups"}, Items: []layout.Item{{ID: "outer-a", Label: "Group", Content: inner}, {ID: "outer-b", Label: "Other group", Content: templ.Raw(`<p>Other</p>`)}}}).Render(r.Context(), w))
		_, _ = io.WriteString(w, `</main></body></html>`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}
func layoutReady(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.WaitForFunction(`() => document.getElementById('alpha')?.dataset.layoutReorderState==='ready' && document.getElementById('alpha').dataset.layoutGeometryState==='settled'`, nil)
	require.NoError(t, err)
}
func layoutEval(t *testing.T, page playwright.Page, script string, args ...any) any {
	t.Helper()
	var arg any
	if len(args) > 0 {
		arg = args[0]
	}
	v, err := page.Evaluate(script, arg)
	require.NoError(t, err)
	return v
}

const layoutOrder = `id => Array.from(document.getElementById(id).children).filter(n=>n.dataset.goshtosoLayout==='item').map(n=>n.dataset.layoutItemId).join(',')`
const layoutRequest = `([item,target,anchor]) => document.getElementById('alpha').dispatchEvent(new CustomEvent('layout:request-move',{bubbles:true,detail:{itemId:item,fromContainerId:'alpha',toContainerId:target,beforeItemId:anchor ?? null}}))`

func TestLayout_CurrentSourceGeometry(t *testing.T) {
	server := layoutFixture(t)
	for _, theme := range []string{"goshtoso", "minimal"} {
		for _, dark := range []bool{false, true} {
			for _, width := range []int{390, 1440} {
				t.Run(fmt.Sprintf("%s/dark=%t/width=%d", theme, dark, width), func(t *testing.T) {
					page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: width, Height: 1100}})
					failures := watchPageFailures(page)
					_, err := page.Goto(server.URL)
					require.NoError(t, err)
					layoutReady(t, page)
					layoutEval(t, page, `([theme,dark])=>{document.documentElement.dataset.theme=theme;document.documentElement.classList.toggle('dark',dark)}`, []any{theme, dark})
					for _, gap := range []int{12, 16, 24} {
						layoutEval(t, page, `gap=>document.getElementById('alpha').dataset.layoutGap=String(gap)`, gap)
						layoutReady(t, page)
						result := layoutEval(t, page, `()=>{const r=document.getElementById('alpha'),g=+r.dataset.layoutGap;const rows=Array.from(r.children).filter(n=>n.dataset.goshtosoLayout==='item');return rows.every(n=>{const s=getComputedStyle(n),h=n.getBoundingClientRect().height+parseFloat(s.marginBlockStart)+parseFloat(s.marginBlockEnd),span=parseFloat(n.style.getPropertyValue('--layout-row-span'));return span-(h+g)>=0&&span-(h+g)<1}) && rows.every((n,i)=>rows.slice(i+1).every(m=>{const a=n.getBoundingClientRect(),b=m.getBoundingClientRect();return a.right<=b.left+0.5||b.right<=a.left+0.5||a.bottom<=b.top+0.5||b.bottom<=a.top+0.5}))}`)
						require.Equal(t, true, result)
					}
					layoutEval(t, page, `()=>document.getElementById('bounds').style.width='220px'`)
					layoutReady(t, page)
					_, err = page.WaitForFunction(`()=>getComputedStyle(document.getElementById('alpha')).gridTemplateColumns.split(' ').length===1`, nil)
					require.NoError(t, err)
					require.Equal(t, "a,b,c,d", layoutEval(t, page, layoutOrder, "alpha"))
					for _, mode := range []string{"", "grid", "masonry"} {
						layoutEval(t, page, `mode=>document.getElementById('alpha').dataset.layoutMode=mode`, mode)
						layoutReady(t, page)
					}
					failures.RequireEmpty(t)
				})
			}
		}
	}
}
func TestLayout_CurrentSourceMovement(t *testing.T) {
	server := layoutFixture(t)
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(server.URL)
	require.NoError(t, err)
	layoutReady(t, page)
	layoutEval(t, page, `()=>{window.moves=[];window.errors=[];document.addEventListener('layout:move',e=>moves.push(e.detail));document.addEventListener('layout:error',e=>errors.push(e.detail.code));window.retained=document.getElementById('layout-draft');}`)
	require.NoError(t, page.Locator("#layout-draft").Fill("retained draft"))
	require.NoError(t, page.Locator("#alpha [data-layout-handle]").First().Focus())
	require.NoError(t, page.Keyboard().Press("Alt+ArrowDown"))
	require.Equal(t, "b,a,c,d", layoutEval(t, page, layoutOrder, "alpha"))
	require.NoError(t, page.Keyboard().Press("Alt+Home"))
	require.Equal(t, "a,b,c,d", layoutEval(t, page, layoutOrder, "alpha"))
	require.NoError(t, page.Keyboard().Press("Alt+End"))
	require.Equal(t, "b,c,d,a", layoutEval(t, page, layoutOrder, "alpha"))
	require.NoError(t, page.Keyboard().Press("Alt+PageDown"))
	require.Equal(t, "a", layoutEval(t, page, layoutOrder, "inbox"))
	require.Equal(t, "inbox--layout-instructions", layoutEval(t, page, `()=>document.activeElement.getAttribute('aria-describedby')`))
	require.Equal(t, "retained draft", layoutEval(t, page, `()=>retained.value`))
	require.Equal(t, true, layoutEval(t, page, `()=>retained===document.getElementById('layout-draft')`))
	require.Equal(t, "A <unsafe> moved to inbox, position 1.", layoutEval(t, page, `()=>document.querySelector('#inbox [data-layout-status]').textContent`))
	require.NoError(t, page.Keyboard().Press("Alt+PageUp"))
	require.Equal(t, "b,c,d,a", layoutEval(t, page, layoutOrder, "alpha"))
	layoutEval(t, page, `()=>{window.before=e=>e.preventDefault();document.getElementById('alpha').addEventListener('layout:before-move',before);}`)
	layoutEval(t, page, layoutRequest, []any{"a", "inbox", nil})
	require.Equal(t, "", layoutEval(t, page, layoutOrder, "inbox"))
	layoutEval(t, page, `()=>document.getElementById('alpha').removeEventListener('layout:before-move',before)`)
	layoutEval(t, page, layoutRequest, []any{"a", "unavailable", nil})
	require.Equal(t, "", layoutEval(t, page, layoutOrder, "unavailable"))
	require.Equal(t, "reorder_unavailable", layoutEval(t, page, `()=>errors.at(-1)`))
	layoutEval(t, page, layoutRequest, []any{"a", "legend-allowed", nil})
	require.Equal(t, "a", layoutEval(t, page, layoutOrder, "legend-allowed"))
	// Programmatic commands and keyboard movement share the same event detail contract.
	require.Equal(t, true, layoutEval(t, page, `()=>moves.every(m=>Object.keys(m).sort().join(',')==='beforeItemId,fromContainerId,itemId,toContainerId')`))
}
func TestLayout_CurrentSourceDisableAndCollisions(t *testing.T) {
	server := layoutFixture(t)
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(server.URL)
	require.NoError(t, err)
	layoutReady(t, page)
	layoutEval(t, page, `()=>{window.moves=0;window.errors=[];document.addEventListener('layout:move',()=>moves++);document.addEventListener('layout:error',e=>errors.push(e.detail.code));const root=document.getElementById('alpha');root.addEventListener('layout:before-move',()=>document.getElementById('disabled').append(document.getElementById('inbox')),{once:true});}`)
	layoutEval(t, page, layoutRequest, []any{"a", "inbox", nil})
	require.Equal(t, 0, layoutEval(t, page, `()=>moves`))
	require.Equal(t, "a,b,c,d", layoutEval(t, page, layoutOrder, "alpha"))
	layoutEval(t, page, `()=>{const collision=document.createElement('div');collision.id='alpha--layout-instructions';collision.dataset.testCollision='';document.body.append(collision)}`)
	_, err = page.WaitForFunction(`()=>document.getElementById('alpha').dataset.layoutReorderState==='blocked'`, nil)
	require.NoError(t, err)
	require.Equal(t, true, layoutEval(t, page, `()=>Array.from(document.querySelectorAll('#alpha [data-layout-handle]')).every(n=>n.hidden&&n.disabled)`))
	layoutEval(t, page, `()=>document.querySelector('[data-test-collision]').remove()`)
	layoutReady(t, page)
	// An explicit container ID colliding with an instruction ID blocks both owners.
	layoutEval(t, page, `()=>{const r=document.getElementById('unavailable');r.id='alpha--layout-instructions';r.dataset.layoutId=r.id}`)
	_, err = page.WaitForFunction(`()=>document.getElementById('alpha').dataset.layoutReorderState==='blocked'`, nil)
	require.NoError(t, err)
	layoutEval(t, page, `()=>{const r=document.querySelector('[data-layout-id="alpha--layout-instructions"]');r.id='unavailable';r.dataset.layoutId=r.id}`)
	layoutReady(t, page)
	layoutEval(t, page, `()=>{document.getElementById('status').focus();document.getElementById('disabled').append(document.getElementById('alpha'))}`)
	layoutEval(t, page, layoutRequest, []any{"a", "legend-allowed", nil})
	require.Equal(t, 0, layoutEval(t, page, `()=>moves`))
	require.Equal(t, true, layoutEval(t, page, `()=>{const b=document.querySelector('#alpha [data-layout-handle]');return !b.hidden&&b.matches(':disabled')&&b.getBoundingClientRect().height>=44}`))
	layoutEval(t, page, `()=>document.getElementById('disabled').disabled=false`)
	layoutEval(t, page, layoutRequest, []any{"a", "legend-allowed", nil})
	require.Equal(t, 1, layoutEval(t, page, `()=>moves`))
}
func TestLayout_CurrentSourceMorphAndFallback(t *testing.T) {
	server := layoutFixture(t)
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(server.URL)
	require.NoError(t, err)
	layoutReady(t, page)
	layoutEval(t, page, `()=>{window.previous=document.getElementById('alpha');document.getElementById('layout-draft').focus();document.getElementById('layout-draft').value='unsaved';document.getElementById('refresh').click()}`)
	_, err = page.WaitForFunction(`()=>document.getElementById('alpha').getAttribute('title')==='Updated'`, nil)
	require.NoError(t, err)
	layoutReady(t, page)
	require.Equal(t, true, layoutEval(t, page, `()=>document.getElementById('alpha')===previous`))
	require.Equal(t, "Alpha updated", layoutEval(t, page, `()=>previous.getAttribute('aria-label')`))
	require.Equal(t, "unsaved", layoutEval(t, page, `()=>document.getElementById('layout-draft').value`))
	layoutEval(t, page, `()=>{window.count=0;document.addEventListener('layout:move',()=>count++);const r=document.getElementById('alpha'),next=r.cloneNode(true);r.replaceWith(next);window.Alpine.initTree(next);}`)
	layoutReady(t, page)
	layoutEval(t, page, layoutRequest, []any{"a", "inbox", nil})
	require.Equal(t, 1, layoutEval(t, page, `()=>count`))
	// Masonry falls back to normal responsive grid without ResizeObserver.
	require.NoError(t, page.AddInitScript(playwright.Script{Content: playwright.String(`window.ResizeObserver=undefined`)}))
	_, err = page.Goto(server.URL)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`()=>document.getElementById('alpha').dataset.layoutGeometryState==='fallback'`, nil)
	require.NoError(t, err)
	require.Equal(t, "16px", layoutEval(t, page, `()=>getComputedStyle(document.getElementById('alpha')).rowGap`))
	noJS := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{JavaScriptEnabled: playwright.Bool(false), Viewport: &playwright.Size{Width: 390, Height: 844}})
	_, err = noJS.Goto(server.URL)
	require.NoError(t, err)
	require.Equal(t, true, layoutEval(t, noJS, `()=>Array.from(document.querySelectorAll('#alpha [data-layout-handle]')).every(n=>n.disabled&&n.hidden)`))
}
func TestLayout_CurrentSourcePointer(t *testing.T) {
	server := layoutFixture(t)
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(server.URL)
	require.NoError(t, err)
	layoutReady(t, page)
	box, err := page.Locator("#alpha [data-layout-handle]").First().BoundingBox()
	require.NoError(t, err)
	target, err := page.Locator("#inbox").BoundingBox()
	require.NoError(t, err)
	require.NoError(t, page.Mouse().Move(box.X+10, box.Y+10))
	require.NoError(t, page.Mouse().Down())
	require.NoError(t, page.Mouse().Move(target.X+20, target.Y+20, playwright.MouseMoveOptions{Steps: playwright.Int(5)}))
	require.NoError(t, page.Mouse().Up())
	require.Equal(t, "a", layoutEval(t, page, layoutOrder, "inbox"))
}

func TestLayout_CurrentSourceDemo(t *testing.T) {
	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 390, Height: 844}})
	failures := watchPageFailures(page)
	_, err := page.Goto(baseURL + "/components/layout")
	require.NoError(t, err)
	_, err = page.WaitForFunction(`()=>document.getElementById('layout-mixed')?.dataset.layoutGeometryState==='settled'`, nil)
	require.NoError(t, err)
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Move service to inbox", Exact: playwright.Bool(true)}).Click())
	require.Equal(t, "service", layoutEval(t, page, layoutOrder, "layout-inbox"))
	require.Equal(t, "notes,metrics,chart,short", layoutEval(t, page, layoutOrder, "layout-mixed"))
	require.Equal(t, 2, layoutEval(t, page, `()=>document.querySelectorAll('#layout-nested > [data-goshtoso-layout="item"]').length`))
	failures.RequireEmpty(t)
}

func layoutPickup(t *testing.T, page playwright.Page, root, item string) {
	t.Helper()
	loc := page.Locator(fmt.Sprintf(`#%s > [data-layout-item-id="%s"] > [data-layout-handle-row] > [data-layout-handle]`, root, item))
	require.NoError(t, loc.ScrollIntoViewIfNeeded())
	box, err := loc.BoundingBox()
	require.NoError(t, err)
	require.NoError(t, page.Mouse().Move(box.X+10, box.Y+10))
	require.NoError(t, page.Mouse().Down())
}
func layoutDrop(t *testing.T, page playwright.Page, target string, nearStart bool) {
	t.Helper()
	box, err := page.Locator(target).BoundingBox()
	require.NoError(t, err)
	y := box.Y + box.Height - 2
	if nearStart {
		y = box.Y + 2
	}
	require.NoError(t, page.Mouse().Move(box.X+box.Width/2, y, playwright.MouseMoveOptions{Steps: playwright.Int(5)}))
	require.NoError(t, page.Mouse().Up())
}
func TestLayout_PickupIdentityAndFocus(t *testing.T) {
	server := layoutFixture(t)
	for _, change := range []string{"identity", "membership", "handle", "handle-row", "disable", "content-mode"} {
		t.Run(change, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			_, err := page.Goto(server.URL)
			require.NoError(t, err)
			layoutReady(t, page)
			layoutEval(t, page, `()=>{window.moves=0;window.beforeMoves=0;document.addEventListener('layout:move',()=>moves++);document.addEventListener('layout:before-move',()=>beforeMoves++)}`)
			layoutPickup(t, page, "alpha", "a")
			switch change {
			case "content-mode":
				layoutEval(t, page, `()=>document.getElementById("alpha").dataset.layoutContentDrag="true"`)
			case "identity":
				layoutEval(t, page, `()=>document.querySelector('#alpha > [data-layout-item-id="a"]').dataset.layoutItemId='changed-during-pickup'`)
			case "membership":
				layoutEval(t, page, `()=>{const root=document.getElementById('alpha');root.insertBefore(root.querySelector('[data-layout-item-id="b"]'),root.firstElementChild)}`)
			case "handle":
				layoutEval(t, page, `()=>{const h=document.querySelector('#alpha > [data-layout-item-id="a"] > [data-layout-handle-row] > [data-layout-handle]');h.replaceWith(h.cloneNode(true))}`)
			case "handle-row":
				layoutEval(t, page, `()=>{const row=document.querySelector('#alpha > [data-layout-item-id="a"] > [data-layout-handle-row]');row.replaceWith(row.cloneNode(true))}`)
			case "disable":
				layoutEval(t, page, `()=>{document.getElementById('status').focus();document.getElementById('disabled').append(document.getElementById('alpha'))}`)
			}
			if change == "disable" {
				require.Equal(t, "status", layoutEval(t, page, `()=>document.activeElement.id`))
				require.Equal(t, true, layoutEval(t, page, `()=>{const h=document.querySelector('#alpha [data-layout-handle]');return !h.hidden&&h.matches(':disabled')&&h.getBoundingClientRect().height>=44}`))
			}
			// Coordinates are read after reparenting; a stale gesture must still be rejected.
			layoutDrop(t, page, "#inbox", true)
			require.Equal(t, 0, layoutEval(t, page, `()=>moves`))
			require.Equal(t, 0, layoutEval(t, page, `()=>beforeMoves`))
			require.Equal(t, "", layoutEval(t, page, layoutOrder, "inbox"))
			if change == "disable" {
				layoutEval(t, page, `()=>document.getElementById('disabled').disabled=false`)
				require.Equal(t, "status", layoutEval(t, page, `()=>document.activeElement.id`))
			}
		})
	}
}
func TestLayout_NestedPointerOwnership(t *testing.T) {
	server := layoutFixture(t)
	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	_, err := page.Goto(server.URL)
	require.NoError(t, err)
	layoutReady(t, page)
	require.NoError(t, page.Locator("#outer").ScrollIntoViewIfNeeded())
	layoutPickup(t, page, "outer", "outer-b")
	layoutDrop(t, page, "#inner", true)
	require.Equal(t, "outer-b,outer-a", layoutEval(t, page, layoutOrder, "outer"))
	require.Equal(t, "inner-a,inner-b", layoutEval(t, page, layoutOrder, "inner"))
	layoutPickup(t, page, "inner", "inner-a")
	layoutDrop(t, page, `#inner > [data-layout-item-id="inner-b"]`, false)
	require.Equal(t, "inner-b,inner-a", layoutEval(t, page, layoutOrder, "inner"))
	require.Equal(t, "outer-b,outer-a", layoutEval(t, page, layoutOrder, "outer"))
	layoutEval(t, page, `()=>{window.moves=0;window.errors=[];document.addEventListener('layout:move',()=>moves++);document.addEventListener('layout:error',e=>errors.push(e.detail.code));document.getElementById('inner').dataset.layoutGroup='groups';document.getElementById('outer').dispatchEvent(new CustomEvent('layout:request-move',{bubbles:true,detail:{itemId:'outer-a',fromContainerId:'outer',toContainerId:'inner',beforeItemId:null}}));}`)
	require.Equal(t, 0, layoutEval(t, page, `()=>moves`))
	require.Equal(t, "invalid_move_request", layoutEval(t, page, `()=>errors.at(-1)`))
	failures.RequireEmpty(t)
}
func TestLayout_AttributeMarkerIsolation(t *testing.T) {
	server := layoutFixture(t)
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(server.URL)
	require.NoError(t, err)
	layoutReady(t, page)
	for _, id := range []string{"inside-markers", "outside-markers"} {
		require.Equal(t, true, layoutEval(t, page, `id=>{const node=document.getElementById(id),s=getComputedStyle(node);return s.display==='inline'&&s.outlineStyle==='none'&&s.marginBottom==='0px'}`, id))
	}
	require.Equal(t, "flow-root", layoutEval(t, page, `()=>getComputedStyle(document.querySelector('#alpha > [data-layout-item-id="a"] > [data-layout-content]')).display`))
}

func TestLayout_AccessibilityStates(t *testing.T) {
	server := layoutFixture(t)
	for _, theme := range []string{"goshtoso", "minimal"} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/dark=%t", theme, dark), func(t *testing.T) {
				page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 1000, Height: 1100}})
				failures := watchPageFailures(page)
				_, err := page.Goto(server.URL)
				require.NoError(t, err)
				layoutReady(t, page)
				layoutEval(t, page, `([theme,dark])=>{document.documentElement.dataset.theme=theme;document.documentElement.classList.toggle('dark',dark)}`, []any{theme, dark})
				root := page.GetByRole("list", playwright.PageGetByRoleOptions{Name: "Alpha", Exact: playwright.Bool(true)})
				count, err := root.Count()
				require.NoError(t, err)
				require.Equal(t, 1, count)
				button := root.GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Move A <unsafe>", Exact: playwright.Bool(true)})
				count, err = button.Count()
				require.NoError(t, err)
				require.Equal(t, 1, count)
				require.NoError(t, page.Keyboard().Press("Tab"))
				require.NoError(t, button.Focus())
				require.Equal(t, true, layoutElementValue(t, button, `el=>el===document.activeElement&&el.matches(':focus-visible')`))
				require.Equal(t, "solid", layoutElementValue(t, button, `el=>getComputedStyle(el).outlineStyle`))
				require.Equal(t, "2px", layoutElementValue(t, button, `el=>getComputedStyle(el).outlineWidth`))
				require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, button, "outlineColor").Ratio, 3.0)
				require.GreaterOrEqual(t, measureRenderedContrast(t, button).Ratio, 4.5)
				require.Equal(t, true, layoutElementValue(t, button, `el=>{const id=el.getAttribute('aria-describedby'),node=document.getElementById(id);return !!node&&!node.hidden&&node.textContent.includes('Alt+Up')&&node.textContent.includes('Escape')&&Array.from(document.querySelectorAll('[id]')).filter(n=>n.id===id).length===1}`))
				require.NoError(t, root.Focus())
				require.Equal(t, true, layoutElementValue(t, root, `el=>el===document.activeElement&&el.matches(':focus-visible')`))
				require.Equal(t, "solid", layoutElementValue(t, root, `el=>getComputedStyle(el).outlineStyle`))
				require.Equal(t, "2px", layoutElementValue(t, root, `el=>getComputedStyle(el).outlineWidth`))
				require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, root, "outlineColor").Ratio, 3.0)
				layoutPickup(t, page, "alpha", "a")
				inbox := page.Locator("#inbox")
				box, err := inbox.BoundingBox()
				require.NoError(t, err)
				require.NoError(t, page.Mouse().Move(box.X+20, box.Y+10))
				require.Equal(t, true, layoutElementValue(t, inbox, `el=>el.hasAttribute('data-layout-drop-target')`))
				require.Equal(t, "solid", layoutElementValue(t, inbox, `el=>getComputedStyle(el).outlineStyle`))
				require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, inbox, "outlineColor").Ratio, 3.0)
				require.NoError(t, page.Keyboard().Press("Escape"))
				require.NoError(t, page.Mouse().Up())
				require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: playwright.ReducedMotionReduce}))
				for _, loc := range []playwright.Locator{root, button} {
					require.Equal(t, "none", layoutElementValue(t, loc, `el=>getComputedStyle(el).transitionProperty`))
					require.Equal(t, "none", layoutElementValue(t, loc, `el=>getComputedStyle(el).animationName`))
				}
				failures.RequireEmpty(t)
			})
		}
	}
}

func layoutElementValue(t *testing.T, locator playwright.Locator, script string) any {
	t.Helper()
	value, err := locator.Evaluate(script, nil)
	require.NoError(t, err)
	return value
}

func TestLayout_ContentPointer(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			server := layoutFixture(t, enabled)
			page := newPage(t, sharedBrowser)
			failures := watchPageFailures(page)
			_, err := page.Goto(server.URL)
			require.NoError(t, err)
			layoutReady(t, page)
			box, err := page.Locator("#alpha p").First().BoundingBox()
			require.NoError(t, err)
			require.NoError(t, page.Mouse().Move(box.X+10, box.Y+5))
			require.NoError(t, page.Mouse().Down())
			layoutDrop(t, page, "#inbox", true)
			want := ""
			if enabled {
				want = "a"
			}
			require.Equal(t, want, layoutEval(t, page, layoutOrder, "inbox"))
			if enabled {
				require.NoError(t, page.Keyboard().Press("Alt+PageUp"))
				require.Equal(t, "b,c,d,a", layoutEval(t, page, layoutOrder, "alpha"))
			}
			failures.RequireEmpty(t)
		})
	}
}
func TestLayout_ContentClicksAndCancellation(t *testing.T) {
	server := layoutFixture(t, true)
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(server.URL)
	require.NoError(t, err)
	layoutReady(t, page)
	layoutEval(t, page, `()=>{window.moves=0;document.addEventListener('layout:move',()=>moves++)}`)
	box, err := page.Locator("#alpha p").First().BoundingBox()
	require.NoError(t, err)
	require.NoError(t, page.Mouse().Move(box.X+10, box.Y+5))
	require.NoError(t, page.Mouse().Down())
	require.NoError(t, page.Mouse().Move(box.X+12, box.Y+7))
	require.NoError(t, page.Mouse().Up())
	require.Equal(t, 0, layoutEval(t, page, `()=>moves`))
	require.Equal(t, "", layoutEval(t, page, `()=>document.querySelector('#alpha [data-layout-status]').textContent`))
	require.NoError(t, page.Locator("#content-link").Click())
	require.Equal(t, "#help", layoutEval(t, page, `()=>location.hash`))
	require.NoError(t, page.Locator("#layout-draft").Fill("still editable"))
	require.Equal(t, "still editable", layoutEval(t, page, `()=>document.getElementById('layout-draft').value`))
	require.NoError(t, page.Mouse().Move(box.X+10, box.Y+5))
	require.NoError(t, page.Mouse().Down())
	target, err := page.Locator("#inbox").BoundingBox()
	require.NoError(t, err)
	require.NoError(t, page.Mouse().Move(target.X+10, target.Y+10, playwright.MouseMoveOptions{Steps: playwright.Int(5)}))
	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, page.Mouse().Up())
	require.Equal(t, 0, layoutEval(t, page, `()=>moves`))
	require.Equal(t, "a,b,c,d", layoutEval(t, page, layoutOrder, "alpha"))
}

func TestLayout_ContentSelectionAndControls(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("selection/%t", enabled), func(t *testing.T) {
			server := layoutFixture(t, enabled)
			page := newPage(t, sharedBrowser)
			_, err := page.Goto(server.URL)
			require.NoError(t, err)
			layoutReady(t, page)
			box, err := page.Locator("#alpha p").First().BoundingBox()
			require.NoError(t, err)
			require.NoError(t, page.Mouse().Move(box.X+2, box.Y+box.Height/2))
			require.NoError(t, page.Mouse().Down())
			require.NoError(t, page.Mouse().Move(box.X+90, box.Y+box.Height/2, playwright.MouseMoveOptions{Steps: playwright.Int(5)}))
			require.NoError(t, page.Mouse().Up())
			require.Equal(t, !enabled, layoutEval(t, page, `()=>getSelection().toString().length>0`))
		})
	}
	server := layoutFixture(t, true)
	for _, id := range []string{"interactive-checkbox", "focusable-content", "media-audio", "media-video"} {
		t.Run(id, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			_, err := page.Goto(server.URL)
			require.NoError(t, err)
			layoutReady(t, page)
			box, err := page.Locator("#" + id).BoundingBox()
			require.NoError(t, err)
			require.NoError(t, page.Mouse().Move(box.X+4, box.Y+4))
			require.NoError(t, page.Mouse().Down())
			layoutDrop(t, page, "#inbox", true)
			require.Equal(t, "a,b,c,d", layoutEval(t, page, layoutOrder, "alpha"))
			require.Equal(t, "", layoutEval(t, page, layoutOrder, "inbox"))
		})
	}
}
func TestLayout_ContentPendingEscape(t *testing.T) {
	server := layoutFixture(t, true)
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(server.URL)
	require.NoError(t, err)
	layoutReady(t, page)
	layoutEval(t, page, `()=>document.addEventListener('keydown',e=>{if(e.key==='Escape')window.escapePrevented=e.defaultPrevented})`)
	box, err := page.Locator("#alpha p").First().BoundingBox()
	require.NoError(t, err)
	require.NoError(t, page.Mouse().Move(box.X+5, box.Y+5))
	require.NoError(t, page.Mouse().Down())
	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, page.Mouse().Up())
	require.Equal(t, false, layoutEval(t, page, `()=>escapePrevented`))
	require.Equal(t, "", layoutEval(t, page, `()=>document.querySelector('#alpha [data-layout-status]').textContent`))
}
func TestLayout_NestedContentOwnership(t *testing.T) {
	server := layoutFixture(t)
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(server.URL)
	require.NoError(t, err)
	layoutReady(t, page)
	layoutEval(t, page, `()=>['inner','outer'].forEach(id=>document.getElementById(id).dataset.layoutContentDrag='true')`)
	require.NoError(t, page.Locator("#outer").ScrollIntoViewIfNeeded())
	box, err := page.Locator("#inner p").BoundingBox()
	require.NoError(t, err)
	require.NoError(t, page.Mouse().Move(box.X+5, box.Y+5))
	require.NoError(t, page.Mouse().Down())
	layoutDrop(t, page, `#inner > [data-layout-item-id="inner-b"]`, false)
	require.Equal(t, "inner-b,inner-a", layoutEval(t, page, layoutOrder, "inner"))
	require.Equal(t, "outer-a,outer-b", layoutEval(t, page, layoutOrder, "outer"))
	box, err = page.Locator(`#outer > [data-layout-item-id="outer-b"] p`).BoundingBox()
	require.NoError(t, err)
	require.NoError(t, page.Mouse().Move(box.X+5, box.Y+5))
	require.NoError(t, page.Mouse().Down())
	layoutDrop(t, page, "#inner", true)
	require.Equal(t, "outer-b,outer-a", layoutEval(t, page, layoutOrder, "outer"))
	require.Equal(t, "inner-b,inner-a", layoutEval(t, page, layoutOrder, "inner"))
}

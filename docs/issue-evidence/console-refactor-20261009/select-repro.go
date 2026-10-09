package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/head"
	goselect "github.com/araihu/goshtoso/components/select"
	"github.com/mxschmitt/playwright-go"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	var requests atomic.Int64
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.HandleFunc("GET /filter", func(w http.ResponseWriter, r *http.Request) {
		n := requests.Add(1)
		fmt.Fprintf(w, "Filter request %d: type=%s", n, r.URL.Query().Get("type"))
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><html lang="en" class="dark" data-theme="dracula"><head><meta charset="utf-8"><title>Select event reproduction</title>`)
		must(head.Dependencies(head.WithLocalRuntime()).Render(context.Background(), w))
		fmt.Fprint(w, `</head><body class="bg-surface dark:bg-surface-dark text-on-surface dark:text-on-surface-dark"><main style="max-width:780px;margin:56px auto;padding:24px;font-family:system-ui,sans-serif"><h1 style="font-size:28px;font-weight:700;margin-bottom:12px">Select + HTMX change event</h1><p style="margin-bottom:28px">Goshtoso v0.3.3 · public components · no consumer watcher</p><form id="filters">`)
		must(goselect.Select(goselect.Config{
			ID: "allocation-type", Name: "type", Label: "Allocation type",
			Options:    []goselect.Option{{Value: "", Label: "All types", Selected: true}, {Value: "allocated", Label: "Allocated"}, {Value: "reserved", Label: "Reserved"}},
			InputAttrs: templ.Attributes{"hx-get": "/filter", "hx-trigger": "change", "hx-target": "#result", "hx-include": "closest form"},
		}).Render(context.Background(), w))
		fmt.Fprint(w, `</form><section style="margin-top:30px;border:1px solid #64748b;border-radius:12px;padding:20px;line-height:2"><p>Submitted field value: <strong id="value">empty</strong></p><p>Bubbling change events: <strong id="changes">0</strong></p><p>HTMX result: <strong id="result">No filter request received</strong></p></section><p style="margin-top:20px">A selection should notify the filter without accessing component internals.</p></main><script>document.addEventListener('change',e=>{if(e.target.name==='type'){const c=document.getElementById('changes');c.textContent=String(Number(c.textContent)+1)}});</script></body></html>`)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	pw, err := playwright.Run()
	must(err)
	defer pw.Stop()
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})
	must(err)
	defer browser.Close()
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 920, Height: 560}})
	must(err)
	_, err = page.Goto(server.URL)
	must(err)
	_, err = page.WaitForFunction("() => !!window.Alpine && !!window.htmx && typeof window.goshtosoSelect === 'function'", nil)
	must(err)
	must(page.GetByRole("combobox", playwright.PageGetByRoleOptions{Name: "Allocation type", Exact: playwright.Bool(true)}).Click())
	must(page.GetByRole("option", playwright.PageGetByRoleOptions{Name: "Reserved", Exact: playwright.Bool(true)}).WaitFor())
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String("/tmp/handoffs/tks-console-segunda-rodada-assets/select-options.png")})
	must(err)
	must(page.GetByRole("option", playwright.PageGetByRoleOptions{Name: "Reserved", Exact: playwright.Bool(true)}).Click())
	_, err = page.WaitForFunction("() => document.querySelector('input[name=type]').value === 'reserved'", nil)
	must(err)
	_, err = page.WaitForFunction("() => Array.from(document.querySelectorAll('[role=option]')).every(el => el.getClientRects().length === 0)", nil)
	must(err)
	_, err = page.Evaluate("() => { document.getElementById('value').textContent = new FormData(document.getElementById('filters')).get('type'); return document.getElementById('changes').textContent; }")
	must(err)
	if requests.Load() != 0 {
		panic("expected reproduction of missing change event")
	}
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String("/tmp/handoffs/tks-console-segunda-rodada-assets/select-no-change.png")})
	must(err)
	fmt.Println("Confirmed: selected value reserved, native change count 0, HTMX filter requests 0")
}

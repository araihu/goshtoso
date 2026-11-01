//go:build e2e && webkit

package e2e

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/modal"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestDialogContentWebKit(t *testing.T) {
	browser, err := sharedPW.WebKit.Launch()
	require.NoError(t, err)
	t.Cleanup(func() { _ = browser.Close() })
	for _, tc := range []struct {
		name       string
		width      int
		compact    bool
		fullscreen bool
	}{
		{"compact-mobile", 390, true, false},
		{"default-desktop", 1440, false, false},
		{"fullscreen-mobile", 390, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var html bytes.Buffer
			html.WriteString(`<!doctype html><html><head><meta name="viewport" content="width=device-width, initial-scale=1">`)
			require.NoError(t, head.Dependencies(head.WithLocalRuntime()).Render(t.Context(), &html))
			html.WriteString(`</head><body><main x-data="{}">`)
			require.NoError(t, button.Button(button.WithID("open-dialog"), button.WithAttrs(templ.Attributes{
				"aria-label": "Open dialog", "x-on:click": "$dispatch('modal:open', {id:'dialog-probe'})",
			})).Render(t.Context(), &html))
			require.NoError(t, modal.Dialog(modal.DialogConfig{
				ID: "dialog-probe", Title: "Search", Compact: tc.compact, FullscreenMobile: tc.fullscreen,
				Content: templ.Join(textinput.TextInput(textinput.Config{ID: "dialog-query", Type: textinput.TypeSearch, Label: "Search", InputAttrs: templ.Attributes{"autofocus": true}}), templ.Raw(`<ul style="max-height:55dvh;overflow-y:auto">`+strings.Repeat("<li>Search result</li>", 33)+"</ul>")),
			}).Render(t.Context(), &html))
			html.WriteString(`</main></body></html>`)
			mux := http.NewServeMux()
			mux.Handle("GET /assets/", assets.Handler())
			mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = io.WriteString(w, html.String())
			})
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)
			page := newPage(t, browser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: tc.width, Height: 844}})
			_, err := page.Goto(server.URL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})
			require.NoError(t, err)
			opener := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Open dialog"})
			require.NoError(t, opener.Focus())
			require.NoError(t, opener.Press("Enter"))
			dialog := page.Locator("#dialog-probe")
			require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
			visible, err := dialog.Evaluate(`el => {
				const body = el.lastElementChild.getBoundingClientRect();
				const input = el.querySelector('input').getBoundingClientRect();
				return body.height >= input.height && input.top >= body.top && input.bottom <= body.bottom && body.bottom <= innerHeight;
			}`, nil)
			require.NoError(t, err)
			require.Equal(t, true, visible, "dialog must expose its search field instead of clipping its body")
			require.NoError(t, page.Locator("#dialog-query").Fill("Grafana"))
			require.NoError(t, page.Keyboard().Press("Escape"))
			require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
			_, err = page.WaitForFunction(`document.activeElement.id === 'open-dialog'`, nil)
			require.NoError(t, err, "keyboard dismissal must restore the previously focused opener")
		})
	}
}

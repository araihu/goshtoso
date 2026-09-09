//go:build e2e && (full || table)

package e2e

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/table"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTableLinkedRowNestedControls(t *testing.T) {
	for _, mode := range []table.LinkMode{table.LinkFull, table.LinkSPA, table.LinkBoost} {
		for _, theme := range []struct {
			name string
			dark bool
		}{{"goshtoso", false}, {"goshtoso", true}, {"minimal", false}, {"minimal", true}} {
			t.Run(string(mode)+"/"+theme.name+map[bool]string{true: "-dark", false: "-light"}[theme.dark], func(t *testing.T) {
				page := newPage(t, sharedBrowser)
				require.NoError(t, page.Route("**/row-destination", func(route playwright.Route) {
					require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{Status: playwright.Int(200), ContentType: playwright.String("text/html"), Body: "<html><body>Destination</body></html>"}))
				}))
				document := renderInteractiveDocument(t, head.DependenciesMinimal(head.WithLocalRuntime()), table.Table(table.Config{
					ID: "main-content-area", ShowCheckbox: true, Columns: []table.Column{{Key: "name", Label: "Name"}, {Key: "controls", Label: "Controls"}},
					Rows: []table.Row{{ID: "app", Link: "/row-destination", LinkMode: mode, Cells: map[string]table.Cell{
						"name": {Text: "Application"}, "controls": {Component: templ.Raw(`<label for="pin">Pinned</label><input id="pin" type="checkbox"><button id="nested-button" onclick="this.dataset.clicked='true'">Action</button><a id="nested-link" href="#nested">Address</a>`)},
					}}},
				}))
				require.NoError(t, page.Route("**/row-fixture", func(route playwright.Route) {
					require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{Status: playwright.Int(200), ContentType: playwright.String("text/html"), Body: document}))
				}))
				_, err := page.Goto(baseURL + "/row-fixture")
				require.NoError(t, err)
				_, err = page.WaitForFunction(`() => typeof window.goshtosoTableRowLinkEvent === 'function' && typeof window.htmx !== 'undefined'`, nil)
				require.NoError(t, err)
				setThemeMode(t, page, theme.name, theme.dark)
				// Count HTMX row requests before cancelling them, keeping the fixture in place.
				_, err = page.Evaluate(`() => {window.rowRequests=0;document.addEventListener('htmx:beforeRequest',e=>{if(e.detail.elt.hasAttribute('data-table-row-link')){window.rowRequests++;e.preventDefault();}});}`)
				require.NoError(t, err)
				require.NoError(t, page.Locator("label[for=pin]").Click())
				checked, err := page.Locator("#pin").IsChecked()
				require.NoError(t, err)
				require.True(t, checked)
				require.NoError(t, page.Locator("#nested-button").Click())
				clicked, err := page.Locator("#nested-button").GetAttribute("data-clicked")
				require.NoError(t, err)
				require.Equal(t, "true", clicked)
				require.NoError(t, page.Locator("#pin").Press("Space"))
				checked, err = page.Locator("#pin").IsChecked()
				require.NoError(t, err)
				require.False(t, checked)
				require.NoError(t, page.Locator("#main-content-area tbody input").First().Check())
				require.NoError(t, page.Locator("#nested-link").Click())
				require.Contains(t, page.URL(), "#nested")
				count, err := page.Evaluate(`() => window.rowRequests`)
				require.NoError(t, err)
				require.EqualValues(t, 0, count)
				if mode == table.LinkFull {
					require.NoError(t, page.Locator("#main-content-area tbody tr").Press("Enter"))
					require.NoError(t, page.WaitForURL("**/row-destination"))
				} else {
					require.NoError(t, page.Locator("#main-content-area tbody tr td").Nth(1).Click())
					_, err = page.WaitForFunction(`() => window.rowRequests === 1`, nil)
					require.NoError(t, err)
				}
			})
		}
	}
}

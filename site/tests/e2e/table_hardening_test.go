package e2e

import (
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/table"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestTableExpandableRowIDEscapesAlpineExpressions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	dialogSeen := listenForDialogs(t, page)

	_, err := page.Goto(baseURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	require.NoError(t, err)

	html := renderInteractiveDocument(t, head.Dependencies(), table.Table(table.Config{
		ID: "security-table",
		Columns: []table.Column{
			{Key: "name", Label: "Name"},
		},
		Rows: []table.Row{
			{
				ID:         `row-1'];alert("table-row-id-xss");//`,
				Expandable: true,
				Detail:     templ.Raw(`<div>Secret details</div>`),
				Cells: map[string]table.Cell{
					"name": {Text: "Alice"},
				},
			},
		},
	}))
	require.NoError(t, page.SetContent(html, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	}))
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil)
	require.NoError(t, err)

	requireNoDialog(t, dialogSeen, "table row ID executed during Alpine init")

	require.NoError(t, page.Locator("#security-table tbody tr").First().Click())
	requireNoDialog(t, dialogSeen, "table row ID executed on expandable row click")
}

func renderInteractiveDocument(t *testing.T, components ...templ.Component) string {
	t.Helper()

	var rendered []string
	for _, component := range components {
		rendered = append(rendered, renderComponentFragment(t, component))
	}

	var headHTML []string
	var bodyHTML []string
	for _, fragment := range rendered {
		switch {
		case strings.Contains(fragment, "<link") || strings.Contains(fragment, "<script"):
			headHTML = append(headHTML, fragment)
		default:
			bodyHTML = append(bodyHTML, fragment)
		}
	}

	return "<!doctype html><html><head>" + strings.Join(headHTML, "") + "</head><body>" + strings.Join(bodyHTML, "") + "</body></html>"
}

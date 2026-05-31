package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// tickerCellText returns the current text of a ticker price cell once it has
// rendered a price.
func tickerCellText(t *testing.T, page playwright.Page, symbol string) string {
	t.Helper()
	sel := "#ticker-cell-" + symbol
	_, err := page.WaitForFunction(
		"(s) => { const el = document.querySelector(s); return el && el.textContent.trim().length > 0; }",
		sel,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(6000)})
	require.NoError(t, err)
	txt, err := page.Locator(sel).TextContent()
	require.NoError(t, err)
	return txt
}

func TestTicker_StreamUpdatesCells(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/examples/ticker")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	before := tickerCellText(t, page, "AAPL")
	_, err = page.WaitForFunction(
		"(b) => { const el = document.querySelector('#ticker-cell-AAPL'); return el && el.textContent.trim() !== b; }",
		before,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(6000)})
	require.NoError(t, err, "price cell should change as SSE ticks arrive")
}

func TestTicker_SpotlightOnRowClick(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/examples/ticker")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	tickerCellText(t, page, "MSFT")

	// The table <tr> does not receive id={row.ID} — Row.ID is only used in
	// Alpine x-on:click expressions. Use :has() to target the MSFT row via its
	// price cell child.
	require.NoError(t, page.Locator("#ticker-table tr:has(#ticker-cell-MSFT)").First().Click())
	_, err = page.WaitForFunction(
		"() => { const el = document.querySelector('#ticker-spotlight'); return el && el.textContent.includes('Microsoft'); }",
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(4000)})
	require.NoError(t, err, "spotlight should show the clicked symbol")
	require.NoError(t, page.Locator("#ticker-card-MSFT").WaitFor())
}

func TestTicker_PauseStopsUpdates(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/examples/ticker")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	tickerCellText(t, page, "AAPL")

	require.NoError(t, page.Locator("label[for='ticker-pause']").Click())
	_, err = page.WaitForFunction(
		"() => Alpine.$data(document.querySelector('#ticker-fragment')).paused === true", nil)
	require.NoError(t, err)
	paused := tickerCellText(t, page, "AAPL")

	stable, err := page.WaitForFunction(
		"(b) => { const el = document.querySelector('#ticker-cell-AAPL'); return el && el.textContent.trim() === b; }",
		paused,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
	require.NotNil(t, stable)

	require.NoError(t, page.Locator("label[for='ticker-pause']").Click())
	_, err = page.WaitForFunction(
		"(b) => { const el = document.querySelector('#ticker-cell-AAPL'); return el && el.textContent.trim() !== b; }",
		paused,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(6000)})
	require.NoError(t, err, "updates should resume after unpausing")
}

func TestTicker_FragmentNavNoErrors(t *testing.T) {
	page := newPage(t, sharedBrowser)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("a[href='/examples/ticker']").First().Click())
	before := tickerCellText(t, page, "AAPL")
	_, err = page.WaitForFunction(
		"(b) => { const el = document.querySelector('#ticker-cell-AAPL'); return el && el.textContent.trim() !== b; }",
		before,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(6000)})
	require.NoError(t, err, "SSE should update cells after fragment nav")

	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav ticker page: %v", jsErrors)
}

//go:build e2e

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

func TestTicker_ProviderBootstrapsFirstPaint(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/examples/ticker")
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => {
		const root = document.querySelector('#ticker-fragment');
		return typeof Alpine !== 'undefined' && Alpine.__tickerPaneRegistered === true &&
			root && root._x_dataStack && typeof Alpine.$data(root).connect === 'function';
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "external tickerPane provider should initialize on first paint")
}

// TestTicker_StreamUpdatesCells loads the page and asserts a table price cell
// changes as ticks stream in.
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
	require.NoError(t, err, "table price cell should change as SSE ticks arrive")
}

// TestTicker_CardsUpdateLive asserts the top card grid updates off the same
// shared SSE stream as the table.
func TestTicker_CardsUpdateLive(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/examples/ticker")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("#ticker-card-MSFT").WaitFor())
	before, err := page.Locator("#ticker-card-MSFT").TextContent()
	require.NoError(t, err)
	_, err = page.WaitForFunction(
		"(b) => { const el = document.querySelector('#ticker-card-MSFT'); return el && el.textContent.trim() !== b; }",
		before,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(6000)})
	require.NoError(t, err, "the top card should update live off the SSE stream")
}

// TestTicker_TableFilterServerSide types a query, asserts the server returns only
// matching rows, and that the filtered-in cell keeps updating after the swap.
func TestTicker_TableFilterServerSide(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/examples/ticker")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	tickerCellText(t, page, "AAPL") // table rendered with all rows

	// Fill the ticker table's own search input (the shared fillSearchInput helper
	// is hardcoded to the components-demo table id). Dispatch input so Alpine's
	// x-model + debounced applyFilters() fires the htmx.ajax to the rows endpoint.
	search := page.Locator("#ticker-table-filters-search")
	require.NoError(t, search.Fill("NVDA"))
	_, err = search.Evaluate(`(el) => el.dispatchEvent(new Event('input', {bubbles: true}))`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(
		"() => { const b = document.querySelector('#ticker-table-tbody'); "+
			"return b && b.querySelector('#ticker-cell-NVDA') && !b.querySelector('#ticker-cell-AAPL'); }",
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(4000)})
	require.NoError(t, err, "server filter should leave only NVDA in the table")

	// The filtered-in cell rebinds to the stream and keeps updating.
	before, err := page.Locator("#ticker-cell-NVDA").TextContent()
	require.NoError(t, err)
	_, err = page.WaitForFunction(
		"(b) => { const el = document.querySelector('#ticker-cell-NVDA'); return el && el.textContent.trim() !== b; }",
		before,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(6000)})
	require.NoError(t, err, "filtered row should still update live after the swap")
}

// TestTicker_PauseStopsUpdates verifies pause freezes the cells and resume
// restarts them. Pause cancels SSE swaps; it does not close the stream.
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

// TestTicker_FragmentNavNoErrors lands elsewhere, navigates to the ticker via the
// sidebar (htmx fragment swap), and asserts no console/page errors plus a live
// cell update — the SPA path a direct load can't catch.
func TestTicker_FragmentNavNoErrors(t *testing.T) {
	page := newPage(t, sharedBrowser)
	dismissCookieBanner(t, page)

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

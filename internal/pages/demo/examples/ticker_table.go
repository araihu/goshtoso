package examples

import (
	"fmt"

	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/internal/examples/ticker"
)

// tickerPrice formats a symbol's price as a fixed 2-decimal string.
func tickerPrice(s ticker.Symbol) string {
	return fmt.Sprintf("%.2f", s.Price)
}

// tickerChange formats the percent change with a sign, e.g. "+0.42%".
func tickerChange(s ticker.Symbol) string {
	return fmt.Sprintf("%+.2f%%", s.ChangePct())
}

// tickerBadgeVariant maps price direction to a badge color.
func tickerBadgeVariant(s ticker.Symbol) badge.Variant {
	switch s.Direction() {
	case "up":
		return badge.Success
	case "down":
		return badge.Danger
	default:
		return badge.Default
	}
}

// tickerRows builds the table rows for the given symbols. Each price cell renders
// a span that subscribes to its symbol's SSE event (via TickerCell), so rows
// stay live after a filter swap once htmx re-binds the swapped-in spans.
func tickerRows(symbols []ticker.Symbol) []table.Row {
	rows := make([]table.Row, 0, len(symbols))
	for _, sym := range symbols {
		rows = append(rows, table.Row{
			ID: sym.Ticker,
			Cells: map[string]table.Cell{
				"symbol": {Text: sym.Ticker, Description: sym.Name},
				"price":  {Component: TickerCell(sym)},
			},
		})
	}
	return rows
}

// TickerTableConfig builds the bottom live board: a filterable table whose
// search box posts to the rows endpoint for a server-side filter. Exported so
// the rows handler can reuse the same Columns when rendering filtered rows.
func TickerTableConfig(symbols []ticker.Symbol) table.Config {
	return table.Config{
		ID:           "ticker-table",
		HTMXEndpoint: "/api/examples/ticker/rows",
		Columns: []table.Column{
			{Key: "symbol", Label: "Symbol"},
			{Key: "price", Label: "Price", Align: "right"},
		},
		Rows: tickerRows(symbols),
		Filters: &table.FilterConfig{
			Filters: []table.Filter{
				{
					Key:         "search",
					Label:       "Filter",
					Type:        table.FilterSearch,
					Placeholder: "Search symbol or name…",
				},
			},
		},
	}
}

// tickerTableConfig is the unexported alias used by the templ view.
func tickerTableConfig(symbols []ticker.Symbol) table.Config {
	return TickerTableConfig(symbols)
}

// tickerPaneJS registers the Alpine component for the ticker pane. `connected`
// hides the spinner once the first SSE message arrives; `paused` cancels swaps
// via the cancelable htmx:sseBeforeMessage event so the connection stays open
// while paused. NOTE: htmx-ext-sse@2.2.3 has NO htmx:sseOpen event, so we set
// connected on the first sseBeforeMessage and clear it on sseError.
const tickerPaneJS = `(() => {
	const register = () => {
		Alpine.data('tickerPane', () => ({
			connected: false,
			paused: false,
			connect(el) {
				if (!el) return;
				el.addEventListener('htmx:sseBeforeMessage', (e) => {
					this.connected = true;
					if (this.paused) e.preventDefault();
				});
				el.addEventListener('htmx:sseError', () => { this.connected = false; });
			},
		}));
	};
	if (window.Alpine && window.Alpine.version) {
		register();
	} else {
		document.addEventListener('alpine:init', register);
	}
})();`

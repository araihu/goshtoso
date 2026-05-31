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

// tickerTableConfig builds the live board: each price cell renders a span that
// subscribes to its symbol's SSE event; each row's click loads the spotlight.
func tickerTableConfig(symbols []ticker.Symbol) table.Config {
	rows := make([]table.Row, 0, len(symbols))
	for _, sym := range symbols {
		rows = append(rows, table.Row{
			ID:       sym.Ticker,
			HXGet:    "/api/examples/ticker/spotlight?symbol=" + sym.Ticker,
			HXTarget: "#ticker-spotlight",
			HXSwap:   "innerHTML",
			Cells: map[string]table.Cell{
				"symbol": {Text: sym.Ticker, Description: sym.Name},
				"price":  {Component: TickerCell(sym)},
			},
		})
	}
	return table.Config{
		ID: "ticker-table",
		Columns: []table.Column{
			{Key: "symbol", Label: "Symbol"},
			{Key: "price", Label: "Price", Align: "right"},
		},
		Rows: rows,
	}
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

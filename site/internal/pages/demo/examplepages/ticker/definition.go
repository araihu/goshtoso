// Package tickerpage owns the Live Ticker runnable example page.
package tickerpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Live Ticker example's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:         "examples/ticker",
	Title:       "Live Ticker",
	Active:      "ticker",
	Description: "Explore a simulated market watchlist with Server-Sent Events, pause and resume controls, and a sortable live table.",
	Type:        "SoftwareSourceCode",
	Content:     TickerContent,
}

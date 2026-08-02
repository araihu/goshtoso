// Package tickerpage owns the Live Ticker runnable example page.
package tickerpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Live Ticker example's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "examples/ticker",
	Title:   "Live Ticker",
	Active:  "ticker",
	Type:    "SoftwareSourceCode",
	Content: TickerContent,
}

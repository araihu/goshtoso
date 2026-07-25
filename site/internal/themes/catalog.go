// Package themes owns the public theme inventory used by the demo site's
// selectors, counts, profile example, and theme previews.
package themes

import "slices"

// Theme is one built-in Goshtoso theme.
type Theme struct {
	Key   string
	Label string
}

var catalog = []Theme{
	{Key: "goshtoso", Label: "Goshtoso"},
	{Key: "minimal", Label: "Minimal"},
	{Key: "modern", Label: "Modern"},
	{Key: "arctic", Label: "Arctic"},
	{Key: "high-contrast", Label: "High Contrast"},
	{Key: "neo-brutalism", Label: "Neo Brutalism"},
	{Key: "news", Label: "News"},
	{Key: "industrial", Label: "Industrial"},
	{Key: "90s", Label: "90s"},
	{Key: "pastel", Label: "Pastel"},
	{Key: "christmas", Label: "Christmas"},
	{Key: "halloween", Label: "Halloween"},
	{Key: "zombie", Label: "Halloween II"},
	{Key: "prototype", Label: "Prototype"},
	{Key: "dracula", Label: "Dracula"},
}

// All returns the built-in themes in stable product order.
func All() []Theme {
	return slices.Clone(catalog)
}

// Count returns the number of built-in themes.
func Count() int {
	return len(catalog)
}

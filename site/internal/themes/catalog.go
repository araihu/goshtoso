// Package themes owns the theme presentation order used by the demo site's
// selectors, counts, profile example, and theme previews.
//
// This compatibility catalog remains local while site/go.mod pins v0.1.12,
// which predates the public root themes package. After that pin is bumped to a
// release containing the package, derive built-in keys from the root catalog
// while retaining this site's presentation labels, order, and defaults.
package themes

import "slices"

// Theme is one built-in Goshtoso theme.
type Theme struct {
	Key   string
	Label string
}

// ZombiePresentationLabelOverride is the demo's intentional presentation copy
// for the root catalog's canonical "Zombie" label.
const ZombiePresentationLabelOverride = "Halloween II"

var catalog = []Theme{
	{Key: "araihu", Label: "Arai Hû"},
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
	{Key: "zombie", Label: ZombiePresentationLabelOverride},
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

// PresentationLabelOverride returns site-owned copy that intentionally differs
// from a canonical root catalog label.
func PresentationLabelOverride(key string) (string, bool) {
	if key == "zombie" {
		return ZombiePresentationLabelOverride, true
	}
	return "", false
}

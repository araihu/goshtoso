// Package themes owns the theme presentation order used by the demo site's
// selectors, counts, profile example, and theme previews.
//
// This compatibility catalog remains local while site/go.mod pins v0.1.12,
// which predates the public root themes package. After that pin is bumped to a
// release containing the package, derive built-in keys from the root catalog
// while retaining this site's presentation labels, order, and defaults.
package themes

import "slices"

// Ownership classifies built-in themes without assigning ownership of their
// CSS token values.
type Ownership string

const (
	OwnershipGeneric      Ownership = "generic"
	OwnershipOrganization Ownership = "organization"
)

// Theme is one built-in Goshtoso theme.
type Theme struct {
	Key       string
	Label     string
	Ownership Ownership
}

// ZombiePresentationLabelOverride is the demo's intentional presentation copy
// for the root catalog's canonical "Zombie" label.
const ZombiePresentationLabelOverride = "Halloween II"

var catalog = []Theme{
	{Key: "araihu", Label: "Arai Hû", Ownership: OwnershipOrganization},
	{Key: "goshtoso", Label: "Goshtoso", Ownership: OwnershipOrganization},
	{Key: "minimal", Label: "Minimal", Ownership: OwnershipGeneric},
	{Key: "modern", Label: "Modern", Ownership: OwnershipGeneric},
	{Key: "arctic", Label: "Arctic", Ownership: OwnershipGeneric},
	{Key: "high-contrast", Label: "High Contrast", Ownership: OwnershipGeneric},
	{Key: "neo-brutalism", Label: "Neo Brutalism", Ownership: OwnershipGeneric},
	{Key: "news", Label: "News", Ownership: OwnershipGeneric},
	{Key: "industrial", Label: "Industrial", Ownership: OwnershipGeneric},
	{Key: "90s", Label: "90s", Ownership: OwnershipGeneric},
	{Key: "pastel", Label: "Pastel", Ownership: OwnershipGeneric},
	{Key: "christmas", Label: "Christmas", Ownership: OwnershipGeneric},
	{Key: "halloween", Label: "Halloween", Ownership: OwnershipGeneric},
	{Key: "zombie", Label: ZombiePresentationLabelOverride, Ownership: OwnershipGeneric},
	{Key: "prototype", Label: "Prototype", Ownership: OwnershipGeneric},
	{Key: "dracula", Label: "Dracula", Ownership: OwnershipGeneric},
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

// Package themes describes themes built into Goshtoso's compiled stylesheet.
// Built-in keys and canonical labels are stable release API. Consumers remain
// responsible for presentation order, defaults, and custom themes.
package themes

import "slices"

// Key identifies a built-in Goshtoso theme.
type Key string

// Ownership classifies a theme as generic design-system styling or an Arai Hû
// organization identity. It does not assign ownership of CSS token values.
type Ownership string

const (
	OwnershipGeneric      Ownership = "generic"
	OwnershipOrganization Ownership = "organization"
)

const (
	Key90s          Key = "90s"
	KeyAraiHu       Key = "araihu"
	KeyArctic       Key = "arctic"
	KeyChristmas    Key = "christmas"
	KeyDracula      Key = "dracula"
	KeyGoshtoso     Key = "goshtoso"
	KeyHalloween    Key = "halloween"
	KeyHighContrast Key = "high-contrast"
	KeyIndustrial   Key = "industrial"
	KeyMinimal      Key = "minimal"
	KeyModern       Key = "modern"
	KeyNeoBrutalism Key = "neo-brutalism"
	KeyNews         Key = "news"
	KeyPastel       Key = "pastel"
	KeyPrototype    Key = "prototype"
	KeyZombie       Key = "zombie"
)

// Theme identifies one theme built into Goshtoso, its canonical design-system
// label, and its source classification. It deliberately excludes CSS tokens,
// defaults, and presentation metadata.
type Theme struct {
	Key       Key
	Label     string
	Ownership Ownership
}

var builtIn = [...]Theme{
	{Key: Key90s, Label: "90s", Ownership: OwnershipGeneric},
	{Key: KeyAraiHu, Label: "Arai Hû", Ownership: OwnershipOrganization},
	{Key: KeyArctic, Label: "Arctic", Ownership: OwnershipGeneric},
	{Key: KeyChristmas, Label: "Christmas", Ownership: OwnershipGeneric},
	{Key: KeyDracula, Label: "Dracula", Ownership: OwnershipGeneric},
	{Key: KeyGoshtoso, Label: "Goshtoso", Ownership: OwnershipOrganization},
	{Key: KeyHalloween, Label: "Halloween", Ownership: OwnershipGeneric},
	{Key: KeyHighContrast, Label: "High Contrast", Ownership: OwnershipGeneric},
	{Key: KeyIndustrial, Label: "Industrial", Ownership: OwnershipGeneric},
	{Key: KeyMinimal, Label: "Minimal", Ownership: OwnershipGeneric},
	{Key: KeyModern, Label: "Modern", Ownership: OwnershipGeneric},
	{Key: KeyNeoBrutalism, Label: "Neo Brutalism", Ownership: OwnershipGeneric},
	{Key: KeyNews, Label: "News", Ownership: OwnershipGeneric},
	{Key: KeyPastel, Label: "Pastel", Ownership: OwnershipGeneric},
	{Key: KeyPrototype, Label: "Prototype", Ownership: OwnershipGeneric},
	{Key: KeyZombie, Label: "Zombie", Ownership: OwnershipGeneric},
}

// BuiltIn returns every built-in theme in deterministic key order. This order
// supports reproducible traversal; it is not presentation order and does not
// select a default. The returned slice is caller-owned.
func BuiltIn() []Theme {
	return slices.Clone(builtIn[:])
}

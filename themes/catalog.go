// Package themes describes themes built into Goshtoso's compiled stylesheet.
// Built-in keys and canonical labels are stable release API. Consumers remain
// responsible for presentation order, defaults, and custom themes.
package themes

import "slices"

// Key identifies a built-in Goshtoso theme.
type Key string

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

// Theme identifies one theme built into Goshtoso and its canonical
// design-system label. It deliberately excludes CSS tokens, defaults, and
// presentation metadata.
type Theme struct {
	Key   Key
	Label string
}

var builtIn = [...]Theme{
	{Key: Key90s, Label: "90s"},
	{Key: KeyAraiHu, Label: "Arai Hû"},
	{Key: KeyArctic, Label: "Arctic"},
	{Key: KeyChristmas, Label: "Christmas"},
	{Key: KeyDracula, Label: "Dracula"},
	{Key: KeyGoshtoso, Label: "Goshtoso"},
	{Key: KeyHalloween, Label: "Halloween"},
	{Key: KeyHighContrast, Label: "High Contrast"},
	{Key: KeyIndustrial, Label: "Industrial"},
	{Key: KeyMinimal, Label: "Minimal"},
	{Key: KeyModern, Label: "Modern"},
	{Key: KeyNeoBrutalism, Label: "Neo Brutalism"},
	{Key: KeyNews, Label: "News"},
	{Key: KeyPastel, Label: "Pastel"},
	{Key: KeyPrototype, Label: "Prototype"},
	{Key: KeyZombie, Label: "Zombie"},
}

// BuiltIn returns every built-in theme in deterministic key order. This order
// supports reproducible traversal; it is not presentation order and does not
// select a default. The returned slice is caller-owned.
func BuiltIn() []Theme {
	return slices.Clone(builtIn[:])
}

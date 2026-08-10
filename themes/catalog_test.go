package themes

import (
	"slices"
	"testing"
)

func TestBuiltInReturnsStableUniqueCatalogInKeyOrder(t *testing.T) {
	want := []Theme{
		{Key: "90s", Label: "90s"},
		{Key: "araihu", Label: "Arai Hû"},
		{Key: "arctic", Label: "Arctic"},
		{Key: "christmas", Label: "Christmas"},
		{Key: "dracula", Label: "Dracula"},
		{Key: "goshtoso", Label: "Goshtoso"},
		{Key: "halloween", Label: "Halloween"},
		{Key: "high-contrast", Label: "High Contrast"},
		{Key: "industrial", Label: "Industrial"},
		{Key: "minimal", Label: "Minimal"},
		{Key: "modern", Label: "Modern"},
		{Key: "neo-brutalism", Label: "Neo Brutalism"},
		{Key: "news", Label: "News"},
		{Key: "pastel", Label: "Pastel"},
		{Key: "prototype", Label: "Prototype"},
		{Key: "zombie", Label: "Zombie"},
	}

	got := BuiltIn()
	seen := make(map[Key]struct{}, len(got))
	for _, theme := range got {
		if theme.Key == "" || theme.Label == "" {
			t.Fatalf("BuiltIn() contains empty key or label: %#v", theme)
		}
		if _, exists := seen[theme.Key]; exists {
			t.Fatalf("BuiltIn() contains duplicate key %q", theme.Key)
		}
		seen[theme.Key] = struct{}{}
	}
	if !slices.Equal(got, want) {
		t.Fatalf("BuiltIn() = %#v, want %#v", got, want)
	}
}

func TestBuiltInReturnsCallerOwnedCatalog(t *testing.T) {
	first := BuiltIn()
	first[0] = Theme{Key: "consumer-custom", Label: "Consumer Custom"}
	first = append(first, Theme{Key: "another-custom", Label: "Another Custom"})
	if len(first) != 17 {
		t.Fatalf("len(mutated caller catalog) = %d, want 17", len(first))
	}

	second := BuiltIn()
	if len(second) != 16 {
		t.Fatalf("len(BuiltIn()) = %d after caller mutation, want 16", len(second))
	}
	if second[0] != (Theme{Key: "90s", Label: "90s"}) {
		t.Fatalf("BuiltIn()[0] = %#v after caller mutation, want stable built-in", second[0])
	}
}

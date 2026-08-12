package themes

import (
	"slices"
	"testing"
)

func TestBuiltInReturnsStableUniqueCatalogInKeyOrder(t *testing.T) {
	want := []Theme{
		{Key: "90s", Label: "90s", Ownership: OwnershipGeneric},
		{Key: "araihu", Label: "Arai Hû", Ownership: OwnershipOrganization},
		{Key: "arctic", Label: "Arctic", Ownership: OwnershipGeneric},
		{Key: "christmas", Label: "Christmas", Ownership: OwnershipGeneric},
		{Key: "dracula", Label: "Dracula", Ownership: OwnershipGeneric},
		{Key: "goshtoso", Label: "Goshtoso", Ownership: OwnershipOrganization},
		{Key: "halloween", Label: "Halloween", Ownership: OwnershipGeneric},
		{Key: "high-contrast", Label: "High Contrast", Ownership: OwnershipGeneric},
		{Key: "industrial", Label: "Industrial", Ownership: OwnershipGeneric},
		{Key: "minimal", Label: "Minimal", Ownership: OwnershipGeneric},
		{Key: "modern", Label: "Modern", Ownership: OwnershipGeneric},
		{Key: "neo-brutalism", Label: "Neo Brutalism", Ownership: OwnershipGeneric},
		{Key: "news", Label: "News", Ownership: OwnershipGeneric},
		{Key: "pastel", Label: "Pastel", Ownership: OwnershipGeneric},
		{Key: "prototype", Label: "Prototype", Ownership: OwnershipGeneric},
		{Key: "zombie", Label: "Zombie", Ownership: OwnershipGeneric},
	}

	got := BuiltIn()
	seen := make(map[Key]struct{}, len(got))
	for _, theme := range got {
		if theme.Key == "" || theme.Label == "" || theme.Ownership == "" {
			t.Fatalf("BuiltIn() contains empty key, label, or ownership: %#v", theme)
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
	if second[0] != (Theme{Key: "90s", Label: "90s", Ownership: OwnershipGeneric}) {
		t.Fatalf("BuiltIn()[0] = %#v after caller mutation, want stable built-in", second[0])
	}
}

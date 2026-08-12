package themes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalogPreservesDemoPresentationOrder(t *testing.T) {
	require.Equal(t, []Theme{
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
	}, All())
}

func TestCatalogHasSixteenUniqueBuiltInThemes(t *testing.T) {
	all := All()
	require.Len(t, all, 16)
	require.Equal(t, len(all), Count())

	seen := make(map[string]bool, len(all))
	for _, theme := range all {
		require.NotEmpty(t, theme.Key)
		require.NotEmpty(t, theme.Label)
		require.Contains(t, []Ownership{OwnershipGeneric, OwnershipOrganization}, theme.Ownership)
		require.False(t, seen[theme.Key], "duplicate theme key %q", theme.Key)
		seen[theme.Key] = true
	}
	require.False(t, seen["totvs"])
}

func TestCatalogReturnsCallerOwnedSlice(t *testing.T) {
	first := All()
	first[0] = Theme{Key: "consumer-custom", Label: "Consumer Custom"}
	first = append(first, Theme{Key: "another-custom", Label: "Another Custom"})
	require.Len(t, first, 17)

	second := All()
	require.Len(t, second, 16)
	require.Equal(t, Theme{Key: "araihu", Label: "Arai Hû", Ownership: OwnershipOrganization}, second[0])
}

func TestPresentationLabelOverrideIsExplicitAndNarrow(t *testing.T) {
	label, ok := PresentationLabelOverride("zombie")
	require.True(t, ok)
	require.Equal(t, "Halloween II", label)

	label, ok = PresentationLabelOverride("halloween")
	require.False(t, ok)
	require.Empty(t, label)
}

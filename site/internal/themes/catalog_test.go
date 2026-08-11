package themes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalogPreservesDemoPresentationOrder(t *testing.T) {
	require.Equal(t, []Theme{
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
	require.Equal(t, Theme{Key: "araihu", Label: "Arai Hû"}, second[0])
}

func TestPresentationLabelOverrideIsExplicitAndNarrow(t *testing.T) {
	label, ok := PresentationLabelOverride("zombie")
	require.True(t, ok)
	require.Equal(t, "Halloween II", label)

	label, ok = PresentationLabelOverride("halloween")
	require.False(t, ok)
	require.Empty(t, label)
}

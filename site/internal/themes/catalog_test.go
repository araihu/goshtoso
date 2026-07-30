package themes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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

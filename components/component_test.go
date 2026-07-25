package components

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllKindsAreStableAndUnique(t *testing.T) {
	kinds := AllKinds()
	require.Len(t, kinds, 74)

	seen := map[Kind]struct{}{}
	for _, kind := range kinds {
		require.NotEmpty(t, kind)
		_, duplicate := seen[kind]
		require.Falsef(t, duplicate, "duplicate Kind %q", kind)
		seen[kind] = struct{}{}
	}
}

func TestAllKindsReturnsCopy(t *testing.T) {
	kinds := AllKinds()
	kinds[0] = "mutated"
	require.NotEqual(t, Kind("mutated"), AllKinds()[0])
}

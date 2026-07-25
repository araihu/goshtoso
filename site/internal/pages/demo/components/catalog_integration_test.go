package components

import (
	"strings"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/stretchr/testify/require"
)

func TestComponentRegistryAndMetadataFollowCatalog(t *testing.T) {
	pages := catalog.ComponentPages()
	require.Equal(t, len(pages), componentCount())

	componentRegistryKeys := 0
	for key := range Demos {
		if strings.HasPrefix(key, "components/") {
			componentRegistryKeys++
		}
	}
	require.Equal(t, len(pages), componentRegistryKeys)

	for _, page := range pages {
		entry, ok := LookupDemo(page.Key)
		require.Truef(t, ok, "missing demo registry entry for %s", page.Key)
		require.Equal(t, page.Active, entry.Active)
		require.NotNil(t, entry.Content)
		if entry.API != nil {
			require.NotEmpty(t, entry.API, "non-nil transitional API metadata must contain sections")
		}

		meta := MetaForKey(page.Key)
		require.Equal(t, page.Description, meta.Description)
		require.Equal(t, page.Path, meta.Path)
	}
}

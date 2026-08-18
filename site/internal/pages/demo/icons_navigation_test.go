package demo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIconsDocsNavigationContainsIconCatalogAndIconPacks(t *testing.T) {
	t.Parallel()

	navigation := iconsDocsNavigation("icon")

	require.Len(t, navigation.Items, 3)
	require.Equal(t, "component-icon", navigation.Items[0].ID)
	require.Equal(t, "Icon", navigation.Items[0].Label)
	require.Equal(t, "/components/icon", navigation.Items[0].Href)
	require.True(t, navigation.Items[0].Active)
	require.Equal(t, "icon-catalog", navigation.Items[1].ID)
	require.Equal(t, "Icon Catalog", navigation.Items[1].Label)
	require.Equal(t, "/docs/icon-catalog", navigation.Items[1].Href)
	require.Equal(t, "/docs/icon-catalog", navigation.Items[1].LinkAttrs["hx-get"])
	require.False(t, navigation.Items[1].Active)
	require.Equal(t, "iconpack", navigation.Items[2].ID)
	require.Equal(t, "Icon Packs", navigation.Items[2].Label)
	require.Equal(t, "/docs/iconpack", navigation.Items[2].Href)
	require.False(t, navigation.Items[2].Active)
	require.True(t, navigation.DisableSearch)
}

func TestComponentDocsFamilyMapsIconPages(t *testing.T) {
	t.Parallel()

	require.Equal(t, "icon-packs", componentDocsFamily("icon"))
	require.Equal(t, "icon-packs", componentDocsFamily("icon-catalog"))
	require.Equal(t, "icon-packs", componentDocsFamily("iconpack"))
}

func TestIconsDocsNavigationActivatesDedicatedCatalog(t *testing.T) {
	t.Parallel()

	navigation := iconsDocsNavigation("icon-catalog")

	require.False(t, navigation.Items[0].Active)
	require.True(t, navigation.Items[1].Active)
	require.False(t, navigation.Items[2].Active)
}

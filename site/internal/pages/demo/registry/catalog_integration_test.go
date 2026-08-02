package registry

import (
	"context"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/stretchr/testify/require"
)

func TestComponentDemoHeadingsMatchCanonicalCatalogTitles(t *testing.T) {
	tests := []struct {
		key   string
		title string
	}{
		{key: "components/avatar", title: "Avatar"},
		{key: "components/button", title: "Button"},
		{key: "components/toast", title: "Toast"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			entry, ok := Lookup(tt.key)
			require.True(t, ok)

			var rendered strings.Builder
			require.NoError(t, entry.Content().Render(context.Background(), &rendered))
			page := rendered.String()
			headingStart := strings.Index(page, "<h1 ")
			require.NotEqual(t, -1, headingStart)
			textStart := strings.Index(page[headingStart:], ">")
			require.NotEqual(t, -1, textStart)
			textStart += headingStart + 1
			textEnd := strings.Index(page[textStart:], "</h1>")
			require.NotEqual(t, -1, textEnd)

			require.Equal(t, tt.title, page[textStart:textStart+textEnd])
		})
	}
}

func TestComponentRegistryAndMetadataFollowCatalog(t *testing.T) {
	pages := catalog.ComponentPages()
	componentRegistryKeys := 0
	for _, meta := range AllPublicMeta() {
		if strings.HasPrefix(meta.Path, "/components/") {
			componentRegistryKeys++
		}
	}
	require.Equal(t, len(pages), componentRegistryKeys)

	for _, page := range pages {
		entry, ok := Lookup(page.Key)
		require.Truef(t, ok, "missing demo registry entry for %s", page.Key)
		require.Equal(t, page.Active, entry.Active)
		require.NotNil(t, entry.Content)
		require.NotEmpty(t, page.GoPackagePath())

		meta := MetaForKey(page.Key)
		require.Equal(t, page.Description, meta.Description)
		require.Equal(t, page.Path, meta.Path)
	}
}

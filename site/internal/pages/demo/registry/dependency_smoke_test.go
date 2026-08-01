package registry_test

import (
	"context"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	actiongrouppage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/actiongroup"
	buttonpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/button"
	"github.com/araihu/goshtoso/site/internal/pages/demo/registry"
	"github.com/stretchr/testify/require"
)

func TestPilotLeafDefinitionsRenderThroughRegistry(t *testing.T) {
	definitions := []struct {
		key        string
		definition func() string
	}{
		{
			key: "components/button",
			definition: func() string {
				return renderDefinition(t, buttonpage.Definition)
			},
		},
		{
			key: "components/action-group",
			definition: func() string {
				return renderDefinition(t, actiongrouppage.Definition)
			},
		},
	}

	for _, tt := range definitions {
		t.Run(tt.key, func(t *testing.T) {
			require.NotEmpty(t, tt.definition())
		})
	}

	componentCatalog := pilotCatalog(t, "components/button", "components/action-group")
	pages, err := registry.New(
		[]demo.PageDefinition{buttonpage.Definition, actiongrouppage.Definition},
		componentCatalog,
	)
	require.NoError(t, err)
	for _, tt := range definitions {
		_, ok := pages.Lookup(tt.key)
		require.Truef(t, ok, "missing pilot page %s", tt.key)
	}
}

func renderDefinition(t *testing.T, definition demo.PageDefinition) string {
	t.Helper()
	var rendered strings.Builder
	require.NoError(t, definition.Content().Render(context.Background(), &rendered))
	return rendered.String()
}

func pilotCatalog(t *testing.T, keys ...string) []catalog.Entry {
	t.Helper()
	entries := make([]catalog.Entry, 0, len(keys))
	for _, key := range keys {
		entry, ok := catalog.Lookup(key)
		require.True(t, ok)
		entries = append(entries, entry)
	}
	return entries
}

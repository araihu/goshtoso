package registry

import (
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	"github.com/stretchr/testify/require"
)

func TestNewRejectsInvalidDefinitions(t *testing.T) {
	component := catalog.ComponentPages()[0]
	valid := demo.PageDefinition{
		Key:     component.Key,
		Title:   component.Title,
		Active:  component.Active,
		Content: emptyComponent,
	}

	tests := []struct {
		name        string
		definitions []demo.PageDefinition
		catalog     []catalog.Entry
		wantError   string
	}{
		{
			name:        "empty key",
			definitions: []demo.PageDefinition{{Content: emptyComponent}},
			wantError:   "empty key",
		},
		{
			name:        "leading slash",
			definitions: []demo.PageDefinition{{Key: "/docs/theme", Content: emptyComponent}},
			wantError:   "canonical",
		},
		{
			name:        "unclean path",
			definitions: []demo.PageDefinition{{Key: "docs//theme", Content: emptyComponent}},
			wantError:   "canonical",
		},
		{
			name:        "missing content",
			definitions: []demo.PageDefinition{{Key: "docs/theme"}},
			wantError:   "nil content",
		},
		{
			name: "duplicate key",
			definitions: []demo.PageDefinition{
				{Key: "docs/theme", Content: emptyComponent},
				{Key: "docs/theme", Content: emptyComponent},
			},
			wantError: "duplicate key",
		},
		{
			name:        "component missing from catalog",
			definitions: []demo.PageDefinition{{Key: "components/missing", Content: emptyComponent}},
			wantError:   "not present in component catalog",
		},
		{
			name:        "catalog component missing from definitions",
			definitions: nil,
			catalog:     []catalog.Entry{component},
			wantError:   "missing component definition",
		},
		{
			name:        "valid component",
			definitions: []demo.PageDefinition{valid},
			catalog:     []catalog.Entry{component},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.definitions, tt.catalog)
			if tt.wantError == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestRegistryCopiesInputsAndReturnsDefensiveMetadata(t *testing.T) {
	definitions := []demo.PageDefinition{
		{
			Key:         "docs/theme",
			Title:       "Themes",
			Description: "Theme documentation",
			Type:        "TechArticle",
			Content:     emptyComponent,
		},
	}

	registry, err := New(definitions, nil)
	require.NoError(t, err)
	definitions[0].Title = "Mutated"

	definition, ok := registry.Lookup("docs/theme")
	require.True(t, ok)
	require.Equal(t, "Themes", definition.Title)

	metadata := registry.AllPublicMeta()
	require.Len(t, metadata, 2)
	metadata[1].Title = "Mutated"
	require.Equal(t, "Themes - Goshtoso UI Library for Go", registry.MetaForKey("docs/theme").Title)
}

func TestRegistryUsesCatalogDescriptionForComponents(t *testing.T) {
	component := catalog.ComponentPages()[0]
	registry, err := New([]demo.PageDefinition{
		{
			Key:         component.Key,
			Title:       "Wrong title",
			Active:      "wrong-active",
			Description: "Wrong description",
			Content:     emptyComponent,
		},
	}, []catalog.Entry{component})
	require.NoError(t, err)

	definition, ok := registry.Lookup(component.Key)
	require.True(t, ok)
	require.Equal(t, "Wrong title", definition.Title)
	require.Equal(t, "wrong-active", definition.Active)
	require.Equal(t, component.Description, definition.Description)

	meta := registry.MetaForKey(component.Key)
	require.Equal(t, component.Path, meta.Path)
	require.Equal(t, component.Description, meta.Description)
	require.Contains(t, meta.Title, "Wrong title Component")
}

func emptyComponent() templ.Component {
	return templ.NopComponent
}

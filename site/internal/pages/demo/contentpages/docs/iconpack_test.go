package docspages

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIconpackDocsRenderConsumerSurface(t *testing.T) {
	var output bytes.Buffer
	require.NoError(t, iconpackContent().Render(context.Background(), &output))
	html := output.String()

	for _, expected := range []string{
		"Icon Packs",
		"Generate from a verified Assets release",
		"Define sources with .iconpack.yaml",
		"A generated pack keeps the same icon contract",
		"Bootstrap Icons source",
		".iconpack.lock.yaml",
		"source-manifest",
		"brand-developer-icons-tRPC",
		"IconBrandDeveloperIconsTRPC",
		"Lookup",
		"components/icon",
		"/assets/icons/appicons/sprite.svg",
		"manifest.json",
		"provenance.json",
	} {
		require.Contains(t, html, expected)
	}
}

func TestIconCatalogDocsDefinitionIsRoutable(t *testing.T) {
	for _, definition := range Definitions {
		if definition.Key == "docs/icon-catalog" {
			require.Equal(t, "Icon Catalog", definition.Title)
			require.Equal(t, "icon-catalog", definition.Active)
			require.NotNil(t, definition.Content)
			return
		}
	}
	t.Fatalf("docs/icon-catalog definition is missing from Definitions: %s", strings.Join(definitionKeys(), ", "))
}

func TestIconpackDocsDefinitionIsRoutable(t *testing.T) {
	for _, definition := range Definitions {
		if definition.Key == "docs/iconpack" {
			require.Equal(t, "Icon Packs", definition.Title)
			require.Equal(t, "iconpack", definition.Active)
			require.NotNil(t, definition.Content)
			return
		}
	}
	t.Fatalf("docs/iconpack definition is missing from Definitions: %s", strings.Join(definitionKeys(), ", "))
}

func definitionKeys() []string {
	keys := make([]string, 0, len(Definitions))
	for _, definition := range Definitions {
		keys = append(keys, definition.Key)
	}
	return keys
}

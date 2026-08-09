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
		"Consumer-local icon packs",
		"Generate from an Assets release",
		"Bring any SVG pack",
		"Bootstrap Icons",
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

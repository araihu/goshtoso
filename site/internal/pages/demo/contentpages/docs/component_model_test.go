package docspages

import (
	"bytes"
	"context"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/demo"
	startpages "github.com/araihu/goshtoso/site/internal/pages/demo/contentpages/start"
	"github.com/stretchr/testify/require"
)

func TestComponentModelDocumentsPublicAPI(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, componentModelContent().Render(context.Background(), &buf))
	html := buf.String()

	for _, phrase := range []string{
		"components.Component",
		"templ.Component",
		"Kind",
		"Keep the concrete type when needed",
		"Configuration structs",
		"Functional options",
		"Defaults and API reference",
		"countTables",
		"SaveButton",
		"KindTable",
		`href="/components/button"`,
		`href="/docs/theme"`,
		`<pre class="ch-chroma"`,
	} {
		require.Contains(t, html, phrase)
	}

	for _, internalPhrase := range []string{
		"Consumer vocabulary",
		"Configuration dimension",
		"documentation vocabulary",
		"Axis",
		"There is no universal Variant",
		"One primitive or two",
	} {
		require.NotContains(t, html, internalPhrase)
	}
}

func TestComponentModelRouteIsRegistered(t *testing.T) {
	entry := docsDefinition(t, "docs/component-model")
	require.Equal(t, "Component Model", entry.Title)
	require.Equal(t, "component-model", entry.Active)
	require.NotNil(t, entry.Content)
}

func TestComponentModelIsLinkedFromSiteNavigation(t *testing.T) {
	var layout bytes.Buffer
	require.NoError(t, demo.Layout("Component Model", "component-model", componentModelContent()).Render(context.Background(), &layout))
	require.Contains(t, layout.String(), `href="/docs/component-model"`)
	require.Contains(t, layout.String(), "search-component-model")

	var buf bytes.Buffer
	require.NoError(t, startpages.Definitions[0].Content().Render(context.Background(), &buf))
	require.Contains(t, buf.String(), `href="/docs/component-model"`)
}

func docsDefinition(t *testing.T, key string) demo.PageDefinition {
	t.Helper()
	for _, definition := range Definitions {
		if definition.Key == key {
			return definition
		}
	}
	t.Fatalf("missing docs definition %q", key)
	return demo.PageDefinition{}
}

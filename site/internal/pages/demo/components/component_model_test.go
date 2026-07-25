package components

import (
	"bytes"
	"context"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/demo"
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
		"Concrete return values",
		"Configuration structs",
		"Functional options",
		"Rendered defaults",
		"component.Kind()",
		"KindAlertDialog",
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
	entry, ok := LookupDemo("docs/component-model")
	require.True(t, ok)
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
	require.NoError(t, gettingStartedContent().Render(context.Background(), &buf))
	require.Contains(t, buf.String(), `href="/docs/component-model"`)
}

package components

import (
	"bytes"
	"context"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/demo"
	"github.com/stretchr/testify/require"
)

func TestComponentModelDocumentsConsumerVocabulary(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, componentModelContent().Render(context.Background(), &buf))
	html := buf.String()

	for _, phrase := range []string{
		"Theme",
		"Primitive",
		"Kind",
		"Configuration dimension",
		"There is no universal Variant",
		"One primitive or two",
		"effective defaults",
		"zero values",
		"functional options",
		"component.Kind()",
		"case components.KindAlertDialog:",
		"case components.KindTable:",
	} {
		require.Contains(t, html, phrase)
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

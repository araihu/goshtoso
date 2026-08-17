package demo

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSidebarTopItemsUseTheConsumerLocalHeroiconsSprite(t *testing.T) {
	wantSymbols := map[string]string{
		"home":                 "heroicons-optimized-24-outline-arrow-down-tray",
		"component-model":      "heroicons-optimized-24-outline-cube",
		"application-patterns": "heroicons-optimized-24-outline-queue-list",
		"iconpack":             "heroicons-optimized-24-outline-squares-2x2",
		"agents":               "heroicons-optimized-24-outline-sparkles",
		"theme":                "heroicons-optimized-24-outline-swatch",
		"attributions":         "heroicons-optimized-24-outline-heart",
		"license":              "heroicons-optimized-24-outline-clipboard-document-list",
	}

	for _, item := range getSidebarTopItems("") {
		want, ok := wantSymbols[item.ID]
		require.Truef(t, ok, "unexpected sidebar item %q", item.ID)

		var rendered bytes.Buffer
		require.NotNil(t, item.Icon, "sidebar item %q has no icon", item.ID)
		require.NoError(t, item.Icon.Render(context.Background(), &rendered))
		require.Contains(t, rendered.String(), `href="/assets/icons/heroicons/sprite.svg#`+want+`"`, item.ID)
		require.Contains(t, rendered.String(), `aria-hidden="true"`, item.ID)
		require.NotContains(t, rendered.String(), "<path", item.ID)
	}
}

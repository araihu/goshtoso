package components

import (
	"context"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/icon"
	"github.com/araihu/goshtoso/components/icon/heroicons"
	"github.com/stretchr/testify/require"
)

func renderIconShowcase(t *testing.T) string {
	t.Helper()
	var rendered strings.Builder
	require.NoError(t, iconDemoContent().Render(context.Background(), &rendered))
	return rendered.String()
}

func TestIconShowcaseRendersEveryGlyphInResponsiveGrid(t *testing.T) {
	html := renderIconShowcase(t)

	require.Len(t, heroicons.Glyphs, 67)
	require.Equal(t, len(heroicons.Glyphs), strings.Count(html, `data-icon-card=`))
	require.Contains(t, html, `grid-cols-1 sm:grid-cols-3 xl:grid-cols-6`)
}

func TestIconExampleReflectsMeaningfulSelectedOptions(t *testing.T) {
	got := iconExample(heroicons.Glyphs[0], icon.Config{
		Size:      icon.SizeLG,
		Label:     "Search",
		RootClass: "text-accent",
	})

	require.Contains(t, got, "heroicons.SpriteURL")
	require.Contains(t, got, "heroicons."+heroicons.Glyphs[0].GoName)
	require.Contains(t, got, "Size:      icon.SizeLG")
	require.Contains(t, got, `Label:     "Search"`)
	require.Contains(t, got, `RootClass: "text-accent"`)
	require.NotContains(t, got, "Decorative:")
}

func TestIconShowcaseUsesModalSemanticsAndAccessibleControls(t *testing.T) {
	html := renderIconShowcase(t)

	require.Contains(t, html, `role="dialog"`)
	require.Contains(t, html, `aria-modal="true"`)
	require.Contains(t, html, "x-trap.inert.noscroll")
	require.Contains(t, html, "keydown.esc.window")
	for _, control := range []string{"Size", "Label", "Decorative", "Root class", "Copy Go code"} {
		require.Contains(t, html, control)
	}
	require.Contains(t, html, `aria-live="polite"`)
}

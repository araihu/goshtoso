package skeleton

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
	"github.com/stretchr/testify/require"
)

func renderHTML(t *testing.T, component templ.Component) string {
	t.Helper()

	var output bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &output))
	return output.String()
}

func TestSkeletonZeroValueRendersAccessibleTextLoadingState(t *testing.T) {
	html := renderHTML(t, Skeleton(Config{}))

	require.Contains(t, html, `role="status"`)
	require.Contains(t, html, `aria-busy="true"`)
	require.Contains(t, html, "Loading content")
	require.Equal(t, 3, strings.Count(html, `data-skeleton-item`))
	require.Contains(t, html, "animate-pulse")
	require.Contains(t, html, "motion-reduce:animate-none")
	require.Contains(t, html, `data-shape="text"`)
	require.Equal(t, components.KindSkeleton, Skeleton(Config{}).Kind())
}

func TestSkeletonSupportsShapeCountStaticModeAndTargetHooks(t *testing.T) {
	html := renderHTML(t, Skeleton(Config{
		Shape:     ShapeCircle,
		Count:     2,
		Label:     "Loading members",
		Static:    true,
		RootClass: "skeleton-hook",
		RootAttrs: templ.Attributes{"data-loading": "members"},
		ItemClass: "item-hook",
		ItemAttrs: templ.Attributes{"data-region": "placeholder"},
	}))

	require.Contains(t, html, "Loading members")
	require.Equal(t, 2, strings.Count(html, `data-skeleton-item`))
	require.Contains(t, html, `data-shape="circle"`)
	require.Contains(t, html, "rounded-full")
	require.NotContains(t, html, "animate-pulse")
	require.Contains(t, html, "skeleton-hook")
	require.Contains(t, html, `data-loading="members"`)
	require.Contains(t, html, "item-hook")
	require.Contains(t, html, `data-region="placeholder"`)

	rectangle := renderHTML(t, Skeleton(Config{Shape: ShapeRectangle}))
	require.Contains(t, rectangle, `data-shape="rectangle"`)
	require.Equal(t, 1, strings.Count(rectangle, `data-skeleton-item`))
}

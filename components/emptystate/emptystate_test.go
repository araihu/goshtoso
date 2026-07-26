package emptystate

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

func TestEmptyStateZeroValueTeachesWhatHappensNext(t *testing.T) {
	html := renderHTML(t, EmptyState(Config{}))

	require.Contains(t, html, "<section")
	require.Contains(t, html, "Nothing here yet")
	require.Contains(t, html, "Items will appear here when they are available.")
	require.Contains(t, html, "text-center")
	require.Equal(t, components.KindEmptyState, EmptyState(Config{}).Kind())
}

func TestEmptyStateRendersIconAndActionWithTargetHooks(t *testing.T) {
	html := renderHTML(t, EmptyState(Config{
		Title:       "No incidents",
		Description: "Adjust filters or create an incident.",
		Icon:        templ.Raw(`<svg data-slot="icon"></svg>`),
		Action:      templ.Raw(`<button data-slot="action">Create incident</button>`),
		RootClass:   "empty-hook",
		RootAttrs:   templ.Attributes{"data-state": "empty"},
		IconClass:   "icon-hook",
		IconAttrs:   templ.Attributes{"data-region": "icon"},
		ActionClass: "action-hook",
		ActionAttrs: templ.Attributes{"data-region": "action"},
	}))

	require.Contains(t, html, "empty-hook")
	require.Contains(t, html, `data-state="empty"`)
	require.Contains(t, html, "icon-hook")
	require.Contains(t, html, `data-region="icon"`)
	require.Contains(t, html, "action-hook")
	require.Contains(t, html, `data-region="action"`)

	iconIndex := strings.Index(html, `data-slot="icon"`)
	titleIndex := strings.Index(html, "No incidents")
	actionIndex := strings.Index(html, `data-slot="action"`)
	require.Greater(t, titleIndex, iconIndex)
	require.Greater(t, actionIndex, titleIndex)
}

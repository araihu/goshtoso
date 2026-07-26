package toolbar

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

func TestToolbarRendersAccessibleResponsiveGroupByDefault(t *testing.T) {
	html := renderHTML(t, Toolbar(Config{}))

	require.Contains(t, html, `role="toolbar"`)
	require.Contains(t, html, `aria-label="Page tools"`)
	require.Contains(t, html, "flex")
	require.Contains(t, html, "flex-wrap")
	require.Equal(t, components.KindToolbar, Toolbar(Config{}).Kind())
}

func TestToolbarRendersSearchFiltersAndActionsWithTargetHooks(t *testing.T) {
	html := renderHTML(t, Toolbar(Config{
		Label:        "Incident tools",
		Search:       templ.Raw(`<input data-slot="search"/>`),
		Filters:      templ.Raw(`<button data-slot="filters">Severity</button>`),
		Actions:      templ.Raw(`<button data-slot="actions">Create incident</button>`),
		Sticky:       true,
		RootClass:    "toolbar-hook",
		RootAttrs:    templ.Attributes{"data-toolbar": "incidents"},
		SearchClass:  "search-hook",
		SearchAttrs:  templ.Attributes{"data-region": "search"},
		FiltersClass: "filters-hook",
		FiltersAttrs: templ.Attributes{"data-region": "filters"},
		ActionsClass: "actions-hook",
		ActionsAttrs: templ.Attributes{"data-region": "actions"},
	}))

	require.Contains(t, html, `aria-label="Incident tools"`)
	require.Contains(t, html, "sticky")
	require.Contains(t, html, "toolbar-hook")
	require.Contains(t, html, `data-toolbar="incidents"`)
	require.Contains(t, html, "search-hook")
	require.Contains(t, html, `data-region="search"`)
	require.Contains(t, html, "filters-hook")
	require.Contains(t, html, `data-region="filters"`)
	require.Contains(t, html, "actions-hook")
	require.Contains(t, html, `data-region="actions"`)

	searchIndex := strings.Index(html, `data-slot="search"`)
	filtersIndex := strings.Index(html, `data-slot="filters"`)
	actionsIndex := strings.Index(html, `data-slot="actions"`)
	require.Greater(t, filtersIndex, searchIndex)
	require.Greater(t, actionsIndex, filtersIndex)
}

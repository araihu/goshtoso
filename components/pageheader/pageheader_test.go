package pageheader

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

func TestPageHeaderRendersPageHierarchy(t *testing.T) {
	html := renderHTML(t, PageHeader(Config{
		Title:       "Operations",
		Description: "Review and act on current incidents.",
	}))

	require.Contains(t, html, "<header")
	require.Contains(t, html, "<h1")
	require.Contains(t, html, ">Operations</h1>")
	require.Contains(t, html, "Review and act on current incidents.")
	require.Contains(t, html, "max-w-3xl")
	require.Equal(t, components.KindPageHeader, PageHeader(Config{}).Kind())
}

func TestPageHeaderRendersBreadcrumbsBeforeTitleAndActionsAfterDescription(t *testing.T) {
	html := renderHTML(t, PageHeader(Config{
		Title:            "Cluster details",
		Description:      "Production cluster status.",
		Breadcrumbs:      templ.Raw(`<nav data-slot="breadcrumbs">Breadcrumbs</nav>`),
		Actions:          templ.Raw(`<button data-slot="actions">Restart</button>`),
		RootClass:        "page-header-hook",
		RootAttrs:        templ.Attributes{"data-page-header": "cluster"},
		BreadcrumbsClass: "breadcrumbs-hook",
		BreadcrumbsAttrs: templ.Attributes{"data-region": "breadcrumbs"},
		ActionsClass:     "actions-hook",
		ActionsAttrs:     templ.Attributes{"data-region": "actions"},
	}))

	require.Contains(t, html, "page-header-hook")
	require.Contains(t, html, `data-page-header="cluster"`)
	require.Contains(t, html, "breadcrumbs-hook")
	require.Contains(t, html, `data-region="breadcrumbs"`)
	require.Contains(t, html, "actions-hook")
	require.Contains(t, html, `data-region="actions"`)

	breadcrumbsIndex := strings.Index(html, `data-slot="breadcrumbs"`)
	titleIndex := strings.Index(html, ">Cluster details</h1>")
	descriptionIndex := strings.Index(html, "Production cluster status.")
	actionsIndex := strings.Index(html, `data-slot="actions"`)
	require.Greater(t, titleIndex, breadcrumbsIndex)
	require.Greater(t, descriptionIndex, titleIndex)
	require.Greater(t, actionsIndex, descriptionIndex)
}

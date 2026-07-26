package components

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/demo"
	"github.com/stretchr/testify/require"
)

func TestApplicationPatternsRenderCompleteRecipeContracts(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, applicationPatternsContent().Render(context.Background(), &buf))
	html := buf.String()

	for _, pattern := range []string{
		"app-shell",
		"operations-list",
		"detail-workspace",
		"multi-step-workflow",
	} {
		require.Contains(t, html, `data-application-pattern="`+pattern+`"`)
		require.Contains(t, html, `id="`+pattern+`"`)
	}

	for _, marker := range []string{
		"data-pattern-preview",
		"data-pattern-problem",
		"data-pattern-components",
		"data-pattern-states",
		"data-pattern-390",
		"data-pattern-1440",
		"data-pattern-accessibility",
		"data-pattern-app-specific",
		"data-pattern-source-map",
		"data-pattern-done",
	} {
		require.Equal(t, 4, strings.Count(html, marker), marker)
	}

	for _, evidence := range []string{
		`aria-label="Application recipe index"`,
		`aria-label="App shell preview content"`,
		`<table`,
		`<caption class="sr-only">Deployments awaiting operational review</caption>`,
		`role="tablist"`,
		`aria-label="Deployment workflow progress"`,
		"app/ui/shell.templ",
		"app/deployments/query.go",
		"app/releases/detail.go",
		"app/workflow/state.go",
	} {
		require.Contains(t, html, evidence)
	}
}

func TestApplicationPatternsRouteIsRegistered(t *testing.T) {
	entry, ok := LookupDemo("docs/application-patterns")
	require.True(t, ok)
	require.Equal(t, "Application Patterns", entry.Title)
	require.Equal(t, "application-patterns", entry.Active)
	require.NotNil(t, entry.Content)

	meta := DemoMeta("docs/application-patterns", entry)
	require.Equal(t, "Application Patterns for Goshtoso", meta.Title)
	require.Equal(t, "/docs/application-patterns", meta.Path)
	require.Contains(t, meta.Description, "App Shell")
	require.Contains(t, meta.Description, "Multi-step Workflow")
}

func TestApplicationPatternsIsLinkedFromDocsNavigationAndSearch(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, demo.Layout("Application Patterns", "application-patterns", applicationPatternsContent()).Render(context.Background(), &buf))
	html := buf.String()

	require.Contains(t, html, `href="/docs/application-patterns"`)
	require.Contains(t, html, `aria-current="page"`)
	require.Contains(t, html, `id="search-application-patterns"`)
	require.Contains(t, html, "Compose App Shell, Operations List, Detail Workspace, and Multi-step Workflow")
}

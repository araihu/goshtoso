package docspages

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

	require.Equal(t, 6, strings.Count(html, `data-toc-heading class="scroll-mt-20`))
	require.Equal(t, 4, strings.Count(html, `data-pattern-contract class="mt-8 border-t border-outline`))
	require.NotContains(t, html, `class="mt-8 border-y border-outline`)
	require.Contains(t, html, `for="operations-list-search"`)
	require.Contains(t, html, `>Search deployments</label>`)
	require.Contains(t, html, `!bg-transparent`)
	require.Contains(t, html, `inline-flex h-8 items-center`)
}

func TestApplicationPatternsRouteIsRegistered(t *testing.T) {
	entry := docsDefinition(t, "docs/application-patterns")
	require.Equal(t, "Application Patterns", entry.Title)
	require.Equal(t, "application-patterns", entry.Active)
	require.NotNil(t, entry.Content)
	require.Contains(t, entry.Description, "App Shell")
	require.Contains(t, entry.Description, "Multi-step Workflow")
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

func TestApplicationPatternsIsLinkedFromDocsNavigationAndSearch(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, demo.Layout("Application Patterns", "application-patterns", applicationPatternsContent()).Render(context.Background(), &buf))
	html := buf.String()

	require.Contains(t, html, `href="/docs/application-patterns"`)
	require.Contains(t, html, `aria-current="page"`)
	require.Contains(t, html, `id="search-application-patterns"`)
	require.Contains(t, html, "Compose App Shell, Operations List, Detail Workspace, and Multi-step Workflow")
}

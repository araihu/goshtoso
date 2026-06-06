package breadcrumbs

import (
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
)

func TestBreadcrumbsCoverageRendersDefaultStructureAndCurrentPage(t *testing.T) {
	rendered := renderBreadcrumbs(t, Config{
		Items: []Item{
			{Label: "Home", Href: "/"},
			{Label: "Docs", Href: "/docs"},
		},
		Current: "Breadcrumbs",
	})

	assert.Contains(t, rendered, `aria-label="breadcrumb"`)
	assert.Contains(t, rendered, `class="flex flex-wrap items-center gap-1"`)
	assert.Contains(t, rendered, `class="flex items-center gap-1"`)
	assert.Contains(t, rendered, `href="/"`)
	assert.Contains(t, rendered, `href="/docs"`)
	assert.Contains(t, rendered, `aria-current="page"`)
	assert.Contains(t, rendered, `Breadcrumbs`)
	assert.Equal(t, 2, strings.Count(rendered, `<svg xmlns="http://www.w3.org/2000/svg" fill="none"`))
	assert.NotContains(t, rendered, `<span aria-hidden="true">/</span>`)
}

func TestBreadcrumbsCoverageRendersSlashSeparatorClassesAndAttributes(t *testing.T) {
	rendered := renderBreadcrumbs(t, Config{
		Items: []Item{
			{
				Label: "Components",
				Href:  "/components?section=nav",
				LinkAttrs: templ.Attributes{
					"data-track": "breadcrumb-link",
					"rel":        "nofollow",
				},
			},
			{
				Label:   "Breadcrumbs",
				Href:    "/components/breadcrumbs",
				Icon:    templ.Raw(`<svg data-testid="crumb-icon"></svg>`),
				Tooltip: "Breadcrumb component",
			},
		},
		Separator: Slash,
		NavClass:  "custom-nav",
		NavAttrs: templ.Attributes{
			"id":        "docs-breadcrumbs",
			"data-role": "trail",
		},
	})

	assert.Contains(t, rendered, `text-on-surface dark:text-on-surface-dark custom-nav`)
	assert.Contains(t, rendered, `id="docs-breadcrumbs"`)
	assert.Contains(t, rendered, `data-role="trail"`)
	assert.Contains(t, rendered, `class="flex flex-wrap items-center gap-2"`)
	assert.Contains(t, rendered, `class="flex items-center gap-2"`)
	assert.Contains(t, rendered, `href="/components?section=nav"`)
	assert.Contains(t, rendered, `data-track="breadcrumb-link"`)
	assert.Contains(t, rendered, `rel="nofollow"`)
	assert.Contains(t, rendered, `title="Breadcrumb component"`)
	assert.Contains(t, rendered, `data-testid="crumb-icon"`)
	assert.Equal(t, 2, strings.Count(rendered, `<span aria-hidden="true">/</span>`))
	assert.NotContains(t, rendered, `aria-current="page"`)
	assert.NotContains(t, rendered, `fill="none" viewBox="0 0 24 24"`)
}

func TestBreadcrumbsCoverageHelperClassBranches(t *testing.T) {
	assert.Equal(t, "", extraClass(""))
	assert.Equal(t, " mt-2", extraClass("mt-2"))
	assert.Equal(t, "flex flex-wrap items-center gap-1", listClasses(Chevron))
	assert.Equal(t, "flex flex-wrap items-center gap-2", listClasses(Slash))
	assert.Equal(t, "flex items-center gap-1", itemClasses(Chevron))
	assert.Equal(t, "flex items-center gap-2", itemClasses(Slash))
}

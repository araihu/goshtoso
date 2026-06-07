package sidebar

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestCoverageRenderSidebarBranches(t *testing.T) {
	html := renderSidebar(t, Config{
		LogoText:          "Docs",
		LogoHref:          "/docs",
		ShowSearch:        true,
		SearchPlaceholder: "Filter navigation",
		RootClass:         "coverage-root",
		Items: []Item{
			{
				ID:     "overview",
				Label:  "Overview",
				Href:   "/overview",
				Icon:   testComponent(`<svg data-testid="overview-icon"></svg>`),
				Active: true,
				LinkAttrs: templ.Attributes{
					"data-sidebar-item": "Overview",
					"hx-get":            "/overview",
					"hx-target":         "#main-content",
				},
			},
			{
				ID:    "inbox",
				Label: "Inbox",
				Href:  "/inbox",
				Icon:  testComponent(`<svg data-testid="inbox-icon"></svg>`),
				Badge: "7",
				Items: []Item{
					{ID: "archive", Label: "Archive", Href: "/archive", Badge: "new"},
					{ID: "muted", Label: "Muted", Href: "/muted", Disabled: true},
				},
			},
			{
				ID:       "disabled",
				Label:    "Disabled top item",
				Href:     "/disabled",
				Icon:     testComponent(`<svg data-testid="disabled-icon"></svg>`),
				Disabled: true,
			},
		},
		SectionsTitle: "Component groups",
		Sections: []Section{
			{
				Title: "Primary",
				Items: []Item{
					{ID: "button", Label: "Button", Href: "/components/button", Active: true, Badge: "hot"},
					{ID: "card", Label: "Card", Href: "/components/card", Badge: "soon"},
					{
						ID:    "nested",
						Label: "Nested",
						Href:  "/components/nested",
						Items: []Item{
							{ID: "nested-child", Label: "Nested child", Href: "/components/nested-child"},
						},
					},
					{ID: "disabled-section", Label: "Disabled section item", Href: "/components/disabled", Disabled: true},
				},
			},
		},
		FooterSlot: testComponent(`<div data-testid="sidebar-footer">Footer slot</div>`),
	})

	assertContainsAll(t, html,
		`aria-label="sidebar navigation"`,
		`href="/docs"`,
		`Docs`,
		`coverage-root`,
		`placeholder="Filter navigation"`,
		`href="/overview"`,
		`data-sidebar-item="Overview"`,
		`hx-get="/overview"`,
		`hx-target="#main-content"`,
		`text-primary dark:text-primary-dark`,
		`<span class="sr-only">active</span>`,
		`data-testid="overview-icon"`,
		`Inbox`,
		`>7</span>`,
		`Archive`,
		`>new</span>`,
		`Disabled top item`,
		`cursor-not-allowed`,
		`Component groups`,
		`data-sidebar-section="Primary"`,
		`border-l-2 border-primary`,
		`>hot</sup>`,
		`>soon</sup>`,
		`Nested child`,
		`Disabled section item`,
		`data-testid="sidebar-footer"`,
	)

	if strings.Contains(html, `href="/disabled"`) {
		t.Fatalf("disabled top item should not render an href: %s", html)
	}
	if strings.Contains(html, `href="/components/disabled"`) {
		t.Fatalf("disabled section item should not render an href: %s", html)
	}
}

func TestCoverageSearchSlotOverridesDefaultSearch(t *testing.T) {
	html := renderSidebar(t, Config{
		Logo:              testComponent(`<span data-testid="logo-slot">Logo slot</span>`),
		LogoText:          "Fallback text",
		ShowSearch:        true,
		SearchPlaceholder: "Should not render",
		SearchSlot:        testComponent(`<div data-testid="search-slot">Custom search</div>`),
	})

	assertContainsAll(t, html,
		`data-testid="logo-slot"`,
		`data-testid="search-slot"`,
		`Custom search`,
	)

	for _, absent := range []string{`Fallback text`, `name="search"`, `Should not render`} {
		if strings.Contains(html, absent) {
			t.Fatalf("custom slot render should not contain %q: %s", absent, html)
		}
	}
}

func TestSectionItemBadgesCanBeColoredAndRightAligned(t *testing.T) {
	html := renderSidebar(t, Config{
		Sections: []Section{
			{
				Title: "Operations",
				Items: []Item{
					{
						ID:         "get-pets",
						Label:      "/pets",
						Href:       "#operation-get-pets",
						Badge:      "GET",
						BadgeClass: "bg-emerald-600 text-white dark:bg-emerald-400 dark:text-neutral-950",
					},
				},
			},
		},
	})

	assertContainsAll(t, html,
		`<span class="min-w-0 flex-1 truncate">/pets</span>`,
		`ml-auto shrink-0`,
		`bg-emerald-600 text-white dark:bg-emerald-400 dark:text-neutral-950`,
		`>GET</sup>`,
	)
}

func TestSectionItemsWithPersistentChildrenRenderParentWithoutRail(t *testing.T) {
	html := renderSidebar(t, Config{
		Sections: []Section{
			{
				Title: "Operations",
				Items: []Item{
					{
						ID:    "tag-pets",
						Label: "Pets",
						Href:  "#operations-heading",
						Items: []Item{
							{ID: "get-pets", Label: "/pets", Href: "#operation-get-pets", Badge: "GET"},
						},
					},
				},
			},
		},
	})

	assertContainsAll(t, html,
		`<a href="#operations-heading" class="flex items-center gap-2 py-2.5 pl-4 text-sm font-medium text-on-surface transition duration-200 hover:text-on-surface-strong dark:text-on-surface-dark dark:hover:text-on-surface-dark-strong">`,
		`<div class="ml-4 flex flex-col">`,
		`<a href="#operation-get-pets" class="flex items-center gap-2 border-l border-outline py-2.5 pl-4 text-sm font-medium text-on-surface transition duration-200 hover:border-l-2 hover:border-outline-strong hover:text-on-surface-strong dark:border-outline-dark dark:text-on-surface-dark dark:hover:border-outline-dark-strong dark:hover:text-on-surface-dark-strong">`,
	)
	if strings.Contains(html, `<a href="#operations-heading" class="flex items-center gap-2 border-l`) {
		t.Fatalf("parent item with persistent children should not render a leading rail: %s", html)
	}
}

func TestSectionItemsWithChildrenCanCollapse(t *testing.T) {
	html := renderSidebar(t, Config{
		Sections: []Section{
			{
				Title: "Operations",
				Items: []Item{
					{
						ID:          "tag-pets",
						Label:       "Pets",
						Href:        "#operations-heading",
						Collapsible: true,
						Open:        true,
						Items: []Item{
							{ID: "get-pets", Label: "/pets", Href: "#operation-get-pets", Badge: "GET"},
						},
					},
				},
			},
		},
	})

	assertContainsAll(t, html,
		`x-data="{ open: true }"`,
		`x-on:click.prevent="open = !open"`,
		`x-bind:aria-expanded="open.toString()"`,
		`x-show="open"`,
		`id="tag-pets-children"`,
		`<span class="min-w-0 flex-1 truncate">Pets</span>`,
	)
	if count := strings.Count(html, `href="#operation-get-pets"`); count != 1 {
		t.Fatalf("collapsible children rendered %d links, want 1: %s", count, html)
	}
}

func TestCoverageSidebarClassHelpers(t *testing.T) {
	cfg := Config{}

	assertContainsAll(t, cfg.ContainerClasses(),
		`h-full`,
		`border-r`,
		`bg-surface`,
		`dark:bg-surface-dark`,
		`flex flex-col`,
	)
	assertContainsAll(t, cfg.NavClasses(),
		`flex-1`,
		`overflow-y-auto`,
		`sidebar-scroll`,
		`p-4`,
	)
}

func renderSidebar(t *testing.T, cfg Config) string {
	t.Helper()

	var buf bytes.Buffer
	if err := Sidebar(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render sidebar: %v", err)
	}
	return buf.String()
}

func testComponent(markup string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, markup)
		return err
	})
}

func assertContainsAll(t *testing.T, got string, wants ...string) {
	t.Helper()

	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %s", want, got)
		}
	}
}

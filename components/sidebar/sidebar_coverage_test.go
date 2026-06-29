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

func TestSidebarLinksExposeFullLabelTitleByDefault(t *testing.T) {
	html := renderSidebar(t, Config{
		Sections: []Section{{
			Title: "Operations",
			Items: []Item{{
				ID:    "long-operation",
				Label: "Get the download status for a pre-receive environment",
				Href:  "#operation-long",
				Badge: "GET",
			}},
		}},
	})

	assertContainsAll(t, html,
		`title="Get the download status for a pre-receive environment"`,
		`<span class="min-w-0 flex-1 truncate">Get the download status for a pre-receive environment</span>`,
		`>GET</sup>`,
	)
}

func TestSidebarLinkTitleCanBeOverriddenOrDisabled(t *testing.T) {
	html := renderSidebar(t, Config{
		Sections: []Section{{
			Title: "Operations",
			Items: []Item{
				{
					ID:    "custom-title",
					Label: "Short label",
					Title: "Longer browser tooltip",
					Href:  "#custom-title",
				},
				{
					ID:               "no-title",
					Label:            "No title label",
					Href:             "#no-title",
					DisableAutoTitle: true,
				},
				{
					ID:    "linkattrs-title",
					Label: "Generated title should not win",
					Href:  "#linkattrs-title",
					LinkAttrs: templ.Attributes{
						"title": "Consumer title",
					},
				},
			},
		}},
	})

	assertContainsAll(t, html,
		`title="Longer browser tooltip"`,
		`title="Consumer title"`,
	)

	if strings.Contains(html, `title="No title label"`) {
		t.Fatalf("DisableAutoTitle should suppress generated title: %s", html)
	}
	if strings.Contains(html, `title="Generated title should not win"`) {
		t.Fatalf("LinkAttrs title should win over generated title: %s", html)
	}
}

func TestSidebarLinkAriaLabelIsExplicitOnly(t *testing.T) {
	html := renderSidebar(t, Config{
		Sections: []Section{{
			Title: "Operations",
			Items: []Item{
				{
					ID:        "with-aria-label",
					Label:     "List global webhooks",
					AriaLabel: "List global webhooks endpoint",
					Href:      "#with-aria-label",
				},
				{
					ID:    "without-aria-label",
					Label: "Create a global webhook",
					Href:  "#without-aria-label",
				},
			},
		}},
	})

	assertContainsAll(t, html,
		`aria-label="List global webhooks endpoint"`,
		`title="List global webhooks"`,
		`title="Create a global webhook"`,
	)

	if strings.Contains(html, `aria-label="Create a global webhook"`) {
		t.Fatalf("aria-label should not be generated by default: %s", html)
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
						BadgeClass: "border-info bg-info text-on-info",
					},
				},
			},
		},
	})

	assertContainsAll(t, html,
		`title="/pets"`,
		`<span class="min-w-0 flex-1 truncate">/pets</span>`,
		`ml-auto shrink-0`,
		`border-info bg-info text-on-info`,
		`>GET</sup>`,
	)
}

func TestSectionItemBadgesNeutralizeSuperscriptOffset(t *testing.T) {
	html := renderSidebar(t, Config{
		Sections: []Section{
			{
				Title: "Operations",
				Items: []Item{
					{
						ID:    "get-pets",
						Label: "List pets",
						Href:  "#operation-get-pets",
						Badge: "GET",
					},
				},
			},
		},
	})

	assertContainsAll(t, html,
		`<span class="min-w-0 flex-1 truncate">List pets</span>`,
		`ml-auto shrink-0 static inline-flex items-center justify-center`,
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
		`<a href="#operations-heading" title="Pets" class="flex items-center gap-2 py-2.5 pl-4 text-sm font-medium text-on-surface transition duration-200 hover:text-on-surface-strong dark:text-on-surface-dark dark:hover:text-on-surface-dark-strong">`,
		`<div class="ml-4 flex flex-col">`,
		`<a href="#operation-get-pets" title="/pets" class="flex items-center gap-2 border-l border-outline py-2.5 pl-4 text-sm font-medium text-on-surface transition duration-200 hover:border-l-2 hover:border-outline-strong hover:text-on-surface-strong dark:border-outline-dark dark:text-on-surface-dark dark:hover:border-outline-dark-strong dark:hover:text-on-surface-dark-strong">`,
	)
	if strings.Contains(html, `<a href="#operations-heading" title="Pets" class="flex items-center gap-2 border-l`) {
		t.Fatalf("parent item with persistent children should not render a leading rail: %s", html)
	}
}

func TestSectionItemsCanBeIndentedFromSectionTitle(t *testing.T) {
	html := renderSidebar(t, Config{
		Sections: []Section{
			{
				Title:       "Schemas",
				IndentItems: true,
				Items: []Item{
					{ID: "schema-pet", Label: "Pet", Href: "#schema-pet"},
				},
			},
		},
	})

	assertContainsAll(t, html,
		`<div data-sidebar-section="Schemas">`,
		`<div class="ml-4 flex flex-col">`,
		`<a href="#schema-pet" title="Pet" class="flex items-center gap-2 border-l border-outline py-2.5 pl-4 text-sm font-medium text-on-surface transition duration-200 hover:border-l-2 hover:border-outline-strong hover:text-on-surface-strong dark:border-outline-dark dark:text-on-surface-dark dark:hover:border-outline-dark-strong dark:hover:text-on-surface-dark-strong">`,
	)
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
		`aria-controls="tag-pets-children"`,
		`x-show="open"`,
		`id="tag-pets-children"`,
		`<span class="min-w-0 flex-1 truncate">Pets</span>`,
	)
	if strings.Contains(html, `id="tag-pets-children" x-show="open" style="display: none;"`) {
		t.Fatalf("open collapsible child groups should render visible before Alpine initializes: %s", html)
	}
	if count := strings.Count(html, `href="#operation-get-pets"`); count != 1 {
		t.Fatalf("collapsible children rendered %d links, want 1: %s", count, html)
	}
}

func TestCollapsibleSectionCollapsesChildGroupsByDefault(t *testing.T) {
	html := renderSidebar(t, Config{
		Sections: []Section{
			{
				Title:       "Operations",
				Collapsible: true,
				Items: []Item{
					{
						ID:    "tag-pets",
						Label: "Pets",
						Href:  "#tag-pets",
						Items: []Item{
							{ID: "get-pets", Label: "GET /pets", Href: "#operation-get-pets"},
						},
					},
					{
						ID:    "tag-stores",
						Label: "Stores",
						Href:  "#tag-stores",
						Items: []Item{
							{ID: "get-stores", Label: "GET /stores", Href: "#operation-get-stores"},
						},
					},
				},
			},
		},
	})

	assertContainsAll(t, html,
		`aria-controls="tag-pets-children"`,
		`aria-controls="tag-stores-children"`,
		`id="tag-pets-children"`,
		`id="tag-stores-children"`,
		`x-on:click.prevent="open = !open"`,
		`x-show="open"`,
		`style="display: none;"`,
		`href="#operation-get-pets"`,
		`href="#operation-get-stores"`,
	)

	if count := strings.Count(html, `x-data="{ open: false }"`); count != 2 {
		t.Fatalf("collapsible section should render each child group closed with independent state, got %d scopes: %s", count, html)
	}
	if strings.Contains(html, `x-data="{ open: true }"`) {
		t.Fatalf("collapsible section child groups should default closed: %s", html)
	}
}

func TestSidebarOverlayRendersNativeOffCanvasShell(t *testing.T) {
	html := renderOverlay(t, OverlayConfig{
		ID: "docs-nav",
		Sidebar: Config{
			LogoText: "API",
			Items: []Item{{
				ID:     "overview",
				Label:  "Overview",
				Href:   "#overview",
				Active: true,
			}},
		},
		TriggerLabel:          "Open API navigation",
		RootClass:             "lg:hidden",
		PanelPositionClass:    "fixed top-16 bottom-0 left-0",
		PanelWidthClass:       "w-72",
		BackdropPositionClass: "fixed top-16 bottom-0 inset-x-0",
	})

	assertContainsAll(t, html,
		`x-data="{ docsNavOpen: false }"`,
		`class="lg:hidden"`,
		`type="button"`,
		`aria-label="Open API navigation"`,
		`aria-controls="docs-nav-panel"`,
		`x-on:click="docsNavOpen = !docsNavOpen"`,
		`x-bind:aria-expanded="docsNavOpen.toString()"`,
		`x-show="docsNavOpen"`,
		`x-on:click="docsNavOpen = false"`,
		`class="fixed top-16 bottom-0 inset-x-0 z-30 bg-black/50"`,
		`id="docs-nav-panel"`,
		`class="fixed top-16 bottom-0 left-0 z-40 w-72"`,
		`x-on:click="if ($event.target.closest(&#39;a[href]:not([aria-controls])&#39;)) docsNavOpen = false"`,
		`API`,
		`Overview`,
	)

	for _, absent := range []string{`x-trap`, `role="dialog"`, `aria-modal="true"`} {
		if strings.Contains(html, absent) {
			t.Fatalf("sidebar overlay should match the docs navigation drawer and avoid modal focus behavior; found %q in %s", absent, html)
		}
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

func renderOverlay(t *testing.T, cfg OverlayConfig) string {
	t.Helper()

	var buf bytes.Buffer
	if err := Overlay(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render sidebar overlay: %v", err)
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

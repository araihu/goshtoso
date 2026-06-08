package search

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestSearchRendersTriggerKbdAndResults(t *testing.T) {
	var buf bytes.Buffer
	err := Search(Config{
		ID:          "docs-search",
		Label:       "Search components",
		Placeholder: "Search docs...",
		Items: []Item{
			{ID: "result-kbd", Title: "KBD", Description: "Keyboard hints", Href: "/components/kbd", Section: "Display", Keywords: []string{"shortcut"}},
		},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="docs-search"`,
		`x-data="goshtosoSearchField(&#39;docs-search&#39;, false)"`,
		`x-data="goshtosoSearchModal(&#39;docs-search&#39;, 4, 120)"`,
		`aria-haspopup="dialog"`,
		`<kbd`,
		`⌘ K`,
		`data-search-title="KBD"`,
		`data-search-description="Keyboard hints"`,
		`data-search-href="/components/kbd"`,
		`data-search-text="KBD Keyboard hints Display shortcut"`,
		`Search docs...`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("Search render missing %q in %s", want, html)
		}
	}

	if strings.Contains(html, `x-on:keydown.window`) {
		t.Fatalf("Search should not bind a global shortcut unless GlobalShortcut is true: %s", html)
	}
}

func TestSearchRendersGlobalShortcutWhenEnabled(t *testing.T) {
	var buf bytes.Buffer
	err := SearchField(Config{
		ID:             "docs-search",
		GlobalShortcut: true,
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), `x-on:keydown.window="handleWindowKey($event)"`) {
		t.Fatalf("Search render missing global shortcut binding in %s", buf.String())
	}
}

func TestSearchModalCanLoadResultsFromItemsURL(t *testing.T) {
	html := renderHTML(t, SearchModal(Config{
		ID:       "remote-search",
		ItemsURL: "/search.json",
		Items: []Item{{
			ID:    "static-result",
			Title: "Static result that should not render",
		}},
	}))

	for _, want := range []string{
		`data-search-source-url="/search.json"`,
		`<template x-for="(item, index) in visibleResults()"`,
		`x-bind:id="item.id || null"`,
		`x-on:click="selectResult(item)"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("SearchModal with ItemsURL missing %q in\n%s", want, html)
		}
	}
	if strings.Contains(html, "Static result that should not render") || strings.Contains(html, `id="static-result"`) {
		t.Fatalf("ItemsURL mode should not pre-render result records:\n%s", html)
	}
}

func TestSearchScriptCachesDOMResultMatches(t *testing.T) {
	html := renderHTML(t, SearchModal(Config{ID: "docs-search"}))

	for _, want := range []string{
		`cachedDOMTerm: null`,
		`cachedDOMResults: []`,
		`if (term === this.cachedDOMTerm) return this.cachedDOMResults;`,
		`cachedAllResults`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("SearchModal script missing DOM match cache %q in\n%s", want, html)
		}
	}
}

func TestSearchEscapesResultPayloadsBeforeHighlighting(t *testing.T) {
	payload := `<img src=x onerror=alert(1)>`
	var buf bytes.Buffer
	err := Search(Config{
		ID: "docs-search",
		Items: []Item{{
			ID:          "result-xss",
			Title:       payload,
			Description: payload,
			Href:        `javascript:alert(1)`,
			Section:     payload,
			Keywords:    []string{payload},
		}},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	if strings.Contains(html, payload) {
		t.Fatalf("rendered raw result payload:\n%s", html)
	}
	for _, want := range []string{
		`data-search-title="&lt;img src=x onerror=alert(1)&gt;"`,
		`data-search-description="&lt;img src=x onerror=alert(1)&gt;"`,
		`&lt;img src=x onerror=alert(1)&gt;`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("Search render missing escaped payload %q in %s", want, html)
		}
	}
	if strings.Contains(html, `data-search-href="javascript:alert(1)"`) {
		t.Fatalf("Search exposes executable javascript: href through client navigation sink:\n%s", html)
	}
}

func TestSearchSafeHrefFiltersExecutableNavigationTargets(t *testing.T) {
	cases := []struct {
		name string
		href string
		want string
	}{
		{name: "relative path", href: "/components/kbd", want: `/components/kbd`},
		{name: "relative path trims whitespace", href: " /components/kbd ", want: `/components/kbd`},
		{name: "https", href: "https://example.test/docs", want: `https://example.test/docs`},
		{name: "mixed case javascript", href: "JaVaScRiPt:alert(1)", want: ``},
		{name: "javascript with whitespace", href: " javascript:alert(1) ", want: ``},
		{name: "data scheme", href: "data:text/html,<script>alert(1)</script>", want: ``},
		{name: "invalid url", href: ":not-a-url", want: ``},
		{name: "protocol relative is treated as relative", href: "//example.test/docs", want: `//example.test/docs`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Search(Config{
				ID: "docs-search",
				Items: []Item{{
					ID:    "result",
					Title: "Result",
					Href:  tc.href,
				}},
			}).Render(context.Background(), &buf)
			if err != nil {
				t.Fatal(err)
			}

			html := buf.String()
			if tc.want == "" {
				if strings.Contains(html, "data-search-href") {
					t.Fatalf("unsafe href %q should not be emitted:\n%s", tc.href, html)
				}
				return
			}
			wantAttr := `data-search-href="` + tc.want + `"`
			if !strings.Contains(html, wantAttr) {
				t.Fatalf("safe href %q did not render as %q:\n%s", tc.href, wantAttr, html)
			}
		})
	}
}

func TestSearchDefaults(t *testing.T) {
	cfg := Config{}
	if cfg.GetID() != "search" {
		t.Fatalf("default ID = %q", cfg.GetID())
	}
	if cfg.GetLabel() != "Search" {
		t.Fatalf("default label = %q", cfg.GetLabel())
	}
	if cfg.GetShortcutText() == "" || cfg.GetEscapeText() == "" || cfg.GetEmptyText() == "" {
		t.Fatalf("default shortcut, escape, and empty text should be populated")
	}
	if cfg.GetMaxResults() != 4 {
		t.Fatalf("default max results = %d", cfg.GetMaxResults())
	}
	if cfg.GetDescriptionMaxLength() != 120 {
		t.Fatalf("default description max length = %d", cfg.GetDescriptionMaxLength())
	}
}

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

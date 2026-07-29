package search

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// renderSearch renders the given component and returns its HTML, failing the
// test on a render error.
func renderHTML(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	return buf.String()
}

// TestConfigCustomValueBranches exercises the non-default branch of every
// Get* accessor and the custom-class branch of every class helper. The default
// branches are covered by TestSearchDefaults in search_test.go.
func TestConfigCustomValueBranches(t *testing.T) {
	cfg := Config{
		ID:                   "custom-id",
		Label:                "Find anything",
		Placeholder:          "Type to search...",
		ShortcutText:         "Ctrl K",
		EscapeText:           "ESC",
		EmptyText:            "Nothing here.",
		MaxResults:           9,
		DescriptionMaxLength: 42,
		RootClass:            "root-extra",
		TriggerClass:         "trigger-extra",
		DialogClass:          "dialog-extra",
	}

	checks := []struct {
		name string
		got  string
		want string
	}{
		{"ID", cfg.getID(), "custom-id"},
		{"Label", cfg.getLabel(), "Find anything"},
		{"Placeholder", cfg.getPlaceholder(), "Type to search..."},
		{"ShortcutText", cfg.getShortcutText(), "Ctrl K"},
		{"EscapeText", cfg.getEscapeText(), "ESC"},
		{"EmptyText", cfg.getEmptyText(), "Nothing here."},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}

	if cfg.getMaxResults() != 9 {
		t.Errorf("getMaxResults = %d, want 9", cfg.getMaxResults())
	}
	if cfg.getDescriptionMaxLength() != 42 {
		t.Errorf("getDescriptionMaxLength = %d, want 42", cfg.getDescriptionMaxLength())
	}

	classChecks := []struct {
		name string
		got  string
		want string
	}{
		{"rootClasses", cfg.rootClasses(), "root-extra"},
		{"triggerClasses", cfg.triggerClasses(), "trigger-extra"},
		{"dialogClasses", cfg.dialogClasses(), "dialog-extra"},
	}
	for _, c := range classChecks {
		if !strings.Contains(c.got, c.want) {
			t.Errorf("%s = %q, want it to contain %q", c.name, c.got, c.want)
		}
	}
}

// TestRootClassesDefaultOmitsExtra confirms the default branch leaves the base
// classes untouched.
func TestRootClassesDefaultOmitsExtra(t *testing.T) {
	got := Config{}.rootClasses()
	if got != "w-full" {
		t.Fatalf("rootClasses default = %q, want %q", got, "w-full")
	}
}

// TestSearchModalCustomValues renders SearchModal directly so its custom
// max-results, description length, label, placeholder, empty text, escape hint,
// dialog class, and spread InputAttrs branches all execute.
func TestSearchModalCustomValues(t *testing.T) {
	html := renderHTML(t, SearchModal(Config{
		ID:                   "modal-search",
		Label:                "Docs lookup",
		Placeholder:          "Search the docs",
		EscapeText:           "Close",
		EmptyText:            "No docs match.",
		MaxResults:           7,
		DescriptionMaxLength: 80,
		DialogClass:          "ring-2",
		InputAttrs:           templ.Attributes{"data-testid": "modal-input", "maxlength": "64"},
		Items: []Item{
			{ID: "doc-1", Title: "Alpha", Description: "First", Section: "Guides", Href: "/alpha"},
		},
	}))

	for _, want := range []string{
		`x-data="goshtosoSearchModal($el)"`,
		`data-search-id="modal-search"`,
		`data-search-max-results="7"`,
		`data-search-description-max-length="80"`,
		`id="modal-search-dialog"`,
		`aria-labelledby="modal-search-label"`,
		`aria-label="Docs lookup results"`,
		`placeholder="Search the docs"`,
		`data-testid="modal-input"`,
		`maxlength="64"`,
		`Close`,
		`No docs match.`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("SearchModal missing %q in\n%s", want, html)
		}
	}
}

// TestResultItemMinimalBranches renders an item with no ID, section,
// description, or href so the false side of each conditional attribute in
// resultItem executes.
func TestResultItemMinimalBranches(t *testing.T) {
	html := renderHTML(t, SearchModal(Config{
		ID: "minimal-search",
		Items: []Item{
			{Title: "Bare result"},
		},
	}))

	if !strings.Contains(html, `data-search-title="Bare result"`) {
		t.Fatalf("missing bare result title in\n%s", html)
	}
	// No section, description, href, or explicit id should be emitted.
	for _, unwanted := range []string{
		`data-search-section=`,
		`data-search-href=`,
		`id="minimal-search-result"`,
	} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("unexpected %q rendered for minimal item in\n%s", unwanted, html)
		}
	}
	// A description paragraph should not render when Description is empty.
	if strings.Contains(html, `truncate($el.closest`) {
		t.Fatalf("description paragraph rendered for item without description in\n%s", html)
	}
}

// TestResultItemFullBranches renders an item with every optional field set,
// including spread Attrs, so the true side of each conditional executes.
func TestResultItemFullBranches(t *testing.T) {
	html := renderHTML(t, SearchModal(Config{
		ID: "full-search",
		Items: []Item{
			{
				ID:          "full-result",
				Title:       "Complete",
				Description: "Has every field",
				Section:     "Reference",
				Href:        "/complete",
				Keywords:    []string{"kw1", "kw2"},
				Attrs:       templ.Attributes{"data-extra": "yes"},
			},
		},
	}))

	for _, want := range []string{
		`id="full-result"`,
		`data-search-section="Reference"`,
		`data-search-href="/complete"`,
		`data-search-text="Complete Has every field Reference kw1 kw2"`,
		`data-extra="yes"`,
		`Has every field`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("full result missing %q in\n%s", want, html)
		}
	}
}

// TestSearchFieldCustomLabelAndShortcut renders SearchField with custom label
// and shortcut text and without a global shortcut so both the trigger label
// branch and the absent-keydown branch execute.
func TestSearchFieldCustomLabelAndShortcut(t *testing.T) {
	html := renderHTML(t, SearchField(Config{
		ID:           "field-search",
		Label:        "Quick find",
		ShortcutText: "F3",
	}))

	for _, want := range []string{
		`x-data="goshtosoSearchField($el)"`,
		`data-search-id="field-search"`,
		`data-search-global-shortcut="false"`,
		`aria-controls="field-search-dialog"`,
		`Quick find`,
		`F3`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("SearchField missing %q in\n%s", want, html)
		}
	}
	if strings.Contains(html, `x-on:keydown.window`) {
		t.Fatalf("SearchField bound a window shortcut without GlobalShortcut:\n%s", html)
	}
}

// TestSearchComposesFieldAndModal confirms the top-level Search wires both the
// trigger and the modal under one root and emits the empty-state copy.
func TestSearchComposesFieldAndModal(t *testing.T) {
	html := renderHTML(t, Search(Config{
		ID:        "combo-search",
		EmptyText: "Try another term.",
		Items: []Item{
			{ID: "r1", Title: "Result one", Section: "S", Description: "D", Href: "/r1"},
		},
	}))

	for _, want := range []string{
		`x-data="goshtosoSearchField($el)"`,
		`x-data="goshtosoSearchModal($el)"`,
		`data-search-id="combo-search"`,
		`id="combo-search-dialog"`,
		`Try another term.`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("Search missing %q in\n%s", want, html)
		}
	}
}

func TestSearchDynamicConfigUsesInertDataAttributes(t *testing.T) {
	html := renderHTML(t, Search(Config{
		ID:                   `docs'); alert(1); //`,
		GlobalShortcut:       true,
		MaxResults:           7,
		DescriptionMaxLength: 80,
	}))

	for _, want := range []string{
		`data-search-id="docs&#39;); alert(1); //"`,
		`data-search-global-shortcut="true"`,
		`data-search-max-results="7"`,
		`data-search-description-max-length="80"`,
		`x-data="goshtosoSearchField($el)"`,
		`x-data="goshtosoSearchModal($el)"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("Search inert config missing %q in\n%s", want, html)
		}
	}
	if strings.Contains(html, `goshtosoSearchField(&#39;docs`) || strings.Contains(html, `<script`) {
		t.Fatalf("Search embedded instance data into executable JavaScript:\n%s", html)
	}
}

// TestItemSearchTextTrimsAndJoins covers SearchText with and without keywords
// and confirms surrounding whitespace is trimmed.
func TestItemSearchTextTrimsAndJoins(t *testing.T) {
	withKeywords := Item{Title: "T", Description: "D", Section: "S", Keywords: []string{"a", "b"}}
	if got := withKeywords.SearchText(); got != "T D S a b" {
		t.Errorf("SearchText with keywords = %q", got)
	}

	titleOnly := Item{Title: "Only"}
	if got := titleOnly.SearchText(); got != "Only" {
		t.Errorf("SearchText title only = %q, want %q", got, "Only")
	}
}

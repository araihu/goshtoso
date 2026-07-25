package tagslist

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func renderTagsList(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := TagsList(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render tagslist: %v", err)
	}
	return buf.String()
}

func TestGetAddLabelDefaultAndCustom(t *testing.T) {
	if got := (Config{}).getAddLabel(); got != "Add" {
		t.Fatalf("default add label = %q, want %q", got, "Add")
	}
	if got := (Config{AddActionLabel: "Append"}).getAddLabel(); got != "Append" {
		t.Fatalf("custom add label = %q, want %q", got, "Append")
	}
}

func TestGetPlaceholderDefaultAndCustom(t *testing.T) {
	if got := (Config{}).getPlaceholder(); got != "Add a tag..." {
		t.Fatalf("default placeholder = %q, want %q", got, "Add a tag...")
	}
	if got := (Config{Placeholder: "Type here"}).getPlaceholder(); got != "Type here" {
		t.Fatalf("custom placeholder = %q, want %q", got, "Type here")
	}
}

func TestContainerClassesDefaultAndRootClass(t *testing.T) {
	base := (Config{}).containerClasses()
	if base != "flex flex-col gap-2" {
		t.Fatalf("default container classes = %q", base)
	}
	withRoot := (Config{RootClass: "mt-4 custom"}).containerClasses()
	if withRoot != "flex flex-col gap-2 mt-4 custom" {
		t.Fatalf("rootclass container classes = %q", withRoot)
	}
}

func TestAlpineDataEmptyValues(t *testing.T) {
	got := (Config{Name: "tags"}).alpineData()
	if !strings.Contains(got, "items: []") {
		t.Fatalf("empty values should yield empty items array: %s", got)
	}
	if !strings.Contains(got, "name: 'tags'") {
		t.Fatalf("expected name in alpine data: %s", got)
	}
	if !strings.Contains(got, "newTag: ''") {
		t.Fatalf("expected newTag in alpine data: %s", got)
	}
}

func TestAlpineDataMultipleValuesCommaSeparated(t *testing.T) {
	got := (Config{Name: "tags", Values: []string{"a", "b", "c"}}).alpineData()
	if !strings.Contains(got, "items: ['a','b','c']") {
		t.Fatalf("multiple values should be comma-separated single-quoted: %s", got)
	}
}

func TestAlpineDataEscapesBackslash(t *testing.T) {
	got := (Config{Name: `na\me`, Values: []string{`a\b`}}).alpineData()
	if !strings.Contains(got, `'a\\b'`) {
		t.Fatalf("backslash in value should be escaped: %s", got)
	}
	if !strings.Contains(got, `name: 'na\\me'`) {
		t.Fatalf("backslash in name should be escaped: %s", got)
	}
}

func TestRenderWithIDIncludesIDAttr(t *testing.T) {
	html := renderTagsList(t, Config{ID: "mytags", Name: "tags"})
	if !strings.Contains(html, `id="mytags"`) {
		t.Fatalf("expected id attribute, got: %s", html)
	}
}

func TestRenderWithoutIDOmitsIDAttr(t *testing.T) {
	html := renderTagsList(t, Config{Name: "tags"})
	if strings.Contains(html, "id=") {
		t.Fatalf("expected no id attribute when ID empty, got: %s", html)
	}
	if !strings.Contains(html, "data-tagslist") {
		t.Fatalf("expected data-tagslist marker, got: %s", html)
	}
}

func TestRenderEnabledHasInputAddAndRemove(t *testing.T) {
	html := renderTagsList(t, Config{ID: "t", Name: "tags", Values: []string{"x"}})
	for _, want := range []string{
		"data-tagslist-input",
		"data-tagslist-add",
		`aria-label="Remove tag"`,
		"addTag()",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("enabled render missing %q in %s", want, html)
		}
	}
}

func TestRenderDisabledOmitsInputAddAndRemove(t *testing.T) {
	html := renderTagsList(t, Config{
		ID:       "t",
		Name:     "locked",
		Values:   []string{"locked", "readonly"},
		Disabled: true,
	})
	// Chips still render.
	if !strings.Contains(html, "data-tagslist-chips") {
		t.Fatalf("disabled render should still include chips: %s", html)
	}
	for _, unwanted := range []string{
		"data-tagslist-input",
		"data-tagslist-add",
		`aria-label="Remove tag"`,
	} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("disabled render should not include %q in %s", unwanted, html)
		}
	}
}

func TestRenderUsesCustomAddLabelAndPlaceholder(t *testing.T) {
	html := renderTagsList(t, Config{
		ID:             "t",
		Name:           "tags",
		AddActionLabel: "Insert",
		Placeholder:    "New tag name",
	})
	if !strings.Contains(html, "Insert") {
		t.Fatalf("expected custom add label in render: %s", html)
	}
	if !strings.Contains(html, `placeholder="New tag name"`) {
		t.Fatalf("expected custom placeholder in render: %s", html)
	}
}

func TestRenderAppliesRootClass(t *testing.T) {
	html := renderTagsList(t, Config{Name: "tags", RootClass: "border-test"})
	if !strings.Contains(html, "border-test") {
		t.Fatalf("expected RootClass applied to container: %s", html)
	}
}

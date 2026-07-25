package structuredinput

import (
	"context"
	"strings"
	"testing"
)

func renderConfig(t *testing.T, cfg Config) string {
	t.Helper()
	var buf strings.Builder
	if err := StructuredInput(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return buf.String()
}

func TestOptionLabelFallsBackToValue(t *testing.T) {
	if got := (Option{Value: "NoExecute"}).optionLabel(); got != "NoExecute" {
		t.Fatalf("optionLabel() = %q, want value fallback", got)
	}
	if got := (Option{Value: "x", Label: "Explicit"}).optionLabel(); got != "Explicit" {
		t.Fatalf("optionLabel() = %q, want explicit label", got)
	}
}

func TestGetAddLabelUsesCustomAndDefault(t *testing.T) {
	if got := (Config{}).getAddLabel(); got != "Add row" {
		t.Fatalf("getAddLabel() = %q, want default", got)
	}
	if got := (Config{AddActionLabel: "Add taint"}).getAddLabel(); got != "Add taint" {
		t.Fatalf("getAddLabel() = %q, want custom", got)
	}
}

func TestContainerClassesAppendsRootClass(t *testing.T) {
	base := (Config{}).containerClasses()
	if base != "flex flex-col gap-2" {
		t.Fatalf("containerClasses() = %q, want base only", base)
	}
	withRoot := (Config{RootClass: "mt-4 w-full"}).containerClasses()
	if !strings.HasPrefix(withRoot, "flex flex-col gap-2 ") || !strings.Contains(withRoot, "mt-4 w-full") {
		t.Fatalf("containerClasses() = %q, want base plus RootClass", withRoot)
	}
}

func TestDefaultValueBranches(t *testing.T) {
	if got := (Column{Default: "explicit"}).defaultValue(); got != "explicit" {
		t.Fatalf("defaultValue() = %q, want explicit default", got)
	}
	sel := Column{Type: ColumnSelect, Options: []Option{{Value: "first"}, {Value: "second"}}}
	if got := sel.defaultValue(); got != "first" {
		t.Fatalf("defaultValue() = %q, want first option", got)
	}
	if got := (Column{Type: ColumnSelect}).defaultValue(); got != "" {
		t.Fatalf("defaultValue() = %q, want empty for optionless select", got)
	}
	if got := (Column{}).defaultValue(); got != "" {
		t.Fatalf("defaultValue() = %q, want empty text default", got)
	}
}

// TestDisabledRendersDisabledControlsAndHidesActions exercises the disabled
// branch in textColumn/selectColumn and the !cfg.Disabled false branches in
// StructuredInput (no add button, no remove button).
func TestDisabledRendersDisabledControlsAndHidesActions(t *testing.T) {
	html := renderConfig(t, Config{
		ID:       "disabledDemo",
		Name:     "labels",
		Disabled: true,
		Columns: []Column{
			{Key: "key", Label: "Key", Placeholder: "key"},
			{Key: "effect", Label: "Effect", Type: ColumnSelect, Options: []Option{{Value: "NoSchedule"}}},
		},
		Entries: []Entry{{"key": "app", "effect": "NoSchedule"}},
	})

	if strings.Count(html, "disabled") < 2 {
		t.Fatalf("disabled render missing disabled attributes:\n%s", html)
	}
	if strings.Contains(html, "data-add-row") {
		t.Fatalf("disabled render must not include add button:\n%s", html)
	}
	if strings.Contains(html, `aria-label="Remove row"`) {
		t.Fatalf("disabled render must not include remove button:\n%s", html)
	}
}

// TestEnabledRendersAddAndRemoveActions covers the !cfg.Disabled true branches.
func TestEnabledRendersAddAndRemoveActions(t *testing.T) {
	html := renderConfig(t, Config{
		Name:           "labels",
		AddActionLabel: "Add row",
		Columns:        []Column{{Key: "key", Label: "Key"}},
		Entries:        []Entry{{"key": "app"}},
	})

	if !strings.Contains(html, "data-add-row") {
		t.Fatalf("enabled render missing add button:\n%s", html)
	}
	if !strings.Contains(html, `aria-label="Remove row"`) {
		t.Fatalf("enabled render missing remove button:\n%s", html)
	}
}

// TestRenderWithoutIDOmitsIDAttribute covers the cfg.ID == "" branch.
func TestRenderWithoutIDOmitsIDAttribute(t *testing.T) {
	html := renderConfig(t, Config{
		Name:    "labels",
		Columns: []Column{{Key: "key", Label: "Key"}},
	})

	if strings.Contains(html, " id=") {
		t.Fatalf("render without ID must omit id attribute:\n%s", html)
	}
	if !strings.Contains(html, `x-data="structuredInput($el)"`) {
		t.Fatalf("render missing x-data root:\n%s", html)
	}
}

// TestSelectOptionLabelFallbackRenders ensures an option without an explicit
// Label renders its Value as the visible text.
func TestSelectOptionLabelFallbackRenders(t *testing.T) {
	html := renderConfig(t, Config{
		Name: "taints",
		Columns: []Column{
			{Key: "effect", Type: ColumnSelect, Options: []Option{{Value: "NoExecute"}}},
		},
	})

	if !strings.Contains(html, `<option value="NoExecute">NoExecute</option>`) {
		t.Fatalf("select missing value-fallback option text:\n%s", html)
	}
}

// TestEntriesJSONOrdersValuesByColumn confirms entry rows serialize in column
// order, dropping keys not present in the schema.
func TestEntriesJSONOrdersValuesByColumn(t *testing.T) {
	cfg := Config{
		Name: "labels",
		Columns: []Column{
			{Key: "key"},
			{Key: "value"},
		},
		Entries: []Entry{
			{"value": "web", "key": "app", "ignored": "x"},
		},
	}

	if got := cfg.entriesJSON(); got != `[["app","web"]]` {
		t.Fatalf("entriesJSON() = %s, want column-ordered rows", got)
	}
}

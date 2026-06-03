package structuredinput

import (
	"context"
	"strings"
	"testing"
)

func TestNormalizedColumnsDropsEmptyAndDuplicateKeys(t *testing.T) {
	cfg := Config{
		Columns: []Column{
			{Key: "", Placeholder: "skip"},
			{Key: "key", Placeholder: "first"},
			{Key: "key", Placeholder: "duplicate"},
			{Key: "effect", Type: ColumnSelect, Options: []Option{{Value: "NoSchedule", Label: "NoSchedule"}}},
		},
	}

	cols := cfg.NormalizedColumns()

	if len(cols) != 2 {
		t.Fatalf("len(cols) = %d, want 2", len(cols))
	}
	if cols[0].Key != "key" || cols[0].Type != ColumnText {
		t.Fatalf("first column = %#v, want key text column", cols[0])
	}
	if cols[1].Key != "effect" || cols[1].DefaultValue() != "NoSchedule" {
		t.Fatalf("second column = %#v, want effect select defaulting to first option", cols[1])
	}
}

func TestInitialEntriesNeverSerializesNull(t *testing.T) {
	cfg := Config{Name: "labels", Entries: nil}

	data := cfg.EntriesJSON()

	if strings.Contains(data, "entries: null") {
		t.Fatalf("EntriesJSON() = %s, must use [] for nil entries", data)
	}
	if data != "[]" {
		t.Fatalf("EntriesJSON() = %s, want []", data)
	}
}

func TestEntriesJSONEscapesStrings(t *testing.T) {
	cfg := Config{
		Name: "labels",
		Columns: []Column{
			{Key: "key", Placeholder: "owner's key"},
			{Key: "value"},
		},
		Entries: []Entry{
			{"key": "team's app", "value": `web\api`},
		},
	}

	data := cfg.EntriesJSON()

	if !strings.Contains(data, `"team's app"`) {
		t.Fatalf("EntriesJSON() = %s, want JSON entry value", data)
	}
	if !strings.Contains(data, `web\\api`) {
		t.Fatalf("EntriesJSON() = %s, want escaped backslash", data)
	}
}

func TestNewRowJSONUsesColumnDefaults(t *testing.T) {
	cfg := Config{
		Columns: []Column{
			{Key: "key"},
			{Key: "effect", Type: ColumnSelect, Options: []Option{{Value: "NoSchedule", Label: "NoSchedule"}}},
			{Key: "priority", Default: "high"},
		},
	}

	row := cfg.NewRowJSON()

	for _, want := range []string{`""`, `"NoSchedule"`, `"high"`} {
		if !strings.Contains(row, want) {
			t.Fatalf("NewRowJSON() = %s, missing %s", row, want)
		}
	}
}

func TestColumnAccessorsUseBracketNotation(t *testing.T) {
	col := Column{Key: "app.kubernetes.io/name"}

	if got := col.EntryAccessor(2); got != "entry[2]" {
		t.Fatalf("EntryAccessor() = %s, want array index notation", got)
	}
	if got := col.NameBinding(); got != "inputName(index, $el.dataset.columnKey)" {
		t.Fatalf("NameBinding() = %s, want data-key based name binding", got)
	}
}

func TestStructuredInputRendersHiddenStructuredNames(t *testing.T) {
	var buf strings.Builder
	err := StructuredInput(Config{
		ID:   "labelsDemo",
		Name: "labels",
		Columns: []Column{
			{Key: "key", Placeholder: "key"},
			{Key: "value", Placeholder: "value"},
		},
		Entries: []Entry{{"key": "app", "value": "web"}},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="labelsDemo"`,
		`x-data="structuredInput($el)"`,
		`data-column-key="key"`,
		`data-column-key="value"`,
		`x-bind:name="inputName(index, $el.dataset.columnKey)"`,
		`x-bind:value="entry[0]"`,
		`x-bind:value="entry[1]"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %s:\n%s", want, html)
		}
	}
}

func TestStructuredInputRendersSelectColumn(t *testing.T) {
	var buf strings.Builder
	err := StructuredInput(Config{
		ID:   "taintsDemo",
		Name: "taints",
		Columns: []Column{
			{Key: "key"},
			{Key: "effect", Type: ColumnSelect, Options: []Option{{Value: "NoSchedule", Label: "NoSchedule"}}},
		},
		Entries: []Entry{{"key": "node", "effect": "NoSchedule"}},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	html := buf.String()
	for _, want := range []string{`<select`, `x-model="entry[1]"`, `<option value="NoSchedule">NoSchedule</option>`} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %s:\n%s", want, html)
		}
	}
}

func TestStructuredInputRendersColumnSeparators(t *testing.T) {
	var buf strings.Builder
	err := StructuredInput(Config{
		ID:   "taintsDemo",
		Name: "taints",
		Columns: []Column{
			{Key: "key", Separator: "="},
			{Key: "value", Separator: ":"},
			{Key: "effect", Type: ColumnSelect, Options: []Option{{Value: "NoSchedule", Label: "NoSchedule"}}},
		},
		Entries: []Entry{{"key": "node", "value": "true", "effect": "NoSchedule"}},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	html := buf.String()
	for _, want := range []string{`data-column-separator="true">=</span>`, `data-column-separator="true">:</span>`} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %s:\n%s", want, html)
		}
	}
}

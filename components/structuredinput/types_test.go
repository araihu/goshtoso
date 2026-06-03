package structuredinput

import (
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

	data := cfg.AlpineData()

	if strings.Contains(data, "entries: null") {
		t.Fatalf("AlpineData() = %s, must use [] for nil entries", data)
	}
	if !strings.Contains(data, "entries: []") {
		t.Fatalf("AlpineData() = %s, want entries: []", data)
	}
}

func TestAlpineDataEscapesSingleQuotedStrings(t *testing.T) {
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

	data := cfg.AlpineData()

	if !strings.Contains(data, `owner\'s key`) {
		t.Fatalf("AlpineData() = %s, want escaped placeholder", data)
	}
	if !strings.Contains(data, `team\'s app`) {
		t.Fatalf("AlpineData() = %s, want escaped entry value", data)
	}
	if !strings.Contains(data, `web\\api`) {
		t.Fatalf("AlpineData() = %s, want escaped backslash", data)
	}
}

func TestNewRowLiteralUsesColumnDefaults(t *testing.T) {
	cfg := Config{
		Columns: []Column{
			{Key: "key"},
			{Key: "effect", Type: ColumnSelect, Options: []Option{{Value: "NoSchedule", Label: "NoSchedule"}}},
			{Key: "priority", Default: "high"},
		},
	}

	row := cfg.NewRowLiteral()

	for _, want := range []string{"'key': ''", "'effect': 'NoSchedule'", "'priority': 'high'"} {
		if !strings.Contains(row, want) {
			t.Fatalf("NewRowLiteral() = %s, missing %s", row, want)
		}
	}
}

func TestColumnAccessorsUseBracketNotation(t *testing.T) {
	col := Column{Key: "app.kubernetes.io/name"}

	if got := col.EntryAccessor(); got != "entry['app.kubernetes.io/name']" {
		t.Fatalf("EntryAccessor() = %s, want bracket notation", got)
	}
	if got := col.NameBinding(); got != "name + '[' + index + '][app.kubernetes.io/name]'" {
		t.Fatalf("NameBinding() = %s, want structured input name binding", got)
	}
}

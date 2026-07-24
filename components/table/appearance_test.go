package table

import (
	"strings"
	"testing"
)

func TestStripedAppearance(t *testing.T) {
	html := renderT(t, Table(Config{
		Appearance: AppearanceStriped,
		Columns:    []Column{{Key: "name", Label: "Name"}},
		Rows:       []Row{{ID: "1", Cells: map[string]Cell{"name": {Text: "Ada"}}}},
	}))
	if !strings.Contains(html, "odd:bg-surface-alt") {
		t.Fatalf("striped appearance must render striped row classes: %s", html)
	}
}

func TestCheckboxSelectionIsBehaviorNotAppearance(t *testing.T) {
	html := renderT(t, Table(Config{ShowCheckbox: true}))
	if !strings.Contains(html, `type="checkbox"`) {
		t.Fatalf("ShowCheckbox must render selection inputs: %s", html)
	}
}

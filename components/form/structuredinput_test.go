package form

import (
	"context"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/structuredinput"
)

func TestFieldGroupRendersStructuredInput(t *testing.T) {
	var buf strings.Builder
	err := FieldGroup(FieldGroupConfig{
		ID:    "labels",
		Label: "Labels",
		StructuredInput: &structuredinput.Config{
			ID:   "labelsInput",
			Name: "labels",
			Columns: []structuredinput.Column{
				{Key: "key"},
				{Key: "value"},
			},
			Entries: []structuredinput.Entry{{"key": "app", "value": "web"}},
		},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	html := buf.String()
	for _, want := range []string{`Labels`, `id="labelsInput"`, `x-bind:name="inputName(index, $el.dataset.columnKey)"`} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %s:\n%s", want, html)
		}
	}
}

package combobox

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/expressions"
)

func TestExpressionOverrideInOOBSelectionLabel(t *testing.T) {
	cfg := Config{ID: "oob-expressions", Mode: ModeMultiple}
	state := State{Selected: []string{"a", "b"}}
	ctx := expressions.With(context.Background(), expressions.Set{Combobox: expressions.Combobox{SelectedLabel: func(n int) string { return fmt.Sprintf("Escolhidos: %d", n) }}})
	var out bytes.Buffer
	if err := triggerLabelOOB(cfg, state).Render(ctx, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Escolhidos: 2") || strings.Contains(out.String(), "2 selected") {
		t.Fatal(out.String())
	}
}

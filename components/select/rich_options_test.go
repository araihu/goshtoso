package selectfield

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestRichOptionPreservesLabelStyleAndKeyboard(t *testing.T) {
	cfg := Config{ID: "themes", Options: []Option{{Value: "lab", Label: "Lab theme", Selected: true, Leading: templ.Raw(`<span data-swatch>Color</span>`), LabelAttrs: templ.Attributes{"style": `font-family:"Lab Font"`, "data-font": "lab"}}}}
	var output bytes.Buffer
	if e := Select(cfg).Render(t.Context(), &output); e != nil {
		t.Fatal(e)
	}
	html := output.String()
	for _, want := range []string{`data-swatch`, `data-font="lab"`, `id="themes-option-0"`, `moveActiveOption($el,1)`, `selectOption(allOptions[0])`, `selectedOption.labelStyle`, `aria-selected`} {
		if !strings.Contains(html, want) {
			t.Fatal("missing rich option contract", want)
		}
	}
	var data factoryData
	if e := json.Unmarshal([]byte(cfg.factoryDataJSON()), &data); e != nil {
		t.Fatal(e)
	}
	if data.Options[0].LabelStyle != `font-family:"Lab Font"` || data.SelectedValues[0] != "lab" {
		t.Fatal(data)
	}
}

func TestTrailingOptionRendersInListAndTriggerWithoutChangingLabel(t *testing.T) {
	cfg := Config{ID: "origin", Options: []Option{{Value: "custom", Label: "My theme", Trailing: templ.Raw(`<span data-origin>Custom</span>`)}}}
	var output bytes.Buffer
	if err := Select(cfg).Render(t.Context(), &output); err != nil {
		t.Fatal(err)
	}
	if strings.Count(output.String(), `data-origin`) != 2 || !strings.Contains(output.String(), `id="origin-option-0"`) {
		t.Fatal("trailing-only option must render rich list and selected trigger content")
	}
	var data factoryData
	if err := json.Unmarshal([]byte(cfg.factoryDataJSON()), &data); err != nil {
		t.Fatal(err)
	}
	if data.Options[0].Label != "My theme" {
		t.Fatal("badge must not alter the option label")
	}
}

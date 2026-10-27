package themeeditor

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/themes"
)

func TestPublicThemeRuntime(t *testing.T) {
	node, e := exec.LookPath("node")
	if e != nil {
		t.Skip("Node unavailable")
	}
	type preset struct {
		Key        themes.Key        `json:"key"`
		Definition themes.Definition `json:"definition"`
		Edited     themes.Definition `json:"edited"`
	}
	presets := []preset{}
	color, _ := themes.ParseColor("#165d9b")
	for _, theme := range themes.BuiltIn() {
		d, e := themes.Preset(theme.Key)
		if e != nil {
			t.Fatal(e)
		}
		edited, e := themes.Materialize(d, themes.Overrides{Tokens: map[string]themes.Color{"primary": color}})
		if e != nil {
			t.Fatal(e)
		}
		presets = append(presets, preset{theme.Key, d, edited})
	}
	payload, e := json.Marshal(struct {
		Presets []preset            `json:"presets"`
		Tokens  []themes.TokenSpec  `json:"tokens"`
		Fonts   []themes.FontOption `json:"fonts"`
	}{presets, themes.Tokens(), themes.DefaultFonts()})
	if e != nil {
		t.Fatal(e)
	}
	cmd := exec.CommandContext(t.Context(), node, "runtime_test.mjs")
	cmd.Stdin = bytes.NewReader(payload)
	if output, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("theme runtime: %v\n%s", e, output)
	}
}
func TestControlledEditorHasNoPersistenceAndOwnsIDs(t *testing.T) {
	d, _ := themes.Preset(themes.KeyGoshtoso)
	var output bytes.Buffer
	for _, id := range []string{"first", "second"} {
		if e := ThemeEditor(Config{ID: id, Initial: d, PreviewID: id + "-preview"}).Render(t.Context(), &output); e != nil {
			t.Fatal(e)
		}
	}
	for _, bad := range []string{"<form", "localStorage", "/presentation/save"} {
		if bytes.Contains(output.Bytes(), []byte(bad)) {
			t.Fatal("consumer behavior leaked", bad)
		}
	}
	for _, id := range []string{`id="first-title"`, `id="second-title"`, `name="definition"`} {
		if !bytes.Contains(output.Bytes(), []byte(id)) {
			t.Fatal("missing control contract", id)
		}
	}
}

func TestFontPickersUseComboboxAndIsolatedTypefaceLabels(t *testing.T) {
	d, _ := themes.Preset(themes.KeyGoshtoso)
	var output bytes.Buffer
	if e := ThemeEditor(Config{ID: "fonts", Initial: d, PreviewID: "preview"}).Render(t.Context(), &output); e != nil {
		t.Fatal(e)
	}
	html := output.String()
	for _, want := range []string{`data-combobox-mode="client"`, `id="fonts-title-trigger"`, `data-theme-font-option="google/roboto"`, `goshtoso-preview-google-roboto`, `style="font-family:`, `data-combobox-search`, `@font-face{font-family:`} {
		if !strings.Contains(html, want) {
			t.Fatal("missing font picker contract", want)
		}
	}
	if strings.Contains(html, "<select") || strings.Contains(html, "fonts.googleapis.com") || strings.Contains(html, "@templ.Raw(") {
		t.Fatal("native font picker or eager Google stylesheet load")
	}
}

func TestThemeOptionsUseSnapshotColorsAndTitleTypeface(t *testing.T) {
	d, _ := themes.Preset(themes.KeyGoshtoso)
	d.Typography.Title = themes.FontRef{Source: "google", ID: "roboto"}
	cfg := Config{ID: "editor", Fonts: themes.DefaultFonts(), Catalog: []CatalogItem{{ID: "theme:one", Label: "My theme", Definition: d}}}
	options := cfg.PresetOptions()
	if len(options) != 1 || options[0].Value != "theme:one" || !strings.Contains(options[0].LabelAttrs["style"].(string), "goshtoso-preview-google-roboto") {
		t.Fatal(options)
	}
	var output bytes.Buffer
	if e := options[0].Leading.Render(t.Context(), &output); e != nil {
		t.Fatal(e)
	}
	for _, key := range []string{"primary", "secondary", "surface"} {
		if !strings.Contains(output.String(), d.Tokens[key].CSS()) {
			t.Fatal("missing literal swatch", key)
		}
	}
	output.Reset()
	if e := ThemeEditor(Config{ID: "editor", Initial: d}).Render(t.Context(), &output); e != nil {
		t.Fatal(e)
	}
	if strings.Count(output.String(), `data-theme-mode`) != 2 || !strings.Contains(output.String(), `<legend class="text-sm">Preview mode</legend>`) || strings.Contains(output.String(), `data-theme-mode-toggle`) {
		t.Fatal("preview mode should use two labelled radio segments")
	}
}

func TestPresetOptionsLoadNamedFacesWithoutChangingFontReferences(t *testing.T) {
	d, _ := themes.Preset(themes.KeyDracula)
	original := d.Typography.Title
	options := (Config{ID: "editor", Catalog: []CatalogItem{{ID: "dracula", Label: "Dracula", Definition: d}}}).PresetOptions()
	if !strings.Contains(options[0].LabelAttrs["style"].(string), "goshtoso-preview-google-fira-code") {
		t.Fatal(options[0].LabelAttrs)
	}
	if d.Typography.Title != original {
		t.Fatal("preview changed saved reference")
	}
}

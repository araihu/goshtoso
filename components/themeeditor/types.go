// Package themeeditor provides a controlled theme editor without persistence.
package themeeditor

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/araihu/goshtoso/components/combobox"

	"github.com/a-h/templ"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/themes"
)

type CatalogItem struct {
	ID         string            `json:"id"`
	Label      string            `json:"label"`
	Definition themes.Definition `json:"definition"`
}

// Config supplies the draft and public event contract. Actions belongs to the consumer.
type Config struct {
	ID        string
	Initial   themes.Definition
	Catalog   []CatalogItem
	Fonts     []themes.FontOption
	PreviewID string
	// ContrastID associates an external ContrastAdvisory and omits the inline panel.
	ContrastID         string
	InitialPreviewMode string
	InputName          string
	Actions            templ.Component
	RootClass          string
	RootAttrs          templ.Attributes
}

func (c Config) data() string {
	b, _ := json.Marshal(struct {
		ID          string              `json:"id"`
		Initial     themes.Definition   `json:"initial"`
		Catalog     []CatalogItem       `json:"catalog"`
		Fonts       []themes.FontOption `json:"fonts"`
		PreviewID   string              `json:"previewID"`
		ContrastID  string              `json:"contrastID,omitempty"`
		PreviewMode string              `json:"previewMode"`
		Tokens      []themes.TokenSpec  `json:"tokens"`
	}{c.ID, c.Initial, c.Catalog, c.Fonts, c.PreviewID, c.ContrastID, c.InitialPreviewMode, themes.Tokens()})
	return string(b)
}
func (c Config) presetOptions() []selectfield.Option {
	return append([]selectfield.Option{{Value: "", Label: "Choose a base"}}, c.PresetOptions()...)
}

// PresetOptions returns theme-colored options for a Select outside or inside this editor.
// Font previews load as options become visible, even without a mounted editor.
func (c Config) PresetOptions() []selectfield.Option {
	if c.Fonts == nil {
		c.Fonts = themes.DefaultFonts()
	}
	out := make([]selectfield.Option, 0, len(c.Catalog))
	for _, p := range c.Catalog {
		ref := c.previewFont(p.Definition.Typography.Title)
		family, err := c.fontFamily(ref)
		if err != nil {
			family = "system-ui,sans-serif"
		}
		attrs := templ.Attributes{"style": "font-family:" + family, "data-theme-font-option": ref.Source + "/" + ref.ID, "data-theme-editor-for": c.ID}
		for _, font := range c.Fonts {
			if font.Ref() == ref {
				data, _ := json.Marshal(font)
				attrs["data-theme-font-preview"] = string(data)
				break
			}
		}
		out = append(out, selectfield.Option{Value: p.ID, Label: p.Label, Leading: themeSwatches(p.Definition), LabelAttrs: attrs})
	}
	return out
}

func colorValue(c themes.Color) string { return c.CSS() }
func radiusValue(c Config) string      { return strconv.FormatFloat(c.Initial.RadiusRem, 'f', -1, 64) }

func (c Config) fontPicker(role string) combobox.Config {
	ref := c.Initial.Typography.Title
	if role == "body" {
		ref = c.Initial.Typography.Body
	}
	options := make([]combobox.Option, 0, len(c.Fonts))
	for _, font := range c.Fonts {
		family, err := c.fontFamily(font.Ref())
		if err != nil {
			continue
		}
		value := font.Source + "/" + font.ID
		options = append(options, combobox.Option{Value: value, Label: font.Label + " · " + font.Weight + " · " + font.Style, LabelAttrs: templ.Attributes{"style": "font-family:" + family, "data-theme-font-option": value, "title": font.Label + " · " + font.Weight + " · " + font.Style}})
	}
	family, _ := c.fontFamily(ref)
	return combobox.Config{ID: c.ID + "-" + role, Name: c.ID + "-" + role + "-font", Label: role + " font", Source: combobox.Source{Static: options}, Selected: []string{ref.Source + "/" + ref.ID}, EnableSearch: true, DisablePersistence: true, Required: true, TriggerAttrs: templ.Attributes{"data-theme-font-trigger": role, "style": "font-family:" + family}}
}
func (c Config) fontFamily(ref themes.FontRef) (string, error) {
	if ref.Source == "google" {
		if _, err := themes.ResolveFont(c.Fonts, ref); err != nil {
			return "", err
		}
		alias := "goshtoso-preview-google-" + strings.NewReplacer(":", "-", "/", "-", ".", "-").Replace(ref.ID)
		return strconv.Quote(alias) + ",system-ui,sans-serif", nil
	}
	_, family, err := themes.RenderFontCSS(ref, func(ref themes.FontRef) (themes.ResolvedFont, error) { return themes.ResolveFont(c.Fonts, ref) })
	return family, err
}
func (c Config) fontFaces() string {
	var css strings.Builder
	for _, font := range c.Fonts {
		declarations, _, err := themes.RenderFontCSS(font.Ref(), func(ref themes.FontRef) (themes.ResolvedFont, error) { return themes.ResolveFont(c.Fonts, ref) })
		if err != nil {
			continue
		}
		for line := range strings.SplitSeq(declarations, "\n") {
			if strings.HasPrefix(line, "@font-face") {
				css.WriteString(line + "\n")
			}
		}
	}
	return css.String()
}

// Named preset stacks may use a curated Google face for isolated option previews.
func (c Config) previewFont(ref themes.FontRef) themes.FontRef {
	if ref.Source != "system" {
		return ref
	}
	resolved, err := themes.ResolveFont(c.Fonts, ref)
	if err != nil || len(resolved.Families) == 0 {
		return ref
	}
	for _, font := range c.Fonts {
		if font.Source == "google" && len(font.Families) > 0 && font.Families[0] == resolved.Families[0] {
			return font.Ref()
		}
	}
	return ref
}

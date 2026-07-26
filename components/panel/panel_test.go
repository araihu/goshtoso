package panel

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestPanelRendersNeutralSlotsWithoutOwningHeadingSemantics(t *testing.T) {
	panel := Panel(Config{
		Appearance:  AppearanceOutlined,
		Density:     DensityCompact,
		Header:      templ.Raw("<h2>Change review</h2>"),
		Actions:     templ.Raw(`<button type="button">Approve</button>`),
		Footer:      templ.Raw("<p>Updated now</p>"),
		RootClass:   "review-panel",
		RootAttrs:   templ.Attributes{"id": "change-review", "aria-labelledby": "change-title"},
		HeaderAttrs: templ.Attributes{"data-region": "header"},
		BodyAttrs:   templ.Attributes{"data-region": "body"},
		FooterAttrs: templ.Attributes{"data-region": "footer"},
	})

	var output strings.Builder
	ctx := templ.WithChildren(context.Background(), templ.Raw("<dl><dt>Risk</dt><dd>High</dd></dl>"))
	if err := panel.Render(ctx, &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	html := output.String()

	for _, want := range []string{
		`id="change-review"`,
		`aria-labelledby="change-title"`,
		`data-region="header"`,
		`data-region="body"`,
		`data-region="footer"`,
		"review-panel",
		"<h2>Change review</h2>",
		`<button type="button">Approve</button>`,
		"<dl><dt>Risk</dt><dd>High</dd></dl>",
		"<p>Updated now</p>",
		"px-4 py-3",
		"border-outline",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %q in:\n%s", want, html)
		}
	}
	for _, unwanted := range []string{"<article", "<section", "<h3", "max-w-"} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("neutral panel unexpectedly owns %q in:\n%s", unwanted, html)
		}
	}
}

func TestPanelAppearancesAndDensityAreBounded(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want []string
		not  []string
	}{
		{name: "default", cfg: Config{}, want: []string{"border-outline", "px-5 py-4"}},
		{name: "subtle relaxed", cfg: Config{Appearance: AppearanceSubtle, Density: DensityRelaxed}, want: []string{"bg-surface-alt", "px-6 py-5"}, not: []string{"border-outline"}},
		{name: "plain", cfg: Config{Appearance: AppearancePlain}, want: []string{"bg-transparent"}, not: []string{"rounded-radius", "border-outline"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output strings.Builder
			if err := Panel(tt.cfg).Render(context.Background(), &output); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			html := output.String()
			for _, want := range tt.want {
				if !strings.Contains(html, want) {
					t.Fatalf("rendered HTML missing %q in:\n%s", want, html)
				}
			}
			for _, unwanted := range tt.not {
				if strings.Contains(html, unwanted) {
					t.Fatalf("rendered HTML unexpectedly contains %q in:\n%s", unwanted, html)
				}
			}
		})
	}
}

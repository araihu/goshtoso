package icon

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
	"github.com/stretchr/testify/require"
)

func render(t *testing.T, component templ.Component) string {
	t.Helper()

	var output bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &output))
	return output.String()
}

func renderErr(cfg Config) error {
	return Icon(cfg).Render(context.Background(), &bytes.Buffer{})
}

func TestIconAccessibilityMatrix(t *testing.T) {
	tests := []struct {
		name       string
		cfg        Config
		contains   []string
		notContain []string
	}{
		{
			name: "labeled icon is an image",
			cfg: Config{
				SpriteURL: "/sprites/ui.svg",
				Symbol:    "check",
				Label:     "Approved",
			},
			contains:   []string{`role="img"`, `aria-label="Approved"`},
			notContain: []string{`aria-hidden="true"`},
		},
		{
			name: "decorative flag wins over label",
			cfg: Config{
				SpriteURL:  "/sprites/ui.svg",
				Symbol:     "check",
				Label:      "Approved",
				Decorative: true,
			},
			contains:   []string{`aria-hidden="true"`},
			notContain: []string{`role="img"`, `aria-label=`},
		},
		{
			name: "blank label is decorative",
			cfg: Config{
				SpriteURL: "/sprites/ui.svg",
				Symbol:    "check",
				Label:     " \t ",
			},
			contains:   []string{`aria-hidden="true"`},
			notContain: []string{`role="img"`, `aria-label=`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			html := render(t, Icon(test.cfg))
			if test.name == "labeled icon is an image" {
				require.Regexp(t, regexp.MustCompile(`<svg[^>]*role="img"[^>]*aria-label="Approved"[^>]*>`), html)
			}
			if test.name == "decorative flag wins over label" {
				require.Regexp(t, regexp.MustCompile(`<svg[^>]*aria-hidden="true"[^>]*>`), html)
			}
			for _, want := range test.contains {
				require.Contains(t, html, want)
			}
			for _, unwanted := range test.notContain {
				require.NotContains(t, html, unwanted)
			}
		})
	}
}

func TestIconRendersSafeExternalAndDocumentLocalReferences(t *testing.T) {
	external := render(t, Icon(Config{
		SpriteURL: "sprites/ui.svg?version=1",
		Symbol:    "check-circle",
		Size:      SizeLG,
		RootClass: "text-primary",
		Label:     "Verified",
	}))
	require.Contains(t, external, `href="sprites/ui.svg?version=1#check-circle"`)
	require.Contains(t, external, `class="size-6 text-primary"`)
	require.NotContains(t, external, `fill=`)
	require.NotContains(t, external, `stroke=`)

	inline := render(t, Icon(Config{
		Symbol: "check-circle",
		Mode:   ModeInline,
	}))
	require.Contains(t, inline, `href="#check-circle"`)
	require.NotContains(t, inline, `sprites/ui.svg`)
}

func TestIconMapsEverySizeToFixedClasses(t *testing.T) {
	tests := []struct {
		size Size
		want string
	}{
		{SizeXS, "size-3"},
		{SizeSM, "size-4"},
		{SizeMD, "size-5"},
		{SizeLG, "size-6"},
		{SizeXL, "size-8"},
		{"", "size-5"},
		{"unknown", "size-5"},
	}

	for _, test := range tests {
		t.Run(string(test.size), func(t *testing.T) {
			html := render(t, Icon(Config{SpriteURL: "/sprites/ui.svg", Symbol: "check", Size: test.size}))
			require.Contains(t, html, `class="`+test.want+`"`)
		})
	}
}

func TestIconRejectsMissingOrUnsafeReferences(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{"missing external sprite URL", Config{Symbol: "check"}},
		{"unsafe scheme", Config{SpriteURL: "javascript:alert(1)", Symbol: "check"}},
		{"data URL", Config{SpriteURL: "data:image/svg+xml,svg", Symbol: "check"}},
		{"protocol relative URL", Config{SpriteURL: "//cdn.example.test/ui.svg", Symbol: "check"}},
		{"document local URL in external mode", Config{SpriteURL: "#existing", Symbol: "check"}},
		{"sprite URL with fragment", Config{SpriteURL: "/sprites/ui.svg#old", Symbol: "check"}},
		{"missing symbol", Config{SpriteURL: "/sprites/ui.svg"}},
		{"uppercase symbol", Config{SpriteURL: "/sprites/ui.svg", Symbol: "Check"}},
		{"leading hyphen", Config{SpriteURL: "/sprites/ui.svg", Symbol: "-check"}},
		{"double hyphen", Config{SpriteURL: "/sprites/ui.svg", Symbol: "check--circle"}},
		{"quote in symbol", Config{SpriteURL: "/sprites/ui.svg", Symbol: `bad\"x`}},
		{"unicode symbol", Config{SpriteURL: "/sprites/ui.svg", Symbol: "café"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Error(t, renderErr(test.cfg))
		})
	}
}

func TestIconAllowsHTTPExternalReferences(t *testing.T) {
	for _, spriteURL := range []string{
		"http://static.example.test/icons.svg",
		"https://static.example.test/icons.svg",
	} {
		t.Run(spriteURL, func(t *testing.T) {
			html := render(t, Icon(Config{SpriteURL: spriteURL, Symbol: "check"}))
			require.Contains(t, html, `href="`+spriteURL+`#check"`)
		})
	}
}

func TestIconExposesComponentIdentity(t *testing.T) {
	require.Equal(t, components.KindIcon, Icon(Config{}).Kind())
}

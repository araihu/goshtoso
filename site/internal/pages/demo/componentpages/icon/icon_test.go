package iconpage

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/icon/heroicons"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/stretchr/testify/require"
)

func renderIconShowcase(t *testing.T) string {
	t.Helper()
	var rendered strings.Builder
	require.NoError(t, iconDemoContent().Render(context.Background(), &rendered))
	return rendered.String()
}

func TestIconShowcaseRendersEveryGlyphInResponsiveGrid(t *testing.T) {
	html := renderIconShowcase(t)

	require.Len(t, heroicons.Glyphs, 67)
	require.Equal(t, len(heroicons.Glyphs), strings.Count(html, `data-icon-card=`))
	require.Contains(t, html, `grid-cols-1 sm:grid-cols-3 xl:grid-cols-6`)
	require.Contains(t, html, "Paste into a")
	require.Contains(t, html, "https://github.com/tailwindlabs/heroicons/blob/master/LICENSE")
	require.Contains(t, html, "IconBrandDeveloperIconsTRPC")
	require.Contains(t, html, "components/icon")
	require.Contains(t, html, "/assets/icons/appicons/sprite.svg")
	require.Contains(t, html, "Bootstrap Icons, generated locally")
	require.Contains(t, html, "/assets/icons/bootstrapicons/sprite.svg")
}

func TestIconShowcaseUsesCanonicalCatalogHeading(t *testing.T) {
	entry, ok := catalog.Lookup("components/icon")
	require.True(t, ok)

	html := renderIconShowcase(t)
	require.Truef(t, strings.Contains(html, `>`+entry.Title+`</h1>`), "icon H1 must match catalog title %q", entry.Title)
}

func TestIconShowcaseParticipatesInComponentDocsContract(t *testing.T) {
	html := renderIconShowcase(t)

	require.Equal(t, 1, strings.Count(html, `data-component-page`))
	require.Equal(t, 1, strings.Count(html, `data-component-description`))
	require.GreaterOrEqual(t, strings.Count(html, `data-component-preview`), 1)
	require.GreaterOrEqual(t, strings.Count(html, `data-component-code`), 1)
}

func TestIconCodeEncoderReflectsMeaningfulSelectedOptions(t *testing.T) {
	got := encodeIconSource(t, iconCodeInput{
		Glyph:     heroicons.Glyphs[0].GoName,
		Size:      "lg",
		Label:     "Search",
		RootClass: "text-accent",
	})

	require.True(t, strings.HasPrefix(got, "@icon.Icon(icon.Config{"))
	require.NotContains(t, got, "package main")
	require.NotContains(t, got, "import (")
	require.NotContains(t, got, "templ IconExample()")
	require.Contains(t, got, "heroicons.SpriteURL")
	require.Contains(t, got, "heroicons."+heroicons.Glyphs[0].GoName)
	require.Contains(t, got, "Size:      icon.SizeLG")
	require.Contains(t, got, `Label:     "Search"`)
	require.Contains(t, got, `RootClass: "text-accent"`)
	require.NotContains(t, got, "Decorative:")
	require.NotContains(t, got, "Mode:")
}

func TestIconCodeEncoderEscapesGoStrings(t *testing.T) {
	source := encodeIconSource(t, iconCodeInput{
		Glyph:      "Icon16SolidMagnifyingGlass",
		Size:       "xl",
		Label:      "Search \"now\"\nnext\\path",
		RootClass:  "before:content-['\\']",
		Decorative: false,
	})

	require.Contains(t, source, `Label:     "Search \"now\"\nnext\\path"`)
	require.Contains(t, source, `RootClass: "before:content-['\\']"`)
	require.NotContains(t, source, `\\u`)
}

func TestIconCodeEncoderMatchesPasteReadyWorkbenchSnippet(t *testing.T) {
	source := encodeIconSource(t, iconCodeInput{
		Glyph:     "Icon16SolidArrowDown",
		Size:      "xl",
		RootClass: "text-danger",
	})

	require.Equal(t, `@icon.Icon(icon.Config{
    SpriteURL: heroicons.SpriteURL,
    Symbol:    heroicons.Icon16SolidArrowDown,
    Size:      icon.SizeXL,
    RootClass: "text-danger",
})
`, source)
}

func TestIconCodeEncoderMakesDecorativeIntentExplicit(t *testing.T) {
	source := encodeIconSource(t, iconCodeInput{
		Glyph:      "Icon16SolidSparkles",
		Size:       "md",
		Label:      "Ignored label",
		Decorative: true,
	})

	require.Contains(t, source, "Symbol:    heroicons.Icon16SolidSparkles")
	require.Contains(t, source, "Decorative: true")
	require.NotContains(t, source, "Label:")
	require.NotContains(t, source, "Size:")
	require.NotContains(t, source, "RootClass:")
}

type iconCodeInput struct {
	Glyph      string `json:"glyph"`
	Size       string `json:"size"`
	Label      string `json:"label"`
	Decorative bool   `json:"decorative"`
	RootClass  string `json:"rootClass"`
}

func encodeIconSource(t *testing.T, input iconCodeInput) string {
	t.Helper()
	payload, err := json.Marshal(input)
	require.NoError(t, err)

	root := iconRepoRoot(t)
	sourcePath := filepath.Join(root, "site", "assets", "js", "src", "icon-catalog.js")
	script := `
const fs = require("fs");
global.window = {};
global.document = { addEventListener: function () {} };
eval(fs.readFileSync(process.argv[1], "utf8"));
const input = JSON.parse(process.argv[2]);
if (typeof window.goshtosoIconCode !== "function") {
  throw new Error("icon source encoder is not exported");
}
process.stdout.write(window.goshtosoIconCode(input));
`
	command := exec.Command("node", "-e", script, sourcePath, string(payload))
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "node icon encoder: %s", output)
	return string(output)
}

func iconRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "../../../../../../"))
}

func TestIconShowcaseUsesModalSemanticsAndAccessibleControls(t *testing.T) {
	html := renderIconShowcase(t)

	require.Contains(t, html, `role="dialog"`)
	require.Contains(t, html, `aria-modal="true"`)
	require.Contains(t, html, "x-trap.inert.noscroll")
	require.Contains(t, html, "keydown.esc.window")
	for _, control := range []string{"Size", "Label", "Decorative", "Color", "Paste into a .templ file", "Copy Paste into a .templ file code"} {
		require.Contains(t, html, control)
	}
	require.Contains(t, html, `id="icon-workbench-code"`)
	require.Contains(t, html, `data-testid="icon-size-selector"`)
	require.Contains(t, html, `role="combobox"`)
	require.Contains(t, html, `id="icon-decorative"`)
	require.NotContains(t, html, `<select`)
	require.NotContains(t, html, `Inherit currentColor`)
}

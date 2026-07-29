package components

import (
	"context"
	"encoding/json"
	"os"
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
	require.Contains(t, html, "copy a standalone Go program")
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

	require.Contains(t, got, "package main")
	require.Contains(t, got, `"context"`)
	require.Contains(t, got, `"os"`)
	require.Contains(t, got, "heroicons.SpriteURL")
	require.Contains(t, got, "heroicons."+heroicons.Glyphs[0].GoName)
	require.Contains(t, got, "Size:      icon.SizeLG")
	require.Contains(t, got, `Label:     "Search"`)
	require.Contains(t, got, `RootClass: "text-accent"`)
	require.NotContains(t, got, "Decorative:")
	require.NotContains(t, got, "Mode:")
}

func TestIconCodeEncoderEscapesGoStringsAndCompiles(t *testing.T) {
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
	compileIconSource(t, source)
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

func compileIconSource(t *testing.T, source string) {
	t.Helper()
	root := iconRepoRoot(t)
	directory, err := os.MkdirTemp(root, "icon-catalog-example-*")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(directory)) })
	require.NoError(t, os.WriteFile(filepath.Join(directory, "main.go"), []byte(source), 0o600))

	command := exec.Command("go", "run", "./"+filepath.Base(directory))
	command.Dir = root
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "generated icon example did not compile: %s\n%s", err, output)
	require.Contains(t, string(output), `href="/assets/icons/heroicons.svg#hi-16-solid-magnifying-glass"`)
}

func iconRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "../../../../../"))
}

func TestIconShowcaseUsesModalSemanticsAndAccessibleControls(t *testing.T) {
	html := renderIconShowcase(t)

	require.Contains(t, html, `role="dialog"`)
	require.Contains(t, html, `aria-modal="true"`)
	require.Contains(t, html, "x-trap.inert.noscroll")
	require.Contains(t, html, "keydown.esc.window")
	for _, control := range []string{"Size", "Label", "Decorative", "Root class", "Copy compilable Go"} {
		require.Contains(t, html, control)
	}
	require.Contains(t, html, `aria-live="polite"`)
}

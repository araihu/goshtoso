//go:build e2e && full

package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/themes"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func cssImportFile(t *testing.T, page playwright.Page, name string, css []byte) {
	t.Helper()
	require.NoError(t, page.Locator("[data-theme-import-file]").SetInputFiles(playwright.InputFile{Name: name, MimeType: "text/css", Buffer: css}))
}
func waitForCSSImport(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.WaitForFunction(`()=>document.querySelector('[data-theme-import-feedback]').textContent==='CSS imported into the draft.'&&!document.querySelector('[data-theme-import-feedback]').hidden`, nil)
	require.NoError(t, err)
}
func TestThemePage_CSSImportAllPresets(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	count, err := page.Locator("[data-theme-import]").Count()
	require.NoError(t, err)
	require.Zero(t, count)
	require.NoError(t, page.Locator("[data-theme-import-file]").DispatchEvent("change", nil))
	visible, err := page.Locator("[data-theme-import-feedback]").IsVisible()
	require.NoError(t, err)
	require.False(t, visible)
	for _, preset := range themes.BuiltIn() {
		t.Run(string(preset.Key), func(t *testing.T) {
			d, err := themes.Preset(preset.Key)
			require.NoError(t, err)
			css, err := themes.RenderCSS(d, nil)
			require.NoError(t, err)
			_, err = page.Evaluate(`()=>{const root=document.querySelector('[data-theme-editor]'),config=JSON.parse(root.dataset.themeEditor);root.dispatchEvent(new CustomEvent('themeeditor:load',{detail:{id:config.id,definition:config.initial}}))}`, nil)
			require.NoError(t, err)
			cssImportFile(t, page, "theme.css", []byte(css))
			waitForCSSImport(t, page)
			imported := editorDefinition(t, page)
			require.Equal(t, d.Tokens, imported.Tokens)
			require.Equal(t, d.Typography, imported.Typography)
			require.Equal(t, d.RadiusRem, imported.RadiusRem)
			require.Empty(t, imported.Provenance.PresetKey)
		})
	}
}
func TestThemePage_CSSImportExportConfirmationAndReset(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	baseline := editorDefinition(t, page)
	for _, source := range []string{"system", "google", "bundled"} {
		t.Run(source, func(t *testing.T) {
			_, err := page.Evaluate(`source=>{
    const root=document.querySelector('[data-theme-editor]'),config=JSON.parse(root.dataset.themeEditor),d=structuredClone(config.initial),font=config.fonts.find(f=>f.source===source);
    d.typography.title={source:font.source,id:font.id};d.radiusRem=1.25;
    root.dispatchEvent(new CustomEvent('themeeditor:load',{detail:{id:config.id,definition:d}}));
   }`, source)
			require.NoError(t, err)
			exported := editorDefinition(t, page)
			download, err := page.ExpectDownload(func() error { return page.Locator("[data-theme-export]").Click() })
			require.NoError(t, err)
			path, err := download.Path()
			require.NoError(t, err)
			css, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Contains(t, string(css), "Goshtoso theme fonts:")
			_, err = page.Evaluate(`()=>{const root=document.querySelector('[data-theme-editor]'),c=JSON.parse(root.dataset.themeEditor);root.dispatchEvent(new CustomEvent('themeeditor:load',{detail:{id:c.id,definition:c.initial}}))}`, nil)
			require.NoError(t, err)
			require.NoError(t, page.Locator("[data-theme-radius]").Fill("2"))
			require.NoError(t, page.Locator("[data-theme-radius]").DispatchEvent("change", nil))
			before := editorDefinition(t, page)
			cssImportFile(t, page, "theme.css", css)
			require.NoError(t, page.Locator("#site-theme-editor-confirm").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
			require.NoError(t, page.Locator("[data-theme-confirm-cancel]").Click())
			require.Equal(t, before, editorDefinition(t, page))
			_, err = page.WaitForFunction(`()=>document.querySelector('[data-theme-import-file]').files.length===0&&!document.querySelector('[data-theme-import-file]').disabled`, nil)
			require.NoError(t, err)
			cssImportFile(t, page, "theme.css", css)
			require.NoError(t, page.Locator("[data-theme-confirm-accept]").Click())
			waitForCSSImport(t, page)
			imported := editorDefinition(t, page)
			require.Equal(t, exported.Tokens, imported.Tokens)
			require.Equal(t, exported.Typography, imported.Typography)
			require.Equal(t, exported.RadiusRem, imported.RadiusRem)
			require.NoError(t, page.Locator("[data-theme-reset]").Click())
			require.NoError(t, page.Locator("[data-theme-confirm-accept]").Click())
			_, err = page.WaitForFunction(`([radius,id,key])=>{
             const d=JSON.parse(document.querySelector('[data-theme-definition]').value);
             return d.radiusRem===radius&&d.typography.title.id===id&&d.provenance.presetKey===key;
            }`, []any{baseline.RadiusRem, baseline.Typography.Title.ID, string(baseline.Provenance.PresetKey)})
			require.NoError(t, err)
			require.Equal(t, baseline, editorDefinition(t, page))
		})
	}
}
func TestThemePage_CSSImportErrorsPreserveDraftAndStayDetached(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	initial := editorDefinition(t, page)
	good, err := themes.RenderCSS(initial, nil)
	require.NoError(t, err)
	tests := []struct{ name, css string }{
		{"malformed.css", "not a theme"},
		{"empty.css", ""},
		{"wrong.txt", good},
		{"large.css", strings.Repeat("x", 1048577)},
		{"missing.css", strings.Replace(good, "--color-primary:", "--missing:", 1)},
		{"cycle.css", strings.Replace(good, "--color-primary:", "--loop:var(--color-primary);--color-primary:var(--loop);--unused:", 1)},
		{"font.css", strings.Replace(good, "--font-title:", "--font-title:UnknownFont;--unused-font:", 1)},
		{"radius.css", strings.Replace(good, "--radius-radius:", "--radius-radius:5rem;--unused-radius:", 1)},
		{"multiple.css", good + strings.Replace(good, `data-theme="custom"`, `data-theme="another"`, -1)},
	}
	for _, sample := range tests {
		t.Run(sample.name, func(t *testing.T) {
			cssImportFile(t, page, sample.name, []byte(sample.css))
			_, err := page.WaitForFunction(`()=>document.querySelector('[data-theme-import-file]').getAttribute('aria-invalid')==='true'`, nil)
			require.NoError(t, err)
			require.Equal(t, initial, editorDefinition(t, page))
			_, err = page.WaitForFunction(`()=>document.querySelector('[data-theme-import-file]').files.length===0&&!document.querySelector('[data-theme-import-file]').disabled`, nil)
			require.NoError(t, err)
		})
	}
	var unexpected []string
	page.OnRequest(func(r playwright.Request) {
		if strings.Contains(r.URL(), "untrusted.invalid") {
			unexpected = append(unexpected, r.URL())
		}
	})
	unsafe := `@import url("https://untrusted.invalid/style.css");@font-face{font-family:Unexpected;src:url("https://untrusted.invalid/font.woff2")}body{background-image:url("https://untrusted.invalid/image");color:red}` + good
	before, err := page.Locator("body").Evaluate(`el=>getComputedStyle(el).color`, nil)
	require.NoError(t, err)
	cssImportFile(t, page, "inert.css", []byte(unsafe))
	waitForCSSImport(t, page)
	after, err := page.Locator("body").Evaluate(`el=>getComputedStyle(el).color`, nil)
	require.NoError(t, err)
	require.Equal(t, before, after)
	require.Empty(t, unexpected)
}

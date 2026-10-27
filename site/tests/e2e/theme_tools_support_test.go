//go:build e2e && full

package e2e

import (
	"encoding/json"
	"github.com/araihu/goshtoso/themes"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func waitForThemeEditor(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.WaitForFunction(`() => !!document.querySelector('[data-theme-editor] [data-theme-definition]')?.value`, nil)
	require.NoError(t, err)
}

func gotoThemePage(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Goto(baseURL+"/docs/theme", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	waitForThemeEditor(t, page)
}

func choosePreviewPreset(t *testing.T, page playwright.Page, key string) {
	t.Helper()
	_, err := page.Evaluate(`()=>{window.themePresetUpdated=false;document.querySelector('[data-theme-editor]').addEventListener('themeeditor:change',()=>{window.themePresetUpdated=true},{once:true});}`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#site-theme-editor-preset-trigger").Click())
	label, err := page.Evaluate(`key => JSON.parse(document.querySelector("[data-theme-editor]").dataset.themeEditor).catalog.find(p=>p.id===key).label`, key)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#site-theme-editor-preset-listbox").GetByRole("option", playwright.LocatorGetByRoleOptions{Name: label.(string), Exact: new(true)}).Click())
	_, err = page.WaitForFunction(`()=>window.themePresetUpdated||document.querySelector('#site-theme-editor-confirm').open`, nil)
	require.NoError(t, err)

	if visible, e := page.Locator("#site-theme-editor-confirm [data-theme-confirm-accept]").IsVisible(); e == nil && visible {
		require.NoError(t, page.Locator("#site-theme-editor-confirm [data-theme-confirm-accept]").Click())
	}
	_, err = page.WaitForFunction(`key => JSON.parse(document.querySelector('[data-theme-definition]').value).provenance.presetKey===key`, key)
	require.NoError(t, err)
}

func editorDefinition(t *testing.T, page playwright.Page) themes.Definition {
	t.Helper()
	value, err := page.Locator("[data-theme-definition]").InputValue()
	require.NoError(t, err)
	var definition themes.Definition
	require.NoError(t, json.Unmarshal([]byte(value), &definition))
	return definition
}

// Documentation navigation checks use real bundled font bytes without external network timing.
func useLocalGoogleFontFixture(t *testing.T, page playwright.Page) {
	t.Helper()
	data, err := os.ReadFile("../../../assets/fonts/go-regular.197d9f3703b4.ttf")
	require.NoError(t, err)
	require.NoError(t, page.Route("https://fonts.googleapis.com/**", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{Status: new(200), ContentType: new("text/css"), Body: `@font-face{font-family:Fixture;src:url(https://fonts.gstatic.com/goshtoso-e2e-fixture.ttf) format('truetype');font-weight:400;font-style:normal;}`}))
	}))
	require.NoError(t, page.Route("https://fonts.gstatic.com/**", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{Status: new(200), ContentType: new("font/ttf"), Body: data, Headers: map[string]string{"Access-Control-Allow-Origin": "*"}}))
	}))
}

//go:build e2e && full

package e2e

import (
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, page.Locator("#site-theme-editor-preset-trigger").Click())
	label, err := page.Evaluate(`key => JSON.parse(document.querySelector("[data-theme-editor]").dataset.themeEditor).catalog.find(p=>p.id===key).label`, key)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#site-theme-editor-preset-listbox").GetByRole("option", playwright.LocatorGetByRoleOptions{Name: label.(string), Exact: new(true)}).Click())
	if visible, e := page.Locator("#site-theme-editor-confirm [data-theme-confirm-accept]").IsVisible(); e == nil && visible {
		require.NoError(t, page.Locator("#site-theme-editor-confirm [data-theme-confirm-accept]").Click())
	}
	_, err = page.WaitForFunction(`key => JSON.parse(document.querySelector('[data-theme-definition]').value).provenance.presetKey===key`, key)
	require.NoError(t, err)
}

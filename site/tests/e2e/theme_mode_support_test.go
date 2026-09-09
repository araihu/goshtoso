//go:build e2e && full

package e2e

import (
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func setThemeMode(t *testing.T, page playwright.Page, theme string, dark bool) {
	t.Helper()
	_, err := page.Evaluate(`([theme, dark]) => {
		const html = document.documentElement;
		let style = document.getElementById('contrast-test-no-transitions');
		if (!style) {
			style = document.createElement('style');
			style.id = 'contrast-test-no-transitions';
			style.textContent = '* { transition: none !important; }';
			document.head.append(style);
		}
		localStorage.setItem('theme', theme);
		html.setAttribute('data-theme', theme);
		html.classList.toggle('dark', dark);
	}`, []any{theme, dark})
	require.NoError(t, err)
	_, err = page.WaitForFunction("theme => document.documentElement.dataset.theme === theme", theme)
	require.NoError(t, err)
}

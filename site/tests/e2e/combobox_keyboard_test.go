//go:build e2e && (full || combobox)

package e2e

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestComboboxKeyboardSelectClearAndSkipDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()
	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/combobox", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	wait := func(expression string) {
		t.Helper()
		_, err := page.WaitForFunction(expression, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
		require.NoError(t, err)
	}
	require.NoError(t, page.Locator("#skills-trigger").Press("Enter"))
	wait(`() => document.activeElement?.getAttribute('data-value') === 'go'`)
	require.NoError(t, page.Keyboard().Press("Space"))
	wait(`() => document.querySelector('#skills input[type=hidden][value=go]') !== null`)
	// Supporting option content must never leak into the closed selection label.
	label, err := page.Locator("#skills-trigger-label").TextContent()
	require.NoError(t, err)
	require.Equal(t, "Go", strings.TrimSpace(label))
	trailing, err := page.Locator(`#skills [data-value="go"] .ml-auto`).TextContent()
	require.NoError(t, err)
	require.Equal(t, "12", strings.TrimSpace(trailing))
	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	wait(`() => document.activeElement?.getAttribute('data-value') === 'rust'`)
	require.NoError(t, page.Keyboard().Press("Enter"))
	wait(`() => document.querySelectorAll('#skills input[type=hidden][name=skills]').length === 2`)
	// Options are outside the sequential tab order; Shift+Tab reaches Clear all.
	require.NoError(t, page.Keyboard().Press("Shift+Tab"))
	wait(`() => document.activeElement?.hasAttribute('data-combobox-clear')`)
	require.NoError(t, page.Keyboard().Press("Enter"))
	wait(`() => document.querySelectorAll('#skills input[type=hidden][name=skills]').length === 0 && document.activeElement?.id === 'skills-trigger'`)
	require.NoError(t, page.Keyboard().Press("Escape"))
	wait(`() => document.querySelector('#skills-trigger').getAttribute('aria-expanded') === 'false' && document.activeElement?.id === 'skills-trigger'`)
	// Region's final option is disabled. End and wrapping arrows must skip it.
	require.NoError(t, page.Locator("#region-trigger").Press("End"))
	wait(`() => document.activeElement?.getAttribute('data-value') === 'ap-south-1'`)
	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	wait(`() => document.activeElement?.getAttribute('data-value') === 'us-east-1'`)
	require.NoError(t, page.Keyboard().Press("ArrowUp"))
	wait(`() => document.activeElement?.getAttribute('data-value') === 'ap-south-1'`)
	require.NoError(t, page.Keyboard().Press("Enter"))
	wait(`() => document.querySelector('#region input[type=hidden]').value === 'ap-south-1' && document.activeElement?.id === 'region-trigger'`)
	// Server clear removes its button during a full-root morph.
	require.NoError(t, page.Locator("#users-trigger").Click())
	require.NoError(t, page.Locator("#users [data-combobox-search]").Fill("al"))
	wait(`() => document.querySelector('#users [data-value=alice]') !== null`)
	require.NoError(t, page.Locator("#users [data-combobox-search]").Press("ArrowDown"))
	wait(`() => document.activeElement?.getAttribute('data-value') === 'alice'`)
	require.NoError(t, page.Keyboard().Press("Enter"))
	wait(`() => document.querySelector('#users input[type=hidden][value=alice]') !== null`)
	require.NoError(t, page.Locator("#users [data-combobox-clear-all]").Press("Enter"))
	wait(`() => document.querySelectorAll('#users input[type=hidden]').length === 0 && document.activeElement?.id === 'users-trigger'`)

	// A deferred open or Escape must not follow focus into another control.
	for _, key := range []string{"ArrowUp", "Escape"} {
		if key == "Escape" {
			require.NoError(t, page.Locator("#region-trigger").Press("Enter"))
			wait(`() => document.activeElement?.getAttribute('data-value') === 'us-east-1'`)
		}
		_, err := page.Evaluate(`async key => {
			const trigger = document.querySelector('#region-trigger');
			if (key !== 'Escape') trigger.focus();
			document.activeElement.dispatchEvent(new KeyboardEvent('keydown', {key, bubbles:true}));
			document.querySelector('#industry-trigger').focus();
			for (let i=0; i<6; i++) await new Promise(requestAnimationFrame);
		}`, key)
		require.NoError(t, err)
		wait(`() => document.activeElement?.id === 'industry-trigger' && document.querySelector('#region-trigger').getAttribute('aria-expanded') === 'false'`)
	}

}

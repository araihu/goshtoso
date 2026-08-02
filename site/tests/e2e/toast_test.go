//go:build e2e && (full || toast)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestToastDangerTriggerCreatesToast(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/toast", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("#toast-alpine button", playwright.PageLocatorOptions{HasText: "Danger"}).Click())
	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('#toast-container [role="alert"]')).some(el => el.textContent.includes('Oops!'))`,
		nil,
	)
	require.NoError(t, err)
	require.Empty(t, jsErrors, "no JS console/page errors when creating the danger toast: %v", jsErrors)
}

func TestToastServerToastsHaveStackGap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/toast", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#toast-htmx button", playwright.PageLocatorOptions{HasText: "Server Success Toast"}).Click())
	require.NoError(t, page.Locator("#toast-htmx button", playwright.PageLocatorOptions{HasText: "Server Danger Toast"}).Click())
	_, err = page.WaitForFunction(
		`() => document.querySelectorAll('#toast-container-oob [role="alert"]').length >= 2`,
		nil,
	)
	require.NoError(t, err)
	_, err = page.WaitForFunction(
		`() => {
			const style = getComputedStyle(document.querySelector('#toast-container-oob'));
			return style.display === 'flex' && style.flexDirection === 'column' && parseFloat(style.rowGap) >= 8;
		}`,
		nil,
	)
	require.NoError(t, err)
}

func TestToastStaticExamplesStayVisible(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, _ := setupPlaywright(t)
	context, err := browser.NewContext()
	require.NoError(t, err)
	t.Cleanup(func() { _ = context.Close() })
	page, err := context.NewPage()
	require.NoError(t, err)
	page.SetDefaultTimeout(2000)
	page.SetDefaultNavigationTimeout(3000)
	require.NoError(t, page.Clock().Install(playwright.ClockInstallOptions{Time: 0}))
	_, err = page.Goto(baseURL+"/components/toast", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	staticAlerts := page.Locator("#toast-static [role='alert']")
	count, err := staticAlerts.Count()
	require.NoError(t, err)
	require.Equal(t, 5, count, "static docs preview should render all toast primitives")

	require.NoError(t, page.Clock().RunFor(8500))

	count, err = staticAlerts.Count()
	require.NoError(t, err)
	require.Equal(t, 5, count, "static docs preview toasts should not auto-dismiss")
}

//go:build e2e && (full || toast)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToastCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	var jsErrors []string
	page.On("pageerror", func(err error) {
		jsErrors = append(jsErrors, err.Error())
	})
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/toast", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{
		Name:  "Toast",
		Exact: new(true),
	}).WaitFor())
	for _, selector := range []string{"#toast-fragment", "#toast-container", "#toast-alpine", "#toast-htmx", "#toast-static"} {
		require.NoErrorf(t, page.Locator(selector).WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateAttached,
		}), "%s should render", selector)
	}

	require.NoError(t, page.Locator("#toast-alpine button", playwright.PageLocatorOptions{HasText: "Message"}).Click())
	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('#toast-container [role="alert"]')).some(el =>
			el.textContent.includes('Jack Ellis') &&
			el.textContent.includes('Hey, can you review the PR I just submitted?') &&
			el.textContent.includes('Dismiss'))`,
		nil,
	)
	require.NoError(t, err)

	require.NoError(t, page.Locator("#toast-alpine button", playwright.PageLocatorOptions{HasText: "Success"}).Click())
	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('#toast-container [role="alert"]')).some(el =>
			el.textContent.includes('Success!') &&
			el.textContent.includes('Your changes have been saved.'))`,
		nil,
	)
	require.NoError(t, err)

	staticAlerts := page.Locator("#toast-static [role='alert']")
	count, err := staticAlerts.Count()
	require.NoError(t, err)
	assert.Equal(t, 5, count, "static preview should render every documented toast primitive")
	require.NoError(t, page.Locator("#toast-static").GetByText("Jack Ellis", playwright.LocatorGetByTextOptions{
		Exact: new(true),
	}).WaitFor())

	clickUntil(t, page, page.Locator("#toast-htmx button", playwright.PageLocatorOptions{HasText: "Server Info Toast"}),
		`() => Array.from(document.querySelectorAll('#toast-container-oob [role="alert"]')).some(el =>
			el.textContent.includes('Server Info') &&
			el.textContent.includes('This is an informational toast from the server.'))`)
	oobCount, err := page.Locator("#toast-container-oob [role='alert']").Count()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, oobCount, 1)

	require.Empty(t, jsErrors, "no JS console/page errors on toast demo: %v", jsErrors)
}

func TestToastReducedMotionShowsAndDismissesWithoutVisualTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	page := newIsolatedPage(t)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}))
	_, err := page.Goto(baseURL+"/components/toast", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	require.NoError(t, page.Locator("#toast-alpine button", playwright.PageLocatorOptions{HasText: "Success"}).Click())
	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('#toast-container [role="alert"]')).some(el => el.textContent.includes('Success!'))`,
		nil,
	)
	require.NoError(t, err)
	_, err = page.WaitForFunction(
		`() => {
			const el = Array.from(document.querySelectorAll('#toast-container [role="alert"]')).find(el => el.textContent.includes('Success!'));
			if (!el) return false;
			const style = getComputedStyle(el);
			return style.transitionProperty === 'none' && style.opacity === '1' && style.transform === 'none';
		}`,
		nil,
	)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#toast-container [role='alert']").GetByRole("button", playwright.LocatorGetByRoleOptions{
		Name: "dismiss notification",
	}).Click())
	_, err = page.WaitForFunction(
		`() => !Array.from(document.querySelectorAll('#toast-container [role="alert"]')).some(el => el.textContent.includes('Success!'))`,
		nil,
	)
	require.NoError(t, err)

	require.NoError(t, page.Locator("#toast-htmx button", playwright.PageLocatorOptions{HasText: "Server Success Toast"}).Click())
	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('#toast-container-oob [role="alert"]')).some(el => el.textContent.includes('Server Says Hello!'))`,
		nil,
	)
	require.NoError(t, err)
	_, err = page.WaitForFunction(
		`() => {
			const el = Array.from(document.querySelectorAll('#toast-container-oob [role="alert"]')).find(el => el.textContent.includes('Server Says Hello!'));
			if (!el) return false;
			const style = getComputedStyle(el);
			return style.transitionProperty === 'none' && style.opacity === '1' && style.transform === 'none';
		}`,
		nil,
	)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#toast-container-oob [role='alert']").GetByRole("button", playwright.LocatorGetByRoleOptions{
		Name: "dismiss notification",
	}).Click())
	_, err = page.WaitForFunction(
		`() => !Array.from(document.querySelectorAll('#toast-container-oob [role="alert"]')).some(el => el.textContent.includes('Server Says Hello!'))`,
		nil,
	)
	require.NoError(t, err)
}

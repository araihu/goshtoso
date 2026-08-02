package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModal_DefaultDialogOpensAndDismisses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/modal", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	require.NoError(t, page.Locator("#modal-default button").Filter(playwright.LocatorFilterOptions{HasText: "Open Modal"}).Click())
	dialog := page.Locator("#modal-default [role='dialog']")
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	labelledBy, err := dialog.GetAttribute("aria-labelledby")
	require.NoError(t, err)
	assert.NotEmpty(t, labelledBy)
	assert.Equal(t, "Special Offer", mustText(t, page.Locator("#"+labelledBy)))

	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(3000),
	}))
}

func TestModal_HTMXPrimaryActionClosesDialogAndSwapsResult(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/modal", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	require.NoError(t, page.Locator("#modal-htmx button").Filter(playwright.LocatorFilterOptions{HasText: "Open HTMX Modal"}).Click())
	dialog := page.Locator("#modal-htmx [role='dialog']")
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))

	require.NoError(t, dialog.Locator("button").Filter(playwright.LocatorFilterOptions{HasText: "Confirm"}).Click())
	require.NoError(t, page.Locator("#modal-htmx-result").Filter(playwright.LocatorFilterOptions{
		HasText: "Hello from HTMX! Request received at POST /api/hello",
	}).WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(3000)}))
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(3000),
	}))
}

func TestModal_JSActionRunsAfterConfirm(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/modal", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	alertSeen := make(chan string, 1)
	page.On("dialog", func(dialog playwright.Dialog) {
		alertSeen <- dialog.Message()
		require.NoError(t, dialog.Accept())
	})

	require.NoError(t, page.Locator("#modal-js button").Filter(playwright.LocatorFilterOptions{HasText: "Open JS Modal"}).Click())
	require.NoError(t, page.Locator("#modal-js [role='dialog'] button").Filter(playwright.LocatorFilterOptions{HasText: "Delete"}).Click())

	select {
	case msg := <-alertSeen:
		assert.Equal(t, "Item deleted successfully!", msg)
	default:
		t.Fatal("expected JS modal primary action to show alert")
	}
}

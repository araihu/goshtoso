package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCombobox_Toggle_PreservesIsOpen asserts that clicking an option inside
// a multi-select combobox triggers an HTMX swap of the body but the Alpine
// outer shell's isOpen state (driving aria-expanded) survives the swap.
func TestCombobox_Toggle_PreservesIsOpen(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/combobox-new", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	trigger := page.Locator("#skills-trigger").First()
	require.NoError(t, trigger.Click())

	// Scope the click to the skills combobox so we don't accidentally click an option
	// from a sibling combobox in the DOM.
	opt := page.Locator("#skills [role=option]").First()
	require.NoError(t, opt.Click())

	// After HTMX swap of the skills body, the Alpine isOpen state on the outer shell
	// must persist — aria-expanded should still be "true".
	expanded, err := trigger.GetAttribute("aria-expanded")
	require.NoError(t, err)
	assert.Equal(t, "true", expanded, "isOpen must survive HTMX swap of body")
}

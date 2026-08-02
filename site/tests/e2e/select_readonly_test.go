//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelect_ReadonlyDisabledWithHiddenInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/select", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	// The visible trigger should be disabled
	trigger := page.Locator("#os-readonly-trigger")
	require.NoError(t, trigger.WaitFor())

	disabled, err := trigger.IsDisabled()
	require.NoError(t, err)
	assert.True(t, disabled, "readonly select should be rendered as disabled")

	// Should show the selected value "Windows"
	val, err := trigger.InnerText()
	require.NoError(t, err)
	assert.Contains(t, val, "Windows", "readonly select should show the selected label")

	// There should be a hidden input with the same name and value for form submission
	hidden := page.Locator("form#readonlySelectForm input[hidden][name='os-readonly']")
	count, err := hidden.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should have a hidden input for form submission")

	hiddenVal, err := hidden.InputValue()
	require.NoError(t, err)
	assert.Equal(t, "windows", hiddenVal, "hidden input should have the selected value")
}

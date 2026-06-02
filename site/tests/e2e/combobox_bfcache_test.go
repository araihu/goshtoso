package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCombobox_BFCache_RestoresSelection asserts that selecting options,
// navigating away, then pressing the back button restores the combobox
// with options rendered and selection intact.
//
// Against the current Alpine combobox this test is expected to FAIL
// because x-data re-initialization on bfcache restore wipes the options
// array. Against the HTMX SSR rewrite it must PASS.
func TestCombobox_BFCache_RestoresSelection(t *testing.T) {
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

	trigger := page.Locator("#industry-trigger").First()
	require.NoError(t, trigger.Click())

	firstOption := page.Locator("[role=option]").First()
	require.NoError(t, firstOption.Click())

	labelBefore, err := trigger.InnerText()
	require.NoError(t, err)
	require.NotEmpty(t, labelBefore, "option must be selected before navigating away")

	_, err = page.Goto(baseURL+"/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	_, err = page.GoBack(playwright.PageGoBackOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	triggerAfter := page.Locator("#industry-trigger").First()
	labelAfter, err := triggerAfter.InnerText()
	require.NoError(t, err)
	assert.Equal(t, labelBefore, labelAfter, "trigger label must match before/after back navigation")

	require.NoError(t, triggerAfter.Click())
	optionCount, err := page.Locator("[role=option]").Count()
	require.NoError(t, err)
	assert.Greater(t, optionCount, 0, "dropdown must have options after bfcache restore")
}

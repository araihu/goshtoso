package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKbd_PageLoads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/kbd", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	title, err := page.Title()
	require.NoError(t, err)
	assert.Contains(t, title, "KBD")

	require.NoError(t, page.Locator("#kbd-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	keys, err := page.Locator("#kbd-frequently-used kbd").Count()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, keys, 16)

	icons, err := page.Locator("#kbd-icons kbd[aria-label]").Count()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, icons, 9)

	for _, id := range []string{"#kbd-inline", "#kbd-alphabet", "#kbd-numbers", "#kbd-functions", "#kbd-sizes"} {
		count, err := page.Locator(id).Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "%s should exist once", id)
	}
}

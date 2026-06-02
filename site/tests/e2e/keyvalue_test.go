package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyValue_AddAndRemoveRows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/key-value", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	container := page.Locator("#labelsDemo")
	require.NoError(t, container.WaitFor())

	// Should have 2 initial rows (app=web, env=prod)
	rows := container.Locator("template + div") // Alpine x-for rows
	// Count hidden inputs as proxy for row count
	hiddenInputs := container.Locator("input[type='hidden']")
	count, err := hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 2, count, "should have 2 initial key-value rows")

	// Add a row
	addBtn := container.Locator("[data-add-row]")
	err = addBtn.Click()
	require.NoError(t, err)

	count, err = hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 3, count, "should have 3 rows after adding")

	// Remove first row
	removeBtn := container.Locator("button[aria-label='Remove row']").First()
	err = removeBtn.Click()
	require.NoError(t, err)

	count, err = hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 2, count, "should have 2 rows after removing")

	_ = rows
}

func TestKeyValue_EmptyStartAndFillRow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/key-value", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	container := page.Locator("#envVarsDemo")
	require.NoError(t, container.WaitFor())

	// Should start empty
	hiddenInputs := container.Locator("input[type='hidden']")
	count, err := hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 0, count, "should start with 0 rows")

	// Add a row
	addBtn := container.Locator("[data-add-row]")
	err = addBtn.Click()
	require.NoError(t, err)

	count, err = hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should have 1 row after clicking add")
}

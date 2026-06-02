package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTriplet_AddAndRemoveRows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/triplet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	container := page.Locator("#taintsDemo")
	require.NoError(t, container.WaitFor())

	// Should have 1 initial row (the control-plane taint)
	hiddenInputs := container.Locator("input[type='hidden']")
	count, err := hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should have 1 initial triplet row")

	// Should have a select with effect options
	selects := container.Locator("select")
	selectCount, err := selects.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, selectCount, "should have 1 effect dropdown")

	// Add a row
	addBtn := container.Locator("[data-add-row]")
	err = addBtn.Click()
	require.NoError(t, err)

	count, err = hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 2, count, "should have 2 rows after adding")

	// New row should have default effect in select
	selectCount, err = selects.Count()
	require.NoError(t, err)
	assert.Equal(t, 2, selectCount, "should have 2 effect dropdowns")

	// Remove first row
	removeBtn := container.Locator("button[aria-label='Remove row']").First()
	err = removeBtn.Click()
	require.NoError(t, err)

	count, err = hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should have 1 row after removing")
}

func TestTriplet_EmptyStartWithCustomEffects(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/triplet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	container := page.Locator("#priorityDemo")
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

	// Should now have 1 row with a select containing custom effects (high/medium/low)
	selects := container.Locator("select")
	selectCount, err := selects.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, selectCount, "should have 1 effect dropdown after adding")

	// Verify the effect dropdown has the expected options
	options := selects.First().Locator("option")
	optCount, err := options.Count()
	require.NoError(t, err)
	assert.Equal(t, 3, optCount, "should have 3 effect options (high, medium, low)")
}

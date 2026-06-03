package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructuredInput_AddAndRemoveKeyValueRows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/structured-input", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	container := page.Locator("#labelsDemo")
	require.NoError(t, container.WaitFor())

	hiddenInputs := container.Locator("input[type='hidden']")
	count, err := hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 4, count, "two rows with two columns should render four hidden inputs")

	name, err := hiddenInputs.First().GetAttribute("name")
	require.NoError(t, err)
	assert.Equal(t, "labels[0][key]", name)

	addBtn := container.Locator("[data-add-row]")
	require.NoError(t, addBtn.Click())
	count, err = hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 6, count, "adding one two-column row should add two hidden inputs")

	removeBtn := container.Locator("button[aria-label='Remove row']").First()
	require.NoError(t, removeBtn.Click())
	count, err = hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 4, count, "removing one two-column row should remove two hidden inputs")
}

func TestStructuredInput_SelectColumnDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/structured-input", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	container := page.Locator("#rulesDemo")
	require.NoError(t, container.WaitFor())

	hiddenInputs := container.Locator("input[type='hidden']")
	count, err := hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 0, count, "empty starter should start with no rows")

	require.NoError(t, container.Locator("[data-add-row]").Click())

	selects := container.Locator("select")
	selectCount, err := selects.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, selectCount, "new row should include the select column")

	value, err := selects.First().InputValue()
	require.NoError(t, err)
	assert.Equal(t, "high", value, "select should default to first option")

	count, err = hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 3, count, "one three-column row should render three hidden inputs")
}

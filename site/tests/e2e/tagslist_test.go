//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagsList_RendersInitialChips(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	gotoTagsList(t, page)

	container := page.Locator("#tagsDemo")
	require.NoError(t, container.WaitFor())

	chips := container.Locator("[data-tagslist-chips] > span")
	count, err := chips.Count()
	require.NoError(t, err)
	assert.Equal(t, 3, count, "should have 3 initial chips")

	// Hidden inputs submit indexed names
	hidden := container.Locator("input[type='hidden'][name='tags[0]']")
	val, err := hidden.GetAttribute("value")
	require.NoError(t, err)
	assert.Equal(t, "production", val)
}

func TestTagsList_AddTagViaButton(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	gotoTagsList(t, page)

	container := page.Locator("#tagsDemo")
	require.NoError(t, container.WaitFor())

	input := container.Locator("[data-tagslist-input]")
	fillAndDispatchInput(t, input, "staging")

	addBtn := container.Locator("[data-tagslist-add]")
	require.NoError(t, addBtn.Click())

	chips := container.Locator("[data-tagslist-chips] > span")
	count, err := chips.Count()
	require.NoError(t, err)
	assert.Equal(t, 4, count, "should have 4 chips after Add click")

	// Input should be cleared
	val, err := input.InputValue()
	require.NoError(t, err)
	assert.Equal(t, "", val, "input cleared after add")
}

func TestTagsList_AddTagViaEnter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	gotoTagsList(t, page)

	container := page.Locator("#labelsDemo")
	require.NoError(t, container.WaitFor())

	input := container.Locator("[data-tagslist-input]")
	fillAndDispatchInput(t, input, "alpha")
	require.NoError(t, input.Press("Enter"))

	chips := container.Locator("[data-tagslist-chips] > span")
	count, err := chips.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Enter key should add a chip")
}

func TestTagsList_EmptyInputDoesNotAdd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	gotoTagsList(t, page)

	container := page.Locator("#labelsDemo")
	require.NoError(t, container.WaitFor())

	addBtn := container.Locator("[data-tagslist-add]")
	require.NoError(t, addBtn.Click())

	// Whitespace-only too
	input := container.Locator("[data-tagslist-input]")
	fillAndDispatchInput(t, input, "   ")
	require.NoError(t, addBtn.Click())

	chips := container.Locator("[data-tagslist-chips] > span")
	count, err := chips.Count()
	require.NoError(t, err)
	assert.Equal(t, 0, count, "empty/whitespace input should not add chips")
}

func TestTagsList_RemoveChip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	gotoTagsList(t, page)

	container := page.Locator("#tagsDemo")
	require.NoError(t, container.WaitFor())

	removeBtn := container.Locator("button[aria-label='Remove tag']").First()
	require.NoError(t, removeBtn.Click())

	chips := container.Locator("[data-tagslist-chips] > span")
	count, err := chips.Count()
	require.NoError(t, err)
	assert.Equal(t, 2, count, "should have 2 chips after removing one")
}

func TestTagsList_DisabledState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	gotoTagsList(t, page)

	container := page.Locator("#disabledTagsDemo")
	require.NoError(t, container.WaitFor())

	chips := container.Locator("[data-tagslist-chips] > span")
	count, err := chips.Count()
	require.NoError(t, err)
	assert.Equal(t, 2, count, "disabled list still renders chips")

	addCount, err := container.Locator("[data-tagslist-add]").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, addCount, "disabled list has no Add button")

	inputCount, err := container.Locator("[data-tagslist-input]").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, inputCount, "disabled list has no input")

	removeCount, err := container.Locator("button[aria-label='Remove tag']").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, removeCount, "disabled chips have no remove buttons")
}

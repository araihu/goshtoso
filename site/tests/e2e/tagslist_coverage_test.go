//go:build e2e

package e2e

import (
	"strings"
	"sync"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTagsList_CoverageDemoNoConsoleErrors loads the demo page, asserts every
// variant container renders, and guards against console errors during an
// add/remove interaction cycle.
func TestTagsList_CoverageDemoNoConsoleErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	var mu sync.Mutex
	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			mu.Lock()
			consoleErrors = append(consoleErrors, msg.Text())
			mu.Unlock()
		}
	})

	gotoTagsList(t, page)

	// All three demo variant containers render.
	for _, id := range []string{"#tagsDemo", "#labelsDemo", "#disabledTagsDemo"} {
		require.NoError(t, page.Locator(id).WaitFor(), "variant %s should render", id)
	}

	// Add two tags to the empty list, then remove one.
	container := page.Locator("#labelsDemo")
	input := container.Locator("[data-tagslist-input]")
	chips := container.Locator("[data-tagslist-chips] > span")

	fillAndDispatchInput(t, input, "alpha")
	require.NoError(t, input.Press("Enter"))
	fillAndDispatchInput(t, input, "beta")
	require.NoError(t, input.Press("Enter"))

	// Wait for the second chip to exist before counting.
	require.NoError(t, chips.Nth(1).WaitFor())
	count, err := chips.Count()
	require.NoError(t, err)
	assert.Equal(t, 2, count, "two tags added via Enter")

	// Hidden inputs use the indexed name prefix.
	hidden := container.Locator("input[type='hidden'][name='labels[0]']")
	val, err := hidden.GetAttribute("value")
	require.NoError(t, err)
	assert.Equal(t, "alpha", val)

	removeBtn := container.Locator("button[aria-label='Remove tag']").First()
	require.NoError(t, removeBtn.Click())
	require.NoError(t, chips.Nth(1).WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateDetached,
	}))
	count, err = chips.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count, "one tag remains after remove")

	mu.Lock()
	errs := append([]string(nil), consoleErrors...)
	mu.Unlock()
	assert.Empty(t, errs, "no console errors expected, got: %s", strings.Join(errs, "; "))
}

//go:build e2e && (full || structuredinput)

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStructuredInputCoverageDemo exercises the /components/structured-input
// demo end to end: page load, x-model text round-trip into the bound hidden
// input, select-column change propagation, add/remove row name re-indexing,
// and the empty-starter add flow. It asserts the demo stays console-error free.
func TestStructuredInputCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			consoleErrors = append(consoleErrors, msg.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/structured-input", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("#structured-input-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}))

	t.Run("Page_Loads", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		assert.Contains(t, title, "Structured Input")
	})

	t.Run("Text_Model_Round_Trip", func(t *testing.T) {
		container := page.Locator("#labelsDemo")
		require.NoError(t, container.WaitFor())

		firstText := container.Locator("input[type='text']").First()
		require.NoError(t, firstText.Fill("region"))

		// x-model on the visible text input drives the bound hidden input value.
		firstHidden := container.Locator("input[type='hidden']").First()
		require.NoError(t, expectHiddenValue(page, "#labelsDemo", "region"))

		name, err := firstHidden.GetAttribute("name")
		require.NoError(t, err)
		assert.Equal(t, "labels[0][key]", name)
	})

	t.Run("Select_Change_Propagates", func(t *testing.T) {
		container := page.Locator("#taintsDemo")
		require.NoError(t, container.WaitFor())

		sel := container.Locator("select").First()
		_, err := sel.SelectOption(playwright.SelectOptionValues{
			Values: &[]string{"NoExecute"},
		})
		require.NoError(t, err)

		value, err := sel.InputValue()
		require.NoError(t, err)
		assert.Equal(t, "NoExecute", value)
	})

	t.Run("Add_Remove_Reindexes_Names", func(t *testing.T) {
		container := page.Locator("#taintsDemo")
		hidden := container.Locator("input[type='hidden']")

		before, err := hidden.Count()
		require.NoError(t, err)

		require.NoError(t, container.Locator("[data-add-row]").Click())
		after, err := hidden.Count()
		require.NoError(t, err)
		assert.Equal(t, before+3, after, "adding one three-column row adds three hidden inputs")

		// The newly added row indexes its hidden input names at the next index.
		lastName, err := hidden.Last().GetAttribute("name")
		require.NoError(t, err)
		assert.Equal(t, "taints[1][effect]", lastName)

		require.NoError(t, container.Locator("button[aria-label='Remove row']").First().Click())
		reduced, err := hidden.Count()
		require.NoError(t, err)
		assert.Equal(t, before, reduced, "removing a row restores the original hidden input count")
	})

	t.Run("Empty_Starter_Adds_From_Defaults", func(t *testing.T) {
		container := page.Locator("#rulesDemo")
		require.NoError(t, container.WaitFor())

		hidden := container.Locator("input[type='hidden']")
		count, err := hidden.Count()
		require.NoError(t, err)
		assert.Equal(t, 0, count, "empty starter renders no rows")

		require.NoError(t, container.Locator("[data-add-row]").Click())
		count, err = hidden.Count()
		require.NoError(t, err)
		assert.Equal(t, 3, count, "one three-column row renders three hidden inputs")

		sel := container.Locator("select").First()
		value, err := sel.InputValue()
		require.NoError(t, err)
		assert.Equal(t, "high", value, "select column defaults to the first option")
	})

	assert.Empty(t, consoleErrors, "structured input demo should not log console errors")
}

// expectHiddenValue waits until the first hidden input inside the given
// container reports the expected value, which Alpine sets via x-bind:value.
func expectHiddenValue(page playwright.Page, containerSelector, want string) error {
	_, err := page.WaitForFunction(
		"([sel, want]) => { const el = document.querySelector(sel + \" input[type='hidden']\"); return el && el.value === want; }",
		[]any{containerSelector, want},
	)
	return err
}

//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestFormCoverageDemo exercises the form demo page interactive sections:
// the Alpine-driven CollapsibleSection toggle and the FlipSection edit/done
// toggle, asserting live aria/visibility state changes and no console errors.
func TestFormCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			consoleErrors = append(consoleErrors, msg.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/form", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("main").First().WaitFor())
	mainText, err := page.Locator("main").First().InnerText()
	require.NoError(t, err)
	require.Contains(t, mainText, "Form")

	// --- CollapsibleSection (#advanced) starts collapsed ---
	collapsibleHeader := page.Locator("#advanced > button")
	require.NoError(t, collapsibleHeader.WaitFor())

	expanded, err := collapsibleHeader.Evaluate("el => el.getAttribute('aria-expanded')", nil)
	require.NoError(t, err)
	require.Equal(t, "false", expanded, "advanced section should start collapsed")

	// Content region hidden while collapsed.
	content := page.Locator("#advanced-content")
	hidden, err := content.IsHidden()
	require.NoError(t, err)
	require.True(t, hidden, "collapsed content should be hidden")

	require.NoError(t, collapsibleHeader.Click())
	_, err = page.WaitForFunction(
		"() => document.querySelector('#advanced > button').getAttribute('aria-expanded') === 'true'",
		nil,
	)
	require.NoError(t, err)
	require.NoError(t, content.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	// --- FlipSection (#network) starts in read-only mode ---
	editButton := page.Locator("#network button", playwright.PageLocatorOptions{
		HasText: "Edit",
	})
	require.NoError(t, editButton.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, editButton.Click())

	// Edit mode reveals the editable fields and the Done button.
	doneButton := page.Locator("#network button", playwright.PageLocatorOptions{
		HasText: "Done",
	})
	require.NoError(t, doneButton.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	podCIDR := page.Locator("#network input[name='pod_cidr']")
	require.NoError(t, podCIDR.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	require.Empty(t, consoleErrors, "form demo page produced console errors: %v", consoleErrors)
}

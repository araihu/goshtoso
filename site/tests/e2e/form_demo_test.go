package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestFormDemoExternalSubmitUsesModalConfirmation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/components/form", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	_, err = page.Evaluate(`() => {
		window.externalFormSubmitCount = 0;
		const form = document.querySelector('#external-form');
		form.addEventListener('submit', () => { window.externalFormSubmitCount += 1; });
	}`)
	require.NoError(t, err)
	initialURL := page.URL()

	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Upgrade (External Button)",
	}).Click())

	count, err := page.Evaluate(`() => String(window.externalFormSubmitCount)`)
	require.NoError(t, err)
	require.Equal(t, "0", count, "opening the modal must not submit the form")

	dialog := page.Locator("[role='dialog'][aria-labelledby='externalSubmitConfirmTitle']")
	require.NoError(t, dialog.WaitFor())
	require.NoError(t, dialog.GetByRole("button", playwright.LocatorGetByRoleOptions{
		Name: "Confirm upgrade",
	}).Click())

	_, err = page.WaitForFunction(`() => window.externalFormSubmitCount === 1`, nil)
	require.NoError(t, err)

	toast := page.Locator("#toast-container [data-toast-id]", playwright.PageLocatorOptions{
		HasText: "Upgrade request submitted",
	})
	require.NoError(t, toast.WaitFor())
	text, err := toast.InnerText()
	require.NoError(t, err)
	require.Contains(t, text, "v1.31.5")
	require.Equal(t, initialURL, page.URL(), "HTMX submit should not reload or navigate the page")
}

func TestFormDemoCollapsibleComboboxEscapesAccordionClip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/components/form", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#advanced").ScrollIntoViewIfNeeded())
	require.NoError(t, page.Locator("#advanced > button").Click())
	require.NoError(t, page.Locator("#advanced-content").WaitFor())
	require.NoError(t, page.Locator("#log-level-trigger").Click())
	require.NoError(t, page.Locator(`#log-level [data-combobox-option][data-value="debug"]`).WaitFor())

	hitDebugOption, err := page.Evaluate(`() => {
		const option = document.querySelector('#log-level [data-combobox-option][data-value="debug"]');
		if (!option) return false;
		const rect = option.getBoundingClientRect();
		const target = document.elementFromPoint(rect.left + 12, rect.top + rect.height / 2);
		return Boolean(target && target.closest('#log-level [data-combobox-option][data-value="debug"]'));
	}`)
	require.NoError(t, err)
	require.Equal(t, true, hitDebugOption, "expanded accordion content must not clip the open combobox dropdown")
}

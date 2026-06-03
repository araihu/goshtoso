package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestGettingStarted_StarterRepoAndLiveOutcome(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	page.SetDefaultTimeout(5000)

	_, err := page.Goto(baseURL+"/getting-started", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("h2:has-text('Expected outcome')").WaitFor())
	require.NoError(t, page.GetByText("git clone https://github.com/araihu/goshtoso-getting-started dog-breeds").WaitFor())
	require.NoError(t, page.GetByText("Starter app assets").WaitFor())
	require.NoError(t, page.Locator("#getting-started-preview").WaitFor())
	require.NoError(t, page.GetByText("Australian Shepherd").WaitFor())

	search := page.Locator("#getting-started-preview input[type='search']")
	require.NoError(t, search.Fill("husky"))
	_, err = search.Evaluate(`(el) => el.dispatchEvent(new Event('input', {bubbles: true}))`, nil)
	require.NoError(t, err)

	_, err = page.WaitForFunction(`() => {
		const preview = document.querySelector('#getting-started-preview');
		if (!preview) return false;
		const text = preview.innerText;
		return text.includes('Siberian Husky') && !text.includes('Australian Shepherd');
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err)
}

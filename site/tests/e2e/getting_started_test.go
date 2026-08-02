//go:build e2e && full

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

	versionBadge := page.Locator(".component-doc-shell__brand-badge")
	require.NoError(t, versionBadge.WaitFor())
	require.Equal(t, goshtosoDocsVersion, mustText(t, versionBadge))
	require.Equal(t, "https://github.com/araihu/goshtoso/releases/tag/"+goshtosoDocsVersion, mustAttribute(t, versionBadge, "href"))

	contentEndsNearFooter, err := page.Evaluate(`() => {
		const main = document.getElementById('main-content');
		const footer = main && main.querySelector('footer');
		if (!main || !footer) return false;
		return main.getBoundingClientRect().bottom - footer.getBoundingClientRect().bottom <= 96;
	}`, nil)
	require.NoError(t, err)
	require.Equal(t, true, contentEndsNearFooter, "content should end near its footer instead of reserving a viewport-sized scroll tail")

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

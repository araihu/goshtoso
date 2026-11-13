//go:build e2e && (full || combobox)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestCombobox_ClientSearchDoesNotSubmitParentQuery(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/components/combobox")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	require.NoError(t, page.Locator("#industry-trigger").Click())
	search := page.Locator("#industry [data-combobox-search]")
	require.NoError(t, search.Fill("no-such-industry"))
	empty := page.Locator("#industry [data-combobox-empty]")
	require.NoError(t, empty.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	require.NoError(t, search.Fill("tech"))
	require.NoError(t, empty.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
	require.NoError(t, page.Locator(`#industry [data-combobox-option][data-value="tech"]`).WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	label, err := search.GetAttribute("aria-label")
	require.NoError(t, err)
	require.Equal(t, "Search…", label)
	require.NoError(t, page.Locator(`#industry [data-combobox-option][data-value="tech"]`).Click())
	values, err := page.Evaluate(`()=>{
   const form=document.querySelector('#combobox-client-form');
   const data=new FormData(form);
   return {queries:data.getAll('q'),industries:data.getAll('industry')};
 }`)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"queries": []any{"records"}, "industries": []any{"tech"}}, values)
}

//go:build e2e && (full || combobox)

package e2e

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestComboboxDisablePersistence(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/components/combobox")
	require.NoError(t, err)
	_, err = page.Evaluate(`() => sessionStorage.setItem('goshtoso:combobox:skills:selected', '["go"]')`)
	require.NoError(t, err)
	_, err = page.Reload()
	require.NoError(t, err)
	require.Equal(t, 0, mustCount(t, page.Locator(`#skills input[type=hidden][name=skills]`)))
	require.NoError(t, page.Locator("#skills-trigger").Click())
	require.NoError(t, page.Locator(`#skills [data-value=rust]`).Click())
	require.Equal(t, 1, mustCount(t, page.Locator(`#skills input[type=hidden][value=rust]`)))
	stored, err := page.Evaluate(`()=>sessionStorage.getItem('goshtoso:combobox:skills:selected')`)
	require.NoError(t, err)
	require.Equal(t, `["go"]`, stored)
}

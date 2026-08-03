//go:build e2e && (full || table)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestTableDemoIDsAreUnique(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	duplicates, err := page.Evaluate(`() => {
		const counts = new Map();
		document.querySelectorAll('[id]').forEach((element) => {
			counts.set(element.id, (counts.get(element.id) || 0) + 1);
		});
		return Array.from(document.querySelectorAll('table'))
			.map((table) => table.id)
			.filter((id) => !id || counts.get(id) !== 1);
	}`)
	require.NoError(t, err)
	require.Equal(t, []any{}, duplicates, "every demo table must own one unique, non-empty DOM ID")
}

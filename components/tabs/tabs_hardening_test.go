package tabs

import (
	"bytes"
	"context"
	"html"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderTabs(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Tabs(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestTabsData_EscapesDefaultAndSyncHashTabIDs(t *testing.T) {
	cfg := Config{
		DefaultTab: "billing'\\x\n\r\t\u2028\u2029",
		SyncHash:   true,
		Tabs: []Tab{
			{ID: "billing'\\x\n\r\t\u2028\u2029", Label: "Billing"},
			{ID: "usage'\\x\n\r\t\u2028\u2029", Label: "Usage"},
		},
	}

	data := tabsData(cfg)

	assert.Contains(t, data, `selectedTab:'billing\'\\x\n\r\t\u2028\u2029'`)
	assert.Contains(t, data, `var v=['billing\'\\x\n\r\t\u2028\u2029','usage\'\\x\n\r\t\u2028\u2029'];`)
	assert.NotContains(t, data, "billing'\\x\n")
	assert.NotContains(t, data, "\r")
	assert.NotContains(t, data, "\u2028")
	assert.NotContains(t, data, "\u2029")
}

func TestTabs_RenderedAlpineExpressionsEscapeTabID(t *testing.T) {
	rendered := renderTabs(t, Config{
		ID:         "account",
		DefaultTab: `billing'\x`,
		Tabs: []Tab{
			{ID: `billing'\x`, Label: "Billing"},
		},
	})
	browserHTML := html.UnescapeString(rendered)

	assert.Contains(t, browserHTML, `x-on:click="selectedTab = 'billing\'\\x'"`)
	assert.Contains(t, browserHTML, `x-bind:aria-selected="selectedTab === 'billing\'\\x'"`)
	assert.Contains(t, browserHTML, `x-show="selectedTab === 'billing\'\\x'"`)
}

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
		DefaultTab: `billing'\x`,
		SyncHash:   true,
		Tabs: []Tab{
			{ID: `billing'\x`, Label: "Billing"},
			{ID: `usage'\x`, Label: "Usage"},
		},
	}

	data := tabsData(cfg)

	assert.Contains(t, data, `selectedTab:'billing\'\\x'`)
	assert.Contains(t, data, `var v=['billing\'\\x','usage\'\\x'];`)
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

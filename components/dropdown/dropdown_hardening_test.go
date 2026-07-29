package dropdown

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderDropdown(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Dropdown(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestDropdownLink_SanitizesUnsafeHref(t *testing.T) {
	rendered := renderDropdown(t, Config{
		ID: "menu",
		Sections: []Section{{Items: []Item{
			{Label: "Bad", Href: "javascript:alert(1)"},
		}}},
	})

	assert.NotContains(t, rendered, `href="javascript:alert`)
	assert.Contains(t, rendered, `href="about:invalid#TemplFailedSanitizationURL"`)
}

func TestDropdownLink_PreservesRelativeHref(t *testing.T) {
	rendered := renderDropdown(t, Config{
		ID: "menu",
		Sections: []Section{{Items: []Item{
			{Label: "Docs", Href: "/docs?tab=api"},
		}}},
	})

	assert.Contains(t, rendered, `href="/docs?tab=api"`)
}

func TestDropdownKeyboardOpenFocusesFirstEnabledItemAndEscapeReturnsToTrigger(t *testing.T) {
	rendered := renderDropdown(t, Config{
		Label: "Actions",
		Sections: []Section{{Items: []Item{
			{Label: "Unavailable", Disabled: true},
			{Label: "Open", Href: "/open"},
		}}},
	})

	assert.Contains(t, rendered, `x-ref="trigger"`)
	assert.Contains(t, rendered, `x-ref="menu"`)
	assert.Contains(t, rendered, `[role=menuitem]:not([disabled]):not([hidden])`)
	assert.Contains(t, rendered, `item.getClientRects().length &gt; 0`)
	assert.Contains(t, rendered, `focusFirstItem()`)
	assert.Contains(t, rendered, `focusAdjacentItem(1)`)
	assert.Contains(t, rendered, `focusAdjacentItem(-1)`)
	assert.NotContains(t, rendered, `$focus.wrap()`)
	assert.Contains(t, rendered, `x-trap.noreturn="openedWithKeyboard"`)
	assert.Contains(t, rendered, `$refs.trigger.focus()`)
	assert.Equal(t, 1, strings.Count(rendered, `role="menu"`))
}

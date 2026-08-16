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

func TestDropdownHTMXItemUsesButtonAndSuppressesConflictingHrefWhenDisabled(t *testing.T) {
	rendered := renderDropdown(t, Config{Sections: []Section{{Items: []Item{
		{
			Label: "Archive", Href: "/archive", ID: "archive",
			HTMX: &HTMXConfig{Post: "/api/archive", Target: "#result", Swap: "outerHTML", Trigger: "click", Vals: `{"id":42}`, Confirm: "Archive?"},
		},
		{Label: "Locked", HTMX: &HTMXConfig{Post: "/api/locked"}, Disabled: true},
	}}}})

	for _, want := range []string{
		`<button type="button"`, `id="archive"`, `hx-post="/api/archive"`,
		`hx-target="#result"`, `hx-swap="outerHTML"`, `hx-trigger="click"`,
		`hx-vals="{&#34;id&#34;:42}"`, `hx-confirm="Archive?"`,
	} {
		assert.Contains(t, rendered, want)
	}
	assert.NotContains(t, rendered, `href="/archive"`)
	assert.NotContains(t, rendered, `hx-post="/api/locked"`)
}

func TestDropdownHTMXPostTakesPrecedenceOverGet(t *testing.T) {
	rendered := renderDropdown(t, Config{Sections: []Section{{Items: []Item{{
		Label: "Save", HTMX: &HTMXConfig{Get: "/preview", Post: "/save"},
	}}}}})

	assert.Contains(t, rendered, `hx-post="/save"`)
	assert.NotContains(t, rendered, `hx-get="/preview"`)
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
	assert.Contains(t, rendered, `focusFirstItem()`)
	assert.Contains(t, rendered, `focusAdjacentItem(1)`)
	assert.Contains(t, rendered, `focusAdjacentItem(-1)`)
	assert.NotContains(t, rendered, `$focus.wrap()`)
	assert.Contains(t, rendered, `x-trap.noreturn="openedWithKeyboard"`)
	assert.Contains(t, rendered, `x-data="goshtosoDropdown($el)"`)
	assert.Contains(t, rendered, `x-on:keydown.esc.window="closeAndFocus()"`)
	assert.NotContains(t, rendered, `<script`)
	assert.Equal(t, 1, strings.Count(rendered, `role="menu"`))
}

func TestDropdownMenuTransitionsProvideReducedMotionFallback(t *testing.T) {
	for _, cfg := range []Config{
		{
			Label:    "Actions",
			Sections: []Section{{Items: []Item{{Label: "Open", Href: "/open"}}}},
		},
		{
			TriggerMode: TriggerContext,
			Sections:    []Section{{Items: []Item{{Label: "Open", Href: "/open"}}}},
		},
	} {
		rendered := renderDropdown(t, cfg)
		assert.Contains(t, cfg.buttonClasses(), "transition motion-reduce:transition-none")

		assert.Contains(t, rendered, `x-transition:enter="transition ease-out duration-150 motion-reduce:transition-none"`)
		assert.Contains(t, rendered, `x-transition:enter-start="opacity-0 scale-95 motion-reduce:opacity-100 motion-reduce:scale-100"`)
		assert.Contains(t, rendered, `x-transition:enter-end="opacity-100 scale-100"`)
		assert.Contains(t, rendered, `x-transition:leave="transition ease-in duration-100 motion-reduce:transition-none"`)
		assert.Contains(t, rendered, `x-transition:leave-start="opacity-100 scale-100"`)
		assert.Contains(t, rendered, `x-transition:leave-end="opacity-0 scale-95 motion-reduce:opacity-100 motion-reduce:scale-100"`)
	}
}

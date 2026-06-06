package dropdown

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
)

func testDropdownIcon(label string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<svg data-icon="`+label+`" aria-hidden="true"></svg>`)
		return err
	})
}

func TestCoverageRenderHoverDropdownWithIconOnlyTrigger(t *testing.T) {
	rendered := renderDropdown(t, Config{
		ID:              "hover-actions",
		Label:           "More actions",
		TriggerMode:     TriggerHover,
		TriggerIcon:     testDropdownIcon("trigger"),
		TriggerIconOnly: true,
		MenuAlign:       AlignEnd,
		Sections: []Section{{Items: []Item{
			{Label: "Open", Href: "/open", Icon: testDropdownIcon("open")},
			{Label: "Archive", Disabled: true, Tooltip: "Locked"},
		}}},
	})

	assert.Contains(t, rendered, `id="hover-actions"`)
	assert.Contains(t, rendered, `x-on:mouseover="isOpen = true"`)
	assert.Contains(t, rendered, `x-on:mouseleave.prevent=`)
	assert.Contains(t, rendered, `aria-label="More actions"`)
	assert.Contains(t, rendered, `data-icon="trigger"`)
	assert.Contains(t, rendered, `right-0`)
	assert.Contains(t, rendered, `href="/open"`)
	assert.Contains(t, rendered, `disabled`)
	assert.Contains(t, rendered, `title="Locked"`)
	assert.Contains(t, rendered, `opacity-50`)
	assert.NotContains(t, rendered, `More actions<svg`)
}

func TestCoverageRenderContextDropdownItems(t *testing.T) {
	rendered := renderDropdown(t, Config{
		ID:          "context-actions",
		TriggerMode: TriggerContext,
		Sections: []Section{
			{Items: []Item{
				{Label: "Rename", Icon: testDropdownIcon("rename"), Shortcut: "R"},
				{Label: "Duplicate", Icon: testDropdownIcon("duplicate"), Shortcut: "D", ShortcutIcon: testDropdownIcon("shift")},
			}},
			{Items: []Item{
				{Label: "Delete", OnClick: "deleted = true", Danger: true, ID: "delete-item"},
				{Label: "Export", Disabled: true, Tooltip: "Unavailable", ID: "export-item"},
			}},
		},
	})

	for _, want := range []string{
		`id="context-actions"`,
		`aria-label="context menu"`,
		`x-on:contextmenu.prevent="isOpen = true"`,
		`top-8 absolute left-0`,
		`<ul class="flex flex-col py-1.5" role="none">`,
		`tabindex="0"`,
		`data-icon="rename"`,
		`data-icon="shift"`,
		`>R</div>`,
		`id="delete-item"`,
		`x-on:click="deleted = true"`,
		`text-danger`,
		`id="export-item"`,
		`disabled`,
		`title="Unavailable"`,
	} {
		assert.Contains(t, rendered, want)
	}
}

func TestCoverageRenderDividersAndDefaultIconOnlyAriaLabel(t *testing.T) {
	rendered := renderDropdown(t, Config{
		ID:              "unnamed-actions",
		TriggerIcon:     testDropdownIcon("overflow"),
		TriggerIconOnly: true,
		Sections: []Section{
			{Items: []Item{{Label: "One", Href: "#one"}}},
			{Items: []Item{{Label: "Two", OnClick: "two = true"}}},
		},
	})

	assert.Contains(t, rendered, `aria-label="open menu"`)
	assert.Contains(t, rendered, `divide-y divide-outline dark:divide-outline-dark`)
	assert.Contains(t, rendered, `class="flex flex-col py-1.5"`)
	assert.Contains(t, rendered, `x-on:click="two = true"`)
	assert.Equal(t, 2, strings.Count(rendered, `role="menuitem"`))
}

func TestCoverageConfigHasShortcuts(t *testing.T) {
	assert.False(t, Config{}.HasShortcuts())
	assert.False(t, Config{
		Sections: []Section{{Items: []Item{{Label: "Copy"}}}},
	}.HasShortcuts())
	assert.True(t, Config{
		Sections: []Section{{Items: []Item{{Label: "Copy", Shortcut: "C"}}}},
	}.HasShortcuts())
}

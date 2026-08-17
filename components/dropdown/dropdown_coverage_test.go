package dropdown

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	templruntime "github.com/a-h/templ/runtime"
	"github.com/stretchr/testify/assert"
)

func testDropdownIcon(label string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<svg data-icon="`+label+`" aria-hidden="true"></svg>`)
		return err
	})
}

func failingDropdownIcon(err error) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return err
	})
}

type failingDropdownWriter struct {
	failAfter int
	written   int
}

func (w *failingDropdownWriter) Write(p []byte) (int, error) {
	if w.written+len(p) > w.failAfter {
		return 0, errors.New("dropdown write failed")
	}
	w.written += len(p)
	return len(p), nil
}

func tinyDropdownBuffer(w io.Writer) *templruntime.Buffer {
	previousSize := templruntime.DefaultBufferSize
	templruntime.DefaultBufferSize = 1
	defer func() { templruntime.DefaultBufferSize = previousSize }()

	buf := &templruntime.Buffer{}
	buf.Reset(w)
	return buf
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
	assert.Contains(t, rendered, `x-on:mouseenter="clearScheduledClose(); open()"`)
	assert.Contains(t, rendered, `x-on:mouseleave="scheduleClose()"`)
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
		`x-on:contextmenu.prevent="open()"`,
		`left-0 top-full mt-2 absolute`,
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
	assert.False(t, Config{}.hasShortcuts())
	assert.False(t, Config{
		Sections: []Section{{Items: []Item{{Label: "Copy"}}}},
	}.hasShortcuts())
	assert.True(t, Config{
		Sections: []Section{{Items: []Item{{Label: "Copy", Shortcut: "C"}}}},
	}.hasShortcuts())
}

func TestCoverageDropdownRenderPropagatesNestedIconErrors(t *testing.T) {
	renderErr := errors.New("dropdown icon render failed")
	icon := failingDropdownIcon(renderErr)

	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "click icon only trigger",
			cfg: Config{
				TriggerIcon:     icon,
				TriggerIconOnly: true,
				Sections:        []Section{{Items: []Item{{Label: "Open", Href: "/open"}}}},
			},
		},
		{
			name: "hover icon only trigger",
			cfg: Config{
				TriggerMode:     TriggerHover,
				TriggerIcon:     icon,
				TriggerIconOnly: true,
				Sections:        []Section{{Items: []Item{{Label: "Open", Href: "/open"}}}},
			},
		},
		{
			name: "context trigger icon",
			cfg: Config{
				TriggerMode: TriggerContext,
				TriggerIcon: icon,
				Sections:    []Section{{Items: []Item{{Label: "Open", Href: "/open"}}}},
			},
		},
		{
			name: "link item icon",
			cfg: Config{
				Sections: []Section{{Items: []Item{{Label: "Open", Href: "/open", Icon: icon}}}},
			},
		},
		{
			name: "button item icon",
			cfg: Config{
				Sections: []Section{{Items: []Item{{Label: "Run", OnClick: "run()", Icon: icon}}}},
			},
		},
		{
			name: "context item icon",
			cfg: Config{
				TriggerMode: TriggerContext,
				Sections:    []Section{{Items: []Item{{Label: "Open", Href: "/open", Icon: icon}}}},
			},
		},
		{
			name: "context shortcut icon",
			cfg: Config{
				TriggerMode: TriggerContext,
				Sections: []Section{{Items: []Item{{
					Label:        "Open",
					Href:         "/open",
					Shortcut:     "O",
					ShortcutIcon: icon,
				}}}},
			},
		},
		{
			name: "context button item icon",
			cfg: Config{
				TriggerMode: TriggerContext,
				Sections:    []Section{{Items: []Item{{Label: "Run", OnClick: "run()", Icon: icon}}}},
			},
		},
		{
			name: "context button shortcut icon",
			cfg: Config{
				TriggerMode: TriggerContext,
				Sections: []Section{{Items: []Item{{
					Label:        "Run",
					OnClick:      "run()",
					Shortcut:     "R",
					ShortcutIcon: icon,
				}}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rendered strings.Builder
			err := Dropdown(tt.cfg).Render(context.Background(), &rendered)
			assert.ErrorIs(t, err, renderErr)
		})
	}
}

func TestCoverageDropdownHelpersPropagateCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := Config{
		Label: "Actions",
		Sections: []Section{{Items: []Item{{
			Label:    "Open",
			Href:     "/open",
			Shortcut: "O",
		}}}},
	}
	item := cfg.Sections[0].Items[0]

	tests := []struct {
		name      string
		component templ.Component
	}{
		{name: "dropdown", component: Dropdown(cfg)},
		{name: "click dropdown", component: clickDropdown(cfg)},
		{name: "hover dropdown", component: hoverDropdown(cfg)},
		{name: "context dropdown", component: contextDropdown(cfg)},
		{name: "trigger button", component: triggerButton(cfg)},
		{name: "hover trigger button", component: hoverTriggerButton(cfg)},
		{name: "context trigger button", component: contextTriggerButton(cfg)},
		{name: "dropdown menu", component: dropdownMenu(cfg)},
		{name: "context dropdown menu", component: contextDropdownMenu(cfg)},
		{name: "menu item", component: menuItem(cfg, item)},
		{name: "menu item link", component: menuItemLink(cfg, item)},
		{name: "menu item button", component: menuItemButton(cfg, Item{Label: "Run", OnClick: "run()"})},
		{name: "context menu item", component: contextMenuItem(cfg, item)},
		{name: "context menu item plain", component: contextMenuItemPlain(cfg, item)},
		{name: "context menu item button", component: contextMenuItemButton(cfg, Item{Label: "Run", OnClick: "run()"})},
		{name: "chevron icon", component: chevronIcon()},
		{name: "dots icon", component: dotsIcon()},
		{name: "cmd icon", component: cmdIcon()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rendered strings.Builder
			err := tt.component.Render(ctx, &rendered)
			assert.ErrorIs(t, err, context.Canceled)
			assert.Empty(t, rendered.String())
		})
	}
}

func TestCoverageDropdownRenderPropagatesWriteErrors(t *testing.T) {
	configs := []struct {
		name string
		cfg  Config
	}{
		{
			name: "click",
			cfg: Config{
				ID:              "click-actions",
				Label:           "Actions",
				TriggerIcon:     testDropdownIcon("trigger"),
				TriggerIconOnly: true,
				MenuAlign:       AlignEnd,
				Sections: []Section{
					{Items: []Item{
						{Label: "Open", Href: "/open", Icon: testDropdownIcon("open"), Tooltip: "Open file", ID: "open-item"},
						{Label: "Archive", OnClick: "archive()", Icon: testDropdownIcon("archive")},
					}},
					{Items: []Item{
						{Label: "Delete", OnClick: "deleteItem()", Danger: true, Disabled: true, Tooltip: "Locked"},
					}},
				},
			},
		},
		{
			name: "hover",
			cfg: Config{
				ID:              "hover-actions",
				Label:           "More actions",
				TriggerMode:     TriggerHover,
				TriggerIcon:     testDropdownIcon("trigger"),
				TriggerIconOnly: true,
				Sections: []Section{{Items: []Item{
					{Label: "Open", Href: "/open", Icon: testDropdownIcon("open"), Tooltip: "Open file", ID: "hover-open"},
					{Label: "Run", OnClick: "run()", Icon: testDropdownIcon("run"), Tooltip: "Run task", ID: "hover-run"},
				}}},
			},
		},
		{
			name: "context",
			cfg: Config{
				ID:          "context-actions",
				Label:       "Context actions",
				TriggerMode: TriggerContext,
				TriggerIcon: testDropdownIcon("trigger"),
				Sections: []Section{
					{Items: []Item{
						{Label: "Rename", Href: "/rename", Icon: testDropdownIcon("rename"), Shortcut: "R", Tooltip: "Rename item", ID: "rename-item"},
						{Label: "Duplicate", Href: "/duplicate", Icon: testDropdownIcon("duplicate"), Shortcut: "D", ShortcutIcon: testDropdownIcon("shift")},
					}},
					{Items: []Item{
						{Label: "Delete", OnClick: "deleteItem()", Icon: testDropdownIcon("delete"), Shortcut: "X", Danger: true},
						{Label: "Export", OnClick: "exportItem()", Shortcut: "E", ShortcutIcon: testDropdownIcon("cmd"), Disabled: true, Tooltip: "Unavailable"},
					}},
				},
			},
		},
	}

	for _, tt := range configs {
		t.Run(tt.name, func(t *testing.T) {
			rendered := renderDropdown(t, tt.cfg)

			var sawError bool
			for failAfter := 0; failAfter < len(rendered); failAfter++ {
				writer := &failingDropdownWriter{failAfter: failAfter}
				err := Dropdown(tt.cfg).Render(context.Background(), tinyDropdownBuffer(writer))
				if err != nil {
					sawError = true
				}
			}
			assert.True(t, sawError)

			writer := &failingDropdownWriter{failAfter: len(rendered) + 1}
			buf := tinyDropdownBuffer(writer)
			assert.NoError(t, Dropdown(tt.cfg).Render(context.Background(), buf))
			assert.NoError(t, buf.Flush())
			assert.Equal(t, len(rendered), writer.written)
		})
	}
}

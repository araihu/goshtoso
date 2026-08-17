package popover

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
	"github.com/stretchr/testify/require"
)

func renderPopover(t *testing.T, cfg Config) string {
	t.Helper()
	var output bytes.Buffer
	require.NoError(t, Popover(cfg).Render(context.Background(), &output))
	return output.String()
}

func TestPopoverRendersConsumerTriggerAndPanelContract(t *testing.T) {
	html := renderPopover(t, Config{
		ID:        "actions",
		Trigger:   templ.Raw(`<button type="button">Actions</button>`),
		Content:   templ.Raw(`<div data-content>Menu</div>`),
		Placement: PlacementBottomEnd,
		Role:      "menu",
		Label:     "Actions menu",
		TrapFocus: true,
	})

	for _, want := range []string{
		`data-popover-root`,
		`data-popover-trigger`,
		`data-popover-panel`,
		`id="actions-panel"`,
		`role="menu"`,
		`aria-label="Actions menu"`,
		`right-0`,
		`x-trap.noreturn`,
		`<button type="button">Actions</button>`,
		`<div data-content>Menu</div>`,
	} {
		require.Contains(t, html, want)
	}
}

func TestPopoverDefaultsToClickAndBottomStart(t *testing.T) {
	html := renderPopover(t, Config{
		ID:      "default",
		Trigger: templ.Raw(`<button type="button">Open</button>`),
		Content: templ.Raw(`<p>Content</p>`),
	})

	for _, want := range []string{
		`x-on:click="toggle()"`,
		`x-on:keydown.enter.prevent="openFromKeyboard()"`,
		`x-on:keydown.space.prevent="openFromKeyboard()"`,
		`left-0`,
		`top-full`,
		`x-on:keydown.esc.window="closeAndFocus()"`,
	} {
		require.Contains(t, html, want)
	}
}

func TestPopoverIdentity(t *testing.T) {
	require.Equal(t, components.KindPopover, Popover(Config{}).Kind())
}

func TestPopoverPropagatesCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Popover(Config{
		Trigger: templ.Raw(`<button>Open</button>`),
		Content: templ.Raw(`<p>Content</p>`),
	}).Render(ctx, &bytes.Buffer{})

	require.ErrorIs(t, err, context.Canceled)
}

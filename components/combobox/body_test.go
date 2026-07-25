package combobox

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBody_MultiSelect_RendersHiddenInputsPerSelected(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID:              "status",
		Name:            "status",
		Mode:            ModeMultiple,
		ToggleEndpoint:  "/toggle",
		OptionsEndpoint: "/options",
		ClearEndpoint:   "/clear",
		Source:          Source{Static: []Option{{Value: "creating", Label: "Creating"}, {Value: "running", Label: "Running"}}},
	}
	state := State{
		Options:  cfg.Source.Static,
		Selected: []string{"creating", "running"},
	}

	var buf bytes.Buffer
	require.NoError(t, body(cfg, state).Render(context.Background(), &buf))
	html := buf.String()

	assert.Equal(t, 2, strings.Count(html, `type="hidden"`), "one hidden input per selected value")
	assert.Contains(t, html, `name="status"`)
	assert.Contains(t, html, `value="creating"`)
	assert.Contains(t, html, `value="running"`)
	assert.Contains(t, html, `data-combobox-body`)
	assert.Contains(t, html, `id="status-body"`)
}

func TestBody_MultiSelect_TriggerLabelReflectsSelection(t *testing.T) {
	cfg := Config{
		ID:             "status",
		Name:           "status",
		Mode:           ModeMultiple,
		Placeholder:    "Select statuses",
		ToggleEndpoint: "/t", OptionsEndpoint: "/o", ClearEndpoint: "/c",
		Source: Source{Static: []Option{{Value: "a", Label: "A"}, {Value: "b", Label: "B"}}},
	}

	t.Run("no selection shows placeholder", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, triggerLabelOOB(cfg, State{Options: cfg.Source.Static}).Render(context.Background(), &buf))
		assert.Contains(t, buf.String(), "Select statuses")
	})

	t.Run("one selection shows label", func(t *testing.T) {
		var buf bytes.Buffer
		s := State{Options: cfg.Source.Static, Selected: []string{"a"}}
		require.NoError(t, triggerLabelOOB(cfg, s).Render(context.Background(), &buf))
		assert.Contains(t, buf.String(), "A")
	})

	t.Run("multi-selection shows count", func(t *testing.T) {
		var buf bytes.Buffer
		s := State{Options: cfg.Source.Static, Selected: []string{"a", "b"}}
		require.NoError(t, triggerLabelOOB(cfg, s).Render(context.Background(), &buf))
		assert.Contains(t, buf.String(), "2 selected")
	})
}

func TestBody_SingleSelect_TriggerLabelShowsSelectedOption(t *testing.T) {
	cfg := Config{
		ID:             "provider",
		Name:           "provider",
		Mode:           ModeSingle,
		ToggleEndpoint: "/t", OptionsEndpoint: "/o", ClearEndpoint: "/c",
		Source: Source{Static: []Option{{Value: "maas", Label: "MAAS"}, {Value: "eks", Label: "EKS"}}},
	}
	state := State{Options: cfg.Source.Static, Selected: []string{"maas"}}

	t.Run("hidden input in body", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, body(cfg, state).Render(context.Background(), &buf))
		assert.Equal(t, 1, strings.Count(buf.String(), `type="hidden"`), "single-select has one hidden input")
	})

	t.Run("label in trigger label OOB", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, triggerLabelOOB(cfg, state).Render(context.Background(), &buf))
		assert.Contains(t, buf.String(), "MAAS")
	})
}

func TestOptionsList_AriaSelectedMatchesState(t *testing.T) {
	cfg := Config{
		ID: "t", Name: "t", Mode: ModeMultiple,
		ToggleEndpoint: "/t", OptionsEndpoint: "/o", ClearEndpoint: "/c",
		Source: Source{Static: []Option{{Value: "a"}, {Value: "b"}, {Value: "c"}}},
	}
	state := State{Options: cfg.Source.Static, Selected: []string{"a", "c"}}

	var buf bytes.Buffer
	require.NoError(t, optionsList(cfg, state).Render(context.Background(), &buf))
	html := buf.String()

	assert.Contains(t, html, `data-value="a" aria-selected="true"`)
	assert.Contains(t, html, `data-value="b" aria-selected="false"`)
	assert.Contains(t, html, `data-value="c" aria-selected="true"`)
}

func TestOptionsList_LiHasHXAttributesForToggle(t *testing.T) {
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		ToggleEndpoint:  "/ui/combobox/status/toggle",
		OptionsEndpoint: "/ui/combobox/status/options",
		ClearEndpoint:   "/ui/combobox/status/clear",
		Source:          Source{LazyEndpoint: "/ui/combobox/status/options"},
		DependsOn:       []string{"provider"},
	}
	state := State{Options: []Option{{Value: "running", Label: "Running"}}}

	var buf bytes.Buffer
	require.NoError(t, optionsList(cfg, state).Render(context.Background(), &buf))
	html := buf.String()

	assert.Contains(t, html, `hx-post="/ui/combobox/status/toggle"`)
	assert.Contains(t, html, `hx-target="closest [data-combobox-body]"`)
	assert.Contains(t, html, `hx-swap="outerHTML"`)
	assert.Contains(t, html, `hx-vals`)
	assert.Contains(t, html, `&#34;value&#34;:&#34;running&#34;`)
	assert.Contains(t, html, `hx-include="closest [data-combobox] input[type=hidden],[name=&#39;provider&#39;]"`)
}

func TestBody_SearchInput_RenderedWhenEnabled(t *testing.T) {
	cfg := Config{
		ID: "users", Name: "users", Mode: ModeMultiple,
		ToggleEndpoint: "/t", OptionsEndpoint: "/ui/combobox/users/options", ClearEndpoint: "/c",
		Source:       Source{Static: []Option{{Value: "a"}}},
		EnableSearch: true,
	}

	t.Run("renders when enabled", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, body(cfg, State{Options: cfg.Source.Static, Search: "foo"}).Render(context.Background(), &buf))
		html := buf.String()
		assert.Contains(t, html, `data-combobox-search`)
		assert.Contains(t, html, `value="foo"`)
		assert.Contains(t, html, `hx-get="/ui/combobox/users/options"`)
		assert.Contains(t, html, `hx-trigger="input changed delay:200ms"`)
		assert.Contains(t, html, `hx-target="#users-options"`)
		assert.Contains(t, html, `hx-swap="outerHTML"`)
	})

	t.Run("absent when disabled", func(t *testing.T) {
		cfg2 := cfg
		cfg2.EnableSearch = false
		var buf bytes.Buffer
		require.NoError(t, body(cfg2, State{Options: cfg2.Source.Static}).Render(context.Background(), &buf))
		assert.NotContains(t, buf.String(), `data-combobox-search`)
	})
}

func TestBody_ClearAllButton(t *testing.T) {
	cfg := Config{
		ID: "t", Name: "t", Mode: ModeMultiple,
		ToggleEndpoint: "/t", OptionsEndpoint: "/o", ClearEndpoint: "/ui/combobox/t/clear",
		Source:         Source{LazyEndpoint: "/o"},
		EnableClearAll: true,
	}
	opts := []Option{{Value: "a"}}

	t.Run("renders when enabled and selection non-empty", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, body(cfg, State{Options: opts, Selected: []string{"a"}}).Render(context.Background(), &buf))
		html := buf.String()
		assert.Contains(t, html, `data-combobox-clear-all`)
		assert.Contains(t, html, `hx-post="/ui/combobox/t/clear"`)
		assert.Contains(t, html, `hx-target="closest [data-combobox-body]"`)
	})

	t.Run("absent when selection empty", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, body(cfg, State{Options: opts}).Render(context.Background(), &buf))
		assert.NotContains(t, buf.String(), `data-combobox-clear-all`)
	})

	t.Run("absent when disabled", func(t *testing.T) {
		cfg2 := cfg
		cfg2.EnableClearAll = false
		var buf bytes.Buffer
		require.NoError(t, body(cfg2, State{Options: opts, Selected: []string{"a"}}).Render(context.Background(), &buf))
		assert.NotContains(t, buf.String(), `data-combobox-clear-all`)
	})
}

func TestCombobox_OuterShell_HasAlpineDataAndContainsBody(t *testing.T) {
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		ToggleEndpoint: "/t", OptionsEndpoint: "/o", ClearEndpoint: "/c",
		Source: Source{Static: []Option{{Value: "a", Label: "A"}}},
	}
	state := State{Options: cfg.Source.Static}

	var buf bytes.Buffer
	require.NoError(t, Combobox(cfg, state).Render(context.Background(), &buf))
	html := buf.String()

	assert.Contains(t, html, `data-combobox`)
	assert.Contains(t, html, `x-data="{isOpen:false, openedWithKeyboard:false, focusIndex:-1}"`)
	assert.Contains(t, html, `id="status-body"`, "body partial embedded inside shell")
	assert.NotContains(t, html, `allOptions`, "no data state in Alpine")
	assert.NotContains(t, html, `selectedValues`, "no data state in Alpine")
	assert.NotContains(t, html, `filteredOptions`, "no data state in Alpine")
}

func TestOptionsList_ClientMode_NoHXPost(t *testing.T) {
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		Source: Source{Static: []Option{{Value: "running", Label: "Running"}}},
	}
	state := State{Options: cfg.Source.Static}

	var buf bytes.Buffer
	require.NoError(t, optionsList(cfg, state).Render(context.Background(), &buf))
	html := buf.String()

	assert.NotContains(t, html, `hx-post`, "client-mode <li> must not trigger server toggle")
	assert.NotContains(t, html, `hx-vals`)
	assert.NotContains(t, html, `hx-target`)
	assert.Contains(t, html, `data-value="running"`)
	assert.Contains(t, html, `data-combobox-option`, "marker for client-side listener")
}

func TestCombobox_ClientMode_RootHasMarker(t *testing.T) {
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		Source: Source{Static: []Option{{Value: "a", Label: "A"}}},
	}
	state := State{Options: cfg.Source.Static}

	var buf bytes.Buffer
	require.NoError(t, Combobox(cfg, state).Render(context.Background(), &buf))
	html := buf.String()

	assert.Contains(t, html, `data-combobox-mode="client"`)
	assert.Contains(t, html, `hx-disinherit="hx-replace-url"`)
	assert.Contains(t, html, `data-combobox-name="status"`)
	assert.Contains(t, html, `data-combobox-multi="true"`)
}

func TestCombobox_RootDeclaresCloseOnSelectBehavior(t *testing.T) {
	t.Run("single-select closes after selection", func(t *testing.T) {
		cfg := Config{
			ID: "industry", Name: "industry", Mode: ModeSingle,
			Source: Source{Static: []Option{{Value: "tech", Label: "Technology"}}},
		}
		var buf bytes.Buffer
		require.NoError(t, Combobox(cfg, State{Options: cfg.Source.Static}).Render(context.Background(), &buf))

		assert.Contains(t, buf.String(), `data-combobox-close-on-select="true"`)
	})

	t.Run("multi-select stays open after selection", func(t *testing.T) {
		cfg := Config{
			ID: "skills", Name: "skills", Mode: ModeMultiple,
			Source: Source{Static: []Option{{Value: "go", Label: "Go"}}},
		}
		var buf bytes.Buffer
		require.NoError(t, Combobox(cfg, State{Options: cfg.Source.Static}).Render(context.Background(), &buf))

		assert.Contains(t, buf.String(), `data-combobox-close-on-select="false"`)
	})
}

func TestCombobox_TriggerUsesActiveSurfaceVocabulary(t *testing.T) {
	for _, tt := range []struct {
		name  string
		state State
	}{
		{name: "placeholder", state: State{}},
		{name: "selected", state: State{Selected: []string{"maas"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			classes := triggerBtnClass(tt.state)
			for _, want := range []string{
				"bg-surface",
				"text-on-surface-strong",
			} {
				assert.Contains(t, classes, want)
			}
		})
	}
}

func TestCombobox_ServerMode_HasHXPost(t *testing.T) {
	cfg := Config{
		ID: "team", Name: "team", Mode: ModeMultiple,
		ToggleEndpoint: "/t/toggle", OptionsEndpoint: "/t/options", ClearEndpoint: "/t/clear",
		Source: Source{LazyEndpoint: "/t/options"},
	}
	state := State{Options: []Option{{Value: "red", Label: "Red"}}}

	var buf bytes.Buffer
	require.NoError(t, Combobox(cfg, state).Render(context.Background(), &buf))
	html := buf.String()

	assert.Contains(t, html, `data-combobox-mode="server"`)
	assert.Contains(t, html, `hx-post="/t/toggle"`)
}

func TestCombobox_ClientMode_EmitsScript(t *testing.T) {
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		Source: Source{Static: []Option{{Value: "a", Label: "A"}}},
	}
	var buf bytes.Buffer
	require.NoError(t, Combobox(cfg, State{Options: cfg.Source.Static}).Render(context.Background(), &buf))
	html := buf.String()
	assert.Contains(t, html, `__goshtosoComboboxInit`, "listener script inlined")
}

func TestCombobox_ServerMode_NoScript(t *testing.T) {
	cfg := Config{
		ID: "team", Name: "team", Mode: ModeMultiple,
		ToggleEndpoint: "/t/toggle", OptionsEndpoint: "/t/options", ClearEndpoint: "/t/clear",
		Source: Source{LazyEndpoint: "/t/options"},
	}
	var buf bytes.Buffer
	require.NoError(t, Combobox(cfg, State{Options: []Option{{Value: "red"}}}).Render(context.Background(), &buf))
	assert.NotContains(t, buf.String(), `__goshtosoComboboxInit`)
}

func TestOptionsList_DisabledOption_OmitsHXPost(t *testing.T) {
	// Server mode with a disabled option: no hx-post on that li.
	cfg := Config{
		ID: "team", Name: "team", Mode: ModeMultiple,
		ToggleEndpoint: "/t/toggle", OptionsEndpoint: "/t/options", ClearEndpoint: "/t/clear",
		Source: Source{LazyEndpoint: "/t/options"},
	}
	state := State{Options: []Option{
		{Value: "red", Label: "Red"},
		{Value: "blue", Label: "Blue", Disabled: true},
	}}

	var buf bytes.Buffer
	require.NoError(t, optionsList(cfg, state).Render(context.Background(), &buf))
	html := buf.String()

	// Red must have hx-post; blue must not.
	// Split by data-value to check each <li> independently.
	redIdx := strings.Index(html, `data-value="red"`)
	blueIdx := strings.Index(html, `data-value="blue"`)
	require.Greater(t, redIdx, -1)
	require.Greater(t, blueIdx, -1)

	// The hx-post attribute lives within the red <li>; check it appears before blueIdx
	// and not between blueIdx and end.
	assert.Contains(t, html[redIdx:blueIdx], `hx-post`, "enabled option has hx-post")
	assert.NotContains(t, html[blueIdx:], `hx-post`, "disabled option omits hx-post")
	assert.Contains(t, html[blueIdx:], `aria-disabled="true"`)
}

func TestBodyOOB_EmitsSwapAttribute(t *testing.T) {
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		Source: Source{Static: []Option{{Value: "a", Label: "A"}}},
	}
	state := State{Options: cfg.Source.Static, Selected: []string{"a"}}

	var buf bytes.Buffer
	require.NoError(t, bodyOOB(cfg, state).Render(context.Background(), &buf))
	html := buf.String()

	assert.Contains(t, html, `id="status-body"`)
	assert.Contains(t, html, `hx-swap-oob="outerHTML"`)
	assert.Contains(t, html, `value="a"`, "hidden input preserved")
}

func TestTriggerLabelOOB_EmitsSiblingSwap(t *testing.T) {
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		Placeholder: "All Statuses",
		Source:      Source{Static: []Option{{Value: "a", Label: "Alpha"}}},
	}
	state := State{Options: cfg.Source.Static, Selected: []string{"a"}}

	var buf bytes.Buffer
	require.NoError(t, triggerLabelOOB(cfg, state).Render(context.Background(), &buf))
	html := buf.String()

	assert.Contains(t, html, `id="status-trigger-label"`)
	assert.Contains(t, html, `hx-swap-oob="true"`)
	assert.Contains(t, html, `Alpha`)
}

func TestCombobox_OuterTriggerLabel_HasID(t *testing.T) {
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		Source: Source{Static: []Option{{Value: "a", Label: "A"}}},
	}
	var buf bytes.Buffer
	require.NoError(t, Combobox(cfg, State{Options: cfg.Source.Static}).Render(context.Background(), &buf))
	assert.Contains(t, buf.String(), `id="status-trigger-label"`)
}

func TestBody_ClientMode_ClearAllAlwaysRendered(t *testing.T) {
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		Source:         Source{Static: []Option{{Value: "a", Label: "A"}}},
		EnableClearAll: true,
	}

	t.Run("empty selection renders hidden button", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, body(cfg, State{Options: cfg.Source.Static}).Render(context.Background(), &buf))
		html := buf.String()
		assert.Contains(t, html, `data-combobox-clear`)
		assert.Contains(t, html, `hidden`, "button has hidden attribute when empty")
	})

	t.Run("non-empty selection renders visible button", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, body(cfg, State{Options: cfg.Source.Static, Selected: []string{"a"}}).Render(context.Background(), &buf))
		html := buf.String()
		assert.Contains(t, html, `data-combobox-clear`)
		// No `hidden` attribute within the button tag.
		clearIdx := strings.Index(html, `data-combobox-clear`)
		require.Greater(t, clearIdx, -1)
		end := strings.Index(html[clearIdx:], ">") + clearIdx
		buttonTag := html[clearIdx:end]
		assert.NotContains(t, buttonTag, ` hidden`, "visible button has no hidden attribute")
	})
}

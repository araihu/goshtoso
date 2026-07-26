package combobox

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/a-h/templ"
)

// Mode selects single- or multi-select behavior.
type Mode int

const (
	ModeSingle Mode = iota
	ModeMultiple
)

// Source is the option source. Exactly one field is set.
type Source struct {
	Static       []Option // in-memory options; rendered on first paint; used for client-side no-op scenarios
	LazyEndpoint string   // relative URL; component renders "Loading…" until first hx-get completes
}

// Option is one selectable item.
type Option struct {
	Value    string
	Label    string
	Disabled bool
}

// Config holds the combobox configuration. ID must be globally unique per page.
type Config struct {
	ID              string
	Name            string
	Label           string
	Placeholder     string
	Mode            Mode
	Source          Source
	Selected        []string
	EnableSearch    bool
	EnableClearAll  bool
	Required        bool
	DependsOn       []string
	ToggleEndpoint  string
	OptionsEndpoint string
	ClearEndpoint   string
	RootClass       string
	// TriggerAttrs appends non-conflicting HTML attributes to the combobox trigger.
	TriggerAttrs templ.Attributes
	Disabled     bool
}

func (c Config) triggerAttributes() templ.Attributes {
	attrs := make(templ.Attributes, len(c.TriggerAttrs)+1)
	for key, value := range c.TriggerAttrs {
		attrs[key] = value
	}
	if c.Required {
		if _, exists := attrs["aria-required"]; !exists {
			attrs["aria-required"] = "true"
		}
	}
	return attrs
}

// State is the per-request render state.
type State struct {
	Options  []Option
	Selected []string
	Search   string
	Deps     map[string]string
}

// OptionsProvider is the server-side source of truth for options.
// search is the user-typed filter (empty for first paint); deps contains
// the values of cfg.DependsOn observed on this request.
type OptionsProvider func(ctx context.Context, search string, deps map[string]string) ([]Option, error)

// IsClientMode reports whether this combobox toggles locally without a server
// round-trip. Client mode covers any source that isn't a LazyEndpoint: the
// caller is expected to render all options into the initial DOM (either via
// Source.Static or by populating State.Options). Configs with neither
// Source.Static nor Source.LazyEndpoint are rejected by Validate, so in
// practice "not lazy" implies "options are in DOM at paint".
func (c Config) isClientMode() bool {
	return c.Source.LazyEndpoint == ""
}

// Validate returns an error if the Config is not usable.
func (c Config) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("combobox: Config.ID is required")
	}
	if c.Name == "" {
		return fmt.Errorf("combobox: Config.Name is required")
	}
	if c.Source.LazyEndpoint != "" && len(c.Source.Static) != 0 {
		return fmt.Errorf("combobox: Config.Source cannot have both Static and LazyEndpoint")
	}
	if c.Source.LazyEndpoint == "" && len(c.Source.Static) == 0 {
		return fmt.Errorf("combobox: Config.Source must have Static or LazyEndpoint set")
	}
	// Server-mode (lazy or cascading) requires endpoints.
	if !c.isClientMode() {
		if c.ToggleEndpoint == "" {
			return fmt.Errorf("combobox: Config.ToggleEndpoint is required for server mode")
		}
		if c.OptionsEndpoint == "" {
			return fmt.Errorf("combobox: Config.OptionsEndpoint is required for server mode")
		}
		if c.ClearEndpoint == "" {
			return fmt.Errorf("combobox: Config.ClearEndpoint is required for server mode")
		}
	}
	return nil
}

// InitialState returns the first render state for static/client-mode comboboxes.
func (c Config) InitialState() State {
	return State{Options: c.Source.Static, Selected: c.Selected}
}

// IsSelected reports whether value is in the selected set.
func (s State) isSelected(value string) bool {
	return slices.Contains(s.Selected, value)
}

// DepsSelector returns the CSS selector for dependency hidden inputs,
// used by hx-include. Example for DependsOn=["provider","zone"]:
//
//	[name='provider'],[name='zone']
func (c Config) depsSelector() string {
	if len(c.DependsOn) == 0 {
		return ""
	}
	parts := make([]string, len(c.DependsOn))
	for i, name := range c.DependsOn {
		parts[i] = "[name='" + name + "']"
	}
	return strings.Join(parts, ",")
}

// HXIncludeSelector returns the full hx-include selector for toggle/search requests:
// own hidden inputs plus all dependency hidden inputs.
func (c Config) hxIncludeSelector() string {
	base := "closest [data-combobox] input[type=hidden]"
	if deps := c.depsSelector(); deps != "" {
		return base + "," + deps
	}
	return base
}

package splitbutton

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/button"
)

func primaryButtonOptions(cfg Config, action Action) []button.Option {
	options := []button.Option{
		button.WithTone(action.tone(cfg)),
		button.WithSize(cfg.size()),
		button.WithRootClass("rounded-r-none border-r-0"),
	}
	if action.ID != "" {
		options = append(options, button.WithID(action.ID))
	}
	if action.Disabled {
		options = append(options, button.Disabled())
	}
	if action.HTMX != nil && !action.Disabled {
		options = append(options, button.WithHTMX(&button.HTMXConfig{
			Get:     action.HTMX.Get,
			Post:    action.HTMX.Post,
			Target:  action.HTMX.Target,
			Swap:    action.HTMX.Swap,
			Trigger: action.HTMX.Trigger,
			Vals:    action.HTMX.Vals,
			Confirm: action.HTMX.Confirm,
		}))
	}
	attrs := templ.Attributes{}
	if action.Tooltip != "" {
		attrs["title"] = action.Tooltip
	}
	if action.OnClick != "" && !action.Disabled {
		attrs["x-on:click"] = action.OnClick
	}
	if len(attrs) > 0 {
		options = append(options, button.WithAttrs(attrs))
	}
	return options
}

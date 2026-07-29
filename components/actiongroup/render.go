package actiongroup

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/button"
)

func actionButtonOptions(action Action, primary bool) []button.Option {
	options := []button.Option{
		button.WithTone(action.buttonTone(primary)),
		button.WithSize(button.SizeSmall),
		button.WithType("button"),
	}
	if action.ID != "" {
		options = append(options, button.WithID(action.ID))
	}
	if action.Disabled {
		options = append(options, button.Disabled())
	}
	attrs := actionAttrs(action)
	if len(attrs) > 0 {
		options = append(options, button.WithAttrs(attrs))
	}
	return options
}

func actionAttrs(action Action) templ.Attributes {
	attrs := templ.Attributes{}
	if action.Tooltip != "" {
		attrs["title"] = action.Tooltip
	}
	if action.OnClick != "" && !action.Disabled {
		attrs["x-on:click"] = action.OnClick
	}
	return attrs
}

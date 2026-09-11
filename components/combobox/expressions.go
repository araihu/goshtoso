package combobox

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).Combobox
	cfg.Placeholder = expressions.Text(cfg.Placeholder, text.Placeholder)
	return cfg
}

func expressionTriggerLabel(ctx context.Context, cfg Config, state State) string {
	if cfg.Mode != ModeSingle && len(state.Selected) > 1 {
		return expressions.From(ctx).Combobox.SelectedLabel(len(state.Selected))
	}
	return triggerLabelText(cfg.withExpressions(ctx), state)
}
func selectedLabels(ctx context.Context, state State) string {
	count := len(state.Options) + len(state.Selected)
	return expressions.CountLabels(count, expressions.From(ctx).Combobox.SelectedLabel)
}

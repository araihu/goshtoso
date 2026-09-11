package toast

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
)

func (cfg MessageConfig) withExpressions(ctx context.Context) MessageConfig {
	text := expressions.From(ctx).Toast
	cfg.DismissLabel = expressions.Text(cfg.DismissLabel, text.DismissLabel)
	return cfg
}

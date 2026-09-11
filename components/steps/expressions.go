package steps

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).Steps
	cfg.AriaLabel = expressions.Text(cfg.AriaLabel, text.Label)
	return cfg
}

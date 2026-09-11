package selectfield

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).Select
	cfg.Placeholder = expressions.Text(cfg.Placeholder, text.Placeholder)
	return cfg
}

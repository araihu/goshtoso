package selectfield

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"github.com/araihu/goshtoso/internal/expressionutil"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).Select
	cfg.Placeholder = expressionutil.Text(cfg.Placeholder, text.Placeholder)
	return cfg
}

package steps

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"github.com/araihu/goshtoso/internal/expressionutil"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).Steps
	cfg.AriaLabel = expressionutil.Text(cfg.AriaLabel, text.AriaLabel)
	return cfg
}

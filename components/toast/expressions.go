package toast

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"github.com/araihu/goshtoso/internal/expressionutil"
)

func (cfg MessageConfig) withExpressions(ctx context.Context) MessageConfig {
	text := expressions.From(ctx).Toast
	cfg.DismissLabel = expressionutil.Text(cfg.DismissLabel, text.DismissLabel)
	return cfg
}

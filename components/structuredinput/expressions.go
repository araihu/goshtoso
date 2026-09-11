package structuredinput

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"github.com/araihu/goshtoso/internal/expressionutil"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).StructuredInput
	cfg.AddActionLabel = expressionutil.Text(cfg.AddActionLabel, text.AddLabel)
	return cfg
}

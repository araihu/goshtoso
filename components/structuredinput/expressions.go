package structuredinput

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).StructuredInput
	cfg.AddActionLabel = expressions.Text(cfg.AddActionLabel, text.AddLabel)
	return cfg
}

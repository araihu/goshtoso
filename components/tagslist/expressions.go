package tagslist

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).TagsList
	cfg.AddActionLabel = expressions.Text(cfg.AddActionLabel, text.AddLabel)
	cfg.Placeholder = expressions.Text(cfg.Placeholder, text.Placeholder)
	return cfg
}

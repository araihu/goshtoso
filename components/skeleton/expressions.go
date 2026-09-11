package skeleton

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"strings"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).Skeleton
	cfg.Label = expressions.Text(strings.TrimSpace(cfg.Label), text.Label)
	return cfg
}

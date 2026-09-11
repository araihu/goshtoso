package actiongroup

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"github.com/araihu/goshtoso/internal/expressionutil"
	"strings"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).ActionGroup
	cfg.Label = expressionutil.Text(strings.TrimSpace(cfg.Label), text.AriaLabel)
	cfg.OverflowLabel = expressionutil.Text(strings.TrimSpace(cfg.OverflowLabel), text.OverflowAriaLabel)
	return cfg
}

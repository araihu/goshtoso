package splitbutton

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"strings"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).SplitButton
	cfg.MenuLabel = expressions.Text(strings.TrimSpace(cfg.MenuLabel), text.MenuLabel)
	return cfg
}

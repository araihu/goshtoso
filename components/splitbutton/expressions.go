package splitbutton

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"github.com/araihu/goshtoso/internal/expressionutil"
	"strings"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).SplitButton
	cfg.MenuLabel = expressionutil.Text(strings.TrimSpace(cfg.MenuLabel), text.MenuAriaLabel)
	return cfg
}

package search

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"github.com/araihu/goshtoso/internal/expressionutil"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).Search
	cfg.Label = expressionutil.Text(cfg.Label, text.Label)
	cfg.Placeholder = expressionutil.Text(cfg.Placeholder, text.Placeholder)
	cfg.ShortcutText = expressionutil.Text(cfg.ShortcutText, text.ShortcutText)
	cfg.EscapeText = expressionutil.Text(cfg.EscapeText, text.EscapeText)
	cfg.EmptyText = expressionutil.Text(cfg.EmptyText, text.EmptyText)
	return cfg
}

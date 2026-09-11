package search

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
)

func (cfg Config) withExpressions(ctx context.Context) Config {
	text := expressions.From(ctx).Search
	cfg.Label = expressions.Text(cfg.Label, text.Label)
	cfg.Placeholder = expressions.Text(cfg.Placeholder, text.Placeholder)
	cfg.ShortcutText = expressions.Text(cfg.ShortcutText, text.ShortcutText)
	cfg.EscapeText = expressions.Text(cfg.EscapeText, text.EscapeText)
	cfg.EmptyText = expressions.Text(cfg.EmptyText, text.EmptyText)
	return cfg
}

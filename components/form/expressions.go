package form

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
)

func (cfg FlipSectionConfig) withExpressions(ctx context.Context) FlipSectionConfig {
	text := expressions.From(ctx).Form
	cfg.EditLabel = expressions.Text(cfg.EditLabel, text.EditLabel)
	cfg.DoneLabel = expressions.Text(cfg.DoneLabel, text.DoneLabel)
	return cfg
}

func (cfg FormErrorsConfig) withExpressions(ctx context.Context) FormErrorsConfig {
	text := expressions.From(ctx).Form
	cfg.Title = expressions.Text(cfg.Title, text.ErrorTitle)
	return cfg
}

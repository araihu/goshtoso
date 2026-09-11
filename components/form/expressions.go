package form

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"github.com/araihu/goshtoso/internal/expressionutil"
)

func (cfg FlipSectionConfig) withExpressions(ctx context.Context) FlipSectionConfig {
	text := expressions.From(ctx).Form
	cfg.EditLabel = expressionutil.Text(cfg.EditLabel, text.EditLabel)
	cfg.DoneLabel = expressionutil.Text(cfg.DoneLabel, text.DoneLabel)
	return cfg
}

func (cfg FormErrorsConfig) withExpressions(ctx context.Context) FormErrorsConfig {
	text := expressions.From(ctx).Form
	cfg.Title = expressionutil.Text(cfg.Title, text.ErrorTitle)
	return cfg
}

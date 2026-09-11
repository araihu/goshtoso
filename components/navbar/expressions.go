package navbar

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"github.com/araihu/goshtoso/internal/expressionutil"
)

func (cfg SecondaryConfig) withExpressions(ctx context.Context) SecondaryConfig {
	text := expressions.From(ctx).Navbar
	cfg.AriaLabel = expressionutil.Text(normalizeLandmarkLabel(cfg.AriaLabel), text.SecondaryAriaLabel)
	return cfg
}

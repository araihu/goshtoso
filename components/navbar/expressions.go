package navbar

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
)

func (cfg SecondaryConfig) withExpressions(ctx context.Context) SecondaryConfig {
	text := expressions.From(ctx).Navbar
	cfg.AriaLabel = expressions.Text(normalizeLandmarkLabel(cfg.AriaLabel), text.SecondaryLabel)
	return cfg
}

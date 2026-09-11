package sidebar

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"github.com/araihu/goshtoso/internal/expressionutil"
	"strings"
)

func (cfg OverlayConfig) withExpressions(ctx context.Context) OverlayConfig {
	text := expressions.From(ctx).Sidebar
	cfg.TriggerLabel = expressionutil.Text(strings.TrimSpace(cfg.TriggerLabel), text.TriggerLabel)
	return cfg
}

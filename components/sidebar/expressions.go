package sidebar

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"strings"
)

func (cfg OverlayConfig) withExpressions(ctx context.Context) OverlayConfig {
	text := expressions.From(ctx).Sidebar
	cfg.TriggerLabel = expressions.Text(strings.TrimSpace(cfg.TriggerLabel), text.TriggerLabel)
	return cfg
}

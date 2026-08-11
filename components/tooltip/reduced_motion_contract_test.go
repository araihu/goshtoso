package tooltip

import (
	"strings"
	"testing"
)

func TestTooltipVariantsRenderReducedMotionTransitions(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
	}{
		{name: "hover"},
		{name: "rich", options: []Option{WithDescription("More detail")}},
		{name: "click", options: []Option{WithActivation(ActivationClick)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rendered := renderTooltip(t, "motion-"+test.name, "Motion tooltip", test.options...)
			if !strings.Contains(rendered, "motion-reduce:transition-none") {
				t.Fatalf("%s Tooltip lacks reduced-motion transition override: %s", test.name, rendered)
			}
		})
	}
}

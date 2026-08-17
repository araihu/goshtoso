package tooltip

import (
	"os"
	"strings"
	"testing"
)

func TestTooltipPanelsUseSemanticRadius(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
	}{
		{name: "default"},
		{name: "rich", options: []Option{WithDescription("More detail")}},
		{name: "click", options: []Option{WithActivation(ActivationClick)}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rendered := renderTooltip(t, "semantic-radius-"+test.name, "Semantic radius", test.options...)
			panel := renderedTooltipPanelTag(t, rendered)
			if !strings.Contains(panel, "rounded-radius") {
				t.Errorf("%s Tooltip panel must use semantic rounded-radius; panel: %s", test.name, panel)
			}
			if strings.Contains(panel, "rounded-sm") {
				t.Errorf("%s Tooltip panel must not use fixed rounded-sm; panel: %s", test.name, panel)
			}
		})
	}
}

func TestTooltipSemanticRadiusProvenance(t *testing.T) {
	authoredBytes, err := os.ReadFile("tooltip.templ")
	if err != nil {
		t.Fatalf("read authored Tooltip template: %v", err)
	}
	var panels []string
	for line := range strings.SplitSeq(string(authoredBytes), "\n") {
		if strings.Contains(line, `role="tooltip"`) {
			panels = append(panels, strings.TrimSpace(line))
		}
	}
	if len(panels) != 3 {
		t.Fatalf("authored Tooltip panel count = %d, want 3; panels: %v", len(panels), panels)
	}
	for index, panel := range panels {
		if !strings.Contains(panel, "rounded-radius") || strings.Contains(panel, "rounded-sm") {
			t.Errorf("authored Tooltip panel %d must use only semantic rounded-radius; panel: %s", index+1, panel)
		}
	}

	themeBytes, err := os.ReadFile("../../all-themes.css")
	if err != nil {
		t.Fatalf("read authoritative theme source: %v", err)
	}
	minimal := themeBlock(t, string(themeBytes), "minimal")
	if !strings.Contains(minimal, "--radius-radius: var(--radius-none);") {
		t.Fatalf("Minimal must map semantic component radius to radius-none; block:\n%s", minimal)
	}
}

func renderedTooltipPanelTag(t *testing.T, rendered string) string {
	t.Helper()

	roleIndex := strings.Index(rendered, ` role="tooltip"`)
	if roleIndex < 0 {
		t.Fatalf("rendered Tooltip has no panel role: %s", rendered)
	}
	start := strings.LastIndex(rendered[:roleIndex], "<div")
	if start < 0 {
		t.Fatalf("rendered Tooltip has no panel element: %s", rendered)
	}
	return rendered[start:roleIndex]
}

func themeBlock(t *testing.T, source, theme string) string {
	t.Helper()

	marker := "[data-theme=" + theme + "]"
	start := strings.Index(source, marker)
	if start < 0 {
		t.Fatalf("authoritative theme source missing %s", marker)
	}
	block := source[start:]
	if next := strings.Index(block[len(marker):], "[data-theme="); next >= 0 {
		block = block[:len(marker)+next]
	}
	return block
}

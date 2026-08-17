package tooltip

import (
	"html"
	"strings"
	"testing"
)

func TestPersistentTooltipRendersLiveStateContract(t *testing.T) {
	rendered := html.UnescapeString(renderTooltip(t,
		"persistent-tip",
		"Persistent help",
		WithActivation(ActivationClick),
		WithTriggerLabel("Show help"),
	))

	for _, want := range []string{
		`aria-describedby="persistent-tip"`,
		`aria-controls="persistent-tip"`,
		`x-bind:aria-expanded="showTooltip"`,
		`x-on:keydown.escape`,
		`role="tooltip"`,
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("persistent Tooltip missing rendered live-state contract %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, `aria-haspopup`) {
		t.Errorf("role=tooltip must not be represented as an ARIA haspopup target:\n%s", rendered)
	}
}

func TestHoverTooltipKeepsDescriptionWithoutPersistentState(t *testing.T) {
	rendered := html.UnescapeString(renderTooltip(t,
		"hover-tip",
		"Hover help",
		WithTriggerLabel("More information"),
	))

	if !strings.Contains(rendered, `aria-describedby="hover-tip"`) {
		t.Fatalf("hover Tooltip lost descriptive relationship:\n%s", rendered)
	}
	for _, forbidden := range []string{`aria-expanded`, `aria-controls`, `aria-haspopup`} {
		if strings.Contains(rendered, forbidden) {
			t.Errorf("hover Tooltip must not render persistent state %q:\n%s", forbidden, rendered)
		}
	}
}

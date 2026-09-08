package tooltip

import (
	"context"
	"strings"
	"testing"
)

func TestHelpUsesNamedCompactButton(t *testing.T) {
	var output strings.Builder
	if err := Help("pin-help", "How pinning works", "Check to pin. Uncheck to remove.").Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	for _, want := range []string{`aria-label="How pinning works"`, `type="button"`, `size-4`, `data-tooltip-activation="click"`, `data-tooltip-content-id="pin-help"`, `role="tooltip"`, `Check to pin. Uncheck to remove.`} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q", want)
		}
	}
}

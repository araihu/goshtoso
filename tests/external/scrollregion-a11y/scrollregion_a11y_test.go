package scrollregioncontract_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/scrollregion"
)

func renderScrollRegion(t *testing.T, component templ.Component) string {
	t.Helper()
	var output bytes.Buffer
	if err := component.Render(context.Background(), &output); err != nil {
		t.Fatalf("render public ScrollRegion: %v", err)
	}
	return output.String()
}

func TestExternalConsumerRendersNamedScrollableRegion(t *testing.T) {
	html := renderScrollRegion(t, scrollregion.Named(scrollregion.Config{}, scrollregion.AccessibleName{Label: "Activity history"}))
	for _, want := range []string{
		`data-goshtoso-scroll-viewport`,
		`tabindex="0"`,
		`role="region"`,
		`aria-label="Activity history"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("public ScrollRegion markup missing %q:\n%s", want, html)
		}
	}
}

func TestExternalConsumerCanReferenceVisibleRegionLabel(t *testing.T) {
	html := renderScrollRegion(t, scrollregion.Named(scrollregion.Config{}, scrollregion.AccessibleName{
		Label:      "Fallback",
		LabelledBy: "activity-history-heading",
	}))
	if !strings.Contains(html, `aria-labelledby="activity-history-heading"`) {
		t.Fatalf("public ScrollRegion omitted aria-labelledby:\n%s", html)
	}
	if strings.Contains(html, `aria-label="Fallback"`) {
		t.Fatalf("public ScrollRegion rendered competing aria-label:\n%s", html)
	}
}

func TestExternalConsumerCanUseLabelledShortcut(t *testing.T) {
	html := renderScrollRegion(t, scrollregion.Labelled(scrollregion.Config{}, "activity-history-heading"))
	if !strings.Contains(html, `aria-labelledby="activity-history-heading"`) {
		t.Fatalf("public Labelled shortcut omitted aria-labelledby:\n%s", html)
	}
	if strings.Contains(html, `aria-label=`) {
		t.Fatalf("public Labelled shortcut rendered competing aria-label:\n%s", html)
	}
}

func TestExternalConsumerKeepsBaseUnkeyedConfigLiteral(t *testing.T) {
	legacy := scrollregion.Config{nil, "legacy-root", "legacy-viewport", false}
	html := renderScrollRegion(t, scrollregion.ScrollRegion(legacy))
	if !strings.Contains(html, `role="group"`) || !strings.Contains(html, `aria-label="Scrollable content"`) {
		t.Fatalf("legacy positional Config literal lost default accessible named group: %s", html)
	}
}

func TestExternalConsumerLegacyInstancesAvoidDuplicateLandmarks(t *testing.T) {
	first := renderScrollRegion(t, scrollregion.ScrollRegion(scrollregion.Config{}))
	second := renderScrollRegion(t, scrollregion.ScrollRegion(scrollregion.Config{}))
	combined := first + second
	if strings.Count(combined, `role="region"`) != 0 || strings.Count(combined, `role="group"`) != 2 || strings.Count(combined, `aria-label="Scrollable content"`) != 2 {
		t.Fatalf("two legacy instances must remain named focusable groups, not duplicate region landmarks: %s", combined)
	}
}

func TestExternalConsumerWideContentKeepsHorizontalAccessUnlessExplicitlyOverridden(t *testing.T) {
	wide := renderScrollRegion(t, scrollregion.Named(scrollregion.Config{
		ViewportClass: "border",
		Content: templ.ComponentFunc(func(_ context.Context, writer io.Writer) error {
			_, err := writer.Write([]byte(`<pre>unbreakable-consumer-token-012345678901234567890123456789</pre>`))
			return err
		}),
	}, scrollregion.AccessibleName{Label: "Consumer data"}))
	if !strings.Contains(wide, `overflow-x-auto`) || !strings.Contains(wide, "unbreakable-consumer-token") {
		t.Fatalf("generic consumer content lost its horizontal access path: %s", wide)
	}
	hidden := renderScrollRegion(t, scrollregion.Named(scrollregion.Config{ViewportClass: "overflow-x-hidden"}, scrollregion.AccessibleName{Label: "Constrained consumer data"}))
	if strings.Contains(hidden, `overflow-x-auto`) || !strings.Contains(hidden, `overflow-x-hidden`) {
		t.Fatalf("explicit consumer horizontal override conflicts with the default: %s", hidden)
	}
}

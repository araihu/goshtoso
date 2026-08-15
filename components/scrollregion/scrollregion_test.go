package scrollregion

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
)

func TestScrollRegionIdentity(t *testing.T) {
	component, ok := ScrollRegion(Config{}).(components.Component)
	if !ok {
		t.Fatalf("ScrollRegion return type %T does not implement components.Component", ScrollRegion(Config{}))
	}
	if got := component.Kind(); got != components.KindScrollRegion {
		t.Fatalf("Kind() = %q; want %q", got, components.KindScrollRegion)
	}
}

// TestConfigRetainsBasePositionalLiteralCompatibility protects consumers that
// compiled against the original four-field public Config shape.
func TestConfigRetainsBasePositionalLiteralCompatibility(t *testing.T) {
	legacy := Config{nil, "legacy-root", "legacy-viewport", false}
	if legacy.RootClass != "legacy-root" || legacy.ViewportClass != "legacy-viewport" || legacy.DisableIndicators {
		t.Fatalf("legacy positional Config literal no longer preserves base field order: %#v", legacy)
	}
}

func TestScrollRegionViewportDeclaresReachableBothAxisOverflow(t *testing.T) {
	classes := Config{}.viewportClasses()
	if !strings.Contains(classes, "overflow-y-auto") {
		t.Fatalf("scroll viewport must retain vertical scrolling: %q", classes)
	}
	if !strings.Contains(classes, "overflow-x-auto") {
		t.Fatalf("scroll viewport must retain horizontal access for generic content: %q", classes)
	}
	if got := (Config{ViewportClass: "overflow-x-hidden"}).viewportClasses(); strings.Contains(got, "overflow-x-auto") || !strings.Contains(got, "overflow-x-hidden") {
		t.Fatalf("consumer horizontal override must not conflict with default utility: %q", got)
	}
}

func TestScrollRegionRendersIndependentBoundaryCues(t *testing.T) {
	var output bytes.Buffer
	err := ScrollRegion(Config{
		RootClass:     "test-root",
		ViewportClass: "test-viewport",
		Content: templ.ComponentFunc(func(_ context.Context, writer io.Writer) error {
			_, err := io.WriteString(writer, `<p>content</p>`)
			return err
		}),
	}).Render(context.Background(), &output)
	if err != nil {
		t.Fatalf("render scroll region: %v", err)
	}

	html := output.String()
	for _, marker := range []string{
		`data-goshtoso-scroll-region`,
		`data-goshtoso-scroll-viewport`,
		`tabindex="0"`,
		`data-goshtoso-scroll-start`,
		`data-goshtoso-scroll-end`,
		`data-goshtoso-scroll-start-indicator`,
		`data-goshtoso-scroll-end-indicator`,
		`pointer-events-none`,
		`aria-hidden="true"`,
		`test-root`,
		`test-viewport`,
	} {
		if !strings.Contains(html, marker) {
			t.Errorf("rendered markup does not contain %q", marker)
		}
	}
	if got := strings.Count(html, ` hidden`); got != 2 {
		t.Fatalf("hidden indicator count = %d, want 2", got)
	}
}

func TestScrollRegionRendersNamedPublicRegion(t *testing.T) {
	render := func(t *testing.T, cfg Config) string {
		t.Helper()
		var output bytes.Buffer
		if err := ScrollRegion(cfg).Render(context.Background(), &output); err != nil {
			t.Fatalf("render ScrollRegion: %v", err)
		}
		return output.String()
	}

	t.Run("uses a stable default accessible non-landmark group", func(t *testing.T) {
		html := render(t, Config{})
		if !strings.Contains(html, `role="group"`) || strings.Contains(html, `role="region"`) {
			t.Fatalf("default ScrollRegion must not manufacture a landmark:\n%s", html)
		}
		if !strings.Contains(html, `aria-label="Scrollable content"`) {
			t.Fatalf("default ScrollRegion lacks its accessible name:\n%s", html)
		}
	})

	t.Run("uses the explicit AccessibleName", func(t *testing.T) {
		var output bytes.Buffer
		if err := Named(Config{}, AccessibleName{Label: "Activity history"}).Render(context.Background(), &output); err != nil {
			t.Fatalf("render named ScrollRegion: %v", err)
		}
		html := output.String()
		if !strings.Contains(html, `role="region"`) || !strings.Contains(html, `aria-label="Activity history"`) {
			t.Fatalf("explicit ScrollRegion label is not rendered:\n%s", html)
		}
	})

	t.Run("prefers LabelledBy over Label", func(t *testing.T) {
		var output bytes.Buffer
		if err := Named(Config{}, AccessibleName{Label: "Fallback", LabelledBy: "activity-history-heading"}).Render(context.Background(), &output); err != nil {
			t.Fatalf("render labelled ScrollRegion: %v", err)
		}
		html := output.String()
		if !strings.Contains(html, `role="region"`) || !strings.Contains(html, `aria-labelledby="activity-history-heading"`) {
			t.Fatalf("labelled ScrollRegion is not rendered:\n%s", html)
		}
		if strings.Contains(html, `aria-label="Fallback"`) {
			t.Fatalf("LabelledBy must not emit a competing aria-label:\n%s", html)
		}
	})

	t.Run("Labelled shortcut renders the visible label reference", func(t *testing.T) {
		var output bytes.Buffer
		if err := Labelled(Config{}, "activity-history-heading").Render(context.Background(), &output); err != nil {
			t.Fatalf("render Labelled ScrollRegion: %v", err)
		}
		html := output.String()
		if !strings.Contains(html, `role="region"`) || !strings.Contains(html, `aria-labelledby="activity-history-heading"`) {
			t.Fatalf("Labelled shortcut is not rendered as a named region:\n%s", html)
		}
		if strings.Contains(html, `aria-label=`) {
			t.Fatalf("Labelled shortcut must not emit a competing aria-label:\n%s", html)
		}
	})
}

func TestScrollRegionCanOmitVisualCues(t *testing.T) {
	var output bytes.Buffer
	if err := ScrollRegion(Config{DisableIndicators: true}).Render(context.Background(), &output); err != nil {
		t.Fatalf("render scroll region: %v", err)
	}

	html := output.String()
	if strings.Contains(html, `data-goshtoso-scroll-start-indicator`) || strings.Contains(html, `data-goshtoso-scroll-end-indicator`) {
		t.Fatal("visual indicators rendered while disabled")
	}
	if !strings.Contains(html, `data-goshtoso-scroll-start`) || !strings.Contains(html, `data-goshtoso-scroll-end`) {
		t.Fatal("boundary sentinels must remain available while indicators are disabled")
	}
}

package scrollregion

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

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

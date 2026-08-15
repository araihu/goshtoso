package scrollregioncompat

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/scrollregion"
)

func TestNamedUsesReleasedConfigAndNamesFocusableViewport(t *testing.T) {
	var rendered strings.Builder
	err := Named(scrollregion.Config{Content: templ.Raw("<p>Activity</p>")}, "Activity history").Render(context.Background(), &rendered)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered.String(), `data-goshtoso-scroll-viewport tabindex="0" role="region" aria-label="Activity history"`) {
		t.Fatalf("named viewport = %s", rendered.String())
	}
	if !strings.Contains(rendered.String(), "overflow-x-auto") {
		t.Fatalf("named viewport omits horizontal access: %s", rendered.String())
	}
}

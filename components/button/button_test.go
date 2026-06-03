package button

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func renderButton(t *testing.T, cfg Config, child string) string {
	t.Helper()
	var buf bytes.Buffer
	ctx := templ.WithChildren(context.Background(), templ.Raw(child))
	if err := Button(cfg).Render(ctx, &buf); err != nil {
		t.Fatalf("render Button: %v", err)
	}
	return buf.String()
}

func TestButtonLoadingTextRendersAsText(t *testing.T) {
	html := renderButton(t, Config{
		HTMX:        &HTMXConfig{Post: "/save"},
		LoadingText: "Saving...",
	}, "Save")

	if !strings.Contains(html, ">Saving...<") {
		t.Fatalf("LoadingText not rendered as text:\n%s", html)
	}
	if strings.Contains(html, `class="htmx-indicator hidden Saving..."`) {
		t.Fatalf("LoadingText was emitted as CSS class:\n%s", html)
	}
}

func TestButtonLoadingTextIgnoredWithoutHTMX(t *testing.T) {
	html := renderButton(t, Config{LoadingText: "Saving..."}, "Save")

	if strings.Contains(html, "Saving...") {
		t.Fatalf("LoadingText rendered without HTMX config:\n%s", html)
	}
	if !strings.Contains(html, "Save") {
		t.Fatalf("button child content missing:\n%s", html)
	}
}

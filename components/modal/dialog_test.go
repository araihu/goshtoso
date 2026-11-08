package modal

import (
	"bytes"
	"context"
	"github.com/a-h/templ"
	"io"
	"strings"
	"testing"
)

func TestCompactDialogKeepsAccessibleTitle(t *testing.T) {
	for _, compact := range []bool{false, true} {
		var out bytes.Buffer
		if err := Dialog(DialogConfig{ID: "search", Title: "Search services", Compact: compact, FullscreenMobile: true}).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		html := out.String()
		if !strings.Contains(html, `aria-labelledby="searchTitle"`) || !strings.Contains(html, "Search services") {
			t.Fatal("missing accessible title")
		}
		if strings.Contains(html, "Close dialog") == compact {
			t.Fatal("incorrect header mode")
		}
		if compact && strings.Contains(html, "h-dvh") {
			t.Fatal("compact must override fullscreen mode")
		}
	}
}

func TestCustomDialogHeaderKeepsLabelAndResponsiveMode(t *testing.T) {
	header := templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<header>Host actions</header>")
		return err
	})
	var out bytes.Buffer
	if err := Dialog(DialogConfig{ID: "asset", Title: "Asset details", HeaderContent: header, FullscreenMobile: true}).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	html := out.String()
	for _, want := range []string{`aria-labelledby="assetTitle"`, `id="assetTitle" class="sr-only">Asset details`, "Host actions", "h-dvh"} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %s", want)
		}
	}
	if strings.Contains(html, "Close dialog") {
		t.Fatal("custom header retained builtin actions")
	}
	out.Reset()
	if err := Dialog(DialogConfig{ID: "compact", Title: "Compact", HeaderContent: header, Compact: true}).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "Host actions") {
		t.Fatal("compact dialog must remain headerless")
	}
}

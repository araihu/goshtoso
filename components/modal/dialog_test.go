package modal

import (
	"bytes"
	"context"
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

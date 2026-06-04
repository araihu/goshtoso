package link

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func render(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	err := Link(cfg).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render Link: %v", err)
	}
	return buf.String()
}

func TestLinkZeroValue(t *testing.T) {
	html := render(t, Config{})
	if !strings.Contains(html, `href="#"`) {
		t.Fatalf("expected empty href to default to #: %s", html)
	}
	if !strings.Contains(html, "text-primary") {
		t.Fatalf("expected text link classes: %s", html)
	}
}

func TestLinkBlankTargetAddsSafeRel(t *testing.T) {
	html := render(t, Config{Href: "https://example.com", Target: "_blank"})
	if !strings.Contains(html, `rel="noopener noreferrer"`) {
		t.Fatalf("expected safe rel for blank target: %s", html)
	}
}

func TestLinkButtonStyle(t *testing.T) {
	html := render(t, Config{Style: StyleButton, Size: SizeLarge})
	if !strings.Contains(html, `role="button"`) {
		t.Fatalf("expected button role: %s", html)
	}
	if !strings.Contains(html, "bg-primary") || !strings.Contains(html, "text-base") {
		t.Fatalf("expected button classes: %s", html)
	}
}

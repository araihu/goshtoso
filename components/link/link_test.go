package link

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func render(t *testing.T, href string, options ...Option) string {
	t.Helper()
	var buf bytes.Buffer
	err := Link(href, options...).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render Link: %v", err)
	}
	return buf.String()
}

func TestLinkDefaultsToTextAppearance(t *testing.T) {
	html := render(t, "#")
	if !strings.Contains(html, `href="#"`) {
		t.Fatalf("expected required href: %s", html)
	}
	if !strings.Contains(html, "text-primary") {
		t.Fatalf("expected text link classes: %s", html)
	}
}

func TestLinkBlankTargetAddsSafeRel(t *testing.T) {
	html := render(t, "https://example.com", WithTarget("_blank"))
	if !strings.Contains(html, `rel="noopener noreferrer"`) {
		t.Fatalf("expected safe rel for blank target: %s", html)
	}
}

func TestLinkButtonStyle(t *testing.T) {
	html := render(t, "#", WithAppearance(AppearanceButton), WithSize(SizeLarge))
	if !strings.Contains(html, `role="button"`) {
		t.Fatalf("expected button role: %s", html)
	}
	if !strings.Contains(html, "bg-primary") || !strings.Contains(html, "text-base") {
		t.Fatalf("expected button classes: %s", html)
	}
}

func TestLinkButtonAppearanceDefaultsToMediumSize(t *testing.T) {
	html := render(t, "#", WithAppearance(AppearanceButton))
	if !strings.Contains(html, "text-sm") {
		t.Fatalf("expected medium button appearance classes: %s", html)
	}
}

func TestLinkRequiresHrefAndDefaultsToTextAppearance(t *testing.T) {
	var buf bytes.Buffer
	ctx := templ.WithChildren(context.Background(), templ.Raw("Docs"))
	if err := Link("/docs").Render(ctx, &buf); err != nil {
		t.Fatalf("render Link: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `href="/docs"`) {
		t.Fatalf("link did not render required href:\n%s", html)
	}
	if !strings.Contains(html, "underline-offset-2") {
		t.Fatalf("link did not default to text appearance:\n%s", html)
	}
	if !strings.Contains(html, "Docs") {
		t.Fatalf("link child content missing:\n%s", html)
	}
}

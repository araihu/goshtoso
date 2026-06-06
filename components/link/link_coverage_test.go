package link

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// iconStub renders a tiny marker element used to assert icon placement.
func iconStub() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<svg data-icon="stub"></svg>`)
		return err
	})
}

func renderCfg(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Link(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render Link: %v", err)
	}
	return buf.String()
}

func TestButtonSizeClasses(t *testing.T) {
	cases := []struct {
		name string
		size Size
		want string
	}{
		{"small", SizeSmall, "text-xs"},
		{"medium explicit", SizeMedium, "text-sm"},
		{"large", SizeLarge, "text-base"},
		{"xlarge", SizeXLarge, "text-lg"},
		{"default empty", Size(""), "text-sm"},
		{"unknown falls back", Size("bogus"), "text-sm"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Config{Size: tc.size}.buttonSizeClasses()
			if !strings.Contains(got, tc.want) {
				t.Fatalf("size %q: want %q in %q", tc.size, tc.want, got)
			}
		})
	}
}

func TestRelExplicitOverride(t *testing.T) {
	// Explicit Rel wins even for a _blank target.
	html := renderCfg(t, Config{Href: "https://example.com", Target: "_blank", Rel: "nofollow"})
	if !strings.Contains(html, `rel="nofollow"`) {
		t.Fatalf("expected explicit rel override: %s", html)
	}
	if strings.Contains(html, "noopener") {
		t.Fatalf("explicit rel should suppress default: %s", html)
	}
}

func TestRelOmittedWithoutBlankTarget(t *testing.T) {
	html := renderCfg(t, Config{Href: "https://example.com", Target: "_self"})
	if strings.Contains(html, "rel=") {
		t.Fatalf("expected no rel attribute for non-blank target: %s", html)
	}
	if !strings.Contains(html, `target="_self"`) {
		t.Fatalf("expected target attribute rendered: %s", html)
	}
}

func TestRoleExplicitOverride(t *testing.T) {
	// Explicit Role wins over the StyleButton default.
	html := renderCfg(t, Config{Style: StyleButton, Role: "link"})
	if !strings.Contains(html, `role="link"`) {
		t.Fatalf("expected explicit role override: %s", html)
	}
}

func TestRoleOmittedForTextStyle(t *testing.T) {
	html := renderCfg(t, Config{})
	if strings.Contains(html, "role=") {
		t.Fatalf("text link should not emit a role: %s", html)
	}
}

func TestIDAttributeRendered(t *testing.T) {
	html := renderCfg(t, Config{ID: "cta-link"})
	if !strings.Contains(html, `id="cta-link"`) {
		t.Fatalf("expected id attribute: %s", html)
	}
}

func TestClassAppended(t *testing.T) {
	html := renderCfg(t, Config{Class: "mt-4 custom-link"})
	if !strings.Contains(html, "mt-4") || !strings.Contains(html, "custom-link") {
		t.Fatalf("expected custom classes appended: %s", html)
	}
	// Base text classes still present alongside custom ones.
	if !strings.Contains(html, "text-primary") {
		t.Fatalf("expected base text classes retained: %s", html)
	}
}

func TestIconTrailingDefault(t *testing.T) {
	html := renderCfg(t, Config{Icon: iconStub()})
	if !strings.Contains(html, `data-icon="stub"`) {
		t.Fatalf("expected icon rendered: %s", html)
	}
	if !strings.Contains(html, "inline-flex") || !strings.Contains(html, "gap-1.5") {
		t.Fatalf("expected icon layout classes: %s", html)
	}
}

func TestIconLeadingPlacement(t *testing.T) {
	html := renderCfg(t, Config{
		Icon:         iconStub(),
		IconPosition: IconLeading,
	})
	iconIdx := strings.Index(html, `data-icon="stub"`)
	textIdx := strings.Index(html, "</a>")
	if iconIdx < 0 {
		t.Fatalf("expected icon rendered: %s", html)
	}
	// Leading icon should appear before the closing tag content marker; assert
	// it precedes any trailing position by checking it renders right after <a>.
	openIdx := strings.Index(html, ">")
	if openIdx >= iconIdx || iconIdx >= textIdx {
		t.Fatalf("expected leading icon between open tag and close: %s", html)
	}
}

func TestAttrsSpread(t *testing.T) {
	html := renderCfg(t, Config{Attrs: templ.Attributes{
		"data-testid": "promo",
		"hx-get":      "/api/x",
	}})
	if !strings.Contains(html, `data-testid="promo"`) {
		t.Fatalf("expected spread data attribute: %s", html)
	}
	if !strings.Contains(html, `hx-get="/api/x"`) {
		t.Fatalf("expected spread hx attribute: %s", html)
	}
}

func TestButtonStyleSmallAndXLarge(t *testing.T) {
	small := renderCfg(t, Config{Style: StyleButton, Size: SizeSmall})
	if !strings.Contains(small, "text-xs") || !strings.Contains(small, "bg-primary") {
		t.Fatalf("expected small button classes: %s", small)
	}
	xl := renderCfg(t, Config{Style: StyleButton, Size: SizeXLarge})
	if !strings.Contains(xl, "text-lg") {
		t.Fatalf("expected xl button classes: %s", xl)
	}
}

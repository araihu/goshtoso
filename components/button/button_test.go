package button

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func renderButton(t *testing.T, child string, options ...Option) string {
	t.Helper()
	var buf bytes.Buffer
	ctx := templ.WithChildren(context.Background(), templ.Raw(child))
	if err := Button(options...).Render(ctx, &buf); err != nil {
		t.Fatalf("render Button: %v", err)
	}
	return buf.String()
}

func TestButtonLoadingTextRendersAsText(t *testing.T) {
	html := renderButton(t, "Save",
		WithHTMX(&HTMXConfig{Post: "/save"}),
		WithLoadingText("Saving..."),
	)

	if !strings.Contains(html, ">Saving...<") {
		t.Fatalf("LoadingText not rendered as text:\n%s", html)
	}
	if strings.Contains(html, `class="htmx-indicator hidden Saving..."`) {
		t.Fatalf("LoadingText was emitted as CSS class:\n%s", html)
	}
}

func TestButtonLoadingTextSupportsAncestorHTMXForm(t *testing.T) {
	html := renderButton(t, "Save", WithLoadingText("Saving..."))

	for _, want := range []string{
		`data-goshtoso-loading`,
		`class="goshtoso-loading-content"`,
		`class="goshtoso-loading-label"`,
		`>Saving...<`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("loading contract missing %q:\n%s", want, html)
		}
	}
}

func TestButtonNilIntegrationOptionsKeepHTMXAndAlpineDisabled(t *testing.T) {
	html := renderButton(t, "Save", WithHTMX(nil), WithAlpine(nil), WithLoadingText("Saving..."))

	for _, unwanted := range []string{"hx-", "x-data", "x-on:"} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("nil integration options rendered %q:\n%s", unwanted, html)
		}
	}
	if !strings.Contains(html, "Saving...") {
		t.Fatalf("loading text should support an ancestor HTMX form:\n%s", html)
	}
}

func TestButtonDefaultsToTouchSafeTarget(t *testing.T) {
	for _, size := range []Size{SizeSmall, SizeMedium, SizeLarge, SizeXLarge} {
		classes := buttonClasses(newConfig([]Option{WithSize(size)}))
		if !strings.Contains(classes, "min-h-11") || !strings.Contains(classes, "min-w-11") {
			t.Fatalf("button size %q lacks a 44px minimum target: %s", size, classes)
		}
	}
}

func TestButtonOptionsApplyOverDefaults(t *testing.T) {
	var buf bytes.Buffer
	ctx := templ.WithChildren(context.Background(), templ.Raw("Delete"))
	err := Button(
		WithTone(ToneDanger),
		WithSize(SizeSmall),
		WithID("delete"),
	).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("render Button: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `id="delete"`) {
		t.Fatalf("button option did not set ID:\n%s", html)
	}
	if !strings.Contains(html, "bg-danger-action") {
		t.Fatalf("button option did not set danger tone:\n%s", html)
	}
	if strings.Contains(html, "danger-dark") {
		t.Fatalf("button must not depend on undefined status-dark tokens:\n%s", html)
	}
	if !strings.Contains(html, "text-xs") {
		t.Fatalf("button option did not set small size:\n%s", html)
	}
	if !strings.Contains(html, `type="button"`) {
		t.Fatalf("button default type is not button:\n%s", html)
	}
}

func TestButtonHoverDoesNotReduceWholeControlOpacity(t *testing.T) {
	html := renderButton(t, "Save")

	if strings.Contains(html, "hover:opacity-75") {
		t.Fatalf("button hover must not blend text and background toward the page surface:\n%s", html)
	}
	if !strings.Contains(html, "transition motion-reduce:transition-none") {
		t.Fatalf("button transition must honor reduced motion:\n%s", html)
	}
}

func TestButtonWithAttrsRendersNativeFormAndDataAttributes(t *testing.T) {
	html := renderButton(t, "Continue",
		WithType("submit"),
		WithAttrs(templ.Attributes{
			"name":        "action",
			"value":       "continue",
			"formaction":  "/deployments/continue",
			"data-intent": "advance",
		}),
	)

	for _, attribute := range []string{
		`name="action"`,
		`value="continue"`,
		`formaction="/deployments/continue"`,
		`data-intent="advance"`,
	} {
		if !strings.Contains(html, attribute) {
			t.Fatalf("button attribute %s missing:\n%s", attribute, html)
		}
	}
}

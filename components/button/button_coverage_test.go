package button

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// render renders Button(cfg) with the given child text into a string.
func render(t *testing.T, cfg Config, child string) string {
	t.Helper()
	var buf bytes.Buffer
	ctx := templ.WithChildren(context.Background(), templ.Raw(child))
	if err := Button(cfg).Render(ctx, &buf); err != nil {
		t.Fatalf("render Button: %v", err)
	}
	return buf.String()
}

func TestCoverageRenderDefaultButton(t *testing.T) {
	html := render(t, Config{}, "Click")
	for _, want := range []string{"<button", "class=", "Click", "</button>"} {
		if !strings.Contains(html, want) {
			t.Fatalf("default render missing %q in %s", want, html)
		}
	}
	// Empty variant hits the default branches of variantClasses (bg-primary)
	// and sizeClasses (text-sm), plus an empty outline color suffix.
	if !strings.Contains(html, "bg-primary") {
		t.Fatalf("default variant should use bg-primary:\n%s", html)
	}
	if !strings.Contains(html, "text-sm") {
		t.Fatalf("default size should use text-sm:\n%s", html)
	}
}

// TestVariantClasses exercises every variant branch (variantClasses, 40% → 100%).
func TestVariantClasses(t *testing.T) {
	cases := []struct {
		variant Variant
		want    string
	}{
		{Primary, "bg-primary"},
		{Secondary, "bg-secondary"},
		{Alternate, "bg-surface-alt"},
		{Inverse, "bg-surface-dark"},
		{Info, "bg-info"},
		{Danger, "bg-danger"},
		{Warning, "bg-warning"},
		{Success, "bg-success"},
	}
	for _, tc := range cases {
		t.Run(string(tc.variant), func(t *testing.T) {
			html := render(t, Config{Variant: tc.variant}, "x")
			if !strings.Contains(html, tc.want) {
				t.Fatalf("variant %q missing %q:\n%s", tc.variant, tc.want, html)
			}
			// The outline color suffix is derived from the variant string.
			if !strings.Contains(html, "focus-visible:outline-"+string(tc.variant)) {
				t.Fatalf("variant %q missing outline color suffix:\n%s", tc.variant, html)
			}
		})
	}
}

// TestVariantClassesUnknownDefaults covers the default branch of variantClasses
// when an unrecognized variant value is supplied.
func TestVariantClassesUnknownDefaults(t *testing.T) {
	html := render(t, Config{Variant: Variant("nope")}, "x")
	if !strings.Contains(html, "bg-primary text-on-primary border-primary") {
		t.Fatalf("unknown variant should fall back to primary classes:\n%s", html)
	}
}

// TestSizeClasses exercises every size branch (sizeClasses, 60% → 100%).
func TestSizeClasses(t *testing.T) {
	cases := []struct {
		size Size
		want string
	}{
		{SizeSmall, "text-xs"},
		{SizeMedium, "text-sm"},
		{SizeLarge, "text-base"},
		{SizeXLarge, "text-lg"},
		{Size("unknown"), "text-sm"}, // default branch
	}
	for _, tc := range cases {
		t.Run(string(tc.size), func(t *testing.T) {
			html := render(t, Config{Variant: Primary, Size: tc.size}, "x")
			if !strings.Contains(html, tc.want) {
				t.Fatalf("size %q missing %q:\n%s", tc.size, tc.want, html)
			}
		})
	}
}

// TestButtonAttributes covers the ID, Type, Disabled, and RootClass branches.
func TestButtonAttributes(t *testing.T) {
	html := render(t, Config{
		ID:        "save-btn",
		Type:      "submit",
		Disabled:  true,
		Variant:   Primary,
		RootClass: "w-full custom-x",
	}, "Save")

	for _, want := range []string{
		`id="save-btn"`,
		`type="submit"`,
		`" disabled`,
		"w-full custom-x",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("attributes render missing %q:\n%s", want, html)
		}
	}
}

// TestButtonNoIDOmitsAttribute confirms an empty ID emits no id attribute.
func TestButtonNoIDOmitsAttribute(t *testing.T) {
	html := render(t, Config{Variant: Primary}, "x")
	if strings.Contains(html, "id=") {
		t.Fatalf("empty ID should not emit id attribute:\n%s", html)
	}
	// The base class list contains "disabled:opacity-75" utilities, so match the
	// boolean attribute form (quote-space-disabled) instead of the bare word.
	if strings.Contains(html, `" disabled`) {
		t.Fatalf("non-disabled button should not emit disabled attribute:\n%s", html)
	}
}

// TestAlpineAttributes covers every AlpineConfig branch in the templ.
func TestAlpineAttributes(t *testing.T) {
	html := render(t, Config{
		Variant: Primary,
		Alpine: &AlpineConfig{
			OnClick:      "open = true",
			Show:         "open",
			BindDisabled: "busy",
			Transition:   true,
			Data:         "{ open: false }",
		},
	}, "x")

	for _, want := range []string{
		`x-on:click="open = true"`,
		`x-show="open"`,
		`:disabled="busy"`,
		"x-transition",
		`x-data="{ open: false }"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("alpine render missing %q:\n%s", want, html)
		}
	}
}

// TestAlpineEmptyOmitsAttributes confirms a zero-value AlpineConfig pointer
// emits no Alpine directives (hasAlpine true, but all fields empty).
func TestAlpineEmptyOmitsAttributes(t *testing.T) {
	html := render(t, Config{Variant: Primary, Alpine: &AlpineConfig{}}, "x")
	for _, unwanted := range []string{"x-on:click", "x-show", ":disabled", "x-transition", "x-data"} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("empty AlpineConfig should not emit %q:\n%s", unwanted, html)
		}
	}
}

// TestHTMXAttributes covers every HTMXConfig branch in the templ.
func TestHTMXAttributes(t *testing.T) {
	html := render(t, Config{
		Variant: Primary,
		HTMX: &HTMXConfig{
			Get:       "/g",
			Post:      "/p",
			Put:       "/u",
			Delete:    "/d",
			Patch:     "/pa",
			Target:    "#out",
			Swap:      "innerHTML",
			Trigger:   "click",
			Indicator: "#spin",
			PushURL:   true,
			Confirm:   "Sure?",
			Vals:      `{"a":1}`,
		},
	}, "x")

	for _, want := range []string{
		`hx-get="/g"`,
		`hx-post="/p"`,
		`hx-put="/u"`,
		`hx-delete="/d"`,
		`hx-patch="/pa"`,
		`hx-target="#out"`,
		`hx-swap="innerHTML"`,
		`hx-trigger="click"`,
		`hx-indicator="#spin"`,
		`hx-push-url="true"`,
		`hx-confirm="Sure?"`,
		"hx-vals",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("htmx render missing %q:\n%s", want, html)
		}
	}
}

// TestHTMXEmptyOmitsAttributes confirms a zero-value HTMXConfig pointer emits
// no hx-* attributes (hasHTMX true, but all fields empty/false).
func TestHTMXEmptyOmitsAttributes(t *testing.T) {
	html := render(t, Config{Variant: Primary, HTMX: &HTMXConfig{}}, "x")
	for _, unwanted := range []string{"hx-get", "hx-post", "hx-put", "hx-delete", "hx-patch", "hx-push-url"} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("empty HTMXConfig should not emit %q:\n%s", unwanted, html)
		}
	}
}

// TestLoadingTextRendersIndicatorSpans covers the LoadingText + HTMX branch
// that wraps children in an htmx-indicator-content span and adds a hidden
// indicator span with the loading text.
func TestLoadingTextRendersIndicatorSpans(t *testing.T) {
	html := render(t, Config{
		Variant:     Primary,
		HTMX:        &HTMXConfig{Post: "/save"},
		LoadingText: "Saving...",
	}, "Save")

	for _, want := range []string{
		`<span class="htmx-indicator-content">Save</span>`,
		`<span class="htmx-indicator hidden">Saving...</span>`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("loading text render missing %q:\n%s", want, html)
		}
	}
}

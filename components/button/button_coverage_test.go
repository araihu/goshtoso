package button

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// render renders Button(options...) with the given child text into a string.
func render(t *testing.T, child string, options ...Option) string {
	t.Helper()
	var buf bytes.Buffer
	ctx := templ.WithChildren(context.Background(), templ.Raw(child))
	if err := Button(options...).Render(ctx, &buf); err != nil {
		t.Fatalf("render Button: %v", err)
	}
	return buf.String()
}

func TestCoverageRenderDefaultButton(t *testing.T) {
	html := render(t, "Click")
	for _, want := range []string{"<button", "class=", "Click", "</button>"} {
		if !strings.Contains(html, want) {
			t.Fatalf("default render missing %q in %s", want, html)
		}
	}
	if !strings.Contains(html, "bg-primary") {
		t.Fatalf("default tone should use bg-primary:\n%s", html)
	}
	if !strings.Contains(html, "text-sm") {
		t.Fatalf("default size should use text-sm:\n%s", html)
	}
}

func TestToneClasses(t *testing.T) {
	cases := []struct {
		tone        Tone
		want        string
		wantOutline string
	}{
		{TonePrimary, "bg-primary", "focus-visible:outline-primary dark:focus-visible:outline-primary-dark"},
		{ToneSecondary, "bg-secondary", "focus-visible:outline-secondary dark:focus-visible:outline-secondary-dark"},
		{ToneAlternate, "bg-surface-alt", "focus-visible:outline-on-surface-strong dark:focus-visible:outline-on-surface-dark-strong"},
		{ToneInverse, "bg-surface-dark", "focus-visible:outline-on-surface-strong dark:focus-visible:outline-on-surface-dark-strong"},
		{ToneInfo, "bg-info-action", "focus-visible:outline-info-action dark:focus-visible:outline-info-action-dark"},
		{ToneDanger, "bg-danger-action", "focus-visible:outline-danger-action dark:focus-visible:outline-danger-action-dark"},
		{ToneWarning, "bg-warning-action", "focus-visible:outline-warning-action dark:focus-visible:outline-warning-action-dark"},
		{ToneSuccess, "bg-success-action", "focus-visible:outline-success-action dark:focus-visible:outline-success-action-dark"},
	}
	for _, tc := range cases {
		t.Run(string(tc.tone), func(t *testing.T) {
			html := render(t, "x", WithTone(tc.tone))
			if !strings.Contains(html, tc.want) {
				t.Fatalf("tone %q missing %q:\n%s", tc.tone, tc.want, html)
			}
			if !strings.Contains(html, tc.wantOutline) {
				t.Fatalf("tone %q missing outline color suffix:\n%s", tc.tone, html)
			}
		})
	}
}

func TestToneClassesUnknownDefaults(t *testing.T) {
	html := render(t, "x", WithTone(Tone("nope")))
	if !strings.Contains(html, "bg-primary text-on-primary border-primary") {
		t.Fatalf("unknown tone should fall back to primary classes:\n%s", html)
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
			html := render(t, "x", WithSize(tc.size))
			if !strings.Contains(html, tc.want) {
				t.Fatalf("size %q missing %q:\n%s", tc.size, tc.want, html)
			}
		})
	}
}

// TestButtonAttributes covers the ID, Type, Disabled, and RootClass branches.
func TestButtonAttributes(t *testing.T) {
	html := render(t, "Save",
		WithID("save-btn"),
		WithType("submit"),
		Disabled(),
		WithRootClass("w-full custom-x"),
	)

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
	html := render(t, "x")
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
	html := render(t, "x",
		WithAlpine(&AlpineConfig{
			OnClick:      "open = true",
			Show:         "open",
			BindDisabled: "busy",
			Transition:   true,
			Data:         "{ open: false }",
		}),
	)

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
	html := render(t, "x", WithAlpine(&AlpineConfig{}))
	for _, unwanted := range []string{"x-on:click", "x-show", ":disabled", "x-transition", "x-data"} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("empty AlpineConfig should not emit %q:\n%s", unwanted, html)
		}
	}
}

// TestHTMXAttributes covers every HTMXConfig branch in the templ.
func TestHTMXAttributes(t *testing.T) {
	html := render(t, "x",
		WithHTMX(&HTMXConfig{
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
		}),
	)

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
	html := render(t, "x", WithHTMX(&HTMXConfig{}))
	for _, unwanted := range []string{"hx-get", "hx-post", "hx-put", "hx-delete", "hx-patch", "hx-push-url"} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("empty HTMXConfig should not emit %q:\n%s", unwanted, html)
		}
	}
}

// TestLoadingTextRendersIndicatorSpans covers the LoadingText + HTMX branch.
func TestLoadingTextRendersIndicatorSpans(t *testing.T) {
	html := render(t, "Save",
		WithHTMX(&HTMXConfig{Post: "/save"}),
		WithLoadingText("Saving..."),
	)

	for _, want := range []string{
		`data-goshtoso-loading`,
		`hx-disabled-elt="this"`,
		`<span class="goshtoso-loading-content">Save</span>`,
		`<span class="goshtoso-loading-label" aria-live="polite">Saving...</span>`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("loading text render missing %q:\n%s", want, html)
		}
	}
}

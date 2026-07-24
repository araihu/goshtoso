package spinner

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func renderSpinner(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Spinner(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render spinner %+v: %v", cfg, err)
	}
	return buf.String()
}

func TestCoverageRenderDefaultSpinner(t *testing.T) {
	html := renderSpinner(t, Config{})

	for _, want := range []string{
		`class=`,
		`aria-hidden="true"`,
		`viewBox="0 0 24 24"`,
		`motion-safe:animate-spin`,
		"size-5",          // default size
		"fill-on-surface", // default variant fill
		"dark:fill-on-surface-dark",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("default render missing %q in %s", want, html)
		}
	}
}

func TestCoverageSizeClasses(t *testing.T) {
	cases := []struct {
		size Size
		want string
	}{
		{SizeSM, "size-4"},
		{SizeMD, "size-5"},
		{SizeLG, "size-8"},
		{SizeXL, "size-12"},
		{Size("bogus"), "size-5"}, // unknown falls back to default
		{"", "size-5"},            // zero value falls back to default
	}

	for _, tc := range cases {
		t.Run(string(tc.size), func(t *testing.T) {
			if got := (Config{Size: tc.size}).sizeClasses(); got != tc.want {
				t.Fatalf("sizeClasses(%q) = %q, want %q", tc.size, got, tc.want)
			}
			if html := renderSpinner(t, Config{Size: tc.size}); !strings.Contains(html, tc.want) {
				t.Fatalf("render size %q missing class %q in %s", tc.size, tc.want, html)
			}
		})
	}
}

func TestCoverageFillClasses(t *testing.T) {
	cases := []struct {
		variant Tone
		want    string
	}{
		{ToneDefault, "fill-on-surface dark:fill-on-surface-dark"},
		{TonePrimary, "fill-primary dark:fill-primary-dark"},
		{ToneSecondary, "fill-secondary dark:fill-secondary-dark"},
		{ToneInfo, "fill-info dark:fill-info"},
		{ToneSuccess, "fill-success dark:fill-success"},
		{ToneWarning, "fill-warning dark:fill-warning"},
		{ToneDanger, "fill-danger dark:fill-danger"},
		{Tone("bogus"), "fill-on-surface dark:fill-on-surface-dark"}, // unknown falls back
		{"", "fill-on-surface dark:fill-on-surface-dark"},            // zero value falls back
	}

	for _, tc := range cases {
		t.Run(string(tc.variant), func(t *testing.T) {
			if got := (Config{Tone: tc.variant}).fillClasses(); got != tc.want {
				t.Fatalf("fillClasses(%q) = %q, want %q", tc.variant, got, tc.want)
			}
			if html := renderSpinner(t, Config{Tone: tc.variant}); !strings.Contains(html, tc.want) {
				t.Fatalf("render variant %q missing fill %q in %s", tc.variant, tc.want, html)
			}
		})
	}
}

func TestCoverageRootClassAppended(t *testing.T) {
	html := renderSpinner(t, Config{RootClass: "custom-extra-class"})
	if !strings.Contains(html, "custom-extra-class") {
		t.Fatalf("expected RootClass appended, got %s", html)
	}
	// RootClass should sit alongside size, fill, and animation classes.
	for _, want := range []string{"size-5", "fill-on-surface", "motion-safe:animate-spin", "custom-extra-class"} {
		if !strings.Contains(html, want) {
			t.Fatalf("render with RootClass missing %q in %s", want, html)
		}
	}
}

func TestCoverageRootClassEmptyNotAppended(t *testing.T) {
	html := renderSpinner(t, Config{})
	// An empty RootClass must not introduce a trailing space artifact in the class list.
	if strings.Contains(html, "animate-spin  ") || strings.Contains(html, `class=" `) {
		t.Fatalf("empty RootClass should not add stray spaces: %s", html)
	}
}

func TestCoverageVariantAndSizeCombined(t *testing.T) {
	html := renderSpinner(t, Config{Tone: ToneDanger, Size: SizeXL, RootClass: "mx-auto"})
	for _, want := range []string{"size-12", "fill-danger", "dark:fill-danger", "motion-safe:animate-spin", "mx-auto"} {
		if !strings.Contains(html, want) {
			t.Fatalf("combined render missing %q in %s", want, html)
		}
	}
}

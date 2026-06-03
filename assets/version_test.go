package assets

import (
	"strings"
	"testing"
)

func TestTailwindVersion(t *testing.T) {
	got := TailwindVersion()
	if got != "4.3.0" {
		t.Fatalf("TailwindVersion() = %q, want %q", got, "4.3.0")
	}
	if strings.HasPrefix(got, "v") {
		t.Fatalf("TailwindVersion() must not include a leading v: %q", got)
	}
}

func TestThemeCSS(t *testing.T) {
	b, err := ThemeCSS()
	if err != nil {
		t.Fatalf("ThemeCSS() error: %v", err)
	}
	s := string(b)
	for _, want := range []string{"@custom-variant dark", "[data-theme=minimal]", "@theme"} {
		if !strings.Contains(s, want) {
			t.Errorf("ThemeCSS() missing %q", want)
		}
	}
	if strings.Contains(s, `@import "tailwindcss"`) {
		t.Error("ThemeCSS() must not contain the tailwind import")
	}
}

func TestVendorVersions(t *testing.T) {
	cases := map[string]string{
		"Alpine":     AlpineVersion(),
		"HTMX":       HTMXVersion(),
		"HTMXExtSSE": HTMXExtSSEVersion(),
		"HTMXExtWS":  HTMXExtWSVersion(),
	}
	want := map[string]string{
		"Alpine": "3.14.9", "HTMX": "2.0.8", "HTMXExtSSE": "2.2.3", "HTMXExtWS": "2.0.3",
	}
	for k, got := range cases {
		if got != want[k] {
			t.Errorf("%sVersion() = %q, want %q", k, got, want[k])
		}
	}
}

func TestVendorFilesEmbedded(t *testing.T) {
	for _, p := range []string{
		"js/vendor/alpinejs/3.14.9/alpine.min.js",
		"js/vendor/alpinejs-collapse/3.14.9/alpine-collapse.min.js",
		"js/vendor/alpinejs-focus/3.14.9/alpine-focus.min.js",
		"js/vendor/htmx.org/2.0.8/htmx.min.js",
		"js/vendor/htmx-ext-sse/2.2.3/htmx-ext-sse.min.js",
		"js/vendor/htmx-ext-ws/2.0.3/htmx-ext-ws.js",
	} {
		if _, err := files.ReadFile(p); err != nil {
			t.Errorf("embedded file missing: %s: %v", p, err)
		}
	}
}

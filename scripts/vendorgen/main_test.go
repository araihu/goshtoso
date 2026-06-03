package main

import "testing"

func TestURLPath(t *testing.T) {
	got := urlPath("alpinejs", dep{Version: "3.14.9", File: "alpine.min.js"})
	want := "/assets/js/vendor/alpinejs/3.14.9/alpine.min.js"
	if got != want {
		t.Fatalf("urlPath = %q, want %q", got, want)
	}
}

func TestConstNameMapComplete(t *testing.T) {
	// Every module in the canonical list must have a Go constant name.
	for _, k := range []string{
		"alpinejs", "alpinejs-collapse", "alpinejs-focus",
		"htmx.org", "htmx-ext-sse", "htmx-ext-ws",
	} {
		if constName[k] == "" {
			t.Errorf("missing constName for %q", k)
		}
	}
}

func TestRenderDeterministic(t *testing.T) {
	deps := map[string]dep{
		"htmx.org": {Version: "2.0.8", File: "htmx.min.js"},
		"alpinejs": {Version: "3.14.9", File: "alpine.min.js"},
	}
	a := render(deps)
	b := render(deps)
	if a != b {
		t.Fatal("render not deterministic")
	}
	// Constants emitted in sorted-by-Go-name order: AlpineJSURL before HTMXURL.
	if ai, hi := indexOf(a, "AlpineJSURL"), indexOf(a, "HTMXURL"); ai < 0 || hi < 0 || ai > hi {
		t.Fatalf("ordering wrong: AlpineJSURL@%d HTMXURL@%d", ai, hi)
	}
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

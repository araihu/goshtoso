package vendorgen

import "testing"

func TestURLPath(t *testing.T) {
	got := urlPath("alpinejs", dep{Version: "3.14.9", File: "alpine.min.js"})
	want := "/assets/js/runtime/alpinejs/3.14.9/alpine.min.js"
	if got != want {
		t.Fatalf("urlPath = %q, want %q", got, want)
	}
	if indexOf(got, "/vendor/") >= 0 {
		t.Fatalf("urlPath must avoid /vendor/ because Go module zips omit vendor dirs: %q", got)
	}
}

func TestConstNameMapComplete(t *testing.T) {
	// Every module in the canonical list must have a Go constant name.
	for _, k := range []string{
		"alpinejs", "alpinejs-collapse", "alpinejs-focus", "alpinejs-mask",
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

func TestRenderIncludesExpandedCDNURLs(t *testing.T) {
	got := render(map[string]dep{
		"alpinejs": {
			Version:   "3.14.9",
			File:      "alpine.min.js",
			URL:       "https://unpkg.com/alpinejs@{v}/dist/cdn.min.js",
			Integrity: "sha384-example",
		},
	})

	for _, want := range []string{
		`AlpineJSURL       = "/assets/js/runtime/alpinejs/3.14.9/alpine.min.js"`,
		`AlpineJSCDNURL    = "https://unpkg.com/alpinejs@3.14.9/dist/cdn.min.js"`,
		`AlpineJSIntegrity = "sha384-example"`,
	} {
		if indexOf(got, want) < 0 {
			t.Fatalf("generated constants missing %q:\n%s", want, got)
		}
	}
}

func TestIntegrityForBytesUsesSHA384SRIFormat(t *testing.T) {
	got := integrityForBytes([]byte("abc"))
	want := "sha384-ywB1P0WjXou1oD1pmsZQBycsMqsO3tFjGotgWkP/W+2AhgcroefMI1i67KE0yCWn"
	if got != want {
		t.Fatalf("integrityForBytes = %q, want %q", got, want)
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

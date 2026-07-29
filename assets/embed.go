// Package assets provides embedded static files and their public runtime
// contract for Goshtoso components. Use Handler to serve them at /assets/,
// DefaultRuntimeManifest to inspect the ordered dependency/fallback set, and
// GoshtosoVersion to identify the exact library module linked into a consumer.
//
// Usage:
//
//	mux := http.NewServeMux()
//	mux.Handle("/assets/", assets.Handler())
//
// This serves:
//   - /assets/styles.css — compiled Tailwind CSS with all theme definitions
//   - /assets/js/runtime/alpinejs/3.14.9/alpine.min.js — Alpine.js
//   - /assets/js/runtime/htmx.org/2.0.8/htmx.min.js — HTMX
//   - /assets/js/runtime/htmx-ext-sse/2.2.3/htmx-ext-sse.min.js — HTMX SSE extension
//   - /assets/js/runtime/alpinejs-collapse/3.14.9/alpine-collapse.min.js — Alpine collapse plugin
//   - /assets/js/runtime/alpinejs-focus/3.14.9/alpine-focus.min.js — Alpine focus plugin
//   - /assets/js/runtime/alpinejs-mask/3.14.9/alpine-mask.min.js — Alpine mask plugin
//   - /assets/js/dependency-loader.js — ordered CDN loader with local fallback
//   - /assets/js/goshtoso.min.js — minified reusable component behavior
//   - /assets/js/combobox.js — standalone compatibility build
//   - /assets/js/action-group.js — responsive ActionGroup measurement
//   - vendored JS versions are pinned in js/runtime/versions.json (see AlpineVersion()/HTMXVersion())
//   - /assets/js/darkmode.js — Alpine dark mode store
//   - /assets/images/* — brand artwork (mascot, logos)
package assets

import (
	"embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed styles.css goshtoso-theme.css tailwind.version js/*.js js/runtime images
var files embed.FS

// Handler returns an http.Handler that serves the embedded Goshtoso assets
// (styles.css, js/, images/) — the generated library files head.Dependencies() links.
//
// Mount it at /assets/ WITHOUT wrapping it in your own StripPrefix — the
// handler already strips the /assets/ prefix internally:
//
//	http.Handle("/assets/", assets.Handler())             // correct
//	http.Handle("/assets/", http.StripPrefix("/assets/", assets.Handler())) // WRONG: double-strip → 404
func Handler() http.Handler {
	return http.StripPrefix("/assets/", http.FileServer(http.FS(files)))
}

// StylesCSS returns the compiled Goshtoso Tailwind CSS.
// Use this to extract the CSS to disk for Tailwind's @import directive.
func StylesCSS() ([]byte, error) {
	return files.ReadFile("styles.css")
}

// TailwindVersion returns the Tailwind CSS version that styles.css and
// goshtoso-theme.css were built with — the single-source pin in
// assets/tailwind.version (e.g. "4.3.0", no leading "v"). Match your own
// Tailwind build to this when compiling Goshtoso's theme source yourself.
func TailwindVersion() string {
	b, err := files.ReadFile("tailwind.version")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// ThemeCSS returns the Goshtoso theme SOURCE (tokens, @custom-variant, the 13
// [data-theme] blocks, base + utility layers) for importing into your OWN
// Tailwind v4 build. Unlike StylesCSS (compiled output you serve directly),
// this is source your Tailwind compiles. Pair it with a @source pointing at
// Goshtoso's components dir (see `goshtoso -source-path`).
func ThemeCSS() ([]byte, error) {
	return files.ReadFile("goshtoso-theme.css")
}

// vendorDep mirrors one entry in js/runtime/versions.json.
type vendorDep struct {
	Version string `json:"version"`
	File    string `json:"file"`
	URL     string `json:"url"`
}

// vendorVersion returns the pinned version of a vendored module from the
// embedded manifest (js/runtime/versions.json), or "" if absent.
func vendorVersion(module string) string {
	b, err := files.ReadFile("js/runtime/versions.json")
	if err != nil {
		return ""
	}
	var m map[string]vendorDep
	if json.Unmarshal(b, &m) != nil {
		return ""
	}
	return m[module].Version
}

// AlpineVersion returns the Alpine.js version Goshtoso vendors (core, collapse,
// focus, and mask share this version). Pinned in js/runtime/versions.json.
func AlpineVersion() string { return vendorVersion("alpinejs") }

// HTMXVersion returns the vendored HTMX version (js/runtime/versions.json).
func HTMXVersion() string { return vendorVersion("htmx.org") }

// HTMXExtSSEVersion returns the vendored htmx-ext-sse version.
func HTMXExtSSEVersion() string { return vendorVersion("htmx-ext-sse") }

// HTMXExtWSVersion returns the vendored htmx-ext-ws version.
func HTMXExtWSVersion() string { return vendorVersion("htmx-ext-ws") }

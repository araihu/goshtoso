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
// This serves compiled CSS, the Muamba-backed JavaScript runtime, first-party
// JavaScript, brand images, and icons. Use DefaultRuntimeManifest to inspect
// exact versions and local URLs.
package assets

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed styles.css goshtoso-theme.css tailwind.version js/*.js images icons
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
	fileServer := http.FileServer(http.FS(files))
	return http.StripPrefix("/assets/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if ref, ok := muambaHTTPFiles[request.URL.Path]; ok {
			serveMuambaFile(writer, request, ref)
			return
		}
		if strings.HasPrefix(request.URL.Path, "icons/") {
			writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		fileServer.ServeHTTP(writer, request)
	}))
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

// AlpineVersion returns the Alpine.js version Goshtoso vendors (core, collapse,
// focus, and mask share this version). Generated from Muamba and the runtime overlay.
func AlpineVersion() string { return runtimeVersionAlpineJS }

// HTMXVersion returns the vendored HTMX version from the canonical manifest.
func HTMXVersion() string { return runtimeVersionHTMX }

// HTMXExtSSEVersion returns the vendored htmx-ext-sse version.
func HTMXExtSSEVersion() string { return runtimeVersionHTMXExtSSE }

// HTMXExtWSVersion returns the vendored htmx-ext-ws version.
func HTMXExtWSVersion() string { return runtimeVersionHTMXExtWS }

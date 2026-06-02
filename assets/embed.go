// Package assets provides embedded static files (CSS, JS, fonts) for Goshtoso components.
// Use Handler() to serve them at /assets/ in your HTTP server.
//
// Usage:
//
//	mux := http.NewServeMux()
//	mux.Handle("/assets/", assets.Handler())
//
// This serves:
//   - /assets/styles.css — compiled Tailwind CSS with all theme definitions
//   - /assets/js/vendor/alpine.min.js — Alpine.js
//   - /assets/js/vendor/htmx.min.js — HTMX
//   - /assets/js/vendor/htmx-ext-sse.min.js — HTMX SSE extension
//   - /assets/js/vendor/alpine-collapse.min.js — Alpine collapse plugin
//   - /assets/js/vendor/alpine-focus.min.js — Alpine focus plugin
//   - /assets/js/darkmode.js — Alpine dark mode store
//   - /assets/fonts/* — TOTVS brand font files
//   - /assets/images/* — brand artwork (mascot, logos)
package assets

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed styles.css goshtoso-theme.css tailwind.version js fonts images
var files embed.FS

// Handler returns an http.Handler that serves the embedded Goshtoso assets
// (styles.css, js/, fonts/, images/) — the files head.Dependencies() links.
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

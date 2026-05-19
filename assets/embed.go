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
//   - /assets/js/vendor/alpine-collapse.min.js — Alpine collapse plugin
//   - /assets/js/vendor/alpine-focus.min.js — Alpine focus plugin
//   - /assets/js/darkmode.js — Alpine dark mode store
//   - /assets/fonts/* — TOTVS brand font files
//   - /assets/images/* — brand artwork (mascot, logos)
package assets

import (
	"embed"
	"net/http"
)

//go:embed styles.css js fonts images
var files embed.FS

// Handler returns an http.Handler that serves the embedded Goshtoso assets.
// Mount it at /assets/ in your router.
func Handler() http.Handler {
	return http.StripPrefix("/assets/", http.FileServer(http.FS(files)))
}

// StylesCSS returns the compiled Goshtoso Tailwind CSS.
// Use this to extract the CSS to disk for Tailwind's @import directive.
func StylesCSS() ([]byte, error) {
	return files.ReadFile("styles.css")
}

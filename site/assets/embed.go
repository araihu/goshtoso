// Package siteassets owns browser assets used only by the Goshtoso demo site.
// It is intentionally separate from the publishable library's assets package:
// head.Dependencies never references this handler or its URLs.
package siteassets

import (
	"embed"
	"net/http"
)

const (
	// DemoBundleURL is the site-owned Alpine provider bundle served by Handler.
	DemoBundleURL = "/site-assets/js/goshtoso-demo.min.js"
)

//go:embed js/goshtoso-demo.min.js
var files embed.FS

// Handler serves generated demo-site assets at /site-assets/. Authored sources
// remain outside the embed pattern and cannot be requested from this handler.
func Handler() http.Handler {
	return http.StripPrefix("/site-assets/", http.FileServer(http.FS(files)))
}

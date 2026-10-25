//go:build !js

package server

import (
	"net/http"
	"path/filepath"
	"strings"

	shellassets "github.com/araihu/goshtoso-app-shells/componentdocshell/assets"
	consoleassets "github.com/araihu/goshtoso-app-shells/consoleshell/assets"
	landingassets "github.com/araihu/goshtoso-app-shells/landingshell/assets"
	chartassets "github.com/araihu/goshtoso-charts/assets"
	libraryassets "github.com/araihu/goshtoso/assets"
	siteassets "github.com/araihu/goshtoso/site/assets"
)

func (s *Server) setupAssetRoutes() {
	s.mux.Handle("/consoleshell/assets/", consoleassets.Handler())
	s.mux.Handle("/landingshell/assets/", landingassets.Handler())
	// Compiled assets (CSS, JS)
	assetsDir := filepath.Join(s.projectRoot, "assets")
	assetsHandler := http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir)))
	s.mux.Handle("/assets/", libraryassets.WithCacheControl(assetsHandler))
	// Versioned dependencies must match the linked module, including standalone builds.
	s.mux.Handle("/assets/js/runtime/", libraryassets.Handler())
	bootstrapSprite := filepath.Join(s.projectRoot, "site", "internal", "demoicons", "bootstrapicons", "sprite.svg")
	s.mux.HandleFunc("GET /assets/icons/bootstrapicons/sprite.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		http.ServeFile(w, r, bootstrapSprite)
	})
	heroiconsSidebarSprite := filepath.Join(s.projectRoot, "site", "internal", "demoicons", "heroicons", "sprite.svg")
	s.mux.HandleFunc("GET /assets/icons/heroicons/sprite.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		http.ServeFile(w, r, heroiconsSidebarSprite)
	})
	s.mux.Handle("/site-assets/", libraryassets.WithCacheControl(siteassets.Handler()))
	s.mux.Handle("/componentdocshell/assets/", libraryassets.WithCacheControl(shellassets.Handler()))
	s.mux.Handle("GET "+chartassets.Prefix, withImmutableCache(chartassets.Handler()))

	// Favicons are referenced at root paths from <head>, so they are served from
	// the root mux, but the files live alongside the other assets on disk.
	// Exact-match patterns; no path traversal possible.
	for _, name := range []string{
		"favicon.ico", "favicon.svg", "favicon-96x96.png", "apple-touch-icon.png",
		"site.webmanifest", "web-app-manifest-192x192.png", "web-app-manifest-512x512.png",
	} {
		routeName := name
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(routeName, ".webmanifest") {
				w.Header().Set("Content-Type", "application/manifest+json")
			}
			http.ServeFile(w, r, filepath.Join(assetsDir, routeName))
		})
		s.mux.Handle("/"+routeName, libraryassets.WithCacheControl(handler))
	}
}

func (s *Server) registerDogImages() {
	dogsDir := filepath.Join(s.projectRoot, "examples", "getting-started", "assets", "dogs")
	s.mux.Handle("/dog-images/", http.StripPrefix("/dog-images/", http.FileServer(http.Dir(dogsDir))))
}

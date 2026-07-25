package server

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/site/internal/pages/demo/components"
)

func (s *Server) registerGettingStartedRoutes() {
	dogsDir := filepath.Join(s.projectRoot, "examples", "getting-started", "assets", "dogs")
	s.mux.Handle("/dog-images/", http.StripPrefix("/dog-images/", http.FileServer(http.Dir(dogsDir))))
	s.mux.HandleFunc("/api/getting-started/breeds", s.handleGettingStartedBreeds)
}

func (s *Server) handleGettingStartedBreeds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cfg := components.GettingStartedPreviewConfigFromQuery(r.URL.Query())
	for _, row := range cfg.Rows {
		_ = table.TableRow(cfg, row).Render(r.Context(), w)
	}

	_, _ = fmt.Fprintf(w, `<template><thead id="%s" hx-swap-oob="outerHTML" class="%s">`,
		resolvedTableID(cfg)+"-thead", tableHeadClasses)
	_ = table.TableHeadContent(cfg).Render(r.Context(), w)
	_, _ = fmt.Fprintf(w, `</thead></template>`)

	if cfg.Pagination != nil {
		_, _ = fmt.Fprintf(w, `<div id="%s" hx-swap-oob="true" class="flex items-center justify-between border-t border-outline px-4 py-3 dark:border-outline-dark">`, resolvedTableID(cfg)+"-pagination")
		_, _ = fmt.Fprintf(w, `<div class="text-sm text-on-surface/70 dark:text-on-surface-dark/70">Page %d of %d</div>`, cfg.Pagination.CurrentPage, cfg.Pagination.TotalPages)
		_ = table.TablePaginationNav(cfg).Render(r.Context(), w)
		_, _ = fmt.Fprintf(w, `</div>`)
	}
}

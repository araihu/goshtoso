package server

import (
	"net/http"
	"strconv"

	"github.com/araihu/goshtoso/site/internal/examples/deployments"
	deploymentspage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/deployments"
)

func (s *Server) renderDeploymentConsole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Enter the standalone shell instead of swapping it into the docs layout.
	if r.Method == http.MethodGet && r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", r.URL.RequestURI())
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	state := deployments.Decode(r.FormValue("state"))
	screen := r.FormValue("view")
	if screen == "" {
		screen = "overview"
	}
	v := deploymentspage.View{State: state, Screen: screen, Query: r.FormValue("q"), Created: r.FormValue("created") == "1"}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		switch r.PostForm.Get("action") {
		case "create":
			v.Service = r.PostForm.Get("service")
			v.Version = r.PostForm.Get("version")
			v.Environment = r.PostForm.Get("environment")
			v.Errors = deployments.Validate(v.Service, v.Version, v.Environment)
			if len(v.Errors) > 0 {
				v.Screen = "new"
				status = http.StatusUnprocessableEntity
			} else {
				id := state.Create(v.Service, v.Version, v.Environment)
				http.Redirect(w, r, state.URL("detail", id)+"&created=1", http.StatusSeeOther)
				return
			}
		case "approve":
			id, _ := strconv.Atoi(r.PostForm.Get("id"))
			if _, ok := state.Find(id); !ok {
				http.NotFound(w, r)
				return
			}
			state.Approve(id)
			http.Redirect(w, r, state.URL("detail", id), http.StatusSeeOther)
			return
		default:
			http.Error(w, "Unknown action", http.StatusBadRequest)
			return
		}
	}
	switch v.Screen {
	case "overview", "list", "new":
	case "detail":
		id, _ := strconv.Atoi(r.FormValue("id"))
		record, ok := state.Find(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		v.Record = record
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = deploymentspage.Layout(v).Render(r.Context(), w)
}

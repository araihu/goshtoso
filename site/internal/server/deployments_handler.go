package server

import (
	"net/http"
	"strconv"

	"github.com/araihu/goshtoso/site/internal/examples/deployments"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	deploymentspage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/deployments"
	demoregistry "github.com/araihu/goshtoso/site/internal/pages/demo/registry"
)

func (s *Server) renderDeploymentConsole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	partial := r.Header.Get("HX-Request-Type") == "partial"
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
		var handled bool
		status, handled = submitDeployment(w, r, &v, partial)
		if handled {
			return
		}
	}

	switch v.Screen {
	case "overview", "list", "new":
	case "detail":
		if v.Record.ID != 0 {
			break
		}
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
	content := deploymentspage.App(v)
	meta := demoregistry.MetaForKey("examples/deployments")
	if partial {
		_ = demo.ComponentDocsFragment(meta, "deployments", content, storageAllowed(r)).Render(r.Context(), w)
		return
	}
	_ = demo.ComponentDocsLayout(meta, "deployments", content, storageAllowed(r)).Render(r.Context(), w)
}

// submitDeployment preserves normal form redirects while returning the updated
// view and its bookmarkable URL for fragment submissions.
func submitDeployment(w http.ResponseWriter, r *http.Request, v *deploymentspage.View, partial bool) (int, bool) {
	var id int
	switch r.PostForm.Get("action") {
	case "create":
		v.Service = r.PostForm.Get("service")
		v.Version = r.PostForm.Get("version")
		v.Environment = r.PostForm.Get("environment")
		v.Errors = deployments.Validate(v.Service, v.Version, v.Environment)
		if len(v.Errors) > 0 {
			v.Screen = "new"
			return http.StatusUnprocessableEntity, false
		}
		id = v.State.Create(v.Service, v.Version, v.Environment)
		v.Created = true
	case "approve":
		id, _ = strconv.Atoi(r.PostForm.Get("id"))
		if _, ok := v.State.Find(id); !ok {
			http.NotFound(w, r)
			return http.StatusNotFound, true
		}
		v.State.Approve(id)
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
		return http.StatusBadRequest, true
	}
	v.Screen = "detail"
	v.Record, _ = v.State.Find(id)
	target := v.State.URL("detail", id)
	if v.Created {
		target += "&created=1"
	}
	if !partial {
		http.Redirect(w, r, target, http.StatusSeeOther)
		return http.StatusSeeOther, true
	}
	w.Header().Set("HX-Push-Url", target)
	return http.StatusOK, false
}

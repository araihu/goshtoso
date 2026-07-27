package server

import (
	"net/http"

	"github.com/araihu/goshtoso/components/toast"
	"github.com/araihu/goshtoso/site/internal/examples/wizard"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	"github.com/araihu/goshtoso/site/internal/pages/demo/components"
	"github.com/araihu/goshtoso/site/internal/pages/demo/examples"
)

// persistWizard saves the wizard state to a cookie only when the visitor has not
// opted out of browser storage (audit BLOCKER B1: never persist unconditionally).
func persistWizard(r *http.Request, w http.ResponseWriter, st wizard.WizardState) {
	if storageAllowed(r) {
		wizard.SetCookie(w, st)
	}
}

// registerWizardRoutes wires all /api/examples/wizard/* endpoints.
func (s *Server) registerWizardRoutes() {
	s.mux.HandleFunc("/api/examples/wizard/next", s.handleWizardNext)
	s.mux.HandleFunc("/api/examples/wizard/back", s.handleWizardBack)
	s.mux.HandleFunc("/api/examples/wizard/confirm", s.handleWizardConfirm)
	s.mux.HandleFunc("/api/examples/wizard/reset", s.handleWizardReset)
}

// renderWizardPage is the first-load handler for /examples/wizard. It reads the
// (normalized) state from the cookie and renders either a full Layout or an HTMX
// Fragment depending on the HX-Request header. Unlike todo, the wizard starts
// empty, so there is no sample seed.
func (s *Server) renderWizardPage(w http.ResponseWriter, r *http.Request) {
	st := wizard.FromRequest(r).Normalized()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content := examples.WizardApp(st, nil)
	meta := components.MetaForKey("examples/wizard")
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		_ = demo.ComponentDocsFragment(meta, "wizard", content, storageAllowed(r)).Render(r.Context(), w)
		return
	}
	_ = demo.ComponentDocsLayout(meta, "wizard", content, storageAllowed(r)).Render(r.Context(), w)
}

// applyStepInput parses the posted form for the current step and stores it on st.
// Values are persisted even when invalid so Back/Next never loses entered data.
func applyStepInput(st *wizard.WizardState, r *http.Request) {
	switch st.Step {
	case 1:
		st.SetAccount(r.FormValue("name"), r.FormValue("email"), r.FormValue("password"))
	case 2:
		st.SetAddress(r.FormValue("line1"), r.FormValue("city"), r.FormValue("country"), r.FormValue("postal"))
	case 3:
		st.SetPlan(r.FormValue("plan"))
	}
}

// firstInvalidStep returns the lowest input step (1..3) that fails validation and
// its errors, or (0, nil) when every input step passes.
func firstInvalidStep(st wizard.WizardState) (int, map[string]string) {
	for step := wizard.FirstStep; step < wizard.LastStep; step++ {
		if errs := wizard.ValidateStep(step, st); len(errs) > 0 {
			return step, errs
		}
	}
	return 0, nil
}

// writeWizardBody renders the swappable body fragment and sets the content type.
func writeWizardBody(r *http.Request, w http.ResponseWriter, st wizard.WizardState, errs map[string]string) {
	writeHTML(w)
	_ = examples.WizardBody(st, errs).Render(r.Context(), w)
}

func (s *Server) handleWizardNext(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := wizard.FromRequest(r).Normalized()
	applyStepInput(&st, r)

	if errs := wizard.ValidateStep(st.Step, st); len(errs) > 0 {
		// Re-render the same step with inline errors; persist entered data so it
		// survives the failed submit. HTTP 200 so htmx swaps the fragment in.
		persistWizard(r, w, st)
		writeWizardBody(r, w, st, errs)
		_ = toast.OOBToast(toast.Config{
			Tone:    toast.ToneWarning,
			Title:   "Check your entries",
			Message: "Please fix the highlighted fields.",
		}).Render(r.Context(), w)
		return
	}

	st.Advance()
	persistWizard(r, w, st)
	writeWizardBody(r, w, st, nil)
	_ = toast.OOBToast(toast.Config{
		Tone:    toast.ToneSuccess,
		Title:   "Saved",
		Message: "Step saved.",
	}).Render(r.Context(), w)
}

func (s *Server) handleWizardBack(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := wizard.FromRequest(r).Normalized()
	st.Back()
	persistWizard(r, w, st)
	writeWizardBody(r, w, st, nil)
}

func (s *Server) handleWizardConfirm(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := wizard.FromRequest(r).Normalized()

	// Defense in depth: never confirm a flow whose earlier steps are invalid
	// (e.g. a stale or hand-crafted cookie). Jump back to the first bad step.
	if step, errs := firstInvalidStep(st); step != 0 {
		st.Step = step
		persistWizard(r, w, st)
		writeWizardBody(r, w, st, errs)
		_ = toast.OOBToast(toast.Config{
			Tone:    toast.ToneWarning,
			Title:   "Almost there",
			Message: "Some details still need fixing.",
		}).Render(r.Context(), w)
		return
	}

	st.Confirm()
	persistWizard(r, w, st)
	writeWizardBody(r, w, st, nil)
	_ = toast.OOBToast(toast.Config{
		Tone:    toast.ToneSuccess,
		Title:   "Welcome aboard",
		Message: "Your account is ready.",
	}).Render(r.Context(), w)
}

func (s *Server) handleWizardReset(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := wizard.FromRequest(r).Normalized()
	st.Reset()
	persistWizard(r, w, st)
	writeWizardBody(r, w, st, nil)
}

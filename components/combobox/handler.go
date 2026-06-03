package combobox

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler returns an http.Handler that routes three sub-paths:
//
//	GET  /options   → OptionsList
//	POST /toggle    → Body, sets HX-Trigger
//	POST /clear     → Body (Selected=[]), sets HX-Trigger
//
// Sub-paths are matched by the suffix of the request URL path.
func Handler(cfg Config, provider OptionsProvider) http.Handler {
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	registerID(cfg.ID)
	return &comboHandler{cfg: cfg, provider: provider}
}

type comboHandler struct {
	cfg      Config
	provider OptionsProvider
}

func (h *comboHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch {
	case strings.HasSuffix(r.URL.Path, "/options") && r.Method == http.MethodGet:
		h.serveOptions(w, r)
	case strings.HasSuffix(r.URL.Path, "/toggle") && r.Method == http.MethodPost:
		h.serveToggle(w, r)
	case strings.HasSuffix(r.URL.Path, "/clear") && r.Method == http.MethodPost:
		h.serveClear(w, r)
	default:
		http.Error(w, "combobox: unsupported route", http.StatusNotFound)
	}
}

// parseRequest extracts search, deps, and current selection from the request form.
func (h *comboHandler) parseRequest(r *http.Request) (search string, selected []string, deps map[string]string) {
	search = r.Form.Get("q")
	selected = r.Form[h.cfg.Name]
	deps = make(map[string]string, len(h.cfg.DependsOn))
	for _, dep := range h.cfg.DependsOn {
		deps[dep] = r.Form.Get(dep)
	}
	return
}

func (h *comboHandler) serveOptions(w http.ResponseWriter, r *http.Request) {
	search, selected, deps := h.parseRequest(r)
	opts, err := h.provider(r.Context(), search, deps)
	if err != nil {
		writeProviderError(w, r, h.cfg)
		return
	}
	state := State{Options: opts, Selected: selected, Search: search, Deps: deps}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = OptionsList(h.cfg, state).Render(r.Context(), w)
}

func (h *comboHandler) serveToggle(w http.ResponseWriter, r *http.Request) {
	search, selected, deps := h.parseRequest(r)
	value := r.Form.Get("value")

	switch h.cfg.Mode {
	case ModeSingle:
		selected = []string{value}
	case ModeMultiple:
		selected = toggleMembership(selected, value)
	}

	opts, err := h.provider(r.Context(), search, deps)
	if err != nil {
		writeProviderError(w, r, h.cfg)
		return
	}

	// Cascading invalidation: drop selections not present in current options.
	selected = filterToExistingOptions(selected, opts)

	state := State{Options: opts, Selected: selected, Search: search, Deps: deps}
	writeHXTrigger(w, h.cfg.ID, selected)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = Body(h.cfg, state).Render(r.Context(), w)
	_ = TriggerLabelOOB(h.cfg, state).Render(r.Context(), w)
}

func (h *comboHandler) serveClear(w http.ResponseWriter, r *http.Request) {
	search, _, deps := h.parseRequest(r)
	opts, err := h.provider(r.Context(), search, deps)
	if err != nil {
		writeProviderError(w, r, h.cfg)
		return
	}
	state := State{Options: opts, Selected: nil, Search: search, Deps: deps}
	writeHXTrigger(w, h.cfg.ID, nil)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = Body(h.cfg, state).Render(r.Context(), w)
	_ = TriggerLabelOOB(h.cfg, state).Render(r.Context(), w)
}

func toggleMembership(selected []string, value string) []string {
	for i, v := range selected {
		if v == value {
			return append(selected[:i], selected[i+1:]...)
		}
	}
	return append(selected, value)
}

func filterToExistingOptions(selected []string, options []Option) []string {
	if len(selected) == 0 {
		return selected
	}
	valid := make(map[string]struct{}, len(options))
	for _, o := range options {
		valid[o.Value] = struct{}{}
	}
	out := make([]string, 0, len(selected))
	for _, v := range selected {
		if _, ok := valid[v]; ok {
			out = append(out, v)
		}
	}
	return out
}

func writeHXTrigger(w http.ResponseWriter, id string, values []string) {
	payload, _ := json.Marshal(map[string]any{
		"combobox:change": map[string]any{"id": id, "values": values},
	})
	w.Header().Set("HX-Trigger", string(payload))
}

func writeProviderError(w http.ResponseWriter, r *http.Request, cfg Config) {
	w.Header().Set("HX-Retarget", "#"+cfg.ID+"-options")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadGateway)
	_ = ProviderError(cfg).Render(r.Context(), w)
}

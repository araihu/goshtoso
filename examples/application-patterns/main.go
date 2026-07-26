package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/steps"
	"github.com/araihu/goshtoso/components/table"
)

const defaultAddr = ":3000"

//go:embed app.css
var appCSS []byte

type viewState string

const (
	stateLoading viewState = "loading"
	stateEmpty   viewState = "empty"
	stateError   viewState = "error"
	stateSuccess viewState = "success"
)

type appearance struct {
	Theme string
	Mode  string
}

func (a appearance) dark() bool {
	return a.Mode == "dark"
}

type pageMeta struct {
	Title       string
	Pattern     string
	CurrentPath string
	Appearance  appearance
	State       viewState
	Step        int
}

type operation struct {
	ID          string
	Service     string
	Summary     string
	Status      string
	Owner       string
	Environment string
	Started     string
	Change      string
	Evidence    []string
}

type operationsView struct {
	State      viewState
	Table      table.Config
	Appearance appearance
}

type workflowView struct {
	Step        int
	Environment string
	Version     string
	Error       string
	Success     bool
}

type application struct {
	operations []operation
}

func newApplication() *application {
	return &application{
		operations: []operation{
			{
				ID:          "op-104",
				Service:     "Payments API",
				Summary:     "Roll out retry budget v2",
				Status:      "running",
				Owner:       "SRE",
				Environment: "production",
				Started:     "8 minutes ago",
				Change:      "release/payments-v1.18.0",
				Evidence:    []string{"Error rate steady at 0.08%", "p95 latency below 210 ms", "3 of 5 regions complete"},
			},
			{
				ID:          "op-103",
				Service:     "Ledger Worker",
				Summary:     "Drain delayed reconciliation queue",
				Status:      "attention",
				Owner:       "Finance Platform",
				Environment: "production",
				Started:     "27 minutes ago",
				Change:      "runbook/ledger-drain",
				Evidence:    []string{"Queue depth reduced by 61%", "Two partitions remain above target", "No duplicate settlements observed"},
			},
			{
				ID:          "op-102",
				Service:     "Identity",
				Summary:     "Rotate signing keys",
				Status:      "completed",
				Owner:       "Security",
				Environment: "staging",
				Started:     "1 hour ago",
				Change:      "security/key-rotation-2026-07",
				Evidence:    []string{"New key active in staging", "Old key removed from discovery", "Token verification smoke tests passed"},
			},
		},
	}
}

func (a *application) handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.HandleFunc("GET /app.css", a.styles)
	mux.HandleFunc("GET /{$}", a.appShell)
	mux.HandleFunc("GET /operations", a.operationsList)
	mux.HandleFunc("GET /operations/{id}", a.operationDetail)
	mux.HandleFunc("GET /workflows/deploy", a.workflow)
	mux.HandleFunc("POST /workflows/deploy", a.workflow)
	return mux
}

func appMux() http.Handler {
	return newApplication().handler()
}

func (a *application) styles(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(appCSS)
}

func (a *application) appShell(w http.ResponseWriter, r *http.Request) {
	prefs := appearanceFromRequest(r)
	meta := pageMeta{
		Title:       "Control room",
		Pattern:     "App Shell",
		CurrentPath: "/",
		Appearance:  prefs,
	}
	render(w, r, http.StatusOK, appShellPage(meta, a.operations))
}

func (a *application) operationsList(w http.ResponseWriter, r *http.Request) {
	prefs := appearanceFromRequest(r)
	state := stateFromRequest(r)
	view := operationsView{
		State:      state,
		Table:      operationsTable(a.operations, prefs),
		Appearance: prefs,
	}
	if r.Header.Get("HX-Request") == "true" {
		render(w, r, http.StatusOK, operationsStatePanel(view))
		return
	}
	meta := pageMeta{
		Title:       "Operations",
		Pattern:     "Operations List",
		CurrentPath: "/operations",
		Appearance:  prefs,
		State:       state,
	}
	render(w, r, http.StatusOK, operationsPage(meta, view))
}

func (a *application) operationDetail(w http.ResponseWriter, r *http.Request) {
	op, ok := a.findOperation(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	prefs := appearanceFromRequest(r)
	meta := pageMeta{
		Title:       op.Service,
		Pattern:     "Detail Workspace",
		CurrentPath: "/operations/" + op.ID,
		Appearance:  prefs,
	}
	render(w, r, http.StatusOK, operationDetailPage(meta, op))
}

func (a *application) workflow(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		a.advanceWorkflow(w, r)
		return
	}
	prefs := appearanceFromRequest(r)
	step := boundedStep(r.FormValue("step"))
	view := workflowView{
		Step:        step,
		Environment: r.FormValue("environment"),
		Version:     r.FormValue("version"),
	}
	a.renderWorkflow(w, r, http.StatusOK, prefs, view)
}

func (a *application) advanceWorkflow(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid workflow form", http.StatusBadRequest)
		return
	}
	prefs := appearanceFromRequest(r)
	view := workflowView{
		Step:        boundedStep(r.FormValue("step")),
		Environment: strings.TrimSpace(r.FormValue("environment")),
		Version:     strings.TrimSpace(r.FormValue("version")),
	}
	if r.FormValue("action") == "back" {
		if view.Step > 1 {
			view.Step--
		}
		a.renderWorkflow(w, r, http.StatusOK, prefs, view)
		return
	}

	switch view.Step {
	case 1:
		if view.Environment == "" {
			view.Error = "Choose an environment before continuing."
			a.renderWorkflow(w, r, http.StatusUnprocessableEntity, prefs, view)
			return
		}
		view.Step = 2
	case 2:
		if view.Version == "" {
			view.Error = "Enter a release version before reviewing the deployment."
			a.renderWorkflow(w, r, http.StatusUnprocessableEntity, prefs, view)
			return
		}
		view.Step = 3
	case 3:
		view.Success = true
	}

	a.renderWorkflow(w, r, http.StatusOK, prefs, view)
}

func (a *application) renderWorkflow(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	prefs appearance,
	view workflowView,
) {
	meta := pageMeta{
		Title:       "Deploy release",
		Pattern:     "Multi-step Workflow",
		CurrentPath: "/workflows/deploy",
		Appearance:  prefs,
		Step:        view.Step,
	}
	render(w, r, status, workflowPage(meta, view))
}

func (a *application) findOperation(id string) (operation, bool) {
	for _, op := range a.operations {
		if op.ID == id {
			return op, true
		}
	}
	return operation{}, false
}

func render(w http.ResponseWriter, r *http.Request, status int, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("render %s: %v", r.URL.Path, err)
	}
}

func appearanceFromRequest(r *http.Request) appearance {
	theme := strings.ToLower(r.FormValue("theme"))
	if theme != "minimal" {
		theme = "goshtoso"
	}
	mode := strings.ToLower(r.FormValue("mode"))
	if mode != "dark" {
		mode = "light"
	}
	return appearance{Theme: theme, Mode: mode}
}

func stateFromRequest(r *http.Request) viewState {
	switch viewState(strings.ToLower(r.URL.Query().Get("state"))) {
	case stateLoading:
		return stateLoading
	case stateEmpty:
		return stateEmpty
	case stateError:
		return stateError
	default:
		return stateSuccess
	}
}

func boundedStep(raw string) int {
	step, err := strconv.Atoi(raw)
	if err != nil || step < 1 || step > 3 {
		return 1
	}
	return step
}

func routeURL(path string, prefs appearance) string {
	return routeURLWith(path, prefs)
}

func routeURLWith(path string, prefs appearance, keyValues ...string) string {
	values := url.Values{
		"theme": []string{prefs.Theme},
		"mode":  []string{prefs.Mode},
	}
	for i := 0; i+1 < len(keyValues); i += 2 {
		if keyValues[i+1] != "" {
			values.Set(keyValues[i], keyValues[i+1])
		}
	}
	return path + "?" + values.Encode()
}

func operationsTable(operations []operation, prefs appearance) table.Config {
	rows := make([]table.Row, 0, len(operations))
	for _, op := range operations {
		rows = append(rows, table.Row{
			ID: op.ID,
			Cells: map[string]table.Cell{
				"operation": {
					Text:        op.Summary,
					Description: op.ID + " · " + op.Service,
				},
				"status": {
					Component: badge.Badge(statusBadge(op.Status)),
				},
				"environment": {Text: op.Environment},
				"owner":       {Text: op.Owner},
				"started":     {Text: op.Started},
			},
			Link:     routeURL("/operations/"+op.ID, prefs),
			LinkMode: table.LinkFull,
		})
	}
	return table.Config{
		ID:      "operations",
		Caption: "Active and recently completed operations",
		Columns: []table.Column{
			{Key: "operation", Label: "Operation"},
			{Key: "status", Label: "Status"},
			{Key: "environment", Label: "Environment"},
			{Key: "owner", Label: "Owner"},
			{Key: "started", Label: "Started"},
		},
		Rows:       rows,
		Appearance: table.AppearanceStriped,
	}
}

func statusBadge(status string) badge.Config {
	cfg := badge.Config{
		Label:      status,
		Appearance: badge.AppearanceSoft,
		Size:       badge.SizeSM,
		Indicator:  true,
	}
	switch status {
	case "completed":
		cfg.Tone = badge.ToneSuccess
	case "attention":
		cfg.Tone = badge.ToneWarning
	default:
		cfg.Tone = badge.ToneInfo
	}
	return cfg
}

func operationStateTone(state viewState) badge.Tone {
	switch state {
	case stateError:
		return badge.ToneDanger
	case stateEmpty:
		return badge.ToneDefault
	case stateLoading:
		return badge.ToneInfo
	default:
		return badge.ToneSuccess
	}
}

func operationStateLabel(state viewState) string {
	return strings.ToUpper(string(state[:1])) + string(state[1:])
}

func workflowSteps(current int, success bool) steps.Config {
	items := []steps.Step{
		{ID: "workflow-step-1", Label: "Target"},
		{ID: "workflow-step-2", Label: "Release"},
		{ID: "workflow-step-3", Label: "Review"},
	}
	for i := range items {
		number := i + 1
		switch {
		case success || number < current:
			items[i].Status = steps.StatusCompleted
		case number == current:
			items[i].Status = steps.StatusCurrent
		default:
			items[i].Status = steps.StatusUpcoming
		}
	}
	return steps.Config{
		ID:          "deployment-progress",
		Steps:       items,
		Orientation: steps.OrientationHorizontal,
		ShowLabels:  true,
		AriaLabel:   "Deployment progress",
		LiveRegion:  true,
	}
}

func workflowButtonLabel(view workflowView) string {
	if view.Step == 3 {
		return "Deploy release"
	}
	return "Continue"
}

func workflowBackURL(prefs appearance, view workflowView) string {
	if view.Step <= 1 {
		return routeURL("/operations", prefs)
	}
	return routeURLWith(
		"/workflows/deploy",
		prefs,
		"step", strconv.Itoa(view.Step-1),
		"environment", view.Environment,
		"version", view.Version,
	)
}

func main() {
	log.Printf("application patterns benchmark running at http://localhost%s", defaultAddr)
	log.Fatal(http.ListenAndServe(defaultAddr, appMux()))
}

func init() {
	if len(appCSS) == 0 {
		panic(fmt.Errorf("embedded app.css is empty"))
	}
}

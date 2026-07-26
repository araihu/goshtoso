package main

import (
	"bytes"
	"go/parser"
	"go/token"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestApplicationRoutesRenderTheirPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "app shell",
			path: "/",
			want: []string{`data-pattern="app-shell"`, "Operate the platform without losing context", `href="#main-content"`, `data-scroll-region="main"`},
		},
		{
			name: "operations list",
			path: "/operations",
			want: []string{`data-pattern="operations-list"`, "Payments API", "op-104"},
		},
		{
			name: "detail workspace",
			path: "/operations/op-104",
			want: []string{`data-pattern="detail-workspace"`, "Roll out retry budget v2", "Live decision context"},
		},
		{
			name: "multi-step workflow",
			path: "/workflows/deploy",
			want: []string{`data-pattern="multi-step-workflow"`, "Deploy a release", `data-step="1"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := serve(t, http.MethodGet, tt.path, nil, nil)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d\n%s", rec.Code, http.StatusOK, rec.Body.String())
			}
			assertContains(t, rec.Body.String(), tt.want...)
		})
	}
}

func TestOperationsStateMatrixRenders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		state viewState
		want  string
	}{
		{state: stateLoading, want: "Refreshing operations"},
		{state: stateEmpty, want: "Start a deployment"},
		{state: stateError, want: "Retry operations"},
		{state: stateSuccess, want: "Operations loaded"},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			t.Parallel()
			view := operationsView{
				State:      tt.state,
				Table:      operationsTable(newApplication().operations, appearance{Theme: "goshtoso", Mode: "light"}),
				Appearance: appearance{Theme: "goshtoso", Mode: "light"},
			}
			body := renderString(t, operationsStatePanel(view))
			assertContains(t, body, `data-state="`+string(tt.state)+`"`, tt.want)
		})
	}
}

func TestOperationsStateMatrixSupportsFullAndHTMXRequests(t *testing.T) {
	t.Parallel()

	for _, state := range []viewState{stateLoading, stateEmpty, stateError, stateSuccess} {
		state := state
		t.Run(string(state), func(t *testing.T) {
			t.Parallel()
			path := "/operations?state=" + string(state)
			full := serve(t, http.MethodGet, path, nil, nil)
			assertContains(t, full.Body.String(), "<!doctype html>", `data-state="`+string(state)+`"`)

			fragment := serve(t, http.MethodGet, path, nil, map[string]string{"HX-Request": "true"})
			assertContains(t, fragment.Body.String(), `id="operations-state"`, `data-state="`+string(state)+`"`)
			if strings.Contains(strings.ToLower(fragment.Body.String()), "<!doctype html>") {
				t.Fatalf("HTMX response unexpectedly contains a full document:\n%s", fragment.Body.String())
			}
		})
	}
}

func TestAssetsAndAppearanceAreLocal(t *testing.T) {
	t.Parallel()

	goshtosoCSS := serve(t, http.MethodGet, "/assets/styles.css", nil, nil)
	if goshtosoCSS.Code != http.StatusOK {
		t.Fatalf("goshtoso CSS status = %d, want 200", goshtosoCSS.Code)
	}
	assertContains(t, goshtosoCSS.Body.String(), "[data-theme=minimal]", "[data-theme=goshtoso]")

	appStyles := serve(t, http.MethodGet, "/app.css", nil, nil)
	if appStyles.Code != http.StatusOK {
		t.Fatalf("app CSS status = %d, want 200", appStyles.Code)
	}
	assertContains(t, appStyles.Body.String(), ".app-shell", "var(--color-surface-dark)", "grid-template-columns: minmax(0, 1fr)", ".app-stack > *", ".state-card > *")

	page := serve(t, http.MethodGet, "/operations?theme=minimal&mode=dark", nil, nil)
	assertContains(t, page.Body.String(), `data-theme="minimal"`, `class="dark"`, `href="/assets/styles.css"`, `href="/app.css"`)
	for _, remote := range []string{"cdn.", "fonts.googleapis.com", "unpkg.com", "jsdelivr.net"} {
		if strings.Contains(page.Body.String(), remote) {
			t.Fatalf("page contains remote runtime reference %q", remote)
		}
	}
}

func TestWorkflowTransitionsAndValidation(t *testing.T) {
	t.Parallel()

	invalid := serveForm(t, "/workflows/deploy", url.Values{
		"step":  {"1"},
		"theme": {"goshtoso"},
		"mode":  {"light"},
	})
	if invalid.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid status = %d, want %d\n%s", invalid.Code, http.StatusUnprocessableEntity, invalid.Body.String())
	}
	assertContains(t, invalid.Body.String(), `data-state="error"`, "Choose an environment before continuing")

	stepTwo := serveForm(t, "/workflows/deploy", url.Values{
		"step":        {"1"},
		"environment": {"production"},
		"theme":       {"minimal"},
		"mode":        {"dark"},
	})
	if stepTwo.Code != http.StatusOK {
		t.Fatalf("step two status = %d, want 200\n%s", stepTwo.Code, stepTwo.Body.String())
	}
	assertContains(t, stepTwo.Body.String(), `data-step="2"`, `data-theme="minimal"`, `class="dark"`, "Release version")
	assertContains(t, stepTwo.Body.String(), `name="action"`, `value="back"`, `value="continue"`)

	backToStepOne := serveForm(t, "/workflows/deploy", url.Values{
		"action":      {"back"},
		"step":        {"2"},
		"environment": {"production"},
		"theme":       {"minimal"},
		"mode":        {"dark"},
	})
	if backToStepOne.Code != http.StatusOK {
		t.Fatalf("back status = %d, want 200\n%s", backToStepOne.Code, backToStepOne.Body.String())
	}
	assertContains(t, backToStepOne.Body.String(), `data-step="1"`, `data-theme="minimal"`, `class="dark"`)

	stepThree := serveForm(t, "/workflows/deploy", url.Values{
		"step":        {"2"},
		"environment": {"production"},
		"version":     {"v1.18.0"},
		"theme":       {"goshtoso"},
		"mode":        {"light"},
	})
	assertContains(t, stepThree.Body.String(), `data-step="3"`, "5% canary, then regional")

	success := serveForm(t, "/workflows/deploy", url.Values{
		"step":        {"3"},
		"environment": {"production"},
		"version":     {"v1.18.0"},
		"theme":       {"goshtoso"},
		"mode":        {"light"},
	})
	assertContains(t, success.Body.String(), `data-state="success"`, "Deployment queued", "v1.18.0")
}

func TestMethodQualifiedRoutesRejectWrongMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		method string
		path   string
		allow  string
	}{
		{method: http.MethodPost, path: "/operations", allow: "GET, HEAD"},
		{method: http.MethodDelete, path: "/workflows/deploy", allow: "GET, HEAD, POST"},
		{method: http.MethodPost, path: "/app.css", allow: "GET, HEAD"},
	}
	for _, tt := range tests {
		rec := serve(t, tt.method, tt.path, nil, nil)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s %s status = %d, want %d", tt.method, tt.path, rec.Code, http.StatusMethodNotAllowed)
		}
		if allow := rec.Header().Get("Allow"); allow != tt.allow {
			t.Fatalf("%s %s Allow = %q, want %q", tt.method, tt.path, allow, tt.allow)
		}
	}
}

func TestUnknownOperationIsNotFound(t *testing.T) {
	t.Parallel()

	rec := serve(t, http.MethodGet, "/operations/op-missing", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestModuleHasNoSiteImports(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, importSpec := range file.Imports {
			importPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				return err
			}
			if strings.Contains(importPath, "github.com/araihu/goshtoso/site") {
				t.Errorf("%s imports forbidden site package %q", path, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"go.mod", "views.templ"} {
		content, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(content), "github.com/araihu/goshtoso/site") {
			t.Fatalf("%s references forbidden site import path", path)
		}
	}
}

func serve(
	t *testing.T,
	method string,
	path string,
	body *strings.Reader,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == nil {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, body)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	appMux().ServeHTTP(rec, req)
	return rec
}

func serveForm(t *testing.T, path string, values url.Values) *httptest.ResponseRecorder {
	t.Helper()
	body := values.Encode()
	rec := serve(t, http.MethodPost, path, strings.NewReader(body), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	return rec
}

func renderString(t *testing.T, component templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := component.Render(t.Context(), &buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func assertContains(t *testing.T, body string, values ...string) {
	t.Helper()
	lowerBody := strings.ToLower(body)
	for _, value := range values {
		if !strings.Contains(lowerBody, strings.ToLower(value)) {
			t.Fatalf("body does not contain %q:\n%s", value, body)
		}
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve test file")
	}
	return filepath.Dir(filename)
}

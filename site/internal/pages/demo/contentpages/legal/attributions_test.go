package legalpages

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/assets"
)

func TestAttributionsRenderExactGeneratedRuntimeInventory(t *testing.T) {
	var rendered bytes.Buffer
	if err := attributionsContent().Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render attributions: %v", err)
	}
	html := rendered.String()
	wants := []string{"DefaultRuntimeManifest", "CDN-first", "local fallback"}
	for _, runtime := range runtimeAttributionRows() {
		wants = append(wants, runtime.Name, runtime.Version, runtime.LocalURL, runtime.License)
		if runtime.LicenseURL != "" {
			wants = append(wants, runtime.LicenseURL)
		}
	}
	for _, want := range wants {
		if !strings.Contains(html, want) {
			t.Errorf("attributions missing %q", want)
		}
	}
	if !strings.Contains(html, `href="/assets/js/runtime/versions.json"`) {
		t.Error("attributions do not link to the embedded compatibility manifest")
	}
}

func TestDisplayedRuntimeLicenseLinksAreServedByLinkedAssets(t *testing.T) {
	for _, runtime := range runtimeAttributionRows() {
		if runtime.LicenseURL == "" {
			continue
		}
		request := httptest.NewRequest(http.MethodGet, runtime.LicenseURL, nil)
		response := httptest.NewRecorder()
		assets.Handler().ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Errorf("GET %s status = %d", runtime.LicenseURL, response.Code)
		}
	}
}

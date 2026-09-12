package consumer_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/head"
)

func TestPublishedPackage(t *testing.T) {
	output, err := exec.Command("go", "run", "./cmd/probe").Output()
	if err != nil {
		t.Fatal(err)
	}
	var identity assets.VersionInfo
	if err := json.Unmarshal(output, &identity); err != nil {
		t.Fatal(err)
	}
	if identity.Status != assets.VersionExact || identity.Version == "" {
		t.Fatalf("consumer did not use a published module: %#v", identity)
	}
	if expected := os.Getenv("GOSHTOSO_EXPECTED_VERSION"); expected != "" && identity.Version != expected {
		t.Fatalf("version = %q, want %q", identity.Version, expected)
	}
	var markup strings.Builder
	if err := button.Button(button.WithType("submit")).Render(context.Background(), &markup); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(markup.String(), `type="submit"`) {
		t.Fatalf("button markup: %s", &markup)
	}
	markup.Reset()
	if err := head.Dependencies().Render(context.Background(), &markup); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(markup.String(), "script") {
		t.Fatalf("runtime markup: %s", &markup)
	}
	manifest := assets.DefaultRuntimeManifest()
	for _, path := range []string{"/assets/styles.css", "/assets/goshtoso-theme.css", manifest.Loader.LocalURL} {
		w := httptest.NewRecorder()
		assets.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK || w.Body.Len() == 0 {
			t.Fatalf("asset %s: status=%d bytes=%d", path, w.Code, w.Body.Len())
		}
	}
}

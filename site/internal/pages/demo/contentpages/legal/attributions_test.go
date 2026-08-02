package legalpages

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestAttributionsRenderExactGeneratedRuntimeInventory(t *testing.T) {
	var rendered bytes.Buffer
	if err := attributionsContent().Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render attributions: %v", err)
	}
	html := rendered.String()
	wants := []string{"DefaultRuntimeManifest", "CDN-first", "local fallback"}
	for _, runtime := range runtimeAttributions {
		wants = append(wants, runtime.Name, runtime.Version, runtime.LocalURL, runtime.License)
	}
	for _, want := range wants {
		if !strings.Contains(html, want) {
			t.Errorf("attributions missing %q", want)
		}
	}
	if !strings.Contains(html, `href="https://github.com/araihu/goshtoso/blob/main/assets/js/runtime/manifest.json"`) {
		t.Error("attributions do not link to the canonical manifest source")
	}
	if !strings.Contains(html, `href="/assets/js/runtime/versions.json"`) {
		t.Error("attributions do not link to the embedded compatibility manifest")
	}
}

package runtimecontract_test

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
	"github.com/araihu/goshtoso/components/head"
)

func TestExternalConsumerCanBindManifestAndReplacementIdentity(t *testing.T) {
	command := exec.Command("go", "run", "./cmd/probe")
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.Output()
	if err != nil {
		t.Fatalf("run external consumer probe: %v", err)
	}
	var identity assets.VersionInfo
	if err := json.Unmarshal(output, &identity); err != nil {
		t.Fatalf("decode external consumer identity: %v: %s", err, output)
	}
	if identity.Status != assets.VersionReplaced {
		t.Fatalf("GoshtosoVersion status = %q, want %q: %#v", identity.Status, assets.VersionReplaced, identity)
	}
	if identity.Version != "" {
		t.Fatalf("replaced build exposed exact version %q", identity.Version)
	}
	if identity.Requested.Path != assets.GoshtosoModulePath || identity.Requested.Version != "v0.0.13" {
		t.Fatalf("requested module = %#v", identity.Requested)
	}
	if identity.Replacement.Path == "" {
		t.Fatalf("replacement metadata is empty: %#v", identity)
	}

	manifest := assets.DefaultRuntimeManifest()
	if len(manifest.Dependencies) != 11 {
		t.Fatalf("dependency count = %d, want 11", len(manifest.Dependencies))
	}
	if manifest.Dependencies[0].Role != assets.RuntimeRoleAlpineCollapse {
		t.Fatalf("first dependency = %q, want %q", manifest.Dependencies[0].Role, assets.RuntimeRoleAlpineCollapse)
	}

	server := httptest.NewServer(assets.Handler())
	t.Cleanup(server.Close)
	response, err := http.Get(server.URL + manifest.Loader.LocalURL)
	if err != nil {
		t.Fatalf("GET loader: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET loader status = %d, want 200", response.StatusCode)
	}
}

func TestExternalConsumerCanRenderCustomPublicRuntimeManifest(t *testing.T) {
	manifest := assets.DefaultRuntimeManifest()
	manifest.Stylesheet.PrimaryURL = "/consumer/goshtoso.css"
	manifest.Loader.PrimaryURL = "/consumer/loader.js"
	for index := range manifest.Dependencies {
		dependency := &manifest.Dependencies[index]
		if dependency.Role == assets.RuntimeRoleDarkMode || dependency.Role == assets.RuntimeRoleHTMXExtSSE || dependency.Role == assets.RuntimeRoleHTMXExtWS {
			dependency.Enabled = true
		}
	}
	manifest.Dependencies = append(manifest.Dependencies, assets.RuntimeAsset{
		Role: "consumer-runtime", Kind: assets.RuntimeAssetScript,
		PrimaryURL: "/consumer/runtime.js", LocalURL: "/consumer/runtime.js",
		Enabled: true, IncludeInMinimal: true,
	})

	var output strings.Builder
	err := head.Dependencies(
		head.WithDependencyCDNURL(head.DependencyHTMX, "/consumer/htmx.js"),
		head.WithRuntimeManifest(manifest),
		head.WithoutLocalFallback(),
	).Render(context.Background(), &output)
	if err != nil {
		t.Fatalf("render custom public manifest: %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`href="/consumer/goshtoso.css"`,
		`src="/consumer/loader.js"`,
		`consumer-runtime`,
		`/consumer/htmx.js`,
		`htmx-ext-sse`,
		`htmx-ext-ws`,
		`dark-mode`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("custom public manifest markup missing %q:\n%s", want, markup)
		}
	}
}

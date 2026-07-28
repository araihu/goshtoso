package runtimecontract_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/araihu/goshtoso/assets"
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
	if len(manifest.Dependencies) != 6 {
		t.Fatalf("dependency count = %d, want 6", len(manifest.Dependencies))
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

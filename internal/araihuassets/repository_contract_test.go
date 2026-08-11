package araihuassets

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRepositoryProducerSetIsPinnedToAssetsV021(t *testing.T) {
	t.Parallel()

	repoRoot := repositoryRoot(t)
	manifestBytes, err := os.ReadFile(filepath.Join(repoRoot, DefaultManifestPath))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeManifest(manifestBytes)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Release != "v0.2.1" {
		t.Fatalf("manifest release = %q, want v0.2.1", manifest.Release)
	}

	wantMappings := map[string]string{
		"examples/getting-started/araihu.css": "themes/araihu.css",
		"assets/icons/heroicons.svg":          "icons/ui/sprite.svg",
		"assets/images/goshtoso-logo.svg":     "brand/goshtoso/logo/adaptive-transparent-optical.svg",
		"assets/images/goshtoso-mark.svg":     "icons/brand/goshtoso-icon-adaptive-transparent-optical.svg",
		"assets/favicon.svg":                  "platform/web/goshtoso/favicon.svg",
	}
	if len(manifest.Mappings) != len(wantMappings) {
		t.Fatalf("manifest mapping count = %d, want %d", len(manifest.Mappings), len(wantMappings))
	}
	for _, mapping := range manifest.Mappings {
		wantSource, ok := wantMappings[mapping.Destination]
		if !ok {
			t.Errorf("unexpected producer destination %q", mapping.Destination)
			continue
		}
		if mapping.Source != wantSource {
			t.Errorf("producer for %q = %q, want %q", mapping.Destination, mapping.Source, wantSource)
		}
	}

	goMod, err := os.ReadFile(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	version, ok := moduleRequirementVersion(goMod, "github.com/araihu/assets")
	if !ok || version != "v0.2.1" {
		t.Fatalf("go.mod authenticates github.com/araihu/assets at %q, want %s", version, manifest.Release)
	}
	goSum, err := os.ReadFile(filepath.Join(repoRoot, "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"github.com/araihu/assets v0.2.1 h1:I/vsIwNahHh3Ip+KzH/Zq2WkSMoq3Z9T5JpQCuvHgxU=",
		"github.com/araihu/assets v0.2.1/go.mod h1:3D6/Mm498XXjNkFIZAbSXSu6tvFFiPE/S4/8aKc2/BI=",
	} {
		if !lineExists(goSum, want) {
			t.Errorf("go.sum missing authenticated producer entry %q", want)
		}
	}
}

func moduleRequirementVersion(contents []byte, module string) (string, bool) {
	for _, line := range strings.Split(string(contents), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == module {
			return fields[1], true
		}
	}
	return "", false
}

func lineExists(contents []byte, want string) bool {
	for _, line := range strings.Split(string(contents), "\n") {
		if line == want {
			return true
		}
	}
	return false
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

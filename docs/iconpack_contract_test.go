package docs_test

import (
	"strings"
	"testing"
)

func TestIconpackReleaseDocumentationContract(t *testing.T) {
	guide := readDoc(t, "ICONPACK.md")
	checklist := readDoc(t, "RELEASE_CHECKLIST.md")
	changelog := readDoc(t, "../CHANGELOG.md")

	for _, required := range []string{
		"github.com/araihu/goshtoso/cmd/iconpack@v0.2.0",
		"https://github.com/araihu/assets/releases/download/v0.2.0/araihu-assets-v0.2.0.tar.gz",
		"a0e8e5c8928e37de979ce9a60f3d66fad1aa1b4c7d2904f9275f0be9932a33d6",
		"dcb97bbbbf98fb2e3c0e96b63eefb17b9b60eb2b3d8097fa6b4e2876f3f19271",
		"JSON is the canonical manifest form",
		"repeated `-name` flags are also supported",
		"ui/heroicons",
		"brand/developer-icons",
	} {
		if !strings.Contains(guide, required) {
			t.Errorf("ICONPACK.md missing %q", required)
		}
	}
	for _, required := range []string{
		"## v0.2.0 iconpack release gate",
		"site/go.mod",
		"both site module contracts pass",
	} {
		if !strings.Contains(checklist, required) {
			t.Errorf("RELEASE_CHECKLIST.md missing %q", required)
		}
	}
	if !strings.Contains(changelog, "[v0.2.0]") {
		t.Error("CHANGELOG.md missing v0.2.0 release reference")
	}
}

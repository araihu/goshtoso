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
		"5d7d691e22d4071507b0bf2248713d7008adf57c18840cfd46e20901db0b78e5",
		"JSON is the canonical manifest form",
		"repeated `-name` flags are also supported",
		"not by the names of its parent directories",
		"below a consumer path such as",
		"`internal/`",
		"`vendor/`",
		"`acquisition/`",
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

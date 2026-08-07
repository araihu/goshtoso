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
		"77c696ae5eceb5e7bc11d19affb7c2c7b7e8afc6414882b9b059239e315f2260",
		"334005c77622250a1e827b9472161cd6e56c82d487fc0d44023d49261f8dbee5",
		"publicly released",
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
		"77c696ae5eceb5e7bc11d19affb7c2c7b7e8afc6414882b9b059239e315f2260",
		"334005c77622250a1e827b9472161cd6e56c82d487fc0d44023d49261f8dbee5",
		"5d7d691e22d4071507b0bf2248713d7008adf57c18840cfd46e20901db0b78e5",
	} {
		if !strings.Contains(checklist, required) {
			t.Errorf("RELEASE_CHECKLIST.md missing %q", required)
		}
	}
	const expectedCompare = "[v0.2.0]: https://github.com/araihu/goshtoso/compare/v0.1.8...v0.2.0"
	if !strings.Contains(changelog, expectedCompare) {
		t.Errorf("CHANGELOG.md missing %q", expectedCompare)
	}
	if strings.Contains(changelog, "[v0.2.0]: https://github.com/araihu/goshtoso/compare/v0.1.7...v0.2.0") {
		t.Error("CHANGELOG.md still compares v0.2.0 against the unpublished v0.1.7 base")
	}
}

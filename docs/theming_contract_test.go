package docs_test

import (
	"strings"
	"testing"
)

func TestThemingDocsPublishPublicCatalogBoundary(t *testing.T) {
	content := readDoc(t, "THEMING.md")
	start := strings.Index(content, "## Built-in Theme Catalog")
	if start < 0 {
		t.Fatal("THEMING.md missing Built-in Theme Catalog section")
	}
	end := strings.Index(content[start+len("## Built-in Theme Catalog"):], "\n## ")
	if end < 0 {
		end = len(content) - start
	}
	section := content[start : start+len("## Built-in Theme Catalog")+end]

	for _, want := range []string{
		"github.com/araihu/goshtoso/themes",
		"themes.BuiltIn()",
		"caller-owned values",
		"deterministic key order",
		"presentation order",
		"defaults",
		"custom themes",
		"| `90s` | 90s |",
		"| `araihu` | Arai Hû |",
		"| `arctic` | Arctic |",
		"| `christmas` | Christmas |",
		"| `dracula` | Dracula |",
		"| `goshtoso` | Goshtoso |",
		"| `halloween` | Halloween |",
		"| `high-contrast` | High Contrast |",
		"| `industrial` | Industrial |",
		"| `minimal` | Minimal |",
		"| `modern` | Modern |",
		"| `neo-brutalism` | Neo Brutalism |",
		"| `news` | News |",
		"| `pastel` | Pastel |",
		"| `prototype` | Prototype |",
		"| `zombie` | Zombie |",
	} {
		if !strings.Contains(section, want) {
			t.Errorf("THEMING.md public catalog section missing %q", want)
		}
	}
	if strings.Contains(section, "Halloween II") {
		t.Error("THEMING.md public catalog section promotes site presentation override Halloween II")
	}
	if strings.Contains(section, "--color-") {
		t.Error("THEMING.md public catalog section leaks CSS token details")
	}
}

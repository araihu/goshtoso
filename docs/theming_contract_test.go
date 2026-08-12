package docs_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/themes"
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
		"theme.Ownership",
		"caller-owned values",
		"deterministic key order",
		"presentation order",
		"defaults",
		"custom themes",
		"classification only",
	} {
		if !strings.Contains(section, want) {
			t.Errorf("THEMING.md public catalog section missing %q", want)
		}
	}
	catalog := themes.BuiltIn()
	for _, theme := range catalog {
		ownership := strings.ToUpper(string(theme.Ownership[:1])) + string(theme.Ownership[1:])
		row := fmt.Sprintf("| `%s` | %s | %s |", theme.Key, theme.Label, ownership)
		if count := strings.Count(section, row); count != 1 {
			t.Errorf("THEMING.md public catalog row %q count = %d, want 1", row, count)
		}
	}
	if rows := strings.Count(section, "| `"); rows != len(catalog) {
		t.Errorf("THEMING.md public catalog has %d theme rows, want %d", rows, len(catalog))
	}
	if strings.Contains(section, "Halloween II") {
		t.Error("THEMING.md public catalog section promotes site presentation override Halloween II")
	}
	if strings.Contains(section, "--color-") {
		t.Error("THEMING.md public catalog section leaks CSS token details")
	}
}

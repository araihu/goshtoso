package docs_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestUsageDocumentsScrollRegionPublicContract(t *testing.T) {
	content, err := os.ReadFile("USAGE.md")
	if err != nil {
		t.Fatalf("read USAGE.md: %v", err)
	}
	usage := string(content)
	for _, want := range []string{
		"components/scrollregion",
		"scrollregion.Named(scrollregion.Config{",
		"scrollregion.Labelled(scrollregion.Config{",
		"AccessibleName{Label:",
		"LabelledBy",
		"named `region`",
		"unique `Label`",
		"DisableIndicators",
		"keyboard",
		"touch",
	} {
		if !strings.Contains(usage, want) {
			t.Errorf("USAGE.md missing Scroll Region contract %q", want)
		}
	}
}

// TestScrollRegionGeneratedReferencePreservesConfigCompatibility makes the
// generated skill reference an API contract. Config deliberately retains its
// original positional shape; naming belongs to the separate AccessibleName
// value and Named/Labelled constructors, not extra Config fields.
func TestScrollRegionGeneratedReferencePreservesConfigCompatibility(t *testing.T) {
	sourcePath := filepath.Join("..", "components", "scrollregion", "types.go")
	configFields := scrollRegionStructFields(t, sourcePath, "Config")
	if want := []string{"Content", "RootClass", "ViewportClass", "DisableIndicators"}; !reflect.DeepEqual(configFields, want) {
		t.Fatalf("ScrollRegion Config fields = %v, want exact compatibility order %v", configFields, want)
	}
	if got, want := scrollRegionStructFields(t, sourcePath, "AccessibleName"), []string{"Label", "LabelledBy"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("AccessibleName fields = %v, want %v", got, want)
	}

	const externalPath = "../.agents/skills/using-goshtoso/references/components-reference.md"
	const legacyPath = "../.claude/skills/using-goshtoso/components-reference.md"
	external, err := os.ReadFile(externalPath)
	if err != nil {
		t.Fatalf("read external generated component reference: %v", err)
	}
	legacy, err := os.ReadFile(legacyPath)
	if err != nil {
		t.Fatalf("read legacy generated component reference: %v", err)
	}
	if string(external) != string(legacy) {
		t.Fatal("generated external and legacy component references differ")
	}

	section := scrollRegionReferenceSection(t, string(external))
	for _, want := range []string{
		"**Entry points:** `Labelled(cfg Config, labelledBy string)` · `Named(cfg Config, name AccessibleName)` · `ScrollRegion(cfg Config)`",
		"### Usage",
		"@scrollregion.Labelled(scrollregion.Config{",
		"\"activity-history-heading\"",
		"**AccessibleName**",
		"| `Label` | `string` | Label sets the accessible name.",
		"| `LabelledBy` | `string` | LabelledBy sets one or more existing element IDs that name the region.",
		"**Config**",
		"| `Content` | `templ.Component` | Content is rendered inside the scroll viewport. |",
		"| `RootClass` | `string` | RootClass appends classes to the positioning root. |",
		"| `ViewportClass` | `string` | ViewportClass appends classes to the scroll viewport. |",
		"| `DisableIndicators` | `bool` | DisableIndicators keeps the sentinels but omits the visual boundary cues. |",
	} {
		if !strings.Contains(section, want) {
			t.Errorf("generated Scroll Region reference missing %q", want)
		}
	}
	configStart := strings.Index(section, "**Config**")
	if configStart < 0 {
		t.Fatal("generated Scroll Region reference has no Config section")
	}
	configSection := section[configStart:]
	if next := strings.Index(configSection[len("**Config**"):], "**"); next >= 0 {
		configSection = configSection[:len("**Config**")+next]
	}
	for _, stale := range []string{"| `Label` |", "| `LabelledBy` |"} {
		if strings.Contains(configSection, stale) {
			t.Errorf("generated Config section incorrectly documents %s as a Config field", stale)
		}
	}
}

func scrollRegionStructFields(t *testing.T, sourcePath, typeName string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), sourcePath, nil, 0)
	if err != nil {
		t.Fatalf("parse Scroll Region types: %v", err)
	}
	for _, declaration := range parsed.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, spec := range general.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != typeName {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				t.Fatalf("Scroll Region %s is not a struct", typeName)
			}
			var fields []string
			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					fields = append(fields, name.Name)
				}
			}
			return fields
		}
	}
	t.Fatalf("Scroll Region %s is absent", typeName)
	return nil
}

func scrollRegionReferenceSection(t *testing.T, reference string) string {
	t.Helper()
	const heading = "## scrollregion\n"
	start := strings.Index(reference, heading)
	if start < 0 {
		t.Fatal("generated reference has no Scroll Region section")
	}
	section := reference[start:]
	if next := strings.Index(section[len(heading):], "\n## "); next >= 0 {
		section = section[:len(heading)+next]
	}
	return section
}

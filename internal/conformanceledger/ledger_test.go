package conformanceledger

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestDeriveInventoryPackageSetMatchesGoList(t *testing.T) {
	inventory, err := DeriveInventory(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "list", "./components/...")
	command.Dir = repoRoot(t)
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Fields(string(output))
	got := make([]string, len(inventory.Packages))
	for index, item := range inventory.Packages {
		got[index] = item.Value
	}
	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("derived packages do not match GOWORK=off go list\ngot=%v\nwant=%v", got, want)
	}
}

func TestDeriveRoutesRejectsDuplicateValues(t *testing.T) {
	repo := t.TempDir()
	path := filepath.Join(repo, "site", "internal", "pages", "catalog")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	source := []byte(`package catalog
var pages = []struct{ Path string }{
	{
		Path: "/components/button",
	},
	{
		Path: "/components/button",
	},
}
`)
	if err := os.WriteFile(filepath.Join(path, "catalog.go"), source, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := deriveRoutes(repo); err == nil || !strings.Contains(err.Error(), "duplicate route") {
		t.Fatalf("duplicate route error = %v", err)
	}
}

func TestDeriveStateMetadataRejectsParserErrorsAndDuplicates(t *testing.T) {
	t.Run("parser", func(t *testing.T) {
		repo := t.TempDir()
		path := filepath.Join(repo, "components", "broken")
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "types.go"), []byte("package broken\nconst ("), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := deriveStateMetadata(repo, nil); err == nil || !strings.Contains(err.Error(), "parse state metadata") {
			t.Fatalf("parser error = %v", err)
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		repo := t.TempDir()
		writeMaintainedConfigStateAuthorities(t, repo)
		path := filepath.Join(repo, "components", "duplicate")
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		source := []byte("package duplicate\ntype Tone string\nconst (\nToneA Tone = \"a\"\nToneA Tone = \"b\"\n)\n")
		if err := os.WriteFile(filepath.Join(path, "types.go"), source, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := deriveStateMetadata(repo, nil); err == nil || !strings.Contains(err.Error(), "duplicate state") {
			t.Fatalf("duplicate state error = %v", err)
		}
	})
}

func writeMaintainedConfigStateAuthorities(t *testing.T, repo string) {
	t.Helper()
	for _, fixture := range []struct {
		packageName string
		fields      string
	}{
		{packageName: "checkbox", fields: "Checked bool\nDisabled bool"},
		{packageName: "textinput", fields: "Disabled bool\nRequired bool\nReadonly bool"},
	} {
		path := filepath.Join(repo, "components", fixture.packageName)
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		source := "package " + fixture.packageName + "\ntype Config struct {\n" + fixture.fields + "\n}\n"
		if err := os.WriteFile(filepath.Join(path, "types.go"), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestValidateRejectsDuplicateRowIDs(t *testing.T) {
	row := Row{ID: "duplicate", Class: ClassInventory, Sources: []SourceRef{{Path: "source.go", Symbol: "Source"}}, Applicability: Applicable, ReceiptStatus: StatusExecuted, Receipt: "source"}
	err := Validate(Ledger{SchemaVersion: SchemaVersion, Rows: []Row{row, row}}, Inventory{})
	if err == nil || !contains(err.Error(), "duplicate id") {
		t.Fatalf("duplicate row id error = %v", err)
	}
}

func TestDeriveInventoryUsesCurrentSourceAuthorities(t *testing.T) {
	inventory, err := DeriveInventory(repoRoot(t))
	if err != nil {
		t.Fatalf("derive current source inventory: %v", err)
	}

	assertLen(t, "packages", inventory.Packages, 55)
	assertLen(t, "renderables", inventory.Renderables, 84)
	assertLen(t, "kinds", inventory.Kinds, 84)
	assertLen(t, "routes", inventory.Routes, 51)
	if len(inventory.States) == 0 {
		t.Fatal("source-derived configuration/default states are empty")
	}
	assertLen(t, "dynamic lifecycle states", inventory.LifecycleStates, 13)
	assertLen(t, "themes", inventory.Themes, 16)
	wantEdges := []int{640, 768, 1024, 1280, 1536}
	if !equalInts(inventory.BreakpointEdges, wantEdges) {
		t.Fatalf("compiled breakpoint edges = %v, want %v", inventory.BreakpointEdges, wantEdges)
	}
}

func TestDeriveInventoryIncludesSourceVisiblePublicConfigurationStates(t *testing.T) {
	inventory, err := DeriveInventory(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	states := make(map[string]bool, len(inventory.States))
	for _, state := range inventory.States {
		states[state.Value] = true
	}
	for _, want := range []string{
		"textinput/Config.Disabled",
		"textinput/Config.Required",
		"textinput/Config.Readonly",
		"checkbox/Config.Checked",
		"checkbox/Config.Disabled",
	} {
		if !states[want] {
			t.Errorf("source-visible public configuration state %s missing from inventory", want)
		}
	}
}

func TestDeriveLifecycleStatesRejectsAuthorityWithoutRealActionContract(t *testing.T) {
	original := lifecycleStateAuthorities[0].Action
	lifecycleStateAuthorities[0].Action = ""
	t.Cleanup(func() { lifecycleStateAuthorities[0].Action = original })
	if _, err := deriveLifecycleStates(repoRoot(t)); err == nil || !strings.Contains(err.Error(), "lacks action contract") {
		t.Fatalf("missing lifecycle action contract error = %v", err)
	}
}

func TestValidateRejectsIncompleteMandatoryRows(t *testing.T) {
	required := Inventory{
		Packages:        []SourceItem{{Value: "github.com/araihu/goshtoso/components/button", Source: SourceRef{Path: "components/button", Symbol: "package button"}}},
		Renderables:     []SourceItem{{Value: "button.Button", Source: SourceRef{Path: "components/button/component.go", Symbol: "Button"}}},
		Kinds:           []SourceItem{{Value: "button", Source: SourceRef{Path: "components/component.go", Symbol: "KindButton"}}},
		Routes:          []SourceItem{{Value: "/components/button", Source: SourceRef{Path: "site/internal/pages/catalog/catalog.go", Symbol: "components/button"}}},
		Themes:          []SourceItem{{Value: "araihu", Source: SourceRef{Path: "all-themes.css", Symbol: "[data-theme=araihu]"}}},
		BreakpointEdges: []int{640},
	}

	err := Validate(Ledger{SchemaVersion: SchemaVersion}, required)
	if err == nil {
		t.Fatal("incomplete ledger unexpectedly valid")
	}
	for _, missing := range []string{"package", "renderable", "kind", "route", "state", "theme", "mode", "breakpoint", "zoom", "motion", "input", "AT"} {
		if !contains(err.Error(), missing) {
			t.Errorf("error %q does not report missing %s coverage", err, missing)
		}
	}
}

func TestValidateRequiresProvenanceApplicabilityAndReceiptStatus(t *testing.T) {
	required := Inventory{}
	row := Row{ID: "button/default"}

	err := Validate(Ledger{SchemaVersion: SchemaVersion, Rows: []Row{row}}, required)
	if err == nil {
		t.Fatal("row without provenance, applicability, and receipt status unexpectedly valid")
	}
	for _, missing := range []string{"source path", "source symbol", "applicability", "receipt status"} {
		if !contains(err.Error(), missing) {
			t.Errorf("error %q does not report missing %s", err, missing)
		}
	}
}

func TestValidateRejectsMandatoryExecutionRowMarkedNotApplicable(t *testing.T) {
	ledger, inventory, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for index := range ledger.Rows {
		row := &ledger.Rows[index]
		if row.ID != "execution/kind/button" {
			continue
		}
		row.Applicability = NotApplicable
		row.ReceiptStatus = StatusNotApplicable
		row.Rationale = "claimant supplied N/A"
		if err := Validate(ledger, inventory); err == nil || !strings.Contains(err.Error(), "mandatory execution row cannot be N/A") {
			t.Fatalf("mandatory direct N/A error = %v", err)
		}
		return
	}
	t.Fatal("mandatory execution row fixture missing")
}

func TestRequiredViewportWidthsIncludeEachCompiledEdgeAndIntermediate(t *testing.T) {
	got := RequiredViewportWidths([]int{640, 768, 1024, 1280, 1536})
	want := []int{390, 639, 640, 641, 704, 767, 768, 769, 896, 1023, 1024, 1025, 1152, 1279, 1280, 1281, 1408, 1440, 1535, 1536, 1537}
	if !equalInts(got, want) {
		t.Fatalf("required viewport widths = %v, want %v", got, want)
	}
}

func TestValidateRejectsApplicableNotApplicableStatus(t *testing.T) {
	row := Row{
		ID:            "execution/kind/button",
		Class:         ClassExecution,
		Sources:       []SourceRef{{Path: "components/component.go", Symbol: "KindButton"}},
		Applicability: Applicable,
		ReceiptStatus: StatusNotApplicable,
		Receipt:       "receipt.log",
		Rationale:     "invalid mixed state",
	}
	if err := Validate(Ledger{SchemaVersion: SchemaVersion, Rows: []Row{row}}, Inventory{}); err == nil || !contains(err.Error(), "applicable row cannot use not-applicable receipt status") {
		t.Fatalf("mixed applicability error = %v", err)
	}
}

func assertLen[T any](t *testing.T, name string, values []T, want int) {
	t.Helper()
	if len(values) != want {
		t.Fatalf("%s count = %d, want %d", name, len(values), want)
	}
}

func equalInts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func contains(value, substring string) bool {
	for index := 0; index+len(substring) <= len(value); index++ {
		if value[index:index+len(substring)] == substring {
			return true
		}
	}
	return false
}

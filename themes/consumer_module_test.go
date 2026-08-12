package themes_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

func TestPublicCatalogCompilesInStandaloneModule(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	fixture := t.TempDir()

	goMod := fmt.Sprintf(`module example.com/theme-consumer

go 1.26.5

require github.com/araihu/goshtoso v0.0.0

replace github.com/araihu/goshtoso => %s
`, strconv.Quote(repoRoot))
	consumerTest := `package consumer_test

import (
	"testing"

	"github.com/araihu/goshtoso/themes"
)

func TestCatalog(t *testing.T) {
	var key themes.Key = themes.KeyGoshtoso
	catalog := themes.BuiltIn()
	want := []themes.Theme{
		{Key: themes.Key90s, Label: "90s", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyAraiHu, Label: "Arai Hû", Ownership: themes.OwnershipOrganization},
		{Key: themes.KeyArctic, Label: "Arctic", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyChristmas, Label: "Christmas", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyDracula, Label: "Dracula", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyGoshtoso, Label: "Goshtoso", Ownership: themes.OwnershipOrganization},
		{Key: themes.KeyHalloween, Label: "Halloween", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyHighContrast, Label: "High Contrast", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyIndustrial, Label: "Industrial", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyMinimal, Label: "Minimal", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyModern, Label: "Modern", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyNeoBrutalism, Label: "Neo Brutalism", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyNews, Label: "News", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyPastel, Label: "Pastel", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyPrototype, Label: "Prototype", Ownership: themes.OwnershipGeneric},
		{Key: themes.KeyZombie, Label: "Zombie", Ownership: themes.OwnershipGeneric},
	}
	if key == "" || len(catalog) != len(want) {
		t.Fatal("public catalog contract unavailable")
	}
	for i := range want {
		if catalog[i] != want[i] {
			t.Fatalf("catalog[%d] = %#v, want %#v", i, catalog[i], want[i])
		}
	}
}
`
	if err := os.WriteFile(filepath.Join(fixture, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture, "catalog_test.go"), []byte(consumerTest), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "test", "./...", "-count=1")
	cmd.Dir = fixture
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("standalone consumer compile failed: %v\n%s", err, output)
	}
}

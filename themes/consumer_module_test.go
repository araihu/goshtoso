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

go 1.27.0

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
	if key == "" || len(catalog) == 0 || catalog[0].Label == "" {
		t.Fatal("public catalog contract unavailable")
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

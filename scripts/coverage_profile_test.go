package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFilterAuthoredCoverageExcludesGeneratedTempl(t *testing.T) {
	t.Parallel()

	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	filter := filepath.Join(repoRoot, "scripts", "filter-authored-coverage")

	fixture := t.TempDir()
	input := filepath.Join(fixture, "coverage.out")
	output := filepath.Join(fixture, "coverage-authored.out")
	writeFile(t, input, `mode: atomic
github.com/araihu/goshtoso/components/button/button_templ.go:10.1,12.2 2 1
github.com/araihu/goshtoso/components/button/types.go:20.1,24.2 3 1
github.com/araihu/goshtoso/components/card/card_templ.go:30.1,36.2 5 0
github.com/araihu/goshtoso/components/card/types.go:40.1,48.2 7 0
`)

	command := exec.Command(filter, input, output)
	if combined, runErr := command.CombinedOutput(); runErr != nil {
		t.Fatalf("filter authored coverage: %v\n%s", runErr, combined)
	}

	want := `mode: atomic
github.com/araihu/goshtoso/components/button/types.go:20.1,24.2 3 1
github.com/araihu/goshtoso/components/card/types.go:40.1,48.2 7 0
`
	assertFileContent(t, output, want)
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s content mismatch\nwant:\n%s\ngot:\n%s", path, want, got)
	}
}

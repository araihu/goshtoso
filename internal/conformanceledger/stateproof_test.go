package conformanceledger

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEverySourceDerivedStateHasExecutableTestReference(t *testing.T) {
	inventory, err := DeriveInventory(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}

	var testSource strings.Builder
	for _, root := range []string{"components", "site/tests/e2e"} {
		err := filepath.WalkDir(filepath.Join(repoRoot(t), root), func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, "_test.go") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			testSource.Write(content)
			return nil
		})
		if err != nil {
			t.Fatalf("read executable state proof sources: %v", err)
		}
	}

	var missing []string
	for _, state := range inventory.States {
		if !strings.Contains(testSource.String(), state.Source.Symbol) {
			missing = append(missing, state.Value+" ("+state.Source.Path+":"+state.Source.Symbol+")")
		}
	}
	if len(missing) != 0 {
		t.Fatalf("source-derived states without executable test reference (%d): %s", len(missing), strings.Join(missing, ", "))
	}
}

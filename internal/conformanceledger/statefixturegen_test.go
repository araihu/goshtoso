package conformanceledger

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateStateFixtureSourceCoversEveryStateExactlyOnce(t *testing.T) {
	source, err := GenerateStateFixtureSource(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := DeriveInventory(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range inventory.States {
		needle := []byte(`State: "` + state.Value + `"`)
		if count := bytes.Count(source, needle); count != 1 {
			t.Fatalf("generated state %s count = %d, want 1", state.Value, count)
		}
	}
}

func TestGenerateStateFixtureSourceFailsClosedForUncontractedPublicConfigStates(t *testing.T) {
	_, err := GenerateStateFixtureSource(repoRoot(t))
	if err == nil || !strings.Contains(err.Error(), "actiongroup.Config.Label") || !strings.Contains(err.Error(), "alert.Config.Dismissible") {
		t.Fatalf("uncontracted public configuration state error = %v", err)
	}
}

func TestGeneratedStateFixtureTrackedSourceMatchesGenerator(t *testing.T) {
	repo := repoRoot(t)
	want, err := GenerateStateFixtureSource(repo)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repo, "site", "tests", "e2e", "conformance_state_fixtures_generated_support_test.go")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("generated state fixture drift: %s must byte-match GenerateStateFixtureSource output", path)
	}
}

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

func TestGenerateStateFixtureSourceUsesMaintainedPublicConfigContracts(t *testing.T) {
	source, err := GenerateStateFixtureSource(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{
		"textinput/Config.Disabled",
		"textinput/Config.Required",
		"textinput/Config.Readonly",
		"checkbox/Config.Checked",
		"checkbox/Config.Disabled",
	} {
		if !strings.Contains(string(source), `State: "`+state+`"`) {
			t.Errorf("generated fixtures omit maintained configuration state %s", state)
		}
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

func TestGeneratedSiteLedgerTrackedSourceMatchesGenerator(t *testing.T) {
	repo := repoRoot(t)
	want, err := GenerateSiteLedgerSource(repo)
	if err != nil {
		t.Fatal(err)
	}
	for relative, source := range want {
		path := filepath.Join(repo, "site", "internal", "conformanceledger", relative)
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read generated site ledger %s: %v", path, err)
		}
		if !bytes.Equal(got, source) {
			t.Fatalf("generated site ledger drift: %s must byte-match GenerateSiteLedgerSource output", path)
		}
	}
}

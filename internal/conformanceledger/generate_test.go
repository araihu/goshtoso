package conformanceledger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSkeletonAuthenticatesEveryPriorReceipt(t *testing.T) {
	config := generationFixture(t)
	config.Receipts = config.Receipts[1:]
	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "missing mandatory receipt") {
		t.Fatalf("missing receipt error = %v", err)
	}
}

func TestRequiredReceiptTasksExactlyMatchTGS010RoadmapPrerequisites(t *testing.T) {
	want := []string{
		"T-GS-008",
		"T-GS-011",
		"T-GS-012",
		"T-GS-013",
		"T-GS-014",
		"T-GS-017",
		"T-GS-018",
		"T-GS-019",
		"T-GS-020",
		"T-GS-021",
		"T-GS-022",
	}
	if strings.Join(RequiredReceiptTasks, ",") != strings.Join(want, ",") {
		t.Fatalf("T-GS-010 prerequisite tasks = %v, want %v", RequiredReceiptTasks, want)
	}
}

func TestGenerateSkeletonRejectsOpaquePriorReceiptEvenWhenHashMatches(t *testing.T) {
	config := generationFixture(t)
	content := []byte("receipt T-GS-008\n")
	if err := os.WriteFile(config.Receipts[0].Path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	config.Receipts[0].ExpectedSHA256 = hex.EncodeToString(digest[:])

	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "prerequisite receipt") {
		t.Fatalf("opaque receipt error = %v", err)
	}
}

func TestGenerateSkeletonRejectsAliasedPrerequisiteReceiptIdentity(t *testing.T) {
	config := generationFixture(t)
	rewritePrerequisiteReceipt(t, &config.Receipts[1], func(receipt map[string]any) {
		receipt["receipt_id"] = "receipt-" + RequiredReceiptTasks[0]
	})
	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "aliased prerequisite receipt") {
		t.Fatalf("aliased receipt error = %v", err)
	}
}

func TestGenerateSkeletonRejectsStalePrerequisiteCandidateIdentity(t *testing.T) {
	config := generationFixture(t)
	rewritePrerequisiteReceipt(t, &config.Receipts[0], func(receipt map[string]any) {
		receipt["candidate"].(map[string]any)["commit"] = strings.Repeat("0", 40)
	})
	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "candidate commit") {
		t.Fatalf("stale candidate error = %v", err)
	}
}

func TestGenerateSkeletonRejectsNonAcceptedPrerequisiteDisposition(t *testing.T) {
	config := generationFixture(t)
	rewritePrerequisiteReceipt(t, &config.Receipts[0], func(receipt map[string]any) {
		receipt["disposition"] = "failed"
	})
	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "accepted disposition") {
		t.Fatalf("disposition error = %v", err)
	}
}

func TestGenerateSkeletonRejectsPrerequisiteWithoutGateEvidence(t *testing.T) {
	config := generationFixture(t)
	rewritePrerequisiteReceipt(t, &config.Receipts[0], func(receipt map[string]any) {
		receipt["gate_evidence"] = []any{}
	})
	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "gate evidence") {
		t.Fatalf("gate evidence error = %v", err)
	}
}

func TestGenerateSkeletonRejectsUnapprovedPrerequisiteProvenance(t *testing.T) {
	config := generationFixture(t)
	rewritePrerequisiteReceipt(t, &config.Receipts[0], func(receipt map[string]any) {
		receipt["provenance"] = "claimant"
	})
	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "authorized provenance") {
		t.Fatalf("provenance error = %v", err)
	}
}

func TestGenerateSkeletonRejectsTamperedReceipt(t *testing.T) {
	config := generationFixture(t)
	config.Receipts[0].ExpectedSHA256 = strings.Repeat("0", 64)
	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("tampered receipt error = %v", err)
	}
}

func TestGenerateSkeletonRejectsSourceIdentityMismatch(t *testing.T) {
	config := generationFixture(t)
	config.SourceTree = strings.Repeat("0", 40)
	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "source tree") {
		t.Fatalf("identity error = %v", err)
	}
}

func TestGenerateSkeletonIsCompleteButCannotCloseBlockedExecutionRows(t *testing.T) {
	config := generationFixture(t)
	ledger, inventory, err := GenerateSkeleton(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(ledger, inventory); err != nil {
		t.Fatalf("structural validation: %v", err)
	}
	if err := ValidateClosure(ledger, inventory); err == nil || !strings.Contains(err.Error(), "at/safari-voiceover/public=blocked") {
		t.Fatalf("closure error = %v", err)
	}
	if ledger.Metadata["current.kind_count"] != "84" || ledger.Metadata["historical.kind_count"] != "83" {
		t.Fatalf("kind drift metadata = %#v", ledger.Metadata)
	}
}

func generationFixture(t *testing.T) GenerationConfig {
	t.Helper()
	directory := t.TempDir()
	commit := fixtureGitIdentity(t, "HEAD^{commit}")
	tree := fixtureGitIdentity(t, "HEAD^{tree}")
	receipts := make([]ReceiptInput, 0, len(RequiredReceiptTasks))
	for _, task := range RequiredReceiptTasks {
		path := filepath.Join(directory, task+".md")
		content, err := json.Marshal(map[string]any{
			"schema_version": 1,
			"task":           task,
			"disposition":    "accepted",
			"candidate": map[string]any{
				"commit": commit,
				"tree":   tree,
			},
			"report": map[string]any{"sha256": strings.Repeat("a", 64)},
			"gate_evidence": []any{map[string]any{
				"name":   "independent-review",
				"tool":   "reviewer",
				"sha256": strings.Repeat("b", 64),
			}},
			"provenance": "independent-review",
			"receipt_id": "receipt-" + task,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(content)
		receipts = append(receipts, ReceiptInput{Task: task, Path: path, ExpectedSHA256: hex.EncodeToString(digest[:])})
	}
	return GenerationConfig{
		RepoRoot:         repoRoot(t),
		SourceCommit:     commit,
		SourceTree:       tree,
		Receipts:         receipts,
		ATBlockerReceipt: "AT capability receipt",
	}
}

func rewritePrerequisiteReceipt(t *testing.T, input *ReceiptInput, mutate func(map[string]any)) {
	t.Helper()
	content, err := os.ReadFile(input.Path)
	if err != nil {
		t.Fatal(err)
	}
	var receipt map[string]any
	if err := json.Unmarshal(content, &receipt); err != nil {
		t.Fatal(err)
	}
	mutate(receipt)
	content, err = json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input.Path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	input.ExpectedSHA256 = hex.EncodeToString(digest[:])
}

package conformanceledger

import (
	"crypto/sha256"
	"encoding/hex"
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
	receipts := make([]ReceiptInput, 0, len(RequiredReceiptTasks))
	for _, task := range RequiredReceiptTasks {
		path := filepath.Join(directory, task+".md")
		content := []byte("receipt " + task + "\n")
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(content)
		receipts = append(receipts, ReceiptInput{Task: task, Path: path, ExpectedSHA256: hex.EncodeToString(digest[:])})
	}
	return GenerationConfig{
		RepoRoot:         repoRoot(t),
		SourceCommit:     fixtureGitIdentity(t, "HEAD^{commit}"),
		SourceTree:       fixtureGitIdentity(t, "HEAD^{tree}"),
		Receipts:         receipts,
		ATBlockerReceipt: "AT capability receipt",
	}
}

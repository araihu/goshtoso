package conformanceledger

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyExecutionReceiptsAuthenticatesAndUpdatesExactRows(t *testing.T) {
	ledger, _, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	receipt := executionFixture(t, "root", []string{"execution/kind/button"}, StatusExecuted)
	if err := ApplyExecutionReceipts(&ledger, []ExecutionReceipt{receipt}); err != nil {
		t.Fatal(err)
	}
	for _, row := range ledger.Rows {
		if row.ID == "execution/kind/button" {
			if row.ReceiptStatus != StatusExecuted || row.EvidenceHashes["receipt_sha256"] != receipt.ExpectedSHA256 {
				t.Fatalf("updated row = %#v", row)
			}
			return
		}
	}
	t.Fatal("updated row not found")
}

func TestApplyExecutionReceiptsRejectsUnknownDuplicateAndTamperedMappings(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*ExecutionReceipt)
		wantErr string
	}{
		{name: "unknown", mutate: func(receipt *ExecutionReceipt) { receipt.RowIDs = []string{"execution/kind/not-real"} }, wantErr: "unknown row"},
		{name: "inventory", mutate: func(receipt *ExecutionReceipt) { receipt.RowIDs = []string{"inventory/kind/button"} }, wantErr: "non-execution row"},
		{name: "tampered", mutate: func(receipt *ExecutionReceipt) { receipt.ExpectedSHA256 = strings.Repeat("0", 64) }, wantErr: "SHA-256"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ledger, _, err := GenerateSkeleton(generationFixture(t))
			if err != nil {
				t.Fatal(err)
			}
			receipt := executionFixture(t, test.name, []string{"execution/kind/button"}, StatusExecuted)
			test.mutate(&receipt)
			if err := ApplyExecutionReceipts(&ledger, []ExecutionReceipt{receipt}); err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("error = %v, want %q", err, test.wantErr)
			}
		})
	}

	ledger, _, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	first := executionFixture(t, "first", []string{"execution/kind/button"}, StatusExecuted)
	second := executionFixture(t, "second", []string{"execution/kind/button"}, StatusExecuted)
	if err := ApplyExecutionReceipts(&ledger, []ExecutionReceipt{first, second}); err == nil || !strings.Contains(err.Error(), "mapped by both") {
		t.Fatalf("duplicate mapping error = %v", err)
	}
}

func TestReadAndApplyExecutionReceiptEnvelopeAuthenticatesWrapperFields(t *testing.T) {
	ledger, _, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	receipt := executionFixture(t, "envelope", []string{"execution/kind/button"}, StatusExecuted)
	envelope := ExecutionReceiptEnvelope{SchemaVersion: SchemaVersion, SourceCommit: ledger.SourceCommit, SourceTree: ledger.SourceTree, Receipts: []ExecutionReceipt{receipt}}
	content, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "execution-envelope.json")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	if err := ReadAndApplyExecutionReceiptEnvelope(&ledger, path, hex.EncodeToString(digest[:])); err != nil {
		t.Fatal(err)
	}

	ledger, _, err = GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	envelope.Receipts[0].Status = StatusNotApplicable
	content, err = json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ReadAndApplyExecutionReceiptEnvelope(&ledger, path, hex.EncodeToString(digest[:])); err == nil || !strings.Contains(err.Error(), "envelope SHA-256") {
		t.Fatalf("wrapper tamper error = %v", err)
	}
}

func TestApplyExecutionReceiptsRejectsMissingIdentityContext(t *testing.T) {
	ledger, _, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	receipt := executionFixture(t, "missing-context", []string{"execution/kind/button"}, StatusExecuted)
	receipt.Context.DependencyPins = nil
	if err := ApplyExecutionReceipts(&ledger, []ExecutionReceipt{receipt}); err == nil || !strings.Contains(err.Error(), "dependency pins") {
		t.Fatalf("missing context error = %v", err)
	}
}

func TestApplyExecutionReceiptsRequiresAuthenticatedATArtifactMetadata(t *testing.T) {
	ledger, _, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	receipt := executionFixture(t, "at", []string{"at/safari-voiceover/public"}, StatusExecuted)
	artifactPath := filepath.Join(t.TempDir(), "caption.png")
	content := []byte("caption evidence")
	if err := os.WriteFile(artifactPath, content, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	receipt.Artifacts = []EvidenceArtifact{{Kind: "at-caption-screenshot", Path: artifactPath, ExpectedSHA256: hex.EncodeToString(digest[:])}}
	if err := ApplyExecutionReceipts(&ledger, []ExecutionReceipt{receipt}); err == nil || !strings.Contains(err.Error(), "requires route, state, browser, and AT version") {
		t.Fatalf("AT metadata error = %v", err)
	}
}

func TestApplyExecutionReceiptsRequiresExactATArtifactPairAndIdentity(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*ExecutionReceipt)
		wantErr string
	}{
		{name: "zero artifacts", mutate: func(receipt *ExecutionReceipt) { receipt.Artifacts = nil }, wantErr: "exactly one at-caption-screenshot and one at-cursor-trace"},
		{name: "missing trace", mutate: func(receipt *ExecutionReceipt) { receipt.Artifacts = receipt.Artifacts[:1] }, wantErr: "exactly one at-caption-screenshot and one at-cursor-trace"},
		{name: "wrong route", mutate: func(receipt *ExecutionReceipt) { receipt.Artifacts[0].Route = "/wrong" }, wantErr: "does not match row"},
		{name: "wrong receipt context", mutate: func(receipt *ExecutionReceipt) { receipt.Context.Route = "/wrong" }, wantErr: "AT context"},
		{name: "wrong browser", mutate: func(receipt *ExecutionReceipt) { receipt.Artifacts[0].Browser = "Wrong Browser 1" }, wantErr: "does not match tool browser"},
		{name: "wrong AT", mutate: func(receipt *ExecutionReceipt) { receipt.Artifacts[0].ATVersion = "Wrong AT 1" }, wantErr: "does not match tool screen reader"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ledger, _, err := GenerateSkeleton(generationFixture(t))
			if err != nil {
				t.Fatal(err)
			}
			receipt := executionFixture(t, "at-exact", []string{"at/safari-voiceover/public"}, StatusExecuted)
			receipt.Context.AT = "safari-voiceover"
			receipt.Context.Route = "/getting-started"
			receipt.Context.State = "initial"
			receipt.ToolVersions = map[string]string{"browser": "Safari 26.5.2", "screen_reader": "VoiceOver bundle 10"}
			receipt.Artifacts = []EvidenceArtifact{
				evidenceArtifactFixture(t, "at-caption-screenshot", "/getting-started", "initial", "Safari 26.5.2", "VoiceOver bundle 10"),
				evidenceArtifactFixture(t, "at-cursor-trace", "/getting-started", "initial", "Safari 26.5.2", "VoiceOver bundle 10"),
			}
			test.mutate(&receipt)
			if err := ApplyExecutionReceipts(&ledger, []ExecutionReceipt{receipt}); err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("AT artifact error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestApplyExecutionReceiptsRejectsForgedATPNG(t *testing.T) {
	ledger, _, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	receipt := executionFixture(t, "forged-png", []string{"at/safari-voiceover/public"}, StatusExecuted)
	receipt.Context.AT = "safari-voiceover"
	receipt.Context.Route = "/getting-started"
	receipt.Context.State = "initial"
	receipt.ToolVersions = map[string]string{"browser": "Safari 26.5.2", "screen_reader": "VoiceOver bundle 10"}
	caption := evidenceArtifactFixture(t, "at-caption-screenshot", "/getting-started", "initial", "Safari 26.5.2", "VoiceOver bundle 10")
	forged := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}
	if err := os.WriteFile(caption.Path, forged, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(forged)
	caption.ExpectedSHA256 = hex.EncodeToString(digest[:])
	receipt.Artifacts = []EvidenceArtifact{
		caption,
		evidenceArtifactFixture(t, "at-cursor-trace", "/getting-started", "initial", "Safari 26.5.2", "VoiceOver bundle 10"),
	}
	if err := ApplyExecutionReceipts(&ledger, []ExecutionReceipt{receipt}); err == nil || !strings.Contains(err.Error(), "decodable non-empty PNG") {
		t.Fatalf("forged PNG error = %v", err)
	}
}

func TestApplyExecutionReceiptsRejectsClaimantNotApplicableForMandatoryExecutionRow(t *testing.T) {
	ledger, _, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	receipt := executionFixture(t, "not-applicable", []string{"execution/kind/button"}, StatusNotApplicable)
	receipt.Rationale = "static fixture has no runtime motion"
	if err := ApplyExecutionReceipts(&ledger, []ExecutionReceipt{receipt}); err == nil || !strings.Contains(err.Error(), "mandatory execution row") {
		t.Fatalf("claimant N/A error = %v", err)
	}
}

func TestApplyExecutionReceiptsRejectsDuplicateReceiptIDs(t *testing.T) {
	ledger, _, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	first := executionFixture(t, "duplicate", []string{"execution/kind/button"}, StatusExecuted)
	second := executionFixture(t, "duplicate", []string{"execution/kind/accordion"}, StatusExecuted)
	if err := ApplyExecutionReceipts(&ledger, []ExecutionReceipt{first, second}); err == nil || !strings.Contains(err.Error(), "duplicate execution receipt id") {
		t.Fatalf("duplicate receipt id error = %v", err)
	}
}

func TestApplyExecutionReceiptsCannotCloseStateRowsWithoutExactBFullManifest(t *testing.T) {
	ledger, _, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	receipt := executionFixture(t, "state-without-browser-matrix", []string{"execution/state/button_default"}, StatusExecuted)
	if err := ApplyExecutionReceipts(&ledger, []ExecutionReceipt{receipt}); err == nil || !strings.Contains(err.Error(), "requires exactly one b-full-state-matrix artifact") {
		t.Fatalf("state execution without B-FULL error = %v", err)
	}
}

func executionFixture(t *testing.T, id string, rows []string, status ReceiptStatus) ExecutionReceipt {
	t.Helper()
	path := filepath.Join(t.TempDir(), id+".log")
	content := []byte("execution " + id + "\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	commit := fixtureGitIdentity(t, "HEAD^{commit}")
	tree := fixtureGitIdentity(t, "HEAD^{tree}")
	return ExecutionReceipt{ID: id, Path: path, ExpectedSHA256: hex.EncodeToString(digest[:]), Status: status, RowIDs: rows, Context: ExecutionContext{RepoRoot: repoRoot(t), SourceCommit: commit, SourceTree: tree, DependencyPins: map[string]string{"github.com/araihu/goshtoso": "fixture"}}}
}

func evidenceArtifactFixture(t *testing.T, kind, route, state, browser, atVersion string) EvidenceArtifact {
	t.Helper()
	path := filepath.Join(t.TempDir(), strings.ReplaceAll(kind, "/", "-")+".evidence")
	content := []byte(kind + " evidence\n")
	if kind == "at-caption-screenshot" {
		var screenshot bytes.Buffer
		if err := png.Encode(&screenshot, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
			t.Fatal(err)
		}
		content = screenshot.Bytes()
	}
	if kind == "at-cursor-trace" {
		content, _ = json.Marshal(map[string]any{"route": route, "state": state, "browser": browser, "screen_reader": atVersion, "events": []map[string]string{{"event": "focus", "text": "State"}}})
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	return EvidenceArtifact{Kind: kind, Path: path, ExpectedSHA256: hex.EncodeToString(digest[:]), Route: route, State: state, Browser: browser, ATVersion: atVersion}
}

func fixtureGitIdentity(t *testing.T, revision string) string {
	t.Helper()
	command := exec.Command("git", "rev-parse", revision)
	command.Dir = repoRoot(t)
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(output))
}

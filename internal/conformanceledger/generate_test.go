package conformanceledger

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var fixturePrerequisiteSigner = func() ed25519.PrivateKey {
	seed := sha256.Sum256([]byte("T-GS-010 prerequisite fixture signer"))
	return ed25519.NewKeyFromSeed(seed[:])
}()

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

func TestGenerateSkeletonRejectsSelfMintedPrerequisiteHashAndProvenance(t *testing.T) {
	config := generationFixture(t)
	rewritePrerequisiteReceipt(t, &config.Receipts[0], func(receipt map[string]any) {
		receipt["attestation"] = map[string]any{"key_id": "claimant", "signature": ""}
	})
	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "trusted prerequisite signature") {
		t.Fatalf("self-minted prerequisite authority error = %v", err)
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

func TestGenerateSkeletonRejectsPrerequisiteWithoutBoundReportAndGateBytes(t *testing.T) {
	config := generationFixture(t)
	rewritePrerequisiteReceipt(t, &config.Receipts[0], func(receipt map[string]any) {
		receipt["report"].(map[string]any)["path"] = ""
	})
	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "report bytes") {
		t.Fatalf("unbound prerequisite report error = %v", err)
	}
}

func TestGenerateSkeletonRejectsTamperedBoundPrerequisiteBytes(t *testing.T) {
	config := generationFixture(t)
	content, err := os.ReadFile(config.Receipts[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	var receipt PrerequisiteReceipt
	if err := json.Unmarshal(content, &receipt); err != nil {
		t.Fatal(err)
	}
	requireWriteFile(t, receipt.Report.Path, []byte("claimant replacement report\n"))
	if _, _, err := GenerateSkeleton(config); err == nil || !strings.Contains(err.Error(), "bound report bytes SHA-256") {
		t.Fatalf("tampered bound report error = %v", err)
	}
}

func TestReviewerTrustIsSeparateFromATRecorderTrust(t *testing.T) {
	reviewers, err := loadConformanceReviewerTrustedKeys(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(reviewers) != 1 || len(reviewers["t-gs-010-reviewer-20260815"]) != ed25519.PublicKeySize {
		t.Fatalf("reviewer trust = %#v", reviewers)
	}
	if _, exists := reviewers["macos-voiceover-20260812"]; exists {
		t.Fatal("AT recorder key reused as prerequisite reviewer authority")
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
	delete(ledger.Metadata, "current.uncovered_config_contracts")
	delete(ledger.Metadata, "current.unverified_lifecycle_action_contracts") // Isolate the pre-existing blocked-AT closure assertion below.
	if err := ValidateClosure(ledger, inventory); err == nil || !strings.Contains(err.Error(), "at/safari-voiceover/public=blocked") {
		t.Fatalf("closure error = %v", err)
	}
	if ledger.Metadata["current.kind_count"] != "84" || ledger.Metadata["historical.kind_count"] != "83" {
		t.Fatalf("kind drift metadata = %#v", ledger.Metadata)
	}
}

func TestGenerateSkeletonRecordsUncoveredConfigContractsAndClosureFailsClosed(t *testing.T) {
	ledger, inventory, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	uncovered := ledger.Metadata["current.uncovered_config_contracts"]
	if !strings.Contains(uncovered, "skeleton.Config.Count") {
		t.Fatalf("uncovered Config contracts = %q, want skeleton.Config.Count", uncovered)
	}
	for index := range ledger.Rows {
		ledger.Rows[index].Applicability = Applicable
		ledger.Rows[index].ReceiptStatus = StatusExecuted
	}
	if err := ValidateClosure(ledger, inventory); err == nil || !strings.Contains(err.Error(), "uncovered source Config contracts") {
		t.Fatalf("uncovered Config contract closure error = %v", err)
	}
}

func TestGenerateSkeletonRecordsUnverifiedLifecycleActionContractsAndClosureFailsClosed(t *testing.T) {
	ledger, inventory, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	delete(ledger.Metadata, "current.uncovered_config_contracts")
	if lifecycle := ledger.Metadata["current.unverified_lifecycle_action_contracts"]; !strings.Contains(lifecycle, "button/lifecycle/loading") {
		t.Fatalf("unverified lifecycle action contracts = %q, want button/lifecycle/loading", lifecycle)
	}
	if err := ValidateClosure(ledger, inventory); err == nil || !strings.Contains(err.Error(), "lifecycle action/outcome artifacts unavailable") {
		t.Fatalf("unverified lifecycle action closure error = %v", err)
	}
}

func TestValidateRequiresExactInventoryExecutionBijection(t *testing.T) {
	ledger, inventory, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	value := inventory.Packages[0].Value
	ledger.Rows = withoutRow(ledger.Rows, "inventory/package/"+strings.ReplaceAll(value, "/", "_"))
	if err := Validate(ledger, inventory); err == nil || !strings.Contains(err.Error(), "inventory package "+value+" count = 0, want 1") {
		t.Fatalf("missing inventory bijection error = %v", err)
	}

	ledger, inventory, err = GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range ledger.Rows {
		if row.ID == "execution/package/"+strings.ReplaceAll(value, "/", "_") {
			duplicate := row
			duplicate.ID += "/duplicate"
			ledger.Rows = append(ledger.Rows, duplicate)
			break
		}
	}
	if err := Validate(ledger, inventory); err == nil || !strings.Contains(err.Error(), "execution package row ID") {
		t.Fatalf("duplicate execution bijection error = %v", err)
	}
}

func TestValidateRejectsAxisPrefixCrossAxisEscape(t *testing.T) {
	ledger, inventory, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	packageValue := inventory.Packages[0].Value
	for index := range ledger.Rows {
		row := &ledger.Rows[index]
		if row.ID != "execution/package/"+strings.ReplaceAll(packageValue, "/", "_") {
			continue
		}
		row.ID = "execution/package/" + strings.ReplaceAll(inventory.States[0].Value, "/", "_")
		if err := Validate(ledger, inventory); err == nil || !strings.Contains(err.Error(), "execution package row ID") {
			t.Fatalf("axis prefix escape error = %v", err)
		}
		return
	}
	t.Fatal("package execution row fixture missing")
}

func TestValidateRejectsExtraOrEmptySourceAxisRows(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*Ledger, Inventory)
		want   string
	}{
		{
			name: "extra source row outside canonical namespace",
			mutate: func(ledger *Ledger, inventory Inventory) {
				for _, row := range ledger.Rows {
					if row.ID == "inventory/package/"+strings.ReplaceAll(inventory.Packages[0].Value, "/", "_") {
						duplicate := row
						duplicate.ID = "claimant-row-outside-source-namespace"
						ledger.Rows = append(ledger.Rows, duplicate)
						return
					}
				}
				t.Fatal("inventory package row fixture missing")
			},
			want: "source-bearing row has noncanonical ID",
		},
		{
			name: "source axis ID without source discriminator",
			mutate: func(ledger *Ledger, inventory Inventory) {
				for _, row := range ledger.Rows {
					if row.ID == "inventory/package/"+strings.ReplaceAll(inventory.Packages[0].Value, "/", "_") {
						duplicate := row
						duplicate.ID = "inventory/package/"
						duplicate.Package = ""
						ledger.Rows = append(ledger.Rows, duplicate)
						return
					}
				}
				t.Fatal("inventory package row fixture missing")
			},
			want: "source axis row has no source discriminator",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ledger, inventory, err := GenerateSkeleton(generationFixture(t))
			if err != nil {
				t.Fatal(err)
			}
			test.mutate(&ledger, inventory)
			if err := Validate(ledger, inventory); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestValidateRejectsMultiAxisSourceRowAndChecklistEscape(t *testing.T) {
	ledger, inventory, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	packageValue := inventory.Packages[0].Value
	for index := range ledger.Rows {
		row := &ledger.Rows[index]
		if row.ID != "inventory/package/"+strings.ReplaceAll(packageValue, "/", "_") {
			continue
		}
		row.State = inventory.States[0].Value
		row.ChecklistURLs = []string{ChecklistA11Y}
		row.ChecklistMappings = []ChecklistMapping{{URL: ChecklistA11Y, Kind: ChecklistFoundation, Rationale: "forged route mapping"}}
		if err := Validate(ledger, inventory); err == nil || !strings.Contains(err.Error(), "projects unexpected state") || !strings.Contains(err.Error(), "non-route source row carries checklist") {
			t.Fatalf("multi-axis source row error = %v", err)
		}
		return
	}
	t.Fatal("package inventory row fixture missing")
}

func TestValidateRejectsUnexpectedSourceNamespaceRow(t *testing.T) {
	ledger, inventory, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range ledger.Rows {
		if row.ID != "inventory/package/"+strings.ReplaceAll(inventory.Packages[0].Value, "/", "_") {
			continue
		}
		row.ID = "inventory/unmapped/claimant"
		ledger.Rows = append(ledger.Rows, row)
		if err := Validate(ledger, inventory); err == nil || !strings.Contains(err.Error(), "source-bearing row has noncanonical ID") {
			t.Fatalf("unexpected source namespace row error = %v", err)
		}
		return
	}
	t.Fatal("inventory package row fixture missing")
}

func TestGenerateSkeletonPersistsChecklistMappingKindAndRationale(t *testing.T) {
	ledger, _, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range ledger.Rows {
		if row.Route == "" || row.Class != ClassInventory {
			continue
		}
		encoded, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := json.Unmarshal(encoded, &document); err != nil {
			t.Fatal(err)
		}
		mappings, ok := document["checklist_mappings"].([]any)
		if !ok || len(mappings) == 0 {
			t.Fatalf("route %s omitted structured checklist mappings: %s", row.Route, encoded)
		}
		for _, raw := range mappings {
			mapping, ok := raw.(map[string]any)
			if !ok || mapping["url"] == "" || mapping["kind"] == "" {
				t.Fatalf("route %s malformed checklist mapping %#v", row.Route, raw)
			}
			if mapping["kind"] != string(ChecklistExact) && mapping["rationale"] == "" {
				t.Fatalf("route %s non-exact mapping lacks rationale %#v", row.Route, raw)
			}
		}
	}
}

func TestGenerateSkeletonBindsStateActionContractsToSourceAndExecutionRows(t *testing.T) {
	ledger, _, err := GenerateSkeleton(generationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for id, expected := range map[string]string{
		"inventory/lifecycle-state/button_lifecycle_disabled": "focus and activation are absent",
		"execution/lifecycle-state/button_lifecycle_disabled": "native disabled control",
		"inventory/state/checkbox_Config.Checked":             "keyboard Space changes checked state",
		"execution/state/checkbox_Config.Checked":             "keyboard Space changes checked state",
	} {
		var found *Row
		for index := range ledger.Rows {
			if ledger.Rows[index].ID == id {
				found = &ledger.Rows[index]
				break
			}
		}
		if found == nil || !strings.Contains(found.Rationale, expected) {
			t.Fatalf("state row %s omits action contract: %#v", id, found)
		}
	}
}

func generationFixture(t *testing.T) GenerationConfig {
	t.Helper()
	directory := t.TempDir()
	commit := fixtureGitIdentity(t, "HEAD^{commit}")
	tree := fixtureGitIdentity(t, "HEAD^{tree}")
	public := fixturePrerequisiteSigner.Public().(ed25519.PublicKey)
	receipts := make([]ReceiptInput, 0, len(RequiredReceiptTasks))
	for _, task := range RequiredReceiptTasks {
		path := filepath.Join(directory, task+".md")
		reportPath := filepath.Join(directory, task+"-report.md")
		reportBytes := []byte("independent review report for " + task + "\ncommit=" + commit + "\ntree=" + tree + "\n")
		requireWriteFile(t, reportPath, reportBytes)
		reportDigest := sha256.Sum256(reportBytes)
		gatePath := filepath.Join(directory, task+"-gate.txt")
		gateBytes := []byte("reviewer gate for " + task + "\n")
		requireWriteFile(t, gatePath, gateBytes)
		gateDigest := sha256.Sum256(gateBytes)
		document := map[string]any{
			"schema_version": 1,
			"task":           task,
			"disposition":    "accepted",
			"candidate": map[string]any{
				"commit": commit,
				"tree":   tree,
			},
			"report": map[string]any{"path": reportPath, "sha256": hex.EncodeToString(reportDigest[:])},
			"gate_evidence": []any{map[string]any{
				"name":   "independent-review",
				"tool":   "reviewer",
				"path":   gatePath,
				"sha256": hex.EncodeToString(gateDigest[:]),
			}},
			"provenance": "independent-review",
			"receipt_id": "receipt-" + task,
		}
		unsigned, err := json.Marshal(document)
		if err != nil {
			t.Fatal(err)
		}
		var receipt PrerequisiteReceipt
		if err := json.Unmarshal(unsigned, &receipt); err != nil {
			t.Fatal(err)
		}
		payload, err := prerequisiteReceiptPayload(receipt)
		if err != nil {
			t.Fatal(err)
		}
		document["attestation"] = map[string]any{"key_id": "fixture-reviewer", "signature": base64.StdEncoding.EncodeToString(ed25519.Sign(fixturePrerequisiteSigner, payload))}
		content, err := json.Marshal(document)
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
		RepoRoot:                repoRoot(t),
		SourceCommit:            commit,
		SourceTree:              tree,
		Receipts:                receipts,
		ATBlockerReceipt:        "AT capability receipt",
		trustedPrerequisiteKeys: map[string]ed25519.PublicKey{"fixture-reviewer": public},
	}
}

func requireWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
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
	if attestation, ok := receipt["attestation"].(map[string]any); ok && attestation["key_id"] == "fixture-reviewer" {
		unsigned, err := json.Marshal(receipt)
		if err != nil {
			t.Fatal(err)
		}
		var typed PrerequisiteReceipt
		if err := json.Unmarshal(unsigned, &typed); err != nil {
			t.Fatal(err)
		}
		payload, err := prerequisiteReceiptPayload(typed)
		if err != nil {
			t.Fatal(err)
		}
		attestation["signature"] = base64.StdEncoding.EncodeToString(ed25519.Sign(fixturePrerequisiteSigner, payload))
	}
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

func withoutRow(rows []Row, id string) []Row {
	result := make([]Row, 0, len(rows))
	for _, row := range rows {
		if row.ID != id {
			result = append(result, row)
		}
	}
	return result
}

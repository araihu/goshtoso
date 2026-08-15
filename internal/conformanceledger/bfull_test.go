package conformanceledger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateBFullManifestRequiresExactCartesianSet(t *testing.T) {
	axes := BFullAxes{
		States:    []string{"button/default", "tooltip/click"},
		Themes:    []string{"araihu"},
		Modes:     []string{"light", "dark"},
		Viewports: []int{390},
		Zooms:     []int{100},
		Motions:   []string{"normal"},
		Inputs:    []string{"mouse", "keyboard"},
	}
	manifest := bfullFixture(t, axes)
	if err := ValidateBFullManifest(manifest, "commit", "tree", axes); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		mutate  func(*BFullManifest)
		wantErr string
	}{
		{name: "missing", mutate: func(got *BFullManifest) { got.Cells = got.Cells[1:] }, wantErr: "missing 1 mandatory cells"},
		{name: "duplicate", mutate: func(got *BFullManifest) { got.Cells[1] = got.Cells[0] }, wantErr: "duplicate cell"},
		{name: "extra", mutate: func(got *BFullManifest) { got.Cells[0].Theme = "extra" }, wantErr: "extra cell"},
		{name: "identity", mutate: func(got *BFullManifest) { got.SourceTree = "wrong" }, wantErr: "source identity"},
		{name: "diagnostic non-closure", mutate: func(got *BFullManifest) {
			got.Diagnostic = &BFullDiagnostic{NonClosure: true, StateCap: 1, Reason: "bounded feasibility slice"}
		}, wantErr: "cannot close execution rows"},
		{name: "missing batch", mutate: func(got *BFullManifest) { got.Batches = nil }, wantErr: "missing 2 mandatory batches"},
		{name: "unauthenticated batch", mutate: func(got *BFullManifest) { got.Batches[0].EvidenceSHA256 = strings.Repeat("b", 64) }, wantErr: "evidence SHA-256"},
		{name: "cell batch evidence", mutate: func(got *BFullManifest) { got.Cells[0].EvidenceSHA256 = strings.Repeat("b", 64) }, wantErr: "does not authenticate batch"},
		{name: "applicable N/A", mutate: func(got *BFullManifest) { got.Cells[0].ReceiptStatus = StatusNotApplicable }, wantErr: "must be executed"},
		{name: "N/A rationale", mutate: func(got *BFullManifest) {
			got.Cells[0].Applicability = NotApplicable
			got.Cells[0].ReceiptStatus = StatusNotApplicable
			got.Cells[0].Rationale = ""
		}, wantErr: "requires not-applicable status and rationale"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := manifest
			got.Cells = append([]BFullCell(nil), manifest.Cells...)
			got.Batches = append([]BFullBatch(nil), manifest.Batches...)
			test.mutate(&got)
			if err := ValidateBFullManifest(got, "commit", "tree", axes); err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func bfullFixture(t *testing.T, axes BFullAxes) BFullManifest {
	t.Helper()
	manifest := BFullManifest{SchemaVersion: SchemaVersion, SourceCommit: "commit", SourceTree: "tree", Identity: BFullIdentity{Mode: BFullIdentityCommitted}, Browser: "Chromium 151"}
	evidenceDir := t.TempDir()
	axeArchivePath := filepath.Join(evidenceDir, "axe-core-4.10.3.tgz")
	if err := os.WriteFile(axeArchivePath, []byte("fixture axe-core archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	axeArchiveSHA256, err := SHA256File(axeArchivePath)
	if err != nil {
		t.Fatal(err)
	}
	initialHTMLPath := filepath.Join(evidenceDir, "initial.html")
	if err := os.WriteFile(initialHTMLPath, []byte("<!doctype html><main>fixture</main>"), 0o600); err != nil {
		t.Fatal(err)
	}
	initialHTMLSHA256, err := SHA256File(initialHTMLPath)
	if err != nil {
		t.Fatal(err)
	}
	batchEvidence := map[string]string{}
	for _, theme := range axes.Themes {
		for _, mode := range axes.Modes {
			for _, viewport := range axes.Viewports {
				for _, zoom := range axes.Zooms {
					for _, motion := range axes.Motions {
						evidence := BFullBatchEvidence{
							SchemaVersion: SchemaVersion, SourceCommit: "commit", SourceTree: "tree", Browser: "Chromium 151",
							Identity: manifest.Identity,
							Theme:    theme, Mode: mode, Viewport: viewport, Zoom: zoom, Motion: motion,
							InitialHTMLPath: initialHTMLPath, InitialHTMLSHA256: initialHTMLSHA256,
							Accessibility: BFullAccessibilityScan{Engine: "axe-core", Version: AxeCoreVersion, ArchivePath: axeArchivePath, ArchiveSHA256: axeArchiveSHA256, ScriptSHA256: axeArchiveSHA256, Rules: []string{"button-name", "landmark-one-main"}, ScannedStates: len(axes.States), ChecklistResults: []BFullChecklistResult{{Criterion: "axe button-name", URL: ChecklistA11Y, RuleID: "button-name", Outcome: "pass"}, {Criterion: "axe landmark-one-main", URL: ChecklistA11Y, RuleID: "landmark-one-main", Outcome: "pass"}}},
						}
						for _, state := range axes.States {
							passed := BFullSemanticAssertion{Applicability: Applicable, Passed: true}
							observation := BFullStateObservation{State: state, Exists: true, DOMNodes: 1, Visible: true, Width: 10, Height: 10, Color: "rgb(0, 0, 0)", Background: "rgba(0, 0, 0, 0)", Theme: theme, Mode: mode, Motion: motion, Zoom: zoom, TextContrast: passed, BoundaryContrast: passed, MotionOutcome: passed, OverlayProvenance: passed}
							evidence.FirstPaint = append(evidence.FirstPaint, observation)
							evidence.Persistence = append(evidence.Persistence, observation)
							for _, input := range axes.Inputs {
								evidence.Inputs = append(evidence.Inputs, bfullInputFixture(state, input))
							}
						}
						evidencePath := filepath.Join(evidenceDir, strings.NewReplacer("|", "-", "/", "-").Replace(bfullBatchKey(theme, mode, viewport, zoom, motion))+".json")
						content, err := json.Marshal(evidence)
						if err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(evidencePath, content, 0o600); err != nil {
							t.Fatal(err)
						}
						evidenceSHA256, err := SHA256File(evidencePath)
						if err != nil {
							t.Fatal(err)
						}
						batch := BFullBatch{Theme: theme, Mode: mode, Viewport: viewport, Zoom: zoom, Motion: motion, EvidencePath: evidencePath, EvidenceSHA256: evidenceSHA256}
						manifest.Batches = append(manifest.Batches, batch)
						batchEvidence[bfullBatchKey(theme, mode, viewport, zoom, motion)] = batch.EvidenceSHA256
					}
				}
			}
		}
	}
	for _, state := range axes.States {
		for _, theme := range axes.Themes {
			for _, mode := range axes.Modes {
				for _, viewport := range axes.Viewports {
					for _, zoom := range axes.Zooms {
						for _, motion := range axes.Motions {
							for _, input := range axes.Inputs {
								manifest.Cells = append(manifest.Cells, BFullCell{State: state, Theme: theme, Mode: mode, Viewport: viewport, Zoom: zoom, Motion: motion, Input: input, Applicability: Applicable, ReceiptStatus: StatusExecuted, EvidenceSHA256: batchEvidence[bfullBatchKey(theme, mode, viewport, zoom, motion)]})
							}
						}
					}
				}
			}
		}
	}
	return manifest
}

func bfullInputFixture(state, input string) BFullInputObservation {
	notApplicable := func(rationale string) BFullSemanticAssertion {
		return BFullSemanticAssertion{Applicability: NotApplicable, Rationale: rationale}
	}
	passed := BFullSemanticAssertion{Applicability: Applicable, Passed: true}
	observation := BFullInputObservation{
		State: state, Input: input, Applicability: Applicable, ReceiptStatus: StatusExecuted,
		TargetSelector: "button", TargetRole: "button", AccessibleName: state,
		ARIAState: map[string]string{}, EventCount: 1, Driver: "fixture Playwright", SourceGrounding: "fixture source",
		Before: BFullInteractionOutcome{TargetConnected: true, TargetVisible: true}, Action: BFullInteractionOutcome{TargetConnected: true, TargetVisible: true, EventTypes: []string{"fixture"}}, Return: BFullInteractionOutcome{TargetConnected: true, TargetVisible: true},
		FocusVisible: notApplicable("not keyboard input"), MovementReturn: notApplicable("not pointer movement input"), Escape: BFullEscapeOutcome{Applicability: NotApplicable, Rationale: "not keyboard input"},
	}
	switch input {
	case "mouse":
		observation.MovementReturn = passed
	case "keyboard":
		observation.FocusVisible = passed
		observation.Escape = BFullEscapeOutcome{Applicability: Applicable, Passed: true, Opened: true, Closed: true, FocusReturned: true, SurfaceSelector: "[role=dialog]"}
	case "touch":
	default:
		panic("unsupported fixture input " + input)
	}
	return observation
}

func TestValidateBFullManifestRejectsStructuredEvidenceClaimsWithoutExecution(t *testing.T) {
	axes := BFullAxes{States: []string{"button/default"}, Themes: []string{"araihu"}, Modes: []string{"light"}, Viewports: []int{390}, Zooms: []int{100}, Motions: []string{"normal"}, Inputs: []string{"mouse", "keyboard", "touch"}}
	tests := []struct {
		name    string
		mutate  func(*BFullBatchEvidence)
		wantErr string
	}{
		{name: "accessibility violation", mutate: func(evidence *BFullBatchEvidence) { evidence.Accessibility.Violations = []string{"button has no name"} }, wantErr: "accessibility scan has 1 violations"},
		{name: "no-op input", mutate: func(evidence *BFullBatchEvidence) { evidence.Inputs[0].EventCount = 0 }, wantErr: "requires target, role, name, ARIA state, and event"},
		{name: "missing focus-visible", mutate: func(evidence *BFullBatchEvidence) { evidence.Inputs[1].FocusVisible.Passed = false }, wantErr: "focus-visible applicable assertion must pass"},
		{name: "duplicate input", mutate: func(evidence *BFullBatchEvidence) { evidence.Inputs = append(evidence.Inputs, evidence.Inputs[0]) }, wantErr: "duplicate input observation"},
		{name: "wrong state", mutate: func(evidence *BFullBatchEvidence) { evidence.FirstPaint[0].State = "invented" }, wantErr: "first paint has extra state"},
		{name: "browser error", mutate: func(evidence *BFullBatchEvidence) {
			evidence.Errors = []BFullBrowserError{{Kind: "csp", Message: "blocked"}}
		}, wantErr: "browser/page/network/CSP errors = 1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := bfullFixture(t, axes)
			rewriteBFullEvidence(t, &manifest.Batches[0], test.mutate)
			if err := ValidateBFullManifest(manifest, "commit", "tree", axes); err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func rewriteBFullEvidence(t *testing.T, batch *BFullBatch, mutate func(*BFullBatchEvidence)) {
	t.Helper()
	content, err := os.ReadFile(batch.EvidencePath)
	if err != nil {
		t.Fatal(err)
	}
	var evidence BFullBatchEvidence
	if err := json.Unmarshal(content, &evidence); err != nil {
		t.Fatal(err)
	}
	mutate(&evidence)
	content, err = json.Marshal(evidence)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(batch.EvidencePath, content, 0o600); err != nil {
		t.Fatal(err)
	}
	batch.EvidenceSHA256, err = SHA256File(batch.EvidencePath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExpectedBFullAxesIncludesSourceDerivedConsumerThemeAndExactCellCount(t *testing.T) {
	ledger := Ledger{}
	for _, theme := range []string{"araihu", "goshtoso", "minimal", "modern"} {
		ledger.Rows = append(ledger.Rows, Row{Class: ClassInventory, Theme: theme})
	}
	for index := 0; index < 347; index++ {
		ledger.Rows = append(ledger.Rows, Row{ID: "execution/state/" + string(rune(index+1)), Class: ClassExecution, State: string(rune(index + 1))})
	}
	for _, viewport := range []int{390, 639, 640, 641, 704, 767, 768, 769, 896, 1023, 1024, 1025, 1152, 1279, 1280, 1281, 1408, 1440, 1535, 1536, 1537} {
		ledger.Rows = append(ledger.Rows, Row{Class: ClassExecution, Viewport: viewport})
	}
	axes := ExpectedBFullAxes(ledger)
	if err := ValidateBFullAxesAgainstLedger(axes, ledger); err != nil {
		t.Fatal(err)
	}
	if got, want := axes.Themes, []string{"araihu", "goshtoso", "minimal", "modern"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("themes = %v, want %v", got, want)
	}
	if got := bfullCellCount(axes); got != 699552 {
		t.Fatalf("cell count = %d, want 699552", got)
	}

	axes.Themes = axes.Themes[:3]
	if err := ValidateBFullAxesAgainstLedger(axes, ledger); err == nil || !strings.Contains(err.Error(), "deep theme axis") {
		t.Fatalf("missing consumer theme error = %v", err)
	}
	axes.Themes = []string{"araihu", "goshtoso", "minimal", "invented"}
	if err := ValidateBFullAxesAgainstLedger(axes, ledger); err == nil || !strings.Contains(err.Error(), "deep theme axis") {
		t.Fatalf("non-source theme error = %v", err)
	}
}

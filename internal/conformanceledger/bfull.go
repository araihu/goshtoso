package conformanceledger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	BFullIdentityCommitted  = "committed"
	BFullIdentityDirtyBound = "dirty-bound"
	AxeCoreVersion          = "4.10.3"
	AxeCoreArchiveSHA256    = "0f2b4d7dcdf7d1219df8d1959ad68e565f51d14c3f0d88bb71cd59abeb956292"
)

type BFullAxes struct {
	States    []string `json:"states"`
	Themes    []string `json:"themes"`
	Modes     []string `json:"modes"`
	Viewports []int    `json:"viewports"`
	Zooms     []int    `json:"zooms"`
	Motions   []string `json:"motions"`
	Inputs    []string `json:"inputs"`
}

type BFullManifest struct {
	SchemaVersion int              `json:"schema_version"`
	SourceCommit  string           `json:"source_commit"`
	SourceTree    string           `json:"source_tree"`
	Identity      BFullIdentity    `json:"identity"`
	Diagnostic    *BFullDiagnostic `json:"diagnostic,omitempty"`
	Coverage      BFullCoverage    `json:"coverage"`
	Browser       string           `json:"browser"`
	Batches       []BFullBatch     `json:"batches"`
	Cells         []BFullCell      `json:"cells"`
}

// BFullDiagnostic makes a deliberately partial feasibility run impossible to
// confuse with closure evidence. ValidateBFullManifest rejects it outright.
type BFullDiagnostic struct {
	NonClosure bool   `json:"non_closure"`
	StateCap   int    `json:"state_cap"`
	Reason     string `json:"reason"`
}

// BFullIdentity distinguishes a final immutable candidate from a diagnostic
// taken while this evidence tooling is still dirty. Diagnostics may bind every
// changed byte, but they can never close execution rows.
type BFullIdentity struct {
	Mode                string `json:"mode"`
	RepoRoot            string `json:"repo_root,omitempty"`
	DirtyManifestPath   string `json:"dirty_manifest_path,omitempty"`
	DirtyManifestSHA256 string `json:"dirty_manifest_sha256,omitempty"`
}

type BFullDirtyManifest struct {
	SchemaVersion int               `json:"schema_version"`
	SourceCommit  string            `json:"source_commit"`
	SourceTree    string            `json:"source_tree"`
	Entries       []BFullDirtyEntry `json:"entries"`
}

type BFullDirtyEntry struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	SHA256 string `json:"sha256"`
}

// BFullCoverage closes source-derived axes that deliberately sit outside the
// 347-state Cartesian product: public packages/renderables/Kinds/routes and
// the maintained dynamic lifecycle authority. Values are source identifiers,
// never hand-picked labels.
type BFullCoverage struct {
	Packages        []string                     `json:"packages"`
	Renderables     []string                     `json:"renderables"`
	Kinds           []string                     `json:"kinds"`
	Routes          []string                     `json:"routes"`
	LifecycleStates []string                     `json:"lifecycle_states"`
	RouteEvidence   []BFullRouteObservation      `json:"route_evidence"`
	ATExemplars     []BFullATExemplarObservation `json:"at_exemplars"`
	ConsumerFixture BFullConsumerFixture         `json:"consumer_fixture"`
}

type BFullRouteObservation struct {
	Route                string `json:"route"`
	URL                  string `json:"url"`
	SourceResponsePath   string `json:"source_response_path"`
	SourceResponseSHA256 string `json:"source_response_sha256"`
	MainVisible          bool   `json:"main_visible"`
	MainOverflowX        bool   `json:"main_overflow_x"`
	Background           string `json:"background"`
	ConsumerOverrideID   string `json:"consumer_override_id,omitempty"`
}

// AT observations bind actual public/authenticated/data/form/navigation
// exemplars to the two required assistive-technology pairs. Artifacts are
// separately authenticated by execution receipts at closure time.
type BFullATExemplarObservation struct {
	AT      string `json:"at"`
	Name    string `json:"name"`
	Route   string `json:"route"`
	State   string `json:"state"`
	Browser string `json:"browser"`
}

type BFullConsumerFixture struct {
	ID               string `json:"id"`
	BaseTheme        string `json:"base_theme"`
	StylesheetPath   string `json:"stylesheet_path"`
	StylesheetSHA256 string `json:"stylesheet_sha256"`
}

type BFullBatch struct {
	Theme          string `json:"theme"`
	Mode           string `json:"mode"`
	Viewport       int    `json:"viewport"`
	Zoom           int    `json:"zoom"`
	Motion         string `json:"motion"`
	EvidencePath   string `json:"evidence_path"`
	EvidenceSHA256 string `json:"evidence_sha256"`
}

type BFullBatchEvidence struct {
	SchemaVersion     int                     `json:"schema_version"`
	SourceCommit      string                  `json:"source_commit"`
	SourceTree        string                  `json:"source_tree"`
	Identity          BFullIdentity           `json:"identity"`
	Browser           string                  `json:"browser"`
	Theme             string                  `json:"theme"`
	Mode              string                  `json:"mode"`
	Viewport          int                     `json:"viewport"`
	Zoom              int                     `json:"zoom"`
	Motion            string                  `json:"motion"`
	InitialHTMLPath   string                  `json:"initial_html_path"`
	InitialHTMLSHA256 string                  `json:"initial_html_sha256"`
	FirstPaint        []BFullStateObservation `json:"first_paint"`
	Persistence       []BFullStateObservation `json:"persistence"`
	Accessibility     BFullAccessibilityScan  `json:"accessibility"`
	Errors            []BFullBrowserError     `json:"errors"`
	Inputs            []BFullInputObservation `json:"inputs"`
}

type BFullStateObservation struct {
	State             string                 `json:"state"`
	Exists            bool                   `json:"exists"`
	DOMNodes          int                    `json:"dom_nodes"`
	Visible           bool                   `json:"visible"`
	Width             int                    `json:"width"`
	Height            int                    `json:"height"`
	Color             string                 `json:"color"`
	Background        string                 `json:"background"`
	OverflowX         bool                   `json:"overflow_x"`
	Theme             string                 `json:"theme"`
	Mode              string                 `json:"mode"`
	Motion            string                 `json:"motion"`
	Zoom              int                    `json:"zoom"`
	TextContrast      BFullSemanticAssertion `json:"text_contrast"`
	BoundaryContrast  BFullSemanticAssertion `json:"boundary_contrast"`
	MotionOutcome     BFullSemanticAssertion `json:"motion_outcome"`
	OverlayProvenance BFullSemanticAssertion `json:"overlay_provenance"`
}

type BFullAccessibilityScan struct {
	Engine           string                 `json:"engine"`
	Version          string                 `json:"version"`
	ArchivePath      string                 `json:"archive_path"`
	ArchiveSHA256    string                 `json:"archive_sha256"`
	ScriptSHA256     string                 `json:"script_sha256"`
	Rules            []string               `json:"rules"`
	ScannedStates    int                    `json:"scanned_states"`
	Violations       []string               `json:"violations"`
	Incomplete       []string               `json:"incomplete"`
	ChecklistResults []BFullChecklistResult `json:"checklist_results"`
}

type BFullChecklistResult struct {
	Criterion string `json:"criterion"`
	URL       string `json:"url"`
	RuleID    string `json:"rule_id"`
	Outcome   string `json:"outcome"`
	Targets   int    `json:"targets"`
}

type BFullBrowserError struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

type BFullInputObservation struct {
	State           string                  `json:"state"`
	Input           string                  `json:"input"`
	Applicability   Applicability           `json:"applicability"`
	ReceiptStatus   ReceiptStatus           `json:"receipt_status"`
	Rationale       string                  `json:"rationale,omitempty"`
	TargetSelector  string                  `json:"target_selector,omitempty"`
	TargetRole      string                  `json:"target_role,omitempty"`
	AccessibleName  string                  `json:"accessible_name,omitempty"`
	ARIAState       map[string]string       `json:"aria_state"`
	EventCount      int                     `json:"event_count,omitempty"`
	Driver          string                  `json:"driver,omitempty"`
	SourceGrounding string                  `json:"source_grounding,omitempty"`
	Before          BFullInteractionOutcome `json:"before"`
	Action          BFullInteractionOutcome `json:"action"`
	Return          BFullInteractionOutcome `json:"return"`
	FocusVisible    BFullSemanticAssertion  `json:"focus_visible"`
	MovementReturn  BFullSemanticAssertion  `json:"movement_return"`
	Escape          BFullEscapeOutcome      `json:"escape"`
}

// BFullInteractionOutcome preserves the observable before/action/return
// phases of a real browser input. EventTypes are listener observations only;
// they are never generated with dispatchEvent.
type BFullInteractionOutcome struct {
	TargetConnected bool              `json:"target_connected"`
	TargetVisible   bool              `json:"target_visible"`
	TargetFocused   bool              `json:"target_focused"`
	ActiveSelector  string            `json:"active_selector,omitempty"`
	ARIAState       map[string]string `json:"aria_state,omitempty"`
	Hovered         bool              `json:"hovered"`
	EventTypes      []string          `json:"event_types,omitempty"`
}

type BFullEscapeOutcome struct {
	Applicability   Applicability `json:"applicability"`
	Passed          bool          `json:"passed,omitempty"`
	Rationale       string        `json:"rationale,omitempty"`
	Opened          bool          `json:"opened,omitempty"`
	Closed          bool          `json:"closed,omitempty"`
	FocusReturned   bool          `json:"focus_returned,omitempty"`
	SurfaceSelector string        `json:"surface_selector,omitempty"`
}

type BFullSemanticAssertion struct {
	Applicability Applicability `json:"applicability"`
	Passed        bool          `json:"passed,omitempty"`
	Rationale     string        `json:"rationale,omitempty"`
}

type BFullCell struct {
	State          string        `json:"state"`
	Theme          string        `json:"theme"`
	Mode           string        `json:"mode"`
	Viewport       int           `json:"viewport"`
	Zoom           int           `json:"zoom"`
	Motion         string        `json:"motion"`
	Input          string        `json:"input"`
	Applicability  Applicability `json:"applicability"`
	ReceiptStatus  ReceiptStatus `json:"receipt_status"`
	Rationale      string        `json:"rationale,omitempty"`
	EvidenceSHA256 string        `json:"evidence_sha256,omitempty"`
}

func ExpectedBFullAxes(ledger Ledger) BFullAxes {
	axes := BFullAxes{
		Themes:  []string{"araihu", "goshtoso", "minimal", "modern"},
		Modes:   []string{"light", "dark"},
		Zooms:   []int{100, 200},
		Motions: []string{"normal", "reduced"},
		Inputs:  []string{"mouse", "keyboard", "touch"},
	}
	for _, row := range ledger.Rows {
		if row.Class == ClassExecution && strings.HasPrefix(row.ID, "execution/state/") {
			axes.States = append(axes.States, row.State)
		}
		if row.Class == ClassExecution && row.Viewport != 0 {
			axes.Viewports = append(axes.Viewports, row.Viewport)
		}
	}
	sort.Strings(axes.States)
	sort.Ints(axes.Viewports)
	return axes
}

func ExpectedBFullAxesFromInventory(inventory Inventory) (BFullAxes, error) {
	axes := BFullAxes{
		Themes: []string{"araihu", "goshtoso", "minimal", "modern"}, Modes: []string{"light", "dark"},
		Viewports: RequiredViewportWidths(inventory.BreakpointEdges), Zooms: []int{100, 200},
		Motions: []string{"normal", "reduced"}, Inputs: []string{"mouse", "keyboard", "touch"},
	}
	for _, state := range inventory.States {
		axes.States = append(axes.States, state.Value)
	}
	sort.Strings(axes.States)
	sourceThemes := make(map[string]struct{}, len(inventory.Themes))
	for _, theme := range inventory.Themes {
		sourceThemes[theme.Value] = struct{}{}
	}
	for _, theme := range axes.Themes {
		if _, ok := sourceThemes[theme]; !ok {
			return BFullAxes{}, fmt.Errorf("B-FULL theme %q is absent from source inventory", theme)
		}
	}
	return axes, nil
}

func ReadAndValidateBFullManifest(path string, ledger Ledger) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read B-FULL manifest: %w", err)
	}
	var manifest BFullManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return fmt.Errorf("parse B-FULL manifest: %w", err)
	}
	axes := ExpectedBFullAxes(ledger)
	if err := ValidateBFullAxesAgainstLedger(axes, ledger); err != nil {
		return err
	}
	if err := ValidateBFullManifest(manifest, ledger.SourceCommit, ledger.SourceTree, axes); err != nil {
		return err
	}
	if manifest.Identity.Mode != BFullIdentityCommitted || manifest.Diagnostic != nil {
		return fmt.Errorf("B-FULL execution closure requires a clean committed non-diagnostic identity")
	}
	return ValidateBFullCoverage(manifest.Coverage, requiredBFullCoverageFromLedger(ledger))
}

type BFullRequiredCoverage struct {
	Packages        []string
	Renderables     []string
	Kinds           []string
	Routes          []string
	LifecycleStates []string
}

func requiredBFullCoverageFromLedger(ledger Ledger) BFullRequiredCoverage {
	coverage := BFullRequiredCoverage{}
	for _, row := range ledger.Rows {
		if row.Class != ClassInventory {
			continue
		}
		switch {
		case row.Package != "":
			coverage.Packages = append(coverage.Packages, row.Package)
		case row.Renderable != "":
			coverage.Renderables = append(coverage.Renderables, row.Renderable)
		case row.Kind != "":
			coverage.Kinds = append(coverage.Kinds, row.Kind)
		case row.Route != "":
			coverage.Routes = append(coverage.Routes, row.Route)
		case strings.HasPrefix(row.ID, "inventory/lifecycle-state/"):
			coverage.LifecycleStates = append(coverage.LifecycleStates, row.State)
		}
	}
	for _, values := range [][]string{coverage.Packages, coverage.Renderables, coverage.Kinds, coverage.Routes, coverage.LifecycleStates} {
		sort.Strings(values)
	}
	return coverage
}

func RequiredBFullCoverageFromInventory(inventory Inventory) BFullRequiredCoverage {
	coverage := BFullRequiredCoverage{}
	for _, item := range inventory.Packages {
		coverage.Packages = append(coverage.Packages, item.Value)
	}
	for _, item := range inventory.Renderables {
		coverage.Renderables = append(coverage.Renderables, item.Value)
	}
	for _, item := range inventory.Kinds {
		coverage.Kinds = append(coverage.Kinds, item.Value)
	}
	for _, item := range inventory.Routes {
		coverage.Routes = append(coverage.Routes, item.Value)
	}
	for _, item := range inventory.LifecycleStates {
		coverage.LifecycleStates = append(coverage.LifecycleStates, item.Value)
	}
	for _, values := range [][]string{coverage.Packages, coverage.Renderables, coverage.Kinds, coverage.Routes, coverage.LifecycleStates} {
		sort.Strings(values)
	}
	return coverage
}

// ValidateBFullCoverage is separate from the state Cartesian validator so a
// package, Kind, route, renderable, lifecycle state, or AT exemplar cannot be
// silently closed by an unrelated state batch.
func ValidateBFullCoverage(coverage BFullCoverage, required BFullRequiredCoverage) error {
	for _, axis := range []struct {
		name string
		got  []string
		want []string
	}{
		{"package", coverage.Packages, required.Packages},
		{"renderable", coverage.Renderables, required.Renderables},
		{"Kind", coverage.Kinds, required.Kinds},
		{"route", coverage.Routes, required.Routes},
		{"lifecycle state", coverage.LifecycleStates, required.LifecycleStates},
	} {
		if err := validateExactBFullCoverageAxis(axis.name, axis.got, axis.want); err != nil {
			return err
		}
	}
	if len(coverage.RouteEvidence) != len(required.Routes) {
		return fmt.Errorf("B-FULL route evidence = %d, want %d", len(coverage.RouteEvidence), len(required.Routes))
	}
	routes := map[string]struct{}{}
	for _, observation := range coverage.RouteEvidence {
		if observation.Route == "" || observation.URL == "" || observation.SourceResponsePath == "" || !validSHA256(observation.SourceResponseSHA256) || !observation.MainVisible || observation.MainOverflowX || strings.TrimSpace(observation.Background) == "" {
			return fmt.Errorf("B-FULL route evidence for %q is incomplete", observation.Route)
		}
		got, err := SHA256File(observation.SourceResponsePath)
		if err != nil || got != observation.SourceResponseSHA256 {
			return fmt.Errorf("B-FULL route evidence response authentication failed for %q", observation.Route)
		}
		if _, duplicate := routes[observation.Route]; duplicate {
			return fmt.Errorf("B-FULL duplicate route evidence %s", observation.Route)
		}
		routes[observation.Route] = struct{}{}
	}
	for _, route := range required.Routes {
		if _, ok := routes[route]; !ok {
			return fmt.Errorf("B-FULL missing route evidence %s", route)
		}
	}
	if err := validateBFullATExemplars(coverage.ATExemplars); err != nil {
		return err
	}
	return validateBFullConsumerFixture(coverage.ConsumerFixture)
}

func validateExactBFullCoverageAxis(axis string, got, want []string) error {
	actual := append([]string(nil), got...)
	expected := append([]string(nil), want...)
	sort.Strings(actual)
	sort.Strings(expected)
	if len(actual) != len(expected) {
		return fmt.Errorf("B-FULL %s coverage = %d, want %d", axis, len(actual), len(expected))
	}
	for index := range expected {
		if actual[index] != expected[index] {
			return fmt.Errorf("B-FULL %s coverage mismatch at %d: got %q, want %q", axis, index, actual[index], expected[index])
		}
		if index > 0 && actual[index] == actual[index-1] {
			return fmt.Errorf("B-FULL duplicate %s coverage %s", axis, actual[index])
		}
	}
	return nil
}

func validateBFullATExemplars(observations []BFullATExemplarObservation) error {
	want := map[string]ATExemplar{}
	for _, at := range []string{"safari-voiceover", "chromium-screen-reader"} {
		for _, exemplar := range RequiredATExemplars {
			want[at+"/"+exemplar.Name] = exemplar
		}
	}
	seen := map[string]struct{}{}
	for _, observation := range observations {
		key := observation.AT + "/" + observation.Name
		exemplar, ok := want[key]
		if !ok {
			return fmt.Errorf("B-FULL unexpected AT exemplar %s", key)
		}
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("B-FULL duplicate AT exemplar %s", key)
		}
		if observation.Route != exemplar.Route || observation.State != exemplar.State || observation.Browser == "" {
			return fmt.Errorf("B-FULL AT exemplar %s route/state/browser context does not match maintained authority", key)
		}
		seen[key] = struct{}{}
	}
	if len(seen) != len(want) {
		return fmt.Errorf("B-FULL AT exemplars = %d, want %d", len(seen), len(want))
	}
	return nil
}

func validateBFullConsumerFixture(fixture BFullConsumerFixture) error {
	if fixture.ID == "" || fixture.BaseTheme != "modern" || fixture.StylesheetPath == "" || !validSHA256(fixture.StylesheetSHA256) {
		return fmt.Errorf("B-FULL explicit Modern consumer fixture identity is required")
	}
	got, err := SHA256File(fixture.StylesheetPath)
	if err != nil {
		return fmt.Errorf("authenticate B-FULL consumer fixture: %w", err)
	}
	if got != fixture.StylesheetSHA256 {
		return fmt.Errorf("B-FULL consumer fixture SHA-256 = %s, want %s", got, fixture.StylesheetSHA256)
	}
	return nil
}

func ValidateBFullAxesAgainstLedger(axes BFullAxes, ledger Ledger) error {
	sourceThemes := make(map[string]struct{})
	for _, row := range ledger.Rows {
		if row.Class == ClassInventory && row.Theme != "" {
			sourceThemes[row.Theme] = struct{}{}
		}
	}
	want := []string{"araihu", "goshtoso", "minimal", "modern"}
	if strings.Join(axes.Themes, "\x00") != strings.Join(want, "\x00") {
		return fmt.Errorf("B-FULL deep theme axis = %v, want source-derived %v", axes.Themes, want)
	}
	for _, theme := range axes.Themes {
		if _, ok := sourceThemes[theme]; !ok {
			return fmt.Errorf("B-FULL theme %q is absent from source inventory", theme)
		}
	}
	return nil
}

func ValidateBFullManifest(manifest BFullManifest, sourceCommit, sourceTree string, axes BFullAxes) error {
	if manifest.SchemaVersion != SchemaVersion {
		return fmt.Errorf("B-FULL schema version %d, want %d", manifest.SchemaVersion, SchemaVersion)
	}
	if manifest.SourceCommit != sourceCommit || manifest.SourceTree != sourceTree {
		return fmt.Errorf("B-FULL source identity %s/%s, want %s/%s", manifest.SourceCommit, manifest.SourceTree, sourceCommit, sourceTree)
	}
	if strings.TrimSpace(manifest.Browser) == "" {
		return fmt.Errorf("B-FULL browser identity is required")
	}
	if manifest.Diagnostic != nil {
		return fmt.Errorf("B-FULL diagnostic/non-closure evidence cannot close execution rows")
	}
	if err := VerifyBFullIdentity(manifest.Identity, sourceCommit, sourceTree); err != nil {
		return err
	}

	expectedBatches := make(map[string]struct{})
	for _, theme := range axes.Themes {
		for _, mode := range axes.Modes {
			for _, viewport := range axes.Viewports {
				for _, zoom := range axes.Zooms {
					for _, motion := range axes.Motions {
						expectedBatches[bfullBatchKey(theme, mode, viewport, zoom, motion)] = struct{}{}
					}
				}
			}
		}
	}
	batchEvidence := make(map[string]string, len(manifest.Batches))
	batchInputs := make(map[string]map[string]BFullInputObservation, len(manifest.Batches))
	for index, batch := range manifest.Batches {
		key := bfullBatchKey(batch.Theme, batch.Mode, batch.Viewport, batch.Zoom, batch.Motion)
		if _, ok := expectedBatches[key]; !ok {
			return fmt.Errorf("B-FULL extra batch %s at index %d", key, index)
		}
		if _, duplicate := batchEvidence[key]; duplicate {
			return fmt.Errorf("B-FULL duplicate batch %s", key)
		}
		inputs, err := validateBFullBatchEvidence(batch, manifest, axes)
		if err != nil {
			return fmt.Errorf("B-FULL batch %s: %w", key, err)
		}
		batchEvidence[key] = batch.EvidenceSHA256
		batchInputs[key] = inputs
	}
	if len(batchEvidence) != len(expectedBatches) {
		return fmt.Errorf("B-FULL missing %d mandatory batches", len(expectedBatches)-len(batchEvidence))
	}

	expected := make(map[string]struct{}, bfullCellCount(axes))
	for _, state := range axes.States {
		for _, theme := range axes.Themes {
			for _, mode := range axes.Modes {
				for _, viewport := range axes.Viewports {
					for _, zoom := range axes.Zooms {
						for _, motion := range axes.Motions {
							for _, input := range axes.Inputs {
								expected[bfullKey(state, theme, mode, viewport, zoom, motion, input)] = struct{}{}
							}
						}
					}
				}
			}
		}
	}

	seen := make(map[string]struct{}, len(manifest.Cells))
	for index, cell := range manifest.Cells {
		key := bfullKey(cell.State, cell.Theme, cell.Mode, cell.Viewport, cell.Zoom, cell.Motion, cell.Input)
		batchKey := bfullBatchKey(cell.Theme, cell.Mode, cell.Viewport, cell.Zoom, cell.Motion)
		if _, ok := expected[key]; !ok {
			return fmt.Errorf("B-FULL extra cell %s at index %d", key, index)
		}
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("B-FULL duplicate cell %s", key)
		}
		seen[key] = struct{}{}
		switch cell.Applicability {
		case Applicable:
			if cell.ReceiptStatus != StatusExecuted {
				return fmt.Errorf("B-FULL applicable cell %s must be executed", key)
			}
		case NotApplicable:
			if cell.ReceiptStatus != StatusNotApplicable || strings.TrimSpace(cell.Rationale) == "" {
				return fmt.Errorf("B-FULL N/A cell %s requires not-applicable status and rationale", key)
			}
		default:
			return fmt.Errorf("B-FULL cell %s has invalid applicability %q", key, cell.Applicability)
		}
		if cell.EvidenceSHA256 != batchEvidence[batchKey] {
			return fmt.Errorf("B-FULL cell %s evidence does not authenticate batch %s", key, batchKey)
		}
		observation, ok := batchInputs[batchKey][bfullInputKey(cell.State, cell.Input)]
		if !ok {
			return fmt.Errorf("B-FULL cell %s lacks exact input observation", key)
		}
		if cell.Applicability != observation.Applicability || cell.ReceiptStatus != observation.ReceiptStatus || cell.Rationale != observation.Rationale {
			return fmt.Errorf("B-FULL cell %s status does not match authenticated input observation", key)
		}
	}
	if len(seen) != len(expected) {
		var missing []string
		for key := range expected {
			if _, ok := seen[key]; !ok {
				missing = append(missing, key)
			}
		}
		sort.Strings(missing)
		return fmt.Errorf("B-FULL missing %d mandatory cells; first=%s", len(missing), missing[0])
	}
	return nil
}

func validateBFullBatchEvidence(batch BFullBatch, manifest BFullManifest, axes BFullAxes) (map[string]BFullInputObservation, error) {
	if strings.TrimSpace(batch.EvidencePath) == "" || !validSHA256(batch.EvidenceSHA256) {
		return nil, fmt.Errorf("evidence path and SHA-256 are required")
	}
	gotDigest, err := SHA256File(batch.EvidencePath)
	if err != nil {
		return nil, fmt.Errorf("authenticate evidence: %w", err)
	}
	if gotDigest != batch.EvidenceSHA256 {
		return nil, fmt.Errorf("evidence SHA-256 = %s, want %s", gotDigest, batch.EvidenceSHA256)
	}
	content, err := os.ReadFile(batch.EvidencePath)
	if err != nil {
		return nil, err
	}
	var evidence BFullBatchEvidence
	if err := json.Unmarshal(content, &evidence); err != nil {
		return nil, fmt.Errorf("parse structured evidence: %w", err)
	}
	if evidence.SchemaVersion != SchemaVersion || evidence.SourceCommit != manifest.SourceCommit || evidence.SourceTree != manifest.SourceTree || evidence.Browser != manifest.Browser || evidence.Identity != manifest.Identity {
		return nil, fmt.Errorf("structured evidence identity mismatch")
	}
	if evidence.Theme != batch.Theme || evidence.Mode != batch.Mode || evidence.Viewport != batch.Viewport || evidence.Zoom != batch.Zoom || evidence.Motion != batch.Motion {
		return nil, fmt.Errorf("structured evidence axes mismatch")
	}
	if strings.TrimSpace(evidence.InitialHTMLPath) == "" || !validSHA256(evidence.InitialHTMLSHA256) {
		return nil, fmt.Errorf("initial HTML artifact path and SHA-256 are required")
	}
	initialDigest, err := SHA256File(evidence.InitialHTMLPath)
	if err != nil {
		return nil, fmt.Errorf("authenticate initial HTML: %w", err)
	}
	if initialDigest != evidence.InitialHTMLSHA256 {
		return nil, fmt.Errorf("initial HTML SHA-256 = %s, want %s", initialDigest, evidence.InitialHTMLSHA256)
	}
	if err := validateBFullStateObservations("first paint", evidence.FirstPaint, axes.States, evidence); err != nil {
		return nil, err
	}
	if err := validateBFullStateObservations("persistence", evidence.Persistence, axes.States, evidence); err != nil {
		return nil, err
	}
	if err := validateBFullAccessibility(evidence.Accessibility, len(axes.States)); err != nil {
		return nil, err
	}
	if len(evidence.Accessibility.Violations) != 0 {
		return nil, fmt.Errorf("automated accessibility scan has %d violations", len(evidence.Accessibility.Violations))
	}
	if len(evidence.Errors) != 0 {
		return nil, fmt.Errorf("browser/page/network/CSP errors = %d", len(evidence.Errors))
	}

	wantInputs := make(map[string]struct{}, len(axes.States)*len(axes.Inputs))
	for _, state := range axes.States {
		for _, input := range axes.Inputs {
			wantInputs[bfullInputKey(state, input)] = struct{}{}
		}
	}
	inputs := make(map[string]BFullInputObservation, len(evidence.Inputs))
	for _, observation := range evidence.Inputs {
		key := bfullInputKey(observation.State, observation.Input)
		if _, ok := wantInputs[key]; !ok {
			return nil, fmt.Errorf("extra input observation %s", key)
		}
		if _, duplicate := inputs[key]; duplicate {
			return nil, fmt.Errorf("duplicate input observation %s", key)
		}
		if err := validateBFullInputObservation(observation); err != nil {
			return nil, fmt.Errorf("input observation %s: %w", key, err)
		}
		inputs[key] = observation
	}
	if len(inputs) != len(wantInputs) {
		return nil, fmt.Errorf("input observations = %d, want %d", len(inputs), len(wantInputs))
	}
	return inputs, nil
}

func validateBFullStateObservations(label string, observations []BFullStateObservation, states []string, evidence BFullBatchEvidence) error {
	want := make(map[string]struct{}, len(states))
	for _, state := range states {
		want[state] = struct{}{}
	}
	seen := make(map[string]struct{}, len(observations))
	for _, observation := range observations {
		if _, ok := want[observation.State]; !ok {
			return fmt.Errorf("%s has extra state %s", label, observation.State)
		}
		if _, duplicate := seen[observation.State]; duplicate {
			return fmt.Errorf("%s has duplicate state %s", label, observation.State)
		}
		seen[observation.State] = struct{}{}
		if !observation.Exists || observation.DOMNodes < 1 {
			return fmt.Errorf("%s state %s is absent", label, observation.State)
		}
		if !observation.Visible {
			return fmt.Errorf("%s state %s is invisible", label, observation.State)
		}
		if observation.Width < 1 || observation.Height < 1 {
			return fmt.Errorf("%s state %s has no rendered dimensions", label, observation.State)
		}
		if observation.OverflowX {
			return fmt.Errorf("%s state %s overflows horizontally", label, observation.State)
		}
		if observation.Theme != evidence.Theme || observation.Mode != evidence.Mode || observation.Motion != evidence.Motion || observation.Zoom != evidence.Zoom {
			return fmt.Errorf("%s state %s axes mismatch", label, observation.State)
		}
		if strings.TrimSpace(observation.Color) == "" || strings.TrimSpace(observation.Background) == "" {
			return fmt.Errorf("%s state %s lacks computed paint", label, observation.State)
		}
		for name, assertion := range map[string]BFullSemanticAssertion{
			"text contrast":      observation.TextContrast,
			"boundary contrast":  observation.BoundaryContrast,
			"motion outcome":     observation.MotionOutcome,
			"overlay provenance": observation.OverlayProvenance,
		} {
			if assertion.Applicability == Applicable && !assertion.Passed {
				return fmt.Errorf("%s state %s %s did not pass", label, observation.State, name)
			}
			if assertion.Applicability == NotApplicable && strings.TrimSpace(assertion.Rationale) == "" {
				return fmt.Errorf("%s state %s %s N/A lacks source-grounded rationale", label, observation.State, name)
			}
			if assertion.Applicability != Applicable && assertion.Applicability != NotApplicable {
				return fmt.Errorf("%s state %s %s lacks applicability", label, observation.State, name)
			}
		}
	}
	if len(seen) != len(want) {
		return fmt.Errorf("%s states = %d, want %d", label, len(seen), len(want))
	}
	return nil
}

func validateBFullInputObservation(observation BFullInputObservation) error {
	assertNotApplicable := func(name string, assertion BFullSemanticAssertion) error {
		if assertion.Applicability != NotApplicable || strings.TrimSpace(assertion.Rationale) == "" {
			return fmt.Errorf("%s N/A rationale is required", name)
		}
		return nil
	}
	assertPassed := func(name string, assertion BFullSemanticAssertion) error {
		if assertion.Applicability != Applicable || !assertion.Passed {
			return fmt.Errorf("%s applicable assertion must pass", name)
		}
		return nil
	}
	if observation.Applicability == NotApplicable {
		if observation.ReceiptStatus != StatusNotApplicable || strings.TrimSpace(observation.Rationale) == "" {
			return fmt.Errorf("N/A input requires status and rationale")
		}
		if observation.TargetSelector != "" || observation.EventCount != 0 || strings.TrimSpace(observation.SourceGrounding) == "" {
			return fmt.Errorf("N/A input cannot claim target or events")
		}
		for name, assertion := range map[string]BFullSemanticAssertion{"focus-visible": observation.FocusVisible, "movement-return": observation.MovementReturn} {
			if err := assertNotApplicable(name, assertion); err != nil {
				return err
			}
		}
		return assertEscapeNotApplicable(observation.Escape)
	}
	if observation.Applicability != Applicable || observation.ReceiptStatus != StatusExecuted {
		return fmt.Errorf("applicable input must be executed")
	}
	if strings.TrimSpace(observation.TargetSelector) == "" || strings.TrimSpace(observation.TargetRole) == "" || strings.TrimSpace(observation.AccessibleName) == "" || observation.ARIAState == nil || observation.EventCount < 1 || strings.TrimSpace(observation.Driver) == "" || strings.TrimSpace(observation.SourceGrounding) == "" {
		return fmt.Errorf("executed input requires target, role, name, ARIA state, and event")
	}
	if !observation.Before.TargetConnected || !observation.Before.TargetVisible || !observation.Action.TargetConnected || !observation.Action.TargetVisible || !observation.Return.TargetConnected || !observation.Return.TargetVisible {
		return fmt.Errorf("executed input requires connected visible before/action/return outcomes")
	}
	if len(observation.Action.EventTypes) == 0 {
		return fmt.Errorf("executed input requires browser-observed action events")
	}
	switch observation.Input {
	case "mouse":
		if err := assertPassed("movement-return", observation.MovementReturn); err != nil {
			return err
		}
		if err := assertNotApplicable("focus-visible", observation.FocusVisible); err != nil {
			return err
		}
		return assertEscapeNotApplicable(observation.Escape)
	case "keyboard":
		if err := assertPassed("focus-visible", observation.FocusVisible); err != nil {
			return err
		}
		if observation.Escape.Applicability == Applicable {
			if !observation.Escape.Passed || !observation.Escape.Opened || !observation.Escape.Closed || !observation.Escape.FocusReturned || strings.TrimSpace(observation.Escape.SurfaceSelector) == "" {
				return fmt.Errorf("Escape applicable assertion must open, close, and return focus")
			}
		} else if err := assertEscapeNotApplicable(observation.Escape); err != nil {
			return err
		}
		return assertNotApplicable("movement-return", observation.MovementReturn)
	case "touch":
		for name, assertion := range map[string]BFullSemanticAssertion{"focus-visible": observation.FocusVisible, "movement-return": observation.MovementReturn} {
			if err := assertNotApplicable(name, assertion); err != nil {
				return err
			}
		}
		return assertEscapeNotApplicable(observation.Escape)
	default:
		return fmt.Errorf("unsupported input %q", observation.Input)
	}
}

func assertEscapeNotApplicable(outcome BFullEscapeOutcome) error {
	if outcome.Applicability != NotApplicable || strings.TrimSpace(outcome.Rationale) == "" {
		return fmt.Errorf("Escape N/A rationale is required")
	}
	if outcome.Opened || outcome.Closed || outcome.FocusReturned || outcome.SurfaceSelector != "" {
		return fmt.Errorf("Escape N/A cannot claim a dismissible surface")
	}
	return nil
}

func validateBFullAccessibility(scan BFullAccessibilityScan, stateCount int) error {
	if scan.Engine != "axe-core" || scan.Version != AxeCoreVersion || scan.ArchivePath == "" || !validSHA256(scan.ArchiveSHA256) || !validSHA256(scan.ScriptSHA256) || len(scan.Rules) == 0 || scan.ScannedStates != stateCount {
		return fmt.Errorf("maintained axe-core accessibility scan is incomplete")
	}
	if got, err := SHA256File(scan.ArchivePath); err != nil {
		return fmt.Errorf("authenticate axe-core archive: %w", err)
	} else if got != scan.ArchiveSHA256 {
		return fmt.Errorf("axe-core archive SHA-256 = %s, want %s", got, scan.ArchiveSHA256)
	}
	if len(scan.Violations) != 0 {
		return fmt.Errorf("automated accessibility scan has %d violations", len(scan.Violations))
	}
	if len(scan.Incomplete) != 0 {
		return fmt.Errorf("automated accessibility scan has %d incomplete results", len(scan.Incomplete))
	}
	seen := map[string]struct{}{}
	for _, result := range scan.ChecklistResults {
		if result.RuleID == "" || result.Criterion == "" || result.URL != ChecklistA11Y || result.Outcome == "" {
			return fmt.Errorf("axe checklist criterion result is incomplete")
		}
		seen[result.RuleID] = struct{}{}
	}
	for _, rule := range scan.Rules {
		if _, ok := seen[rule]; !ok {
			return fmt.Errorf("axe rule %s lacks Checklist Design criterion result", rule)
		}
	}
	return nil
}

func SHA256File(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:]), nil
}

// BuildBFullIdentity makes the worktree boundary explicit. A clean committed
// checkout is the only identity suitable for closure. A dirty-bound identity
// is permitted solely for diagnostic evidence and carries a complete manifest
// of every changed path and byte hash.
func BuildBFullIdentity(repoRoot, sourceCommit, sourceTree, mode, manifestPath string) (BFullIdentity, error) {
	if strings.TrimSpace(repoRoot) == "" {
		return BFullIdentity{}, fmt.Errorf("B-FULL repository root is required")
	}
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return BFullIdentity{}, fmt.Errorf("resolve B-FULL repository root: %w", err)
	}
	if err := verifyBFullGitHead(repoRoot, sourceCommit, sourceTree); err != nil {
		return BFullIdentity{}, err
	}
	entries, err := deriveBFullDirtyEntries(repoRoot)
	if err != nil {
		return BFullIdentity{}, err
	}
	switch mode {
	case BFullIdentityCommitted:
		if len(entries) != 0 {
			return BFullIdentity{}, fmt.Errorf("B-FULL committed identity requires clean worktree; changed paths=%d", len(entries))
		}
		return BFullIdentity{Mode: BFullIdentityCommitted, RepoRoot: repoRoot}, nil
	case BFullIdentityDirtyBound:
		if strings.TrimSpace(manifestPath) == "" {
			return BFullIdentity{}, fmt.Errorf("B-FULL dirty-bound identity requires a manifest path")
		}
		manifest := BFullDirtyManifest{SchemaVersion: SchemaVersion, SourceCommit: sourceCommit, SourceTree: sourceTree, Entries: entries}
		encoded, err := json.Marshal(manifest)
		if err != nil {
			return BFullIdentity{}, fmt.Errorf("encode B-FULL dirty manifest: %w", err)
		}
		if err := os.WriteFile(manifestPath, append(encoded, '\n'), 0o600); err != nil {
			return BFullIdentity{}, fmt.Errorf("write B-FULL dirty manifest: %w", err)
		}
		digest, err := SHA256File(manifestPath)
		if err != nil {
			return BFullIdentity{}, err
		}
		return BFullIdentity{Mode: BFullIdentityDirtyBound, RepoRoot: repoRoot, DirtyManifestPath: manifestPath, DirtyManifestSHA256: digest}, nil
	default:
		return BFullIdentity{}, fmt.Errorf("B-FULL identity mode %q is invalid", mode)
	}
}

// VerifyBFullIdentity authenticates the recorded mode and, for a diagnostic,
// confirms that the source worktree still has exactly the bound dirty bytes.
func VerifyBFullIdentity(identity BFullIdentity, sourceCommit, sourceTree string) error {
	switch identity.Mode {
	case BFullIdentityCommitted:
		if identity.DirtyManifestPath != "" || identity.DirtyManifestSHA256 != "" {
			return fmt.Errorf("B-FULL committed identity cannot carry a dirty manifest")
		}
		if identity.RepoRoot == "" {
			return nil // unit fixtures authenticate shape; production identities carry a root.
		}
		if err := verifyBFullGitHead(identity.RepoRoot, sourceCommit, sourceTree); err != nil {
			return err
		}
		entries, err := deriveBFullDirtyEntries(identity.RepoRoot)
		if err != nil {
			return err
		}
		if len(entries) != 0 {
			return fmt.Errorf("B-FULL committed identity worktree is dirty; changed paths=%d", len(entries))
		}
		return nil
	case BFullIdentityDirtyBound:
		if identity.RepoRoot == "" || identity.DirtyManifestPath == "" || !validSHA256(identity.DirtyManifestSHA256) {
			return fmt.Errorf("B-FULL dirty-bound identity requires repo root and authenticated manifest")
		}
		if err := verifyBFullGitHead(identity.RepoRoot, sourceCommit, sourceTree); err != nil {
			return err
		}
		got, err := SHA256File(identity.DirtyManifestPath)
		if err != nil {
			return fmt.Errorf("authenticate B-FULL dirty manifest: %w", err)
		}
		if got != identity.DirtyManifestSHA256 {
			return fmt.Errorf("B-FULL dirty manifest SHA-256 = %s, want %s", got, identity.DirtyManifestSHA256)
		}
		content, err := os.ReadFile(identity.DirtyManifestPath)
		if err != nil {
			return err
		}
		var manifest BFullDirtyManifest
		if err := json.Unmarshal(content, &manifest); err != nil {
			return fmt.Errorf("parse B-FULL dirty manifest: %w", err)
		}
		if manifest.SchemaVersion != SchemaVersion || manifest.SourceCommit != sourceCommit || manifest.SourceTree != sourceTree {
			return fmt.Errorf("B-FULL dirty manifest source identity mismatch")
		}
		current, err := deriveBFullDirtyEntries(identity.RepoRoot)
		if err != nil {
			return err
		}
		if !sameBFullDirtyEntries(manifest.Entries, current) {
			return fmt.Errorf("B-FULL dirty-bound worktree changed after manifest capture")
		}
		return nil
	default:
		return fmt.Errorf("B-FULL identity mode %q is invalid", identity.Mode)
	}
}

func verifyBFullGitHead(repoRoot, sourceCommit, sourceTree string) error {
	for _, check := range []struct {
		name string
		rev  string
		want string
	}{{"commit", "HEAD^{commit}", sourceCommit}, {"tree", "HEAD^{tree}", sourceTree}} {
		command := exec.Command("git", "-C", repoRoot, "rev-parse", check.rev)
		output, err := command.Output()
		if err != nil {
			return fmt.Errorf("resolve B-FULL source %s: %w", check.name, err)
		}
		if got := strings.TrimSpace(string(output)); got != check.want {
			return fmt.Errorf("B-FULL source %s = %s, want %s", check.name, got, check.want)
		}
	}
	return nil
}

func deriveBFullDirtyEntries(repoRoot string) ([]BFullDirtyEntry, error) {
	command := exec.Command("git", "-C", repoRoot, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("read B-FULL worktree status: %w", err)
	}
	parts := strings.Split(string(output), "\x00")
	entries := make([]BFullDirtyEntry, 0, len(parts))
	for index := 0; index < len(parts); index++ {
		item := parts[index]
		if item == "" {
			continue
		}
		if len(item) < 4 {
			return nil, fmt.Errorf("parse B-FULL git status entry %q", item)
		}
		status, relative := item[:2], item[3:]
		entry := BFullDirtyEntry{Path: filepath.ToSlash(relative), Status: status}
		path := filepath.Join(repoRoot, relative)
		info, statErr := os.Stat(path)
		if statErr == nil && !info.IsDir() {
			digest, digestErr := SHA256File(path)
			if digestErr != nil {
				return nil, digestErr
			}
			entry.SHA256 = digest
		} else if statErr != nil && !os.IsNotExist(statErr) {
			return nil, statErr
		}
		entries = append(entries, entry)
		// In porcelain v1 -z, a rename/copy record carries a second NUL-delimited
		// source path. It has no current worktree bytes to hash, but it is still
		// identity-significant and must not be mistaken for another status record.
		if strings.Contains(status, "R") || strings.Contains(status, "C") {
			if index+1 >= len(parts) || parts[index+1] == "" {
				return nil, fmt.Errorf("parse B-FULL rename/copy status entry %q", item)
			}
			index++
			entries = append(entries, BFullDirtyEntry{Path: filepath.ToSlash(parts[index]), Status: "source:" + status})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Path == entries[j].Path {
			return entries[i].Status < entries[j].Status
		}
		return entries[i].Path < entries[j].Path
	})
	return entries, nil
}

func sameBFullDirtyEntries(left, right []BFullDirtyEntry) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func bfullCellCount(axes BFullAxes) int {
	return len(axes.States) * len(axes.Themes) * len(axes.Modes) * len(axes.Viewports) * len(axes.Zooms) * len(axes.Motions) * len(axes.Inputs)
}

func bfullKey(state, theme, mode string, viewport, zoom int, motion, input string) string {
	return fmt.Sprintf("%s|%s|%s|%d|%d|%s|%s", state, theme, mode, viewport, zoom, motion, input)
}

func bfullBatchKey(theme, mode string, viewport, zoom int, motion string) string {
	return fmt.Sprintf("%s|%s|%d|%d|%s", theme, mode, viewport, zoom, motion)
}

func bfullInputKey(state, input string) string {
	return state + "|" + input
}

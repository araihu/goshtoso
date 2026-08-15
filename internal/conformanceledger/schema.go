// Package conformanceledger derives and validates Goshtoso conformance evidence.
// It is evidence tooling only and is not used by component runtime code.
package conformanceledger

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
)

const SchemaVersion = 1

type Applicability string

const (
	Applicable    Applicability = "applicable"
	NotApplicable Applicability = "not-applicable"
)

type ReceiptStatus string

const (
	StatusExecuted      ReceiptStatus = "executed"
	StatusFailed        ReceiptStatus = "failed"
	StatusBlocked       ReceiptStatus = "blocked"
	StatusNotApplicable ReceiptStatus = "not-applicable"
)

type EvidenceClass string

const (
	ClassInventory EvidenceClass = "inventory"
	ClassExecution EvidenceClass = "execution"
)

type SourceRef struct {
	Path   string `json:"path"`
	Symbol string `json:"symbol"`
}

type SourceItem struct {
	Value  string    `json:"value"`
	Source SourceRef `json:"source"`
	// Action is the maintained real interaction contract when this authority
	// represents an executable configuration or lifecycle state.
	Action string `json:"action,omitempty"`
	// Outcome is the source-grounded postcondition for an executable action.
	// Browser evidence must later bind a real before/action/after artifact to it.
	Outcome string `json:"outcome,omitempty"`
}

type Inventory struct {
	Packages        []SourceItem `json:"packages"`
	Renderables     []SourceItem `json:"renderables"`
	Kinds           []SourceItem `json:"kinds"`
	Routes          []SourceItem `json:"routes"`
	States          []SourceItem `json:"states"`
	LifecycleStates []SourceItem `json:"lifecycle_states"`
	Themes          []SourceItem `json:"themes"`
	BreakpointEdges []int        `json:"breakpoint_edges"`
}

// RequiredATExemplars is the maintained source authority for the five real
// AT journeys required for each supported AT pair. It is intentionally shared
// by ledger generation and B-FULL validation so a label cannot drift onto an
// arbitrary route or state.
type ATExemplar struct {
	Name   string
	Route  string
	State  string
	Source SourceRef
}

var RequiredATExemplars = []ATExemplar{
	{Name: "public", Route: "/getting-started", State: "initial", Source: SourceRef{Path: "site/internal/pages/demo/contentpages/start/getting_started_templ.go", Symbol: "Getting Started"}},
	{Name: "authenticated", Route: "/examples/profile", State: "profile-initial", Source: SourceRef{Path: "site/internal/pages/demo/examplepages/profile/profile.templ", Symbol: "Profile"}},
	{Name: "data", Route: "/components/table", State: "table-initial-sortable", Source: SourceRef{Path: "site/internal/pages/demo/componentpages/table/table.templ", Symbol: "Table"}},
	{Name: "form", Route: "/components/form", State: "form-initial", Source: SourceRef{Path: "site/internal/pages/demo/componentpages/form/form.templ", Symbol: "Form"}},
	{Name: "nav", Route: "/components/sidebar", State: "sidebar-navigation", Source: SourceRef{Path: "site/internal/pages/demo/componentpages/sidebar/sidebar.templ", Symbol: "sidebar navigation"}},
}

type Row struct {
	ID                string                  `json:"id"`
	Class             EvidenceClass           `json:"class"`
	Package           string                  `json:"package,omitempty"`
	Renderable        string                  `json:"renderable,omitempty"`
	Kind              string                  `json:"kind,omitempty"`
	Route             string                  `json:"route,omitempty"`
	State             string                  `json:"state,omitempty"`
	Theme             string                  `json:"theme,omitempty"`
	Mode              string                  `json:"mode,omitempty"`
	Viewport          int                     `json:"viewport,omitempty"`
	Breakpoint        string                  `json:"breakpoint,omitempty"`
	Zoom              int                     `json:"zoom,omitempty"`
	Motion            string                  `json:"motion,omitempty"`
	Input             string                  `json:"input,omitempty"`
	AT                string                  `json:"at,omitempty"`
	ChecklistURLs     []string                `json:"checklist_urls,omitempty"`
	ChecklistMappings []ChecklistMapping      `json:"checklist_mappings,omitempty"`
	Sources           []SourceRef             `json:"sources"`
	Applicability     Applicability           `json:"applicability"`
	ReceiptStatus     ReceiptStatus           `json:"receipt_status"`
	Receipt           string                  `json:"receipt,omitempty"`
	Rationale         string                  `json:"rationale,omitempty"`
	EvidenceHashes    map[string]string       `json:"evidence_hashes,omitempty"`
	Reproductions     []ExecutionReproduction `json:"reproductions,omitempty"`
	ExecutionAttempts []ExecutionAttempt      `json:"execution_attempts,omitempty"`
}

// ExecutionReproduction identifies one independently-produced execution
// evidence set. Closure-critical rows require two non-aliased records: a
// second receipt cannot merely relabel the same runner, recorder, or capture
// artifacts.
type ExecutionReproduction struct {
	SourceCommit    string   `json:"source_commit"`
	SourceTree      string   `json:"source_tree"`
	Producer        string   `json:"producer"`
	RunID           string   `json:"run_id"`
	Recorder        string   `json:"recorder"`
	ReceiptSHA256   string   `json:"receipt_sha256"`
	ArtifactSHA256s []string `json:"artifact_sha256s,omitempty"`
}

// ExecutionAttempt preserves every authenticated receipt application. Failed,
// blocked, and N/A attempts are history only; they cannot become closure
// reproductions or overwrite an earlier successful receipt.
type ExecutionAttempt struct {
	ReceiptID       string        `json:"receipt_id"`
	ReceiptPath     string        `json:"receipt_path"`
	Status          ReceiptStatus `json:"status"`
	Rationale       string        `json:"rationale,omitempty"`
	SourceCommit    string        `json:"source_commit"`
	SourceTree      string        `json:"source_tree"`
	Producer        string        `json:"producer"`
	RunID           string        `json:"run_id"`
	Recorder        string        `json:"recorder"`
	ReceiptSHA256   string        `json:"receipt_sha256"`
	ArtifactSHA256s []string      `json:"artifact_sha256s,omitempty"`
}

type Ledger struct {
	SchemaVersion int               `json:"schema_version"`
	SourceCommit  string            `json:"source_commit"`
	SourceTree    string            `json:"source_tree"`
	Rows          []Row             `json:"rows"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// Validate checks schema integrity and reports every absent mandatory axis.
// Blocked and failed receipts remain structurally valid evidence; ValidateClosure
// rejects them when an ACCEPT disposition is requested.
func Validate(ledger Ledger, required Inventory) error {
	var problems []string
	if ledger.SchemaVersion != SchemaVersion {
		problems = append(problems, fmt.Sprintf("schema version %d, want %d", ledger.SchemaVersion, SchemaVersion))
	}

	rowIDs := make(map[string]int, len(ledger.Rows))
	for index, row := range ledger.Rows {
		prefix := fmt.Sprintf("row %d", index)
		if row.ID == "" {
			problems = append(problems, prefix+" missing id")
		} else if previous, duplicate := rowIDs[row.ID]; duplicate {
			problems = append(problems, fmt.Sprintf("row %d duplicate id %q first seen at row %d", index, row.ID, previous))
		} else {
			rowIDs[row.ID] = index
		}
		if row.Class != ClassInventory && row.Class != ClassExecution {
			problems = append(problems, prefix+" missing evidence class")
		}
		if len(row.Sources) == 0 || strings.TrimSpace(row.Sources[0].Path) == "" {
			problems = append(problems, prefix+" missing source path")
		}
		if len(row.Sources) == 0 || strings.TrimSpace(row.Sources[0].Symbol) == "" {
			problems = append(problems, prefix+" missing source symbol")
		}
		if row.Applicability != Applicable && row.Applicability != NotApplicable {
			problems = append(problems, prefix+" missing applicability")
		}
		if !validReceiptStatus(row.ReceiptStatus) {
			problems = append(problems, prefix+" missing receipt status")
		}
		if row.Applicability == NotApplicable {
			if row.ReceiptStatus != StatusNotApplicable {
				problems = append(problems, prefix+" N/A applicability requires not-applicable receipt status")
			}
			if strings.TrimSpace(row.Rationale) == "" {
				problems = append(problems, prefix+" N/A row missing rationale")
			}
			if mandatoryExecutionRow(row) {
				problems = append(problems, prefix+" mandatory execution row cannot be N/A")
			}
		} else {
			if row.ReceiptStatus == StatusNotApplicable {
				problems = append(problems, prefix+" applicable row cannot use not-applicable receipt status")
			}
			if validReceiptStatus(row.ReceiptStatus) && row.ReceiptStatus != StatusNotApplicable && strings.TrimSpace(row.Receipt) == "" {
				problems = append(problems, prefix+" missing receipt")
			}
		}
	}

	problems = append(problems, sourceAxisBijection("package", required.Packages, ledger.Rows, func(row Row) string { return row.Package })...)
	problems = append(problems, sourceAxisBijection("renderable", required.Renderables, ledger.Rows, func(row Row) string { return row.Renderable })...)
	problems = append(problems, sourceAxisBijection("kind", required.Kinds, ledger.Rows, func(row Row) string { return row.Kind })...)
	problems = append(problems, sourceAxisBijection("route", required.Routes, ledger.Rows, func(row Row) string { return row.Route })...)
	problems = append(problems, sourceAxisBijection("theme", required.Themes, ledger.Rows, func(row Row) string { return row.Theme })...)
	problems = append(problems, sourceAxisBijection("state", required.States, ledger.Rows, func(row Row) string { return row.State })...)
	problems = append(problems, sourceAxisBijection("lifecycle-state", required.LifecycleStates, ledger.Rows, func(row Row) string { return row.State })...)
	problems = append(problems, validateSourceAxisNamespace(ledger.Rows)...)
	problems = append(problems, validateChecklistProvenance(ledger.Rows)...)

	if !anyRow(ledger.Rows, func(row Row) bool { return row.State != "" }) {
		problems = append(problems, "missing mandatory state row")
	}
	for _, mode := range []string{"light", "dark"} {
		if !anyRow(ledger.Rows, func(row Row) bool { return row.Mode == mode }) {
			problems = append(problems, "missing mandatory mode "+mode)
		}
	}
	for _, zoom := range []int{100, 200} {
		if !anyRow(ledger.Rows, func(row Row) bool { return row.Zoom == zoom }) {
			problems = append(problems, fmt.Sprintf("missing mandatory zoom %d", zoom))
		}
	}
	for _, motion := range []string{"normal", "reduced"} {
		if !anyRow(ledger.Rows, func(row Row) bool { return row.Motion == motion }) {
			problems = append(problems, "missing mandatory motion "+motion)
		}
	}
	for _, input := range []string{"mouse", "keyboard", "touch"} {
		if !anyRow(ledger.Rows, func(row Row) bool { return row.Input == input }) {
			problems = append(problems, "missing mandatory input "+input)
		}
	}
	for _, at := range []string{"safari-voiceover", "chromium-screen-reader"} {
		if !anyRow(ledger.Rows, func(row Row) bool { return row.AT == at }) {
			problems = append(problems, "missing mandatory AT "+at)
		}
	}

	for _, width := range RequiredViewportWidths(required.BreakpointEdges) {
		if !anyRow(ledger.Rows, func(row Row) bool { return row.Viewport == width }) {
			problems = append(problems, fmt.Sprintf("missing mandatory breakpoint viewport %d", width))
		}
	}

	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return errors.New(strings.Join(problems, "; "))
}

func validateSourceAxisNamespace(rows []Row) []string {
	known := map[string]bool{
		"package": true, "renderable": true, "kind": true, "route": true,
		"theme": true, "state": true, "lifecycle-state": true,
	}
	var problems []string
	for _, row := range rows {
		parts := strings.Split(row.ID, "/")
		if len(parts) < 2 || (parts[0] != string(ClassInventory) && parts[0] != string(ClassExecution)) {
			continue
		}
		if len(parts) != 3 || !known[parts[1]] {
			problems = append(problems, fmt.Sprintf("row %s has unexpected source axis namespace", row.ID))
		}
	}
	return problems
}

func ValidateClosure(ledger Ledger, required Inventory) error {
	if err := Validate(ledger, required); err != nil {
		return err
	}
	if uncovered := strings.TrimSpace(ledger.Metadata["current.uncovered_config_contracts"]); uncovered != "" {
		return fmt.Errorf("closure has open rows: uncovered source Config contracts=%s", uncovered)
	}
	if lifecycle := strings.TrimSpace(ledger.Metadata["current.unverified_lifecycle_action_contracts"]); lifecycle != "" {
		return fmt.Errorf("closure has open rows: lifecycle action/outcome artifacts unavailable=%s", lifecycle)
	}
	var open []string
	for _, row := range ledger.Rows {
		if row.ReceiptStatus == StatusBlocked || row.ReceiptStatus == StatusFailed {
			open = append(open, fmt.Sprintf("%s=%s", row.ID, row.ReceiptStatus))
		}
		if mandatoryExecutionRow(row) && row.ReceiptStatus == StatusExecuted {
			if err := validateIndependentReproductions(row, ledger.SourceCommit, ledger.SourceTree); err != nil {
				open = append(open, fmt.Sprintf("%s=%v", row.ID, err))
			}
		}
	}
	if len(open) > 0 {
		sort.Strings(open)
		return fmt.Errorf("closure has open rows: %s", strings.Join(open, ", "))
	}
	return nil
}

func validateIndependentReproductions(row Row, sourceCommit, sourceTree string) error {
	if len(row.Reproductions) < 2 {
		return fmt.Errorf("requires two independent reproductions")
	}
	successfulAttempts := make(map[string]ExecutionAttempt, len(row.ExecutionAttempts))
	for _, attempt := range row.ExecutionAttempts {
		if attempt.Status != StatusExecuted {
			continue
		}
		if attempt.SourceCommit != sourceCommit || attempt.SourceTree != sourceTree || strings.TrimSpace(attempt.ReceiptID) == "" || strings.TrimSpace(attempt.ReceiptPath) == "" || strings.TrimSpace(attempt.Producer) == "" || strings.TrimSpace(attempt.RunID) == "" || strings.TrimSpace(attempt.Recorder) == "" || !validSHA256(attempt.ReceiptSHA256) {
			return fmt.Errorf("has incomplete successful immutable attempt")
		}
		key := attempt.ReceiptSHA256 + "\x00" + attempt.Producer + "\x00" + attempt.RunID + "\x00" + attempt.Recorder
		if _, duplicate := successfulAttempts[key]; duplicate {
			return fmt.Errorf("has duplicate successful immutable attempt")
		}
		successfulAttempts[key] = attempt
	}
	if len(successfulAttempts) != len(row.Reproductions) {
		return fmt.Errorf("successful immutable attempts = %d, want reproduction count %d", len(successfulAttempts), len(row.Reproductions))
	}
	seenProducer := map[string]struct{}{}
	seenRun := map[string]struct{}{}
	seenRecorder := map[string]struct{}{}
	seenArtifact := map[string]struct{}{}
	for _, reproduction := range row.Reproductions {
		if reproduction.SourceCommit != sourceCommit || reproduction.SourceTree != sourceTree || strings.TrimSpace(reproduction.Producer) == "" || strings.TrimSpace(reproduction.RunID) == "" || strings.TrimSpace(reproduction.Recorder) == "" || !validSHA256(reproduction.ReceiptSHA256) {
			return fmt.Errorf("has incomplete reproduction identity")
		}
		if _, duplicate := seenProducer[reproduction.Producer]; duplicate {
			return fmt.Errorf("has duplicate reproduction producer")
		}
		seenProducer[reproduction.Producer] = struct{}{}
		if _, duplicate := seenRun[reproduction.RunID]; duplicate {
			return fmt.Errorf("has duplicate reproduction run id")
		}
		seenRun[reproduction.RunID] = struct{}{}
		if _, duplicate := seenRecorder[reproduction.Recorder]; duplicate {
			return fmt.Errorf("has duplicate reproduction recorder")
		}
		seenRecorder[reproduction.Recorder] = struct{}{}
		for _, digest := range reproduction.ArtifactSHA256s {
			if !validSHA256(digest) {
				return fmt.Errorf("has invalid reproduction artifact SHA-256")
			}
			if _, duplicate := seenArtifact[digest]; duplicate {
				return fmt.Errorf("has aliased reproduction artifact SHA-256")
			}
			seenArtifact[digest] = struct{}{}
		}
		if len(reproduction.ArtifactSHA256s) == 0 {
			return fmt.Errorf("has no reproduction artifacts")
		}
		key := reproduction.ReceiptSHA256 + "\x00" + reproduction.Producer + "\x00" + reproduction.RunID + "\x00" + reproduction.Recorder
		attempt, ok := successfulAttempts[key]
		if !ok || !slices.Equal(attempt.ArtifactSHA256s, reproduction.ArtifactSHA256s) {
			return fmt.Errorf("reproduction is not backed by an immutable successful attempt")
		}
	}
	return nil
}

func mandatoryExecutionRow(row Row) bool {
	if row.Class != ClassExecution {
		return false
	}
	switch {
	case row.AT != "":
		return strings.HasPrefix(row.ID, "at/"+row.AT+"/")
	case row.Package != "":
		return isSourceAxisRow(row, ClassExecution, "package", row.Package)
	case row.Renderable != "":
		return isSourceAxisRow(row, ClassExecution, "renderable", row.Renderable)
	case row.Kind != "":
		return isSourceAxisRow(row, ClassExecution, "kind", row.Kind)
	case row.Route != "":
		return isSourceAxisRow(row, ClassExecution, "route", row.Route)
	case row.Theme != "":
		return isSourceAxisRow(row, ClassExecution, "theme", row.Theme)
	case row.State != "":
		return isSourceAxisRow(row, ClassExecution, "state", row.State) || isSourceAxisRow(row, ClassExecution, "lifecycle-state", row.State)
	case row.Mode != "":
		return row.ID == "mode/"+row.Mode
	case row.Viewport != 0:
		return row.ID == fmt.Sprintf("viewport/%d", row.Viewport)
	case row.Zoom != 0:
		return row.ID == fmt.Sprintf("zoom/%d", row.Zoom)
	case row.Motion != "":
		return row.ID == "motion/"+row.Motion
	case row.Input != "":
		return row.ID == "input/"+row.Input
	default:
		return false
	}
}

func RequiredViewportWidths(edges []int) []int {
	widths := []int{390, 768, 1440}
	for index, edge := range edges {
		widths = append(widths, edge-1, edge, edge+1)
		if index > 0 {
			widths = append(widths, (edges[index-1]+edge)/2)
		}
	}
	sort.Ints(widths)
	return slices.Compact(widths)
}

func validReceiptStatus(status ReceiptStatus) bool {
	return status == StatusExecuted || status == StatusFailed || status == StatusBlocked || status == StatusNotApplicable
}

func sourceAxisBijection(axis string, items []SourceItem, rows []Row, value func(Row) string) []string {
	allowed := make(map[string]bool, len(items))
	for _, item := range items {
		allowed[item.Value] = true
	}
	var problems []string
	for _, expected := range []struct {
		name  string
		class EvidenceClass
	}{
		{name: "inventory", class: ClassInventory},
		{name: "execution", class: ClassExecution},
	} {
		prefix := expected.name + "/" + axis + "/"
		counts := make(map[string]int, len(items))
		for _, row := range rows {
			if !strings.HasPrefix(row.ID, prefix) {
				continue
			}
			if row.Class != expected.class {
				problems = append(problems, fmt.Sprintf("%s %s row %s has class %s, want %s", expected.name, axis, row.ID, row.Class, expected.class))
				continue
			}
			axisValue := value(row)
			if axisValue != "" && !allowed[axisValue] {
				problems = append(problems, fmt.Sprintf("unexpected %s %s %s", expected.name, axis, axisValue))
				continue
			}
			if row.ID != ledgerSourceRowID(expected.class, axis, axisValue) {
				problems = append(problems, fmt.Sprintf("%s %s row ID %s does not match source value %s", expected.name, axis, row.ID, axisValue))
				continue
			}
			problems = append(problems, sourceAxisProjectionProblems(expected.name, axis, row)...)
			counts[axisValue]++
		}
		for _, item := range items {
			if counts[item.Value] != 1 {
				problems = append(problems, fmt.Sprintf("%s %s %s count = %d, want 1", expected.name, axis, item.Value, counts[item.Value]))
			}
		}
	}
	return problems
}

// sourceAxisProjectionProblems prevents one source-axis row from satisfying
// unrelated axes or from carrying a checklist mapping that only a route row
// can own. Exact IDs alone are insufficient because a caller could retain a
// valid package ID while smuggling a state or route value into the same row.
func sourceAxisProjectionProblems(className, axis string, row Row) []string {
	allowed := map[string]bool{}
	switch axis {
	case "package":
		allowed["package"] = true
	case "renderable":
		allowed["renderable"] = true
	case "kind":
		allowed["kind"] = true
	case "route":
		allowed["route"] = true
	case "state", "lifecycle-state":
		allowed["state"] = true
	case "theme":
		allowed["theme"] = true
	}
	values := []struct {
		name  string
		value string
	}{
		{name: "package", value: row.Package},
		{name: "renderable", value: row.Renderable},
		{name: "kind", value: row.Kind},
		{name: "route", value: row.Route},
		{name: "state", value: row.State},
		{name: "theme", value: row.Theme},
		{name: "mode", value: row.Mode},
		{name: "breakpoint", value: row.Breakpoint},
		{name: "input", value: row.Input},
		{name: "AT", value: row.AT},
	}
	var problems []string
	for _, value := range values {
		if value.value != "" && !allowed[value.name] {
			problems = append(problems, fmt.Sprintf("%s %s row %s projects unexpected %s", className, axis, row.ID, value.name))
		}
	}
	if row.Viewport != 0 && !allowed["viewport"] {
		problems = append(problems, fmt.Sprintf("%s %s row %s projects unexpected viewport", className, axis, row.ID))
	}
	if row.Zoom != 0 && !allowed["zoom"] {
		problems = append(problems, fmt.Sprintf("%s %s row %s projects unexpected zoom", className, axis, row.ID))
	}
	if row.Motion != "" && !allowed["motion"] {
		problems = append(problems, fmt.Sprintf("%s %s row %s projects unexpected motion", className, axis, row.ID))
	}
	if axis != "route" && (len(row.ChecklistURLs) != 0 || len(row.ChecklistMappings) != 0) {
		problems = append(problems, fmt.Sprintf("%s %s row %s non-route source row carries checklist", className, axis, row.ID))
	}
	return problems
}

func ledgerSourceRowID(class EvidenceClass, axis, value string) string {
	prefix := string(class)
	if axis == "route" {
		value = strings.TrimPrefix(value, "/components/")
	}
	return prefix + "/" + axis + "/" + strings.ReplaceAll(value, "/", "_")
}

func isSourceAxisRow(row Row, class EvidenceClass, axis, value string) bool {
	return row.Class == class && row.ID == ledgerSourceRowID(class, axis, value)
}

func validateChecklistProvenance(rows []Row) []string {
	var problems []string
	for _, row := range rows {
		if row.Route == "" || !isSourceAxisRow(row, row.Class, "route", row.Route) {
			continue
		}
		expected, err := ChecklistMappingsForRoute(row.Route)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s checklist mapping: %v", row.ID, err))
			continue
		}
		if !equalChecklistMappings(row.ChecklistMappings, expected) {
			problems = append(problems, fmt.Sprintf("%s checklist mapping provenance does not match %s", row.ID, row.Route))
		}
		urls := make([]string, 0, len(expected))
		for _, mapping := range expected {
			urls = append(urls, mapping.URL)
		}
		if !slices.Equal(row.ChecklistURLs, urls) {
			problems = append(problems, fmt.Sprintf("%s checklist URLs do not match structured mappings", row.ID))
		}
	}
	return problems
}

func equalChecklistMappings(left, right []ChecklistMapping) bool {
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

func anyRow(rows []Row, predicate func(Row) bool) bool {
	return slices.ContainsFunc(rows, predicate)
}

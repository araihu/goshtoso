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
	ID             string            `json:"id"`
	Class          EvidenceClass     `json:"class"`
	Package        string            `json:"package,omitempty"`
	Renderable     string            `json:"renderable,omitempty"`
	Kind           string            `json:"kind,omitempty"`
	Route          string            `json:"route,omitempty"`
	State          string            `json:"state,omitempty"`
	Theme          string            `json:"theme,omitempty"`
	Mode           string            `json:"mode,omitempty"`
	Viewport       int               `json:"viewport,omitempty"`
	Breakpoint     string            `json:"breakpoint,omitempty"`
	Zoom           int               `json:"zoom,omitempty"`
	Motion         string            `json:"motion,omitempty"`
	Input          string            `json:"input,omitempty"`
	AT             string            `json:"at,omitempty"`
	ChecklistURLs  []string          `json:"checklist_urls,omitempty"`
	Sources        []SourceRef       `json:"sources"`
	Applicability  Applicability     `json:"applicability"`
	ReceiptStatus  ReceiptStatus     `json:"receipt_status"`
	Receipt        string            `json:"receipt,omitempty"`
	Rationale      string            `json:"rationale,omitempty"`
	EvidenceHashes map[string]string `json:"evidence_hashes,omitempty"`
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
		} else {
			if row.ReceiptStatus == StatusNotApplicable {
				problems = append(problems, prefix+" applicable row cannot use not-applicable receipt status")
			}
			if validReceiptStatus(row.ReceiptStatus) && row.ReceiptStatus != StatusNotApplicable && strings.TrimSpace(row.Receipt) == "" {
				problems = append(problems, prefix+" missing receipt")
			}
		}
	}

	problems = append(problems, missingSourceItems("package", required.Packages, ledger.Rows, func(row Row) string { return row.Package })...)
	problems = append(problems, missingSourceItems("renderable", required.Renderables, ledger.Rows, func(row Row) string { return row.Renderable })...)
	problems = append(problems, missingSourceItems("kind", required.Kinds, ledger.Rows, func(row Row) string { return row.Kind })...)
	problems = append(problems, missingSourceItems("route", required.Routes, ledger.Rows, func(row Row) string { return row.Route })...)
	problems = append(problems, missingSourceItems("theme", required.Themes, ledger.Rows, func(row Row) string { return row.Theme })...)
	problems = append(problems, missingSourceItems("state", required.States, ledger.Rows, func(row Row) string { return row.State })...)
	problems = append(problems, missingSourceItems("lifecycle state", required.LifecycleStates, ledger.Rows, func(row Row) string { return row.State })...)
	problems = append(problems, missingExecutionItems("package", required.Packages, ledger.Rows, func(row Row) string { return row.Package })...)
	problems = append(problems, missingExecutionItems("renderable", required.Renderables, ledger.Rows, func(row Row) string { return row.Renderable })...)
	problems = append(problems, missingExecutionItems("kind", required.Kinds, ledger.Rows, func(row Row) string { return row.Kind })...)
	problems = append(problems, missingExecutionItems("route", required.Routes, ledger.Rows, func(row Row) string { return row.Route })...)
	problems = append(problems, missingExecutionItems("theme", required.Themes, ledger.Rows, func(row Row) string { return row.Theme })...)
	problems = append(problems, missingExecutionItems("state", required.States, ledger.Rows, func(row Row) string { return row.State })...)
	problems = append(problems, missingExecutionItems("lifecycle state", required.LifecycleStates, ledger.Rows, func(row Row) string { return row.State })...)

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

func ValidateClosure(ledger Ledger, required Inventory) error {
	if err := Validate(ledger, required); err != nil {
		return err
	}
	var open []string
	for _, row := range ledger.Rows {
		if row.ReceiptStatus == StatusBlocked || row.ReceiptStatus == StatusFailed {
			open = append(open, fmt.Sprintf("%s=%s", row.ID, row.ReceiptStatus))
		}
	}
	if len(open) > 0 {
		sort.Strings(open)
		return fmt.Errorf("closure has open rows: %s", strings.Join(open, ", "))
	}
	return nil
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

func missingSourceItems(axis string, items []SourceItem, rows []Row, value func(Row) string) []string {
	var missing []string
	for _, item := range items {
		if !anyRow(rows, func(row Row) bool { return value(row) == item.Value }) {
			missing = append(missing, fmt.Sprintf("missing mandatory %s %s", axis, item.Value))
		}
	}
	return missing
}

func missingExecutionItems(axis string, items []SourceItem, rows []Row, value func(Row) string) []string {
	var missing []string
	for _, item := range items {
		if !anyRow(rows, func(row Row) bool { return row.Class == ClassExecution && value(row) == item.Value }) {
			missing = append(missing, fmt.Sprintf("missing mandatory %s execution %s", axis, item.Value))
		}
	}
	return missing
}

func anyRow(rows []Row, predicate func(Row) bool) bool {
	for _, row := range rows {
		if predicate(row) {
			return true
		}
	}
	return false
}

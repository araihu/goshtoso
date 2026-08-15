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
	"strconv"
	"strings"
)

var RequiredReceiptTasks = []string{
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

const prerequisiteReceiptSchemaVersion = 1

var authorizedPrerequisiteProvenance = map[string]bool{
	"independent-review": true,
}

type ReceiptInput struct {
	Task           string `json:"task"`
	Path           string `json:"path"`
	ExpectedSHA256 string `json:"expected_sha256"`
}

// PrerequisiteReceipt is the typed, content-authenticated prerequisite input
// for T-GS-010. ReceiptInput authenticates these bytes; this envelope binds
// their semantic claim to the exact candidate identity.
type PrerequisiteReceipt struct {
	SchemaVersion int                   `json:"schema_version"`
	Task          string                `json:"task"`
	Disposition   string                `json:"disposition"`
	Candidate     PrerequisiteCandidate `json:"candidate"`
	Report        PrerequisiteReport    `json:"report"`
	GateEvidence  []PrerequisiteGate    `json:"gate_evidence"`
	Provenance    string                `json:"provenance"`
	ReceiptID     string                `json:"receipt_id"`
}

type PrerequisiteCandidate struct {
	Commit string `json:"commit"`
	Tree   string `json:"tree"`
}

type PrerequisiteReport struct {
	SHA256 string `json:"sha256"`
}

type PrerequisiteGate struct {
	Name   string `json:"name"`
	Tool   string `json:"tool"`
	SHA256 string `json:"sha256"`
}

type GenerationConfig struct {
	RepoRoot         string
	SourceCommit     string
	SourceTree       string
	Receipts         []ReceiptInput
	ATBlockerReceipt string
}

func GenerateSkeleton(config GenerationConfig) (Ledger, Inventory, error) {
	if config.RepoRoot == "" || config.SourceCommit == "" || config.SourceTree == "" {
		return Ledger{}, Inventory{}, fmt.Errorf("repo root, source commit, and source tree are required")
	}
	if err := verifyGitIdentity(config.RepoRoot, config.SourceCommit, config.SourceTree); err != nil {
		return Ledger{}, Inventory{}, err
	}
	receiptMetadata, err := authenticateReceipts(config.Receipts, config.SourceCommit, config.SourceTree)
	if err != nil {
		return Ledger{}, Inventory{}, err
	}
	inventory, err := DeriveInventory(config.RepoRoot)
	if err != nil {
		return Ledger{}, Inventory{}, err
	}

	ledger := Ledger{
		SchemaVersion: SchemaVersion,
		SourceCommit:  config.SourceCommit,
		SourceTree:    config.SourceTree,
		Metadata:      receiptMetadata,
	}
	ledger.Metadata["historical.package_count"] = "54"
	ledger.Metadata["historical.directory_count"] = "52"
	ledger.Metadata["historical.kind_count"] = "83"
	ledger.Metadata["historical.route_count"] = "50"
	ledger.Metadata["historical.theme_count"] = "16"
	ledger.Metadata["current.package_count"] = strconv.Itoa(len(inventory.Packages))
	ledger.Metadata["current.renderable_count"] = strconv.Itoa(len(inventory.Renderables))
	ledger.Metadata["current.kind_count"] = strconv.Itoa(len(inventory.Kinds))
	ledger.Metadata["current.route_count"] = strconv.Itoa(len(inventory.Routes))
	ledger.Metadata["current.lifecycle_state_count"] = strconv.Itoa(len(inventory.LifecycleStates))
	ledger.Metadata["current.theme_count"] = strconv.Itoa(len(inventory.Themes))

	ledger.Rows = appendSourceRows(ledger.Rows, "package", inventory.Packages, func(row *Row, value string) { row.Package = value })
	ledger.Rows = appendSourceRows(ledger.Rows, "renderable", inventory.Renderables, func(row *Row, value string) { row.Renderable = value })
	ledger.Rows = appendSourceRows(ledger.Rows, "kind", inventory.Kinds, func(row *Row, value string) { row.Kind = value })
	ledger.Rows = appendSourceRows(ledger.Rows, "state", inventory.States, func(row *Row, value string) { row.State = value })
	ledger.Rows = appendSourceRows(ledger.Rows, "lifecycle-state", inventory.LifecycleStates, func(row *Row, value string) { row.State = value })
	ledger.Rows = appendSourceRows(ledger.Rows, "theme", inventory.Themes, func(row *Row, value string) { row.Theme = value })
	ledger.Rows = appendExecutionRows(ledger.Rows, "package", inventory.Packages, config.ATBlockerReceipt, func(row *Row, value string) { row.Package = value })
	ledger.Rows = appendExecutionRows(ledger.Rows, "renderable", inventory.Renderables, config.ATBlockerReceipt, func(row *Row, value string) { row.Renderable = value })
	ledger.Rows = appendExecutionRows(ledger.Rows, "kind", inventory.Kinds, config.ATBlockerReceipt, func(row *Row, value string) { row.Kind = value })
	ledger.Rows = appendExecutionRows(ledger.Rows, "state", inventory.States, config.ATBlockerReceipt, func(row *Row, value string) { row.State = value })
	ledger.Rows = appendExecutionRows(ledger.Rows, "lifecycle-state", inventory.LifecycleStates, config.ATBlockerReceipt, func(row *Row, value string) { row.State = value })
	ledger.Rows = appendExecutionRows(ledger.Rows, "theme", inventory.Themes, config.ATBlockerReceipt, func(row *Row, value string) { row.Theme = value })

	for _, route := range inventory.Routes {
		mappings, err := ChecklistMappingsForRoute(route.Value)
		if err != nil {
			return Ledger{}, Inventory{}, err
		}
		urls := make([]string, 0, len(mappings))
		for _, mapping := range mappings {
			urls = append(urls, mapping.URL)
		}
		ledger.Rows = append(ledger.Rows, Row{
			ID:                ledgerSourceRowID(ClassInventory, "route", route.Value),
			Class:             ClassInventory,
			Route:             route.Value,
			ChecklistURLs:     urls,
			ChecklistMappings: mappings,
			Sources:           []SourceRef{route.Source},
			Applicability:     Applicable,
			ReceiptStatus:     StatusExecuted,
			Receipt:           "source-derived route and checklist mapping",
		})
		execution := blockedRow(ledgerSourceRowID(ClassExecution, "route", route.Value), route.Source, config.ATBlockerReceipt, func(row *Row) { row.Route = route.Value })
		execution.ChecklistURLs = urls
		execution.ChecklistMappings = mappings
		ledger.Rows = append(ledger.Rows, execution)
	}

	roadmap := SourceRef{Path: "roadmap:compliance-roadmap-final-v6.md", Symbol: "B-FULL"}
	for _, mode := range []string{"light", "dark"} {
		ledger.Rows = append(ledger.Rows, blockedRow("mode/"+mode, roadmap, config.ATBlockerReceipt, func(row *Row) { row.Mode = mode }))
	}
	for _, width := range RequiredViewportWidths(inventory.BreakpointEdges) {
		value := width
		ledger.Rows = append(ledger.Rows, blockedRow(fmt.Sprintf("viewport/%d", width), SourceRef{Path: "assets/styles.css", Symbol: fmt.Sprintf("compiled viewport %d", width)}, config.ATBlockerReceipt, func(row *Row) {
			row.Viewport = value
			row.Breakpoint = breakpointLabel(value, inventory.BreakpointEdges)
		}))
	}
	for _, zoom := range []int{100, 200} {
		value := zoom
		ledger.Rows = append(ledger.Rows, blockedRow(fmt.Sprintf("zoom/%d", zoom), roadmap, config.ATBlockerReceipt, func(row *Row) { row.Zoom = value }))
	}
	for _, motion := range []string{"normal", "reduced"} {
		value := motion
		ledger.Rows = append(ledger.Rows, blockedRow("motion/"+motion, roadmap, config.ATBlockerReceipt, func(row *Row) { row.Motion = value }))
	}
	for _, input := range []string{"mouse", "keyboard", "touch"} {
		value := input
		ledger.Rows = append(ledger.Rows, blockedRow("input/"+input, roadmap, config.ATBlockerReceipt, func(row *Row) { row.Input = value }))
	}
	for _, at := range []string{"safari-voiceover", "chromium-screen-reader"} {
		for _, exemplar := range RequiredATExemplars {
			value := at
			row := blockedRow("at/"+at+"/"+exemplar.Name, exemplar.Source, config.ATBlockerReceipt, func(row *Row) {
				row.AT = value
				row.Route = exemplar.Route
				row.State = exemplar.State
			})
			row.Rationale = "real AT caption/cursor screenshot and trace pending; automated scans are not substitution"
			ledger.Rows = append(ledger.Rows, row)
		}
	}

	sort.Slice(ledger.Rows, func(left, right int) bool { return ledger.Rows[left].ID < ledger.Rows[right].ID })
	if err := Validate(ledger, inventory); err != nil {
		return Ledger{}, Inventory{}, fmt.Errorf("generated skeleton is structurally invalid: %w", err)
	}
	return ledger, inventory, nil
}

func verifyGitIdentity(repoRoot, commit, tree string) error {
	for _, check := range []struct {
		name string
		args []string
		want string
	}{
		{name: "commit", args: []string{"rev-parse", "HEAD"}, want: commit},
		{name: "tree", args: []string{"rev-parse", "HEAD^{tree}"}, want: tree},
	} {
		command := exec.Command("git", append([]string{"-C", repoRoot}, check.args...)...)
		output, err := command.Output()
		if err != nil {
			return fmt.Errorf("resolve source %s: %w", check.name, err)
		}
		got := strings.TrimSpace(string(output))
		if got != check.want {
			return fmt.Errorf("source %s = %s, want %s", check.name, got, check.want)
		}
	}
	return nil
}

func authenticateReceipts(inputs []ReceiptInput, sourceCommit, sourceTree string) (map[string]string, error) {
	byTask := make(map[string]ReceiptInput, len(inputs))
	paths := make(map[string]string, len(inputs))
	for _, input := range inputs {
		if input.Task == "" || input.Path == "" || input.ExpectedSHA256 == "" {
			return nil, fmt.Errorf("receipt task, path, and expected SHA-256 are required")
		}
		if _, duplicate := byTask[input.Task]; duplicate {
			return nil, fmt.Errorf("duplicate receipt task %s", input.Task)
		}
		canonicalPath, err := filepath.Abs(input.Path)
		if err != nil {
			return nil, fmt.Errorf("resolve receipt %s path: %w", input.Task, err)
		}
		if other, duplicate := paths[canonicalPath]; duplicate {
			return nil, fmt.Errorf("aliased prerequisite receipt path %s serves both %s and %s", canonicalPath, other, input.Task)
		}
		paths[canonicalPath] = input.Task
		byTask[input.Task] = input
	}
	metadata := map[string]string{}
	receiptIDs := make(map[string]string, len(inputs))
	for _, task := range RequiredReceiptTasks {
		input, ok := byTask[task]
		if !ok {
			return nil, fmt.Errorf("missing mandatory receipt %s", task)
		}
		content, err := os.ReadFile(input.Path)
		if err != nil {
			return nil, fmt.Errorf("read receipt %s: %w", task, err)
		}
		digest := sha256.Sum256(content)
		got := hex.EncodeToString(digest[:])
		if got != strings.ToLower(input.ExpectedSHA256) {
			return nil, fmt.Errorf("receipt %s SHA-256 = %s, want %s", task, got, input.ExpectedSHA256)
		}
		var receipt PrerequisiteReceipt
		decoder := json.NewDecoder(strings.NewReader(string(content)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&receipt); err != nil {
			return nil, fmt.Errorf("parse prerequisite receipt %s: %w", task, err)
		}
		if err := validatePrerequisiteReceipt(receipt, task, sourceCommit, sourceTree); err != nil {
			return nil, err
		}
		if other, duplicate := receiptIDs[receipt.ReceiptID]; duplicate {
			return nil, fmt.Errorf("aliased prerequisite receipt identity %s used by %s and %s", receipt.ReceiptID, other, task)
		}
		receiptIDs[receipt.ReceiptID] = task
		metadata["receipt."+task+".path"] = input.Path
		metadata["receipt."+task+".sha256"] = got
		metadata["receipt."+task+".id"] = receipt.ReceiptID
		metadata["receipt."+task+".report_sha256"] = receipt.Report.SHA256
		metadata["receipt."+task+".provenance"] = receipt.Provenance
	}
	if len(byTask) != len(RequiredReceiptTasks) {
		return nil, fmt.Errorf("receipt set has %d entries, want exactly %d", len(byTask), len(RequiredReceiptTasks))
	}
	return metadata, nil
}

func validatePrerequisiteReceipt(receipt PrerequisiteReceipt, task, sourceCommit, sourceTree string) error {
	if receipt.SchemaVersion != prerequisiteReceiptSchemaVersion {
		return fmt.Errorf("prerequisite receipt %s schema version %d, want %d", task, receipt.SchemaVersion, prerequisiteReceiptSchemaVersion)
	}
	if receipt.Task != task {
		return fmt.Errorf("prerequisite receipt task %s, want %s", receipt.Task, task)
	}
	if receipt.Disposition != "accepted" {
		return fmt.Errorf("prerequisite receipt %s requires accepted disposition", task)
	}
	if receipt.Candidate.Commit != sourceCommit {
		return fmt.Errorf("prerequisite receipt %s candidate commit %s, want %s", task, receipt.Candidate.Commit, sourceCommit)
	}
	if receipt.Candidate.Tree != sourceTree {
		return fmt.Errorf("prerequisite receipt %s candidate tree %s, want %s", task, receipt.Candidate.Tree, sourceTree)
	}
	if !validPrerequisiteSHA256(receipt.Report.SHA256) {
		return fmt.Errorf("prerequisite receipt %s missing valid report SHA-256", task)
	}
	if !authorizedPrerequisiteProvenance[receipt.Provenance] {
		return fmt.Errorf("prerequisite receipt %s lacks authorized provenance", task)
	}
	if strings.TrimSpace(receipt.ReceiptID) == "" {
		return fmt.Errorf("prerequisite receipt %s missing receipt identity", task)
	}
	if len(receipt.GateEvidence) == 0 {
		return fmt.Errorf("prerequisite receipt %s missing gate evidence", task)
	}
	for index, gate := range receipt.GateEvidence {
		if strings.TrimSpace(gate.Name) == "" || strings.TrimSpace(gate.Tool) == "" || !validPrerequisiteSHA256(gate.SHA256) {
			return fmt.Errorf("prerequisite receipt %s gate evidence %d is invalid", task, index)
		}
	}
	return nil
}

func validPrerequisiteSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func appendSourceRows(rows []Row, axis string, items []SourceItem, assign func(*Row, string)) []Row {
	for _, item := range items {
		row := Row{
			ID:            ledgerSourceRowID(ClassInventory, axis, item.Value),
			Class:         ClassInventory,
			Sources:       []SourceRef{item.Source},
			Applicability: Applicable,
			ReceiptStatus: StatusExecuted,
			Receipt:       "current-source derivation",
		}
		if item.Action != "" {
			row.Rationale = "source-grounded action contract: " + item.Action
		}
		assign(&row, item.Value)
		rows = append(rows, row)
	}
	return rows
}

func appendExecutionRows(rows []Row, axis string, items []SourceItem, receipt string, assign func(*Row, string)) []Row {
	for _, item := range items {
		row := blockedRow(ledgerSourceRowID(ClassExecution, axis, item.Value), item.Source, receipt, func(row *Row) {
			assign(row, item.Value)
		})
		if item.Action != "" {
			row.Rationale += "; source-grounded action contract: " + item.Action
		}
		rows = append(rows, row)
	}
	return rows
}

func blockedRow(id string, source SourceRef, receipt string, assign func(*Row)) Row {
	if receipt == "" {
		receipt = "pending execution receipt"
	}
	row := Row{
		ID:            id,
		Class:         ClassExecution,
		Sources:       []SourceRef{source},
		Applicability: Applicable,
		ReceiptStatus: StatusBlocked,
		Receipt:       receipt,
		Rationale:     "mandatory execution row remains open",
	}
	assign(&row)
	return row
}

func breakpointLabel(width int, edges []int) string {
	for _, edge := range edges {
		switch width {
		case edge - 1:
			return fmt.Sprintf("edge-%d-minus-1", edge)
		case edge:
			return fmt.Sprintf("edge-%d", edge)
		case edge + 1:
			return fmt.Sprintf("edge-%d-plus-1", edge)
		}
	}
	return "baseline-or-intermediate"
}

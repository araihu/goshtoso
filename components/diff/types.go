// Package diff compares and presents complete source lines from any text.
// RowsFromText calculates line differences; Diff renders supplied rows without
// interpreting the source language.
package diff

import "github.com/a-h/templ"

// Operation identifies the comparison represented by a row.
type Operation string

const (
	// OperationUnchanged presents matching lines without change markers.
	OperationUnchanged Operation = ""
	// OperationInsert marks the after line as added; Before should be nil.
	OperationInsert Operation = "insert"
	// OperationRemove marks the before line as removed; After should be nil.
	OperationRemove Operation = "remove"
	// OperationReplace marks the before line as removed and the after line as added.
	OperationReplace Operation = "replace"
)

// Line is one complete source line. Text is escaped and preserves whitespace.
// An empty Text is a real blank line; a nil *Line is alignment padding.
type Line struct {
	// Text is the complete line without its newline separator. Spaces, tabs,
	// Unicode, and HTML-looking characters are preserved and escaped.
	Text string
	// Number is the optional source line number. Nonpositive numbers are omitted.
	Number int
}

// Row pairs complete lines in before-then-after reading order. Use RowsFromText
// or supply comparisons with application-owned normalization and validation.
// For unequal replacement blocks, use OperationReplace with a nil side for
// each unmatched line.
// Both-nil rows are omitted. Unknown operations render neutrally. Populated
// sides always render, even when they do not match the operation's contract.
type Row struct {
	Operation Operation
	Before    *Line
	After     *Line
}

// Config configures a server-rendered comparison. At widths below 640px each
// pair stacks before then after; absent sides are hidden. Long lines scroll
// horizontally within the shared region without changing the supplied text.
type Config struct {
	// ID optionally identifies the scroll region for fragment replacement.
	ID string
	// BeforeLabel names the baseline source; defaults to the Before expression.
	BeforeLabel string
	// AfterLabel names the proposed source; defaults to the After expression.
	AfterLabel string
	// Rows are comparisons in source order, supplied directly or by RowsFromText.
	// Empty or all-padding rows display the empty state. Identical sources use
	// unchanged rows.
	Rows []Row
	// MaxHeight bounds the scroll region; defaults to "32rem".
	MaxHeight string
	// AriaLabel names the focusable region; defaults to "BeforeLabel / AfterLabel".
	AriaLabel string
	// EmptyText is shown when Rows is empty; defaults to the EmptyText expression.
	EmptyText string
	// RootClass appends classes to the scroll region.
	RootClass string
	// RootAttrs appends integration attributes to the scroll region.
	RootAttrs templ.Attributes
}

func (cfg Config) hasLines() bool {
	for _, row := range cfg.Rows {
		if row.Before != nil || row.After != nil {
			return true
		}
	}
	return false
}

func (cfg Config) maxHeightStyle() string {
	height := cfg.MaxHeight
	if height == "" {
		height = "32rem"
	}
	return "max-height: " + height
}

func (row Row) beforeMarker() string {
	if row.Operation == OperationRemove || row.Operation == OperationReplace {
		return "−"
	}
	return " "
}

func (row Row) afterMarker() string {
	if row.Operation == OperationInsert || row.Operation == OperationReplace {
		return "+"
	}
	return " "
}

package diff

import (
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// RowsFromText compares complete lines using sergi/go-diff and returns rows with
// independent one-based source numbers. Adjacent removals and additions become
// replacement rows, with nil padding for the shorter side.
//
// Empty strings contain no lines. LF and CRLF terminate lines; a final separator
// does not create an extra blank line. Only separators are removed from displayed
// text. Separator differences still mark lines changed, even when their displayed
// text matches. All other whitespace and source bytes are preserved.
//
// The library's default one-second comparison timeout may produce coarser changes
// for difficult inputs without dropping lines. Rendering never invokes this
// helper implicitly; consumers may also provide their own Config.Rows.
func RowsFromText(before, after string) []Row {
	dmp := diffmatchpatch.New()
	beforeTokens, afterTokens, source := dmp.DiffLinesToRunes(before, after)
	changes := dmp.DiffCharsToLines(dmp.DiffMainRunes(beforeTokens, afterTokens, false), source)
	var rows []Row
	var removed, added []Line
	beforeNumber, afterNumber := 1, 1
	for _, change := range changes {
		switch change.Type {
		case diffmatchpatch.DiffEqual:
			rows = append(rows, changedRows(removed, added)...)
			removed, added = nil, nil
			left := numberedLines(change.Text, beforeNumber)
			right := numberedLines(change.Text, afterNumber)
			rows = append(rows, pairedRows(left, right, OperationUnchanged)...)
			beforeNumber += len(left)
			afterNumber += len(right)
		case diffmatchpatch.DiffDelete:
			lines := numberedLines(change.Text, beforeNumber)
			removed = append(removed, lines...)
			beforeNumber += len(lines)
		case diffmatchpatch.DiffInsert:
			lines := numberedLines(change.Text, afterNumber)
			added = append(added, lines...)
			afterNumber += len(lines)
		}
	}
	return append(rows, changedRows(removed, added)...)
}

func numberedLines(text string, first int) []Line {
	parts := strings.SplitAfter(text, "\n")
	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	lines := make([]Line, len(parts))
	for index, part := range parts {
		if before, ok := strings.CutSuffix(part, "\n"); ok {
			part = strings.TrimSuffix(before, "\r")
		}
		lines[index] = Line{Text: part, Number: first + index}
	}
	return lines
}

func changedRows(before, after []Line) []Row {
	operation := OperationReplace
	if len(before) == 0 {
		operation = OperationInsert
	} else if len(after) == 0 {
		operation = OperationRemove
	}
	return pairedRows(before, after, operation)
}

func pairedRows(before, after []Line, operation Operation) []Row {
	rows := make([]Row, max(len(before), len(after)))
	for index := range rows {
		rows[index].Operation = operation
		if index < len(before) {
			rows[index].Before = &before[index]
		}
		if index < len(after) {
			rows[index].After = &after[index]
		}
	}
	return rows
}

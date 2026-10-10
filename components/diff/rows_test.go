package diff_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/diff"
	"github.com/stretchr/testify/require"
)

func TestRowsFromText(t *testing.T) {
	line := func(text string, number int) *diff.Line { return &diff.Line{Text: text, Number: number} }
	for _, test := range []struct {
		name, before, after string
		want                []diff.Row
	}{
		{name: "empty"},
		{name: "identical without terminator", before: "same", after: "same", want: []diff.Row{{Before: line("same", 1), After: line("same", 1)}}},
		{name: "blank lines", before: "a\n\n", after: "a\n\n", want: []diff.Row{{Before: line("a", 1), After: line("a", 1)}, {Before: line("", 2), After: line("", 2)}}},
		{name: "only blank line added", after: "\n", want: []diff.Row{{Operation: diff.OperationInsert, After: line("", 1)}}},
		{name: "entire addition", after: "one\ntwo", want: []diff.Row{{Operation: diff.OperationInsert, After: line("one", 1)}, {Operation: diff.OperationInsert, After: line("two", 2)}}},
		{name: "entire removal", before: "one\ntwo\n", want: []diff.Row{{Operation: diff.OperationRemove, Before: line("one", 1)}, {Operation: diff.OperationRemove, Before: line("two", 2)}}},
		{name: "unequal replacement", before: "same\nold\nend\n", after: "same\nnew\nextra\nend\n", want: []diff.Row{
			{Before: line("same", 1), After: line("same", 1)},
			{Operation: diff.OperationReplace, Before: line("old", 2), After: line("new", 2)},
			{Operation: diff.OperationReplace, After: line("extra", 3)},
			{Before: line("end", 3), After: line("end", 4)},
		}},
		{name: "insert keeps independent numbers", before: "a\nb\nc\n", after: "a\nadded\nb\nc\n", want: []diff.Row{
			{Before: line("a", 1), After: line("a", 1)},
			{Operation: diff.OperationInsert, After: line("added", 2)},
			{Before: line("b", 2), After: line("b", 3)},
			{Before: line("c", 3), After: line("c", 4)},
		}},
		{name: "remove keeps independent numbers", before: "a\nremoved\nb\n", after: "a\nb\n", want: []diff.Row{
			{Before: line("a", 1), After: line("a", 1)},
			{Operation: diff.OperationRemove, Before: line("removed", 2)},
			{Before: line("b", 3), After: line("b", 2)},
		}},
		{name: "CRLF", before: "a\r\n\r\n", after: "a\r\n\r\n", want: []diff.Row{{Before: line("a", 1), After: line("a", 1)}, {Before: line("", 2), After: line("", 2)}}},
		{name: "newline style changes", before: "a\r\n", after: "a\n", want: []diff.Row{{Operation: diff.OperationReplace, Before: line("a", 1), After: line("a", 1)}}},
		{name: "final newline changes", before: "a", after: "a\n", want: []diff.Row{{Operation: diff.OperationReplace, Before: line("a", 1), After: line("a", 1)}}},
		{name: "bare carriage return", before: "a\rb\r", after: "a\rb\r", want: []diff.Row{{Before: line("a\rb\r", 1), After: line("a\rb\r", 1)}}},
		{name: "complete verbatim lines", before: "\t  Olá 世界 👋 <b>old</b>  \n", after: "\t  Olá 世界 👋 <b>new</b>  \n", want: []diff.Row{{Operation: diff.OperationReplace, Before: line("\t  Olá 世界 👋 <b>old</b>  ", 1), After: line("\t  Olá 世界 👋 <b>new</b>  ", 1)}}},
		{name: "invalid UTF8 bytes preserved", before: "\xffold\n", after: "\xffnew\n", want: []diff.Row{{Operation: diff.OperationReplace, Before: line("\xffold", 1), After: line("\xffnew", 1)}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, diff.RowsFromText(test.before, test.after))
		})
	}
}

func TestRowsFromTextPreservesRepeatedLinesAndLargeTokenSets(t *testing.T) {
	// Cross the ASCII and surrogate rune ranges used by the library's line
	// encoding, while retaining repeated lines that make alignment ambiguous.
	before := make([]string, 56000)
	for index := range before {
		before[index] = fmt.Sprintf("line %d · 世界", index)
		if index%100 == 0 {
			before[index] = "repeated"
		}
	}
	after := append([]string(nil), before...)
	after = append(after[:100:100], after[103:]...)
	after = append(after[:400:400], append([]string{"", "\tadded  "}, after[400:]...)...)
	rows := diff.RowsFromText(strings.Join(before, "\n"), strings.Join(after, "\n"))
	var left, right []string
	for _, row := range rows {
		require.True(t, row.Before != nil || row.After != nil)
		if row.Before != nil {
			left = append(left, row.Before.Text)
			require.Equal(t, len(left), row.Before.Number)
		}
		if row.After != nil {
			right = append(right, row.After.Text)
			require.Equal(t, len(right), row.After.Number)
		}
		if row.Operation == diff.OperationUnchanged {
			require.NotNil(t, row.Before)
			require.NotNil(t, row.After)
			require.Equal(t, row.Before.Text, row.After.Text)
		}
	}
	// Both source sequences must survive the algorithm and our row conversion,
	// regardless of where it aligns the repeated lines.
	require.Equal(t, before, left)
	require.Equal(t, after, right)
}

package diff_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/diff"
	"github.com/araihu/goshtoso/expressions"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func render(t *testing.T, component templ.Component) *html.Node {
	t.Helper()
	var out bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &out))
	doc, err := html.Parse(&out)
	require.NoError(t, err)
	return doc
}

func attr(node *html.Node, key string) string {
	for _, value := range node.Attr {
		if value.Key == key {
			return value.Val
		}
	}
	return ""
}

func find(node *html.Node, match func(*html.Node) bool) []*html.Node {
	var nodes []*html.Node
	if match(node) {
		nodes = append(nodes, node)
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		nodes = append(nodes, find(child, match)...)
	}
	return nodes
}

func byClass(node *html.Node, class string) []*html.Node {
	return find(node, func(n *html.Node) bool {
		return strings.Contains(" "+attr(n, "class")+" ", " "+class+" ")
	})
}

func nodeText(node *html.Node) string {
	var out strings.Builder
	for _, text := range find(node, func(n *html.Node) bool { return n.Type == html.TextNode }) {
		out.WriteString(text.Data)
	}
	return out.String()
}

func TestDiffOperationsKeepEveryPopulatedSideAndNumber(t *testing.T) {
	rows := []diff.Row{
		{Before: &diff.Line{Text: "same", Number: 4}, After: &diff.Line{Text: "same", Number: 8}},
		{Operation: diff.OperationInsert, After: &diff.Line{Text: "added", Number: 9}},
		{Operation: diff.OperationRemove, Before: &diff.Line{Text: "removed", Number: 5}},
		{Operation: diff.OperationReplace, Before: &diff.Line{Text: "old", Number: 6}, After: &diff.Line{Text: "new", Number: 10}},
		{Operation: diff.OperationReplace, After: &diff.Line{Text: "extra", Number: 11}},
		{Operation: "unknown", Before: &diff.Line{Text: "unknown before", Number: -1}, After: &diff.Line{Text: "unknown after"}},
		{Operation: diff.OperationInsert, Before: &diff.Line{Text: "keep malformed before"}, After: &diff.Line{Text: "still added"}},
		{Operation: diff.OperationRemove, Before: &diff.Line{Text: "still removed"}, After: &diff.Line{Text: "keep malformed after"}},
	}
	doc := render(t, diff.Diff(diff.Config{Rows: rows, BeforeLabel: "Original", AfterLabel: "Revised"}))
	populated := byClass(doc, "gs-diff-cell")
	wantText := []string{"same", "same", "added", "removed", "old", "new", "extra", "unknown before", "unknown after", "keep malformed before", "still added", "still removed", "keep malformed after"}
	wantMarkers := []string{" ", " ", "+", "−", "−", "+", "+", " ", " ", " ", "+", "−", " "}
	wantNumbers := []string{"4", "8", "9", "5", "6", "10", "11", "", "", "", "", "", ""}
	wantLabels := []string{"Original", "Revised", "Revised", "Original", "Original", "Revised", "Revised", "Original", "Revised", "Original", "Revised", "Original", "Revised"}
	require.Len(t, populated, len(wantText))
	for index, cell := range populated {
		code := find(cell, func(n *html.Node) bool { return n.Data == "code" && n.Type == html.ElementNode })
		require.Len(t, code, 1)
		require.Equal(t, wantText[index], nodeText(code[0]))
		require.Equal(t, wantMarkers[index], attr(cell, "data-diff-marker"))
		require.Equal(t, wantNumbers[index], nodeText(byClass(cell, "gs-diff-number")[0]))
		require.Equal(t, wantLabels[index], attr(cell, "aria-label"))
	}
	require.Len(t, byClass(doc, "gs-diff-row"), len(rows))
}

func TestDiffPreservesBlankLinesAndEscapedSource(t *testing.T) {
	source := "\t  <script>alert('x')</script><img src=x onerror=alert(1)> & \"Olá 世界 👋\"  " + strings.Repeat(" long", 100)
	doc := render(t, diff.Diff(diff.Config{Rows: []diff.Row{
		{},
		{Operation: diff.OperationReplace, Before: &diff.Line{Text: source, Number: 123456}, After: &diff.Line{Text: ""}},
		{Operation: diff.OperationReplace, After: &diff.Line{Text: "third replacement line"}},
	}}))
	require.Empty(t, find(doc, func(n *html.Node) bool { return n.Type == html.ElementNode && (n.Data == "script" || n.Data == "img") }))
	code := find(doc, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "code" })
	require.Len(t, code, 3)
	require.Equal(t, source, nodeText(code[0]))
	require.Empty(t, nodeText(code[1]), "a blank line remains a populated source element")
	require.Len(t, byClass(doc, "gs-diff-row"), 2, "both-nil rows are omitted")
	padding := byClass(doc, "gs-diff-padding")
	require.Len(t, padding, 1)
	require.Equal(t, "true", attr(padding[0], "aria-hidden"))
	require.Empty(t, nodeText(padding[0]))
	require.Equal(t, "123456", nodeText(byClass(doc, "gs-diff-number")[0]))
}

func TestDiffEmptyAndIdenticalInputs(t *testing.T) {
	for _, rows := range [][]diff.Row{nil, {}, {{}}, {{}, {}}} {
		doc := render(t, diff.Diff(diff.Config{Rows: rows}))
		require.Equal(t, "No lines to compare.", nodeText(byClass(doc, "gs-diff-empty")[0]))
		require.Empty(t, byClass(doc, "gs-diff-row"))
	}
	for _, text := range []string{"identical", ""} {
		doc := render(t, diff.Diff(diff.Config{Rows: []diff.Row{{Before: &diff.Line{Text: text}, After: &diff.Line{Text: text}}}}))
		require.Empty(t, byClass(doc, "gs-diff-empty"))
		cells := byClass(doc, "gs-diff-cell")
		require.Len(t, cells, 2)
		for _, cell := range cells {
			require.Equal(t, " ", attr(cell, "data-diff-marker"))
			require.Equal(t, "Unchanged", attr(byClass(cell, "gs-diff-marker")[0], "aria-label"))
		}
	}
}

func TestDiffExpressionsRespectInheritanceAndExplicitConfig(t *testing.T) {
	renderer := expressions.NewRenderer(expressions.Set{Diff: expressions.Diff{BeforeLabel: "application before", AfterLabel: "application after", AddedLabel: "application added", EmptyText: "application empty"}})
	ctx := expressions.With(context.Background(), expressions.Set{Diff: expressions.Diff{AfterLabel: "request after", RemovedLabel: "request removed"}})
	original := diff.Diff(diff.Config{Rows: []diff.Row{{Operation: diff.OperationReplace, Before: &diff.Line{Text: "old"}, After: &diff.Line{Text: "new"}}}})
	translated := original.WithExpressions(expressions.Diff{BeforeLabel: "component before", AddedLabel: "component added"})
	changed := translated.WithExpressions(expressions.Diff{BeforeLabel: "<original>"})
	var out bytes.Buffer
	require.NoError(t, renderer.Render(ctx, &out, changed))
	require.Contains(t, out.String(), `aria-label="&lt;original&gt; / request after"`)
	require.Contains(t, out.String(), `aria-label="component added"`)
	require.Contains(t, out.String(), `aria-label="request removed"`)
	require.Equal(t, components.KindDiff, changed.Kind())
	out.Reset()
	require.NoError(t, renderer.Render(ctx, &out, translated))
	require.Contains(t, out.String(), `aria-label="component before / request after"`)
	out.Reset()
	require.NoError(t, renderer.Render(ctx, &out, original))
	require.Contains(t, out.String(), `aria-label="application before / request after"`)
	explicit := diff.Diff(diff.Config{BeforeLabel: "explicit before", AfterLabel: "explicit after", EmptyText: "explicit empty"}).WithExpressions(expressions.Diff{BeforeLabel: "component before", EmptyText: "component empty"})
	out.Reset()
	require.NoError(t, renderer.Render(ctx, &out, explicit))
	require.Contains(t, out.String(), `aria-label="explicit before / explicit after"`)
	require.Contains(t, out.String(), "explicit empty")
	out.Reset()
	require.NoError(t, renderer.Render(ctx, &out, diff.Diff(diff.Config{})))
	require.Contains(t, out.String(), "application empty")
}

func TestDiffRootHooksDefaultsAndWriteErrors(t *testing.T) {
	doc := render(t, diff.Diff(diff.Config{ID: "comparison", RootClass: "consumer-class", RootAttrs: templ.Attributes{"data-consumer": "value"}, AriaLabel: "Text review", MaxHeight: "180px"}))
	root := byClass(doc, "gs-diff")[0]
	require.Equal(t, "comparison", attr(root, "id"))
	require.Contains(t, attr(root, "class"), "consumer-class")
	require.Equal(t, "value", attr(root, "data-consumer"))
	require.Equal(t, "Text review", attr(root, "aria-label"))
	require.Equal(t, "region", attr(root, "role"))
	require.Equal(t, "0", attr(root, "tabindex"))
	require.Equal(t, "max-height: 180px;", attr(root, "style"))
	defaults := byClass(render(t, diff.Diff(diff.Config{})), "gs-diff")[0]
	require.Empty(t, attr(defaults, "id"))
	require.Equal(t, "max-height: 32rem;", attr(defaults, "style"))
	require.Equal(t, "Before / After", attr(defaults, "aria-label"))
	err := diff.Diff(diff.Config{}).Render(context.Background(), failedWriter{})
	require.ErrorIs(t, err, io.ErrClosedPipe)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	require.True(t, errors.Is(diff.Diff(diff.Config{}).Render(cancelled, io.Discard), context.Canceled))
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

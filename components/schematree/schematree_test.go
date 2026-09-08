package schematree_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/schematree"
	"golang.org/x/net/html"
)

func TestDescriptionsSupportEscapedTextAndBlockComponents(t *testing.T) {
	cfg := schematree.Config{Nodes: []schematree.Node{
		{Name: "plain", Description: `<script>alert(1)</script>`},
		{Name: "parent", Description: "fallback must not render", DescriptionContent: templ.Raw(`<p>Use <strong>either</strong> value.</p><ul><li><a href="/help">Help</a></li></ul>`), Children: []schematree.Node{{Name: "child", Required: true}}},
	}}
	var out bytes.Buffer
	if err := schematree.SchemaTree(cfg).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	markup := out.String()
	for _, want := range []string{"&lt;script&gt;", "<strong>either</strong>", `href="/help"`, "required", "<details open"} {
		if !strings.Contains(markup, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(markup, "fallback must not render") || strings.Contains(markup, "<script>") {
		t.Fatal("description precedence/escaping failed")
	}
	doc, err := html.Parse(strings.NewReader(markup))
	if err != nil {
		t.Fatal(err)
	}
	var visit func(*html.Node, bool)
	visit = func(n *html.Node, summary bool) {
		summary = summary || n.Data == "summary"
		if summary && (n.Data == "p" || n.Data == "ul" || n.Data == "a") {
			t.Errorf("rich description entered disclosure header: %s", n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c, summary)
		}
	}
	visit(doc, false)
}

func TestSlotErrorAndCancellationPropagate(t *testing.T) {
	failure := errors.New("render failed")
	content := templ.ComponentFunc(func(context.Context, io.Writer) error { return failure })
	err := schematree.SchemaTree(schematree.Config{Nodes: []schematree.Node{{Name: "field", DescriptionContent: content}}}).Render(context.Background(), io.Discard)
	if !errors.Is(err, failure) {
		t.Fatalf("slot error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = schematree.SchemaTree(schematree.Config{Nodes: []schematree.Node{{Name: "field"}}}).Render(ctx, io.Discard)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation = %v", err)
	}
}

func TestOptionalLabelsAreOptIn(t *testing.T) {
	for _, showOptional := range []bool{false, true} {
		var out bytes.Buffer
		cfg := schematree.Config{Nodes: []schematree.Node{
			{Name: "optionalField", ShowOptional: showOptional},
			{Name: "requiredField", Required: true, ShowOptional: showOptional},
		}}
		if err := schematree.SchemaTree(cfg).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		if got := strings.Contains(out.String(), `data-required="false"`); got != showOptional {
			t.Errorf("optional label visible=%t with ShowOptional=%t", got, showOptional)
		}
		if strings.Count(out.String(), `data-required="true"`) != 1 {
			t.Fatal("required field must remain labeled")
		}
	}
}

func TestLazyChildrenRemainCallerOwned(t *testing.T) {
	var out bytes.Buffer
	cfg := schematree.Config{Nodes: []schematree.Node{{
		Name: "reference", Collapsed: true,
		RootAttrs:       templ.Attributes{"data-schema-id": "stable-id"},
		ChildrenContent: templ.Raw(`<div hx-get="/fragments/schema.html" hx-trigger="revealed once" hx-swap="outerHTML">Load schema</div>`),
	}}}
	if err := schematree.SchemaTree(cfg).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`data-schema-id="stable-id"`, `hx-get="/fragments/schema.html"`, `hx-trigger="revealed once"`, `hx-swap="outerHTML"`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing lazy fragment contract %q", want)
		}
	}
	if strings.Contains(out.String(), "<details open") {
		t.Fatal("lazy branch should start closed")
	}
}

package inlinecode

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
)

func TestInlineCodeRendersSemanticEscapedText(t *testing.T) {
	var output bytes.Buffer
	component := InlineCode(`<script>alert("x")</script>`)
	if err := component.Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}

	html := output.String()
	for _, want := range []string{
		"<code",
		"font-mono",
		"bg-surface-alt",
		"text-on-surface-strong",
		"dark:bg-surface-dark-alt",
		"dark:text-on-surface-dark-strong",
		`&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt;`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("InlineCode render missing %q in %s", want, html)
		}
	}
	if strings.Contains(html, "<script>") {
		t.Fatalf("InlineCode rendered unsafe text: %s", html)
	}
	if component.Kind() != components.KindInlineCode {
		t.Fatalf("InlineCode kind = %q, want %q", component.Kind(), components.KindInlineCode)
	}
}

func TestInlineCodeSupportsRootHooks(t *testing.T) {
	var output bytes.Buffer
	if err := InlineCode("muamba.yaml",
		WithRootClass("consumer-code"),
		WithRootAttrs(templ.Attributes{"data-inline-code": "manifest"}),
	).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}

	html := output.String()
	for _, want := range []string{`class="`, "consumer-code", `data-inline-code="manifest"`, "muamba.yaml"} {
		if !strings.Contains(html, want) {
			t.Fatalf("InlineCode root hooks missing %q in %s", want, html)
		}
	}
}

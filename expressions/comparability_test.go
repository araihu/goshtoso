package expressions_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/codeblock"
	"github.com/araihu/goshtoso/components/pagination"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/expressions"
)

// These declarations also protect source compatibility for map keys.
var (
	_ = map[codeblock.Instance]bool{}
	_ = map[pagination.Instance]bool{}
	_ = map[table.ImageCellInstance]bool{}
)

func TestComparableInstancesAndImmutableOverrides(t *testing.T) {
	original := codeblock.CodeBlock(codeblock.Config{Language: "go", Code: "hello"})
	identical := codeblock.CodeBlock(codeblock.Config{Language: "go", Code: "hello"})
	if original != identical {
		t.Fatal("unconfigured instances no longer compare by config")
	}
	translated := original.WithExpressions(expressions.CodeBlock{CopyLabel: "Copiar"})
	copy := translated
	changed := translated.WithExpressions(expressions.CodeBlock{CopyLabel: "Kopieren"})
	if translated != copy || translated == changed {
		t.Fatal("override identity is not preserved across copies")
	}
	// Interface comparisons must not panic after adding function-bearing expressions.
	var a, b any = translated, copy
	if a != b {
		t.Fatal("interface equality changed")
	}
	for _, tc := range []struct {
		component codeblock.Instance
		want      string
	}{{original, "Copy"}, {translated, "Copiar"}, {changed, "Kopieren"}} {
		var out bytes.Buffer
		if err := tc.component.Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `data-code-block-idle-text="`+tc.want+`"`) {
			t.Fatalf("override mutated another instance: %s", out.String())
		}
	}
}

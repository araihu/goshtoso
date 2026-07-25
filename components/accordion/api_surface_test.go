package accordion

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccordionItemRenderDataIsAbsentFromConsumerAPI(t *testing.T) {
	typesFile, err := parser.ParseFile(
		token.NewFileSet(),
		"types.go",
		nil,
		parser.SkipObjectResolution,
	)
	require.NoError(t, err)

	for _, declaration := range typesFile.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range general.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			require.NotEqual(
				t,
				"AccordionItemData",
				typeSpec.Name.Name,
				"per-item render data must not appear in go doc",
			)
		}
	}

	for _, path := range []string{
		filepath.Join("..", "..", ".claude", "skills", "using-goshtoso", "components-reference.md"),
		filepath.Join("..", "..", ".agents", "skills", "using-goshtoso", "references", "components-reference.md"),
	} {
		reference, err := os.ReadFile(path)
		require.NoError(t, err)
		require.NotContains(t, strings.ToLower(string(reference)), "accordionitemdata")
	}
}

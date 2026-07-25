package radio

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInputAttrsPublicDocsDescribeAppendSemantics(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "types.go", nil, parser.ParseComments)
	require.NoError(t, err)
	require.NotNil(t, file.Doc, "types.go must retain the radio package comment")

	packageDoc := normalizeDoc(file.Doc.Text())
	require.Contains(
		t,
		packageDoc,
		"Appended after modeled attributes; conflicting keys serialize duplicates and are not reliable overrides. Use typed Config fields for modeled attributes.",
	)
	requireNoOverridePromise(t, packageDoc)

	var inputAttrsDoc string
	ast.Inspect(file, func(node ast.Node) bool {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok || typeSpec.Name.Name != "Config" {
			return true
		}
		config, ok := typeSpec.Type.(*ast.StructType)
		require.True(t, ok)
		for _, field := range config.Fields.List {
			if len(field.Names) == 1 && field.Names[0].Name == "InputAttrs" {
				require.NotNil(t, field.Doc, "Config.InputAttrs must have a public Go doc comment")
				inputAttrsDoc = field.Doc.Text()
				return false
			}
		}
		return false
	})

	require.Equal(
		t,
		"InputAttrs contains additional non-conflicting attributes appended to the input. Conflicting modeled keys serialize duplicate attributes rather than reliably overriding them; use typed fields for modeled attributes.",
		normalizeDoc(inputAttrsDoc),
	)
	requireNoOverridePromise(t, inputAttrsDoc)
}

func normalizeDoc(doc string) string {
	return strings.Join(strings.Fields(doc), " ")
}

func requireNoOverridePromise(t *testing.T, doc string) {
	t.Helper()
	lowerDoc := strings.ToLower(doc)
	require.NotContains(t, lowerDoc, "always override")
	require.NotContains(t, lowerDoc, "wins on conflict")
}

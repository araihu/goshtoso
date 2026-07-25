package components

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/demo"
	"github.com/stretchr/testify/require"
)

const componentModulePath = "github.com/araihu/goshtoso/components"

type packageDeclarations struct {
	name         string
	declarations map[string]struct{}
}

func TestEveryDocumentedDeclarationSignatureMatchesComponentSource(t *testing.T) {
	index := componentDeclarationIndex(t)
	var constructorCount int
	var declarationCount int

	keys := make([]string, 0, len(Demos))
	for key := range Demos {
		if strings.HasPrefix(key, "components/") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	for _, key := range keys {
		entry := Demos[key]
		eligible := declarationPackagesForDemo(t, key, entry.API)
		for _, section := range entry.API {
			if section.Constructor != "" {
				constructorCount++
				t.Run(key+"/"+section.ID+"/constructor", func(t *testing.T) {
					require.NoError(t, documentedConstructorError(index, eligible, section.Constructor))
				})
			}
			for _, prop := range section.Props {
				signature := strings.TrimSpace(prop.Signature)
				if !isDeclarationSignature(signature) {
					continue
				}
				declarationCount++
				t.Run(key+"/"+section.ID+"/"+prop.Name, func(t *testing.T) {
					require.NoError(t, documentedDeclarationError(index, eligible, signature))
				})
			}
		}
	}
	require.Positive(t, constructorCount)
	require.Positive(t, declarationCount)
	t.Logf("validated %d constructors and %d declaration signature rows", constructorCount, declarationCount)
}

func TestDeclarationIndexRejectsReturnTypeAndVariadicDrift(t *testing.T) {
	index := componentDeclarationIndex(t)

	t.Run("return type", func(t *testing.T) {
		err := documentedDeclarationError(
			index,
			[]string{componentModulePath + "/combobox"},
			"func Handler(cfg Config, provider OptionsProvider) http.HandlerFunc",
		)
		require.ErrorContains(t, err, "does not match an exported declaration")
	})

	t.Run("variadic parameter", func(t *testing.T) {
		err := documentedDeclarationError(
			index,
			[]string{componentModulePath + "/tooltip"},
			"func Tooltip(id, label string, options []Option) Instance",
		)
		require.ErrorContains(t, err, "does not match an exported declaration")
	})
}

func componentDeclarationIndex(t *testing.T) map[string]packageDeclarations {
	t.Helper()

	root := filepath.Clean("../../../../../components")
	fset := token.NewFileSet()
	index := make(map[string]packageDeclarations)

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() ||
			!strings.HasSuffix(path, ".go") ||
			strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		importPath, err := componentImportPath(root, path)
		if err != nil {
			return err
		}
		pkg := index[importPath]
		if pkg.name == "" {
			pkg.name = file.Name.Name
			pkg.declarations = make(map[string]struct{})
		}
		if pkg.name != file.Name.Name {
			return fmt.Errorf("%s mixes packages %s and %s", importPath, pkg.name, file.Name.Name)
		}

		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				if !declaration.Name.IsExported() {
					continue
				}
				text, err := formatFunctionDeclaration(fset, declaration)
				if err != nil {
					return fmt.Errorf("%s: %w", path, err)
				}
				pkg.declarations[text] = struct{}{}
			case *ast.GenDecl:
				texts, err := formatExportedGeneralDeclarations(fset, declaration)
				if err != nil {
					return fmt.Errorf("%s: %w", path, err)
				}
				for _, text := range texts {
					pkg.declarations[text] = struct{}{}
				}
			}
		}
		index[importPath] = pkg
		return nil
	})
	require.NoError(t, err)
	return index
}

func componentImportPath(root string, path string) (string, error) {
	relativeDir, err := filepath.Rel(root, filepath.Dir(path))
	if err != nil {
		return "", err
	}
	if relativeDir == "." {
		return componentModulePath, nil
	}
	return componentModulePath + "/" + filepath.ToSlash(relativeDir), nil
}

func formatFunctionDeclaration(fset *token.FileSet, declaration *ast.FuncDecl) (string, error) {
	bodyless := &ast.FuncDecl{
		Recv: declaration.Recv,
		Name: declaration.Name,
		Type: declaration.Type,
	}
	return formatDeclarationNode(fset, bodyless)
}

func formatExportedGeneralDeclarations(fset *token.FileSet, declaration *ast.GenDecl) ([]string, error) {
	if declaration.Tok != token.TYPE && declaration.Tok != token.CONST {
		return nil, nil
	}

	var texts []string
	for _, spec := range declaration.Specs {
		switch spec := spec.(type) {
		case *ast.TypeSpec:
			if declaration.Tok != token.TYPE || !spec.Name.IsExported() {
				continue
			}
			text, err := formatDeclarationNode(fset, &ast.GenDecl{
				Tok:   token.TYPE,
				Specs: []ast.Spec{spec},
			})
			if err != nil {
				return nil, err
			}
			texts = append(texts, text)
		case *ast.ValueSpec:
			if declaration.Tok != token.CONST || len(spec.Values) == 0 {
				continue
			}
			for index, name := range spec.Names {
				if !name.IsExported() || len(spec.Names) != len(spec.Values) {
					continue
				}
				single := &ast.ValueSpec{
					Names:  []*ast.Ident{name},
					Type:   spec.Type,
					Values: []ast.Expr{spec.Values[index]},
				}
				text, err := formatDeclarationNode(fset, &ast.GenDecl{
					Tok:   token.CONST,
					Specs: []ast.Spec{single},
				})
				if err != nil {
					return nil, err
				}
				texts = append(texts, text)
			}
		}
	}
	return texts, nil
}

func formatDeclarationNode(fset *token.FileSet, node ast.Node) (string, error) {
	var buffer bytes.Buffer
	if err := format.Node(&buffer, fset, node); err != nil {
		return "", err
	}
	return normalizeDeclaration(buffer.String())
}

func normalizeDeclaration(text string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "contract.go", "package contract\n"+text, 0)
	if err != nil {
		return "", fmt.Errorf("parse %q: %w", text, err)
	}
	if len(file.Decls) != 1 {
		return "", fmt.Errorf("expected one declaration in %q", text)
	}
	var buffer bytes.Buffer
	if err := format.Node(&buffer, fset, file.Decls[0]); err != nil {
		return "", err
	}
	return strings.TrimSpace(buffer.String()), nil
}

func declarationPackagesForDemo(t *testing.T, key string, sections []demo.APISection) []string {
	t.Helper()

	seen := make(map[string]struct{})
	for _, section := range sections {
		if section.StructType != nil {
			seen[section.StructType.PkgPath()] = struct{}{}
		}
	}
	if len(seen) == 0 {
		fallbacks := map[string]string{
			"components/dependencies": componentModulePath + "/head",
			"components/kbd":          componentModulePath + "/kbd",
			"components/link":         componentModulePath + "/link",
			"components/tooltip":      componentModulePath + "/tooltip",
		}
		fallback, ok := fallbacks[key]
		require.Truef(t, ok, "%s has declaration metadata but no StructType package or fallback", key)
		seen[fallback] = struct{}{}
	}

	paths := make([]string, 0, len(seen))
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func isDeclarationSignature(signature string) bool {
	return strings.HasPrefix(signature, "func ") ||
		strings.HasPrefix(signature, "type ") ||
		strings.HasPrefix(signature, "const ")
}

func documentedDeclarationError(
	index map[string]packageDeclarations,
	eligible []string,
	signature string,
) error {
	normalized, err := normalizeDeclaration(signature)
	if err != nil {
		return err
	}
	return exactDeclarationMatchError(index, eligible, "", normalized)
}

func documentedConstructorError(
	index map[string]packageDeclarations,
	eligible []string,
	constructor string,
) error {
	openParen := strings.Index(constructor, "(")
	if openParen < 1 {
		return fmt.Errorf("constructor %q has no parameter list", constructor)
	}
	qualifiedName := constructor[:openParen]
	dot := strings.LastIndex(qualifiedName, ".")
	if dot < 1 || dot == len(qualifiedName)-1 {
		return fmt.Errorf("constructor %q must use a package-qualified name", constructor)
	}
	packageName := qualifiedName[:dot]
	functionName := qualifiedName[dot+1:]
	normalized, err := normalizeDeclaration("func " + functionName + constructor[openParen:])
	if err != nil {
		return err
	}
	return exactDeclarationMatchError(index, eligible, packageName, normalized)
}

func exactDeclarationMatchError(
	index map[string]packageDeclarations,
	eligible []string,
	requiredPackageName string,
	normalized string,
) error {
	var matches []string
	for _, path := range eligible {
		pkg, ok := index[path]
		if !ok || (requiredPackageName != "" && pkg.name != requiredPackageName) {
			continue
		}
		if _, ok := pkg.declarations[normalized]; ok {
			matches = append(matches, path)
		}
	}
	if len(matches) != 1 {
		return fmt.Errorf(
			"%q does not match an exported declaration in eligible packages %v (matches %v)",
			normalized,
			eligible,
			matches,
		)
	}
	return nil
}

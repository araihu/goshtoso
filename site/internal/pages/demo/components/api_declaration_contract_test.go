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
	"slices"
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

func TestEveryDeclarationSectionBindsToOneExactSourcePackage(t *testing.T) {
	for key, entry := range Demos {
		if !strings.HasPrefix(key, "components/") {
			continue
		}
		for _, section := range entry.API {
			sourcePackage, err := sourcePackageForSection(section)
			require.NoErrorf(t, err, "%s.%s", key, section.ID)
			require.Truef(
				t,
				strings.HasPrefix(sourcePackage, componentModulePath+"/"),
				"%s.%s source package %q",
				key,
				section.ID,
				sourcePackage,
			)
		}
	}
}

func TestCanonicalDeclarationInventoryIsExact(t *testing.T) {
	got := collectDeclarationInventory(Demos)
	want := strings.Fields(canonicalDeclarationInventory)
	require.Len(t, want, 139)
	require.Len(t, filterInventory(want, "constructor|"), 59)
	require.Len(t, filterInventory(want, "declaration|"), 80)
	require.NoError(t, declarationInventoryError(want, got))
}

func TestCanonicalDeclarationInventoryRejectsDeletedRows(t *testing.T) {
	want := strings.Fields(canonicalDeclarationInventory)

	t.Run("Table.NextSortDir", func(t *testing.T) {
		entries := cloneDemoEntries()
		entry := entries["components/table"]
		for sectionIndex := range entry.API {
			section := &entry.API[sectionIndex]
			if section.ID != "config-helpers" {
				continue
			}
			for propIndex := range section.Props {
				if section.Props[propIndex].Name == "NextSortDir" {
					section.Props = slices.Delete(section.Props, propIndex, propIndex+1)
					break
				}
			}
		}
		entries["components/table"] = entry

		err := declarationInventoryError(want, collectDeclarationInventory(entries))
		require.ErrorContains(
			t,
			err,
			"declaration|components/table|config-helpers|NextSortDir",
		)
	})

	t.Run("constructor", func(t *testing.T) {
		entries := cloneDemoEntries()
		entry := entries["components/accordion"]
		for sectionIndex := range entry.API {
			if entry.API[sectionIndex].ID == "accordionconfig" {
				entry.API[sectionIndex].Constructor = ""
			}
		}
		entries["components/accordion"] = entry

		err := declarationInventoryError(want, collectDeclarationInventory(entries))
		require.ErrorContains(
			t,
			err,
			"constructor|components/accordion|accordionconfig",
		)
	})
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
		for _, section := range entry.API {
			sourcePackage, err := sourcePackageForSection(section)
			require.NoErrorf(t, err, "%s.%s source package", key, section.ID)
			if section.Constructor != "" {
				constructorCount++
				t.Run(key+"/"+section.ID+"/constructor", func(t *testing.T) {
					require.NoError(t, documentedConstructorError(index, sourcePackage, section.Constructor))
				})
			}
			for _, prop := range section.Props {
				signature := strings.TrimSpace(prop.Signature)
				if !isDeclarationSignature(signature) {
					continue
				}
				declarationCount++
				t.Run(key+"/"+section.ID+"/"+prop.Name, func(t *testing.T) {
					require.NoError(t, documentedDeclarationError(index, sourcePackage, signature))
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
			componentModulePath+"/combobox",
			"func Handler(cfg Config, provider OptionsProvider) http.HandlerFunc",
		)
		require.ErrorContains(t, err, "does not match an exported declaration")
	})

	t.Run("variadic parameter", func(t *testing.T) {
		err := documentedDeclarationError(
			index,
			componentModulePath+"/tooltip",
			"func Tooltip(id, label string, options []Option) Instance",
		)
		require.ErrorContains(t, err, "does not match an exported declaration")
	})
}

func TestDeclarationIndexRejectsWrongPackageMatch(t *testing.T) {
	signature := "type ValidationType int"
	index := map[string]packageDeclarations{
		componentModulePath + "/form": {
			name: "form",
			declarations: map[string]struct{}{
				signature: {},
			},
		},
		componentModulePath + "/form/validation": {
			name:         "validation",
			declarations: map[string]struct{}{},
		},
	}

	err := documentedDeclarationError(
		index,
		componentModulePath+"/form/validation",
		signature,
	)
	require.ErrorContains(t, err, "does not match an exported declaration")
}

func sourcePackageForSection(section demo.APISection) (string, error) {
	if section.SourcePackage != "" {
		return section.SourcePackage, nil
	}
	if section.StructType != nil {
		return section.StructType.PkgPath(), nil
	}
	return "", fmt.Errorf("function-only section %q must declare SourcePackage", section.ID)
}

func collectDeclarationInventory(entries map[string]DemoEntry) []string {
	var inventory []string
	for key, entry := range entries {
		if !strings.HasPrefix(key, "components/") {
			continue
		}
		for _, section := range entry.API {
			if section.Constructor != "" {
				inventory = append(inventory, "constructor|"+key+"|"+section.ID)
			}
			for _, prop := range section.Props {
				if isDeclarationSignature(strings.TrimSpace(prop.Signature)) {
					inventory = append(inventory, "declaration|"+key+"|"+section.ID+"|"+prop.Name)
				}
			}
		}
	}
	sort.Strings(inventory)
	return inventory
}

func filterInventory(inventory []string, prefix string) []string {
	return slices.Collect(func(yield func(string) bool) {
		for _, key := range inventory {
			if strings.HasPrefix(key, prefix) && !yield(key) {
				return
			}
		}
	})
}

func cloneDemoEntries() map[string]DemoEntry {
	cloned := make(map[string]DemoEntry, len(Demos))
	for key, entry := range Demos {
		entry.API = slices.Clone(entry.API)
		for sectionIndex := range entry.API {
			entry.API[sectionIndex].Props = slices.Clone(entry.API[sectionIndex].Props)
		}
		cloned[key] = entry
	}
	return cloned
}

func declarationInventoryError(want []string, got []string) error {
	if slices.Equal(want, got) {
		return nil
	}

	wantSet := make(map[string]struct{}, len(want))
	for _, key := range want {
		wantSet[key] = struct{}{}
	}
	gotSet := make(map[string]struct{}, len(got))
	for _, key := range got {
		gotSet[key] = struct{}{}
	}

	var missing []string
	for _, key := range want {
		if _, ok := gotSet[key]; !ok {
			missing = append(missing, key)
		}
	}
	var unexpected []string
	for _, key := range got {
		if _, ok := wantSet[key]; !ok {
			unexpected = append(unexpected, key)
		}
	}
	return fmt.Errorf(
		"canonical declaration inventory mismatch: want %d, got %d, missing %v, unexpected %v",
		len(want),
		len(got),
		missing,
		unexpected,
	)
}

const canonicalDeclarationInventory = `
constructor|components/accordion|accordionconfig
constructor|components/alert|config
constructor|components/avatar|config
constructor|components/avatar|stackconfig
constructor|components/badge|config
constructor|components/breadcrumbs|config
constructor|components/button|button-options
constructor|components/card|config
constructor|components/carousel|cardconfig
constructor|components/carousel|config
constructor|components/chatbubble|config
constructor|components/checkbox|config
constructor|components/checkbox|groupconfig
constructor|components/codeblock|config
constructor|components/combobox|config
constructor|components/drawer|config
constructor|components/dropdown|config
constructor|components/fileinput|config
constructor|components/form|collapsiblesectionconfig
constructor|components/form|config
constructor|components/form|fieldgroupconfig
constructor|components/form|flipsectionconfig
constructor|components/form|formerrorsconfig
constructor|components/form|sectionconfig
constructor|components/form|subsectionconfig
constructor|components/kbd|kbd-options
constructor|components/link|link-options
constructor|components/modal|alertdialogconfig
constructor|components/modal|config
constructor|components/navbar|config
constructor|components/pagination|config
constructor|components/palette|config
constructor|components/radio|config
constructor|components/radio|groupconfig
constructor|components/range|config
constructor|components/rating|config
constructor|components/rating|displayconfig
constructor|components/schema-form|fieldsconfig
constructor|components/search|config
constructor|components/search|searchfield
constructor|components/search|searchmodal
constructor|components/select|config
constructor|components/sidebar|config
constructor|components/sidebar|overlayconfig
constructor|components/spinner|config
constructor|components/steps|config
constructor|components/structured-input|config
constructor|components/table|config
constructor|components/tabs|config
constructor|components/tags-list|config
constructor|components/text-input|config
constructor|components/textarea|config
constructor|components/toast|config
constructor|components/toast|containerconfig
constructor|components/toast|messageconfig
constructor|components/toast|oobmessagetoast
constructor|components/toast|oobtoast
constructor|components/toggle|config
constructor|components/tooltip|tooltip-options
declaration|components/avatar|helpers|GetInitials
declaration|components/badge|animatingdot|AnimatingDot
declaration|components/badge|notificationbadge|NotificationBadge
declaration|components/badge|notificationdot|NotificationDot
declaration|components/banner|constructors|Banner
declaration|components/banner|constructors|CookieBanner
declaration|components/button|button-options|Disabled
declaration|components/button|button-options|WithAlpine
declaration|components/button|button-options|WithHTMX
declaration|components/button|button-options|WithID
declaration|components/button|button-options|WithLoadingText
declaration|components/button|button-options|WithRootClass
declaration|components/button|button-options|WithSize
declaration|components/button|button-options|WithTone
declaration|components/button|button-options|WithType
declaration|components/chatbubble|typingindicator|TypingIndicator
declaration|components/codeblock|helpers|Render
declaration|components/combobox|combobox-server-contracts|Config.InitialState
declaration|components/combobox|combobox-server-contracts|Config.Validate
declaration|components/combobox|combobox-server-contracts|Handler
declaration|components/combobox|combobox-server-contracts|OptionsProvider
declaration|components/dependencies|dependenciesminimal|DependenciesMinimal
declaration|components/dependencies|dependencies|Dependencies
declaration|components/form|validation-operations|FormDef.Bind
declaration|components/form|validation-operations|FormDef.Dependents
declaration|components/form|validation-operations|FormDef.PopulateValues
declaration|components/form|validation-operations|Handle
declaration|components/form|validation-operations|IsFieldValidation
declaration|components/form|validation-operations|RenderFieldResponse
declaration|components/form|validation-operations|ValidateFunc
declaration|components/form|validation-operations|ValidationType
declaration|components/kbd|kbd-options|WithAttrs
declaration|components/kbd|kbd-options|WithIcon
declaration|components/kbd|kbd-options|WithLabel
declaration|components/kbd|kbd-options|WithRootClass
declaration|components/kbd|kbd-options|WithSize
declaration|components/link|link-options|WithAppearance
declaration|components/link|link-options|WithAttrs
declaration|components/link|link-options|WithID
declaration|components/link|link-options|WithIcon
declaration|components/link|link-options|WithIconPosition
declaration|components/link|link-options|WithRel
declaration|components/link|link-options|WithRole
declaration|components/link|link-options|WithRootClass
declaration|components/link|link-options|WithSize
declaration|components/link|link-options|WithTarget
declaration|components/pagination|config-helpers|HasNext
declaration|components/pagination|config-helpers|HasPrevious
declaration|components/pagination|config-helpers|NextPage
declaration|components/pagination|config-helpers|PageURL
declaration|components/pagination|config-helpers|Pages
declaration|components/pagination|config-helpers|PreviousPage
declaration|components/radio|radiobar|RadioBar
declaration|components/schema-form|allowmode-and-transforms|AllowMode
declaration|components/schema-form|allowmode-and-transforms|AllowModeDisabled
declaration|components/schema-form|allowmode-and-transforms|AllowModeManaged
declaration|components/schema-form|allowmode-and-transforms|FallbackFromDefaults
declaration|components/schema-form|allowmode-and-transforms|FlattenAllowList
declaration|components/schema-form|allowmode-and-transforms|HasOnlySimpleScalars
declaration|components/schema-form|allowmode-and-transforms|PruneDisabled
declaration|components/schema-form|allowmode-and-transforms|Walk
declaration|components/search|item-methods|Item.NormalizedMethod
declaration|components/search|item-methods|Item.SafeHref
declaration|components/search|item-methods|Item.SearchText
declaration|components/table|config-helpers|IsSortedBy
declaration|components/table|config-helpers|NextPageURL
declaration|components/table|config-helpers|NextSortDir
declaration|components/table|config-helpers|PageURL
declaration|components/table|config-helpers|SortURL
declaration|components/table|imagecell|ImageCell
declaration|components/table|tableheadcontent|TableHeadContent
declaration|components/table|tablepaginationnav|TablePaginationNav
declaration|components/table|tablerows|TableRows
declaration|components/table|tablerow|TableRow
declaration|components/textarea|textareawithactions|TextareaWithActions
declaration|components/tooltip|tooltip-options|WithActivation
declaration|components/tooltip|tooltip-options|WithDescription
declaration|components/tooltip|tooltip-options|WithPosition
declaration|components/tooltip|tooltip-options|WithTrigger
declaration|components/tooltip|tooltip-options|WithTriggerLabel
`

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

func isDeclarationSignature(signature string) bool {
	return strings.HasPrefix(signature, "func ") ||
		strings.HasPrefix(signature, "type ") ||
		strings.HasPrefix(signature, "const ")
}

func documentedDeclarationError(
	index map[string]packageDeclarations,
	sourcePackage string,
	signature string,
) error {
	normalized, err := normalizeDeclaration(signature)
	if err != nil {
		return err
	}
	return exactDeclarationMatchError(index, sourcePackage, "", normalized)
}

func documentedConstructorError(
	index map[string]packageDeclarations,
	sourcePackage string,
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
	return exactDeclarationMatchError(index, sourcePackage, packageName, normalized)
}

func exactDeclarationMatchError(
	index map[string]packageDeclarations,
	sourcePackage string,
	requiredPackageName string,
	normalized string,
) error {
	pkg, ok := index[sourcePackage]
	if !ok {
		return fmt.Errorf("source package %q is absent from the component declaration index", sourcePackage)
	}
	if requiredPackageName != "" && pkg.name != requiredPackageName {
		return fmt.Errorf(
			"constructor package qualifier %q does not match package %q declared by %s",
			requiredPackageName,
			pkg.name,
			sourcePackage,
		)
	}
	if _, ok := pkg.declarations[normalized]; !ok {
		return fmt.Errorf(
			"%q does not match an exported declaration in exact source package %s",
			normalized,
			sourcePackage,
		)
	}
	return nil
}

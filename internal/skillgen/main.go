// Command skillgen derives the Goshtoso component API reference from source.
//
// It parses every package under components/ via go/ast and emits the
// per-component reference consumed by the using-goshtoso skill. Because the
// output is derived from each component package's Go source, it can never drift
// from or misrepresent the real API.
//
// Run from the repo root:
//
//	go run ./cmd/skillgen
//
// CI and the pre-commit hook run it and fail if the committed file is stale
// (see .github/workflows/ci.yml and .githooks/pre-commit).
package skillgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	componentsDir        = "components"
	legacyOutPath        = ".claude/skills/using-goshtoso/components-reference.md"
	externalSkillOutPath = ".agents/skills/using-goshtoso/references/components-reference.md"
	modulePath           = "github.com/araihu/goshtoso"
	templImportPath      = "github.com/a-h/templ"
	componentsImportPath = modulePath + "/components"
)

// pkgAPI is the extracted public surface of one component package.
type pkgAPI struct {
	dir     string // directory name under components/ (= import path tail)
	name    string // Go package name (may differ, e.g. select → selectfield)
	entries []string
	options []string
	enums   []enum
	structs []strukt
}

type enum struct {
	typ    string
	values []string // "Name = value"
}

type strukt struct {
	name   string
	fields []field
}

type field struct {
	name string
	typ  string
	doc  string
}

// Run regenerates the component API reference. It must be called from the repo root.
func Run() error {
	dirs, err := componentDirs()
	if err != nil {
		return err
	}
	var pkgs []pkgAPI
	for _, rel := range dirs {
		api, ok, err := parsePkg(filepath.Join(componentsDir, rel), rel)
		if err != nil {
			return err
		}
		if ok {
			pkgs = append(pkgs, api)
		}
	}
	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].dir < pkgs[j].dir })
	output := []byte(render(pkgs))
	for _, path := range []string{legacyOutPath, externalSkillOutPath} {
		if err := writeFile(path, output); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(path string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// componentDirs returns every directory under components/ (recursively, so
// nested component packages are included) that directly contains Go
// source, as paths relative to componentsDir.
func componentDirs() ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(componentsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		matches, gerr := filepath.Glob(filepath.Join(path, "*.go"))
		if gerr != nil {
			return gerr
		}
		rel, rerr := filepath.Rel(componentsDir, path)
		if rerr != nil {
			return rerr
		}
		if len(matches) > 0 && rel != "." {
			dirs = append(dirs, rel)
		}
		return nil
	})
	sort.Strings(dirs)
	return dirs, err
}

// parsePkg extracts the public API of one component package directory.
func parsePkg(dir, dirName string) (pkgAPI, bool, error) {
	fset := token.NewFileSet()
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return pkgAPI{}, false, err
	}
	api := pkgAPI{dir: dirName}
	enums := map[string][]string{}
	parsedFiles := make([]*ast.File, 0, len(files))
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, parser.ParseComments)
		if err != nil {
			return pkgAPI{}, false, err
		}
		api.name = af.Name.Name
		parsedFiles = append(parsedFiles, af)
	}
	errorShadowed := packageDeclares(parsedFiles, "error")
	concreteTypes := collectComponentTypes(parsedFiles, errorShadowed)
	for _, af := range parsedFiles {
		collectFile(fset, af, &api, enums, concreteTypes, importBindings(af))
	}
	if api.name == "" {
		return pkgAPI{}, false, nil
	}
	for t, v := range enums {
		api.enums = append(api.enums, enum{typ: t, values: v})
	}
	sort.Slice(api.enums, func(i, j int) bool { return api.enums[i].typ < api.enums[j].typ })
	sort.Strings(api.entries)
	sort.Strings(api.options)
	sort.Slice(api.structs, func(i, j int) bool { return api.structs[i].name < api.structs[j].name })
	return api, true, nil
}

type componentMethod uint8

const (
	componentKindMethod componentMethod = 1 << iota
	componentRenderMethod
	componentMethods = componentKindMethod | componentRenderMethod
)

type componentTypes struct {
	value   map[string]struct{}
	pointer map[string]struct{}
}

// collectComponentTypes returns package-local types whose value or pointer
// method sets expose the syntactic components.Component contract.
func collectComponentTypes(files []*ast.File, errorShadowed bool) componentTypes {
	valueMethods := map[string]componentMethod{}
	pointerMethods := map[string]componentMethod{}
	for _, af := range files {
		imports := importBindings(af)
		for _, decl := range af.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 {
				continue
			}
			name, pointer := receiverType(fn.Recv.List[0].Type)
			method := componentMethodOf(fn, imports, errorShadowed)
			if name == "" || method == 0 {
				continue
			}
			if pointer {
				pointerMethods[name] |= method
				continue
			}
			valueMethods[name] |= method
			pointerMethods[name] |= method
		}
	}

	types := componentTypes{
		value:   map[string]struct{}{},
		pointer: map[string]struct{}{},
	}
	for name, methods := range valueMethods {
		if methods == componentMethods {
			types.value[name] = struct{}{}
		}
	}
	for name, methods := range pointerMethods {
		if methods == componentMethods {
			types.pointer[name] = struct{}{}
		}
	}
	return types
}

func componentMethodOf(
	fn *ast.FuncDecl,
	imports map[string]string,
	errorShadowed bool,
) componentMethod {
	switch fn.Name.Name {
	case "Kind":
		result, single := singleResult(fn.Type.Results)
		if len(parameterTypes(fn.Type.Params)) == 0 &&
			single &&
			isImportedSelector(result, imports, componentsImportPath, "Kind") {
			return componentKindMethod
		}
	case "Render":
		params := parameterTypes(fn.Type.Params)
		result, single := singleResult(fn.Type.Results)
		if len(params) == 2 &&
			isImportedSelector(params[0], imports, "context", "Context") &&
			isImportedSelector(params[1], imports, "io", "Writer") &&
			single &&
			isPredeclaredError(result, imports, errorShadowed) {
			return componentRenderMethod
		}
	}
	return 0
}

func parameterTypes(fields *ast.FieldList) []ast.Expr {
	if fields == nil {
		return nil
	}
	var types []ast.Expr
	for _, field := range fields.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			types = append(types, field.Type)
		}
	}
	return types
}

func singleResult(results *ast.FieldList) (ast.Expr, bool) {
	if results == nil {
		return nil, false
	}
	var result ast.Expr
	count := 0
	for _, field := range results.List {
		fieldCount := len(field.Names)
		if fieldCount == 0 {
			fieldCount = 1
		}
		count += fieldCount
		if count > 1 {
			return nil, false
		}
		result = field.Type
	}
	return result, count == 1
}

func isImportedSelector(
	expr ast.Expr,
	imports map[string]string,
	importPath string,
	name string,
) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != name {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && imports[pkg.Name] == importPath
}

func isPredeclaredError(
	expr ast.Expr,
	imports map[string]string,
	packageShadowed bool,
) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok || ident.Name != "error" || packageShadowed {
		return false
	}
	_, fileShadowed := imports["error"]
	return !fileShadowed
}

func importBindings(af *ast.File) map[string]string {
	imports := make(map[string]string, len(af.Imports))
	for _, spec := range af.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		name := defaultImportName(importPath)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		if name == "." || name == "_" {
			continue
		}
		imports[name] = importPath
	}
	return imports
}

func defaultImportName(importPath string) string {
	if slash := strings.LastIndexByte(importPath, '/'); slash >= 0 {
		return importPath[slash+1:]
	}
	return importPath
}

func packageDeclares(files []*ast.File, name string) bool {
	for _, af := range files {
		for _, decl := range af.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil && d.Name.Name == name {
					return true
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.Name == name {
							return true
						}
					case *ast.ValueSpec:
						for _, ident := range s.Names {
							if ident.Name == name {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func receiverType(expr ast.Expr) (name string, pointer bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, false
	case *ast.StarExpr:
		ident, ok := t.X.(*ast.Ident)
		if ok {
			return ident.Name, true
		}
	}
	return "", false
}

// collectFile walks one file's top-level decls into the package API.
func collectFile(
	fset *token.FileSet,
	af *ast.File,
	api *pkgAPI,
	enums map[string][]string,
	concreteTypes componentTypes,
	imports map[string]string,
) {
	for _, decl := range af.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if isEntry(d, concreteTypes, imports) {
				api.entries = append(api.entries, entrySig(fset, d))
			}
			if isOption(fset, d) {
				api.options = append(api.options, entrySig(fset, d))
			}
		case *ast.GenDecl:
			if d.Tok == token.CONST {
				collectConsts(fset, d, enums)
			}
			if d.Tok == token.TYPE {
				collectStructs(fset, d, api)
			}
		}
	}
}

// isEntry reports whether fn is an exported component entry point: a top-level
// function with no receiver returning templ.Component or a package-local type
// that exposes both Kind and Render methods.
func isEntry(
	fn *ast.FuncDecl,
	concreteTypes componentTypes,
	imports map[string]string,
) bool {
	if fn.Recv != nil || !fn.Name.IsExported() || fn.Type.Results == nil {
		return false
	}
	result, single := singleResult(fn.Type.Results)
	if !single {
		return false
	}
	if isImportedSelector(result, imports, templImportPath, "Component") {
		return true
	}
	name, pointer := receiverType(result)
	if name == "" {
		return false
	}
	types := concreteTypes.value
	if pointer {
		types = concreteTypes.pointer
	}
	_, ok := types[name]
	return ok
}

// isOption reports whether fn is an exported functional option constructor:
// a top-level function whose single result is the package-local Option type.
func isOption(fset *token.FileSet, fn *ast.FuncDecl) bool {
	if fn.Recv != nil || !fn.Name.IsExported() || fn.Type.Results == nil {
		return false
	}
	results := fn.Type.Results.List
	return len(results) == 1 &&
		len(results[0].Names) <= 1 &&
		exprString(fset, results[0].Type) == "Option"
}

func entrySig(fset *token.FileSet, fn *ast.FuncDecl) string {
	var params []string
	if fn.Type.Params != nil {
		for _, p := range fn.Type.Params.List {
			t := exprString(fset, p.Type)
			if len(p.Names) == 0 {
				params = append(params, t)
				continue
			}
			for _, n := range p.Names {
				params = append(params, n.Name+" "+t)
			}
		}
	}
	return fn.Name.Name + "(" + strings.Join(params, ", ") + ")"
}

// collectConsts groups exported typed constants by their type name (enums).
func collectConsts(fset *token.FileSet, d *ast.GenDecl, enums map[string][]string) {
	var lastType string
	for _, spec := range d.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		if vs.Type != nil {
			lastType = exprString(fset, vs.Type)
		}
		if lastType == "" {
			continue
		}
		for i, n := range vs.Names {
			if !n.IsExported() {
				continue
			}
			val := ""
			if i < len(vs.Values) {
				val = " = " + exprString(fset, vs.Values[i])
			}
			enums[lastType] = append(enums[lastType], n.Name+val)
		}
	}
}

// collectStructs records exported struct types and their exported fields.
func collectStructs(fset *token.FileSet, d *ast.GenDecl, api *pkgAPI) {
	for _, spec := range d.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok || !ts.Name.IsExported() {
			continue
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			continue
		}
		// *Data structs are internal per-item render payloads, not public config.
		if strings.HasSuffix(ts.Name.Name, "Data") {
			continue
		}
		s := strukt{name: ts.Name.Name}
		for _, f := range st.Fields.List {
			for _, n := range f.Names {
				if !n.IsExported() {
					continue
				}
				s.fields = append(s.fields, field{
					name: n.Name,
					typ:  exprString(fset, f.Type),
					doc:  markdownTableText(docOf(f)),
				})
			}
		}
		if len(s.fields) > 0 {
			api.structs = append(api.structs, s)
		}
	}
}

func docOf(f *ast.Field) string {
	if f.Doc != nil {
		return f.Doc.Text()
	}
	if f.Comment != nil {
		return f.Comment.Text()
	}
	return ""
}

func markdownTableText(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	return strings.NewReplacer(
		"|", `\|`,
		"<", "&lt;",
		">", "&gt;",
	).Replace(s)
}

func exprString(fset *token.FileSet, e ast.Expr) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, e)
	return buf.String()
}

// render builds the markdown reference document.
func render(pkgs []pkgAPI) string {
	var b strings.Builder
	b.WriteString("<!-- GENERATED by cmd/skillgen — DO NOT EDIT. Run `go run ./cmd/skillgen`. -->\n")
	b.WriteString("# Goshtoso component reference\n\n")
	b.WriteString("## Component API\n\n")
	b.WriteString("Public constructors return concrete values that implement `components.Component`\n")
	b.WriteString("and `templ.Component`. Use the concrete return type for component-specific code,\n")
	b.WriteString("or store mixed values through the common interface and inspect their stable\n")
	b.WriteString("`Kind()`. Constructor signatures, config fields, options, and rendered defaults are\n")
	b.WriteString("listed below. See the selected Goshtoso tag's `docs/COMPONENT_MODEL.md`.\n\n")
	fmt.Fprintf(&b, "%d component packages. Each is imported by its directory path; note the\n", len(pkgs))
	b.WriteString("**package name** when it differs from the directory (e.g. `select` → `selectfield`).\n\n")
	for _, p := range pkgs {
		renderPkg(&b, p)
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderPkg(b *strings.Builder, p pkgAPI) {
	fmt.Fprintf(b, "## %s\n\n", p.dir)
	fmt.Fprintf(b, "```go\nimport \"%s/%s/%s\"  // package %s\n```\n\n", modulePath, componentsDir, p.dir, p.name)
	if len(p.entries) > 0 {
		b.WriteString("**Entry points:** ")
		for i, e := range p.entries {
			if i > 0 {
				b.WriteString(" · ")
			}
			fmt.Fprintf(b, "`%s`", e)
		}
		b.WriteString("\n\n")
	}
	if len(p.options) > 0 {
		b.WriteString("**Options:** ")
		for i, option := range p.options {
			if i > 0 {
				b.WriteString(" · ")
			}
			fmt.Fprintf(b, "`%s`", option)
		}
		b.WriteString("\n\n")
	}
	for _, en := range p.enums {
		fmt.Fprintf(b, "- **%s** — %s\n", en.typ, strings.Join(en.values, ", "))
	}
	if len(p.enums) > 0 {
		b.WriteString("\n")
	}
	for _, s := range p.structs {
		renderStruct(b, s)
	}
}

func renderStruct(b *strings.Builder, s strukt) {
	fmt.Fprintf(b, "**%s**\n\n", s.name)
	b.WriteString("| Field | Type | Description |\n|-------|------|-------------|\n")
	for _, f := range s.fields {
		fmt.Fprintf(b, "| `%s` | `%s` | %s |\n", f.name, f.typ, f.doc)
	}
	b.WriteString("\n")
}

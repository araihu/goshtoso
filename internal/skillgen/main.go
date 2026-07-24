// Command skillgen derives the Goshtoso component API reference from source.
//
// It parses every package under components/ via go/ast and emits the
// per-component reference consumed by the using-goshtoso skill. Because the
// output is derived from types.go (and the generated *_templ.go entry points),
// it can never drift from or misrepresent the real API.
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
	"strings"
)

const (
	componentsDir        = "components"
	legacyOutPath        = ".claude/skills/using-goshtoso/components-reference.md"
	externalSkillOutPath = ".agents/skills/using-goshtoso/references/components-reference.md"
	modulePath           = "github.com/araihu/goshtoso"
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
	for _, f := range files {
		af, err := parser.ParseFile(fset, f, nil, parser.ParseComments)
		if err != nil {
			return pkgAPI{}, false, err
		}
		api.name = af.Name.Name
		collectFile(fset, af, &api, enums)
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

// collectFile walks one file's top-level decls into the package API.
func collectFile(fset *token.FileSet, af *ast.File, api *pkgAPI, enums map[string][]string) {
	for _, decl := range af.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if isEntry(fset, d) {
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
// function with no receiver returning templ.Component.
func isEntry(fset *token.FileSet, fn *ast.FuncDecl) bool {
	if fn.Recv != nil || !fn.Name.IsExported() || fn.Type.Results == nil {
		return false
	}
	for _, r := range fn.Type.Results.List {
		if exprString(fset, r.Type) == "templ.Component" {
			return true
		}
	}
	return false
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
					doc:  firstLine(docOf(f)),
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

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.ReplaceAll(s, "|", `\|`)
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

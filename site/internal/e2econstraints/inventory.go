// Package e2econstraints validates build-constraint safety for the E2E suite.
package e2econstraints

import (
	"fmt"
	"go/types"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

// CrossFileDeclaration is a package declaration whose defining source file is
// not guaranteed to be present with all of its consumers.
type CrossFileDeclaration struct {
	Name          string
	DeclaringFile string
	ConsumerFiles []string
}

// FindCrossFileDeclarations type-checks the E2E package and reports declarations
// that are consumed outside their identity-owned file.
func FindCrossFileDeclarations(siteDir string) ([]CrossFileDeclaration, error) {
	config := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		Dir:   siteDir,
		Tests: true,
		BuildFlags: []string{
			"-tags=e2e,full",
		},
	}
	loaded, err := packages.Load(config, "./tests/e2e")
	if err != nil {
		return nil, fmt.Errorf("load E2E package: %w", err)
	}
	if packages.PrintErrors(loaded) > 0 {
		return nil, fmt.Errorf("load E2E package: package contains errors")
	}

	pkg := testVariant(loaded)
	if pkg == nil {
		return nil, fmt.Errorf("load E2E package: test variant not found")
	}

	definitionFiles := make(map[types.Object]string)
	for ident, object := range pkg.TypesInfo.Defs {
		if object == nil || ident == nil || object.Pkg() != pkg.Types {
			continue
		}
		position := pkg.Fset.Position(ident.Pos())
		if position.Filename == "" || isSupportFile(position.Filename) || isTestEntryPoint(object.Name()) {
			continue
		}
		definitionFiles[object] = filepath.Base(position.Filename)
	}

	consumers := make(map[types.Object]map[string]struct{})
	for ident, object := range pkg.TypesInfo.Uses {
		declaringFile, tracked := definitionFiles[object]
		if !tracked {
			continue
		}
		consumerFile := filepath.Base(pkg.Fset.Position(ident.Pos()).Filename)
		if consumerFile == "" || consumerFile == declaringFile {
			continue
		}
		if consumers[object] == nil {
			consumers[object] = make(map[string]struct{})
		}
		consumers[object][consumerFile] = struct{}{}
	}

	findings := make([]CrossFileDeclaration, 0, len(consumers))
	for object, files := range consumers {
		consumerFiles := make([]string, 0, len(files))
		for file := range files {
			consumerFiles = append(consumerFiles, file)
		}
		slices.Sort(consumerFiles)
		findings = append(findings, CrossFileDeclaration{
			Name:          object.Name(),
			DeclaringFile: definitionFiles[object],
			ConsumerFiles: consumerFiles,
		})
	}
	slices.SortFunc(findings, func(a, b CrossFileDeclaration) int {
		if byFile := strings.Compare(a.DeclaringFile, b.DeclaringFile); byFile != 0 {
			return byFile
		}
		return strings.Compare(a.Name, b.Name)
	})
	return findings, nil
}

func testVariant(loaded []*packages.Package) *packages.Package {
	var best *packages.Package
	for _, pkg := range loaded {
		if pkg.PkgPath != "github.com/araihu/goshtoso/site/tests/e2e" {
			continue
		}
		if best == nil || len(pkg.Syntax) > len(best.Syntax) {
			best = pkg
		}
	}
	return best
}

func isSupportFile(filename string) bool {
	base := filepath.Base(filename)
	return base == "e2e_test.go" ||
		base == "class_verifier.go" ||
		strings.HasSuffix(base, "_support_test.go")
}

func isTestEntryPoint(name string) bool {
	return name == "TestMain" || strings.HasPrefix(name, "Test") ||
		strings.HasPrefix(name, "Benchmark") || strings.HasPrefix(name, "Fuzz")
}

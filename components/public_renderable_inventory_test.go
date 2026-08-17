package components_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components"
	"github.com/stretchr/testify/require"
)

func TestPublicRenderableInventoryMatchesAllKinds(t *testing.T) {
	got := make([]components.Kind, 0, 85)
	for _, inventory := range publicRenderableInventories() {
		for want, value := range inventory {
			require.Equal(t, want, value.Kind())
			got = append(got, value.Kind())
		}
	}

	require.ElementsMatch(t, components.AllKinds(), got)
	require.Len(t, got, 85)
	require.Len(t, got, len(components.AllKinds()))
}

func TestPublicFunctionSurfaceMatchesContract(t *testing.T) {
	allowedFunctions := map[string]struct{}{
		"components.AllKinds": {},

		"accordion.Accordion":             {},
		"actiongroup.ActionGroup":         {},
		"alert.Alert":                     {},
		"appshell.AppShell":               {},
		"avatar.Avatar":                   {},
		"avatar.AvatarStack":              {},
		"badge.Badge":                     {},
		"badge.NotificationBadge":         {},
		"badge.NotificationDot":           {},
		"badge.AnimatingDot":              {},
		"banner.Banner":                   {},
		"banner.CookieBanner":             {},
		"breadcrumbs.Breadcrumbs":         {},
		"button.Button":                   {},
		"card.Card":                       {},
		"carousel.Carousel":               {},
		"carousel.CardCarousel":           {},
		"chatbubble.ChatBubble":           {},
		"chatbubble.TypingIndicator":      {},
		"checkbox.Checkbox":               {},
		"checkbox.CheckboxGroup":          {},
		"codeblock.CodeBlock":             {},
		"inlinecode.InlineCode":           {},
		"inlinecode.WithRootAttrs":        {},
		"inlinecode.WithRootClass":        {},
		"combobox.Combobox":               {},
		"drawer.Drawer":                   {},
		"dropdown.Dropdown":               {},
		"popover.Popover":                 {},
		"splitbutton.SplitButton":         {},
		"emptystate.EmptyState":           {},
		"fileinput.FileInput":             {},
		"form.Form":                       {},
		"form.Section":                    {},
		"form.CollapsibleSection":         {},
		"form.FlipSection":                {},
		"form.SubSection":                 {},
		"form.FieldGroup":                 {},
		"form.FormErrors":                 {},
		"head.Dependencies":               {},
		"head.DependenciesMinimal":        {},
		"head.Metadata":                   {},
		"head.WithComboboxURL":            {},
		"head.WithActionGroupURL":         {},
		"head.WithDependencyCDNURL":       {},
		"head.WithDependencyIntegrity":    {},
		"head.WithDependencyLocalURL":     {},
		"head.WithLoaderURL":              {},
		"head.WithLocalRuntime":           {},
		"head.WithRuntimeManifest":        {},
		"head.WithStylesheetOnly":         {},
		"head.WithStylesheetURL":          {},
		"head.WithoutDependency":          {},
		"head.WithoutLocalFallback":       {},
		"icon.Icon":                       {},
		"kbd.Kbd":                         {},
		"link.Link":                       {},
		"modal.Modal":                     {},
		"modal.AlertDialog":               {},
		"navbar.Navbar":                   {},
		"navbar.SecondaryRow":             {},
		"pagination.Pagination":           {},
		"panel.Panel":                     {},
		"pageheader.PageHeader":           {},
		"palette.Palette":                 {},
		"radio.Radio":                     {},
		"radio.RadioBar":                  {},
		"radio.RadioGroup":                {},
		"range.Range":                     {},
		"rating.Rating":                   {},
		"rating.RatingDisplay":            {},
		"schemaform.Fields":               {},
		"scrollregion.ScrollRegion":       {},
		"search.Search":                   {},
		"search.SearchField":              {},
		"search.SearchModal":              {},
		"select.Select":                   {},
		"sidebar.Sidebar":                 {},
		"sidebar.Overlay":                 {},
		"skeleton.Skeleton":               {},
		"spinner.Spinner":                 {},
		"steps.Steps":                     {},
		"structuredinput.StructuredInput": {},
		"table.Table":                     {},
		"table.TableHeadContent":          {},
		"table.TableRows":                 {},
		"table.TableRow":                  {},
		"table.TablePaginationNav":        {},
		"table.ImageCell":                 {},
		"tabs.Tabs":                       {},
		"tagslist.TagsList":               {},
		"textarea.Textarea":               {},
		"textarea.TextareaWithActions":    {},
		"textinput.TextInput":             {},
		"toast.ToastContainer":            {},
		"toast.Toast":                     {},
		"toast.MessageToast":              {},
		"toast.OOBToast":                  {},
		"toast.OOBMessageToast":           {},
		"toggle.Toggle":                   {},
		"toolbar.Toolbar":                 {},
		"tooltip.Tooltip":                 {},

		"button.WithTone":        {},
		"button.WithSize":        {},
		"button.WithType":        {},
		"button.Disabled":        {},
		"button.WithID":          {},
		"button.WithRootClass":   {},
		"button.WithAttrs":       {},
		"button.WithHTMX":        {},
		"button.WithAlpine":      {},
		"button.WithLoadingText": {},

		"link.WithTarget":       {},
		"link.WithRel":          {},
		"link.WithRole":         {},
		"link.WithID":           {},
		"link.WithAppearance":   {},
		"link.WithSize":         {},
		"link.WithIcon":         {},
		"link.WithIconPosition": {},
		"link.WithRootClass":    {},
		"link.WithAttrs":        {},

		"kbd.WithLabel":     {},
		"kbd.WithSize":      {},
		"kbd.WithIcon":      {},
		"kbd.WithRootClass": {},
		"kbd.WithAttrs":     {},

		"tooltip.WithDescription":  {},
		"tooltip.WithPosition":     {},
		"tooltip.WithActivation":   {},
		"tooltip.WithTriggerLabel": {},
		"tooltip.WithTrigger":      {},

		"avatar.GetInitials":                  {},
		"codeblock.Render":                    {},
		"combobox.Handler":                    {},
		"schemaform.FlattenAllowList":         {},
		"schemaform.Walk":                     {},
		"schemaform.FallbackFromDefaults":     {},
		"schemaform.PruneDisabled":            {},
		"schemaform.HasOnlySimpleScalars":     {},
		"form/validation.Handle":              {},
		"form/validation.IsFieldValidation":   {},
		"form/validation.RenderFieldResponse": {},
	}

	allowedMethods := map[string]struct{}{
		"head.metadataComponent.Render":          {},
		"combobox.Config.Validate":               {},
		"combobox.Config.InitialState":           {},
		"combobox.comboHandler.ServeHTTP":        {},
		"form.FieldGroupConfig.FocusTargetID":    {},
		"pagination.Config.HasPrevious":          {},
		"pagination.Config.HasNext":              {},
		"pagination.Config.PreviousPage":         {},
		"pagination.Config.NextPage":             {},
		"pagination.Config.PageURL":              {},
		"pagination.Config.Pages":                {},
		"search.Item.SearchText":                 {},
		"search.Item.NormalizedMethod":           {},
		"search.Item.SafeHref":                   {},
		"navbar.Config.Validate":                 {},
		"navbar.SecondaryConfig.Validate":        {},
		"navbar.ValidationError.Error":           {},
		"table.Config.IsSortedBy":                {},
		"table.Config.NextSortDir":               {},
		"table.Config.SortURL":                   {},
		"table.Config.PageURL":                   {},
		"table.Config.NextPageURL":               {},
		"table.Config.TbodyID":                   {},
		"table.Config.TheadID":                   {},
		"table.Config.PaginationID":              {},
		"form/validation.FormDef.Bind":           {},
		"form/validation.FormDef.Dependents":     {},
		"form/validation.FormDef.PopulateValues": {},
	}

	gotFunctions, gotMethods := exportedGoFunctions(t)
	allowedComponentMethods := allowedRenderableMethods(t)

	unexpected := unexpectedPublicSurface(
		gotFunctions,
		gotMethods,
		allowedFunctions,
		allowedMethods,
		allowedComponentMethods,
	)
	require.Empty(t, unexpected, "unexpected exported component functions or methods")

	var missing []string
	for name := range allowedFunctions {
		if _, ok := gotFunctions[name]; !ok {
			missing = append(missing, name)
		}
	}
	for name := range allowedMethods {
		if _, ok := gotMethods[name]; !ok {
			missing = append(missing, name)
		}
	}
	for name := range allowedComponentMethods {
		if _, ok := gotMethods[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	require.Empty(t, missing, "contracted public component functions or methods disappeared")
}

func TestExportedGoFunctionsIncludesGeneratedTemplDeclarations(t *testing.T) {
	root := t.TempDir()
	writeGoFixture(t, root, "avatar/avatar_templ.go", `package avatar

func UserIcon() {}
`)

	functions, _, err := exportedGoFunctionsFromDir(root)
	require.NoError(t, err)
	require.Contains(t, functions, "avatar.UserIcon")
}

func TestExportedGoFunctionsIncludesPlatformSpecificDeclarations(t *testing.T) {
	root := t.TempDir()
	writeGoFixture(t, root, "avatar/helper_windows.go", `//go:build windows

package avatar

func WindowsOnlyHelper() {}
`)

	functions, _, err := exportedGoFunctionsFromDir(root)
	require.NoError(t, err)
	require.Contains(t, functions, "avatar.WindowsOnlyHelper")
}

func TestPublicSurfaceRejectsKindAndRenderOnNonInventoryReceiver(t *testing.T) {
	root := t.TempDir()
	writeGoFixture(t, root, "fake/component.go", `package fake

type Helper struct{}

func (Helper) Kind() {}
func (Helper) Render() {}
`)

	_, methods, err := exportedGoFunctionsFromDir(root)
	require.NoError(t, err)

	unexpected := unexpectedPublicSurface(
		nil,
		methods,
		nil,
		nil,
		map[string]struct{}{},
	)

	require.ElementsMatch(t, []string{"fake.Helper.Kind", "fake.Helper.Render"}, unexpected)
}

func TestExportedGoFunctionsHandlesGenericReceivers(t *testing.T) {
	root := t.TempDir()
	writeGoFixture(t, root, "generic/helpers.go", `package generic

type Box[T any] struct{}
type Pair[A, B any] struct{}

func (Box[T]) Kind() {}
func (*Pair[A, B]) Render() {}
`)

	_, methods, err := exportedGoFunctionsFromDir(root)
	require.NoError(t, err)
	require.Contains(t, methods, "generic.Box.Kind")
	require.Contains(t, methods, "generic.Pair.Render")
}

func TestExportedGoFunctionsRejectsUnrecognizedReceiver(t *testing.T) {
	root := t.TempDir()
	writeGoFixture(t, root, "invalid/helper.go", `package invalid

import "time"

func (time.Time) HiddenExport() {}
`)

	_, _, err := exportedGoFunctionsFromDir(root)
	require.Error(t, err)
}

func unexpectedPublicSurface(
	gotFunctions map[string]string,
	gotMethods map[string]string,
	allowedFunctions map[string]struct{},
	allowedMethods map[string]struct{},
	allowedComponentMethods map[string]struct{},
) []string {
	var unexpected []string
	for name := range gotFunctions {
		if _, ok := allowedFunctions[name]; !ok {
			unexpected = append(unexpected, name)
		}
	}
	for name := range gotMethods {
		if _, ok := allowedMethods[name]; ok {
			continue
		}
		if _, ok := allowedComponentMethods[name]; ok {
			continue
		}
		unexpected = append(unexpected, name)
	}
	sort.Strings(unexpected)
	return unexpected
}

func publicRenderableInventories() []map[components.Kind]components.Component {
	return []map[components.Kind]components.Component{
		compositionRenderables(),
		displayRenderables(),
		inputRenderables(),
		feedbackRenderables(),
		navigationRenderables(),
	}
}

func allowedRenderableMethods(t *testing.T) map[string]struct{} {
	t.Helper()

	const componentPackagePrefix = "github.com/araihu/goshtoso/components/"

	methods := make(map[string]struct{}, 170)
	for _, inventory := range publicRenderableInventories() {
		for _, value := range inventory {
			valueType := reflect.TypeOf(value)
			for valueType.Kind() == reflect.Pointer {
				valueType = valueType.Elem()
			}
			require.NotEmpty(t, valueType.Name(), "renderable must use a named concrete type")
			require.True(t, strings.HasPrefix(valueType.PkgPath(), componentPackagePrefix),
				"renderable %s must live under the components module", valueType)

			receiver := strings.TrimPrefix(valueType.PkgPath(), componentPackagePrefix) +
				"." + valueType.Name()
			methods[receiver+".Kind"] = struct{}{}
			methods[receiver+".Render"] = struct{}{}
		}
	}
	require.Len(t, methods, 170)
	return methods
}

func exportedGoFunctions(t *testing.T) (map[string]string, map[string]string) {
	t.Helper()

	functions, methods, err := exportedGoFunctionsFromDir(".")
	require.NoError(t, err)
	return functions, methods
}

func exportedGoFunctionsFromDir(root string) (map[string]string, map[string]string, error) {
	functions := make(map[string]string)
	methods := make(map[string]string)
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") ||
			strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		packageDir, relativeErr := filepath.Rel(root, filepath.Dir(path))
		if relativeErr != nil {
			return relativeErr
		}
		packagePath := filepath.ToSlash(packageDir)
		if packagePath == "." {
			packagePath = "components"
		}

		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || !function.Name.IsExported() {
				continue
			}
			if function.Recv == nil {
				functions[packagePath+"."+function.Name.Name] = path
				continue
			}
			receiver, receiverErr := receiverName(function.Recv.List[0].Type)
			if receiverErr != nil {
				return fmt.Errorf("%s: exported method %s: %w", path, function.Name.Name, receiverErr)
			}
			methods[packagePath+"."+receiver+"."+function.Name.Name] = path
		}
		return nil
	})

	return functions, methods, err
}

func receiverName(expression ast.Expr) (string, error) {
	switch expression := expression.(type) {
	case *ast.Ident:
		return expression.Name, nil
	case *ast.StarExpr:
		return receiverName(expression.X)
	case *ast.IndexExpr:
		return receiverName(expression.X)
	case *ast.IndexListExpr:
		return receiverName(expression.X)
	case *ast.ParenExpr:
		return receiverName(expression.X)
	default:
		return "", fmt.Errorf("unsupported receiver syntax %T", expression)
	}
}

func writeGoFixture(t *testing.T, root, name, contents string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(name))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
}

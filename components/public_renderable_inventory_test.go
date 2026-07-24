package components_test

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components"
	"github.com/stretchr/testify/require"
)

func TestPublicRenderableInventoryMatchesAllKinds(t *testing.T) {
	inventories := []map[components.Kind]components.Component{
		displayRenderables(),
		inputRenderables(),
		feedbackRenderables(),
		navigationRenderables(),
	}

	got := make([]components.Kind, 0, 74)
	for _, inventory := range inventories {
		for want, value := range inventory {
			require.Equal(t, want, value.Kind())
			got = append(got, value.Kind())
		}
	}

	require.ElementsMatch(t, components.AllKinds(), got)
	require.Len(t, got, 74)
	require.Len(t, got, len(components.AllKinds()))
}

func TestPublicFunctionSurfaceMatchesContract(t *testing.T) {
	allowedFunctions := map[string]struct{}{
		"components.AllKinds": {},

		"accordion.Accordion":             {},
		"alert.Alert":                     {},
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
		"combobox.Combobox":               {},
		"drawer.Drawer":                   {},
		"dropdown.Dropdown":               {},
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
		"kbd.Kbd":                         {},
		"link.Link":                       {},
		"modal.Modal":                     {},
		"modal.AlertDialog":               {},
		"navbar.Navbar":                   {},
		"pagination.Pagination":           {},
		"palette.Palette":                 {},
		"radio.Radio":                     {},
		"radio.RadioBar":                  {},
		"radio.RadioGroup":                {},
		"range.Range":                     {},
		"rating.Rating":                   {},
		"rating.RatingDisplay":            {},
		"schemaform.Fields":               {},
		"search.Search":                   {},
		"search.SearchField":              {},
		"search.SearchModal":              {},
		"select.Select":                   {},
		"sidebar.Sidebar":                 {},
		"sidebar.Overlay":                 {},
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
		"tooltip.Tooltip":                 {},

		"button.WithTone":        {},
		"button.WithSize":        {},
		"button.WithType":        {},
		"button.Disabled":        {},
		"button.WithID":          {},
		"button.WithRootClass":   {},
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
		"combobox.Config.Validate":               {},
		"combobox.Config.InitialState":           {},
		"combobox.comboHandler.ServeHTTP":        {},
		"pagination.Config.HasPrevious":          {},
		"pagination.Config.HasNext":              {},
		"pagination.Config.PreviousPage":         {},
		"pagination.Config.NextPage":             {},
		"pagination.Config.PageURL":              {},
		"pagination.Config.Pages":                {},
		"search.Item.SearchText":                 {},
		"search.Item.NormalizedMethod":           {},
		"search.Item.SafeHref":                   {},
		"table.Config.IsSortedBy":                {},
		"table.Config.NextSortDir":               {},
		"table.Config.SortURL":                   {},
		"table.Config.PageURL":                   {},
		"table.Config.NextPageURL":               {},
		"form/validation.FormDef.Bind":           {},
		"form/validation.FormDef.Dependents":     {},
		"form/validation.FormDef.PopulateValues": {},
	}

	gotFunctions, gotMethods := exportedGoFunctions(t)

	var unexpected []string
	for name := range gotFunctions {
		if _, ok := allowedFunctions[name]; !ok {
			unexpected = append(unexpected, name)
		}
	}
	for name, filename := range gotMethods {
		if _, ok := allowedMethods[name]; ok {
			continue
		}
		if (strings.HasSuffix(name, ".Kind") || strings.HasSuffix(name, ".Render")) &&
			filepath.Base(filename) == "component.go" {
			continue
		}
		unexpected = append(unexpected, name)
	}
	sort.Strings(unexpected)
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
	sort.Strings(missing)
	require.Empty(t, missing, "contracted public component functions or methods disappeared")

	require.Empty(t, exportedTemplComponents(t), "templ render helpers must remain package-private")
}

func exportedGoFunctions(t *testing.T) (map[string]string, map[string]string) {
	t.Helper()

	functions := make(map[string]string)
	methods := make(map[string]string)
	fset := token.NewFileSet()
	require.NoError(t, filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") ||
			strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_templ.go") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		packagePath := filepath.ToSlash(filepath.Dir(path))
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
			receiver := receiverName(function.Recv.List[0].Type)
			methods[packagePath+"."+receiver+"."+function.Name.Name] = path
		}
		return nil
	}))

	return functions, methods
}

func receiverName(expression ast.Expr) string {
	switch expression := expression.(type) {
	case *ast.Ident:
		return expression.Name
	case *ast.StarExpr:
		return receiverName(expression.X)
	default:
		return "<unknown>"
	}
}

func exportedTemplComponents(t *testing.T) []string {
	t.Helper()

	var exported []string
	require.NoError(t, filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".templ") {
			return nil
		}

		file, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}

		scanner := bufio.NewScanner(file)
		line := 0
		for scanner.Scan() {
			line++
			fields := strings.Fields(scanner.Text())
			if len(fields) < 2 || fields[0] != "templ" {
				continue
			}
			name, _, _ := strings.Cut(fields[1], "(")
			if ast.IsExported(name) {
				exported = append(exported, path+":"+name)
			}
		}
		if scanErr := scanner.Err(); scanErr != nil {
			_ = file.Close()
			return scanErr
		}
		return file.Close()
	}))
	sort.Strings(exported)
	return exported
}

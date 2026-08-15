package conformanceledger

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var stateFixtureAdapters = map[string]string{
	"accordion/Appearance":       `accordion.Accordion(accordion.AccordionConfig{ID:%[2]q, Appearance:accordion.%[1]s})`,
	"alert/Tone":                 `alert.Alert(alert.Config{Title:"State", Description:"State", Tone:alert.%[1]s})`,
	"avatar/Radius":              `avatar.Avatar(avatar.Config{Name:"State", Shape:avatar.ShapeSquare, Radius:avatar.%[1]s})`,
	"avatar/Shape":               `avatar.Avatar(avatar.Config{Name:"State", Shape:avatar.%[1]s})`,
	"avatar/Size":                `avatar.Avatar(avatar.Config{Name:"State", Size:avatar.%[1]s})`,
	"avatar/Status":              `avatar.Avatar(avatar.Config{Name:"State", Status:avatar.%[1]s})`,
	"avatar/Tone":                `avatar.Avatar(avatar.Config{Name:"State", Tone:avatar.%[1]s})`,
	"badge/Appearance":           `badge.Badge(badge.Config{Label:"State", Appearance:badge.%[1]s})`,
	"badge/Size":                 `badge.Badge(badge.Config{Label:"State", Size:badge.%[1]s})`,
	"badge/Tone":                 `badge.Badge(badge.Config{Label:"State", Tone:badge.%[1]s})`,
	"banner/Position":            `banner.Banner(banner.Config{Description:"State", Position:banner.%[1]s})`,
	"banner/Tone":                `banner.Banner(banner.Config{Description:"State", Tone:banner.%[1]s})`,
	"breadcrumbs/SeparatorStyle": `breadcrumbs.Breadcrumbs(breadcrumbs.Config{Current:"State", Separator:breadcrumbs.%[1]s})`,
	"button/Size":                `conformanceWithChildren(button.Button(button.WithSize(button.%[1]s), button.WithID(%[2]q)), "State")`,
	"button/Tone":                `conformanceWithChildren(button.Button(button.WithTone(button.%[1]s), button.WithID(%[2]q)), "State")`,
	"card/Appearance":            `card.Card(card.Config{Title:"State", Appearance:card.%[1]s})`,
	"card/Interaction":           `card.Card(card.Config{Title:"State", Interaction:card.%[1]s})`,
	"card/Layout":                `card.Card(card.Config{Title:"State", Layout:card.%[1]s})`,
	"chatbubble/Side":            `chatbubble.ChatBubble(chatbubble.Config{Message:"State", Side:chatbubble.%[1]s})`,
	"chatbubble/Status":          `chatbubble.ChatBubble(chatbubble.Config{Message:"State", Status:chatbubble.%[1]s})`,
	"checkbox/Animation":         `checkbox.Checkbox(checkbox.Config{ID:%[2]q, Label:"State", Animation:checkbox.%[1]s})`,
	"checkbox/Icon":              `checkbox.Checkbox(checkbox.Config{ID:%[2]q, Label:"State", Icon:checkbox.%[1]s})`,
	"checkbox/Tone":              `checkbox.Checkbox(checkbox.Config{ID:%[2]q, Label:"State", Tone:checkbox.%[1]s})`,
	"codeblock/Density":          `codeblock.CodeBlock(codeblock.Config{ID:%[2]q, Language:"go", Code:"package state", Density:codeblock.%[1]s})`,
	"combobox/Mode":              `combobox.Combobox(combobox.Config{ID:%[2]q, Label:"State", Mode:combobox.%[1]s}, combobox.State{})`,
	"drawer/Height":              `drawer.Drawer(drawer.Config{ID:%[2]q, Title:"State", Side:drawer.SideTop, Height:drawer.%[1]s})`,
	"drawer/Side":                `drawer.Drawer(drawer.Config{ID:%[2]q, Title:"State", Side:drawer.%[1]s})`,
	"drawer/Width":               `drawer.Drawer(drawer.Config{ID:%[2]q, Title:"State", Width:drawer.%[1]s})`,
	"dropdown/MenuAlign":         `dropdown.Dropdown(dropdown.Config{ID:%[2]q, Label:"State", MenuAlign:dropdown.%[1]s})`,
	"dropdown/TriggerMode":       `dropdown.Dropdown(dropdown.Config{ID:%[2]q, Label:"State", TriggerMode:dropdown.%[1]s})`,
	"fileinput/Appearance":       `fileinput.FileInput(fileinput.Config{ID:%[2]q, Label:"State", Appearance:fileinput.%[1]s})`,
	"head/Dependency":            `conformanceDependencyFixture(head.%[1]s)`,
	"head/OpenGraphType":         `head.Metadata(conformanceMetadata(head.%[1]s, head.TwitterCardSummaryLargeImage))`,
	"head/TwitterCard":           `head.Metadata(conformanceMetadata(head.OpenGraphTypeWebsite, head.%[1]s))`,
	"icon/Mode":                  `icon.Icon(icon.Config{SpriteURL:heroicons.SpriteURL, Symbol:heroicons.Icon16SolidCheckCircle, Label:"State", Mode:icon.%[1]s})`,
	"icon/Size":                  `icon.Icon(icon.Config{SpriteURL:heroicons.SpriteURL, Symbol:heroicons.Icon16SolidCheckCircle, Label:"State", Size:icon.%[1]s})`,
	"kbd/Size":                   `kbd.Kbd("State", kbd.WithSize(kbd.%[1]s))`,
	"link/Appearance":            `conformanceWithChildren(link.Link("/state", link.WithAppearance(link.%[1]s), link.WithID(%[2]q)), "State")`,
	"link/IconPosition":          `conformanceWithChildren(link.Link("/state", link.WithIcon(icon.Icon(icon.Config{SpriteURL:heroicons.SpriteURL, Symbol:heroicons.Icon16SolidCheckCircle, Decorative:true})), link.WithIconPosition(link.%[1]s), link.WithID(%[2]q)), "State")`,
	"link/Size":                  `conformanceWithChildren(link.Link("/state", link.WithAppearance(link.AppearanceButton), link.WithSize(link.%[1]s), link.WithID(%[2]q)), "State")`,
	"modal/Tone":                 `modal.AlertDialog(modal.AlertDialogConfig{ID:%[2]q, Title:"State", Body:"State", TriggerLabel:"Open state", ActionLabel:"Close state", Tone:modal.%[1]s})`,
	"navbar/ActionPosition":      `navbar.Navbar(navbar.Config{Actions:[]navbar.ActionItem{{Content:conformanceWithChildren(button.Button(button.WithAttrs(templ.Attributes{"aria-label":"State action"})), "State"), Position:navbar.%[1]s}}})`,
	"pagination/Mode":            `pagination.Pagination(pagination.Config{ID:%[2]q, Mode:pagination.%[1]s, CurrentPage:1, TotalPages:3, BaseURL:"/state"})`,
	"panel/Appearance":           `panel.Panel(panel.Config{Appearance:panel.%[1]s, Body:templ.Raw("State")})`,
	"panel/Density":              `panel.Panel(panel.Config{Density:panel.%[1]s, Body:templ.Raw("State")})`,
	"radio/Size":                 `radio.Radio(radio.Config{ID:%[2]q, Name:%[2]q, Label:"State", Size:radio.%[1]s})`,
	"radio/Tone":                 `radio.Radio(radio.Config{ID:%[2]q, Name:%[2]q, Label:"State", Tone:radio.%[1]s})`,
	"rating/Appearance":          `rating.Rating(rating.Config{ID:%[2]q, Name:%[2]q, Label:"State", Appearance:rating.%[1]s})`,
	"rating/Size":                `rating.Rating(rating.Config{ID:%[2]q, Name:%[2]q, Label:"State", Size:rating.%[1]s})`,
	"schemaform/AllowMode":       `schemaform.Fields(schemaform.FieldsConfig{Fields:[]schemaform.Field{{Path:%[2]q, Name:%[2]q, Label:"State", HelperText:"State help", Kind:schemaform.KindString, Managed:schemaform.%[1]s==schemaform.AllowModeManaged}}})`,
	"schemaform/Kind":            `schemaform.Fields(schemaform.FieldsConfig{Fields:[]schemaform.Field{{Path:%[2]q, Name:%[2]q, Label:"State", HelperText:"State help", Kind:schemaform.%[1]s}}})`,
	"search/MatchMode":           `search.Search(search.Config{ID:%[2]q, Label:"State", MatchMode:search.%[1]s})`,
	"selectfield/State":          `selectcomponent.Select(selectcomponent.Config{ID:%[2]q, Name:%[2]q, Label:"State", State:selectcomponent.%[1]s})`,
	"skeleton/Shape":             `skeleton.Skeleton(skeleton.Config{Shape:skeleton.%[1]s, Label:"State"})`,
	"spinner/Size":               `spinner.Spinner(spinner.Config{Size:spinner.%[1]s})`,
	"spinner/Tone":               `spinner.Spinner(spinner.Config{Tone:spinner.%[1]s})`,
	"steps/Orientation":          `steps.Steps(steps.Config{ID:%[2]q, Orientation:steps.%[1]s, Steps:[]steps.Step{{ID:%[2]q+"-step", Label:"State"}}})`,
	"steps/Status":               `steps.Steps(steps.Config{ID:%[2]q, Steps:[]steps.Step{{ID:%[2]q+"-step", Label:"State", Status:steps.%[1]s}}})`,
	"structuredinput/ColumnType": `structuredinput.StructuredInput(structuredinput.Config{ID:%[2]q, Name:%[2]q, Columns:[]structuredinput.Column{{Key:"value", Label:"State", Type:structuredinput.%[1]s}}})`,
	"table/Appearance":           `table.Table(table.Config{ID:%[2]q, Appearance:table.%[1]s, Columns:conformanceColumns(), Rows:conformanceRows()})`,
	"table/FilterAppearance":     `table.Table(table.Config{ID:%[2]q, Columns:conformanceColumns(), Rows:conformanceRows(), Filters:&table.FilterConfig{Appearance:table.%[1]s}})`,
	"table/FilterType":           `table.Table(table.Config{ID:%[2]q, Columns:conformanceColumns(), Rows:conformanceRows(), Filters:&table.FilterConfig{Filters:[]table.Filter{{Key:"query", Label:"State", Placeholder:"State", Type:table.%[1]s}}}})`,
	"table/LinkMode":             `table.Table(table.Config{ID:%[2]q, Columns:conformanceColumns(), Rows:[]table.Row{{ID:"row", Link:"/state", LinkMode:table.%[1]s, Cells:map[string]table.Cell{"name":{Text:"State"}}}}})`,
	"table/PaginationMode":       `table.Table(table.Config{ID:%[2]q, Columns:conformanceColumns(), Rows:conformanceRows(), Pagination:&table.PaginationConfig{Mode:table.%[1]s, CurrentPage:1, TotalPages:2}})`,
	"table/SortDir":              `table.Table(table.Config{ID:%[2]q, Columns:conformanceColumns(), Rows:conformanceRows(), SortBy:"name", SortDir:table.%[1]s})`,
	"textarea/State":             `textarea.Textarea(textarea.Config{ID:%[2]q, Name:%[2]q, Label:"State", State:textarea.%[1]s})`,
	"textinput/InputType":        `textinput.TextInput(textinput.Config{ID:%[2]q, Name:%[2]q, Label:"State", Type:textinput.%[1]s})`,
	"textinput/State":            `textinput.TextInput(textinput.Config{ID:%[2]q, Name:%[2]q, Label:"State", State:textinput.%[1]s})`,
	"toast/Tone":                 `toast.Toast(toast.Config{Title:"State", Message:"State", DisplayDuration:-1, Tone:toast.%[1]s})`,
	"toggle/Appearance":          `toggle.Toggle(toggle.Config{ID:%[2]q, Label:"State", Appearance:toggle.%[1]s})`,
	"toggle/Tone":                `toggle.Toggle(toggle.Config{ID:%[2]q, Label:"State", Tone:toggle.%[1]s})`,
	"tooltip/Activation":         `tooltip.Tooltip(%[2]q, "State", tooltip.WithActivation(tooltip.%[1]s))`,
	"tooltip/Position":           `tooltip.Tooltip(%[2]q, "State", tooltip.WithPosition(tooltip.%[1]s))`,
}

var defaultStateFixtureAdapters = map[string]string{
	"KindActionGroup": `actiongroup.ActionGroup(actiongroup.Config{Label:"State actions", Primary:actiongroup.Action{ID:%[1]q, Label:"State"}})`,
	"KindAlertDialog": `modal.AlertDialog(modal.AlertDialogConfig{ID:%[1]q, Title:"State", Body:"State", TriggerLabel:"Open state", ActionLabel:"Close state"})`,
	"KindButton":      `conformanceWithChildren(button.Button(button.WithID(%[1]q)), "State")`,
	// The public Kind inventory deliberately instantiates these constructors with
	// zero Config values to prove kind identity.  That is not a behavioral
	// carousel fixture: carousel navigation has a source-defined Slides input.
	// Keep the state row itself at the default (no exported variant selected),
	// while giving the browser harness the minimum two-slide data needed to
	// exercise the documented navigation and touch paths.  Do not replace this
	// with a synthetic DOM fixture; it remains the real public constructor.
	"KindCardCarousel":           `carousel.CardCarousel(carousel.CardConfig{ID:%[1]q, Touch:true, Slides:[]carousel.Slide{{ImgSrc:"/assets/images/carousel/slide-1.webp", ImgAlt:"Conformance slide A", Title:"Conformance A", Description:"State fixture"}, {ImgSrc:"/assets/images/carousel/slide-2.webp", ImgAlt:"Conformance slide B", Title:"Conformance B", Description:"State fixture"}}})`,
	"KindCarousel":               `carousel.Carousel(carousel.Config{ID:%[1]q, Touch:true, AspectRatio:"3/1", Slides:[]carousel.Slide{{ImgSrc:"/assets/images/carousel/slide-1.webp", ImgAlt:"Conformance slide A"}, {ImgSrc:"/assets/images/carousel/slide-2.webp", ImgAlt:"Conformance slide B"}}})`,
	"KindCheckbox":               `checkbox.Checkbox(checkbox.Config{ID:%[1]q, Label:"State"})`,
	"KindDependencies":           `conformanceDependenciesFixture(false)`,
	"KindDependenciesMinimal":    `conformanceDependenciesFixture(true)`,
	"KindDropdown":               `dropdown.Dropdown(dropdown.Config{ID:%[1]q, Label:"State", Sections:[]dropdown.Section{{Items:[]dropdown.Item{{Label:"State item", Href:"/state"}}}}})`,
	"KindFileInput":              `fileinput.FileInput(fileinput.Config{ID:%[1]q, Name:%[1]q, Label:"State"})`,
	"KindFormCollapsibleSection": `form.CollapsibleSection(form.CollapsibleSectionConfig{SectionConfig:form.SectionConfig{ID:%[1]q, Title:"State"}})`,
	"KindFormFlipSection":        `form.FlipSection(form.FlipSectionConfig{SectionConfig:form.SectionConfig{ID:%[1]q, Title:"State"}}, templ.Raw("State"))`,
	"KindLink":                   `conformanceWithChildren(link.Link("/state", link.WithID(%[1]q)), "State")`,
	"KindModal":                  `modal.Modal(modal.Config{ID:%[1]q, Title:"State", Body:"State", TriggerLabel:"Open state", PrimaryLabel:"Close state"})`,
	"KindRadio":                  `radio.Radio(radio.Config{ID:%[1]q, Name:%[1]q, Label:"State"})`,
	"KindRange":                  `rangecomponent.Range(rangecomponent.Config{ID:%[1]q, Name:%[1]q, Label:"State"})`,
	"KindRating":                 `rating.Rating(rating.Config{ID:%[1]q, Name:%[1]q, Label:"State"})`,
	"KindRatingDisplay":          `rating.RatingDisplay(rating.DisplayConfig{ID:%[1]q, Label:"State"})`,
	"KindSearch":                 `search.Search(search.Config{ID:%[1]q, Label:"State"})`,
	"KindSearchField":            `search.Search(search.Config{ID:%[1]q, Label:"State"})`,
	"KindSearchModal":            `search.SearchModal(search.Config{ID:%[1]q, Label:"State"})`,
	// A zero ScrollRegion config has no content and therefore no real scroll
	// lifecycle to inspect. This valid public configuration supplies bounded
	// overflow content while keeping the generated state at its default Kind.
	"KindScrollRegion":        `scrollregion.ScrollRegion(scrollregion.Config{RootClass:"h-48", Content:templ.Raw("<div class=\\\"h-96\\\" data-conformance-scroll-content>State scroll content</div>")})`,
	"KindTextarea":            `textarea.Textarea(textarea.Config{ID:%[1]q, Name:%[1]q, Label:"State"})`,
	"KindTextareaWithActions": `textarea.TextareaWithActions(textarea.Config{ID:%[1]q, Name:%[1]q, Label:"State", InputAttrs:templ.Attributes{"aria-label":"State"}})`,
	"KindTextInput":           `textinput.TextInput(textinput.Config{ID:%[1]q, Name:%[1]q, Label:"State"})`,
	"KindToggle":              `toggle.Toggle(toggle.Config{ID:%[1]q, Label:"State"})`,
}

func GenerateStateFixtureSource(repoRoot string) ([]byte, error) {
	if err := ValidateStateExecutionContracts(repoRoot); err != nil {
		return nil, err
	}
	inventory, err := DeriveInventory(repoRoot)
	if err != nil {
		return nil, err
	}
	defaults, imports, err := deriveDefaultFixtureExpressions(repoRoot)
	if err != nil {
		return nil, err
	}
	for _, path := range stateFixtureImportPaths() {
		imports[filepath.Base(path)] = path
	}
	imports["templ"] = "github.com/a-h/templ"
	imports["heroicons"] = "github.com/araihu/goshtoso/components/icon/heroicons"

	var cases []string
	for _, state := range inventory.States {
		parts := strings.SplitN(state.Value, "/", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid state %q", state.Value)
		}
		var expression string
		if strings.HasSuffix(parts[1], "/default") || parts[1] == "default" {
			if adapter := defaultStateFixtureAdapters[state.Source.Symbol]; adapter != "" {
				expression = adapter
				if strings.Contains(adapter, "%[1]") {
					expression = fmt.Sprintf(adapter, stateFixtureID(state.Value))
				}
			} else {
				expression = defaults[state.Source.Symbol]
			}
		} else {
			typeAndConstant := strings.Split(parts[1], ".")
			if len(typeAndConstant) != 2 {
				return nil, fmt.Errorf("invalid typed state %q", state.Value)
			}
			adapter := stateFixtureAdapters[parts[0]+"/"+typeAndConstant[0]]
			if adapter == "" {
				return nil, fmt.Errorf("missing state fixture adapter for %s", parts[0]+"/"+typeAndConstant[0])
			}
			expression = fmt.Sprintf(adapter, typeAndConstant[1], stateFixtureID(state.Value))
		}
		if expression == "" {
			return nil, fmt.Errorf("missing default fixture for %s (%s)", state.Value, state.Source.Symbol)
		}
		cases = append(cases, fmt.Sprintf("{State:%q, Component:%s},", state.Value, expression))
	}
	// Identity authorities import every package they reference, while a valid
	// behavioral adapter may replace the one expression that used an imported
	// package. Emit only imports still referenced by generated expressions so
	// a source-valid fixture can never leave the tracked Go test uncompilable.
	allCases := strings.Join(cases, "\n")
	for alias := range imports {
		if !strings.Contains(allCases, alias+".") {
			delete(imports, alias)
		}
	}

	var source bytes.Buffer
	source.WriteString("// Code generated by conformanceledger.GenerateStateFixtureSource; DO NOT EDIT.\n//go:build e2e && full\n\npackage e2e\n\nimport (\n")
	aliases := make([]string, 0, len(imports))
	for alias := range imports {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	for _, alias := range aliases {
		fmt.Fprintf(&source, "%s %q\n", alias, imports[alias])
	}
	source.WriteString(")\n\nfunc conformanceStateFixtures() []conformanceStateFixture { return []conformanceStateFixture{\n")
	for _, item := range cases {
		source.WriteString(item + "\n")
	}
	source.WriteString("}}\n")
	formatted, err := format.Source(source.Bytes())
	if err != nil {
		lines := strings.Split(source.String(), "\n")
		return nil, fmt.Errorf("format generated state fixtures: %w; around line 186: %s", err, strings.Join(lines[180:min(195, len(lines))], "\n"))
	}
	return formatted, nil
}

func deriveDefaultFixtureExpressions(repoRoot string) (map[string]string, map[string]string, error) {
	expressions := map[string]string{}
	imports := map[string]string{}
	fset := token.NewFileSet()
	for _, relative := range []string{"components/composition_identity_test.go", "components/display_identity_test.go", "components/input_identity_test.go", "components/feedback_navigation_identity_test.go"} {
		file, err := parser.ParseFile(fset, filepath.Join(repoRoot, relative), nil, 0)
		if err != nil {
			return nil, nil, err
		}
		for _, spec := range file.Imports {
			path, _ := strconv.Unquote(spec.Path.Value)
			alias := filepath.Base(path)
			if spec.Name != nil {
				alias = spec.Name.Name
			}
			if alias != "_" && alias != "." && (path == "github.com/araihu/goshtoso/components" || strings.HasPrefix(path, "github.com/araihu/goshtoso/components/")) {
				imports[alias] = path
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}
			for _, element := range literal.Elts {
				entry, ok := element.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				selector, ok := entry.Key.(*ast.SelectorExpr)
				if !ok || !strings.HasPrefix(selector.Sel.Name, "Kind") {
					continue
				}
				qualifier, ok := selector.X.(*ast.Ident)
				if !ok || qualifier.Name != "components" {
					continue
				}
				var output bytes.Buffer
				_ = format.Node(&output, fset, entry.Value)
				expressions[selector.Sel.Name] = output.String()
			}
			return true
		})
	}
	return expressions, imports, nil
}

func stateFixtureID(state string) string {
	replacer := strings.NewReplacer("/", "-", ".", "-", "_", "-")
	return "conformance-" + strings.ToLower(replacer.Replace(state))
}

func stateFixtureImportPaths() []string {
	paths := make([]string, 0, len(stateFixtureAdapters))
	seen := map[string]struct{}{}
	for key := range stateFixtureAdapters {
		pkg := strings.SplitN(key, "/", 2)[0]
		path := "github.com/araihu/goshtoso/components/" + pkg
		if pkg == "selectfield" {
			continue
		}
		if _, ok := seen[path]; !ok {
			seen[path] = struct{}{}
			paths = append(paths, path)
		}
	}
	return paths
}

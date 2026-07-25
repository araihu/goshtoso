package docs_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/kbd"
	"github.com/araihu/goshtoso/components/link"
	"github.com/araihu/goshtoso/components/modal"
	"github.com/araihu/goshtoso/components/pagination"
	"github.com/araihu/goshtoso/components/tooltip"
)

const (
	releaseBaseTag    = "v0.0.11"
	releaseBaseCommit = "10b4dcbf3da3c1dd534d8d2baa949d043b9d0f1f"
)

func Example_sharedComponentIdentity() {
	pageComponents := []components.Component{
		badge.Badge(badge.Config{Label: "New"}),
		button.Button(),
	}

	for _, component := range pageComponents {
		switch component.Kind() {
		case components.KindBadge:
			fmt.Println("badge")
		case components.KindButton:
			fmt.Println("button")
		}
	}

	// Output:
	// badge
	// button
}

func Example_configurationDimensions() {
	badgeValue := badge.Badge(badge.Config{
		Label:      "Requires attention",
		Tone:       badge.ToneDanger,
		Appearance: badge.AppearanceSoft,
		Size:       badge.SizeSM,
	})
	paginationValue := pagination.Pagination(pagination.Config{
		Mode:        pagination.ModeSimple,
		CurrentPage: 2,
		TotalPages:  5,
	})

	fmt.Println(badgeValue.Kind())
	fmt.Println(paginationValue.Kind())

	// Output:
	// badge
	// pagination
}

func Example_splitPrimitives() {
	modalValue := modal.Modal(modal.Config{
		ID:           "profile",
		Title:        "Edit profile",
		TriggerLabel: "Edit",
	})
	alertValue := modal.AlertDialog(modal.AlertDialogConfig{
		ID:           "remove-profile",
		Title:        "Remove profile?",
		TriggerLabel: "Remove",
		ActionLabel:  "Remove",
		Tone:         modal.ToneDanger,
	})

	fmt.Println(modalValue.Kind())
	fmt.Println(alertValue.Kind())

	// Output:
	// modal
	// alert-dialog
}

func Example_atomicFunctionalOptions() {
	values := []components.Component{
		button.Button(
			button.WithTone(button.ToneDanger),
			button.WithSize(button.SizeSmall),
			button.WithType("submit"),
		),
		link.Link("/docs",
			link.WithAppearance(link.AppearanceButton),
			link.WithTarget("_blank"),
		),
		kbd.Kbd("K", kbd.WithSize(kbd.SizeSM), kbd.WithLabel("Command K")),
		tooltip.Tooltip("save-help", "Saves changes",
			tooltip.WithPosition(tooltip.PositionBottom),
			tooltip.WithActivation(tooltip.ActivationClick),
		),
	}

	for _, value := range values {
		fmt.Println(value.Kind())
	}

	// Output:
	// button
	// link
	// kbd
	// tooltip
}

func Example_concreteComponentReturn() {
	concrete := button.Button(button.WithID("save"))
	var component components.Component = concrete

	switch component.Kind() {
	case components.KindButton:
		fmt.Printf("%T: %s\n", concrete, component.Kind())
	}

	// Output:
	// button.Instance: button
}

func TestReleaseMigrationDocsLockBaseAndEntryPoints(t *testing.T) {
	changelog := readDoc(t, "../CHANGELOG.md")
	guide := readDoc(t, "MIGRATING_COMPONENT_API.md")
	readme := readDoc(t, "../README.md")

	for name, document := range map[string]string{
		"CHANGELOG.md":               changelog,
		"MIGRATING_COMPONENT_API.md": guide,
	} {
		if !strings.Contains(document, "`"+releaseBaseTag+"`") {
			t.Errorf("%s does not lock release base tag %s", name, releaseBaseTag)
		}
		if !strings.Contains(document, "`"+releaseBaseCommit+"`") {
			t.Errorf("%s does not lock release base commit %s", name, releaseBaseCommit)
		}
	}

	for _, required := range []string{
		"Unreleased",
		"Breaking",
		"v0.0.11",
		"docs/MIGRATING_COMPONENT_API.md",
	} {
		if !strings.Contains(changelog, required) {
			t.Errorf("CHANGELOG.md missing %q", required)
		}
	}
	for _, required := range []string{
		"v0.0.11",
		"Source-breaking",
		"Behavior and effective defaults",
		"Mechanical upgrade checklist",
	} {
		if !strings.Contains(guide, required) {
			t.Errorf("MIGRATING_COMPONENT_API.md missing %q", required)
		}
	}
	for _, required := range []string{
		"CHANGELOG.md",
		"docs/MIGRATING_COMPONENT_API.md",
	} {
		if !strings.Contains(readme, required) {
			t.Errorf("README.md missing release-documentation link %q", required)
		}
	}
}

func TestMigrationGuideCoversEveryBreakingFamily(t *testing.T) {
	guide := readDoc(t, "MIGRATING_COMPONENT_API.md")
	for _, required := range []string{
		"components.Component",
		"Kind()",
		"AllKinds",
		"Tone",
		"Appearance",
		"Mode",
		"no universal `Variant`",
		"split primitives",
		"functional options",
		"concrete return",
		"removed and private",
		"effective default",
		"documentation contract",
		"smoke-test contract",
	} {
		if !strings.Contains(guide, required) {
			t.Errorf("migration guide missing breaking family marker %q", required)
		}
	}
}

func TestChangelogCoversEveryReleaseNoteFamily(t *testing.T) {
	changelog := readDoc(t, "../CHANGELOG.md")
	for _, required := range []string{
		"components.Component",
		"AllKinds()",
		"Tone",
		"Appearance",
		"Mode",
		"Split primitives",
		"functional options",
		"concrete instance types",
		"Removed or privatized",
		"Behavior and effective defaults",
		"Documentation",
		"smoke coverage",
	} {
		if !strings.Contains(changelog, required) {
			t.Errorf("CHANGELOG.md missing release-note family %q", required)
		}
	}
}

func TestMigrationGuideContainsExactOldToNewMappings(t *testing.T) {
	guide := readDoc(t, "MIGRATING_COMPONENT_API.md")
	mappings := [][2]string{
		{"accordion.Variant", "accordion.Appearance"},
		{"accordion.Default", "accordion.AppearanceDefault"},
		{"accordion.NoBackground", "accordion.AppearancePlain"},
		{"accordion.Split", "accordion.AppearanceSplit"},
		{"accordion.SingleOpen", "AllowMultiple"},
		{"alert.Variant", "alert.Tone"},
		{"alert.Info", "alert.ToneInfo"},
		{"alert.Success", "alert.ToneSuccess"},
		{"alert.Warning", "alert.ToneWarning"},
		{"alert.Danger", "alert.ToneDanger"},
		{"avatar.Variant", "avatar.Tone"},
		{"badge.Variant", "badge.Tone"},
		{"badge.Style", "badge.Appearance"},
		{"banner.Variant", "banner.Tone"},
		{"button.Config", "button.Option"},
		{"button.Variant", "button.Tone"},
		{"card.Variant", "card.Appearance"},
		{"carousel.OnCard", "carousel.CardCarousel"},
		{"chatbubble.Config.AvatarVariant", "chatbubble.Config.AvatarTone"},
		{"checkbox.Variant", "checkbox.Tone"},
		{"fileinput.Variant", "fileinput.Appearance"},
		{"kbd.Config", "kbd.Option"},
		{"link.Config", "link.Option"},
		{"link.Style", "link.Appearance"},
		{"modal.Config.AlertMode", "modal.AlertDialog"},
		{"modal.Variant", "modal.Tone"},
		{"pagination.Variant", "pagination.Mode"},
		{"radio.Variant", "radio.Tone"},
		{"rating.Style", "rating.Appearance"},
		{"rating.Config.ReadOnly", "rating.RatingDisplay"},
		{"spinner.Variant", "spinner.Tone"},
		{"table.Variant", "table.Appearance"},
		{"table.FilterVariant", "table.FilterAppearance"},
		{"toast.Variant", "toast.Tone"},
		{"toast.Message", "toast.MessageToast"},
		{"toast.Container", "toast.ToastContainer"},
		{"toggle.Variant", "toggle.Tone"},
		{"toggle.Style", "toggle.Appearance"},
		{"tooltip.Config", "tooltip.Option"},
		{"tooltip.Trigger", "tooltip.Activation"},
		{"tooltip.Top", "tooltip.PositionTop"},
	}

	for _, mapping := range mappings {
		if !lineContainsBoth(guide, mapping[0], mapping[1]) {
			t.Errorf("migration guide missing same-row mapping %q -> %q", mapping[0], mapping[1])
		}
	}
}

func TestMigrationGuideNamesExactCuratedRemovals(t *testing.T) {
	guide := readDoc(t, "MIGRATING_COMPONENT_API.md")
	removed := []string{
		"accordion.AccordionItemData",
		"avatar.UserIcon",
		"card.Config.Price",
		"card.Config.Rating",
		"card.Config.HasPrice",
		"card.Config.HasRating",
		"card.StarRating",
		"combobox.Body",
		"combobox.OptionsList",
		"combobox.ClientScript",
		"combobox.ProviderError",
		"combobox.BodyOOB",
		"combobox.TriggerLabelOOB",
		"combobox.Option.Meta",
		"combobox.Option.Img",
		"combobox.Option.Initials",
		"combobox.Option.Badge",
		"combobox.Option.BadgeColor",
		"navbar.LinkClasses",
		"navbar.MenuItemClasses",
		"radio.BadgeClasses",
		"search.JSString",
		"select.ToOptions",
		"table.ActionButton",
		"table.StatusBadge",
		"table.TableHead",
		"table.TableBody",
		"table.TablePagination",
		"table.ActionButtonClasses",
		"table.StatusBadgeClasses",
		"table.ColumnCellClasses",
		"table.ColumnHeaderClasses",
		"table.BadgeCellClasses",
		"tabs.ActiveClasses",
		"tabs.InactiveClasses",
		"tabs.BadgeActiveClasses",
		"tabs.BadgeInactiveClasses",
		"palette.DefaultHues",
		"palette.DefaultShades",
		"rating.DefaultEmojiOptions",
	}
	for _, symbol := range removed {
		if !strings.Contains(guide, "`"+symbol+"`") {
			t.Errorf("migration guide missing removed symbol %q", symbol)
		}
	}
}

func TestMigrationGuideIncludesMechanicalSearches(t *testing.T) {
	guide := readDoc(t, "MIGRATING_COMPONENT_API.md")
	for _, search := range []string{
		"rg -n",
		"\\.Variant\\b",
		"button|link|kbd|tooltip",
		"AccordionItemData",
		"ActionButton|StatusBadge|TableHead|TableBody|TablePagination",
		"Body|OptionsList|ClientScript|ProviderError|BodyOOB|TriggerLabelOOB",
	} {
		if !strings.Contains(guide, search) {
			t.Errorf("migration guide missing mechanical search fragment %q", search)
		}
	}
}

func TestMigrationGuideDocumentsEveryRemovedPublicMethod(t *testing.T) {
	guide := readDoc(t, "MIGRATING_COMPONENT_API.md")
	for _, symbol := range strings.Fields(removedPublicMethodsV0011) {
		if !strings.Contains(guide, "`"+symbol+"`") {
			t.Errorf("migration guide missing exact removed public method %q", symbol)
		}
	}
}

func TestMigrationGuideDocumentsEveryRemovedPublicNonMethodSymbol(t *testing.T) {
	guide := readDoc(t, "MIGRATING_COMPONENT_API.md")
	for _, symbol := range strings.Fields(removedPublicNonMethodSymbolsV0011) {
		if !strings.Contains(guide, "`"+symbol+"`") {
			t.Errorf("migration guide missing exact removed public symbol %q", symbol)
		}
	}
}

func TestMigrationGuideMethodSearchFindsEveryRemovedMethodName(t *testing.T) {
	guide := readDoc(t, "MIGRATING_COMPONENT_API.md")
	checklistStart := strings.Index(guide, "## Mechanical upgrade checklist")
	if checklistStart < 0 {
		t.Fatal("migration guide missing mechanical upgrade checklist")
	}
	checklist := guide[checklistStart:]

	uniqueNames := make(map[string]struct{})
	for _, symbol := range strings.Fields(removedPublicMethodsV0011) {
		name := symbol[strings.LastIndex(symbol, ".")+1:]
		uniqueNames[name] = struct{}{}
	}
	names := make([]string, 0, len(uniqueNames))
	for name := range uniqueNames {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if !strings.Contains(checklist, name) {
			t.Errorf("mechanical checklist does not search for removed method name %q", name)
		}
	}
}

func TestCurrentMigrationGuideSnippetsCompileExternally(t *testing.T) {
	guide := readDoc(t, "MIGRATING_COMPONENT_API.md")
	snippets := extractCurrentGoSnippets(t, guide)
	requiredLabels := []string{
		"button",
		"dimensions",
		"kbd",
		"kind",
		"link",
		"modal",
		"tooltip",
	}
	labels := make([]string, 0, len(snippets))
	for label := range snippets {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	if strings.Join(labels, ",") != strings.Join(requiredLabels, ",") {
		t.Fatalf("compile-current snippet labels = %v, want %v", labels, requiredLabels)
	}

	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	tempModule := t.TempDir()
	goMod := fmt.Sprintf(`module example.com/goshtoso-migration-guide

go 1.26.5

require github.com/araihu/goshtoso v0.0.0

replace github.com/araihu/goshtoso => %s
`, filepath.ToSlash(root))
	if err := os.WriteFile(filepath.Join(tempModule, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}

	var source strings.Builder
	source.WriteString(`package migrationguide

import (
	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/kbd"
	"github.com/araihu/goshtoso/components/link"
	"github.com/araihu/goshtoso/components/modal"
	"github.com/araihu/goshtoso/components/pagination"
	"github.com/araihu/goshtoso/components/tooltip"
)

func compileCurrentGuideSnippets() {
	_ = components.KindBadge
	_ = badge.Badge
	_ = button.Button
	_ = kbd.Kbd
	_ = link.Link
	_ = modal.Modal
	_ = pagination.Pagination
	_ = tooltip.Tooltip
`)
	for _, label := range requiredLabels {
		source.WriteString("\n// " + label + "\n")
		source.WriteString(snippets[label])
		source.WriteByte('\n')
	}
	source.WriteString("}\n")
	if err := os.WriteFile(filepath.Join(tempModule, "migration_guide_test.go"), []byte(source.String()), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(
		"go", "test", "-mod=mod", "-c",
		"-o", filepath.Join(tempModule, "migration-guide.test"),
		".",
	)
	command.Dir = tempModule
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("compile current fenced Go snippets: %v\n%s", err, output)
	}
}

func extractCurrentGoSnippets(t *testing.T, guide string) map[string]string {
	t.Helper()
	lines := strings.Split(guide, "\n")
	snippets := make(map[string]string)
	const markerPrefix = "<!-- compile-current: "

	for index, line := range lines {
		if !strings.HasPrefix(line, markerPrefix) {
			continue
		}
		if !strings.HasSuffix(line, " -->") {
			t.Fatalf("malformed compile-current marker on line %d", index+1)
		}
		label := strings.TrimSuffix(strings.TrimPrefix(line, markerPrefix), " -->")
		if label == "" {
			t.Fatalf("empty compile-current label on line %d", index+1)
		}
		if _, duplicate := snippets[label]; duplicate {
			t.Fatalf("duplicate compile-current label %q", label)
		}
		if index+1 >= len(lines) || lines[index+1] != "```go" {
			t.Fatalf("compile-current marker %q must be immediately followed by a Go fence", label)
		}

		var snippet strings.Builder
		closed := false
		for cursor := index + 2; cursor < len(lines); cursor++ {
			if lines[cursor] == "```" {
				closed = true
				break
			}
			snippet.WriteString(lines[cursor])
			snippet.WriteByte('\n')
		}
		if !closed {
			t.Fatalf("compile-current Go fence %q is not closed", label)
		}
		snippets[label] = snippet.String()
	}

	return snippets
}

func lineContainsBoth(content, left, right string) bool {
	for line := range strings.Lines(content) {
		if strings.Contains(line, left) && strings.Contains(line, right) {
			return true
		}
	}
	return false
}

const removedPublicMethodsV0011 = `
accordion.AccordionConfig.ContainerClasses
accordion.AccordionItemData.CollapsedClasses
accordion.AccordionItemData.ContentClasses
accordion.AccordionItemData.ExpandedClasses
accordion.AccordionItemData.ItemButtonClasses
accordion.AccordionItemData.ItemContainerClasses
alert.Config.ContainerClasses
alert.Config.IconBadgeClasses
alert.Config.InnerClasses
alert.Config.LinkClasses
alert.Config.ListClasses
alert.Config.PrimaryActionClasses
alert.Config.TitleClasses
avatar.Config.BorderClasses
avatar.Config.HasImage
avatar.Config.HasInitials
avatar.Config.RadiusClasses
avatar.Config.ResolvedInitials
avatar.Config.ShapeClasses
avatar.Config.SizeClasses
avatar.Config.SpinnerSizeClasses
avatar.Config.StatusClasses
avatar.Config.StatusSizeClasses
avatar.Config.VariantClasses
avatar.Config.VariantFillClasses
badge.Config.IndicatorClasses
badge.Config.IsSoft
badge.Config.SizeClasses
badge.Config.SizeTextClass
badge.Config.SoftInnerClasses
badge.Config.SoftVariantClasses
badge.Config.VariantClasses
banner.Config.CTAClasses
banner.Config.ContainerClasses
banner.Config.CookieContainerClasses
banner.Config.LinkClasses
banner.Config.TextClasses
card.Config.ContainerClasses
card.Config.ContentClasses
card.Config.DescriptionClasses
card.Config.HasImage
card.Config.HasPrice
card.Config.HasRating
card.Config.ImageClasses
card.Config.ImageContainerClasses
card.Config.TagClasses
card.Config.TitleClasses
chatbubble.Config.BubbleClasses
chatbubble.Config.DataMine
chatbubble.Config.HasAvatar
chatbubble.Config.HasHeader
chatbubble.Config.IsMine
chatbubble.Config.RowClasses
checkbox.Config.InputClasses
checkbox.Config.SvgClasses
codeblock.Config.GetID
codeblock.Config.GetLabel
combobox.Config.DepsSelector
combobox.Config.HXIncludeSelector
combobox.Config.IsClientMode
combobox.State.IsSelected
drawer.Config.EnterEnd
drawer.Config.EnterStart
drawer.Config.GetBodyID
drawer.Config.OverlayClasses
drawer.Config.PanelClasses
drawer.Config.StateVar
drawer.Config.TitleID
dropdown.Config.ButtonClasses
dropdown.Config.DangerClasses
dropdown.Config.DisabledClasses
dropdown.Config.GetTriggerMode
dropdown.Config.HasDividers
dropdown.Config.HasIcons
dropdown.Config.HasShortcuts
dropdown.Config.IsContextMenu
dropdown.Config.ItemClasses
dropdown.Config.MenuClasses
dropdown.Config.UseIconOnlyTrigger
dropdown.Item.IsButton
fileinput.Config.BrowseLabelClasses
fileinput.Config.ContainerClasses
fileinput.Config.DropZoneClasses
fileinput.Config.HelperTextClasses
fileinput.Config.IsUpload
fileinput.Config.LabelClasses
fileinput.Config.UploadButtonClasses
fileinput.Config.UploadContainerClasses
fileinput.Config.UploadControlClasses
fileinput.Config.UploadFileNameClasses
form.FormErrorsConfig.GetID
form.FormErrorsConfig.GetTitle
kbd.Config.AccessibleLabel
kbd.Config.IconClasses
kbd.Config.RootClasses
kbd.Config.SizeClasses
modal.Config.AlertCTAClasses
modal.Config.DialogClasses
modal.Config.HeaderClasses
modal.Config.IconBadgeClasses
modal.Config.StateVar
modal.Config.TitleID
modal.Config.TriggerClasses
navbar.Config.LeftActions
navbar.Config.NavClasses
navbar.Config.RightActions
pagination.Config.EllipsisClasses
pagination.Config.ListClasses
pagination.Config.NavClasses
pagination.Config.PageClasses
pagination.Config.PrevNextClasses
pagination.Config.SwapStrategy
palette.Config.ContainerClasses
radio.Config.HasAlpine
radio.Config.HasHTMX
radio.Config.InputClasses
radio.Config.SegmentedLabelClasses
radio.HTMXConfig.EffectiveTrigger
radio.HTMXConfig.HasHxVerb
range.Config.AlpineData
range.Config.ControlClasses
range.Config.IconClasses
range.Config.InputClasses
range.Config.LabelClasses
range.Config.MaxOrDefault
range.Config.MinOrDefault
range.Config.RootClasses
range.Config.StepOrDefault
range.Config.TickClasses
range.Config.TickLabels
range.Config.ValueClasses
range.Config.ValueOrDefault
rating.Config.ActiveIconClasses
rating.Config.BindClass
rating.Config.ControlClasses
rating.Config.EmojiIcon
rating.Config.IconClasses
rating.Config.InactiveIconClasses
rating.Config.InputID
rating.Config.IsActive
rating.Config.ResolvedID
rating.Config.ResolvedLabel
rating.Config.ResolvedMax
rating.Config.ResolvedName
rating.Config.ResolvedValue
rating.Config.RootClasses
rating.Config.ValueLabel
rating.Config.XData
search.Config.DialogClasses
search.Config.GetDescriptionMaxLength
search.Config.GetEmptyText
search.Config.GetEscapeText
search.Config.GetID
search.Config.GetLabel
search.Config.GetMaxResults
search.Config.GetPlaceholder
search.Config.GetShortcutText
search.Config.RootClasses
search.Config.TriggerClasses
select.Config.ContainerClasses
select.Config.GetPlaceholder
select.Config.IsEffectivelyDisabled
select.Config.LabelClasses
select.Config.SelectClasses
select.Config.SelectedValue
select.Config.TriggerClasses
sidebar.Config.ContainerClasses
sidebar.Config.NavClasses
sidebar.OverlayConfig.BackdropClasses
sidebar.OverlayConfig.PanelClasses
sidebar.OverlayConfig.PanelID
sidebar.OverlayConfig.RootClasses
sidebar.OverlayConfig.StateVar
sidebar.OverlayConfig.TriggerClasses
sidebar.OverlayConfig.TriggerLabelText
spinner.Config.FillClasses
spinner.Config.SizeClasses
structuredinput.Column.DefaultValue
structuredinput.Column.EntryAccessor
structuredinput.Column.NameBinding
structuredinput.Config.ContainerClasses
structuredinput.Config.EntriesJSON
structuredinput.Config.GetAddLabel
structuredinput.Config.NewRowJSON
structuredinput.Config.NormalizedColumns
structuredinput.Option.OptionLabel
table.Config.CellClasses
table.Config.CheckboxClasses
table.Config.ColCount
table.Config.ContainerClasses
table.Config.FilterBarID
table.Config.GetID
table.Config.HTMXEndpointValue
table.Config.HTMXTargetValue
table.Config.HasActionableRows
table.Config.HasActions
table.Config.HasExpandableRows
table.Config.HasFilters
table.Config.HasLinkedRows
table.Config.HasSortableColumns
table.Config.HeaderCellClasses
table.Config.LazyLoadTrigger
table.Config.PaginationBaseURL
table.Config.PaginationID
table.Config.RowClasses
table.Config.SortableHeaderClasses
table.Config.TableClasses
table.Config.TbodyClasses
table.Config.TbodyID
table.Config.TheadClasses
table.Config.TheadID
table.FilterConfig.ResolvedHxSwap
table.FilterConfig.ResolvedHxTarget
table.PaginationConfig.GetContainerHeight
table.PaginationConfig.IsContained
table.PaginationConfig.IsInfiniteScroll
table.PaginationConfig.NextPage
table.PaginationConfig.PaginationPages
table.Row.ClickableRole
table.Row.HasHTMXAction
table.Row.IsActionable
tagslist.Config.AlpineData
tagslist.Config.ContainerClasses
tagslist.Config.GetAddLabel
tagslist.Config.GetPlaceholder
textarea.Config.ContainerClasses
textarea.Config.GetRows
textarea.Config.HelperTextClasses
textarea.Config.LabelClasses
textarea.Config.TextareaClasses
textinput.Config.ContainerClasses
textinput.Config.GetType
textinput.Config.HasMask
textinput.Config.HasMaxLength
textinput.Config.HasPattern
textinput.Config.HelperTextClasses
textinput.Config.InputClasses
textinput.Config.IsPassword
textinput.Config.IsSearch
textinput.Config.LabelClasses
textinput.Config.MaxLengthStr
toast.Config.BgClass
toast.Config.BorderClass
toast.Config.HasAction
toast.Config.IconBgClass
toast.Config.TitleClass
toggle.Config.LabelClasses
toggle.Config.ToggleClasses
`

const removedPublicNonMethodSymbolsV0011 = `
accordion.AccordionConfig.Variant
accordion.AccordionItemData
accordion.AccordionItemData.AllowMultiple
accordion.AccordionItemData.ContainerID
accordion.AccordionItemData.Index
accordion.AccordionItemData.Item
accordion.AccordionItemData.Variant
accordion.Default
accordion.NoBackground
accordion.SingleOpen
accordion.Split
accordion.Variant
alert.Config.Variant
alert.Danger
alert.Info
alert.Success
alert.Variant
alert.Warning
avatar.Config.Variant
avatar.Danger
avatar.Default
avatar.Info
avatar.Inverse
avatar.Primary
avatar.Secondary
avatar.Success
avatar.UserIcon
avatar.Variant
avatar.Warning
badge.Config.Style
badge.Config.Variant
badge.Danger
badge.Default
badge.Info
badge.Inverse
badge.Primary
badge.Secondary
badge.Style
badge.StyleSoft
badge.StyleSolid
badge.Success
badge.Variant
badge.Warning
banner.Config.CookieBanner
banner.Config.CookieConfig
banner.Config.Variant
banner.Danger
banner.Default
banner.Info
banner.Primary
banner.Success
banner.Variant
banner.Warning
button.Alternate
button.Config
button.Config.Alpine
button.Config.Disabled
button.Config.HTMX
button.Config.ID
button.Config.LoadingText
button.Config.RootClass
button.Config.Size
button.Config.Type
button.Config.Variant
button.Danger
button.Info
button.Inverse
button.Primary
button.Secondary
button.Success
button.Variant
button.Warning
card.Config.Price
card.Config.Rating
card.Config.Variant
card.Default
card.Primary
card.StarRating
card.Variant
carousel.Config.Variant
carousel.Default
carousel.OnCard
carousel.Variant
carousel.WithCTA
carousel.WithText
chatbubble.Config.AvatarVariant
checkbox.Config.Variant
checkbox.Danger
checkbox.Info
checkbox.Primary
checkbox.Secondary
checkbox.Success
checkbox.Variant
checkbox.Warning
combobox.Body
combobox.BodyOOB
combobox.ClientScript
combobox.Option.Badge
combobox.Option.BadgeColor
combobox.Option.Img
combobox.Option.Initials
combobox.Option.Meta
combobox.OptionsList
combobox.ProviderError
combobox.TriggerLabelOOB
fileinput.Config.Variant
fileinput.Variant
fileinput.VariantDropZone
fileinput.VariantUpload
kbd.Config
kbd.Config.Attrs
kbd.Config.Class
kbd.Config.Icon
kbd.Config.Label
kbd.Config.Size
kbd.Config.Text
link.Config
link.Config.Attrs
link.Config.Class
link.Config.Href
link.Config.ID
link.Config.Icon
link.Config.IconPosition
link.Config.Rel
link.Config.Role
link.Config.Size
link.Config.Style
link.Config.Target
link.Style
link.StyleButton
link.StyleText
modal.Config.AlertMode
modal.Config.Variant
modal.Danger
modal.Default
modal.Info
modal.Success
modal.Variant
modal.Warning
navbar.LinkClasses
navbar.MenuItemClasses
pagination.Config.Variant
pagination.Simple
pagination.Variant
pagination.WithEllipsis
palette.DefaultHues
palette.DefaultShades
radio.BadgeClasses
radio.Config.Variant
radio.Danger
radio.Info
radio.Primary
radio.Secondary
radio.Success
radio.Variant
radio.Warning
rating.Config.Attrs
rating.Config.Class
rating.Config.ReadOnly
rating.Config.Style
rating.DefaultEmojiOptions
rating.EmojiOption
rating.EmojiOption.Icon
rating.EmojiOption.Label
rating.EmojiOption.Value
rating.Style
rating.StyleEmoji
rating.StyleStars
search.JSString
select.ToOptions
spinner.Config.Variant
spinner.Danger
spinner.Default
spinner.Info
spinner.Primary
spinner.Secondary
spinner.Success
spinner.Variant
spinner.Warning
table.ActionButton
table.ActionButtonClasses
table.BadgeCellClasses
table.ColumnCellClasses
table.ColumnHeaderClasses
table.Config.Variant
table.Default
table.FilterConfig.Variant
table.FilterVariant
table.FilterVariantBar
table.FilterVariantInline
table.StatusBadge
table.StatusBadgeClasses
table.Striped
table.TableBody
table.TableHead
table.TablePagination
table.Variant
table.WithCheckbox
tabs.ActiveClasses
tabs.BadgeActiveClasses
tabs.BadgeInactiveClasses
tabs.InactiveClasses
toast.Config.Sender
toast.Config.Variant
toast.Container
toast.Danger
toast.Info
toast.Message
toast.Success
toast.Variant
toast.Warning
toggle.Config.Style
toggle.Config.Variant
toggle.Danger
toggle.Info
toggle.Primary
toggle.Secondary
toggle.Style
toggle.StyleContainer
toggle.StyleDefault
toggle.Success
toggle.Variant
toggle.Warning
tooltip.Bottom
tooltip.Click
tooltip.Config
tooltip.Config.Description
tooltip.Config.ID
tooltip.Config.Label
tooltip.Config.Position
tooltip.Config.Trigger
tooltip.Config.TriggerLabel
tooltip.Config.TriggerMode
tooltip.Hover
tooltip.Left
tooltip.Right
tooltip.Top
tooltip.Trigger
`

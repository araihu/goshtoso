package docs_test

import (
	"fmt"
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

func lineContainsBoth(content, left, right string) bool {
	for line := range strings.Lines(content) {
		if strings.Contains(line, left) && strings.Contains(line, right) {
			return true
		}
	}
	return false
}

// Package actiongroup provides responsive groups of primary, secondary, and
// stacked actions.
package actiongroup

import (
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/dropdown"
)

// Action describes one action or one stacked group of actions.
//
// Set Items to render a labeled flat Dropdown at normal widths. When the
// ActionGroup is constrained, those child actions are flattened into a
// dedicated section of the shared overflow Dropdown. Deeper groups are also
// flattened with disabled context labels; Goshtoso Dropdown menus remain flat.
type Action struct {
	// Label is the visible action label or grouped-action trigger label.
	Label string
	// Href renders a native link when OnClick and HTMX are empty and Disabled is false.
	Href string
	// Icon renders before Label.
	Icon templ.Component
	// OnClick is an Alpine.js expression for button actions.
	OnClick string
	// HTMX configures a declarative server action. It may be combined with
	// OnClick. Href is ignored when HTMX is set.
	HTMX *dropdown.HTMXConfig
	// Disabled renders an inert native button.
	Disabled bool
	// Danger applies the destructive action treatment to buttons and Dropdown
	// items. Links retain Link's navigation treatment.
	Danger bool
	// Tooltip sets a native title attribute.
	Tooltip string
	// ID sets the native action or grouped Dropdown root ID. Flattened overflow
	// copies append "-overflow" so the rendered document keeps IDs unique.
	ID string
	// Items turns this action into a stacked group backed by Dropdown.
	Items []Action
}

// Config holds ActionGroup actions, accessible labeling, and root hooks.
type Config struct {
	// Label is the ActionGroup's accessible name. Default is "Actions".
	Label string
	// Primary is required, always rendered, and never moved into overflow.
	Primary Action
	// Secondary is ordered from highest to lowest priority. Lower-priority
	// actions at the end move into overflow first.
	Secondary []Action
	// OverflowLabel is the accessible label for the ellipsis trigger.
	// Default is "More actions".
	OverflowLabel string
	// RootClass appends CSS classes to the ActionGroup root.
	RootClass string
	// RootAttrs appends arbitrary HTML attributes to the ActionGroup root.
	RootAttrs templ.Attributes
}

func (cfg Config) label() string {
	if label := strings.TrimSpace(cfg.Label); label != "" {
		return label
	}
	return "Actions"
}

func (cfg Config) overflowLabel() string {
	if label := strings.TrimSpace(cfg.OverflowLabel); label != "" {
		return label
	}
	return "More actions"
}

func (cfg Config) rootClasses() string {
	base := "flex w-full min-w-0 flex-wrap items-center justify-end gap-2"
	if extra := strings.TrimSpace(cfg.RootClass); extra != "" {
		return base + " " + extra
	}
	return base
}

func (action Action) isGroup() bool {
	return len(action.Items) > 0
}

func (action Action) buttonTone(primary bool) button.Tone {
	if action.Danger {
		return button.ToneDanger
	}
	if primary {
		return button.TonePrimary
	}
	return button.ToneAlternate
}

func (action Action) dropdownItem(overflow bool) dropdown.Item {
	id := action.ID
	if overflow && id != "" {
		id += "-overflow"
	}
	return dropdown.Item{
		Label:    action.Label,
		Href:     action.Href,
		Icon:     action.Icon,
		OnClick:  action.OnClick,
		HTMX:     action.HTMX,
		Disabled: action.Disabled,
		Danger:   action.Danger,
		Tooltip:  action.Tooltip,
		ID:       id,
	}
}

func (action Action) dropdownSections() []dropdown.Section {
	items := make([]dropdown.Item, 0, len(action.Items))
	items = appendFlattenedItems(items, action.Items, false)
	return []dropdown.Section{{Items: items}}
}

func overflowSections(actions []Action) ([]dropdown.Section, []int) {
	sections := make([]dropdown.Section, 0, len(actions))
	counts := make([]int, 0, len(actions))
	for _, action := range actions {
		if !action.isGroup() {
			sections = append(sections, dropdown.Section{
				Items: []dropdown.Item{action.dropdownItem(true)},
			})
			counts = append(counts, 1)
			continue
		}

		items := make([]dropdown.Item, 0, len(action.Items)+1)
		items = append(items, groupLabelItem(action))
		items = appendFlattenedItems(items, action.Items, true)
		sections = append(sections, dropdown.Section{Items: items})
		counts = append(counts, len(items))
	}
	return sections, counts
}

func appendFlattenedItems(items []dropdown.Item, actions []Action, overflow bool) []dropdown.Item {
	for _, action := range actions {
		if action.isGroup() {
			items = append(items, groupLabelItem(action))
			items = appendFlattenedItems(items, action.Items, overflow)
			continue
		}
		items = append(items, action.dropdownItem(overflow))
	}
	return items
}

func groupLabelItem(action Action) dropdown.Item {
	return dropdown.Item{
		Label:    action.Label,
		Disabled: true,
		Tooltip:  action.Tooltip,
	}
}

func overflowCountsValue(counts []int) string {
	if len(counts) == 0 {
		return ""
	}
	values := make([]string, len(counts))
	for index, count := range counts {
		values[index] = strconv.Itoa(count)
	}
	return strings.Join(values, ",")
}

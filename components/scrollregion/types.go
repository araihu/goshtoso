package scrollregion

import (
	"strings"

	"github.com/a-h/templ"
)

// Config describes a bounded, independently scrollable content region.
type Config struct {
	// Content is rendered inside the scroll viewport.
	Content templ.Component
	// RootClass appends classes to the positioning root.
	RootClass string
	// ViewportClass appends classes to the scroll viewport.
	ViewportClass string
	// DisableIndicators keeps the sentinels but omits the visual boundary cues.
	DisableIndicators bool
}

// AccessibleName describes the accessible name of a Scroll Region without
// changing Config's original four-field public shape. LabelledBy takes
// precedence over Label when both are set.
type AccessibleName struct {
	// Label sets the accessible name. It defaults to "Scrollable content" when
	// LabelledBy is not set.
	Label string
	// LabelledBy sets one or more existing element IDs that name the region.
	LabelledBy string
}

// ScrollRegion creates a bounded scroll group with automatic boundary cues.
// It keeps Config's original positional compatibility and uses the stable
// fallback accessible name "Scrollable content". The compatibility
// constructor intentionally uses role=group: repeating unnamed legacy calls
// must not manufacture duplicate ARIA region landmarks. Use Named or Labelled
// when the content is an intentional, uniquely named landmark.
func ScrollRegion(cfg Config) templ.Component {
	return Instance{cfg: cfg, name: AccessibleName{}}
}

// Named creates a bounded scroll region with an explicit accessible name.
// Use an AccessibleName value rather than adding fields to Config so existing
// keyed and unkeyed Config literals remain source compatible.
func Named(cfg Config, name AccessibleName) templ.Component {
	return Instance{cfg: cfg, name: name, landmark: true}
}

// Labelled creates a Scroll Region named by existing visible label element IDs.
// GoshtosoSkillExample: @scrollregion.Labelled(scrollregion.Config{RootClass: "h-64", Content: activityRows()}, "activity-history-heading")
func Labelled(cfg Config, labelledBy string) templ.Component {
	return Named(cfg, AccessibleName{LabelledBy: labelledBy})
}

func (cfg Config) rootClasses() string {
	return joinClasses("relative min-h-0", cfg.RootClass)
}

func (cfg Config) viewportClasses() string {
	// Keep wide generic content reachable by keyboard, mouse, and touch. If a
	// consumer deliberately supplies an axis utility, omit the conflicting
	// default so Tailwind's generated utility order cannot silently reverse the
	// consumer contract (for example, overflow-x-hidden).
	horizontal := "overflow-x-auto"
	if hasViewportAxisClass(cfg.ViewportClass, "overflow-x-") {
		horizontal = ""
	}
	vertical := "overflow-y-auto"
	if hasViewportAxisClass(cfg.ViewportClass, "overflow-y-") {
		vertical = ""
	}
	return joinClasses("h-full", horizontal, vertical, "rounded-radius focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:focus-visible:outline-primary-dark", cfg.ViewportClass)
}

func hasViewportAxisClass(classes, prefix string) bool {
	for class := range strings.FieldsSeq(classes) {
		if strings.HasPrefix(class, prefix) {
			return true
		}
	}
	return false
}

func scrollRegionRole(landmark bool) string {
	if landmark {
		return "region"
	}
	return "group"
}

func (name AccessibleName) accessibleLabel() string {
	if label := strings.TrimSpace(name.Label); label != "" {
		return label
	}
	return "Scrollable content"
}

func (name AccessibleName) labelledBy() string {
	return strings.TrimSpace(name.LabelledBy)
}

func joinClasses(values ...string) string {
	classes := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			classes = append(classes, value)
		}
	}
	return strings.Join(classes, " ")
}

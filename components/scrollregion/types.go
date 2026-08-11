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

// ScrollRegion creates a bounded scroll region with automatic boundary cues.
func ScrollRegion(cfg Config) templ.Component {
	return Instance{cfg: cfg}
}

func (cfg Config) rootClasses() string {
	return joinClasses("relative min-h-0", cfg.RootClass)
}

func (cfg Config) viewportClasses() string {
	return joinClasses("h-full overflow-y-auto rounded-radius focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:focus-visible:outline-primary-dark", cfg.ViewportClass)
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

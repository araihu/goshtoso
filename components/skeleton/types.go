package skeleton

import (
	"strings"

	"github.com/a-h/templ"
)

// Shape controls the geometry of each skeleton placeholder.
type Shape string

const (
	// ShapeText renders line placeholders and is the default.
	ShapeText Shape = ""
	// ShapeRectangle renders rectangular content placeholders.
	ShapeRectangle Shape = "rectangle"
	// ShapeCircle renders avatar-sized circular placeholders.
	ShapeCircle Shape = "circle"
)

// Config holds skeleton geometry, accessible labeling, animation, and target-specific hooks.
type Config struct {
	// Shape selects text, rectangle, or circle placeholders; default is text.
	Shape Shape
	// Count controls placeholder count; default is three for text and one otherwise.
	Count int
	// Label describes the loading state for assistive technology; default is "Loading content".
	Label string
	// Static disables pulse animation while preserving the loading semantics.
	Static bool
	// RootClass appends CSS classes to the skeleton root.
	RootClass string
	// RootAttrs appends arbitrary HTML attributes to the skeleton root.
	RootAttrs templ.Attributes
	// ItemClass appends CSS classes to every placeholder item.
	ItemClass string
	// ItemAttrs appends arbitrary HTML attributes to every placeholder item.
	ItemAttrs templ.Attributes
}

func (cfg Config) shape() Shape {
	switch cfg.Shape {
	case ShapeRectangle, ShapeCircle:
		return cfg.Shape
	default:
		return ShapeText
	}
}

func (cfg Config) shapeName() string {
	if cfg.shape() == ShapeText {
		return "text"
	}
	return string(cfg.shape())
}

func (cfg Config) count() int {
	if cfg.Count > 0 {
		return cfg.Count
	}
	if cfg.shape() == ShapeText {
		return 3
	}
	return 1
}

func (cfg Config) label() string {
	if label := strings.TrimSpace(cfg.Label); label != "" {
		return label
	}
	return "Loading content"
}

func (cfg Config) rootClasses() string {
	base := "w-full"
	if cfg.shape() == ShapeCircle {
		base += " flex flex-wrap gap-3"
	} else {
		base += " space-y-3"
	}
	return appendClass(base, cfg.RootClass)
}

func (cfg Config) itemClasses(index int) string {
	base := "block bg-on-surface/10 dark:bg-on-surface-dark/10"
	switch cfg.shape() {
	case ShapeRectangle:
		base += " h-32 w-full rounded-radius"
	case ShapeCircle:
		base += " size-12 rounded-full"
	default:
		base += " h-3 rounded-radius"
		switch {
		case cfg.count() == 1:
			base += " w-3/4"
		case index == cfg.count()-1:
			base += " w-2/3"
		default:
			base += " w-full"
		}
	}
	if !cfg.Static {
		base += " animate-pulse motion-reduce:animate-none"
	}
	return appendClass(base, cfg.ItemClass)
}

func appendClass(base, extra string) string {
	if extra = strings.TrimSpace(extra); extra != "" {
		return base + " " + extra
	}
	return base
}

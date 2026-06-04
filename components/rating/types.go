// Package rating renders PenguinUI-styled star and emoji rating controls.
package rating

import (
	"fmt"
	"strconv"

	"github.com/a-h/templ"
)

// Style represents the rating visual style.
type Style string

const (
	StyleStars Style = "stars"
	StyleEmoji Style = "emoji"
)

// Size represents icon sizes.
type Size string

const (
	SizeSM Size = "sm"
	SizeMD Size = "md" // default
	SizeLG Size = "lg"
	SizeXL Size = "xl"
)

// EmojiOption describes one emoji rating step.
type EmojiOption struct {
	Value int
	Label string
	Icon  string
}

// Config holds configuration for a rating control.
type Config struct {
	// ID prefixes the generated input IDs. Defaults to Name, then "rating".
	ID string
	// Name is the radio group field name.
	Name string
	// Value is the initially selected rating. Values outside 0..Max are clamped.
	Value int
	// Max is the number of rating options. Defaults to 5.
	Max int
	// Label gives the group an accessible label and visible label when ShowLabel is true.
	Label string
	// ShowLabel renders Label visibly above the control.
	ShowLabel bool
	// Style switches between star and emoji visuals. Defaults to StyleStars.
	Style Style
	// Size sets icon size. Defaults to SizeMD.
	Size Size
	// Disabled disables all radio inputs.
	Disabled bool
	// ReadOnly renders a non-interactive meter instead of radios.
	ReadOnly bool
	// Class is appended to the root element.
	Class string
	// Attrs is an escape hatch applied last to the root element.
	Attrs templ.Attributes
}

// DefaultEmojiOptions are the five standard sentiment choices.
var DefaultEmojiOptions = []EmojiOption{
	{Value: 1, Label: "very dissatisfied", Icon: "😠"},
	{Value: 2, Label: "dissatisfied", Icon: "🙁"},
	{Value: 3, Label: "neutral", Icon: "😐"},
	{Value: 4, Label: "satisfied", Icon: "🙂"},
	{Value: 5, Label: "very satisfied", Icon: "😍"},
}

// ResolvedID returns the root ID prefix.
func (cfg Config) ResolvedID() string {
	if cfg.ID != "" {
		return cfg.ID
	}
	if cfg.Name != "" {
		return cfg.Name
	}
	return "rating"
}

// ResolvedName returns the radio group name.
func (cfg Config) ResolvedName() string {
	if cfg.Name != "" {
		return cfg.Name
	}
	return cfg.ResolvedID()
}

// ResolvedMax returns the number of rating options.
func (cfg Config) ResolvedMax() int {
	if cfg.Max > 0 {
		return cfg.Max
	}
	return 5
}

// ResolvedValue returns Value clamped to 0..Max.
func (cfg Config) ResolvedValue() int {
	return min(max(cfg.Value, 0), cfg.ResolvedMax())
}

// ResolvedLabel returns the accessible group label.
func (cfg Config) ResolvedLabel() string {
	if cfg.Label != "" {
		return cfg.Label
	}
	return "Rating"
}

// RootClasses returns the root classes.
func (cfg Config) RootClasses() string {
	base := "inline-flex flex-col gap-2"
	if cfg.Class != "" {
		base += " " + cfg.Class
	}
	return base
}

// ControlClasses returns classes for the option row.
func (cfg Config) ControlClasses() string {
	base := "inline-flex w-fit items-center gap-1 rounded-radius"
	if !cfg.ReadOnly {
		base += " focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-outline-strong dark:focus-within:outline-outline-dark-strong"
	}
	return base
}

// IconClasses returns base classes for each visual option.
func (cfg Config) IconClasses() string {
	base := "block transition"
	switch cfg.Size {
	case SizeSM:
		base += " size-5 text-lg"
	case SizeLG:
		base += " size-8 text-3xl"
	case SizeXL:
		base += " size-10 text-4xl"
	default:
		base += " size-6 text-2xl"
	}
	if cfg.Disabled {
		base += " opacity-60"
	}
	return base
}

// ActiveIconClasses returns classes for selected values.
func (cfg Config) ActiveIconClasses() string {
	if cfg.Style == StyleEmoji {
		return "scale-110 opacity-100 grayscale-0"
	}
	return "text-warning"
}

// InactiveIconClasses returns classes for unselected values.
func (cfg Config) InactiveIconClasses() string {
	if cfg.Style == StyleEmoji {
		return "opacity-45 grayscale"
	}
	return "text-on-surface-muted dark:text-on-surface-dark-muted"
}

// BindClass returns an Alpine class binding for the given value.
func (cfg Config) BindClass(value int) string {
	if cfg.Style == StyleEmoji {
		return fmt.Sprintf("currentVal === %d ? '%s' : '%s'", value, cfg.ActiveIconClasses(), cfg.InactiveIconClasses())
	}
	return fmt.Sprintf("currentVal >= %d ? '%s' : '%s'", value, cfg.ActiveIconClasses(), cfg.InactiveIconClasses())
}

// IsActive reports whether a value should render active on first paint.
func (cfg Config) IsActive(value int) bool {
	if cfg.Style == StyleEmoji {
		return cfg.ResolvedValue() == value
	}
	return value <= cfg.ResolvedValue()
}

// XData returns the simple Alpine state object.
func (cfg Config) XData() string {
	return fmt.Sprintf("{ currentVal: %d }", cfg.ResolvedValue())
}

// InputID returns the generated input ID for a value.
func (cfg Config) InputID(value int) string {
	return cfg.ResolvedID() + "-" + strconv.Itoa(value)
}

// ValueLabel returns the accessible label for a numeric rating.
func (cfg Config) ValueLabel(value int) string {
	if cfg.Style == StyleEmoji {
		for _, opt := range DefaultEmojiOptions {
			if opt.Value == value {
				return opt.Label
			}
		}
	}
	if value == 1 {
		return "one star"
	}
	return fmt.Sprintf("%d stars", value)
}

// EmojiIcon returns the configured emoji icon for a value.
func (cfg Config) EmojiIcon(value int) string {
	for _, opt := range DefaultEmojiOptions {
		if opt.Value == value {
			return opt.Icon
		}
	}
	return "🙂"
}

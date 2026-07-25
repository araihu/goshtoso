// Package rating renders PenguinUI-styled star and emoji rating controls.
package rating

import (
	"fmt"
	"strconv"

	"github.com/a-h/templ"
)

// Appearance represents the rating visual treatment.
type Appearance string

const (
	AppearanceStars Appearance = "stars"
	AppearanceEmoji Appearance = "emoji"
)

// Size represents icon sizes.
type Size string

const (
	SizeSM Size = "sm"
	SizeMD Size = "md" // default
	SizeLG Size = "lg"
	SizeXL Size = "xl"
)

type emojiOption struct {
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
	// Appearance switches between star and emoji visuals. Defaults to AppearanceStars.
	Appearance Appearance
	// Size sets icon size. Defaults to SizeMD.
	Size Size
	// Disabled disables all radio inputs.
	Disabled bool
	// RootClass is appended to the root element.
	RootClass string
	// RootAttrs is an escape hatch applied last to the root element.
	RootAttrs templ.Attributes
}

// DisplayConfig holds configuration for a non-interactive rating display.
type DisplayConfig struct {
	// ID identifies the root element. Defaults to "rating".
	ID string
	// Value is the displayed rating. Values outside 0..Max are clamped.
	Value int
	// Max is the number of displayed rating options. Defaults to 5.
	Max int
	// Label is the accessible label and visible label when ShowLabel is true.
	Label string
	// ShowLabel renders Label visibly above the display.
	ShowLabel bool
	// Appearance switches between star and emoji visuals. Defaults to AppearanceStars.
	Appearance Appearance
	// Size sets icon size. Defaults to SizeMD.
	Size Size
	// RootClass is appended to the root element.
	RootClass string
	// RootAttrs is an escape hatch applied last to the root element.
	RootAttrs templ.Attributes
}

var defaultEmojiOptions = []emojiOption{
	{Value: 1, Label: "very dissatisfied", Icon: "😠"},
	{Value: 2, Label: "dissatisfied", Icon: "🙁"},
	{Value: 3, Label: "neutral", Icon: "😐"},
	{Value: 4, Label: "satisfied", Icon: "🙂"},
	{Value: 5, Label: "very satisfied", Icon: "😍"},
}

// ResolvedID returns the root ID prefix.
func (cfg Config) resolvedID() string {
	if cfg.ID != "" {
		return cfg.ID
	}
	if cfg.Name != "" {
		return cfg.Name
	}
	return "rating"
}

// ResolvedName returns the radio group name.
func (cfg Config) resolvedName() string {
	if cfg.Name != "" {
		return cfg.Name
	}
	return cfg.resolvedID()
}

// ResolvedMax returns the number of rating options.
func (cfg Config) resolvedMax() int {
	if cfg.Max > 0 {
		return cfg.Max
	}
	return 5
}

// ResolvedValue returns Value clamped to 0..Max.
func (cfg Config) resolvedValue() int {
	return min(max(cfg.Value, 0), cfg.resolvedMax())
}

// ResolvedLabel returns the accessible group label.
func (cfg Config) resolvedLabel() string {
	if cfg.Label != "" {
		return cfg.Label
	}
	return "Rating"
}

// RootClasses returns the root classes.
func (cfg Config) rootClasses() string {
	base := "inline-flex flex-col gap-2"
	if cfg.RootClass != "" {
		base += " " + cfg.RootClass
	}
	return base
}

// ControlClasses returns classes for the option row.
func (cfg Config) controlClasses() string {
	return "inline-flex w-fit items-center gap-1 rounded-radius focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-outline-strong dark:focus-within:outline-outline-dark-strong"
}

// IconClasses returns base classes for each visual option.
func (cfg Config) iconClasses() string {
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
func (cfg Config) activeIconClasses() string {
	if cfg.Appearance == AppearanceEmoji {
		return "scale-110 opacity-100 grayscale-0"
	}
	return "text-warning"
}

// InactiveIconClasses returns classes for unselected values.
func (cfg Config) inactiveIconClasses() string {
	if cfg.Appearance == AppearanceEmoji {
		return "opacity-45 grayscale"
	}
	return "text-on-surface-muted dark:text-on-surface-dark-muted"
}

// BindClass returns an Alpine class binding for the given value.
func (cfg Config) bindClass(value int) string {
	if cfg.Appearance == AppearanceEmoji {
		return fmt.Sprintf("currentVal === %d ? '%s' : '%s'", value, cfg.activeIconClasses(), cfg.inactiveIconClasses())
	}
	return fmt.Sprintf("currentVal >= %d ? '%s' : '%s'", value, cfg.activeIconClasses(), cfg.inactiveIconClasses())
}

// IsActive reports whether a value should render active on first paint.
func (cfg Config) isActive(value int) bool {
	if cfg.Appearance == AppearanceEmoji {
		return cfg.resolvedValue() == value
	}
	return value <= cfg.resolvedValue()
}

// XData returns the simple Alpine state object.
func (cfg Config) xData() string {
	return fmt.Sprintf("{ currentVal: %d }", cfg.resolvedValue())
}

// InputID returns the generated input ID for a value.
func (cfg Config) inputID(value int) string {
	return cfg.resolvedID() + "-" + strconv.Itoa(value)
}

// ValueLabel returns the accessible label for a numeric rating.
func (cfg Config) valueLabel(value int) string {
	if cfg.Appearance == AppearanceEmoji {
		for _, opt := range defaultEmojiOptions {
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
func (cfg Config) emojiIcon(value int) string {
	for _, opt := range defaultEmojiOptions {
		if opt.Value == value {
			return opt.Icon
		}
	}
	return "🙂"
}

func (cfg DisplayConfig) resolvedID() string {
	if cfg.ID != "" {
		return cfg.ID
	}
	return "rating"
}

func (cfg DisplayConfig) resolvedMax() int {
	if cfg.Max > 0 {
		return cfg.Max
	}
	return 5
}

func (cfg DisplayConfig) resolvedValue() int {
	return min(max(cfg.Value, 0), cfg.resolvedMax())
}

func (cfg DisplayConfig) resolvedLabel() string {
	if cfg.Label != "" {
		return cfg.Label
	}
	return "Rating"
}

func (cfg DisplayConfig) rootClasses() string {
	base := "inline-flex flex-col gap-2"
	if cfg.RootClass != "" {
		base += " " + cfg.RootClass
	}
	return base
}

func (cfg DisplayConfig) controlClasses() string {
	return "inline-flex w-fit items-center gap-1 rounded-radius"
}

func (cfg DisplayConfig) iconClasses() string {
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
	return base
}

func (cfg DisplayConfig) activeIconClasses() string {
	if cfg.Appearance == AppearanceEmoji {
		return "scale-110 opacity-100 grayscale-0"
	}
	return "text-warning"
}

func (cfg DisplayConfig) inactiveIconClasses() string {
	if cfg.Appearance == AppearanceEmoji {
		return "opacity-45 grayscale"
	}
	return "text-on-surface-muted dark:text-on-surface-dark-muted"
}

func (cfg DisplayConfig) isActive(value int) bool {
	if cfg.Appearance == AppearanceEmoji {
		return cfg.resolvedValue() == value
	}
	return value <= cfg.resolvedValue()
}

func (cfg DisplayConfig) valueLabel(value int) string {
	if cfg.Appearance == AppearanceEmoji {
		for _, opt := range defaultEmojiOptions {
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

func (cfg DisplayConfig) emojiIcon(value int) string {
	for _, opt := range defaultEmojiOptions {
		if opt.Value == value {
			return opt.Icon
		}
	}
	return "🙂"
}

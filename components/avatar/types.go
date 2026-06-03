package avatar

import "strings"

import "github.com/a-h/templ"

// Size represents avatar size variants
type Size string

const (
	SizeXS  Size = "xs"  // Extra small
	SizeSM  Size = "sm"  // Small
	SizeMD  Size = "md"  // Medium (default)
	SizeLG  Size = "lg"  // Large
	SizeXL  Size = "xl"  // Extra large
	Size2XL Size = "2xl" // 2x extra large (matches PenguinUI)
)

// Variant represents avatar style variants
type Variant string

const (
	Default   Variant = "default"
	Inverse   Variant = "inverse"
	Primary   Variant = "primary"
	Secondary Variant = "secondary"
	Info      Variant = "info"
	Success   Variant = "success"
	Warning   Variant = "warning"
	Danger    Variant = "danger"
)

// Shape represents avatar shape
type Shape string

const (
	ShapeCircle Shape = "circle" // Rounded full (default)
	ShapeSquare Shape = "square" // Rounded corners
)

// Radius represents square-avatar corner radius variants.
type Radius string

const (
	RadiusDefault Radius = ""     // Default square radius
	RadiusNone    Radius = "none" // No radius
	RadiusXS      Radius = "xs"   // Extra small radius
	RadiusSM      Radius = "sm"   // Small radius
	RadiusMD      Radius = "md"   // Medium radius
	RadiusLG      Radius = "lg"   // Large radius
)

// Status represents online/offline status indicator
type Status string

const (
	StatusOffline Status = "offline"
	StatusInfo    Status = "info"
	StatusSuccess Status = "success"
	StatusWarning Status = "warning"
	StatusDanger  Status = "danger"
)

// Config holds configuration for the avatar.
//
// The avatar renders in 3 layers (bottom to top):
//  1. Initials layer — always rendered as the base fallback
//  2. Loading layer — spinner shown while the image loads (only when Src is set)
//  3. Image layer — the actual image (only when Src is set, hidden on load error)
//
// When Src is provided, Alpine.js handles image load/error events:
// on load success the spinner hides, on error both image and spinner hide
// (falling back to the initials layer) and a console warning is logged.
type Config struct {
	// Src is the image URL. When set, the image and loading layers are rendered.
	Src string
	// Alt is the alt text for accessibility
	Alt string
	// Name is used to auto-derive initials via GetInitials(Name, "").
	// Ignored if Initials is set explicitly.
	Name string
	// Initials are displayed as the base fallback (e.g., "JS").
	// If empty, derived from Name via GetInitials.
	Initials string
	// Size of the avatar
	Size Size
	// Variant determines the color scheme (for initials/icon placeholders)
	Variant Variant
	// Shape of the avatar (circle or square)
	Shape Shape
	// Radius controls the corner radius for square avatars.
	Radius Radius
	// Border adds a colored border (for image avatars)
	Border bool
	// BorderColor is the border color (defaults to variant color if empty)
	BorderColor string
	// Status adds a status indicator dot
	Status Status
	// Icon is an optional icon component (replaces initials in the base layer)
	Icon templ.Component
	// Class allows additional CSS classes
	Class string
	// SrcExpr is an Alpine expression evaluated in the parent scope that yields
	// the image src at runtime (e.g. "avatarSrc"). When set, the image layer is
	// always rendered and binds x-bind:src to this expression; the initials
	// layer shows whenever the expression is falsy. Use for client-set sources
	// (object URLs, late-loaded images) where no static Src exists at render.
	// SrcExpr is a trusted developer-authored Alpine expression — never interpolate untrusted/user input into it (it is evaluated as JavaScript on the client).
	SrcExpr string
	// Reactive defers the size + status-indicator size classes to the parent
	// Alpine scope. When true, the avatar root and status dot read their size
	// class from `avatarSizeClass` and `avatarStatusSizeClass` respectively
	// (via x-bind:class). This lets a demo page or container drive multiple
	// avatars from a single Alpine state — e.g. an interactive size selector.
	//
	// The outer Alpine scope must define both expressions, for example:
	//
	//   x-data="{
	//     selected: 'md',
	//     sizeMap: { xs:'size-8 text-xs', md:'size-14 text-2xl', ... },
	//     statusSizeMap: { xs:'size-2', md:'size-4', ... },
	//     get avatarSizeClass() { return this.sizeMap[this.selected]; },
	//     get avatarStatusSizeClass() { return this.statusSizeMap[this.selected]; },
	//   }"
	Reactive bool
	// ReactiveRadius defers square-avatar radius classes to the parent Alpine
	// scope via `avatarRadiusClass`. Use with ShapeSquare for live radius demos.
	ReactiveRadius bool
}

// StackConfig holds configuration for a stacked avatar group.
type StackConfig struct {
	// Items are rendered as overlapping avatars from left to right.
	Items []Config
	// Label describes the group for assistive technology.
	Label string
	// Class allows additional CSS classes on the stack root.
	Class string
	// Reactive defers each avatar's size classes to the parent Alpine scope.
	Reactive bool
}

// ResolvedInitials returns the initials to display: explicit Initials, or derived from Name.
func (cfg Config) ResolvedInitials() string {
	if cfg.Initials != "" {
		return cfg.Initials
	}
	if cfg.Name != "" {
		return GetInitials(cfg.Name, "")
	}
	return "?"
}

// SizeClasses returns the CSS classes for the size
func (cfg Config) SizeClasses() string {
	switch cfg.Size {
	case SizeXS:
		return "size-8 text-xs"
	case SizeSM:
		return "size-10 text-sm"
	case SizeMD:
		return "size-14 text-2xl"
	case SizeLG:
		return "size-20 text-3xl"
	case SizeXL:
		return "size-24 text-4xl"
	case Size2XL:
		return "size-32 text-5xl"
	default:
		return "size-14 text-2xl"
	}
}

// ShapeClasses returns the CSS classes for the shape
func (cfg Config) ShapeClasses() string {
	switch cfg.Shape {
	case ShapeSquare:
		return cfg.RadiusClasses()
	default:
		return "rounded-full"
	}
}

// RadiusClasses returns the CSS classes for square avatar radius.
func (cfg Config) RadiusClasses() string {
	switch cfg.Radius {
	case RadiusNone:
		return "rounded-none"
	case RadiusXS:
		return "rounded-xs"
	case RadiusSM:
		return "rounded-sm"
	case RadiusLG:
		return "rounded-lg"
	default:
		return "rounded-md"
	}
}

// VariantClasses returns the CSS classes for the variant (for initials/icon)
func (cfg Config) VariantClasses() string {
	switch cfg.Variant {
	case Inverse:
		return "border border-outline-dark bg-surface-dark-alt text-on-surface-dark/80 dark:border-outline dark:bg-surface-alt dark:text-on-surface/80"
	case Primary:
		return "border border-primary bg-primary text-on-primary/80 dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark/80"
	case Secondary:
		return "border border-secondary bg-secondary text-on-secondary/80 dark:border-secondary-dark dark:bg-secondary-dark dark:text-on-secondary-dark/80"
	case Info:
		return "border border-info bg-info text-on-info/80"
	case Success:
		return "border border-success bg-success text-on-success/80"
	case Warning:
		return "border border-warning bg-warning text-on-warning/80"
	case Danger:
		return "border border-danger bg-danger text-on-danger/80"
	default:
		return "border border-outline bg-surface-alt text-on-surface/80 dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark/80"
	}
}

// VariantFillClasses returns variant classes without border color.
func (cfg Config) VariantFillClasses() string {
	switch cfg.Variant {
	case Inverse:
		return "bg-surface-dark-alt text-on-surface-dark/80 dark:bg-surface-alt dark:text-on-surface/80"
	case Primary:
		return "bg-primary text-on-primary/80 dark:bg-primary-dark dark:text-on-primary-dark/80"
	case Secondary:
		return "bg-secondary text-on-secondary/80 dark:bg-secondary-dark dark:text-on-secondary-dark/80"
	case Info:
		return "bg-info text-on-info/80"
	case Success:
		return "bg-success text-on-success/80"
	case Warning:
		return "bg-warning text-on-warning/80"
	case Danger:
		return "bg-danger text-on-danger/80"
	default:
		return "bg-surface-alt text-on-surface/80 dark:bg-surface-dark-alt dark:text-on-surface-dark/80"
	}
}

// BorderClasses returns border classes if border is enabled
func (cfg Config) BorderClasses() string {
	if !cfg.Border {
		return ""
	}

	color := cfg.BorderColor
	if color == "" {
		switch cfg.Variant {
		case Info:
			color = "border-info"
		case Success:
			color = "border-success"
		case Warning:
			color = "border-warning"
		case Danger:
			color = "border-danger"
		default:
			color = "border-primary"
		}
	}

	return "border-2 " + color + " p-0.5"
}

// StatusClasses returns status dot classes
func (cfg Config) StatusClasses() string {
	switch cfg.Status {
	case StatusOffline:
		return "bg-outline dark:bg-outline-dark"
	case StatusInfo:
		return "bg-info"
	case StatusSuccess:
		return "bg-success"
	case StatusWarning:
		return "bg-warning"
	case StatusDanger:
		return "bg-danger"
	default:
		return ""
	}
}

// StatusSizeClasses returns status dot size based on avatar size
func (cfg Config) StatusSizeClasses() string {
	switch cfg.Size {
	case SizeXS:
		return "size-2"
	case SizeSM:
		return "size-2.5"
	case SizeLG:
		return "size-5"
	case SizeXL:
		return "size-6"
	case Size2XL:
		return "size-7"
	default:
		return "size-4"
	}
}

// SpinnerSizeClasses returns spinner size classes based on avatar size
func (cfg Config) SpinnerSizeClasses() string {
	switch cfg.Size {
	case SizeXS:
		return "size-4"
	case SizeSM:
		return "size-5"
	case SizeLG:
		return "size-8"
	case SizeXL:
		return "size-10"
	case Size2XL:
		return "size-12"
	default:
		return "size-6"
	}
}

// HasImage returns true if avatar uses an image
func (cfg Config) HasImage() bool {
	return cfg.Src != "" || cfg.SrcExpr != ""
}

// HasInitials returns true if avatar uses initials
func (cfg Config) HasInitials() bool {
	return cfg.Initials != ""
}

// GetInitials derives 1-2 character initials from a name, falling back to email.
// Splits on whitespace, hyphens, and underscores.
// Examples: "John Doe" → "JD", "dev-ops" → "DO", "Engineering" → "EN", "" with "alice@x.com" → "AL", "" with "" → "?"
func GetInitials(name, email string) string {
	if name != "" {
		parts := splitWords(name)
		if len(parts) >= 2 {
			return toUpper(parts[0][0]) + toUpper(parts[1][0])
		}
		if len(parts) == 1 && len(parts[0]) >= 2 {
			return toUpper(parts[0][0]) + toUpper(parts[0][1])
		}
		if len(parts) == 1 && len(parts[0]) == 1 {
			return toUpper(parts[0][0])
		}
	}
	if len(email) >= 2 {
		return toUpper(email[0]) + toUpper(email[1])
	}
	if len(email) == 1 {
		return toUpper(email[0])
	}
	return "?"
}

func toUpper(b byte) string {
	if b >= 'a' && b <= 'z' {
		return string(b - 32)
	}
	return string(b)
}

// jsEscapeSingle escapes a string for safe embedding inside a single-quoted
// JS string literal. Without this, a Src containing a single quote (data URIs
// frequently do, e.g. `viewBox='0 0 64 64'`) would terminate the string early
// and break the x-on:error handler with a SyntaxError at Alpine compile time.
func jsEscapeSingle(s string) string {
	var result strings.Builder
	for _, c := range s {
		switch c {
		case '\'':
			result.WriteString(`\'`)
		case '\\':
			result.WriteString(`\\`)
		default:
			result.WriteString(string(c))
		}
	}
	return result.String()
}

func splitWords(s string) []string {
	var result []string
	var current string
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '-' || s[i] == '_' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(s[i])
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

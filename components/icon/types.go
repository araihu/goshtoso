package icon

import (
	"fmt"
	"strings"

	"github.com/araihu/goshtoso/internal/iconurl"
)

// Symbol identifies an SVG symbol in a sprite.
type Symbol string

// Mode selects where an icon resolves its symbol reference.
type Mode string

const (
	// ModeExternal resolves the symbol from SpriteURL. It is the default.
	ModeExternal Mode = ""
	// ModeInline resolves the symbol from the current document.
	ModeInline Mode = "inline"
)

// Size controls fixed icon dimensions.
type Size string

const (
	// SizeXS renders a 0.75rem icon.
	SizeXS Size = "xs"
	// SizeSM renders a 1rem icon.
	SizeSM Size = "sm"
	// SizeMD renders a 1.25rem icon. It is the default.
	SizeMD Size = "md"
	// SizeLG renders a 1.5rem icon.
	SizeLG Size = "lg"
	// SizeXL renders a 2rem icon.
	SizeXL Size = "xl"
)

// Config configures a sprite icon.
type Config struct {
	// SpriteURL identifies the external sprite when Mode is ModeExternal. Prefer a relative same-origin URL. Cross-origin external <use> references depend on browser and CORS compatibility. HTTPS pages cannot reliably load HTTP sprite URLs because browsers may block mixed content.
	SpriteURL string
	// Symbol identifies the sprite symbol to render.
	Symbol Symbol
	// Size controls the root SVG dimensions.
	Size Size
	// Label gives the icon an accessible image name.
	Label string
	// Decorative hides the icon from assistive technology, even when Label is set.
	Decorative bool
	// RootClass appends CSS classes to the root SVG.
	RootClass string
	// Mode selects external sprite or document-local symbol resolution.
	Mode Mode
}

func (cfg Config) validate() error {
	if !validSymbol(cfg.Symbol) {
		return fmt.Errorf("icon: invalid symbol %q", cfg.Symbol)
	}

	switch cfg.Mode {
	case ModeExternal:
		if err := iconurl.ValidateSpriteURL(cfg.SpriteURL); err != nil {
			return fmt.Errorf("icon: %w", err)
		}
		return nil
	case ModeInline:
		return nil
	default:
		return fmt.Errorf("icon: invalid mode %q", cfg.Mode)
	}
}

func validSymbol(symbol Symbol) bool {
	if symbol == "" {
		return false
	}

	previousHyphen := false
	for index := 0; index < len(symbol); index++ {
		character := symbol[index]
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') {
			previousHyphen = false
			continue
		}
		if character == '-' && index > 0 && index < len(symbol)-1 && !previousHyphen {
			previousHyphen = true
			continue
		}
		return false
	}
	return true
}

func (cfg Config) href() string {
	if cfg.Mode == ModeInline {
		return "#" + string(cfg.Symbol)
	}
	return cfg.SpriteURL + "#" + string(cfg.Symbol)
}

func (cfg Config) decorative() bool {
	return cfg.Decorative || strings.TrimSpace(cfg.Label) == ""
}

func (cfg Config) label() string {
	return strings.TrimSpace(cfg.Label)
}

func (cfg Config) classes() string {
	base := "size-5"
	switch cfg.Size {
	case SizeXS:
		base = "size-3"
	case SizeSM:
		base = "size-4"
	case SizeLG:
		base = "size-6"
	case SizeXL:
		base = "size-8"
	}
	if rootClass := strings.TrimSpace(cfg.RootClass); rootClass != "" {
		return base + " " + rootClass
	}
	return base
}

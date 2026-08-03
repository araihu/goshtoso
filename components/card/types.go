package card

import "github.com/a-h/templ"

// Appearance represents the card's visual treatment.
type Appearance string

const (
	AppearanceDefault Appearance = ""
	AppearancePrimary Appearance = "primary"
)

// Interaction selects optional pointer and keyboard motion for a card.
type Interaction string

const (
	InteractionDefault Interaction = ""
	// InteractionPressed gives linked and clickable cards a physical press response.
	InteractionPressed Interaction = "pressed"
)

// Layout represents card layout
type Layout string

const (
	LayoutVertical   Layout = "vertical"   // Default (image top, content bottom)
	LayoutHorizontal Layout = "horizontal" // Side by side
)

// Config holds configuration for the card
type Config struct {
	// Image is the card image URL
	Image string
	// ImageAlt is the image alt text
	ImageAlt string
	// Media replaces Image with arbitrary card media or decorative content.
	Media templ.Component
	// MediaClass allows additional CSS classes on the media container.
	MediaClass string
	// Tag is an optional category/tag (shown above title)
	Tag string
	// Title is the card title
	Title string
	// Description is the card body text
	Description string
	// Body renders arbitrary content between Description and Footer.
	Body templ.Component
	// Footer is optional footer content (buttons, links, etc.)
	Footer templ.Component
	// Appearance determines the card's visual treatment.
	Appearance Appearance
	// Layout determines vertical or horizontal layout
	Layout Layout
	// Interaction determines optional card motion.
	Interaction Interaction
	// RootClass allows additional CSS classes on the card root.
	RootClass string
}

// ContainerClasses returns the container CSS classes
func (cfg Config) containerClasses() string {
	base := "group flex rounded-radius overflow-hidden border bg-surface-alt text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark"

	// Appearance
	if cfg.Appearance == AppearancePrimary {
		base += " border-2 border-primary dark:border-primary-dark"
	} else {
		base += " border-outline"
	}

	if cfg.Interaction == InteractionPressed {
		base += " shadow-lg transition-[translate,box-shadow] duration-150 ease-out hover:translate-y-1.5 hover:shadow-sm active:translate-y-2 active:shadow-none motion-reduce:hover:translate-none motion-reduce:active:translate-none motion-reduce:transition-none"
	}

	// Layout
	if cfg.Layout == LayoutHorizontal {
		base += " max-w-2xl grid grid-cols-1 md:grid-cols-8"
	} else {
		base += " max-w-sm flex-col"
	}

	return base + " " + cfg.RootClass
}

// ImageContainerClasses returns the image container classes
func (cfg Config) imageContainerClasses() string {
	if cfg.Layout == LayoutHorizontal {
		return "col-span-3 overflow-hidden " + cfg.MediaClass
	}
	return "h-44 md:h-64 overflow-hidden " + cfg.MediaClass
}

// ImageClasses returns the image classes
func (cfg Config) imageClasses() string {
	if cfg.Layout == LayoutHorizontal {
		return "h-52 md:h-full w-full object-cover transition duration-700 ease-out group-hover:scale-105"
	}
	return "object-cover transition duration-700 ease-out group-hover:scale-105"
}

// ContentClasses returns the content container classes
func (cfg Config) contentClasses() string {
	if cfg.Layout == LayoutHorizontal {
		return "flex flex-col justify-center p-6 col-span-5"
	}
	return "flex flex-col gap-4 p-6"
}

// TagClasses returns the tag classes
func (cfg Config) tagClasses() string {
	return "text-sm font-medium"
}

// TitleClasses returns the title classes
func (cfg Config) titleClasses() string {
	return "text-balance text-xl lg:text-2xl font-bold text-on-surface-strong dark:text-on-surface-dark-strong"
}

// DescriptionClasses returns the description classes
func (cfg Config) descriptionClasses() string {
	return "text-pretty text-sm"
}

// HasImage returns true if card has an image
func (cfg Config) hasImage() bool {
	return cfg.Image != ""
}

func (cfg Config) hasMedia() bool {
	return cfg.Media != nil || cfg.hasImage()
}

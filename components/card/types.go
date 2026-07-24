package card

import "github.com/a-h/templ"

// Appearance represents the card's visual treatment.
type Appearance string

const (
	AppearanceDefault Appearance = ""
	AppearancePrimary Appearance = "primary"
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
	// Tag is an optional category/tag (shown above title)
	Tag string
	// Title is the card title
	Title string
	// Description is the card body text
	Description string
	// Footer is optional footer content (buttons, links, etc.)
	Footer templ.Component
	// Appearance determines the card's visual treatment.
	Appearance Appearance
	// Layout determines vertical or horizontal layout
	Layout Layout
	// RootClass allows additional CSS classes on the card root.
	RootClass string
}

// ContainerClasses returns the container CSS classes
func (cfg Config) ContainerClasses() string {
	base := "group flex rounded-radius overflow-hidden border bg-surface-alt text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark"

	// Appearance
	if cfg.Appearance == AppearancePrimary {
		base += " border-2 border-primary dark:border-primary-dark"
	} else {
		base += " border-outline"
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
func (cfg Config) ImageContainerClasses() string {
	if cfg.Layout == LayoutHorizontal {
		return "col-span-3 overflow-hidden"
	}
	return "h-44 md:h-64 overflow-hidden"
}

// ImageClasses returns the image classes
func (cfg Config) ImageClasses() string {
	if cfg.Layout == LayoutHorizontal {
		return "h-52 md:h-full w-full object-cover transition duration-700 ease-out group-hover:scale-105"
	}
	return "object-cover transition duration-700 ease-out group-hover:scale-105"
}

// ContentClasses returns the content container classes
func (cfg Config) ContentClasses() string {
	if cfg.Layout == LayoutHorizontal {
		return "flex flex-col justify-center p-6 col-span-5"
	}
	return "flex flex-col gap-4 p-6"
}

// TagClasses returns the tag classes
func (cfg Config) TagClasses() string {
	return "text-sm font-medium"
}

// TitleClasses returns the title classes
func (cfg Config) TitleClasses() string {
	return "text-balance text-xl lg:text-2xl font-bold text-on-surface-strong dark:text-on-surface-dark-strong"
}

// DescriptionClasses returns the description classes
func (cfg Config) DescriptionClasses() string {
	return "text-pretty text-sm"
}

// HasImage returns true if card has an image
func (cfg Config) HasImage() bool {
	return cfg.Image != ""
}
